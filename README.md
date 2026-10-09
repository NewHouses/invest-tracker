# Control de Investimentos

Ferramenta para o seguimento mensual dunha carteira de investimentos en USD.
Tes dúas formas de usala, que comparten a mesma base de datos (`investimentos.db`):

- **Interface web**: no navegador do propio ordenador ou desde outros dispositivos
  da rede local, como o móbil. Está protexida con contrasinal.
- **Terminal (CLI)**: os menús de texto de sempre.

> `investimentos.db` non se sube a GitHub (está en `.gitignore`): garda ti unha
> copia de seguridade dela.

## Requisitos para compilar

- [Go](https://go.dev/dl/) 1.25 ou superior.
- Node.js 22.22 ou superior, só para compilar a interface web. Se o Node do
  sistema é máis antigo, copia un Node 24 LTS portátil en `.tools\node`: o script
  de compilación úsao automaticamente.

## Compilar

```powershell
powershell -NoProfile -ExecutionPolicy Bypass -File .\build.ps1
```

O script fai isto, en orde:
1. Instala as dependencias do frontend se fai falta.
2. Pasa o lint e os tests do frontend e compílao.
3. Pasa `go vet` e os tests de Go.
4. Xera `invest-tracker.exe`, coa web xa embebida.

Opcións:
- `-SkipTests`: non executa os tests.
- `-SkipFrontend`: reutiliza a web xa compilada.
- `-CleanInstall`: reinstala as dependencias de npm.

## Usar

### Interface web

Fai dobre clic en `invest-tracker-web.cmd`, ou executa:

```powershell
.\invest-tracker.exe web
```

Ábrese o navegador en `http://localhost:8080`. A consola mostra tamén os
enderezos da rede local, por exemplo `http://192.168.1.20:8080`.

Opcións:
- `-addr 0.0.0.0:8080`: enderezo e porto.
- `-db ruta.db`: outra base de datos.
- `-no-open`: non abrir o navegador.

Para parar o servidor, preme `Ctrl+C` na consola.

### Terminal

```powershell
.\invest-tracker.exe
```

Escribe `:q` ou `cancelar` en calquera pregunta para volver ao menú.

## Contrasinal

- **Primeiro arranque:** o contrasinal créase no navegador, e por seguridade só
  desde o propio ordenador (`http://localhost:8080`).
- **Cambialo:** en *Axustes*. Ao cambialo péchanse as sesións dos demais dispositivos.
- **Esquecino:** executa `.\invest-tracker.exe web -reset-password` e volve
  crealo desde o ordenador.
- A conexión é HTTP simple: úsaa só nunha rede doméstica de confianza.

## Acceso desde o móbil

1. Arranca o modo web no ordenador.
2. A primeira vez, o Firewall de Windows pregunta se permite o acceso:
   acéptao **só para redes privadas**.
3. No móbil, conectado á mesma wifi, abre o enderezo "Rede local" que mostra a
   consola e inicia sesión.

## Desenvolvemento

Usa dúas terminais:

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
- `internal/web`: servidor HTTP, API JSON (`/api`), autenticación e web embebida
  (`dist/`).
- `frontend/`: React + TypeScript + Mantine (Vite). Compílase en `internal/web/dist`.

## Tests

```powershell
go test ./...
cd frontend; npm test
```
