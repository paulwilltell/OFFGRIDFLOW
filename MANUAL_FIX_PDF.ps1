# Manual fix - no regex, just clean file operations

Write-Host "🔧 Manually fixing all 5 compliance PDF files..." -ForegroundColor Cyan
Write-Host ""

$files = @("csrd", "sec", "california", "cbam", "ifrs")

foreach ($name in $files) {
    $file = "C:\Users\pault\OffGridFlow\internal\compliance\$name.go"
    
    Write-Host "📝 Fixing $name.go..." -ForegroundColor Yellow
    
    # Read entire file
    $content = Get-Content $file -Raw
    
    # Fix import block - remove the broken backtick-n-backtick-t
    $content = $content -replace 'import \(\s+"bytes"\`n\`t"fmt"', 'import (`n`t"bytes"`n`t"fmt"'
    
    # Also handle the other broken version
    $content = $content -replace '"bytes"``n``t"fmt"', '"bytes"`n`t"fmt"'
    
    # Fix the Output call - remove the broken backtick-n-backtick-t  
    $content = $content -replace 'var buf bytes\.Buffer``n``tif err :=', 'var buf bytes.Buffer`n`tif err :='
    
    # Write back
    $content | Set-Content $file -NoNewline
    
    Write-Host "   ✅ Fixed" -ForegroundColor Green
}

Write-Host ""
Write-Host "🎉 All files should be fixed!" -ForegroundColor Green
Write-Host ""
Write-Host "Try again: .\EXECUTE_SECTION5.ps1" -ForegroundColor Cyan
