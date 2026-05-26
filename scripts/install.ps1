# gm installer for Windows (PowerShell 5.1+)
# Usage:
#   irm https://raw.githubusercontent.com/angei24/go-manager/main/scripts/install.ps1 | iex
#   .\scripts\install.ps1
#   .\scripts\install.ps1 -FromSource

[CmdletBinding()]
param(
    [switch]$FromSource,
    [string]$InstallDir = "",
    [string]$Repo = "angei24/go-manager",
    [string]$Branch = "main",
    [string]$Version = "",
    [switch]$AddToPath
)

$ErrorActionPreference = "Stop"

function Write-Step([string]$Message) { Write-Host "==> $Message" -ForegroundColor Cyan }
function Write-Warn([string]$Message) { Write-Host "warning: $Message" -ForegroundColor Yellow }
function Write-Err([string]$Message) { Write-Host "error: $Message" -ForegroundColor Red; exit 1 }

function Get-DefaultInstallDir {
    if ($InstallDir) { return $InstallDir }
    $local = Join-Path $env:USERPROFILE ".local\bin"
    if (-not (Test-Path (Split-Path $local))) {
        New-Item -ItemType Directory -Force -Path (Split-Path $local) | Out-Null
    }
    return $local
}

function Get-RepoRootFromScript {
    $root = Split-Path -Parent $PSScriptRoot
    if ((Test-Path (Join-Path $root "go.mod")) -and (Test-Path (Join-Path $root "cmd\gm\main.go"))) {
        return $root
    }
    return $null
}

function Get-PlatformAsset {
    if ($env:PROCESSOR_ARCHITECTURE -match "ARM64") {
        return "windows", "arm64"
    }
    if ([Environment]::Is64BitOperatingSystem) {
        return "windows", "amd64"
    }
    return "windows", "386"
}

function Test-InPath([string]$Dir) {
    $parts = $env:Path -split ';' | ForEach-Object { $_.TrimEnd('\') }
    $norm = $Dir.TrimEnd('\')
    return $parts -contains $norm
}

function Show-PathHint([string]$Dir) {
    if (Test-InPath $Dir) { return }
    Write-Host ""
    Write-Host "Add gm to your PATH:"
    Write-Host "  `$env:Path = `"$Dir;`$env:Path`""
    Write-Host "  Or: Settings -> System -> Environment Variables -> User Path -> Add -> $Dir"
    if ($AddToPath) {
        Write-Step "Adding $Dir to user PATH ..."
        $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
        if (-not (Test-InPath $Dir)) {
            [Environment]::SetEnvironmentVariable("Path", "$Dir;$userPath", "User")
            $env:Path = "$Dir;$env:Path"
            Write-Host "PATH updated for current user."
        }
    }
}

function Build-FromDir([string]$Root, [string]$Dest) {
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) {
        Write-Err "Go 1.21+ is required to build from source. Install from https://go.dev/dl/"
    }
    Write-Step "Building gm from $Root ..."
    Push-Location $Root
    try {
        & go build -ldflags="-s -w" -o $Dest .\cmd\gm
    } finally {
        Pop-Location
    }
}

function Build-FromGit([string]$Dest) {
    if (-not (Get-Command git -ErrorAction SilentlyContinue)) { Write-Err "git is required" }
    if (-not (Get-Command go -ErrorAction SilentlyContinue)) { Write-Err "Go 1.21+ is required" }
    $tmp = Join-Path $env:TEMP ("gm-install-" + [guid]::NewGuid().ToString("n"))
    New-Item -ItemType Directory -Force -Path $tmp | Out-Null
    try {
        Write-Step "Cloning https://github.com/$Repo.git (branch $Branch) ..."
        & git clone --depth 1 --branch $Branch "https://github.com/$Repo.git" (Join-Path $tmp "repo")
        Build-FromDir (Join-Path $tmp "repo") $Dest
    } finally {
        Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
    }
}

function Get-LatestReleaseTag {
    if ($Version -and $Version -ne "latest") { return $Version }
    try {
        $resp = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest" -UseBasicParsing
        return $resp.tag_name
    } catch {
        return $null
    }
}

function Install-FromRelease([string]$Dest) {
    $tag = Get-LatestReleaseTag
    if (-not $tag) { return $false }
    $os, $arch = Get-PlatformAsset
    $asset = "gm_${tag}_${os}_${arch}.zip"
    $url = "https://github.com/$Repo/releases/download/$tag/$asset"
    $tmp = Join-Path $env:TEMP ("gm-dl-" + [guid]::NewGuid().ToString("n"))
    New-Item -ItemType Directory -Force -Path $tmp | Out-Null
    try {
        Write-Step "Downloading $url ..."
        $zip = Join-Path $tmp $asset
        Invoke-WebRequest -Uri $url -OutFile $zip -UseBasicParsing
        Expand-Archive -Path $zip -DestinationPath $tmp -Force
        $bin = Join-Path $tmp "gm.exe"
        if (-not (Test-Path $bin)) { $bin = Join-Path $tmp "gm" }
        if (-not (Test-Path $bin)) {
            Write-Warn "release archive missing gm binary"
            return $false
        }
        Copy-Item -Force $bin $Dest
        return $true
    } catch {
        return $false
    } finally {
        Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
    }
}

function Install-Gm {
    $installDir = Get-DefaultInstallDir
    if (-not (Test-Path $installDir)) {
        New-Item -ItemType Directory -Force -Path $installDir | Out-Null
    }
    $dest = Join-Path $installDir "gm.exe"
    $os, $arch = Get-PlatformAsset
    Write-Step "Platform: $os/$arch"
    Write-Step "Install dir: $installDir"

    if ($FromSource -or $Version -eq "source") {
        $root = Get-RepoRootFromScript
        if ($root) { Build-FromDir $root $dest } else { Build-FromGit $dest }
        return $installDir, $dest
    }

    if (Install-FromRelease $dest) {
        Write-Step "Installed release binary to $dest"
        return $installDir, $dest
    }

    Write-Warn "No GitHub release found (or download failed); building from source ..."
    $root = Get-RepoRootFromScript
    if ($root) { Build-FromDir $root $dest } else { Build-FromGit $dest }
    return $installDir, $dest
}

function Verify-Install([string]$Dest) {
    if (-not (Test-Path $Dest)) { Write-Err "install failed: $Dest not found" }
    Write-Step "Verifying installation ..."
    & $Dest --help | Out-Null
    Write-Step "Success! gm is ready at $Dest"
}

$dir, $binary = Install-Gm
Verify-Install $binary
Show-PathHint $dir
