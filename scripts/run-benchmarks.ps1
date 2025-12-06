# ========================================
# PERFORMANCE REGRESSION TEST RUNNER (PowerShell)
# ========================================
# Runs Go benchmarks and compares against baseline

param(
    [switch]$Compare,
    [switch]$SaveBaseline,
    [switch]$Verbose
)

$BASELINE_FILE = "reports/performance-baseline.json"
$timestamp = Get-Date -Format "yyyyMMdd-HHmmss"
$RESULTS_FILE = "reports/performance-results-$timestamp.json"

Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host "🔬 OffGridFlow Performance Regression Test" -ForegroundColor Cyan
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
Write-Host ""

# Create reports directory
New-Item -ItemType Directory -Force -Path "reports" | Out-Null

Write-Host "Running Go Benchmarks..." -ForegroundColor Yellow
Write-Host ""

# Run benchmarks
$benchOutput = go test -bench=. -benchmem -benchtime=5s ./internal/performance/... 2>&1 | Out-String

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Benchmarks failed to run" -ForegroundColor Red
    Write-Host $benchOutput
    exit 1
}

Write-Host "✅ Benchmarks completed" -ForegroundColor Green
Write-Host ""

# Save benchmark output to temp file
$benchOutput | Out-File -FilePath "temp/benchmark-output.txt" -Encoding UTF8

# Parse benchmark results
$cacheOps = if ($benchOutput -match "BenchmarkCacheOperations-\d+\s+(\d+)") { $matches[1] } else { "0" }
$queryOpt = if ($benchOutput -match "BenchmarkQueryOptimization-\d+\s+(\d+)") { $matches[1] } else { "0" }
$loadTest = if ($benchOutput -match "BenchmarkLoadTesterMetrics-\d+\s+(\d+)") { $matches[1] } else { "0" }

# Extract allocations
$cacheAllocs = if ($benchOutput -match "BenchmarkCacheOperations.*?(\d+) B/op") { $matches[1] } else { "0" }
$queryAllocs = if ($benchOutput -match "BenchmarkQueryOptimization.*?(\d+) B/op") { $matches[1] } else { "0" }
$loadAllocs = if ($benchOutput -match "BenchmarkLoadTesterMetrics.*?(\d+) B/op") { $matches[1] } else { "0" }

# Get git info
try {
    $gitCommit = git rev-parse HEAD 2>$null
    $gitBranch = git rev-parse --abbrev-ref HEAD 2>$null
} catch {
    $gitCommit = "unknown"
    $gitBranch = "unknown"
}

# Create results JSON
$results = @{
    timestamp = (Get-Date -Format "o")
    git_commit = $gitCommit
    git_branch = $gitBranch
    benchmarks = @{
        cache_operations = @{
            ns_per_op = [int]$cacheOps
            allocs_per_op = [int]$cacheAllocs
        }
        query_optimization = @{
            ns_per_op = [int]$queryOpt
            allocs_per_op = [int]$queryAllocs
        }
        load_tester_metrics = @{
            ns_per_op = [int]$loadTest
            allocs_per_op = [int]$loadAllocs
        }
    }
}

$results | ConvertTo-Json -Depth 10 | Out-File -FilePath $RESULTS_FILE -Encoding UTF8

Write-Host "📄 Results saved: $RESULTS_FILE" -ForegroundColor Cyan
Write-Host ""

