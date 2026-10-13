# Control de Investimentos

Ferramenta para o seguimento mensual dunha carteira de investimentos en USD.
Funciona en Windows e en macOS (Apple Silicon e Intel). Tes dúas formas de
usala, que comparten a mesma base de datos (`investimentos.db`):

- **Interface web**: no navegador do propio ordenador ou desde outros dispositivos
  da rede local, como o móbil. Non ten contrasinal.
- **Terminal (CLI)**: os menús de texto de sempre.

> `investimentos.db` non se sube a GitHub (está en `.gitignore`): garda ti unha
> copia de seguridade dela. Se non existe, a aplicación créaa baleira ao
> arrancar, no cartafol desde o que se executa. Os lanzadores
> (`invest-tracker-web.cmd` en Windows, os `.command` en macOS) usan sempre o
> cartafol do programa, e ao arrancar móstrase a ruta da base de datos.

## Idioma / Language

A aplicación está en galego e en inglés (*available in Galician and English*):

- **Web**: por defecto, en inglés se o navegador está en inglés e en galego en
  calquera outro caso. Cámbiase co selector da cabeceira (`GL`/`EN`) ou en
  *Axustes*, e cada navegador lémbrao.
- **Terminal**: por defecto, o idioma do sistema (inglés se o sistema está en
  inglés; se non, galego). Cámbiase coa opción *Idioma / Language* do menú
  principal e gárdase en `idioma.txt`, no cartafol de configuración
  (`%APPDATA%\Control de Investimentos` en Windows,
  `~/Library/Application Support/Control de Investimentos` en macOS).
- **Instaladores**: o de Windows está nos dous idiomas e o que se escolla na
  instalación pasa a ser o da terminal; a app de macOS leva `LEME.txt` e
  `README.txt`.

## Instalar

