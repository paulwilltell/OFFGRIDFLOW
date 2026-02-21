@echo off
echo ========================================
echo Deploying REAL Demo Page
echo ========================================

cd C:\Users\pault\OffGridFlow

echo [1/3] Staging demo page...
git add web/app/demo/page.tsx

echo.
echo [2/3] Committing...
git commit -m "Fix: Replace placeholder demo page with interactive preview

- Add tabbed interface (Dashboard, Reports, Compliance)
- Show real emissions KPIs and quarterly trends  
- Display sample report sections for CSRD, SB 253, SEC, CBAM
- Add compliance status dashboard with completion tracking
- Include data quality metrics
- Add connected data sources visualization
- Professional enterprise UI matching homepage design"

echo.
echo [3/3] Pushing to GitHub...
git push origin main

echo.
echo ========================================
echo DONE! Railway will deploy in 2-3 minutes.
echo Visit https://off-grid-flow.com/demo to see the new page.
echo ========================================
pause
