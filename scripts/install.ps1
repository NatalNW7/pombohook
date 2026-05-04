$ErrorActionPreference = "Stop"

Write-Host "🕊️  Installing PomboHook CLI..." -ForegroundColor Cyan

# Check Arch
$Arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }
$Os = "Windows"

# Fetch latest release from GitHub API
Write-Host "🔍 Fetching latest release version..." -ForegroundColor Yellow
Try {
    $Release = Invoke-RestMethod -Uri "https://api.github.com/repos/NatalNW7/pombohook/releases/latest"
    $Version = $Release.tag_name.TrimStart('v')
} Catch {
    Write-Host "❌ Failed to fetch the latest release version. Please check your internet connection." -ForegroundColor Red
    Exit 1
}

$ZipName = "pombo_${Version}_${Os}_${Arch}.zip"
$DownloadUrl = "https://github.com/NatalNW7/pombohook/releases/latest/download/$ZipName"

Write-Host "⬇️  Downloading $ZipName..." -ForegroundColor Yellow

$TempDir = Join-Path $env:TEMP "pombohook_install"
If (Test-Path $TempDir) { Remove-Item -Recurse -Force $TempDir }
New-Item -ItemType Directory -Path $TempDir | Out-Null

$ZipPath = Join-Path $TempDir $ZipName

Try {
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $ZipPath
} Catch {
    Write-Host "❌ Failed to download the binary. Please check if the release exists." -ForegroundColor Red
    Exit 1
}

Write-Host "📦 Extracting..." -ForegroundColor Yellow
$InstallDir = Join-Path $env:LOCALAPPDATA "PomboHook"
If (!(Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir | Out-Null
}

Expand-Archive -Path $ZipPath -DestinationPath $InstallDir -Force

# Add to PATH
Write-Host "🔧 Configuring PATH..." -ForegroundColor Yellow
$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
If ($UserPath -notmatch [regex]::Escape($InstallDir)) {
    $NewPath = "$UserPath;$InstallDir"
    [Environment]::SetEnvironmentVariable("PATH", $NewPath, "User")
    $env:PATH = "$env:PATH;$InstallDir" # Update current session
    Write-Host "✅ Added $InstallDir to your PATH." -ForegroundColor Green
} Else {
    Write-Host "✅ $InstallDir is already in your PATH." -ForegroundColor Green
}

# Clean up
Remove-Item -Recurse -Force $TempDir

Write-Host ""
Write-Host "🎉 PomboHook CLI installed successfully!" -ForegroundColor Cyan
Write-Host "You can now run 'pombo' from anywhere in your terminal."
Write-Host "(Note: You may need to restart your terminal for PATH changes to take effect)" -ForegroundColor DarkGray
Write-Host ""
Write-Host "🚀 Getting started:" -ForegroundColor Cyan
Write-Host "  1. Authenticate:  pombo ping --server `"ws://your-server:8080`" --token `"your-token`""
Write-Host "  2. Setup route:   pombo route --path=`"/webhook`" --port=3000"
Write-Host "  3. Start proxy:   pombo go"
Write-Host ""
Write-Host "📂 Local configurations are saved in: ~/.pombohook" -ForegroundColor Cyan
Write-Host ""
