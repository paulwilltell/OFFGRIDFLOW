@echo off
echo ========================================
echo Generating package-lock.json
echo ========================================

cd C:\Users\pault\OffGridFlow\web

echo [1/4] Running npm install to generate package-lock.json...
npm install

echo.
echo [2/4] Removing problematic config files...
del nixpacks.toml 2>nul
del railway-build.sh 2>nul

echo.
echo [3/4] Going back to root and committing...
cd ..
git add web/package-lock.json
git add -A
git commit -m "Fix: Generate package-lock.json for Railway build"

echo.
echo [4/4] Pushing to GitHub...
git push origin main

echo.
echo ========================================
echo DONE! Railway will rebuild with package-lock.json
echo ========================================
pause
