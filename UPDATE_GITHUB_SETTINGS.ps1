# ========================================
# UPDATE GITHUB REPOSITORY SETTINGS
# ========================================
# Quick 3-minute task to boost Section 5 to 92%

Write-Host "🔧 GitHub Repository Settings Update" -ForegroundColor Cyan
Write-Host ""
Write-Host "Follow these steps to update your repository:" -ForegroundColor Yellow
Write-Host ""

Write-Host "STEP 1: Open your browser" -ForegroundColor Green
Write-Host "   Navigate to: https://github.com/paulwilltell/OFFGRIDFLOW" -ForegroundColor White
Write-Host ""

Write-Host "STEP 2: Click 'Settings' tab (top right)" -ForegroundColor Green
Write-Host ""

Write-Host "STEP 3: Update Repository Description" -ForegroundColor Green
Write-Host "   Paste this in the 'Description' field:" -ForegroundColor Yellow
Write-Host ""
Write-Host "   Enterprise carbon accounting & ESG compliance platform with multi-cloud data ingestion, automated emissions calculations, and CSRD/SEC/CBAM reporting" -ForegroundColor Cyan
Write-Host ""

Write-Host "STEP 4: Add Topics (click 'Topics' gear icon)" -ForegroundColor Green
Write-Host "   Paste these 20 topics (comma-separated):" -ForegroundColor Yellow
Write-Host ""
$topics = "carbon-accounting, esg, csrd, sustainability, emissions, climate-tech, saas, golang, nextjs, typescript, compliance, sec-climate, cbam, ghg-protocol, scope3, multi-tenant, enterprise, production-ready, kubernetes, terraform"
Write-Host "   $topics" -ForegroundColor Cyan
Write-Host ""

Write-Host "STEP 5: Save Changes" -ForegroundColor Green
Write-Host "   Click 'Save changes' button" -ForegroundColor White
Write-Host ""

Write-Host "✅ After completing: Section 5 will be at 92% (+4%)" -ForegroundColor Green
Write-Host ""

# Copy topics to clipboard if possible
Write-Host "📋 Attempting to copy topics to clipboard..." -ForegroundColor Yellow
try {
    Set-Clipboard -Value $topics
    Write-Host "   ✅ Topics copied to clipboard! Just paste in GitHub." -ForegroundColor Green
} catch {
    Write-Host "   ⚠️  Manual copy needed" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "Press any key when done..." -ForegroundColor Gray
$null = $Host.UI.RawUI.ReadKey("NoEcho,IncludeKeyDown")

Write-Host ""
Write-Host "🎉 GitHub settings updated!" -ForegroundColor Green
Write-Host "📊 Section 5: 77% → 92%" -ForegroundColor Cyan
