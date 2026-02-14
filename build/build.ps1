# GoSer Build Script for Windows
# Usage: .\build\build.ps1 [-Installer]
#
# Prerequisites:
#   - Go 1.21+
#   - Node.js 18+
#   - Inno Setup 6 (only for -Installer flag)
#     Download: https://jrsoftware.org/isinfo.php

param(
    [switch]$Installer  # Also build the installer
)

$ErrorActionPreference = "Stop"

$ROOT = Split-Path -Parent (Split-Path -Parent $MyInvocation.MyCommand.Path)
$OUTPUT = Join-Path $ROOT "dist"
$FRONTEND = Join-Path $ROOT "cmd\app\frontend"

Write-Host "=== GoSer Build ===" -ForegroundColor Cyan
Write-Host "Root: $ROOT" -ForegroundColor Gray

# Clean dist (but keep installer subfolder)
if (Test-Path $OUTPUT) {
    Get-ChildItem $OUTPUT -Exclude "installer" | Remove-Item -Recurse -Force
}
New-Item -ItemType Directory -Path $OUTPUT -Force | Out-Null

# Step 0: Embed icon into exe resources using go-winres
Write-Host "`n[0/4] Embedding icon resources..." -ForegroundColor Yellow
$LOGO = Join-Path $ROOT "build\windows\logo.ico"
if (Test-Path $LOGO) {
    $targets = @(
        @{ Dir = "cmd\app";    Manifest = "gui"; Name = "GoSer";        Desc = "GoSer Service Manager" },
        @{ Dir = "cmd\goserd"; Manifest = "cli"; Name = "GoSer Daemon"; Desc = "GoSer Daemon Service" },
        @{ Dir = "cmd\goser";  Manifest = "cli"; Name = "GoSer CLI";    Desc = "GoSer Command Line Tool" }
    )
    foreach ($t in $targets) {
        Push-Location (Join-Path $ROOT $t.Dir)
        go run github.com/tc-hib/go-winres@latest simply --icon $LOGO --manifest $t.Manifest --product-name $t.Name --file-description $t.Desc --arch amd64 2>$null
        Pop-Location
    }
    Write-Host "  Icon embedded." -ForegroundColor Green
} else {
    Write-Host "  logo.png not found, skipping" -ForegroundColor Gray
}

# Step 1: Build frontend
Write-Host "`n[1/4] Building frontend..." -ForegroundColor Yellow
Push-Location $FRONTEND
npm install --silent
npx vite build
Pop-Location
Write-Host "  Frontend built." -ForegroundColor Green

# Step 2: Build daemon
Write-Host "`n[2/4] Building goserd.exe..." -ForegroundColor Yellow
Push-Location $ROOT
go build -ldflags="-s -w" -o "$OUTPUT\goserd.exe" ./cmd/goserd
Pop-Location
Write-Host "  goserd.exe built." -ForegroundColor Green

# Step 3: Build CLI
Write-Host "`n[3/4] Building goser.exe..." -ForegroundColor Yellow
Push-Location $ROOT
go build -ldflags="-s -w" -o "$OUTPUT\goser.exe" ./cmd/goser
Pop-Location
Write-Host "  goser.exe built." -ForegroundColor Green

# Step 4: Build GUI app (requires Wails build tags + hide console window)
Write-Host "`n[4/4] Building goser-app.exe..." -ForegroundColor Yellow
Push-Location $ROOT
go build -tags "desktop,production" -ldflags="-s -w -H=windowsgui" -o "$OUTPUT\goser-app.exe" ./cmd/app
Pop-Location
Write-Host "  goser-app.exe built." -ForegroundColor Green

# Summary
Write-Host "`n=== Binaries ===" -ForegroundColor Cyan
Get-ChildItem "$OUTPUT\*.exe" | Format-Table Name, @{N="Size(KB)";E={[math]::Round($_.Length/1KB)}} -AutoSize

# Step 5: Build installer (optional)
if ($Installer) {
    Write-Host "`n[5] Building installer..." -ForegroundColor Yellow

    # Find Inno Setup compiler
    $ISCC = $null
    $searchPaths = @(
        "${env:ProgramFiles(x86)}\Inno Setup 6\ISCC.exe",
        "${env:ProgramFiles}\Inno Setup 6\ISCC.exe",
        "C:\Program Files (x86)\Inno Setup 6\ISCC.exe",
        "C:\Program Files\Inno Setup 6\ISCC.exe"
    )
    foreach ($p in $searchPaths) {
        if (Test-Path $p) { $ISCC = $p; break }
    }

    if (-not $ISCC) {
        Write-Host "  ERROR: Inno Setup 6 not found!" -ForegroundColor Red
        Write-Host "  Download from: https://jrsoftware.org/isinfo.php" -ForegroundColor Yellow
        Write-Host "  After installing, re-run: .\build\build.ps1 -Installer" -ForegroundColor Yellow
    } else {
        $ISS = Join-Path $ROOT "build\windows\installer.iss"
        New-Item -ItemType Directory -Path "$OUTPUT\installer" -Force | Out-Null
        & $ISCC $ISS
        if ($LASTEXITCODE -eq 0) {
            Write-Host "  Installer built." -ForegroundColor Green
            $installerFile = Get-ChildItem "$OUTPUT\installer\*.exe" | Select-Object -First 1
            if ($installerFile) {
                Write-Host "  Output: $($installerFile.FullName)" -ForegroundColor Gray
                Write-Host "  Size: $([math]::Round($installerFile.Length/1MB, 1)) MB" -ForegroundColor Gray
            }
        } else {
            Write-Host "  ERROR: Installer build failed!" -ForegroundColor Red
        }
    }
}

Write-Host "`n=== Build Complete ===" -ForegroundColor Green
