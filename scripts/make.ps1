param(
    [Parameter(Mandatory = $true)]
    [ValidateSet("help", "tidy", "build", "run", "run-worker", "migrate-up", "migrate-down", "test-unit", "test", "swagger", "clean")]
    [string]$Target,
    [string]$Env = "dev",
    [string]$Dsn = "",
    [string]$Driver = ""
)

$ErrorActionPreference = "Stop"
Set-Location (Join-Path $PSScriptRoot "..")

switch ($Target) {
    "help" {
        Write-Host "Available targets:"
        Write-Host "  .\scripts\make.ps1 -Target tidy"
        Write-Host "  .\scripts\make.ps1 -Target build"
        Write-Host "  .\scripts\make.ps1 -Target run -Env dev"
        Write-Host "  .\scripts\make.ps1 -Target run-worker -Env dev"
        Write-Host "  .\scripts\make.ps1 -Target migrate-up -Driver mysql -Dsn <database_dsn>"
        Write-Host "  .\scripts\make.ps1 -Target migrate-down -Driver mysql -Dsn <database_dsn>"
        Write-Host "  .\scripts\make.ps1 -Target test-unit"
        Write-Host "  .\scripts\make.ps1 -Target test"
        Write-Host "  .\scripts\make.ps1 -Target swagger"
        Write-Host "  .\scripts\make.ps1 -Target clean"
    }
    "tidy" { go mod tidy }
    "build" {
        if (!(Test-Path "bin")) { New-Item -ItemType Directory -Path "bin" | Out-Null }
        go build -o "bin/server.exe" ./cmd/server
    }
    "run" { go run ./cmd/server server --env $Env }
    "run-worker" { go run ./cmd/server worker --env $Env }
    "migrate-up" {
        if ([string]::IsNullOrWhiteSpace($Dsn)) { throw "Dsn is required for migrate-up" }
        if ([string]::IsNullOrWhiteSpace($Driver)) { throw "Driver is required for migrate-up" }
        go run ./cmd/migrate up --driver $Driver --dsn $Dsn
    }
    "migrate-down" {
        if ([string]::IsNullOrWhiteSpace($Dsn)) { throw "Dsn is required for migrate-down" }
        if ([string]::IsNullOrWhiteSpace($Driver)) { throw "Driver is required for migrate-down" }
        go run ./cmd/migrate down --driver $Driver --dsn $Dsn
    }
    "test-unit" { go test ./tests/unit/... }
    "test" { go test ./... }
    "swagger" { go run github.com/swaggo/swag/cmd/swag@latest init -g main.go -o docs -d ./cmd/server,./api }
    "clean" {
        if (Test-Path "bin") { Remove-Item -Recurse -Force "bin" }
    }
}
