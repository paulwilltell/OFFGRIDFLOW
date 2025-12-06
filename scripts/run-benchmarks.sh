#!/bin/bash
# ========================================
# PERFORMANCE REGRESSION TEST RUNNER
# ========================================
# Runs Go benchmarks and compares against baseline

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
BASELINE_FILE="$PROJECT_ROOT/reports/performance-baseline.json"
RESULTS_FILE="$PROJECT_ROOT/reports/performance-results-$(date +%Y%m%d-%H%M%S).json"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo -e "${CYAN}🔬 OffGridFlow Performance Regression Test${NC}"
echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
echo ""

# Parse command line arguments
COMPARE=false
SAVE_BASELINE=false
VERBOSE=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --compare)
            COMPARE=true
            shift
            ;;
        --save-baseline)
            SAVE_BASELINE=true
            shift
            ;;
        --verbose)
            VERBOSE=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            echo "Usage: $0 [--compare] [--save-baseline] [--verbose]"
            exit 1
            ;;
    esac
done

cd "$PROJECT_ROOT"

# Create reports directory
mkdir -p reports

echo -e "${YELLOW}Running Go Benchmarks...${NC}"
echo ""

# Run benchmarks with increased iterations for accuracy
BENCH_OUTPUT=$(go test -bench=. -benchmem -benchtime=5s ./internal/performance/... 2>&1)

if [ $? -ne 0 ]; then
    echo -e "${RED}❌ Benchmarks failed to run${NC}"
    echo "$BENCH_OUTPUT"
    exit 1
fi

echo -e "${GREEN}✅ Benchmarks completed${NC}"
echo ""

# Parse benchmark results
echo "$BENCH_OUTPUT" | tee /tmp/benchmark-output.txt

# Extract key metrics
CACHE_OPS=$(echo "$BENCH_OUTPUT" | grep "BenchmarkCacheOperations" | awk '{print $3}')
QUERY_OPT=$(echo "$BENCH_OUTPUT" | grep "BenchmarkQueryOptimization" | awk '{print $3}')
LOAD_TEST=$(echo "$BENCH_OUTPUT" | grep "BenchmarkLoadTesterMetrics" | awk '{print $3}')

CACHE_ALLOCS=$(echo "$BENCH_OUTPUT" | grep "BenchmarkCacheOperations" | awk '{print $5}')
QUERY_ALLOCS=$(echo "$BENCH_OUTPUT" | grep "BenchmarkQueryOptimization" | awk '{print $5}')
LOAD_ALLOCS=$(echo "$BENCH_OUTPUT" | grep "BenchmarkLoadTesterMetrics" | awk '{print $5}')

# Create results JSON
cat > "$RESULTS_FILE" << EOF
{
  "timestamp": "$(date -Iseconds)",
  "git_commit": "$(git rev-parse HEAD 2>/dev/null || echo 'unknown')",
  "git_branch": "$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo 'unknown')",
  "benchmarks": {
    "cache_operations": {
      "ns_per_op": ${CACHE_OPS:-0},
      "allocs_per_op": ${CACHE_ALLOCS:-0}
    },
    "query_optimization": {
      "ns_per_op": ${QUERY_OPT:-0},
      "allocs_per_op": ${QUERY_ALLOCS:-0}
    },
    "load_tester_metrics": {
      "ns_per_op": ${LOAD_TEST:-0},
      "allocs_per_op": ${LOAD_ALLOCS:-0}
    }
  }
}
EOF

echo -e "${CYAN}📄 Results saved: $RESULTS_FILE${NC}"
echo ""

