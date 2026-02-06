; =====================================================================
; NEXA OS ULTIMATE v4.0.0-PRO | MASSIVE UNIFIED DEPLOYMENT 
; =====================================================================
; This is the heavyweight installer including ALL binaries and source.

#define AppName "Nexa OS Ultimate"
#define AppVersion "4.0.0-PRO"
#define AppPublisher "Nexa Intelligence Systems"
#define AppURL "http://hub.n"
#define AppExeName "nexa.exe"
#define AppGuid "5C1B4A5D-8E23-4B9F-AF12-0F98E8D93C22"

[Setup]
AppId={{{#AppGuid}}}
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
OutputBaseFilename=Nexa_OS_Ultimate_V4_Setup
; Using High compression but including massive files
Compression=lzma2/ultra64
SolidCompression=yes
PrivilegesRequired=admin
ArchitecturesAllowed=x64
ArchitecturesInstallIn64BitMode=x64
CloseApplications=yes
WizardStyle=modern
DisableWelcomePage=no
LicenseFile=LICENSE.txt
InfoBeforeFile=WELCOME.txt
UninstallDisplayIcon={app}\{#AppExeName}
VersionInfoVersion=4.0.0.0
VersionInfoCompany={#AppPublisher}
VersionInfoDescription="Nexa OS Unified Intelligent Environment"
VersionInfoProductName={#AppName}
VersionInfoTextVersion={#AppVersion}

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"
Name: "arabic"; MessagesFile: "compiler:Languages\Arabic.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked
Name: "startup"; Description: "تشغيل النظام تلقائياً عند بدء ويندوز (Launch on System Boot)"; GroupDescription: "إعدادات التشغيل:"
Name: "masterprep"; Description: "تفعيل وضع 'Nexa Master' (تجهيز المنافذ وجدار الحماية)"; GroupDescription: "تهيئة الشبكة الذكية:"; Flags: checkedonce
Name: "addpath"; Description: "إضافة Nexa إلى متغيرات النظام (Add to PATH Env Variable)"; GroupDescription: "إعدادات المطورين:"

[Files]
; --- 1. THE ULTIMATE BINARY SUITE (MODULAR & UNIFIED) ---
; Including the entire BIN folder with all individual service EXEs
Source: "..\bin\*"; DestDir: "{app}\bin"; Flags: ignoreversion recursesubdirs createallsubdirs
; The main unified executable at root
Source: "..\nexa.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\NEXA_MASTER_READY.bat"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\BUILD_INSTALLER.bat"; DestDir: "{app}"; Flags: ignoreversion
Source: "SETUP_ENVIRONMENT.bat"; DestDir: "{app}"; Flags: ignoreversion

; --- 2. SOURCE CODE & SYSTEMS (FOR FULL TRANSPARENCY) ---
Source: "..\pkg\*"; DestDir: "{app}\pkg"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\cmd\*"; DestDir: "{app}\cmd"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\scripts\*"; DestDir: "{app}\scripts"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\tools\*"; DestDir: "{app}\tools"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\docs\*"; DestDir: "{app}\docs"; Flags: ignoreversion recursesubdirs createallsubdirs

; --- 3. INFRASTRUCTURE & PERSISTENCE ---
Source: "..\data\*"; DestDir: "{app}\data"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\config\*"; DestDir: "{app}\config"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\sites\*"; DestDir: "{app}\sites"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\storage\*"; DestDir: "{app}\storage"; Flags: ignoreversion recursesubdirs createallsubdirs
Source: "..\certs\*"; DestDir: "{app}\certs"; Flags: ignoreversion recursesubdirs createallsubdirs

; --- 4. ROOT METADATA ---
Source: "..\go.mod"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\go.sum"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\readme.md"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\config.yaml"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\ledger.json"; DestDir: "{app}"; Flags: ignoreversion skipifsourcedoesntexist
Source: "..\dns_records.json"; DestDir: "{app}"; Flags: ignoreversion skipifsourcedoesntexist
Source: "..\users.json"; DestDir: "{app}"; Flags: ignoreversion skipifsourcedoesntexist
Source: "..\data\analytics.json"; DestDir: "{app}\data"; Flags: ignoreversion skipifsourcedoesntexist

[Icons]
Name: "{group}\{#AppName}"; Filename: "{app}\{#AppExeName}"
Name: "{group}\Unified Admin Console"; Filename: "{app}\NEXA_MASTER_READY.bat"
Name: "{group}\Matrix Analytics Hub"; Filename: "http://localhost:8000/analytics"
Name: "{group}\Individual Services\Gateway Control"; Filename: "{app}\bin\nexa_gateway.exe"
Name: "{group}\Individual Services\Admin Center"; Filename: "{app}\bin\nexa_admin.exe"
Name: "{group}\Individual Services\Intelligence Hub"; Filename: "{app}\bin\nexa_dashboard.exe"
Name: "{group}\Documentation"; Filename: "{app}\readme.md"
Name: "{group}\{cm:UninstallProgram,{#AppName}}"; Filename: "{uninstallexe}"
Name: "{commondesktop}\{#AppName}"; Filename: "{app}\{#AppExeName}"; Tasks: desktopicon

[Registry]
; Auto-Run Entry
Root: HKCU; Subkey: "Software\Microsoft\Windows\CurrentVersion\Run"; ValueType: string; ValueName: "{#AppName}"; ValueData: """{app}\{#AppExeName}"""; Flags: uninsdeletevalue; Tasks: startup
; PATH Integration
Root: HKLM; Subkey: "SYSTEM\CurrentControlSet\Control\Session Manager\Environment"; \
    ValueType: expandsz; ValueName: "Path"; ValueData: "{olddata};{app}"; \
    Check: NeedsAddPath; Tasks: addpath

[Run]
; Deploy Matrix Firewall & Port 80 Liberation via internal setup script
Filename: "{app}\SETUP_ENVIRONMENT.bat"; Description: "تثبيت قواعد الشبكة وتحرير المنفذ 80 (المستوى الاحترافي)"; Flags: postinstall runascurrentuser; Tasks: masterprep

; Launch Unified Engine
Filename: "{app}\{#AppExeName}"; Description: "بدء نظام نكسا الموحد (Start Nexa OS)"; Flags: nowait postinstall skipifsilent

[UninstallRun]
; Detailed Cleanup
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall delete rule name=""NEXA BINARY"""; Flags: runhidden
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall delete rule name=""NEXA WEB"""; Flags: runhidden
Filename: "{sys}\netsh.exe"; Parameters: "advfirewall firewall delete rule name=""NEXA DNS"""; Flags: runhidden

[Code]
function NeedsAddPath: Boolean;
var
  Path: string;
begin
  if RegQueryStringValue(HKLM, 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment', 'Path', Path) then
    Result := Pos(';' + ExpandConstant('{app}'), Path) = 0
  else
    Result := True;
end;

function GetUninstallString: string;
var
  sUnInstPath: string;
  sUnInstallString: String;
begin
  sUnInstPath := 'Software\Microsoft\Windows\CurrentVersion\Uninstall\{' + '{#AppGuid}' + '}_is1';
  sUnInstallString := '';
  if not RegQueryStringValue(HKLM, sUnInstPath, 'UninstallString', sUnInstallString) then
    RegQueryStringValue(HKCU, sUnInstPath, 'UninstallString', sUnInstallString);
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
    V := MsgBox('لقد تم اكتشاف وجود نسخة سابقة من Nexa OS. هل تريد إزالتها تلقائياً قبل متابعة التثبيت الجديد؟' + #13#10#13#10 + 'Previous version detected. Uninstall automatically?', mbConfirmation, MB_YESNO);
    if V = IDYES then
    begin
      sUnInstallString := GetUninstallString;
      sUnInstallString := RemoveQuotes(sUnInstallString);
      Exec(sUnInstallString, '/SILENT /NORESTART /SUPPRESSMSGBOXES', '', SW_SHOW, ewWaitUntilTerminated, iResultCode);
      Result := True;
    end
    else
      Result := True; 
  end;
end;

