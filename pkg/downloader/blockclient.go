package downloader

// Control de las peticiones de bloques de archivo a Telegram.
//
// Todo lo que pide bytes de un archivo —el downloader de gotd para una descarga
// nueva y downloadMissingParts para una reanudación— pasa por el mismo sitio:
// el envoltorio de este archivo. Así hay un único punto donde se limita cuántas
// peticiones van en vuelo y donde se atiende el FLOOD_WAIT, en vez de dos
// caminos con comportamientos distintos, que era la razón de que una descarga
// empezara rápida por un lado y siguiera arrastrándose por el otro.

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gotd/td/tg"

	"tgdown/pkg/logbus"
	"tgdown/pkg/telegram"
)

// maxBloquesEnVuelo es el tope de peticiones de bloque simultáneas de TODO el
// programa, sumando todos los workers de todas las descargas.
//
// Antes no había ninguno: con 6 descargas a la vez y 4 workers cada una se
// pedían 24 bloques al mismo tiempo, y con los topes de la configuración se
// podía llegar a 256. Pedir más de lo que las conexiones pueden servir no
// acelera nada —la cola se forma igual, solo que dentro del cliente— y es la
// forma más segura de que Telegram empiece a contestar FLOOD_WAIT.
//
// Dos por conexión es lo que hace falta para que ninguna se quede parada
// esperando la ida y vuelta de la anterior.
var maxBloquesEnVuelo = telegram.ConexionesDescarga * 2

// bloquesEnVuelo reparte esas ranuras entre todos los workers del programa.
var bloquesEnVuelo = make(chan struct{}, maxBloquesEnVuelo)

func adquirirRanura(ctx context.Context) error {
	select {
	case bloquesEnVuelo <- struct{}{}:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func liberarRanura() {
	select {
	case <-bloquesEnVuelo:
	default:
	}
}

// intentosPorFlood es cuántas veces se aguanta una pausa de Telegram para la
// misma petición antes de devolver el error hacia arriba.
const intentosPorFlood = 3

// puertaFlood para a TODOS los workers cuando Telegram pide esperar.
//
// El límite de Telegram es por cuenta y método, no por conexión ni por worker.
// Si uno se para y los demás siguen pidiendo, la cuenta sigue castigada y lo
// único que se consigue es que el castigo se renueve: cada worker se lleva su
// propia pausa, se duermen escalonados y la tubería nunca vuelve a llenarse.
// Eso es exactamente lo que se ve como «los workers no trabajan todos juntos».
// Parando todos a la vez, el tiempo se cumple una sola vez y la descarga
// recupera el ritmo completo de golpe.
var puertaFlood struct {
	mu    sync.Mutex
	hasta time.Time
}

// cerrarPuertaFlood deja a todos los workers parados hasta dentro de espera. Si
// ya había una pausa en curso más larga, se respeta la más larga.
func cerrarPuertaFlood(espera time.Duration) {
	limite := time.Now().Add(espera)
	puertaFlood.mu.Lock()
	if limite.After(puertaFlood.hasta) {
		puertaFlood.hasta = limite
	}
	puertaFlood.mu.Unlock()
}

// esperarPuertaFlood duerme lo que quede de la pausa global, si hay alguna.
func esperarPuertaFlood(ctx context.Context) error {
	for {
		puertaFlood.mu.Lock()
		hasta := puertaFlood.hasta
		puertaFlood.mu.Unlock()

		restante := time.Until(hasta)
		if restante <= 0 {
			return nil
		}
		if restante > esperaMaximaFlood {
			restante = esperaMaximaFlood
		}

		temporizador := time.NewTimer(restante)
		select {
		case <-ctx.Done():
			temporizador.Stop()
			return ctx.Err()
		case <-temporizador.C:
		}
	}
}

// clienteBloques es el cliente de Telegram con el control de bloques puesto.
//
// Lleva *tg.Client incrustado a propósito: así hereda tal cual todos los
// métodos que el downloader de gotd necesita (hashes, CDN, archivos web) y solo
// se cambia el que importa, UploadGetFile. Escribir a mano esa lista de métodos
// sería pedir que se rompa la compilación cada vez que gotd toque su interfaz.
type clienteBloques struct {
	*tg.Client
}

// envolverCliente prepara el cliente de bloques. Devuelve el original sin
// envolver si no hay cliente, para no esconder un nil detrás de un struct.
func envolverCliente(cliente *tg.Client) clienteBloques {
	return clienteBloques{Client: cliente}
}

// UploadGetFile pide un bloque respetando la pausa global y la ranura, y
// aguanta el FLOOD_WAIT en vez de dejarlo subir.
//
// Que esto viva aquí y no en el worker tiene una consecuencia importante: el
// downloader de gotd, que se usa en las descargas nuevas y que no sabe nada de
// FLOOD_WAIT, ahora también queda cubierto. Antes, una pausa de Telegram en una
// descarga nueva mataba el intento, y la descarga solo continuaba en el
// reintento, ya por el camino de reanudación.
func (c clienteBloques) UploadGetFile(ctx context.Context, request *tg.UploadGetFileRequest) (tg.UploadFileClass, error) {
	var (
		respuesta tg.UploadFileClass
		err       error
	)

	for intento := 0; intento < intentosPorFlood; intento++ {
		if errEspera := esperarPuertaFlood(ctx); errEspera != nil {
			return nil, errEspera
		}
		if errRanura := adquirirRanura(ctx); errRanura != nil {
			return nil, errRanura
		}

		respuesta, err = c.Client.UploadGetFile(ctx, request)
		liberarRanura()

		if err == nil {
			return respuesta, nil
		}

		espera, esFlood := esperaPorFlood(err)
		if !esFlood {
			return respuesta, err
		}

		// Se avisa en el registro porque es la única forma de saber desde el
		// panel que una descarga lenta es Telegram frenándonos y no un problema
		// del programa.
		logbus.Warn(logbus.CatDownloads,
			fmt.Sprintf("Telegram pide esperar %s antes de seguir", espera),
			fmt.Sprintf("Bloque en el byte %d · intento %d de %d · se para toda la descarga hasta que pase",
				request.Offset, intento+1, intentosPorFlood))

		// El segundo extra es margen para no volver justo en el borde y que nos
		// lo repitan. La espera la cumple la puerta al principio de la vuelta
		// siguiente, y la cumplen todos los workers a la vez.
		cerrarPuertaFlood(espera + time.Second)
	}

	return respuesta, err
}
