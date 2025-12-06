# UI Screenshots Capture Guide

## Quick Capture Instructions

### Prerequisites
```powershell
cd C:\Users\pault\OffGridFlow
docker-compose up -d
```

Wait 30 seconds for startup, then open: **http://localhost:3000**

---

## Screenshots Needed (10 Total)

### 1. Login Page
**File**: `01-login.png`  
**URL**: http://localhost:3000/login  
**Capture**: Full login form with logo

**Press**: `Win + Shift + S` → Select area → Save as `01-login.png`

---

### 2. Dashboard
**File**: `02-dashboard.png`  
**URL**: http://localhost:3000/dashboard  
**Capture**: Main dashboard with emissions summary charts

**Key Elements**:
- Total emissions card
- Scope 1/2/3 breakdown
- Trend chart
- Recent activities

---

### 3. Activities List
**File**: `03-activities.png`  
**URL**: http://localhost:3000/activities  
**Capture**: Activities table showing emissions data

**Key Elements**:
- Activity name, scope, category
- Emissions values
- Date, location
- Action buttons (edit, delete)

---

### 4. Create Activity Form
**File**: `04-activity-form.png`  
**URL**: http://localhost:3000/activities/new  
**Capture**: New activity creation form

**Key Elements**:
- Activity name field
- Scope selector (1, 2, or 3)
- Category dropdown
- Amount/unit inputs
- Location field
- Save button

---

### 5. Emissions Summary
**File**: `05-emissions-summary.png`  
**URL**: http://localhost:3000/emissions  
**Capture**: Emissions summary page with charts

**Key Elements**:
- Scope 1/2/3 totals
- Pie chart or bar chart
- Category breakdown
- Export options

---

### 6. Compliance Reports List
**File**: `06-compliance-reports.png`  
**URL**: http://localhost:3000/compliance  
**Capture**: List of generated compliance reports

**Key Elements**:
- Report type (CSRD, SEC, etc.)
- Reporting year
- Status (Draft, Final)
- Generate/Download buttons

---

### 7. Sample Report Preview
**File**: `07-csrd-report.png`  
**Capture**: Preview of CSRD report (PDF viewer or preview pane)

**Alternative**: Screenshot the first page of a generated PDF

---

### 8. Settings Page
**File**: `08-settings.png`  
**URL**: http://localhost:3000/settings  
**Capture**: Settings/configuration page

**Key Elements**:
- Organization profile
- Integration settings
- User preferences
- Data sources

---

### 9. API Keys Management
**File**: `09-api-keys.png`  
**URL**: http://localhost:3000/settings/api-keys  
**Capture**: API key management interface

**Key Elements**:
- List of API keys
- Create new key button
- Key permissions
- Revoke option

---

### 10. Audit Log
**File**: `10-audit-log.png`  
**URL**: http://localhost:3000/audit  
**Capture**: Audit trail showing user actions

**Key Elements**:
- Timestamp
- User
- Action (login, create, export, etc.)
- Resource
- IP address

---

## Automated Capture (PowerShell)

```powershell
# Start application
cd C:\Users\pault\OffGridFlow
docker-compose up -d

# Wait for startup
Start-Sleep -Seconds 30

# Open browser to each page
$urls = @(
    "http://localhost:3000/login",
    "http://localhost:3000/dashboard",
    "http://localhost:3000/activities",
    "http://localhost:3000/activities/new",
    "http://localhost:3000/emissions",
    "http://localhost:3000/compliance",
    "http://localhost:3000/settings",
    "http://localhost:3000/settings/api-keys",
    "http://localhost:3000/audit"
)

foreach ($url in $urls) {
    Start-Process $url
    Write-Host "Opened: $url"
    Write-Host "Press Win+Shift+S to capture screenshot"
    Write-Host "Save to: docs\screenshots\"
    Write-Host ""
    Read-Host "Press Enter when ready for next page"
}
```

---

## Manual Capture Steps

1. **Start Application**:
   ```powershell
   docker-compose up -d
   ```

2. **Open Browser**: Navigate to http://localhost:3000

3. **For Each Screenshot**:
   - Navigate to the page
   - Press `Win + Shift + S` (Windows Snipping Tool)
   - Select area to capture
   - Click "Copy" or "Save"
   - Save to `C:\Users\pault\OffGridFlow\docs\screenshots\`
   - Name file according to list above

4. **Verify**:
   ```powershell
   Get-ChildItem C:\Users\pault\OffGridFlow\docs\screenshots
   ```

Should show 10 PNG files.

---

## Screenshot Requirements

- **Format**: PNG (preferred) or JPG
- **Resolution**: Minimum 1280x720 (HD)
- **Quality**: High (no compression artifacts)
- **Content**: Clear, readable text
- **Scope**: Capture relevant UI elements, crop unnecessary browser chrome

---

## Alternative: Use Browser DevTools

```javascript
// In browser console
function captureScreenshot(name) {
    html2canvas(document.body).then(canvas => {
        const link = document.createElement('a');
        link.download = name;
        link.href = canvas.toDataURL();
        link.click();
    });
}

// Usage
captureScreenshot('02-dashboard.png');
```

**Note**: Requires html2canvas library loaded

---

## Quick Verification

After capturing all screenshots:

```powershell
# Check count
$screenshots = Get-ChildItem "C:\Users\pault\OffGridFlow\docs\screenshots" -Filter "*.png"
Write-Host "Captured: $($screenshots.Count)/10 screenshots"

# List files
$screenshots | ForEach-Object {
    $sizeKB = [math]::Round($_.Length / 1KB, 1)
    Write-Host "  ✅ $($_.Name) - $sizeKB KB"
}
```

Expected output: 10 PNG files, ~50-500 KB each

---

## Adding to README

Once screenshots are captured, update `README.md`:

```markdown
## Screenshots

### Login
![Login Page](docs/screenshots/01-login.png)

### Dashboard
![Dashboard](docs/screenshots/02-dashboard.png)
*Real-time emissions tracking with scope breakdown*

### Activity Management
![Activities](docs/screenshots/03-activities.png)
*Track emissions across all operations*

### Compliance Reporting
![Reports](docs/screenshots/06-compliance-reports.png)
*Generate CSRD, SEC, CBAM, and California reports*

[View all screenshots →](docs/screenshots/)
```

---

## Estimated Time

- Setup: 2 minutes (start docker-compose)
- Capture 10 screenshots: 15-20 minutes
- Verify and organize: 3 minutes

**Total**: ~25 minutes

---

**Status**: Guide created  
**Location**: `C:\Users\pault\OffGridFlow\docs\screenshots\README.md`  
**Next**: Capture screenshots using Win+Shift+S
