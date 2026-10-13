#!/bin/bash
# Servidor web de Control de Investimentos, aberto pola app de macOS. Os datos
# gárdanse en ~/Library/Application Support/Control de Investimentos.
contents=$(cd "$(dirname "$0")/.." && pwd)
data="$HOME/Library/Application Support/Control de Investimentos"
mkdir -p "$data" && cd "$data" || exit 1
if [ "$(uname -m)" = "arm64" ]; then
  bin="$contents/MacOS/invest-tracker-arm64"
else
  bin="$contents/MacOS/invest-tracker-amd64"
fi
printf '\033]0;Control de Investimentos\007'
exec "$bin" web -db "$data/investimentos.db"
