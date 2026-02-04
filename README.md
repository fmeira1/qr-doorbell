## qr-doorbell

Proyecto personal para reemplazar a Bellify y sus elevados costos al agregar más unidades funcionales.  
Consiste en un backend en Go + un frontend simple para un "timbre por QR" con WebSockets y, opcionalmente, integración con Telegram.

### Estructura rápida

- **cmd/server/main.go**: arranque del servidor HTTP.
- **internal/api**: router y handlers HTTP.
- **internal/config**: carga de configuración desde variables de entorno.
- **internal/doorbell**: servicio de timbres.
- **internal/session**: servicio de sesiones.
- **internal/message**: servicio de mensajes + broadcast WS + Telegram.
- **internal/realtime**: Hub de WebSockets.
- **internal/notify**: integración con Telegram.
- **internal/user**: usuarios (propietarios por unidad y admin).
- **internal/auth**: login, JWT, cambio de contraseña.
- **internal/admin**: alta de usuarios (solo admin).
- **web/public**: frontend estático (UI del chat, login, admin).

### Configuración

La aplicación lee configuración desde variables de entorno (todas opcionales salvo las que indiquen lo contrario):

- **PORT**: puerto HTTP. Por defecto `8081`.
- **BASE_URL**: URL pública del servicio, usada en los links de Telegram.  
  Por defecto `http://localhost:PORT` (ej. `http://localhost:8081`).
- **TELEGRAM_TOKEN**: token del bot de Telegram.
- **TELEGRAM_CHAT_ID**: chat ID o canal donde se enviarán las alertas.
- **JWT_SECRET**: clave para firmar las sesiones (obligatorio para usar login y panel admin).
- **ADMIN_EMAIL** y **ADMIN_PASSWORD**: email y contraseña del primer administrador. Si están definidos, al arrancar se crea o actualiza ese usuario como admin (así podés entrar a `login.html` y luego a `admin.html` para dar de alta a los dueños por departamento).
- **DATABASE_URL**: si está vacía, sesiones y mensajes van en memoria; si está definida, se usa Postgres (incluyendo la tabla `users`).
- **DOORBELL_LAT**, **DOORBELL_LON**, **GEOFENCE_RADIUS_METERS**: geovalla. Si definís los tres (latitud y longitud del timbre en grados decimales, y radio en metros), solo se podrá tocar el timbre estando dentro de ese radio. Si no los configurás, no se pide ubicación.

Si `TELEGRAM_TOKEN` o `TELEGRAM_CHAT_ID` están vacíos, el sistema funciona igual pero **no envía notificaciones de Telegram**.  
Si **JWT_SECRET** está vacío, las rutas `/auth/*` y `/admin/*` no se registran (no hay login ni panel de usuarios).

### Cómo ejecutar en local

1. **Instalar dependencias**
   ```bash
   go mod tidy
   ```

2. **(Opcional) Exportar variables de entorno**
   - Windows PowerShell:
     ```powershell
     $env:PORT="8081"
     $env:BASE_URL="http://localhost:8081"
     # Solo si quieres Telegram real:
     $env:TELEGRAM_TOKEN="TU_TOKEN"
     $env:TELEGRAM_CHAT_ID="TU_CHAT_ID"
     ```

3. **Levantar el servidor**
   ```bash
   go run ./cmd/server
   ```
   El servidor escuchará en `http://localhost:PORT` (por defecto `http://localhost:8081`).

4. **Abrir el frontend**
   - **Timbre (visitante/propietario)**: `http://localhost:8081/` o `http://localhost:8081/index.html`
   - **Login (usuarios y admin)**: `http://localhost:8081/login.html`
   - **Panel admin** (solo si iniciaste sesión como admin): después de login ir a `admin.html` para dar de alta emails con contraseña genérica por departamento.
   - **Cambiar contraseña**: `http://localhost:8081/change-password.html` (cualquier usuario logueado).

   En la pantalla del timbre:
   - **Visitante**: modo por defecto, sin `session_id` en la URL.
   - **Propietario**: URL con `?session_id=...` (p. ej. la que llega por Telegram), o bien iniciar sesión con tu usuario asociado a un departamento.

### Flujo de prueba sin Telegram

1. **Levanta el servidor**:
   ```bash
   go run ./cmd/server
   ```

2. **Abrí el navegador como visitante** en `http://localhost:8081/`.

3. **Iniciar una sesión simulando escanear un QR**:
   - Click en **"Iniciar sesión (simula QR)"**.
   - El frontend hace `POST /session/start` con `doorbell_id = "abc123"` (definido en `internal/doorbell`).
   - Se abre un WebSocket a `/ws/{session_id}` y se asocia el chat.

4. **Simular al propietario (sin Telegram)**:
   - Copiá el `session_id` que aparece en pantalla (por ejemplo `7bf6...`).
   - Abrí otra pestaña o navegador en:  
     `http://localhost:8081/index.html?session_id=EL_ID_COPIADO`
   - Esa vista entra en modo **Propietario**, carga el historial (`GET /messages?session_id=...`) y abre un WebSocket para la misma sesión.

5. **Enviar mensajes en ambos lados**:
   - En la pestaña del visitante: los mensajes se mandan como `sender = "visitor"`.
   - En la pestaña del propietario: se mandan como `sender = "owner"`.
   - Ambos reciben los mensajes en tiempo real vía WebSocket.

### Flujo con Telegram

1. Configura estas variables de entorno antes de arrancar el servidor:
   - `TELEGRAM_TOKEN`
   - `TELEGRAM_CHAT_ID`

2. Al llegar el **primer mensaje de un visitante** en una sesión, el backend:
   - Guarda el mensaje.
   - Envía una notificación a Telegram con:
     - Texto del mensaje.
     - Un link al chat web: `BASE_URL/index.html?session_id={ID}`.
     - Botones de respuesta rápida (ej. "Ya voy", "No puedo atender").

3. Si el propietario pulsa un botón en Telegram:
   - Telegram envía un callback al bot.
   - El poller de `internal/notify/telegram_polling.go` interpreta la acción
     y crea un mensaje dentro de la misma sesión, que se envía a la UI web
     como si fuera escrito por el propietario.

### Flujo con usuarios (admin y propietarios por departamento)

1. Configurá **JWT_SECRET**, **ADMIN_EMAIL** y **ADMIN_PASSWORD** y arrancá el servidor. Se crea (o actualiza) el usuario admin.

2. Entrá a **login.html** con ese email y contraseña. Si sos admin, te redirige a **admin.html**.

3. En **admin.html** agregás usuarios: email + contraseña genérica + departamento. Esos usuarios pueden entrar a **login.html**, iniciar sesión y usar **change-password.html** para cambiar la contraseña.

4. Cada usuario queda asociado a una **unidad** (departamento). Más adelante se puede usar esa asociación para notificar solo al dueño de esa unidad (p. ej. Web Push por unidad).
