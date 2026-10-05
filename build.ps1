<#
.SYNOPSIS
    Compila invest-tracker: frontend (lint, tests, build) e Go (vet, tests, build).

.DESCRIPTION
    Deixa o resultado en invest-tracker.exe, que inclúe a interface web
    embebida. Usa o Node portátil de .tools\node se existe; se non, o do sistema
    (require Node 22.22 ou superior).

.EXAMPLE
    powershell -NoProfile -ExecutionPolicy Bypass -File .\build.ps1

.EXAMPLE
    powershell -NoProfile -ExecutionPolicy Bypass -File .\build.ps1 -SkipTests
#>
[CmdletBinding()]
param(
    # Non executa os tests (nin de Go nin do frontend).
    [switch]$SkipTests,
    # Non recompila o frontend: reutiliza o que xa haxa en internal\web\dist.
    [switch]$SkipFrontend,
    # Reinstala as dependencias do frontend con "npm ci" aínda que xa existan.
    [switch]$CleanInstall
)

$ErrorActionPreference = 'Stop'
Set-Location -LiteralPath $PSScriptRoot

function Invoke-Step([string]$Name, [scriptblock]$Action) {
    Write-Host "==> $Name" -ForegroundColor Cyan
    & $Action
    if ($LASTEXITCODE -ne 0) {
        throw "Fallou o paso '$Name' (código $LASTEXITCODE)."
    }
}

if (-not $SkipFrontend) {
    $portableNode = Join-Path $PSScriptRoot '.tools\node'
    if (Test-Path (Join-Path $portableNode 'node.exe')) {
        $env:PATH = "$portableNode;$env:PATH"
    }
    if (-not (Get-Command node -ErrorAction SilentlyContinue)) {
        throw 'Non se atopou Node.js. Instala Node 24 LTS ou copia un Node portátil en .tools\node.'
    }
    $nodeVersion = [version]((& node --version).TrimStart('v'))
    if ($nodeVersion -lt [version]'22.22.0') {
        throw "Node $nodeVersion é demasiado antigo: o frontend require Node 22.22 ou superior."
    }
    Write-Host "Node $nodeVersion" -ForegroundColor DarkGray

    Push-Location (Join-Path $PSScriptRoot 'frontend')
    try {
        if ($CleanInstall -or -not (Test-Path 'node_modules')) {
            Invoke-Step 'npm ci' { npm ci }
        }
        Invoke-Step 'npm run lint' { npm run lint }
        if (-not $SkipTests) {
            Invoke-Step 'npm test' { npm test }
        }
        Invoke-Step 'npm run build' { npm run build }
    }
    finally {
        Pop-Location
    }
}

Invoke-Step 'go vet' { go vet ./... }
if (-not $SkipTests) {
    Invoke-Step 'go test' { go test ./... }
}
Invoke-Step 'go build' { go build -o invest-tracker.exe . }

Write-Host 'Listo: invest-tracker.exe' -ForegroundColor Green
