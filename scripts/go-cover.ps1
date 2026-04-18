$ErrorActionPreference = "Stop"
Set-Location (Split-Path $PSScriptRoot -Parent)

$thresholdRaw = if ($env:COVERAGE_THRESHOLD) { $env:COVERAGE_THRESHOLD } else { "22" }
$threshold = 0.0
if (-not [double]::TryParse($thresholdRaw, [ref]$threshold)) {
    throw "COVERAGE_THRESHOLD must be numeric, got: $thresholdRaw"
}

if (-not (Test-Path "coverage")) {
    New-Item -ItemType Directory -Path "coverage" | Out-Null
}
$profile = "coverage/coverage.out"
$summary = "coverage/coverage.txt"

# Optional: limit parallel packages if the machine spikes CPU/RAM, e.g. $env:GOTEST_PARALLEL='2'
$parallelArgs = @()
if (-not [string]::IsNullOrWhiteSpace($env:GOTEST_PARALLEL)) {
    $parallelArgs = @("-p", $env:GOTEST_PARALLEL.Trim())
}

& go test @parallelArgs -covermode=atomic -coverprofile $profile ./...
$coverOutput = go tool cover "-func=$profile"
$coverOutput | Set-Content -Encoding UTF8 $summary
$coverOutput

$totalLine = $coverOutput | Where-Object { $_ -match '^total:' }
if (-not $totalLine) {
    throw "Failed to parse total coverage"
}
if ($totalLine -notmatch '([0-9]+(\.[0-9]+)?)%') {
    throw "Failed to parse total coverage percentage from: $totalLine"
}
$total = [double]$Matches[1]
if ($total -lt $threshold) {
    throw "Coverage gate failed: $total% < $threshold%"
}
Write-Host "Coverage gate passed: $total% >= $threshold%"
