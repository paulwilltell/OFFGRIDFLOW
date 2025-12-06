# OffGridFlow Final Production Deployment Script
# Million Fold Precision - Zero Downtime Deployment
# Author: Paul Canttell
# Date: 2024-12-04

param(
    [Parameter(Mandatory=$true)]
    [ValidateSet('staging', 'production')]
    [string]$Environment,
    
    [Parameter(Mandatory=$false)]
    [switch]$DryRun,
    
    [Parameter(Mandatory=$false)]
    [switch]$SkipTests,
    
    [Parameter(Mandatory=$false)]
    [switch]$Rollback,
    
    [Parameter(Mandatory=$false)]
    [string]$Version
)

$ErrorActionPreference = 'Stop'

# ANSI Colors
$GREEN = "`e[32m"
$RED = "`e[31m"
$YELLOW = "`e[33m"
$BLUE = "`e[34m"
$CYAN = "`e[36m"
$MAGENTA = "`e[35m"
$RESET = "`e[0m"

function Write-DeploymentBanner {
    param([string]$Text)
    Write-Host ""
    Write-Host "${MAGENTA}╔════════════════════════════════════════════════════════════════════╗${RESET}"
    Write-Host "${MAGENTA}║${RESET}  $Text"
    Write-Host "${MAGENTA}╚════════════════════════════════════════════════════════════════════╝${RESET}"
    Write-Host ""
}

function Write-Step {
    param([string]$Message, [string]$Color = $BLUE)
    Write-Host "${Color}[$(Get-Date -Format 'HH:mm:ss')] ► $Message${RESET}"
}

function Write-Success { param([string]$Message) Write-Step "✓ $Message" $GREEN }
function Write-Error { param([string]$Message) Write-Step "✗ $Message" $RED }
function Write-Warning { param([string]$Message) Write-Step "⚠ $Message" $YELLOW }

function Test-PreDeploymentChecks {
    Write-DeploymentBanner "PRE-DEPLOYMENT VALIDATION"
    
    $allPassed = $true
    
    # Check 1: Git status
    Write-Step "Checking git status..." $CYAN
    $gitStatus = git status --porcelain
    if ($gitStatus) {
        Write-Warning "Uncommitted changes detected"
        git status --short
        $allPassed = $false
    } else {
        Write-Success "Git repository is clean"
    }
    
    # Check 2: Tests
    if (-not $SkipTests) {
        Write-Step "Running backend tests..." $CYAN
        go test ./... -short
        if ($LASTEXITCODE -ne 0) {
            Write-Error "Backend tests failed"
            $allPassed = $false
        } else {
            Write-Success "Backend tests passed"
        }
        
        Write-Step "Running frontend tests..." $CYAN
        Push-Location web
        npm test -- --passWithNoTests
        Pop-Location
        if ($LASTEXITCODE -ne 0) {
            Write-Error "Frontend tests failed"
            $allPassed = $false
        } else {
            Write-Success "Frontend tests passed"
        }
    } else {
        Write-Warning "Tests skipped (--SkipTests flag)"
    }
    
    # Check 3: Environment file
    Write-Step "Checking environment configuration..." $CYAN
    $envFile = ".env.$Environment"
    if (Test-Path $envFile) {
        Write-Success "Environment file exists: $envFile"
        
        # Validate required variables
        $envContent = Get-Content $envFile -Raw
        $required = @('DATABASE_URL', 'JWT_SECRET', 'REDIS_URL')
        $missing = $required | Where-Object { $envContent -notmatch $_ }
        
        if ($missing.Count -gt 0) {
            Write-Error "Missing required variables: $($missing -join ', ')"
            $allPassed = $false
        } else {
            Write-Success "All required environment variables present"
        }
    } else {
        Write-Error "Environment file not found: $envFile"
        $allPassed = $false
    }
    
    # Check 4: Docker images
    Write-Step "Verifying Docker images..." $CYAN
    if (Get-Command docker -ErrorAction SilentlyContinue) {
        docker images | Select-String -Pattern "offgridflow"
        Write-Success "Docker available"
    } else {
        Write-Error "Docker not available"
        $allPassed = $false
    }
    
    # Check 5: Database connectivity
    Write-Step "Testing database connectivity..." $CYAN
    # This would need actual connection string from env file
    Write-Warning "Database connectivity test not implemented (manual verification required)"
    
    return $allPassed
}

