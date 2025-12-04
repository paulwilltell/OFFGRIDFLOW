# SAP Integration Quick Start Guide

## Prerequisites
- SAP ERP or SAP S/4HANA system with API access
- OAuth2 client credentials from SAP administrator
- OffGridFlow application installed and running

## 5-Minute Setup

### Step 1: Get SAP Credentials
Contact your SAP administrator to obtain:
- API Base URL (e.g., `https://api.sap.yourcompany.com`)
- OAuth2 Client ID
- OAuth2 Client Secret
- Company Code (e.g., `1000`)

### Step 2: Configure Environment Variables
Add to your `.env` file or set in your environment:

```bash
export OFFGRIDFLOW_SAP_INGEST_ENABLED=true
export OFFGRIDFLOW_SAP_BASE_URL=https://api.sap.yourcompany.com
export OFFGRIDFLOW_SAP_CLIENT_ID=your-client-id
export OFFGRIDFLOW_SAP_CLIENT_SECRET=your-secret
export OFFGRIDFLOW_SAP_COMPANY=1000
export OFFGRIDFLOW_SAP_ORG_ID=your-org-id
```

Optional - filter by specific plant:
```bash
export OFFGRIDFLOW_SAP_PLANT=US-TX-001
```

### Step 3: Start OffGridFlow
```bash
./api
```

### Step 4: Verify Integration
Check logs for successful initialization:
```
[offgridflow] sap adapter initialized
```

### Step 5: Trigger Manual Ingestion (Optional)
Use the API or UI to trigger a manual data sync:
```bash
curl -X POST http://localhost:8090/api/v1/ingestion/run \
  -H "Content-Type: application/json" \
  -d '{"source": "sap"}'
```

## What Gets Ingested

### Energy Data
- ✅ Electricity consumption (kWh, MWh)
- ✅ Natural gas consumption (m³, GJ)
- ✅ Diesel and fuel oil (L, gallons)
- ✅ Steam and other utilities

### Emissions Data (if SAP Sustainability module available)
- ✅ Scope 1 emissions (direct)
- ✅ Scope 2 emissions (purchased electricity)
- ✅ Scope 3 emissions (indirect)

## Verification Checklist

- [ ] Environment variables configured
- [ ] SAP credentials verified with administrator
- [ ] OffGridFlow application starts without errors
- [ ] SAP adapter appears in startup logs
- [ ] Manual ingestion completes successfully
- [ ] Activities appear in database/dashboard
- [ ] Data looks correct (quantities, units, dates)

## Troubleshooting

### "authentication failed"
- Verify client ID and secret are correct
- Check OAuth endpoint URL is accessible
- Ensure client has necessary SAP permissions

### "No data ingested"
- Check date range includes actual data
- Verify company code is correct
- Confirm plant code filter (if used) matches SAP
- Review SAP API logs for access issues

### "Connection refused"
- Verify base URL is correct
- Check network connectivity to SAP
- Ensure firewall allows outbound HTTPS

## Next Steps

1. **Monitor Data Quality**: Review ingested data in OffGridFlow dashboard
2. **Set Up Scheduling**: Configure automatic daily/weekly ingestion
3. **Customize Mappings**: Update plant-to-location mappings for your organization
4. **Add Alerts**: Set up notifications for ingestion failures
5. **Scale Up**: Extend date range or add more plants

## Support

- 📖 Full Documentation: `internal/ingestion/sources/sap/README.md`
- 💻 Code Example: `internal/ingestion/sources/sap/example/main.go`
- 🧪 Tests: `internal/ingestion/sources/sap/sap_test.go`
- ⚙️ Config Template: `internal/ingestion/sources/sap/.env.example`

## Success Indicators

✅ Logs show "sap adapter initialized"  
✅ Activities created with source "sap_erp" or "sap_sustainability"  
✅ Metadata includes plant, meter, cost center  
✅ Quantities and units look reasonable  
✅ Emissions calculations using SAP data  

You're ready to track your organization's carbon footprint with SAP data! 🌱
