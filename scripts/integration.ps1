param(
    [Parameter(Mandatory = $false)]
    [ValidateSet("help", "deps-up", "deps-down", "prepare-db", "test", "all")]
    [string]$Action = "help",
    [string]$BaseUrl = "http://127.0.0.1:18080",
    [string]$AdminUsername = "admin",
    [string]$SecretEnvName = "INTEGRATION_ADMIN_PASSWORD",
    [string]$TenantID = ""
)

$ErrorActionPreference = "Stop"
$root = Join-Path $PSScriptRoot ".."
Set-Location $root

function Wait-HttpReady {
    param([string]$Url, [int]$TimeoutSec = 60)
    $deadline = (Get-Date).AddSeconds($TimeoutSec)
    while ((Get-Date) -lt $deadline) {
        try {
            $resp = Invoke-WebRequest -Uri $Url -Method Get -TimeoutSec 3 -UseBasicParsing
            if ($resp.StatusCode -ge 200 -and $resp.StatusCode -lt 500) {
                return
            }
        } catch {
        }
        Start-Sleep -Seconds 2
    }
    throw "timeout waiting for $Url"
}

function Start-ComposeDeps {
    docker compose up -d mysql redis
}

function Initialize-TestDB {
    Write-Host "[integration] waiting mysql container..."
    $ok = $false
    for ($i = 0; $i -lt 30; $i++) {
        try {
            docker compose exec -T mysql mysqladmin ping -h 127.0.0.1 -uroot -proot | Out-Null
            $ok = $true
            break
        } catch {
            Start-Sleep -Seconds 2
        }
    }
    if (-not $ok) {
        throw "mysql not ready"
    }

    docker compose exec -T mysql mysql -uroot -proot -e "CREATE DATABASE IF NOT EXISTS scaffold_test CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;"
    go run ./cmd/migrate up --env test --driver mysql --dsn "root:root@tcp(127.0.0.1:3306)/scaffold_test?charset=utf8mb4&parseTime=True"
    Get-Content ".\tests\integration\fixtures\base.sql" | docker compose exec -T mysql mysql -uroot -proot scaffold_test
}

function Invoke-IntegrationTests {
    $adminSecret = [Environment]::GetEnvironmentVariable($SecretEnvName)
    if ([string]::IsNullOrWhiteSpace($adminSecret)) {
        $adminSecret = "admin123456"
    }
    $env:INTEGRATION_BASE_URL = $BaseUrl
    $env:INTEGRATION_ADMIN_USERNAME = $AdminUsername
    $env:INTEGRATION_ADMIN_PASSWORD = $adminSecret
    $env:INTEGRATION_TENANT_ID = $TenantID
    go test -tags=integration ./tests/integration -v
}

function Invoke-All {
    Start-ComposeDeps
    Initialize-TestDB

    $server = Start-Process -FilePath "go" -ArgumentList @("run", "./cmd/server", "server", "--env", "test") -PassThru -NoNewWindow
    $worker = Start-Process -FilePath "go" -ArgumentList @("run", "./cmd/server", "worker", "--env", "test") -PassThru -NoNewWindow
    try {
        Wait-HttpReady -Url "$BaseUrl/livez" -TimeoutSec 90
        Invoke-IntegrationTests
    } finally {
        if ($null -ne $worker -and -not $worker.HasExited) {
            Stop-Process -Id $worker.Id -Force -ErrorAction SilentlyContinue
        }
        if ($null -ne $server -and -not $server.HasExited) {
            Stop-Process -Id $server.Id -Force -ErrorAction SilentlyContinue
        }
    }
}

switch ($Action) {
    "help" {
        Write-Host "Integration helper:"
        Write-Host "  .\scripts\integration.ps1 -Action deps-up"
        Write-Host "  .\scripts\integration.ps1 -Action prepare-db"
        Write-Host "  .\scripts\integration.ps1 -Action test -BaseUrl http://127.0.0.1:18080"
        Write-Host "  .\scripts\integration.ps1 -Action all"
    }
    "deps-up" { Start-ComposeDeps }
    "deps-down" { docker compose down }
    "prepare-db" { Initialize-TestDB }
    "test" { Invoke-IntegrationTests }
    "all" { Invoke-All }
}
