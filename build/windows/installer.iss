; GoSer Installer Script for Inno Setup
; Download Inno Setup: https://jrsoftware.org/isinfo.php

#define MyAppName "GoSer"
#define MyAppVersion "1.0.0"
#define MyAppPublisher "GoSer"
#define MyAppURL "https://github.com/BAIGUANGMEI/goser"
#define MyAppExeName "goser-app.exe"

[Setup]
AppId={{B8F3A2E1-7C4D-4E5F-9A1B-2D3E4F5A6B7C}
AppName={#MyAppName}
AppVersion={#MyAppVersion}
AppPublisher={#MyAppPublisher}
AppPublisherURL={#MyAppURL}
DefaultDirName={autopf}\{#MyAppName}
DefaultGroupName={#MyAppName}
; Require admin for service installation and PATH modification
PrivilegesRequired=admin
OutputDir=..\..\dist\installer
OutputBaseFilename=GoSer-Setup-{#MyAppVersion}
Compression=lzma2/ultra64
SolidCompression=yes
WizardStyle=modern
SetupIconFile=logo.ico
UninstallDisplayIcon={app}\goser-app.exe
ArchitecturesInstallIn64BitMode=x64compatible
MinVersion=10.0

[Languages]
Name: "english"; MessagesFile: "compiler:Default.isl"
Name: "chinesesimplified"; MessagesFile: "compiler:Languages\ChineseSimplified.isl"

[Tasks]
Name: "desktopicon"; Description: "{cm:CreateDesktopIcon}"; GroupDescription: "{cm:AdditionalIcons}"; Flags: unchecked
Name: "addtopath"; Description: "Add GoSer CLI to system PATH"; GroupDescription: "System Integration:"; Flags: checkedonce
Name: "installservice"; Description: "Install daemon as Windows service (auto-start on boot)"; GroupDescription: "System Integration:"; Flags: unchecked

[Files]
; Main binaries from dist/ folder
Source: "..\..\dist\goserd.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\..\dist\goser.exe"; DestDir: "{app}"; Flags: ignoreversion
Source: "..\..\dist\goser-app.exe"; DestDir: "{app}"; Flags: ignoreversion

[Icons]
; Start Menu
Name: "{group}\GoSer Service Manager"; Filename: "{app}\{#MyAppExeName}"
Name: "{group}\Uninstall GoSer"; Filename: "{uninstallexe}"
; Desktop
Name: "{autodesktop}\GoSer Service Manager"; Filename: "{app}\{#MyAppExeName}"; Tasks: desktopicon

[Run]
; Install Windows service if selected
Filename: "{app}\goserd.exe"; Parameters: "-install"; StatusMsg: "Installing GoSer daemon service..."; Tasks: installservice; Flags: runhidden
; Option to launch after install
Filename: "{app}\{#MyAppExeName}"; Description: "Launch GoSer Service Manager"; Flags: nowait postinstall skipifsilent

[UninstallRun]
; Stop and uninstall Windows service on uninstall
Filename: "sc.exe"; Parameters: "stop GoSerDaemon"; Flags: runhidden; RunOnceId: "StopService"
Filename: "{app}\goserd.exe"; Parameters: "-uninstall"; Flags: runhidden; RunOnceId: "UninstallService"

[Code]
// Add to PATH
const
  EnvironmentKey = 'SYSTEM\CurrentControlSet\Control\Session Manager\Environment';

procedure AddToPath(Dir: String);
var
  OldPath: String;
begin
  if not RegQueryStringValue(HKEY_LOCAL_MACHINE, EnvironmentKey, 'Path', OldPath) then
    OldPath := '';
    
  // Check if already in PATH
  if Pos(Uppercase(Dir), Uppercase(OldPath)) > 0 then
    Exit;
    
  // Append
  if OldPath <> '' then
    OldPath := OldPath + ';';
  OldPath := OldPath + Dir;
  
  RegWriteStringValue(HKEY_LOCAL_MACHINE, EnvironmentKey, 'Path', OldPath);
end;

procedure RemoveFromPath(Dir: String);
var
  OldPath, NewPath: String;
  P: Integer;
begin
  if not RegQueryStringValue(HKEY_LOCAL_MACHINE, EnvironmentKey, 'Path', OldPath) then
    Exit;
    
  P := Pos(Uppercase(Dir), Uppercase(OldPath));
  if P = 0 then
    Exit;
    
  NewPath := Copy(OldPath, 1, P - 1) + Copy(OldPath, P + Length(Dir), MaxInt);
  
  // Clean up extra semicolons
  while Pos(';;', NewPath) > 0 do
    StringChangeEx(NewPath, ';;', ';', True);
  if (Length(NewPath) > 0) and (NewPath[Length(NewPath)] = ';') then
    NewPath := Copy(NewPath, 1, Length(NewPath) - 1);
  if (Length(NewPath) > 0) and (NewPath[1] = ';') then
    NewPath := Copy(NewPath, 2, MaxInt);
    
  RegWriteStringValue(HKEY_LOCAL_MACHINE, EnvironmentKey, 'Path', NewPath);
end;

procedure CurStepChanged(CurStep: TSetupStep);
begin
  if CurStep = ssPostInstall then
  begin
    if IsTaskSelected('addtopath') then
      AddToPath(ExpandConstant('{app}'));
  end;
end;

procedure CurUninstallStepChanged(CurUninstallStep: TUninstallStep);
begin
  if CurUninstallStep = usPostUninstall then
  begin
    RemoveFromPath(ExpandConstant('{app}'));
  end;
end;
