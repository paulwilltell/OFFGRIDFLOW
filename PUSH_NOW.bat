@echo off
echo ========================================
echo OffGridFlow - Quick Deploy to Railway
echo ========================================
echo.

cd C:\Users\pault\OffGridFlow

echo [1/4] Checking git status...
git status

echo.
echo [2/4] Staging all changes...
git add .

echo.
echo [3/4] Committing...
git commit -m "Fix: Premium homepage design + Railway production config - URGENT DEPLOYMENT"

echo.
echo [4/4] Pushing to GitHub (will trigger Railway auto-deploy)...
git push origin main

echo.
echo ========================================
echo DONE! Check Railway dashboard for deployment progress.
echo Railway will auto-build and deploy in 2-5 minutes.
echo ========================================
echo.
echo Next: Visit https://off-grid-flow.com in 5 minutes to see new design
pause
