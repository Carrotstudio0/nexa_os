param(
    [int]$TimeoutSeconds = 20
)

# Suppress errors for cleaner output
$ErrorActionPreference = 'SilentlyContinue'

Write-Host "`n [NEXA] Scanning for active WiFi Hotspot..." -ForegroundColor Cyan
$hotspotFound = $false
$hotspotInfo = @{}

# Method 1: Fast Check - Hosted Network via netsh
try {
    $netshOutput = netsh wlan show hostednetwork 2>$null | Out-String
    if ($netshOutput -match "Status\s*:\s*Started") {
        Write-Host "  ✅ Hosted Network detected" -ForegroundColor Green
        
        if ($netshOutput -match "SSID\s+:\s+(.+?)[\r\n]") {
            $hotspotInfo['SSID'] = $matches[1].Trim()
        }
        if ($netshOutput -match "Number of clients\s+:\s+(\d+)") {
            $hotspotInfo['Clients'] = $matches[1]
        }
        $hotspotFound = $true
    }
}
catch { }

# Method 2: Windows Mobile Hotspot (WinRT)
if (-not $hotspotFound) {
    try {
        Add-Type -AssemblyName System.Runtime.WindowsRuntime | Out-Null
        $connProfile = [Windows.Networking.Connectivity.NetworkInformation,Windows.Networking.Connectivity,ContentType=WindowsRuntime]::GetInternetConnectionProfile()
        
        if ($null -ne $connProfile) {
            $manager = [Windows.Networking.NetworkOperators.NetworkOperatorTetheringManager,Windows.Networking.NetworkOperators,ContentType=WindowsRuntime]::CreateFromConnectionProfile($connProfile)
            
            if ($manager.TetheringOperationalState -eq 1) {
                Write-Host "  ✅ Mobile Hotspot detected" -ForegroundColor Green
                $hotspotInfo['Type'] = 'Windows Mobile Hotspot'
                $hotspotFound = $true
            }
        }
    }
    catch { }
}

# Polling Loop - if not found immediately, keep checking
$elapsed = 0
$checkInterval = 500  # milliseconds
$pollCount = 0

while (-not $hotspotFound -and $elapsed -lt ($TimeoutSeconds * 1000)) {
    $pollCount++
    
    # Try netsh again
    try {
        $netshOutput = netsh wlan show hostednetwork 2>$null | Out-String
        if ($netshOutput -match "Status\s*:\s*Started") {
            Write-Host "  ✅ Hotspot found at attempt $pollCount" -ForegroundColor Green
            $hotspotFound = $true
            break
        }
    }
    catch { }
    
    # Display progress every 2 attempts
    if ($pollCount % 4 -eq 0) {
        $remaining = $TimeoutSeconds - [int]($elapsed / 1000)
        Write-Host "  ⏳ Scanning... ($remaining`s)" -ForegroundColor Gray -NoNewline
        Write-Host "`r" -NoNewline
    }
    
    Start-Sleep -Milliseconds $checkInterval
    $elapsed += $checkInterval
}

Write-Host ""
Write-Host " ════════════════════════════════════════════════════════" -ForegroundColor Gray

if ($hotspotFound) {
    Write-Host " ✅  HOTSPOT ACTIVE - WiFi mode enabled" -ForegroundColor Green
    if ($hotspotInfo['SSID']) {
        Write-Host "     SSID: $($hotspotInfo['SSID'])" -ForegroundColor White
    }
    if ($hotspotInfo['Clients']) {
        Write-Host "     Clients: $($hotspotInfo['Clients'])" -ForegroundColor White
    }
    Write-Host ""
    exit 0
}
else {
    Write-Host " ℹ️  No Hotspot detected - using wired connection" -ForegroundColor Cyan
    Write-Host ""
    exit 1
}
