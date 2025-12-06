# OffGridFlow Build Artifacts Documentation

**Last Updated**: December 5, 2025  
**Version**: 1.0.0

---

## Docker Images

### Production Images
- **offgridflow-api:latest** - Backend API service
- **offgridflow-worker:latest** - Background worker service  
- **offgridflow-web:latest** - Next.js frontend

### Image Registry
- Docker Hub: `paulcanttell/offgridflow-{service}:latest`
- GitHub Container Registry: `ghcr.io/paulcanttell/offgridflow-{service}:latest`

---

## Binary Names

### Linux Binaries
- `offgridflow-api` - API server (built from cmd/api/main.go)
- `offgridflow-worker` - Background worker (built from cmd/worker/main.go)

### Windows Binaries
- `offgridflow-api.exe`
- `offgridflow-worker.exe`

### macOS Binaries (Intel)
- `offgridflow-api-darwin-amd64`
- `offgridflow-worker-darwin-amd64`

### macOS Binaries (Apple Silicon)
- `offgridflow-api-darwin-arm64`
- `offgridflow-worker-darwin-arm64`

---

## Build Commands

### Backend Binaries

**Linux**:
```bash
# API
GOOS=linux GOARCH=amd64 go build -o offgridflow-api cmd/api/main.go

# Worker
GOOS=linux GOARCH=amd64 go build -o offgridflow-worker cmd/worker/main.go
```

**Windows**:
```powershell
# API
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -o offgridflow-api.exe cmd/api/main.go

# Worker
$env:GOOS="windows"; $env:GOARCH="amd64"; go build -o offgridflow-worker.exe cmd/worker/main.go
```

**macOS (Intel)**:
```bash
GOOS=darwin GOARCH=amd64 go build -o offgridflow-api-darwin-amd64 cmd/api/main.go
GOOS=darwin GOARCH=amd64 go build -o offgridflow-worker-darwin-amd64 cmd/worker/main.go
```

**macOS (Apple Silicon)**:
```bash
GOOS=darwin GOARCH=arm64 go build -o offgridflow-api-darwin-arm64 cmd/api/main.go
GOOS=darwin GOARCH=arm64 go build -o offgridflow-worker-darwin-arm64 cmd/worker/main.go
```

**All Platforms (using GoReleaser)**:
```bash
goreleaser build --snapshot --clean
```

---

### Docker Images

**Single Platform Build**:
```bash
# API
docker build -t offgridflow-api:latest -f Dockerfile .

# Web
docker build -t offgridflow-web:latest -f web/Dockerfile web/

# Worker (uses same Dockerfile as API, different entrypoint)
docker build -t offgridflow-worker:latest -f Dockerfile .
```

**Multi-Architecture Build** (amd64 + arm64):
```bash
# Setup buildx (first time only)
docker buildx create --name multiarch --use
docker buildx inspect --bootstrap

# Build and push multi-arch images
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t paulcanttell/offgridflow-api:latest \
  --push \
  .

docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t paulcanttell/offgridflow-web:latest \
  --push \
  web/
```

**Using Docker Compose**:
```bash
docker-compose build
```

---

## Artifact Locations

### Docker Hub
- https://hub.docker.com/r/paulcanttell/offgridflow-api
- https://hub.docker.com/r/paulcanttell/offgridflow-web
- https://hub.docker.com/r/paulcanttell/offgridflow-worker

### GitHub Container Registry (if using)
- ghcr.io/paulcanttell/offgridflow-api:latest
- ghcr.io/paulcanttell/offgridflow-web:latest
- ghcr.io/paulcanttell/offgridflow-worker:latest

### GitHub Releases
Binary releases published to:
- https://github.com/paulcanttell/offgridflow/releases

### Kubernetes Manifests
Located in: `infra/k8s/`
- `deployment.yaml` - Main deployments
- `service.yaml` - Service definitions
- `ingress.yaml` - Ingress configuration

---

