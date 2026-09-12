# Acceso remoto a TelegramDL

TelegramDL expone un servidor HTTP/WebSocket (por defecto en `0.0.0.0:8000`)
que el panel usa tanto dentro de la ventana de escritorio como si lo abres en
un navegador. Toda la API (`/api/*`) exige un **token de acceso** (Ajustes →
Acceso remoto, dentro de la app), así que nadie puede controlar tus descargas
ni leer tus credenciales de Telegram sin ese token.

## Acceso desde tu propia red (WiFi de casa)

`0.0.0.0` significa que el panel atiende en todas las interfaces del equipo, no
solo en la de loopback. Es lo que permite abrirlo desde el móvil o desde otro
ordenador de la casa escribiendo `http://<ip-local-de-la-pc>:8000`; la primera
vez te pedirá el token. Ese valor se escribe en `~/.tgdown/.env` la primera vez
que arranca la aplicación:

```
TGDL_BIND_HOST=0.0.0.0
```

Si prefieres que el panel solo sea accesible desde el propio ordenador, cambia
esa línea a `TGDL_BIND_HOST=127.0.0.1` y reinicia. La aplicación nunca
sobrescribe ese valor una vez elegido.

Ten en cuenta que con `0.0.0.0` el puerto queda visible para **cualquier
dispositivo conectado a tu red local**, no solo para los tuyos. En un WiFi
doméstico es razonable; en una red compartida (residencia, oficina, WiFi
público) conviene usar `127.0.0.1`.

## Acceso desde fuera de casa

Eso resuelve la red local, pero **no** cómo llegar al servidor desde internet.
Para eso, evita a toda costa abrir el puerto 8000 al exterior con
port-forwarding en el router: un puerto así recibe tráfico de escáneres
automáticos en cuestión de minutos. En su lugar, usa una de estas dos opciones.

## Opción recomendada: Tailscale (VPN de malla)

[Tailscale](https://tailscale.com) crea una red privada cifrada entre tus
dispositivos sin abrir ningún puerto en el router.

1. Instala Tailscale en la PC donde corre TelegramDL y en tu celular (o
   cualquier otro dispositivo desde el que quieras controlarlo).
2. Inicia sesión con la misma cuenta en ambos.
3. En la PC, revisa la IP de Tailscale asignada (suele ser `100.x.y.z`);
   Tailscale la muestra en su propio panel/bandeja del sistema.
4. Desde el celular (con Tailscale activo), abre en el navegador:
   `http://100.x.y.z:8000`
5. La primera vez te pedirá el token de acceso: ábrelo en la app de
   escritorio en **Ajustes → Acceso remoto**, cópialo y pégalo. El
   navegador lo recordará para la próxima vez.

Alternativas equivalentes si prefieres no depender de la infraestructura de
Tailscale: **WireGuard** (autogestionado) o **ZeroTier**.

## Alternativa: túnel público (Cloudflare Tunnel)

Si prefieres una URL pública en vez de instalar un cliente VPN en cada
dispositivo:

1. Instala `cloudflared` en la PC con TelegramDL.
2. Crea un túnel apuntando a `http://127.0.0.1:8000`.
3. En el dashboard de Cloudflare, activa **Cloudflare Access** sobre ese
   túnel y restringe el acceso a tu email (login por código de un solo uso)
   u otro proveedor de identidad.
4. Accede desde cualquier dispositivo a la URL pública que te da Cloudflare
   y, la primera vez, pega el token de acceso de la app.

Esto añade una capa de autenticación *antes* de que la petición llegue
siquiera a TelegramDL, además del token propio de la app.

## Qué hace el token y cómo se gestiona

- Se genera automáticamente la primera vez que arranca la app; no hace
  falta configurar nada para seguir usándola normalmente en el mismo
  equipo (la propia ventana de escritorio lo obtiene por su cuenta).
- Se ve y se puede copiar desde **Ajustes → Acceso remoto**.
- Si sospechas que se filtró (por ejemplo, lo compartiste sin querer),
  pulsa **Regenerar token** ahí mismo: el anterior deja de funcionar de
  inmediato y tendrás que volver a introducir el nuevo en cada dispositivo
  remoto.
- Nunca lo compartas por canales inseguros (chat sin cifrar, capturas de
  pantalla públicas, etc.): quien lo tenga puede controlar la app por
  completo mientras siga siendo válido.

## Qué NO hacer

- No abras el puerto 8000 en el router (port-forwarding) para "simplificar"
  el acceso desde fuera de casa. Aunque el token protege la API, estarías
  exponiendo el servicio a escaneos e intentos de fuerza bruta de todo
  internet sin necesidad, teniendo alternativas mejores arriba.
- No dejes `TGDL_BIND_HOST=0.0.0.0` si el equipo está en una red que no
  controlas (WiFi público, residencia, oficina compartida): ahí cualquiera
  conectado a esa red ve el puerto. Usa `127.0.0.1` en esos casos.
- No compartas el token por el mismo canal que uses para compartir el
  enlace/IP de acceso (si uno se filtra, que el otro no se filtre con él).
