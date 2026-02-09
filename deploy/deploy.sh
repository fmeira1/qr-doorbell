#!/usr/bin/env bash
# deploy.sh - Compila, copia y reinicia qr-doorbell en el VPS.
#
# Uso desde tu máquina local:
#   ./deploy/deploy.sh usuario@ip-del-vps
#
# Requiere: Go instalado localmente, ssh configurado con tu VPS.
set -euo pipefail

if [ $# -lt 1 ]; then
  echo "Uso: $0 <usuario@host>"
  echo "Ejemplo: $0 root@123.45.67.89"
  exit 1
fi

SSH_TARGET="$1"
APP_DIR="/opt/qr-doorbell"
PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "==> Compilando binario para Linux amd64..."
cd "$PROJECT_ROOT"
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o qr-doorbell ./cmd/server

echo "==> Subiendo binario y archivos web al VPS..."
scp qr-doorbell "$SSH_TARGET:$APP_DIR/qr-doorbell"
scp -r web/public "$SSH_TARGET:$APP_DIR/web/public"

echo "==> Reiniciando servicio..."
ssh "$SSH_TARGET" "sudo systemctl restart qr-doorbell"

echo "==> Verificando estado..."
ssh "$SSH_TARGET" "sudo systemctl status qr-doorbell --no-pager"

# Limpiar binario local
rm -f qr-doorbell

echo ""
echo "===== Deploy completado ====="
