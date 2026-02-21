# FIX_RAILWAY_DATABASE.ps1
# This script fixes the database connection for OffGridFlow API in Railway

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "FIXING RAILWAY DATABASE CONNECTION" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

Write-Host "CRITICAL ISSUE IDENTIFIED:" -ForegroundColor Yellow
Write-Host "OFFGRIDFLOW_DB_DSN is set to localhost:5432" -ForegroundColor Red
Write-Host "It needs to be set to Railway's Postgres service" -ForegroundColor Yellow
Write-Host ""

Write-Host "TO FIX THIS IN RAILWAY:" -ForegroundColor Green
Write-Host "1. Go to: https://railway.app/project/99b5cf9a-451d-47e5-be0f-fcb8eee95aff" -ForegroundColor White
Write-Host "2. Click on 'offgridflow-api' service" -ForegroundColor White
Write-Host "3. Click 'Variables' tab" -ForegroundColor White
Write-Host "4. Find OFFGRIDFLOW_DB_DSN variable" -ForegroundColor White
Write-Host "5. Click the ... menu button next to it" -ForegroundColor White
Write-Host "6. Click 'Edit'" -ForegroundColor White
Write-Host "7. Change the value from:" -ForegroundColor White
Write-Host "   postgres://offgridflow:changeme@localhost:5432/offgridflow?sslmode=disable" -ForegroundColor Red
Write-Host "   TO:" -ForegroundColor White
Write-Host "   `${{Postgres.DATABASE_URL}}" -ForegroundColor Green
Write-Host "8. Click 'Update'" -ForegroundColor White
Write-Host "9. Railway will automatically redeploy the API" -ForegroundColor White
Write-Host ""

Write-Host "ALSO FIX THESE WHILE YOU'RE THERE:" -ForegroundColor Yellow
Write-Host "- OFFGRIDFLOW_APP_ENV: Change from 'development' to 'production'" -ForegroundColor White
Write-Host "- OFFGRIDFLOW_JWT_SECRET: Generate a new secret with:" -ForegroundColor White
Write-Host "  openssl rand -base64 48" -ForegroundColor Cyan
Write-Host ""

Write-Host "After fixing, the API will connect to the database and registration will work!" -ForegroundColor Green
Write-Host ""

# Generate a JWT secret for them
Write-Host "HERE'S A NEW JWT SECRET FOR YOU:" -ForegroundColor Cyan
$bytes = New-Object byte[] 48
[System.Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($bytes)
$jwt_secret = [Convert]::ToBase64String($bytes)
Write-Host $jwt_secret -ForegroundColor Green
Write-Host ""

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Copy the JWT secret above and use it!" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

pause
