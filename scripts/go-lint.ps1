$ErrorActionPreference = "Stop"
Set-Location (Split-Path $PSScriptRoot -Parent)
if (-not (Get-Command golangci-lint -ErrorAction SilentlyContinue)) {
    Write-Error "golangci-lint not found. Install: https://golangci-lint.run/welcome/install/"
}
& golangci-lint run ./...