Descarga o instalador da túa plataforma desde
[Releases](https://github.com/NewHouses/invest-tracker/releases), ou xérao ti
(ver [Crear os instaladores](#crear-os-instaladores)). Non están asinados, así
que a primeira vez o sistema avisa.

### Windows

1. Executa `ControlInvestimentos-<versión>-windows-setup.exe`. Non pide permisos
   de administrador: instálase só para o teu usuario.
2. Se Windows SmartScreen avisa, preme *Máis información* → *Executar de todos
   modos*.
3. Abre *Control de Investimentos* desde o menú Inicio ou o escritorio. No menú
   Inicio tamén están a versión de terminal e o *Cartafol dos datos*.

Os datos gárdanse en `%APPDATA%\Control de Investimentos\investimentos.db`. Ao
desinstalar (Configuración → Aplicacións) bórrase o programa, pero non os datos.

### macOS

1. Descomprime `ControlInvestimentos-<versión>-macos.zip` e arrastra
   *Control de Investimentos.app* a Aplicacións.
2. A primeira vez macOS non a deixa abrir: vai a *Axustes do Sistema →
   Privacidade e seguranza* e preme *Abrir igualmente* (ou executa
   `xattr -dr com.apple.quarantine "/Applications/Control de Investimentos.app"`).
3. Ao abrila aparece unha xanela do Terminal co servidor e ábrese o navegador.
   Para pechala, preme `Ctrl+C` no Terminal.

Os datos gárdanse en `~/Library/Application Support/Control de Investimentos/investimentos.db`.

> Para levar os datos que xa tes, copia o teu `investimentos.db` a ese cartafol
> antes de abrir a aplicación instalada.

## Requisitos para compilar

- [Go](https://go.dev/dl/) 1.25 ou superior.
- Node.js 22.22 ou superior, só para compilar a interface web. Se o Node do
  sistema é máis antigo, copia un Node 24 LTS portátil en `.tools/node` (en
  macOS, o contido do `.tar.gz`, co cartafol `bin`): os scripts de compilación
  úsano automaticamente.

## Compilar

En Windows:

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\build.ps1
```

En macOS (ou Linux):

```bash
./build.sh
```

Os scripts fan isto, en orde:
1. Instalan as dependencias do frontend se fai falta.
2. Pasan o lint e os tests do frontend e compílano.
3. Pasan `go vet` e os tests de Go.
4. Xeran os binarios, coa web xa embebida: `invest-tracker.exe` (Windows),
   `invest-tracker-macos-arm64` (Mac con Apple Silicon) e
   `invest-tracker-macos-amd64` (Mac con Intel).

Opcións (`build.ps1` / `build.sh`):
- `-SkipTests` / `--skip-tests`: non executa os tests.
- `-SkipFrontend` / `--skip-frontend`: reutiliza a web xa compilada.
- `-CleanInstall` / `--clean-install`: reinstala as dependencias de npm.
- `-Installers` / `--installers`: xera tamén os instaladores.

## Crear os instaladores

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\build.ps1 -Installers
```

Deixa en `dist-installers/` o instalador de Windows (require
[Inno Setup 6](https://jrsoftware.org/isdl.php)) e a app de macOS en `.zip`. En
macOS, `./build.sh --installers` xera só a app de macOS. A versión lese do
ficheiro `VERSION`. As iconas (`installers/windows/app.ico` e
`installers/macos/AppIcon.icns`) xéranse co deseño do favicon con
`go run ./installers/icon`.

## Usar desde o repositorio

### Interface web

- **Windows**: dobre clic en `invest-tracker-web.cmd`, ou `.\invest-tracker.exe web`.
- **macOS**: dobre clic en `invest-tracker-web.command`, ou
  `./invest-tracker-macos-arm64 web` (`-amd64` nun Mac con Intel).

Ábrese o navegador en `http://localhost:8080`. A consola mostra tamén os
enderezos da rede local, por exemplo `http://192.168.1.20:8080`.

Opcións:
- `-addr 0.0.0.0:8080`: enderezo e porto.
- `-db ruta.db`: outra base de datos.
- `-no-open`: non abrir o navegador.

Para parar o servidor, preme `Ctrl+C` na consola.

### Terminal

- **Windows**: `.\invest-tracker.exe`
- **macOS**: dobre clic en `invest-tracker.command`, ou `./invest-tracker-macos-arm64`
  (`-amd64` nun Mac con Intel).

Escribe `:q` ou `cancelar` en calquera pregunta para volver ao menú (en inglés,
`:q` ou `cancel`).

### Primeira vez en macOS

- Se descargaches o proxecto como ZIP en vez de con `git clone`, macOS bloquea
  os programas descargados («non se pode verificar o programador»). Abre o
  Terminal no cartafol do proxecto e executa `xattr -dr com.apple.quarantine .`
  (ou fai clic dereito no `.command` → *Abrir*).
- Ao arrancar a web, macOS pode preguntar se permites conexións entrantes:
  acéptao se a queres usar desde o móbil.

## Seguridade

- A web non ten contrasinal: mentres o servidor está aberto, calquera dispositivo
  da túa rede local pode ver e cambiar os datos. Úsaa só nunha rede doméstica de
  confianza; a conexión é HTTP simple.
- Só se atenden peticións dirixidas á IP do equipo, a `localhost` ou ao nome do
  equipo. Así unha páxina web maliciosa non pode chegar aos datos a través do
  navegador.

## Acceso desde o móbil

1. Arranca o modo web no ordenador.
2. A primeira vez, o firewall pregunta se permite o acceso. En Windows,
   acéptao **só para redes privadas**; en macOS, permite as conexións entrantes.
3. No móbil, conectado á mesma wifi, abre o enderezo "Rede local" que mostra a
   consola.

## Desenvolvemento

Usa dúas terminais (os comandos son iguais en Windows e en macOS):

```powershell
# 1) API en Go (porto 8080)
go run . web -no-open

# 2) Frontend con recarga en quente (redirixe /api ao porto 8080)
cd frontend
npm run dev
```

## Estrutura

- `main.go`: menú da CLI. Co argumento `web` arranca o servidor.
- `internal/domain`: modelos, validacións e proxección.
- `internal/store`: acceso a SQLite.
- `internal/<operación>`: unha operación da CLI cada paquete. Exportan funcións
  `Build…` cos cálculos, que comparten a CLI e a web.
- `internal/web`: servidor HTTP, API JSON (`/api`) e web embebida (`dist/`).
- `internal/i18n`: traducións da CLI e da API. Os textos escríbense en galego e
  úsanse como clave; o inglés de cada paquete está no seu `i18n_en.go`. A API
  responde no idioma da cabeceira `Accept-Language`.
- `frontend/`: React + TypeScript + Mantine (Vite). Compílase en `internal/web/dist`.
  As súas traducións ao inglés están en `frontend/src/i18n/en/`.
- `installers/`: instalador de Windows (Inno Setup), empaquetado da app de macOS
  e xerador das iconas.

## Tests

```powershell
go test ./...
cd frontend; npm test
```

Hai tests que comproban que todos os textos (Go e web) teñen a súa tradución ao
inglés cos mesmos parámetros.
