#!/usr/bin/env bash
# setup.sh - Configura el VPS para correr qr-doorbell como servicio systemd.
# Ejecutar como root o con sudo.
set -euo pipefail

APP_USER="deploy"
APP_DIR="/opt/qr-doorbell"
ENV_DIR="/etc/qr-doorbell"

echo "==> Creando usuario $APP_USER (si no existe)"
id -u "$APP_USER" &>/dev/null || useradd --system --shell /usr/sbin/nologin "$APP_USER"

echo "==> Creando directorio $APP_DIR"
mkdir -p "$APP_DIR"
chown "$APP_USER:$APP_USER" "$APP_DIR"

echo "==> Creando directorio de configuración $ENV_DIR"
mkdir -p "$ENV_DIR"

if [ ! -f "$ENV_DIR/env" ]; then
  cp "$(dirname "$0")/env.example" "$ENV_DIR/env"
  chmod 600 "$ENV_DIR/env"
  chown root:root "$ENV_DIR/env"
  echo "==> Archivo de entorno creado en $ENV_DIR/env"
  echo "    EDITÁ $ENV_DIR/env con tus valores reales antes de arrancar."
else
  echo "==> $ENV_DIR/env ya existe, no se sobreescribe."
fi

echo "==> Instalando servicio systemd"
cp "$(dirname "$0")/qr-doorbell.service" /etc/systemd/system/qr-doorbell.service
systemctl daemon-reload
systemctl enable qr-doorbell

echo ""
echo "===== Setup completado ====="
echo ""
echo "Próximos pasos:"
echo "  1. Editá las variables de entorno:  sudo nano $ENV_DIR/env"
echo "  2. Copiá el binario y la carpeta web/ a $APP_DIR/"
echo "  3. Arrancá el servicio:             sudo systemctl start qr-doorbell"
echo "  4. Verificá:                        sudo systemctl status qr-doorbell"
echo "  5. Ver logs:                        sudo journalctl -u qr-doorbell -f"
echo ""
