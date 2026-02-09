# Guía de Despliegue en VPS

Esta guía te ayudará a desplegar `qr-doorbell` en un VPS para uso en producción.

## Requisitos Previos

- VPS con Linux (Ubuntu 20.04+ recomendado)
- Acceso SSH al servidor
- Dominio apuntando al VPS (opcional pero recomendado)
- Postgres instalado (o usar un servicio gestionado)

## Paso 1: Preparar el Servidor

### Instalar Go (si no está instalado)

```bash
wget https://go.dev/dl/go1.21.5.linux-amd64.tar.gz
sudo rm -rf /usr/local/go
sudo tar -C /usr/local -xzf go1.21.5.linux-amd64.tar.gz
export PATH=$PATH:/usr/local/go/bin
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
```

### Instalar Postgres

```bash
sudo apt update
sudo apt install postgresql postgresql-contrib -y
sudo systemctl start postgresql
sudo systemctl enable postgresql
```

### Crear base de datos

```bash
sudo -u postgres psql
```

Dentro de psql:

```sql
CREATE DATABASE qrdoorbell;
CREATE USER qrdoorbell_user WITH PASSWORD 'tu_password_seguro';
GRANT ALL PRIVILEGES ON DATABASE qrdoorbell TO qrdoorbell_user;
\q
```

### Crear tablas

```bash
sudo -u postgres psql qrdoorbell
```

```sql
CREATE TABLE sessions (
  id          TEXT PRIMARY KEY,
  doorbell_id TEXT NOT NULL,
  unit_id     TEXT NOT NULL,
  status      TEXT NOT NULL,
  created_at  TIMESTAMPTZ NOT NULL,
  expires_at  TIMESTAMPTZ NOT NULL
);

CREATE TABLE messages (
  id         TEXT PRIMARY KEY,
  session_id TEXT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
  sender     TEXT NOT NULL,
  content    TEXT NOT NULL,
  created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_messages_session_id ON messages(session_id);
CREATE INDEX idx_sessions_unit_id ON sessions(unit_id);

-- Tabla de usuarios (propietarios y admin) para login y asociación a unidades
CREATE TABLE users (
  id            TEXT PRIMARY KEY,
  email         TEXT NOT NULL UNIQUE,
  password_hash TEXT NOT NULL,
  unit_id       TEXT NOT NULL DEFAULT '',
  is_admin      BOOLEAN NOT NULL DEFAULT false,
  created_at    TIMESTAMPTZ NOT NULL
);
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_unit_id ON users(unit_id);
```

Salir: `\q`

## Paso 2: Setup automático (recomendado)

El proyecto incluye scripts en la carpeta `deploy/` que automatizan la configuración:

```bash
# En el VPS, cloná el repo o copiá la carpeta deploy/
sudo bash deploy/setup.sh
```

Esto crea:
- Usuario de sistema `deploy` (sin shell de login)
- Directorio `/opt/qr-doorbell/` para el binario y archivos web
- Directorio `/etc/qr-doorbell/` con el archivo de variables de entorno
- Servicio systemd habilitado (pero no arrancado)

### Configurar variables de entorno

```bash
sudo nano /etc/qr-doorbell/env
```

Editá con tus valores reales. El archivo de ejemplo (`deploy/env.example`) tiene todas las variables documentadas:

- `PORT`, `BASE_URL` — puerto y URL pública
- `DATABASE_URL` — conexión a Postgres
- `JWT_SECRET`, `ADMIN_EMAIL`, `ADMIN_PASSWORD` — autenticación
- `TELEGRAM_TOKEN`, `TELEGRAM_CHAT_ID` — notificaciones (opcional)
- `DOORBELL_LAT`, `DOORBELL_LON`, `GEOFENCE_RADIUS_METERS` — geovalla (opcional)

> El archivo tiene permisos `600` (solo root puede leerlo) para proteger las credenciales.

## Paso 3: Compilar y desplegar

### Opción A: Deploy desde tu máquina local (recomendado)

```bash
./deploy/deploy.sh root@tu-vps-ip
```

Esto compila para Linux, sube el binario + archivos web por scp y reinicia el servicio.

### Opción B: Manual

En tu máquina local:

```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o qr-doorbell ./cmd/server
scp qr-doorbell usuario@tu-vps:/opt/qr-doorbell/
scp -r web/public usuario@tu-vps:/opt/qr-doorbell/web/public
```

