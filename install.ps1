# Kernel Windows One-Line Installer
$ErrorActionPreference = "Stop"

Write-Host "`n⚡ Installing Kernel AI Agent..." -ForegroundColor Cyan

$Repo = "404kernel/kernel"
$InstallDir = "$env:USERPROFILE\.kernel\bin"

# Determine architecture
$Arch = "amd64"
if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
    $Arch = "arm64"
}

$BinaryName = "kernel-windows-$Arch.exe"
$DownloadUrl = "https://github.com/$Repo/releases/latest/download/$BinaryName"
$DestPath = "$InstallDir\kernel.exe"

if (!(Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

Write-Host "• Downloading latest release ($Arch)..." -ForegroundColor Yellow
try {
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
    Invoke-WebRequest -Uri $DownloadUrl -OutFile $DestPath -UseBasicParsing
} catch {
    Write-Host "Could not download pre-built binary. Falling back to Go build..." -ForegroundColor DarkYellow
    if (Get-Command go -ErrorAction SilentlyContinue) {
        go install github.com/404kernel/kernel@latest
        exit 0
    } else {
        Write-Error "Failed to download binary from $DownloadUrl: $_"
    }
}

# Add to user PATH if not present
$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("Path", "$UserPath;$InstallDir", "User")
    Write-Host "• Added $InstallDir to User PATH." -ForegroundColor Green
}

Write-Host "`n🎉 Installation complete! Sparky is ready." -ForegroundColor Green
Write-Host "Run 'kernel' in any new terminal window.`n" -ForegroundColor Cyan
