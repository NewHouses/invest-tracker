#!/usr/bin/env bash
# Compila invest-tracker: frontend (lint, tests, build) e Go (vet, tests, build).
#
# Deixa os binarios para Windows (invest-tracker.exe) e macOS
# (invest-tracker-macos-arm64 para Apple Silicon e invest-tracker-macos-amd64
# para Intel), todos coa interface web embebida. Usa o Node portátil de
# .tools/node se existe; se non, o do sistema (require Node 22.22 ou superior).
#
# Uso: ./build.sh [--skip-tests] [--skip-frontend] [--clean-install] [--installers]
#   --skip-tests     non executa os tests (nin de Go nin do frontend)
#   --skip-frontend  reutiliza a web xa compilada en internal/web/dist
#   --clean-install  reinstala as dependencias do frontend con "npm ci"
#   --installers     xera tamén a app de macOS en dist-installers (o instalador
#                    de Windows xérase en Windows con build.ps1 -Installers)
set -euo pipefail

cd "$(dirname "$0")"

skip_tests=0
skip_frontend=0
clean_install=0
installers=0
for arg in "$@"; do
  case "$arg" in
    --skip-tests) skip_tests=1 ;;
    --skip-frontend) skip_frontend=1 ;;
    --clean-install) clean_install=1 ;;
    --installers) installers=1 ;;
    -h | --help)
      sed -n '2,14p' "$0"
      exit 0
      ;;
    *)
      echo "Opción descoñecida: $arg (usa --help)" >&2
      exit 2
      ;;
  esac
done

# step executa un paso e para o script se falla.
step() {
  local name=$1
  shift
  printf '\033[36m==> %s\033[0m\n' "$name"
  if ! "$@"; then
    printf "Fallou o paso '%s'.\n" "$name" >&2
    exit 1
  fi
}

if [ "$skip_frontend" -eq 0 ]; then
  if [ -x .tools/node/bin/node ]; then
    PATH="$PWD/.tools/node/bin:$PATH"
  elif [ -x .tools/node/node ] || [ -x .tools/node/node.exe ]; then
    PATH="$PWD/.tools/node:$PATH"
  fi
  export PATH
  if ! command -v node >/dev/null 2>&1; then
    echo "Non se atopou Node.js. Instala Node 24 LTS ou copia un Node portátil en .tools/node." >&2
    exit 1
  fi
  node_version=$(node --version)
  node_version=${node_version#v}
  node_major=${node_version%%.*}
  node_minor=${node_version#*.}
  node_minor=${node_minor%%.*}
  if [ "$node_major" -lt 22 ] || { [ "$node_major" -eq 22 ] && [ "$node_minor" -lt 22 ]; }; then
    echo "Node $node_version é demasiado antigo: o frontend require Node 22.22 ou superior." >&2
    exit 1
  fi
  printf '\033[90mNode %s\033[0m\n' "$node_version"

  (
    cd frontend
    if [ "$clean_install" -eq 1 ] || [ ! -d node_modules ]; then
      step 'npm ci' npm ci
    fi
    step 'npm run lint' npm run lint
    if [ "$skip_tests" -eq 0 ]; then
      step 'npm test' npm test
    fi
    step 'npm run build' npm run build
  )
fi

step 'go vet' go vet ./...
if [ "$skip_tests" -eq 0 ]; then
  step 'go test' go test ./...
fi

# Binarios que se distribúen no repositorio. -trimpath e -buildvcs=false fan
# que o mesmo código xere sempre o mesmo binario (e non levan rutas locais);
# -s -w quitan os símbolos de depuración para que pesen menos.
build() {
  step "go build $3" env GOOS="$1" GOARCH="$2" CGO_ENABLED=0 \
    go build -trimpath -buildvcs=false -ldflags='-s -w' -o "$3" .
}
build windows amd64 invest-tracker.exe
build darwin arm64 invest-tracker-macos-arm64
build darwin amd64 invest-tracker-macos-amd64
chmod +x invest-tracker-macos-arm64 invest-tracker-macos-amd64

printf '\033[32mListo: invest-tracker.exe, invest-tracker-macos-arm64, invest-tracker-macos-amd64\033[0m\n'

if [ "$installers" -eq 1 ]; then
  step 'instalador de macOS' go run ./installers/macos
  echo 'O instalador de Windows xérase en Windows con build.ps1 -Installers (require Inno Setup).'
fi
