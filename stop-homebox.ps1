$ErrorActionPreference = "Stop"

$root = $PSScriptRoot
$pidDirectory = Join-Path $root ".homebox-pids"

function Stop-ProcessTree {
    param([int]$ProcessId)

    if (-not (Get-Process -Id $ProcessId -ErrorAction SilentlyContinue)) {
        return
    }

    $children = Get-CimInstance Win32_Process -Filter "ParentProcessId = $ProcessId" -ErrorAction SilentlyContinue
    foreach ($child in $children) {
        Stop-ProcessTree -ProcessId ([int]$child.ProcessId)
    }

    Stop-Process -Id $ProcessId -Force -ErrorAction SilentlyContinue
}

foreach ($name in @("backend.pid", "frontend.pid")) {
    $pidFile = Join-Path $pidDirectory $name
    if (-not (Test-Path $pidFile)) {
        continue
    }

    $processId = [int](Get-Content -Raw $pidFile).Trim()
    Stop-ProcessTree -ProcessId $processId
    Remove-Item -Force $pidFile
}

Write-Host "Homebox wurde gestoppt."
