; Instalador de Windows de Control de Investimentos (Inno Setup 6).
;
; Xérase con build.ps1 -Installers, que lle pasa a versión do ficheiro VERSION.
; Instala só para o usuario actual (sen permisos de administrador). Os datos
; (investimentos.db) gárdanse en %APPDATA%\Control de Investimentos e non se
; borran ao desinstalar.

#ifndef AppVersion
  #define AppVersion "0.0.0"
#endif
#define AppName "Control de Investimentos"
#define AppExe "invest-tracker.exe"
#define DataDir "{userappdata}\Control de Investimentos"

[Setup]
; O AppId identifica a aplicación nas actualizacións: non o cambies.
AppId={{4E0B539D-45E2-4C8F-B413-081EFEF8DB65}
AppName={#AppName}
AppVersion={#AppVersion}
AppVerName={#AppName} {#AppVersion}
AppPublisher=NewHouses
AppPublisherURL=https://github.com/NewHouses/invest-tracker
AppSupportURL=https://github.com/NewHouses/invest-tracker
PrivilegesRequired=lowest
DefaultDirName={autopf}\{#AppName}
DefaultGroupName={#AppName}
DisableProgramGroupPage=yes
OutputDir=..\..\dist-installers
OutputBaseFilename=ControlInvestimentos-{#AppVersion}-windows-setup
SetupIconFile=app.ico
UninstallDisplayIcon={app}\app.ico
UninstallDisplayName={#AppName}
ArchitecturesAllowed=x64compatible
ArchitecturesInstallIn64BitMode=x64compatible
WizardStyle=modern
Compression=lzma2/max
SolidCompression=yes
CloseApplications=yes

[Languages]
Name: "gl"; MessagesFile: "Galician.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"

[Dirs]
Name: "{#DataDir}"; Flags: uninsneveruninstall

[Files]
Source: "..\..\invest-tracker.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "app.ico"; DestDir: "{app}"; Flags: ignoreversion

; Os accesos directos arrancan no cartafol dos datos: a aplicación usa o
; investimentos.db do cartafol desde o que se executa.
[Icons]
Name: "{group}\{#AppName}"; Filename: "{app}\{#AppExe}"; Parameters: "web"; WorkingDir: "{#DataDir}"; IconFilename: "{app}\app.ico"; Comment: "Abre a interface web no navegador"
Name: "{group}\{#AppName} (terminal)"; Filename: "{app}\{#AppExe}"; WorkingDir: "{#DataDir}"; IconFilename: "{app}\app.ico"; Comment: "Abre os menús de texto"
Name: "{group}\Cartafol dos datos"; Filename: "{#DataDir}"; Comment: "Onde está investimentos.db (para copias de seguridade)"
Name: "{group}\{cm:UninstallProgram,{#AppName}}"; Filename: "{uninstallexe}"
Name: "{autodesktop}\{#AppName}"; Filename: "{app}\{#AppExe}"; Parameters: "web"; WorkingDir: "{#DataDir}"; IconFilename: "{app}\app.ico"; Tasks: desktopicon

[Run]
Filename: "{app}\{#AppExe}"; Parameters: "web"; WorkingDir: "{#DataDir}"; Description: "{cm:LaunchProgram,{#AppName}}"; Flags: nowait postinstall skipifsilent