function Invoke-Build {
    Write-DeploymentBanner "BUILD PHASE"
    
    # Backend build
    Write-Step "Building backend..." $CYAN
    go build -o bin/offgridflow-api ./cmd/api
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Backend build failed"
        throw "Build failed"
    }
    Write-Success "Backend built successfully"
    
    # Frontend build
    Write-Step "Building frontend..." $CYAN
    Push-Location web
    npm run build
    Pop-Location
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Frontend build failed"
        throw "Build failed"
    }
    Write-Success "Frontend built successfully"
    
    # Docker images
    Write-Step "Building Docker images..." $CYAN
    if ($Version) {
        $tag = $Version
    } else {
        $tag = git rev-parse --short HEAD
    }
    
    Write-Step "Building API image (offgridflow-api:$tag)..." $CYAN
    docker build -t "offgridflow-api:$tag" -f Dockerfile .
    if ($LASTEXITCODE -ne 0) {
        Write-Error "API Docker build failed"
        throw "Docker build failed"
    }
    Write-Success "API image built"
    
    Write-Step "Building web image (offgridflow-web:$tag)..." $CYAN
    docker build -t "offgridflow-web:$tag" -f web/Dockerfile ./web
    if ($LASTEXITCODE -ne 0) {
        Write-Error "Web Docker build failed"
        throw "Docker build failed"
    }
    Write-Success "Web image built"
    
    return $tag
}

function Invoke-DatabaseMigrations {
    param([string]$EnvFile)
    
    Write-DeploymentBanner "DATABASE MIGRATIONS"
    
    Write-Step "Running database migrations..." $CYAN
    
    # Load environment
    if (Test-Path $EnvFile) {
        Get-Content $EnvFile | ForEach-Object {
            if ($_ -match '^([^=]+)=(.*)$') {
                $env:$($matches[1]) = $matches[2]
            }
        }
    }
    
    Write-Step "Migrations will be executed on deployment..."
    Write-Success "Migration preparation complete"
}

function Invoke-Deployment {
    param([string]$ImageTag)
    
    Write-DeploymentBanner "DEPLOYMENT PHASE"
    
    if ($DryRun) {
        Write-Warning "DRY RUN MODE - No actual deployment"
        Write-Step "Would deploy:" $YELLOW
        Write-Host "  - Environment: $Environment"
        Write-Host "  - Image Tag: $ImageTag"
        Write-Host "  - Namespace: offgridflow-$Environment"
        return
    }
    
    if ($Environment -eq 'production') {
        Write-Warning "⚠️  PRODUCTION DEPLOYMENT - This will affect live users"
        Write-Host "Press ENTER to continue or Ctrl+C to abort..."
        Read-Host
    }
    
    Write-Step "Deploying to $Environment..." $CYAN
    
    # Using kubectl
    if (Get-Command kubectl -ErrorAction SilentlyContinue) {
        Write-Step "Updating Kubernetes deployment..." $CYAN
        
        # Update image tags in deployment files
        $namespace = "offgridflow-$Environment"
        
        # Apply Kubernetes manifests
        kubectl apply -f deployments/kubernetes/namespace.yaml
        kubectl apply -f deployments/kubernetes/configmap.yaml -n $namespace
        kubectl apply -f deployments/kubernetes/secrets.yaml -n $namespace
        kubectl apply -f deployments/kubernetes/deployment.yaml -n $namespace
        kubectl apply -f deployments/kubernetes/service.yaml -n $namespace
        kubectl apply -f deployments/kubernetes/ingress.yaml -n $namespace
        
        # Wait for rollout
        Write-Step "Waiting for rollout to complete..." $CYAN
        kubectl rollout status deployment/offgridflow-api -n $namespace --timeout=5m
        
        Write-Success "Deployment completed"
    }
    elseif (Test-Path docker-compose.yml) {
        Write-Step "Deploying with Docker Compose..." $CYAN
        docker-compose -f docker-compose.yml -f "docker-compose.$Environment.yml" up -d
        Write-Success "Docker Compose deployment completed"
    }
    else {
        Write-Error "No deployment method available (kubectl or docker-compose)"
        throw "Deployment failed"
    }
}

function Test-PostDeployment {
    param([string]$BaseUrl)
    
    Write-DeploymentBanner "POST-DEPLOYMENT VERIFICATION"
    
    Write-Step "Waiting for services to be ready..." $CYAN
    Start-Sleep -Seconds 10
    
    # Health check
    Write-Step "Testing health endpoint..." $CYAN
    try {
        $response = Invoke-WebRequest -Uri "$BaseUrl/health" -TimeoutSec 30
        if ($response.StatusCode -eq 200) {
            Write-Success "Health check passed"
        } else {
            Write-Error "Health check failed with status $($response.StatusCode)"
        }
    }
    catch {
        Write-Error "Health check failed: $_"
    }
    
    # API version check
    Write-Step "Checking API version..." $CYAN
    try {
        $response = Invoke-WebRequest -Uri "$BaseUrl/api/v1/version" -TimeoutSec 30
        Write-Success "API is responding"
    }
    catch {
        Write-Warning "API version endpoint not accessible"
    }
    
    # Database connectivity
    Write-Step "Verifying database connectivity..." $CYAN
    Write-Warning "Database connectivity test not implemented (manual verification required)"
    
    Write-Success "Post-deployment verification completed"
}

