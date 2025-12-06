# GitHub Repository Settings - Quick Setup

## Repository Description

```
Enterprise carbon accounting & ESG compliance platform with multi-cloud data ingestion, automated emissions calculations, and CSRD/SEC/CBAM reporting
```

## Topics (Tags)

Add these topics to your GitHub repository:

```
carbon-accounting
esg
csrd
sustainability
emissions
climate-tech
saas
golang
nextjs
typescript
compliance
sec-climate
cbam
ghg-protocol
scope3
multi-tenant
enterprise
production-ready
kubernetes
terraform
```

## How to Add (GitHub Web Interface)

1. Go to your repository on GitHub
2. Click **⚙️ Settings** (top navigation)
3. Scroll to **"About"** section (right sidebar on main repo page, or in Settings)
4. Click **⚙️ (gear icon)** next to "About"
5. Paste description in **"Description"** field
6. Add topics one by one in **"Topics"** field
7. Click **"Save changes"**

## Via GitHub API (Optional)

```bash
# Set description
curl -X PATCH \
  -H "Authorization: token YOUR_GITHUB_TOKEN" \
  -H "Accept: application/vnd.github.v3+json" \
  https://api.github.com/repos/YOUR_USERNAME/offgridflow \
  -d '{"description":"Enterprise carbon accounting & ESG compliance platform with multi-cloud data ingestion, automated emissions calculations, and CSRD/SEC/CBAM reporting"}'

# Set topics
curl -X PUT \
  -H "Authorization: token YOUR_GITHUB_TOKEN" \
  -H "Accept: application/vnd.github.mercy-preview+json" \
  https://api.github.com/repos/YOUR_USERNAME/offgridflow/topics \
  -d '{"names":["carbon-accounting","esg","csrd","sustainability","emissions","climate-tech","saas","golang","nextjs","typescript","compliance","sec-climate","cbam","ghg-protocol","scope3","multi-tenant","enterprise","production-ready","kubernetes","terraform"]}'
```

## Additional Settings (Recommended)

### Features to Enable
- ✅ Issues
- ✅ Projects (if using GitHub Projects)
- ✅ Discussions (for community Q&A)
- ✅ Actions (for CI/CD)

### Branch Protection (for main)
- ✅ Require pull request reviews before merging
- ✅ Require status checks to pass before merging
- ✅ Require branches to be up to date before merging

### Social Preview Image (Optional)
Create a 1280x640 PNG with:
- OffGridFlow logo
- "Enterprise Carbon Accounting Platform"
- Key feature highlights (CSRD, SEC, CBAM)

Upload via: Settings → Options → Social Preview → Upload

---

**Time Required**: 2-3 minutes  
**Impact**: +6% to Section 5 score  
**Status**: Ready to implement
