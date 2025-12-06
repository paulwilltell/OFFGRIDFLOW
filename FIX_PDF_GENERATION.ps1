# Quick fix for PDF generation
# Fixes the gofpdf Output method call in all 5 compliance files

Write-Host "🔧 Fixing PDF generation in compliance files..." -ForegroundColor Cyan
Write-Host ""

$files = @(
    "internal\compliance\csrd.go",
    "internal\compliance\sec.go",
    "internal\compliance\california.go",
    "internal\compliance\cbam.go",
    "internal\compliance\ifrs.go"
)

foreach ($file in $files) {
    $fullPath = "C:\Users\pault\OffGridFlow\$file"
    
    if (Test-Path $fullPath) {
        Write-Host "📝 Fixing $file..." -ForegroundColor Yellow
        
        # Read file
        $content = Get-Content $fullPath -Raw
        
        # Fix 1: Add bytes import if not present
        if ($content -notmatch 'import \(\s*"bytes"') {
            $content = $content -replace '(import \(\s+)"fmt"', '$1"bytes"`n`t"fmt"'
        }
        
        # Fix 2: Change Output call
        $content = $content -replace 'var buf \[\]byte\s+var err error\s+if buf, err = pdf\.Output\(&buf\); err != nil \{', 'var buf bytes.Buffer`n`tif err := pdf.Output(&buf); err != nil {'
        
        # Fix 3: Return bytes
        $content = $content -replace 'return buf, nil', 'return buf.Bytes(), nil'
        
        # Write back
        Set-Content -Path $fullPath -Value $content -NoNewline
        
        Write-Host "   ✅ Fixed" -ForegroundColor Green
    } else {
        Write-Host "   ❌ Not found: $file" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "🎉 All files fixed!" -ForegroundColor Green
Write-Host ""
Write-Host "Now run: .\EXECUTE_SECTION5.ps1" -ForegroundColor Cyan
