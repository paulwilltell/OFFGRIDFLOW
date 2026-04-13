# Railway Deployment Setup

## Web Service (off-grid-flow.com) — Should auto-deploy from push

The web service uses Root Directory `web` and builds with Nixpacks.
After pushing to main, Railway should automatically rebuild and deploy.

**Required env vars for web service:**
```
NEXT_PUBLIC_OFFGRIDFLOW_API_URL=https://<your-api-service>.up.railway.app
NODE_ENV=production
PORT=3000
```

## API Service — VERIFY THESE SETTINGS

The Go backend needs these Railway settings:

### Settings tab:
- **Root Directory**: MUST be blank/empty (NOT "web")
- **Builder**: Dockerfile
- **Dockerfile path**: `Dockerfile` (default)

### Variables tab (copy from .env.railway.api):
```
OFFGRIDFLOW_APP_ENV=production
OFFGRIDFLOW_HTTP_PORT=8090
PORT=8090
OFFGRIDFLOW_SERVICE_ROLE=api
OFFGRIDFLOW_DB_DSN=${{Postgres.DATABASE_URL}}
DATABASE_URL=${{Postgres.DATABASE_URL}}
OFFGRIDFLOW_JWT_SECRET=<generate 64-char random string>
OFFGRIDFLOW_COOKIE_SECURE=true
OFFGRIDFLOW_COOKIE_DOMAIN=.off-grid-flow.com
OFFGRIDFLOW_REQUIRE_AUTH=true
OFFGRIDFLOW_FRONTEND_URL=https://off-grid-flow.com
OFFGRIDFLOW_ALLOWED_ORIGINS=https://off-grid-flow.com
OFFGRIDFLOW_REQUIRE_EMAIL_VERIFICATION=false
```

**Note:** Set OFFGRIDFLOW_REQUIRE_EMAIL_VERIFICATION=false initially
so registration works without SMTP configured. Enable it after
setting up SendGrid or another email provider.

### Networking tab:
- Click "Generate Domain" if no domain exists
- Port: 8090

### After API deploys successfully:
1. Copy the API service's Railway domain (e.g., `offgridflow-api-production.up.railway.app`)
2. Go to the Web service → Variables
3. Set `NEXT_PUBLIC_OFFGRIDFLOW_API_URL=https://<api-domain>`
4. Redeploy the web service

### Quick health check:
```
curl https://<api-domain>/health
```
Should return: `{"status":"ok","timestamp":"...","service":"offgridflow-api"}`