## Version Tagging Strategy

### Semantic Versioning
Format: `vMAJOR.MINOR.PATCH`

Examples:
- `v1.0.0` - Initial release
- `v1.0.1` - Bug fix
- `v1.1.0` - New feature (backwards compatible)
- `v2.0.0` - Breaking changes

### Git Tags
```bash
# Create tag
git tag -a v1.0.0 -m "Release version 1.0.0"

# Push tag
git push origin v1.0.0
```

### Docker Tags
```bash
# Tag with version
docker tag offgridflow-api:latest offgridflow-api:v1.0.0
docker tag offgridflow-api:latest offgridflow-api:v1.0
docker tag offgridflow-api:latest offgridflow-api:v1

# Push all tags
docker push offgridflow-api:latest
docker push offgridflow-api:v1.0.0
docker push offgridflow-api:v1.0
docker push offgridflow-api:v1
```

---

## Build Sizes (Approximate)

### Binaries
- API: ~25 MB (compressed: ~8 MB with UPX)
- Worker: ~22 MB (compressed: ~7 MB with UPX)

### Docker Images
- offgridflow-api: ~50 MB (alpine-based)
- offgridflow-worker: ~50 MB (alpine-based)
- offgridflow-web: ~180 MB (node-based)

---

## CI/CD Integration

### GitHub Actions
Build artifacts automatically on:
- Push to `main` branch → Build and push `latest` tag
- Create Git tag → Build and push version tag
- Pull request → Build only (no push)

Workflow file: `.github/workflows/build.yml`

### Docker Registry Login
```bash
# Docker Hub
echo "$DOCKER_PASSWORD" | docker login -u "$DOCKER_USERNAME" --password-stdin

# GitHub Container Registry
echo "$GITHUB_TOKEN" | docker login ghcr.io -u "$GITHUB_ACTOR" --password-stdin
```

---

## Build Optimization

### Reduce Binary Size
```bash
# Strip debug symbols and compress
go build -ldflags="-s -w" -o offgridflow-api cmd/api/main.go
upx --best --lzma offgridflow-api
```

### Multi-Stage Docker Build
```dockerfile
# Already implemented in Dockerfile
FROM golang:1.21-alpine AS builder
# ... build stage ...

FROM alpine:latest
# ... minimal runtime ...
```

### Layer Caching
```bash
# Use BuildKit for better caching
DOCKER_BUILDKIT=1 docker build -t offgridflow-api:latest .
```

---

## Verification

### Test Binary
```bash
# Check version
./offgridflow-api --version

# Check help
./offgridflow-api --help

# Test startup
./offgridflow-api
```

### Test Docker Image
```bash
# Run container
docker run --rm offgridflow-api:latest --version

# Check image size
docker images offgridflow-api:latest

# Inspect layers
docker history offgridflow-api:latest
```

---

## Troubleshooting

### Build Fails
```bash
# Clean Go cache
go clean -cache -modcache -i -r

# Re-download dependencies
go mod download

# Rebuild
go build ./...
```

### Docker Build Fails
```bash
# Clean Docker cache
docker system prune -a

# Check Dockerfile syntax
docker build --no-cache -t offgridflow-api:latest .
```

### Multi-Arch Build Issues
```bash
# Reset buildx
docker buildx rm multiarch
docker buildx create --name multiarch --use
docker buildx inspect --bootstrap
```

---

## Release Checklist

Before releasing a new version:

- [ ] Update version in code
- [ ] Update CHANGELOG.md
- [ ] Run all tests: `go test ./...`
- [ ] Build all platforms
- [ ] Test binaries on each platform
- [ ] Create Git tag
- [ ] Build and push Docker images with version tags
- [ ] Update documentation
- [ ] Create GitHub release with binaries attached
- [ ] Announce release

---

**Maintained By**: Paul Canttell  
**Repository**: https://github.com/paulcanttell/offgridflow  
**Documentation**: https://docs.offgridflow.com
