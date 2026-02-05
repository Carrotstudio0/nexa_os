; Nexa Installer Script for Inno Setup
; Built by Antigravity AI

[Setup]
AppId={{5C1B4A5D-8E23-4B9F-AF12-0F98E8D93C22}
AppName=Nexa
AppVersion=3.1
AppPublisher=MultiX0
DefaultDirName={autopf}\Nexa
DefaultGroupName=Nexa
AllowNoIcons=yes
; The following line specifies the output directory for the installer
OutputDir=..\installer_output
OutputBaseFilename=Nexa_Setup_v3.1
Compression=lzma
SolidCompression=yes
PrivilegesRequired=admin
CloseApplications=yes

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked
Name: "startup"; Description: "Automatically start Nexa when Windows starts"; GroupDescription: "Additional options:"

[Files]
; Root binaries and scripts
Source: "..\bin\nexa.exe"; DestDir: "{app}\bin"; Flags: ignoreversion
Source: "..\bin\server.exe"; DestDir: "{app}\bin"; Flags: ignoreversion
Source: "..\bin\gateway.exe"; DestDir: "{app}\bin"; Flags: ignoreversion
Source: "..\bin\admin.exe"; DestDir: "{app}\bin"; Flags: ignoreversion
Source: "..\bin\web.exe"; DestDir: "{app}\bin"; Flags: ignoreversion
Source: "..\bin\chat.exe"; DestDir: "{app}\bin"; Flags: ignoreversion
Source: "..\bin\dns.exe"; DestDir: "{app}\bin"; Flags: ignoreversion
Source: "..\bin\dashboard.exe"; DestDir: "{app}\bin"; Flags: ignoreversion

; Root Files
Source: "..\NEXA.bat"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\go.mod"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\go.sum"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\readme.md"; DestDir: "{app}"; Flags: ignoreversion

; Folders (Mirroring the project structure)
Source: "..\bin\start-all.bat"; DestDir: "{app}\bin"; Flags: ignoreversion
Source: "..\config\*"; DestDir: "{app}\config"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\certs\*"; DestDir: "{app}\certs"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\data\*"; DestDir: "{app}\data"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\scripts\*"; DestDir: "{app}\scripts"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\docs\*"; DestDir: "{app}\docs"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\tools\*"; DestDir: "{app}\tools"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\cmd\*"; DestDir: "{app}\cmd"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\pkg\*"; DestDir: "{app}\pkg"; Flags: ignoreversion recursesubdirs createallsubdirs


[Icons]
Name: "{group}\Nexa Launcher"; Filename: "{app}\NEXA.bat"
Name: "{group}\Nexa Dashboard"; Filename: "http://localhost:7000"
Name: "{commondesktop}\Nexa Launcher"; Filename: "{app}\NEXA.bat"; Parameters: "/run"; Tasks: desktopicon
Name: "{commondesktop}\Nexa Dashboard"; Filename: "http://localhost:7000"; Tasks: desktopicon

[Registry]
; Auto-run
Root: HKCU; Subkey: "Software\Microsoft\Windows\CurrentVersion\Run"; ValueType: string; ValueName: "Nexa"; ValueData: """{app}\NEXA.bat"" /run"; Flags: uninsdeletevalue; Tasks: startup

[Run]
; Launch the app after install (Optional)
Filename: "{app}\NEXA.bat"; Parameters: "/run"; Description: "{cm:LaunchProgram,Nexa}"; Flags: nowait postinstall skipifsilent

; Open Ports in Windows Firewall
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall add rule name=""Nexa Hub"" dir=in action=allow protocol=TCP localport=7000"; Flags: runhidden; StatusMsg: "Configuring Firewall (Hub)..."
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall add rule name=""Nexa Gateway"" dir=in action=allow protocol=TCP localport=8000"; Flags: runhidden; StatusMsg: "Configuring Firewall (Gateway)..."
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall add rule name=""Nexa Admin"" dir=in action=allow protocol=TCP localport=8080"; Flags: runhidden; StatusMsg: "Configuring Firewall (Admin)..."
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall add rule name=""Nexa Storage"" dir=in action=allow protocol=TCP localport=8081"; Flags: runhidden; StatusMsg: "Configuring Firewall (Storage)..."
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall add rule name=""Nexa Chat"" dir=in action=allow protocol=TCP localport=8082"; Flags: runhidden; StatusMsg: "Configuring Firewall (Chat)..."
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall add rule name=""Nexa Server"" dir=in action=allow protocol=TCP localport=1413"; Flags: runhidden; StatusMsg: "Configuring Firewall (Server)..."
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall add rule name=""Nexa DNS"" dir=in action=allow protocol=UDP localport=1112"; Flags: runhidden; StatusMsg: "Configuring Firewall (DNS)..."

[UninstallRun]
; Clean up Firewall rules on uninstall
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall delete rule name=""Nexa Hub"""; Flags: runhidden
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall delete rule name=""Nexa Gateway"""; Flags: runhidden
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall delete rule name=""Nexa Admin"""; Flags: runhidden
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall delete rule name=""Nexa Storage"""; Flags: runhidden
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall delete rule name=""Nexa Chat"""; Flags: runhidden
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall delete rule name=""Nexa Server"""; Flags: runhidden
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall delete rule name=""Nexa DNS"""; Flags: runhidden

[Messages]
FinishedHeadingLabel=Setup Complete
FinishedLabelNoIcons=Nexa has been installed on your computer. The system is ready to launch.

[Code]
function GetUninstallString: string;
var
  sUninstPath: string;
  sUnInstallString: String;
begin
  sUninstPath := ExpandConstant('Software\Microsoft\Windows\CurrentVersion\Uninstall\{{5C1B4A5D-8E23-4B9F-AF12-0F98E8D93C22}_is1');
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
  V: Integer;
  iResultCode: Integer;
  sUnInstallString: string;
begin
  Result := True;
  if IsUpgrade then
  begin
    V := MsgBox('An existing version of Nexa was detected. Do you want to uninstall it before continuing?', mbInformation, MB_YESNO);
    if V = IDYES then
    begin
      sUnInstallString := RemoveQuotes(GetUninstallString);
      Exec(sUnInstallString, '/SILENT /NORESTART /SUPPRESSMSGBOXES', '', SW_SHOW, ewWaitUntilTerminated, iResultCode);
      Result := True;
    end
    else
      Result := True;
  end;
end;

