$ErrorActionPreference = "Stop"

$root = $PSScriptRoot
$pidDirectory = Join-Path $root ".homebox-pids"
$taskExecutable = Join-Path $env:USERPROFILE "go\bin\task.exe"
$pnpmExecutable = (Get-Command pnpm.cmd -ErrorAction SilentlyContinue)?.Source

if (-not (Test-Path $taskExecutable)) {
    throw "Task wurde nicht gefunden: $taskExecutable"
}

if (-not $pnpmExecutable) {
    throw "pnpm wurde nicht gefunden. Fuehre 'npm install -g pnpm@10.28.0' aus."
}

New-Item -ItemType Directory -Force -Path $pidDirectory | Out-Null

$backendCommand = @"
Set-Location '$root'
`$env:PATH = '$env:USERPROFILE\go\bin;' + `$env:PATH
& '$taskExecutable' go:run
"@

$frontendCommand = @"
Set-Location '$root\frontend'
& '$pnpmExecutable' dev --no-fork
"@

$backendProcess = Start-Process `
    -FilePath "powershell.exe" `
    -ArgumentList @("-NoProfile", "-NoExit", "-Command", $backendCommand) `
    -WorkingDirectory $root `
    -PassThru

$frontendProcess = Start-Process `
    -FilePath "powershell.exe" `
    -ArgumentList @("-NoProfile", "-NoExit", "-Command", $frontendCommand) `
    -WorkingDirectory (Join-Path $root "frontend") `
    -PassThru

Set-Content -Path (Join-Path $pidDirectory "backend.pid") -Value $backendProcess.Id
Set-Content -Path (Join-Path $pidDirectory "frontend.pid") -Value $frontendProcess.Id

Write-Host "Homebox wird gestartet."
Write-Host "Backend:  http://localhost:7745"
Write-Host "Frontend: http://localhost:3000"