# Compare against baseline if requested
if ($Compare -and (Test-Path $BASELINE_FILE)) {
    Write-Host "Comparing against baseline..." -ForegroundColor Yellow
    Write-Host ""
    
    $baseline = Get-Content $BASELINE_FILE | ConvertFrom-Json
    
    $baselineCache = $baseline.benchmarks.cache_operations.ns_per_op
    $baselineQuery = $baseline.benchmarks.query_optimization.ns_per_op
    $baselineLoad = $baseline.benchmarks.load_tester_metrics.ns_per_op
    
    # Calculate changes
    $cacheChange = if ($baselineCache -gt 0) { (([int]$cacheOps - $baselineCache) / $baselineCache) * 100 } else { 0 }
    $queryChange = if ($baselineQuery -gt 0) { (([int]$queryOpt - $baselineQuery) / $baselineQuery) * 100 } else { 0 }
    $loadChange = if ($baselineLoad -gt 0) { (([int]$loadTest - $baselineLoad) / $baselineLoad) * 100 } else { 0 }
    
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    Write-Host "📊 Regression Analysis" -ForegroundColor Cyan
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    Write-Host ""
    
    $regressionFound = $false
    $threshold = 20  # 20% degradation threshold
    
    # Cache operations
    $cacheFormatted = "{0,10}" -f $cacheOps
    Write-Host "Cache Operations:       $cacheFormatted ns/op " -NoNewline
    if ($cacheChange -gt $threshold) {
        Write-Host "(+$([math]::Round($cacheChange, 2))% ⚠️  REGRESSION)" -ForegroundColor Red
        $regressionFound = $true
    } elseif ($cacheChange -gt 0) {
        Write-Host "(+$([math]::Round($cacheChange, 2))%)" -ForegroundColor Yellow
    } else {
        Write-Host "($([math]::Round($cacheChange, 2))% ✅)" -ForegroundColor Green
    }
    
    # Query optimization
    $queryFormatted = "{0,10}" -f $queryOpt
    Write-Host "Query Optimization:     $queryFormatted ns/op " -NoNewline
    if ($queryChange -gt $threshold) {
        Write-Host "(+$([math]::Round($queryChange, 2))% ⚠️  REGRESSION)" -ForegroundColor Red
        $regressionFound = $true
    } elseif ($queryChange -gt 0) {
        Write-Host "(+$([math]::Round($queryChange, 2))%)" -ForegroundColor Yellow
    } else {
        Write-Host "($([math]::Round($queryChange, 2))% ✅)" -ForegroundColor Green
    }
    
    # Load tester metrics
    $loadFormatted = "{0,10}" -f $loadTest
    Write-Host "Load Tester Metrics:    $loadFormatted ns/op " -NoNewline
    if ($loadChange -gt $threshold) {
        Write-Host "(+$([math]::Round($loadChange, 2))% ⚠️  REGRESSION)" -ForegroundColor Red
        $regressionFound = $true
    } elseif ($loadChange -gt 0) {
        Write-Host "(+$([math]::Round($loadChange, 2))%)" -ForegroundColor Yellow
    } else {
        Write-Host "($([math]::Round($loadChange, 2))% ✅)" -ForegroundColor Green
    }
    
    Write-Host ""
    Write-Host "Threshold: $threshold% degradation" -ForegroundColor Cyan
    Write-Host ""
    
    if ($regressionFound) {
        Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Red
        Write-Host "❌ PERFORMANCE REGRESSION DETECTED" -ForegroundColor Red
        Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Red
        Write-Host ""
        Write-Host "One or more benchmarks show >$threshold% degradation."
        Write-Host "Review changes and optimize before merging."
        Write-Host ""
        exit 1
    } else {
        Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Green
        Write-Host "✅ NO REGRESSION DETECTED" -ForegroundColor Green
        Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Green
        Write-Host ""
    }
}

# Save as new baseline if requested
if ($SaveBaseline) {
    Copy-Item $RESULTS_FILE $BASELINE_FILE -Force
    Write-Host "✅ Baseline updated: $BASELINE_FILE" -ForegroundColor Green
    Write-Host ""
}

# Verbose output
if ($Verbose) {
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    Write-Host "📋 Detailed Benchmark Output" -ForegroundColor Cyan
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Cyan
    Write-Host ""
    Write-Host $benchOutput
    Write-Host ""
}

Write-Host "🎉 Performance regression test complete" -ForegroundColor Green
Write-Host ""
