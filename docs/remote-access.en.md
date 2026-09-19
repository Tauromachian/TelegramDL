# Remote Access & Security

TelegramDL features an integrated HTTP and WebSocket server (defaulting to port `8000`), allowing you to control your downloads both from the native desktop GUI and from web browsers on smartphones, tablets, or remote computers.

---

## 🔐 Security: Bearer Token Authentication

Every REST endpoint (`/api/*`) and WebSocket handshake requires a valid **Access Token**. No unauthorized party on your local network or the internet can view downloads, trigger transfers, or inspect your credentials without this token.

### Managing the Token
- **Automatic Generation**: Randomly created during initial setup and stored securely in SQLite.
- **Viewing the Token**: Inspect or copy it anytime under **Settings ➔ Remote Access**.
- **Instant Revocation**: If your token was compromised, click **"Regenerate token"**. The previous token is immediately invalidated.

---

## 📱 "Send to Telegram" (Saved Messages Delivery)

If you are away from your PC and opening the web dashboard from your smartphone:

1. Open `http://<server-ip>:8000` in your mobile browser.
2. On the remote access login prompt, tap **"Send it to my Telegram"**.
3. TelegramDL dispatches a direct private message containing the token to your **Saved Messages** in Telegram.
4. Switch to your official Telegram app, copy the token, and paste it into your browser to unlock full control.

> [!TIP]
> **Anti-Spam Throttling**: The token delivery endpoint enforces a strict rate limit of **1 request per minute** and logs the requester IP address.

---

## 🏠 Local Network Access (LAN / Wi-Fi)

By default, TelegramDL binds to `0.0.0.0`, listening across all network interfaces:

```env
TGDL_BIND_HOST=0.0.0.0
TGDL_PORT=8000
```

To connect from your phone or secondary computer:
1. Find your machine's local IP via `ipconfig` (Windows) or `ip a` / `ifconfig` (Linux/macOS) (e.g. `192.168.1.50`).
2. Open in your phone browser: `http://192.168.1.50:8000`.
3. Provide your Access Token when prompted.

> [!WARNING]
> On public or shared networks (universities, coffee shops, open offices), set `TGDL_BIND_HOST=127.0.0.1` in your `.env` so only localhost can connect.

---

## 🌐 Remote Access over the Internet

> [!CAUTION]
> **Never expose port 8000 directly via router port-forwarding** without an encrypted tunnel or firewall, as automated internet scanners will probe it within minutes.

Instead, use one of these secure architectures:

### Option 1 (Recommended): Tailscale (Mesh VPN)
[Tailscale](https://tailscale.com) forms an encrypted point-to-point mesh network between your devices without port forwarding:

1. Install Tailscale on the host PC and your mobile phone.
2. Sign in using the same account on both devices.
3. Locate your host PC's Tailscale IP (typically starts with `100.x.y.z`).
4. In your phone's browser, visit `http://100.x.y.z:8000` and enter your token.

### Option 2: Cloudflare Tunnel + Zero Trust Access
For a public URL without client software:
1. Install `cloudflared` on your PC or server.
2. Configure a tunnel targeting `http://127.0.0.1:8000`.
3. Enable **Cloudflare Access** on the tunnel, requiring one-time passcode email login prior to accessing the service.
