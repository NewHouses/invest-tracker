; Instalador de Windows de Control de Investimentos (Inno Setup 6).
;
; Xérase con build.ps1 -Installers, que lle pasa a versión do ficheiro VERSION.
; Instala só para o usuario actual (sen permisos de administrador). Os datos
; (investimentos.db) gárdanse en %APPDATA%\Control de Investimentos e non se
; borran ao desinstalar.
;
; Está en galego e en inglés: escolle o idioma do sistema e, se non é ningún
; dos dous, pregunta (por defecto galego). O idioma escollido gárdase tamén
; como idioma inicial da aplicación (idioma.txt), se aínda non tiña un.

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
ShowLanguageDialog=auto

; O primeiro idioma é o que se usa cando o do sistema non é ningún destes.
[Languages]
Name: "gl"; MessagesFile: "Galician.isl"
Name: "en"; MessagesFile: "compiler:Default.isl"

[CustomMessages]
gl.WebComment=Abre a interface web no navegador
en.WebComment=Opens the web interface in the browser
gl.TerminalComment=Abre os menús de texto
en.TerminalComment=Opens the text menus
gl.DataFolder=Cartafol dos datos
en.DataFolder=Data folder
gl.DataFolderComment=Onde está investimentos.db (para copias de seguridade)
en.DataFolderComment=Where investimentos.db is (for backups)

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
Name: "{group}\{#AppName}"; Filename: "{app}\{#AppExe}"; Parameters: "web"; WorkingDir: "{#DataDir}"; IconFilename: "{app}\app.ico"; Comment: "{cm:WebComment}"
Name: "{group}\{#AppName} (terminal)"; Filename: "{app}\{#AppExe}"; WorkingDir: "{#DataDir}"; IconFilename: "{app}\app.ico"; Comment: "{cm:TerminalComment}"
Name: "{group}\{cm:DataFolder}"; Filename: "{#DataDir}"; Comment: "{cm:DataFolderComment}"
Name: "{group}\{cm:UninstallProgram,{#AppName}}"; Filename: "{uninstallexe}"
Name: "{autodesktop}\{#AppName}"; Filename: "{app}\{#AppExe}"; Parameters: "web"; WorkingDir: "{#DataDir}"; IconFilename: "{app}\app.ico"; Tasks: desktopicon

[Run]
Filename: "{app}\{#AppExe}"; Parameters: "web"; WorkingDir: "{#DataDir}"; Description: "{cm:LaunchProgram,{#AppName}}"; Flags: nowait postinstall skipifsilent

; Ao reinstalar noutro idioma, bórranse os accesos directos co nome do idioma
; anterior para que non queden repetidos (os do idioma actual vólvense crear).
[InstallDelete]
Type: files; Name: "{group}\Cartafol dos datos.lnk"
Type: files; Name: "{group}\Data folder.lnk"
Type: files; Name: "{group}\Desinstalar {#AppName}.lnk"
Type: files; Name: "{group}\Uninstall {#AppName}.lnk"

[Code]
// Garda o idioma da instalación como idioma inicial da aplicación (a CLI leo
// de idioma.txt), sen cambiar o que o usuario xa escollese antes.
procedure CurStepChanged(CurStep: TSetupStep);
var
  LangFile: String;
begin
  if CurStep = ssPostInstall then
  begin
    LangFile := ExpandConstant('{#DataDir}\idioma.txt');
    if not FileExists(LangFile) then
      SaveStringToFile(LangFile, ActiveLanguage + #13#10, False);
  end;
end;