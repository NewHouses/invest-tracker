#!/bin/bash
# Abre os menús de texto de Control de Investimentos en macOS: dobre clic no
# Finder. Execútase desde o cartafol do programa para usar sempre a mesma base
# de datos (se non, o Finder usaría o cartafol persoal).
cd "$(dirname "$0")" || exit 1
if [ "$(uname -m)" = "arm64" ]; then
  bin=./invest-tracker-macos-arm64
else
  bin=./invest-tracker-macos-amd64
fi
[ -x "$bin" ] || chmod +x "$bin"
exec "$bin" "$@"
