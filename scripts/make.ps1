param(
    [Parameter(Mandatory = $true)]
    [ValidateSet("help", "tidy", "build", "run", "run-worker", "migrate-up", "migrate-down", "test-unit", "test", "cover", "quality", "swagger", "integration-test", "integration-all", "clean")]
    [string]$Target,
    [string]$Env = "dev",
    [string]$Profile = "",
    [string]$Dsn = "",
    [string]$Driver = "mysql",
    [string]$TimeZone = "UTC"
)

$ErrorActionPreference = "Stop"
Set-Location (Join-Path $PSScriptRoot "..")

switch ($Target) {
    "help" {
        Write-Host "Available targets:"
        Write-Host "  .\scripts\make.ps1 -Target tidy"
        Write-Host "  .\scripts\make.ps1 -Target build"
        Write-Host "  .\scripts\make.ps1 -Target run -Env dev -Profile order"
        Write-Host "  .\scripts\make.ps1 -Target run-worker -Env dev -Profile order"
        Write-Host "  .\scripts\make.ps1 -Target migrate-up -Dsn <database_dsn>   # schema + seed"
        Write-Host "  .\scripts\make.ps1 -Target migrate-down -Dsn <database_dsn> # one schema rollback"
        Write-Host "  .\scripts\make.ps1 -Target test-unit"
        Write-Host "  .\scripts\make.ps1 -Target test"
        Write-Host "  .\scripts\make.ps1 -Target cover"
        Write-Host "  .\scripts\make.ps1 -Target quality   # gofmt + test + coverage gate"
        Write-Host "  .\scripts\make.ps1 -Target swagger"
        Write-Host "  .\scripts\make.ps1 -Target integration-test"
        Write-Host "  .\scripts\make.ps1 -Target integration-all"
        Write-Host "  .\scripts\make.ps1 -Target clean"
    }
    "tidy" { go mod tidy }
    "build" {
        if (!(Test-Path "bin")) { New-Item -ItemType Directory -Path "bin" | Out-Null }
        go build -o "bin/server.exe" ./cmd/server
    }
    "run" { go run ./cmd/server server --env $Env --profile $Profile }
    "run-worker" { go run ./cmd/server worker --env $Env --profile $Profile }
    "migrate-up" {
        if ([string]::IsNullOrWhiteSpace($Dsn)) { throw "Dsn is required for migrate-up" }
        go run ./cmd/migrate up --driver $Driver --dsn $Dsn --time-zone $TimeZone
        go run ./cmd/migrate seed up --driver $Driver --dsn $Dsn --time-zone $TimeZone
    }
    "migrate-down" {
        if ([string]::IsNullOrWhiteSpace($Dsn)) { throw "Dsn is required for migrate-down" }
        go run ./cmd/migrate down --driver $Driver --dsn $Dsn --time-zone $TimeZone
    }
    "test-unit" { go test ./tests/unit/... }
    "test" { go test ./... }
    "cover" { & ".\scripts\go-cover.ps1" }
    "quality" {
        $unformatted = gofmt -l .
        if ($unformatted) {
            throw "gofmt required on:`n$unformatted"
        }
        go test ./...
        if ($env:CGO_ENABLED -eq "1") {
            go test -race ./...
        }
        & ".\scripts\go-cover.ps1"
    }
    "swagger" { go run github.com/swaggo/swag/cmd/swag@latest init -g main.go -o docs -d ./cmd/server,./api }
    "integration-test" { & ".\scripts\integration.ps1" -Action test }
    "integration-all" { & ".\scripts\integration.ps1" -Action all }
    "clean" {
        if (Test-Path "bin") { Remove-Item -Recurse -Force "bin" }
    }
}