En el VPS:

```bash
sudo systemctl restart qr-doorbell
```

## Paso 4: Verificar servicio systemd

```bash
# Estado
sudo systemctl status qr-doorbell

# Logs en tiempo real
sudo journalctl -u qr-doorbell -f

# Arrancar / parar / reiniciar
sudo systemctl start qr-doorbell
sudo systemctl stop qr-doorbell
sudo systemctl restart qr-doorbell
```

El servicio systemd (`deploy/qr-doorbell.service`) incluye:
- **Restart automático** (`Restart=always`, espera 5s entre intentos)
- **Seguridad** (`NoNewPrivileges=true`, usuario sin privilegios)
- **Logs centralizados** vía journald
- **Arranque con el sistema** (`WantedBy=multi-user.target`)

## Paso 5: Configurar Nginx como Reverse Proxy

Instalar Nginx:

```bash
sudo apt install nginx -y
```

Configurar sitio:

```bash
sudo nano /etc/nginx/sites-available/qr-doorbell
```

Contenido:

```nginx
server {
    listen 80;
    server_name timbre.tudominio.com;

    # Redirigir a HTTPS (después de configurar SSL)
    # return 301 https://$server_name$request_uri;

    location / {
        proxy_pass http://localhost:8081;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

Activar sitio:

```bash
sudo ln -s /etc/nginx/sites-available/qr-doorbell /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl restart nginx
```

## Paso 6: Configurar SSL con Let's Encrypt

```bash
sudo apt install certbot python3-certbot-nginx -y
sudo certbot --nginx -d timbre.tudominio.com
```

Esto configurará HTTPS automáticamente.

## Paso 7: Verificar que Todo Funciona

1. Ver logs del servicio:
   ```bash
   sudo journalctl -u qr-doorbell -f
   ```

2. Probar endpoints:
   ```bash
   curl http://localhost:8081/health
   curl https://timbre.tudominio.com/health
   ```

3. Abrir en navegador:
   - `https://timbre.tudominio.com/select-unit.html?doorbell_id=abc123`

## Paso 8: Configurar Edificios y Departamentos

Por ahora los edificios y departamentos están hardcodeados en `internal/building/service.go`. Para producción, deberías:

1. Crear una tabla `buildings` y `units` en Postgres
2. Crear endpoints/admin panel para gestionarlos
3. O simplemente editar `service.go` con tus departamentos reales

Ejemplo de edición rápida en `internal/building/service.go`:

```go
units := map[string]Unit{
    "unit-1": {
        ID:         "unit-1",
        BuildingID: buildingID,
        Number:     "PB A",
        Name:       "Planta Baja A",
        Active:     true,
    },
    // Agregar tus departamentos reales aquí
}
```

## Mantenimiento

### Ver logs
```bash
sudo journalctl -u qr-doorbell -f
```

### Reiniciar servicio
```bash
sudo systemctl restart qr-doorbell
```

### Actualizar aplicación

Desde tu máquina local:
```bash
./deploy/deploy.sh root@tu-vps-ip
```

O manualmente en el VPS:
```bash
cd /opt/qr-doorbell
# Subí el nuevo binario y archivos web, luego:
sudo systemctl restart qr-doorbell
```

## Seguridad Adicional

1. **Firewall**: Configurar UFW
   ```bash
   sudo ufw allow 22/tcp
   sudo ufw allow 80/tcp
   sudo ufw allow 443/tcp
   sudo ufw enable
   ```

2. **Rate Limiting**: Ya está implementado en el código, pero puedes ajustar los límites en `internal/api/router.go`

3. **Backups**: Configurar backups regulares de la base de datos:
   ```bash
   # Agregar a crontab
   0 2 * * * pg_dump -U qrdoorbell_user qrdoorbell > /backups/qrdoorbell_$(date +\%Y\%m\%d).sql
   ```

## Troubleshooting

- **Servicio no inicia**: Revisar logs con `sudo journalctl -u qr-doorbell`
- **No conecta a DB**: Verificar `DATABASE_URL` y que Postgres esté corriendo
- **WebSocket no funciona**: Verificar headers de proxy en Nginx
- **Telegram no envía**: Verificar `TELEGRAM_TOKEN` y `TELEGRAM_CHAT_ID`
