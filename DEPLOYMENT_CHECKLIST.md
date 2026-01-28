# OffGridFlow Deployment Checklist
**Print this out or keep it open while deploying**

---

## ☐ PRE-DEPLOYMENT (10 minutes)

### Generate JWT Secret
```bash
openssl rand -base64 48
```
Copy the output → You'll paste this into Railway

### Open Railway Dashboard
https://railway.com/project/99b5cf9a-451d-47e5-be0f-fcb8eee95aff

---

## ☐ UPDATE ENVIRONMENT VARIABLES

### For `offgridflow-web` Service

Click on `offgridflow-web` → Settings → Variables

**DELETE THESE:**
- ☐ `OFFGRIDFLOW_DB_DSN` (the localhost one)
- ☐ `OFFGRIDFLOW_HTTP_PORT=3000` (wrong service)

**CHANGE THESE:**
- ☐ `NODE_ENV` → `production`
- ☐ `NEXT_PUBLIC_OFFGRIDFLOW_API_URL` → `https://offgridflow-api-production.up.railway.app`
- ☐ `NEXTAUTH_URL` → `https://off-grid-flow.com`

**ADD THESE IF MISSING:**
- ☐ `DATABASE_URL` → `${{Postgres.DATABASE_URL}}`

### For `offgridflow-api` Service

Click on `offgridflow-api` → Settings → Variables

**CRITICAL CHANGES:**
- ☐ `OFFGRIDFLOW_DB_DSN` → `${{Postgres.DATABASE_URL}}`
- ☐ `OFFGRIDFLOW_APP_ENV` → `production`
- ☐ `OFFGRIDFLOW_HTTP_PORT` → `8090`
- ☐ `PORT` → `8090`
- ☐ `OFFGRIDFLOW_JWT_SECRET` → *[paste your generated secret]*
- ☐ `OFFGRIDFLOW_COOKIE_SECURE` → `true`
- ☐ `OFFGRIDFLOW_COOKIE_DOMAIN` → `.off-grid-flow.com`
- ☐ `OFFGRIDFLOW_REQUIRE_AUTH` → `true`

**ADD IF MISSING:**
- ☐ `NEXTAUTH_URL` → `https://off-grid-flow.com`

---

## ☐ DEPLOY CODE (2 minutes)

### Option A: Automated (Recommended)
```powershell
cd C:\Users\pault\OffGridFlow
.\DEPLOY.ps1
```

### Option B: Manual
```bash
cd C:\Users\pault\OffGridFlow
git add .
git commit -m "Fix: Premium homepage + Railway production config"
git push origin main
```

---

## ☐ MONITOR DEPLOYMENT (5-10 minutes)

### Watch Railway Build Logs

**For offgridflow-web:**
1. ☐ Click on `offgridflow-web` service
2. ☐ Click "Deployments" tab
3. ☐ Watch for "Building..." → "Deploying..." → "Success"
4. ☐ Look for: `npm run build` success
5. ☐ Look for: `npm start` running on port 3000

**For offgridflow-api:**
1. ☐ Click on `offgridflow-api` service
2. ☐ Click "Deployments" tab
3. ☐ Watch for build completion
4. ☐ Look for: "Database connected successfully"
5. ☐ Look for: "HTTP server listening on :8090"

---

## ☐ TEST PRODUCTION (5 minutes)

### Test Homepage
1. ☐ Navigate to https://off-grid-flow.com
2. ☐ Verify premium glassmorphic design loads
3. ☐ Verify animated 3D globe is visible
4. ☐ Verify live statistics display (127.8M tCO₂e, etc.)
5. ☐ Verify feature cards appear
6. ☐ Check browser DevTools console for errors (should be none)

### Test API Health
1. ☐ Navigate to https://offgridflow-api-production.up.railway.app/health
2. ☐ Verify response: `{"status":"ok","timestamp":"...","service":"offgridflow-api"}`
3. ☐ Verify status code: 200 OK

### Test Registration
1. ☐ Click "Get Started" button on homepage
2. ☐ Fill in registration form:
   - First Name: Test
   - Last Name: User
   - Email: your-real-email@example.com
   - Company: Test Company
   - Password: TestPass123!
   - Confirm Password: TestPass123!
3. ☐ Click "Create account"
4. ☐ **VERIFY SUCCESS**: Should see "Check Your Email" page
5. ☐ **NOT AN ERROR**: Should NOT see "An unexpected error occurred"

### Check Database
1. ☐ Railway dashboard → Postgres service
2. ☐ Click "Data" tab
3. ☐ Find "users" table
4. ☐ Verify your test user was created

---

## ☐ TROUBLESHOOTING

### If Homepage Doesn't Load:
- Check Railway logs for Next.js build errors
- Verify `offgridflow-web` deployment succeeded
- Hard refresh browser (Ctrl+Shift+R)

### If Registration Fails:
- Check `offgridflow-api` logs for errors
- Verify database connection in logs
- Verify JWT_SECRET is set
- Verify OFFGRIDFLOW_DB_DSN points to `${{Postgres.DATABASE_URL}}`

### If Build Fails:
- Check Railway build logs for specific error
- Verify all environment variables are set
- Try redeploying previous working deployment
- Contact Claude for assistance

---

## ☐ POST-DEPLOYMENT SUCCESS CRITERIA

**All of these should be TRUE:**

- ☐ Homepage shows premium design (not basic placeholder)
- ☐ Globe animation is running smoothly
- ☐ Live statistics display correctly
- ☐ Registration form is accessible
- ☐ Test registration succeeds (email verification message shown)
- ☐ No console errors in browser DevTools
- ☐ No 500 errors in Railway API logs
- ☐ Health endpoint returns 200 OK
- ☐ Test user exists in Postgres database

**If ALL checkboxes above are checked → DEPLOYMENT SUCCESSFUL! 🎉**

---

## ☐ NEXT STEPS AFTER SUCCESS

1. ☐ Delete test user from database (if desired)
2. ☐ Configure email service for verification emails
3. ☐ Resume prospect email campaign (10 companies ready)
4. ☐ Monitor for first real customer signups
5. ☐ Set up Stripe billing (optional)
6. ☐ Celebrate! You just launched a premium SaaS platform 🚀

---

## 🆘 ROLLBACK PROCEDURE (If Deployment Fails)

### Via Railway Dashboard:
1. Go to failed service deployment
2. Click "Redeploy" on previous working deployment
3. Wait for rollback to complete

### Via Git:
```bash
git revert HEAD
git push origin main
```

---

## 📞 SUPPORT

**Documentation Created:**
- `DEPLOYMENT_FIX_GUIDE.md` - Full instructions
- `REPAIR_REPORT.md` - Technical details
- `.env.railway.web` - Environment template
- `.env.railway.api` - Environment template

**Railway Dashboard:**
https://railway.com/project/99b5cf9a-451d-47e5-be0f-fcb8eee95aff

**Need Help?**
Review DEPLOYMENT_FIX_GUIDE.md for detailed troubleshooting

---

**Last Updated**: January 27, 2026  
**Version**: 1.0 - Initial deployment fix
