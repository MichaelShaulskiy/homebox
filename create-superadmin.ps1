param(
    [Parameter(Mandatory = $true)][string]$Name,
    [Parameter(Mandatory = $true)][string]$Email,
    [Parameter(Mandatory = $true)][string]$Password,
    [string]$GroupName = ""
)

$root = $PSScriptRoot
Set-Location (Join-Path $root "backend")
$env:UNSAFE_DISABLE_PASSWORD_PROJECTION = "yes_i_am_sure"

$arguments = @(
    "run",
    "./app/api",
    "create-superadmin",
    "--name=$Name",
    "--email=$Email",
    "--password=$Password"
)
if ($GroupName) {
    $arguments += "--group-name=$GroupName"
}

& go @arguments
