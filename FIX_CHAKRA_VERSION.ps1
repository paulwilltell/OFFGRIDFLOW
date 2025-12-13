# Downgrade Chakra UI to v2 (compatible version)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Fixing Chakra UI Version Compatibility" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

Set-Location "C:\Users\pault\OffGridFlow\web"

Write-Host "Uninstalling Chakra UI v3..." -ForegroundColor Yellow
npm uninstall @chakra-ui/react @chakra-ui/next-js @emotion/react @emotion/styled framer-motion

Write-Host ""
Write-Host "Installing Chakra UI v2 (compatible version)..." -ForegroundColor Yellow
npm install @chakra-ui/react@^2.8.2 @chakra-ui/next-js@^2.2.0 @emotion/react@^11.11.4 @emotion/styled@^11.11.5 framer-motion@^10.18.0

Write-Host ""
Write-Host "✓ Chakra UI downgraded to v2" -ForegroundColor Green
Write-Host ""
Write-Host "Restarting dev server..." -ForegroundColor Yellow
Write-Host "Press Ctrl+C to stop the current server, then run:" -ForegroundColor Gray
Write-Host "  npm run dev" -ForegroundColor White
