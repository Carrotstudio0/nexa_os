# Nexa Ultimate WiFi Genesis Script v4.1 - Enhanced Stability
# Optimized for Windows 10/11 with improved error handling

$SSID = "NEXA_PRO_MATRIX"
$Password = "nexa_ultimate_2026"

Write-Host " "
Write-Host " [NEXA] Initializing Wireless Matrix Broadcast..." -ForegroundColor Cyan

# 1. Admin Validation
$currentPrincipal = New-Object Security.Principal.WindowsPrincipal([Security.Principal.WindowsIdentity]::GetCurrent())
if (-not $currentPrincipal.IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)) {
    Write-Host " [CRITICAL] Access Denied. Administrative privileges required." -ForegroundColor Red
    Write-Host " [INFO] Please run this script with Administrator rights." -ForegroundColor Yellow
    exit 1
}

# 2. Hardware Capabilities Audit
Write-Host " [AUDIT] Checking wireless hardware synchronization..." -ForegroundColor Gray
$wlanInfo = netsh wlan show drivers | Out-String
$isHostedSupported = $wlanInfo -match "Hosted network supported.*Yes"

if (-not $isHostedSupported) {
    Write-Host " [WARNING] Hosted Network may not be supported by this driver." -ForegroundColor Yellow
}

# Try Legacy Netsh First
Write-Host " [TRY] Attempting Legacy Hosted Network protocol..." -ForegroundColor Cyan

try {
    # Reset existing hosted network
    netsh wlan set hostednetwork mode=allow ssid=$SSID key=$Password | Out-Null
    Start-Sleep -Milliseconds 500
    
    # Start hosted network
    $netshStart = netsh wlan start hostednetwork 2>&1 | Out-String
    
    if ($netshStart -match "started|running|success") {
        Write-Host " ----------------------------------------------" -ForegroundColor Gray
        Write-Host " ✅  WIRELESS MATRIX ONLINE" -ForegroundColor Green
        Write-Host "     SSID:     $SSID" -ForegroundColor White
        Write-Host "     PASS:     $Password" -ForegroundColor White
        Write-Host " ----------------------------------------------" -ForegroundColor Gray
        exit 0
    }
    else {
        Write-Host " [WARN] Netsh start failed: $($netshStart.Trim())" -ForegroundColor Yellow
    }
}
catch {
    Write-Host " [WARN] Legacy method error: $($_.Exception.Message)" -ForegroundColor Yellow
}

# Try WinRT as a secondary option
Write-Host " [TRY] Attempting WinRT Method as backup..." -ForegroundColor Cyan

try {
    Add-Type -AssemblyName System.Runtime.WindowsRuntime | Out-Null
    
    $TetheringType = [Windows.Networking.NetworkOperators.NetworkOperatorTetheringManager, Windows.Networking.NetworkOperators, ContentType = WindowsRuntime]
    $ConnectivityType = [Windows.Networking.Connectivity.NetworkInformation, Windows.Networking.Connectivity, ContentType = WindowsRuntime]

    $wifiProfile = $ConnectivityType::GetInternetConnectionProfile()
    
    if ($null -ne $wifiProfile) {
        $manager = $TetheringType::CreateFromConnectionProfile($wifiProfile)
        
        if ($manager.TetheringOperationalState -eq 1) {
            Write-Host " [INFO] Hotspot is already active via WinRT." -ForegroundColor Green
            exit 0
        }
        
        Write-Host " [EXEC] Powering on Wireless Matrix (WinRT)..." -ForegroundColor Gray
        $startTask = $manager.StartTetheringAsync()
        
        $timer = [System.Diagnostics.Stopwatch]::StartNew()
        while ($startTask.Status -eq 0 -and $timer.Elapsed.TotalSeconds -lt 10) { 
            Start-Sleep -Milliseconds 200 
        }
        
        if ($manager.TetheringOperationalState -eq 1) {
            Write-Host " ----------------------------------------------" -ForegroundColor Gray
            Write-Host " ✅  WIRELESS MATRIX ONLINE (WinRT)" -ForegroundColor Green
            Write-Host "     SSID:     $SSID" -ForegroundColor White
            Write-Host " ----------------------------------------------" -ForegroundColor Gray
            exit 0
        }
    }
}
catch {
    Write-Host " [INFO] WinRT method failed or not supported." -ForegroundColor Gray
}

# Final Diagnostics
Write-Host " "
Write-Host " [DIAGNOSTICS] Wireless setup encountered issues." -ForegroundColor Yellow
Write-Host "   - Ensure Wi-Fi is ON and Adapter is enabled." -ForegroundColor Gray
Write-Host "   - Some newer laptops only support Mobile Hotspot via Settings." -ForegroundColor Gray

Write-Host " "
Write-Host " [INFO] Continuing without wireless hotspot..." -ForegroundColor Cyan
exit 0
