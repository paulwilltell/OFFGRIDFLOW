# OffGridFlow - Local Development Bootstrap (PowerShell)
# Wrapper around dev-start.ps1 for a consistent entrypoint.

param(
    [switch]$Clean,
    [switch]$Logs
)

$ErrorActionPreference = "Stop"

$repoRoot = Resolve-Path (Join-Path $PSScriptRoot ".." "..")
$devStart = Join-Path $PSScriptRoot "dev-start.ps1"

if (-not (Test-Path $devStart)) {
    Write-Error "dev-start.ps1 not found at $devStart"
    exit 1
}

$argsList = @()
if ($Clean) { $argsList += "--clean" }
if ($Logs) { $argsList += "--logs" }

Push-Location $repoRoot
try {
    & $devStart @argsList
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
} finally {
    Pop-Location
}
