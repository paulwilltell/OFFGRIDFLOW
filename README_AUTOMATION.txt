OffGridFlow Automation Bootstrap

Files created:
- offgridflow-automation.ps1 : PowerShell scaffold to orchestrate Claude -> Codex -> CMS -> Ads -> Payment.

Prerequisites:
- PowerShell 7+
- Node/npm
- Codex CLI: npm install -g @openai/codex
- Claude API access: set CLAUDE_API_KEY and optionally CLAUDE_API_URL (defaults to https://api.anthropic.com/v1/complete)
  - Example (Windows PowerShell):
    - setx CLAUDE_API_KEY "your_key_here"
    - setx CLAUDE_API_URL "https://api.anthropic.com/v1/complete"
- Store other keys as env vars: CMS_API_USER, CMS_API_PASSWORD, CMS_BASE_URL, ADS_API_KEY, PAYMENT_API_KEY

How to use (example):
1. Open PowerShell 7+ as user with npm in PATH.
2. Ensure environment variables are set (see above).
3. .\offgridflow-automation.ps1 -StartPipeline

Notes & next steps:
- Implement real API calls in the placeholder functions (Publish-ToCMS, Push-Ads, GenerateContentWithCodex). Publish-ToCMS now creates drafts by default; change the 'status' parameter to 'publish' to auto-publish posts.
- Create a dedicated automation user (e.g., offgridflow-bot) and an Application Password; set env vars CMS_API_USER, CMS_API_PASSWORD, CMS_BASE_URL (e.g., https://off-grid-flow.com/wp-json/wp/v2).
- Review legal/privacy constraints before scraping or messaging users on social platforms.
- Choose which marketing task to prioritize so the scaffold can be extended (SEO/content, outreach, paid ads, or sales automation).
