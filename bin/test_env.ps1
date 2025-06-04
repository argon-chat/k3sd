# PowerShell equivalent of test_env.sh for Multipass on Windows
# Usage: .\test_env.ps1 [-Multinode] [-Teardown]

param(
    [switch]$Multinode,
    [switch]$Teardown
)

function Check-Multipass {
    if (-not (Get-Command multipass -ErrorAction SilentlyContinue)) {
        Write-Error "multipass could not be found. Please install multipass first."
        exit 1
    }
}

Check-Multipass

if ($Teardown) {
    Write-Host "Deleting all multipass nodes..."
    $nodes = multipass list --format csv | Select-Object -Skip 1 | ForEach-Object { ($_ -split ',')[0] }
    if ($nodes) {
        multipass delete $nodes --purge
    } else {
        Write-Host "No multipass nodes to delete."
    }
    multipass list
    exit 0
}

$nodes = if ($Multinode) { @('node1', 'node2') } else { @('node1') }
$memory = if ($Multinode) { '2G' } else { '4G' }
$pass = $env:MPS_PASSWORD
if (-not $pass) { $pass = 'password123' }

foreach ($node in $nodes) {
    multipass launch --name $node --cpus 2 --memory $memory --disk 10G
    multipass exec $node -- sudo bash -c "echo ubuntu:$pass | sudo chpasswd"
    foreach ($key in Get-ChildItem -Path $env:USERPROFILE\.ssh\*.pub) {
        $pubkey = Get-Content $key
        multipass exec $node -- bash -c "echo '$pubkey' >> ~/.ssh/authorized_keys"
    }
    Write-Host "Node $node ready."
}