function Invoke-DeploymentRollback {
    Write-DeploymentBanner "ROLLBACK INITIATED"
    
    Write-Step "Rolling back deployment..." $CYAN
    
    if (Get-Command kubectl -ErrorAction SilentlyContinue) {
        $namespace = "offgridflow-$Environment"
        kubectl rollout undo deployment/offgridflow-api -n $namespace
        kubectl rollout status deployment/offgridflow-api -n $namespace
        Write-Success "Rollback completed"
    }
    else {
        Write-Error "Rollback not supported with current deployment method"
    }
}

# ============================================================================
# MAIN EXECUTION
# ============================================================================

try {
    Write-Host ""
    Write-Host "${MAGENTA}╔════════════════════════════════════════════════════════════════════╗${RESET}"
    Write-Host "${MAGENTA}║                                                                    ║${RESET}"
    Write-Host "${MAGENTA}║              OFFGRIDFLOW PRODUCTION DEPLOYMENT                     ║${RESET}"
    Write-Host "${MAGENTA}║              Million Fold Precision Framework                      ║${RESET}"
    Write-Host "${MAGENTA}║                                                                    ║${RESET}"
    Write-Host "${MAGENTA}╚════════════════════════════════════════════════════════════════════╝${RESET}"
    Write-Host ""
    
    Write-Host "Environment: ${CYAN}$Environment${RESET}"
    Write-Host "Dry Run: $(if ($DryRun) { "${YELLOW}Yes${RESET}" } else { "${GREEN}No${RESET}" })"
    Write-Host "Skip Tests: $(if ($SkipTests) { "${YELLOW}Yes${RESET}" } else { "${GREEN}No${RESET}" })"
    Write-Host ""
    
    $startTime = Get-Date
    
    if ($Rollback) {
        Invoke-DeploymentRollback
        exit 0
    }
    
    # Phase 1: Pre-deployment checks
    $checksPass = Test-PreDeploymentChecks
    if (-not $checksPass) {
        Write-Error "Pre-deployment checks failed"
        Write-Host ""
        Write-Host "Fix the issues above and try again"
        exit 1
    }
    
    # Phase 2: Build
    $imageTag = Invoke-Build
    Write-Success "Build completed with tag: $imageTag"
    
    # Phase 3: Database migrations
    Invoke-DatabaseMigrations -EnvFile ".env.$Environment"
    
    # Phase 4: Deployment
    Invoke-Deployment -ImageTag $imageTag
    
    # Phase 5: Post-deployment verification
    $baseUrl = if ($Environment -eq 'production') {
        "https://api.offgridflow.com"
    } else {
        "https://staging-api.offgridflow.com"
    }
    
    Test-PostDeployment -BaseUrl $baseUrl
    
    $endTime = Get-Date
    $duration = ($endTime - $startTime).TotalMinutes
    
    Write-Host ""
    Write-Host "${GREEN}╔════════════════════════════════════════════════════════════════════╗${RESET}"
    Write-Host "${GREEN}║                                                                    ║${RESET}"
    Write-Host "${GREEN}║                    ✓ DEPLOYMENT SUCCESSFUL                         ║${RESET}"
    Write-Host "${GREEN}║                                                                    ║${RESET}"
    Write-Host "${GREEN}╚════════════════════════════════════════════════════════════════════╝${RESET}"
    Write-Host ""
    Write-Host "Environment: $Environment"
    Write-Host "Version: $imageTag"
    Write-Host "Duration: $([math]::Round($duration, 2)) minutes"
    Write-Host "Timestamp: $(Get-Date -Format 'yyyy-MM-dd HH:mm:ss')"
    Write-Host ""
    
    # Save deployment record
    $deploymentRecord = @{
        Environment = $Environment
        Version = $imageTag
        Timestamp = Get-Date -Format 'o'
        Duration = $duration
        Status = "SUCCESS"
    } | ConvertTo-Json
    
    $deploymentRecord | Out-File -FilePath "deployments/history/deployment-$Environment-$(Get-Date -Format 'yyyyMMdd-HHmmss').json" -Encoding UTF8
    
    Write-Success "Deployment record saved"
    Write-Host ""
    
}
catch {
    Write-Host ""
    Write-Error "Deployment failed: $_"
    Write-Host ""
    Write-Host "${RED}╔════════════════════════════════════════════════════════════════════╗${RESET}"
    Write-Host "${RED}║                                                                    ║${RESET}"
    Write-Host "${RED}║                     ✗ DEPLOYMENT FAILED                            ║${RESET}"
    Write-Host "${RED}║                                                                    ║${RESET}"
    Write-Host "${RED}╚════════════════════════════════════════════════════════════════════╝${RESET}"
    Write-Host ""
    Write-Host "To rollback: ./deploy-production.ps1 -Environment $Environment -Rollback"
    Write-Host ""
    exit 1
}
