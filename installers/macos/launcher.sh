#!/bin/bash
# Lanzador da app de macOS de Control de Investimentos. Abre o servidor web
# nunha xanela do Terminal (para ver o que fai e pechalo con Ctrl+C) ou, se xa
# está aberto, só o navegador.
contents=$(cd "$(dirname "$0")/.." && pwd)
app=$(dirname "$contents")

# A app non está asinada: unha vez que o usuario a autoriza, quítase a marca de
# "descargado de Internet" ao resto dos seus ficheiros para que o Terminal os
# poida executar.
xattr -dr com.apple.quarantine "$app" 2>/dev/null

if curl -fsS -m 2 -o /dev/null http://localhost:8080/api/meta 2>/dev/null; then
  open http://localhost:8080
else
  open -a Terminal "$contents/Resources/servidor.command"
fi
