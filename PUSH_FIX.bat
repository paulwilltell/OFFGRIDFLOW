@echo off
echo ========================================
echo OffGridFlow - FIXED Deployment
echo ========================================
echo.

cd C:\Users\pault\OffGridFlow

echo [1/3] Staging changes (removed bad config files)...
git add -A

echo.
echo [2/3] Committing fix...
git commit -m "Fix: Remove broken Railway config - let Railway auto-detect build"

echo.
echo [3/3] Pushing to GitHub...
git push origin main

echo.
echo ========================================
echo DEPLOYED! Railway will now auto-detect the build process.
echo Wait 3-5 minutes and check Railway dashboard.
echo ========================================
pause
