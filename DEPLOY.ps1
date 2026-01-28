# OffGridFlow Railway Deployment Script
# Run this to deploy fixes to production

Write-Host "🚀 OffGridFlow Production Deployment" -ForegroundColor Green
Write-Host "=====================================" -ForegroundColor Green
Write-Host ""

# Check if we're in the right directory
if (-not (Test-Path "web\app\page.tsx")) {
    Write-Host "❌ Error: Must run from OffGridFlow project root" -ForegroundColor Red
    exit 1
}

Write-Host "📋 Changes to be deployed:" -ForegroundColor Yellow
Write-Host "  ✅ Premium glassmorphic homepage design" -ForegroundColor White
Write-Host "  ✅ Railway deployment configuration" -ForegroundColor White
Write-Host "  ✅ Environment variable templates" -ForegroundColor White
Write-Host "  ✅ Next.js build fixes" -ForegroundColor White
Write-Host ""

Write-Host "⚠️  BEFORE DEPLOYING:" -ForegroundColor Yellow
Write-Host "  1. Update Railway environment variables (see DEPLOYMENT_FIX_GUIDE.md)" -ForegroundColor White
Write-Host "  2. Generate JWT secret: openssl rand -base64 48" -ForegroundColor White
Write-Host "  3. Set OFFGRIDFLOW_DB_DSN=${{Postgres.DATABASE_URL}} in Railway" -ForegroundColor White
Write-Host ""

$response = Read-Host "Have you updated Railway environment variables? (y/n)"
if ($response -ne "y") {
    Write-Host "❌ Deployment cancelled. Update Railway variables first." -ForegroundColor Red
    Write-Host "📖 See DEPLOYMENT_FIX_GUIDE.md for instructions" -ForegroundColor Yellow
    exit 1
}

Write-Host ""
Write-Host "🔄 Staging changes..." -ForegroundColor Cyan
git add .

Write-Host "💾 Creating commit..." -ForegroundColor Cyan
git commit -m "Fix: Premium homepage + Railway production config

- Replace basic homepage with glassmorphic design
- Add animated 3D globe with cloud provider nodes
- Add live statistics dashboard
- Create Railway deployment configuration
- Fix Next.js Server Action errors
- Add proper environment variable templates
- Document API and database configuration fixes"

Write-Host "📤 Pushing to GitHub..." -ForegroundColor Cyan
git push origin main

Write-Host ""
Write-Host "✅ Code deployed to GitHub!" -ForegroundColor Green
Write-Host ""
Write-Host "🔍 Next steps:" -ForegroundColor Yellow
Write-Host "  1. Check Railway dashboard for auto-deployment progress" -ForegroundColor White
Write-Host "  2. Monitor build logs in Railway" -ForegroundColor White
Write-Host "  3. Test https://off-grid-flow.com for new homepage" -ForegroundColor White
Write-Host "  4. Test registration form for proper API communication" -ForegroundColor White
Write-Host ""
Write-Host "📊 Railway Dashboard:" -ForegroundColor Cyan
Write-Host "  https://railway.com/project/99b5cf9a-451d-47e5-be0f-fcb8eee95aff" -ForegroundColor Blue
Write-Host ""
Write-Host "🎉 Deployment initiated!" -ForegroundColor Green
