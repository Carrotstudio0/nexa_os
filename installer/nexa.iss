; Nexa Ultimate Installer Script
; Built by Antigravity (Advanced Agentic AI)

#define AppName "Nexa Ultimate"
#define AppVersion "4.0.0-PRO"
#define AppPublisher "Nexa Intelligence Systems"
#define AppURL "http://nexa.matrix"
#define AppExeName "nexa.exe"
#define AppGuid "{5C1B4A5D-8E23-4B9F-AF12-0F98E8D93C22}"

[Setup]
AppId={{#AppGuid}
AppName={#AppName}
AppVersion={#AppVersion}
AppPublisher={#AppPublisher}
AppPublisherURL={#AppURL}
AppSupportURL={#AppURL}
AppUpdatesURL={#AppURL}
DefaultDirName={autopf}\{#AppName}
DefaultGroupName={#AppName}
AllowNoIcons=yes
OutputDir=..\installer_output
OutputBaseFilename=Nexa_Ultimate_v4.0.0_Setup
Compression=lzma2/ultra64
SolidCompression=yes
PrivilegesRequired=admin
CloseApplications=yes
; Professional Setup Aesthetics
WizardStyle=modern
DisableWelcomePage=no

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked
Name: "startup"; Description: "Launch Nexa Ultimate on Windows Startup"; GroupDescription: "System Integration:"

[Files]
; --- CORE BINARIES (THE MATRIX SUITE) ---
Source: "..\bin\nexa.exe"; DestDir: "{app}\bin"; Flags: ignoreversion
Source: "..\bin\gateway.exe"; DestDir: "{app}\bin"; Flags: ignoreversion
Source: "..\bin\admin.exe"; DestDir: "{app}\bin"; Flags: ignoreversion
Source: "..\bin\dashboard.exe"; DestDir: "{app}\bin"; Flags: ignoreversion
Source: "..\bin\chat.exe"; DestDir: "{app}\bin"; Flags: ignoreversion
Source: "..\bin\web.exe"; DestDir: "{app}\bin"; Flags: ignoreversion
Source: "..\bin\server.exe"; DestDir: "{app}\bin"; Flags: ignoreversion
Source: "..\bin\dns.exe"; DestDir: "{app}\bin"; Flags: ignoreversion
Source: "..\bin\client.exe"; DestDir: "{app}\bin"; Flags: ignoreversion
Source: "..\bin\gen_certs.exe"; DestDir: "{app}\bin"; Flags: ignoreversion
Source: "..\bin\hashgen.exe"; DestDir: "{app}\bin"; Flags: ignoreversion

; --- SYSTEM SOURCE & INFRASTRUCTURE ---
Source: "..\cmd\*"; DestDir: "{app}\cmd"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\pkg\*"; DestDir: "{app}\pkg"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\config\*"; DestDir: "{app}\config"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\certs\*"; DestDir: "{app}\certs"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\data\*"; DestDir: "{app}\data"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\scripts\*"; DestDir: "{app}\scripts"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\tools\*"; DestDir: "{app}\tools"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\docs\*"; DestDir: "{app}\docs"; Flags: ignoreversion recursesubdirs createallsubdirs

; --- ROOT FILES ---
Source: "..\go.mod"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\go.sum"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\readme.md"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\ledger.json"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\dns_records.json"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\users.json"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
Name: "{group}\{#AppName} Launcher"; Filename: "{app}\bin\{#AppExeName}"
Name: "{group}\{#AppName} Intelligence Hub"; Filename: "http://localhost:7000"
Name: "{commondesktop}\{#AppName}"; Filename: "{app}\bin\{#AppExeName}"; Tasks: desktopicon
Name: "{commondesktop}\Intelligence Hub"; Filename: "http://localhost:7000"; Tasks: desktopicon

[Registry]
Root: HKCU; Subkey: "Software\Microsoft\Windows\CurrentVersion\Run"; ValueType: string; ValueName: "{#AppName}"; ValueData: """{app}\bin\{#AppExeName}"""; Flags: uninsdeletevalue; Tasks: startup

[Run]
; Firewall Rules Deployment
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall add rule name=""Nexa Hub"" dir=in action=allow protocol=TCP localport=7000"; Flags: runhidden; StatusMsg: "Authorizing Intelligence Hub (7000)..."
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall add rule name=""Nexa Gateway"" dir=in action=allow protocol=TCP localport=8000"; Flags: runhidden; StatusMsg: "Authorizing Matrix Gateway (8000)..."
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall add rule name=""Nexa Services"" dir=in action=allow protocol=TCP localport=8080-8082"; Flags: runhidden; StatusMsg: "Authorizing Core Services (8080-8082)..."
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall add rule name=""Nexa Server"" dir=in action=allow protocol=TCP localport=1413"; Flags: runhidden; StatusMsg: "Authorizing Nucleus Server (1413)..."
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall add rule name=""Nexa DNS"" dir=in action=allow protocol=TCP localport=53"; Flags: runhidden; StatusMsg: "Authorizing DNS Authority (53)..."

; Finalization
Filename: "{app}\bin\{#AppExeName}"; Description: "{cm:LaunchProgram,{#StringChange(AppName, '&', '&&')}}"; Flags: nowait postinstall skipifsilent

[UninstallRun]
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall delete rule name=""Nexa Hub"""; Flags: runhidden
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall delete rule name=""Nexa Gateway"""; Flags: runhidden
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall delete rule name=""Nexa Services"""; Flags: runhidden
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall delete rule name=""Nexa Server"""; Flags: runhidden
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall delete rule name=""Nexa DNS"""; Flags: runhidden

[Messages]
FinishedHeadingLabel=Intelligence Layer Installed
FinishedLabelNoIcons=The Nexa Ultimate intelligence layer has been successfully deployed. You can now access the Matrix via the Intelligence Hub.

[Code]
function GetUninstallString: string;
var
  sUninstPath: string;
  sUnInstallString: String;
begin
  sUninstPath := 'Software\Microsoft\Windows\CurrentVersion\Uninstall\{' + '{#AppGuid}_is1';
  sUnInstallString := '';
  if not RegQueryStringValue(HKLM, sUninstPath, 'UninstallString', sUnInstallString) then
    RegQueryStringValue(HKCU, sUninstPath, 'UninstallString', sUnInstallString);
  Result := sUnInstallString;
end;

function IsUpgrade: Boolean;
begin
  Result := (GetUninstallString <> '');
end;

function InitializeSetup: Boolean;
var
  iResultCode: Integer;
  sUnInstallString: string;
begin
  Result := True;
  if IsUpgrade then
  begin
    sUnInstallString := RemoveQuotes(GetUninstallString);
    if Exec(sUnInstallString, '/SILENT /NORESTART /SUPPRESSMSGBOXES', '', SW_SHOW, ewWaitUntilTerminated, iResultCode) then
    begin
      Result := True;
    end;
  end;
end;

