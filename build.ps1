<#
.SYNOPSIS
    Compila invest-tracker: frontend (lint, tests, build) e Go (vet, tests, build).

.DESCRIPTION
    Deixa os binarios para Windows (invest-tracker.exe) e macOS
    (invest-tracker-macos-arm64 para Apple Silicon e invest-tracker-macos-amd64
    para Intel), todos coa interface web embebida. Usa o Node portátil de
    .tools\node se existe; se non, o do sistema (require Node 22.22 ou superior).
    En macOS ou Linux, usa build.sh.

.EXAMPLE
    powershell -NoProfile -ExecutionPolicy Bypass -File .\build.ps1

.EXAMPLE
    powershell -NoProfile -ExecutionPolicy Bypass -File .\build.ps1 -SkipTests

.EXAMPLE
    powershell -NoProfile -ExecutionPolicy Bypass -File .\build.ps1 -Installers
#>
[CmdletBinding()]
param(
    # Non executa os tests (nin de Go nin do frontend).
    [switch]$SkipTests,
    # Non recompila o frontend: reutiliza o que xa haxa en internal\web\dist.
    [switch]$SkipFrontend,
    # Reinstala as dependencias do frontend con "npm ci" aínda que xa existan.
    [switch]$CleanInstall,
    # Xera tamén os instaladores de Windows e macOS en dist-installers
    # (o de Windows require Inno Setup 6).
    [switch]$Installers
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

# Binarios que se distribúen no repositorio. -trimpath e -buildvcs=false fan
# que o mesmo código xere sempre o mesmo binario (e non levan rutas locais);
# -s -w quitan os símbolos de depuración para que pesen menos.
$targets = @(
    @{ GOOS = 'windows'; GOARCH = 'amd64'; Output = 'invest-tracker.exe' },
    @{ GOOS = 'darwin'; GOARCH = 'arm64'; Output = 'invest-tracker-macos-arm64' },
    @{ GOOS = 'darwin'; GOARCH = 'amd64'; Output = 'invest-tracker-macos-amd64' }
)
foreach ($target in $targets) {
    $env:GOOS = $target.GOOS
    $env:GOARCH = $target.GOARCH
    $env:CGO_ENABLED = '0'
    try {
        Invoke-Step "go build $($target.Output)" { go build -trimpath -buildvcs=false -ldflags '-s -w' -o $target.Output . }
    }
    finally {
        Remove-Item Env:GOOS, Env:GOARCH, Env:CGO_ENABLED -ErrorAction SilentlyContinue
    }
}

Write-Host "Listo: $(($targets | ForEach-Object { $_.Output }) -join ', ')" -ForegroundColor Green

if ($Installers) {
    Invoke-Step 'instalador de macOS' { go run ./installers/macos }

    $iscc = @(
        (Get-Command iscc -ErrorAction SilentlyContinue).Source,
        "$env:LOCALAPPDATA\Programs\Inno Setup 6\ISCC.exe",
        "${env:ProgramFiles(x86)}\Inno Setup 6\ISCC.exe",
        "$env:ProgramFiles\Inno Setup 6\ISCC.exe"
    ) | Where-Object { $_ -and (Test-Path $_) } | Select-Object -First 1
    if (-not $iscc) {
        throw 'Non se atopou Inno Setup 6 (ISCC.exe): instálao desde https://jrsoftware.org/isdl.php.'
    }
    $version = (Get-Content -LiteralPath (Join-Path $PSScriptRoot 'VERSION') -Raw).Trim()
    Invoke-Step 'instalador de Windows' { & $iscc /Qp "/DAppVersion=$version" installers\windows\ControlInvestimentos.iss }

    Write-Host 'Instaladores en dist-installers:' -ForegroundColor Green
    Get-ChildItem -LiteralPath (Join-Path $PSScriptRoot 'dist-installers') | ForEach-Object { Write-Host "  $($_.Name)" }
}