# Compare against baseline if requested
if [ "$COMPARE" = true ] && [ -f "$BASELINE_FILE" ]; then
    echo -e "${YELLOW}Comparing against baseline...${NC}"
    echo ""
    
    # Extract baseline values
    BASELINE_CACHE=$(jq -r '.benchmarks.cache_operations.ns_per_op' "$BASELINE_FILE")
    BASELINE_QUERY=$(jq -r '.benchmarks.query_optimization.ns_per_op' "$BASELINE_FILE")
    BASELINE_LOAD=$(jq -r '.benchmarks.load_tester_metrics.ns_per_op' "$BASELINE_FILE")
    
    # Calculate changes
    CACHE_CHANGE=$(echo "scale=2; (($CACHE_OPS - $BASELINE_CACHE) / $BASELINE_CACHE) * 100" | bc 2>/dev/null || echo "0")
    QUERY_CHANGE=$(echo "scale=2; (($QUERY_OPT - $BASELINE_QUERY) / $BASELINE_QUERY) * 100" | bc 2>/dev/null || echo "0")
    LOAD_CHANGE=$(echo "scale=2; (($LOAD_TEST - $BASELINE_LOAD) / $BASELINE_LOAD) * 100" | bc 2>/dev/null || echo "0")
    
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}📊 Regression Analysis${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    
    REGRESSION_FOUND=false
    THRESHOLD=20  # 20% degradation threshold
    
    # Cache operations
    printf "Cache Operations:       %10s ns/op " "$CACHE_OPS"
    if (( $(echo "$CACHE_CHANGE > $THRESHOLD" | bc -l) )); then
        echo -e "${RED}(+${CACHE_CHANGE}% ⚠️  REGRESSION)${NC}"
        REGRESSION_FOUND=true
    elif (( $(echo "$CACHE_CHANGE > 0" | bc -l) )); then
        echo -e "${YELLOW}(+${CACHE_CHANGE}%)${NC}"
    else
        echo -e "${GREEN}(${CACHE_CHANGE}% ✅)${NC}"
    fi
    
    # Query optimization
    printf "Query Optimization:     %10s ns/op " "$QUERY_OPT"
    if (( $(echo "$QUERY_CHANGE > $THRESHOLD" | bc -l) )); then
        echo -e "${RED}(+${QUERY_CHANGE}% ⚠️  REGRESSION)${NC}"
        REGRESSION_FOUND=true
    elif (( $(echo "$QUERY_CHANGE > 0" | bc -l) )); then
        echo -e "${YELLOW}(+${QUERY_CHANGE}%)${NC}"
    else
        echo -e "${GREEN}(${QUERY_CHANGE}% ✅)${NC}"
    fi
    
    # Load tester metrics
    printf "Load Tester Metrics:    %10s ns/op " "$LOAD_TEST"
    if (( $(echo "$LOAD_CHANGE > $THRESHOLD" | bc -l) )); then
        echo -e "${RED}(+${LOAD_CHANGE}% ⚠️  REGRESSION)${NC}"
        REGRESSION_FOUND=true
    elif (( $(echo "$LOAD_CHANGE > 0" | bc -l) )); then
        echo -e "${YELLOW}(+${LOAD_CHANGE}%)${NC}"
    else
        echo -e "${GREEN}(${LOAD_CHANGE}% ✅)${NC}"
    fi
    
    echo ""
    echo -e "${CYAN}Threshold: ${THRESHOLD}% degradation${NC}"
    echo ""
    
    if [ "$REGRESSION_FOUND" = true ]; then
        echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${RED}❌ PERFORMANCE REGRESSION DETECTED${NC}"
        echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo ""
        echo "One or more benchmarks show >$THRESHOLD% degradation."
        echo "Review changes and optimize before merging."
        echo ""
        exit 1
    else
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo -e "${GREEN}✅ NO REGRESSION DETECTED${NC}"
        echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
        echo ""
    fi
fi

# Save as new baseline if requested
if [ "$SAVE_BASELINE" = true ]; then
    cp "$RESULTS_FILE" "$BASELINE_FILE"
    echo -e "${GREEN}✅ Baseline updated: $BASELINE_FILE${NC}"
    echo ""
fi

# Verbose output
if [ "$VERBOSE" = true ]; then
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${CYAN}📋 Detailed Benchmark Output${NC}"
    echo -e "${CYAN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo ""
    echo "$BENCH_OUTPUT"
    echo ""
fi

echo -e "${GREEN}🎉 Performance regression test complete${NC}"
echo ""
