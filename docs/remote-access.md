# Acceso Remoto y Seguridad

TelegramDL incorpora un servidor HTTP y WebSocket interno (por defecto en el puerto `8000`) que permite controlar el gestor tanto desde la propia aplicación de escritorio como desde el navegador web de cualquier teléfono móvil, tablet u otro ordenador.

---

## 🔐 Seguridad: Autenticación por Token Bearer

Toda la API REST (`/api/*`) y el flujo WebSocket exigen un **Token de Acceso**. Nadie en tu red local ni en internet puede consultar tus descargas, iniciar transferencias ni acceder a tus credenciales de Telegram sin este token.

### Gestión del Token
- **Generación Automática**: Se genera aleatoriamente en el primer arranque y se almacena en SQLite.
- **Visualización**: Puedes verlo y copiarlo en cualquier momento desde **Ajustes ➔ Acceso Remoto**.
- **Regeneración Instantánea**: Si sospechas que tu token fue comprometido, pulsa **"Regenerar token"** en Ajustes. El token anterior quedará invalidado de inmediato.

---

## 📱 Botón «Enviármelo a Telegram» (Saved Messages)

Si intentas acceder a TelegramDL desde tu smartphone fuera del ordenador y no recuerdas el token:

1. Abre `http://<ip-servidor>:8000` en el navegador de tu móvil.
2. En la pantalla de bloqueo de acceso remoto, pulsa el botón **"Enviármelo a Telegram"**.
3. TelegramDL enviará un mensaje directo con el token a tus **Mensajes Guardados** en Telegram.
4. Entra a tu app oficial de Telegram, copia el token y pégalo en el navegador para desbloquear el panel.

> [!TIP]
> **Protección contra Spam**: El endpoint de envío del token a Telegram implementa un límite estricto de **1 envío por minuto** y registra la dirección IP solicitante para evitar abusos.

---

## 🏠 Acceso en Red Local (LAN / Wi-Fi)

Por defecto, TelegramDL enlaza a `0.0.0.0`, lo que permite que cualquier dispositivo conectado a tu red Wi-Fi doméstica acceda al panel:

```env
TGDL_BIND_HOST=0.0.0.0
TGDL_PORT=8000
```

Para conectarte desde tu móvil o portátil:
1. Averigua la IP local de tu PC ejecutando `ipconfig` (Windows) o `ip a` / `ifconfig` (Linux/Mac) (ejemplo: `192.168.1.50`).
2. Abre en el navegador del móvil: `http://192.168.1.50:8000`.
3. Introduce el Token de acceso cuando te lo solicite.

> [!WARNING]
> Si estás conectado a una red pública (universidad, oficina, cafetería), cambia `TGDL_BIND_HOST=127.0.0.1` en tu archivo `.env` para que el puerto solo responda a peticiones del propio equipo.

---

## 🌐 Acceso desde Fuera de Casa (Internet)

> [!CAUTION]
> **Nunca abras el puerto 8000 en tu router (port-forwarding directo)** sin cifrado ni cortafuegos, ya que los escáneres automáticos de internet lo atacarán en minutos.

En su lugar, utiliza uno de los siguientes métodos seguros:

### Opción 1 (Recomendada): Tailscale (VPN de Malla)
[Tailscale](https://tailscale.com) crea una red privada segura punto a punto entre tus dispositivos sin abrir puertos en el router:

1. Instala Tailscale en el ordenador donde corre TelegramDL y en tu móvil.
2. Inicia sesión con la misma cuenta en ambos.
3. Copia la IP de Tailscale de tu PC (suele empezar por `100.x.y.z`).
4. Abre desde el navegador del móvil: `http://100.x.y.z:8000` e introduce tu token.

### Opción 2: Cloudflare Tunnel + Zero Trust Access
Si prefieres una URL pública sin instalar clientes VPN:
1. Instala `cloudflared` en tu servidor o PC.
2. Crea un túnel apuntando a `http://127.0.0.1:8000`.
3. Configura una regla de **Cloudflare Access** que obligue a iniciar sesión con tu correo electrónico mediante código de un solo uso antes de llegar a la aplicación.
