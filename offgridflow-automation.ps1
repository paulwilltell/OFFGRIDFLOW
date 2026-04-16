<#
OffGridFlow automation scaffold
Purpose: Bootstrap PowerShell orchestration connecting Claude, Codex, CMS, Ads, and payment APIs.
Prereqs: PowerShell 7+, Node/npm, Codex CLI, Claude access. Store API keys in env vars (CLAUDE_API_KEY, CODEX_API_KEY, CMS_API_KEY, ADS_API_KEY, PAYMENT_API_KEY).
#>

param(
    [switch]$StartPipeline,
    [switch]$DryRun
)

function Ensure-PowerShellVersion {
    if ($PSVersionTable.PSVersion.Major -lt 7) {
        Write-Error "PowerShell 7+ is required."
        return $false
    }
    return $true
}

function Install-CodexCli {
    Write-Host "Checking for codex CLI..."
    $found = Get-Command codex -ErrorAction SilentlyContinue
    if (-not $found) {
        Write-Host "Installing Codex CLI (requires npm)..."
        npm install -g @openai/codex
    }
}

function Configure-Claude {
    Write-Host "Ensure CLAUDE_API_KEY set in environment. This function is a placeholder for Claude integration (use Claude API/CLI)."
}

function Get-TopicsFromClaude {
    param(
        $seedKeywords,
        $numTopics = 10
    )

    $apiKey = $env:CLAUDE_API_KEY
    $apiUrl = $env:CLAUDE_API_URL

    if (-not $apiKey) {
        Write-Error "CLAUDE_API_KEY must be set in environment variables."
        return @()
    }

    if (-not $apiUrl) { $apiUrl = "https://api.anthropic.com/v1/complete" }

    $kw = $seedKeywords -join ", "

    $prompt = @"
You are an SEO research assistant. Given the seed keywords: $kw
Return a JSON array (only JSON) with $numTopics objects. Each object should have the fields:
  - title: short SEO-friendly title
  - brief: a 1-2 sentence brief describing the article angle
  - target_keywords: an array of 5 related keywords
  - priority: one of [high, medium, low]
Provide only the JSON array, no commentary.
"@

    $body = @{ model = "claude-2.1"; prompt = $prompt; max_tokens = 800; temperature = 0.2 } | ConvertTo-Json -Depth 6
    $headers = @{ "x-api-key" = $apiKey; "Content-Type" = "application/json" }

    try {
        $resp = Invoke-RestMethod -Uri $apiUrl -Method Post -Headers $headers -Body $body -ErrorAction Stop

        # Extract text from common response shapes
        $text = $null
        if ($resp.completion) { $text = $resp.completion }
        elseif ($resp.output) { $text = $resp.output }
        elseif ($resp.choices -and $resp.choices[0].text) { $text = $resp.choices[0].text }
        else { $text = $resp | Out-String }

        # Attempt to pull a JSON array from the response
        $start = $text.IndexOf('[')
        $end = $text.LastIndexOf(']')
        if ($start -ge 0 -and $end -gt $start) {
            $jsonText = $text.Substring($start, $end - $start + 1)
            try {
                $parsed = $jsonText | ConvertFrom-Json
                return $parsed
            } catch {
                Write-Warning "Failed to parse JSON from Claude response; returning raw text as a single item."
                return @($text)
            }
        } else {
            Write-Warning "No JSON array found in Claude response; returning raw text as a single item."
            return @($text)
        }
    } catch {
        Write-Error "Error calling Claude API: $_"
        return @()
    }
}

function GenerateContentWithCodex {
    param(
        [string]$brief,
        [string]$tone = "informative",
        [ValidateSet('short','medium','long')][string]$length = "medium"
    )

    Write-Host "Generating content for brief: $brief"

    if ($DryRun) { return "DRYRUN: Generated content for $brief" }

    $apiKey = $env:CODEX_API_KEY
    if (-not $apiKey) {
        Write-Error "CODEX_API_KEY must be set in environment variables."
        return $null
    }

    # Construct a clear prompt for the Codex/LLM
    $prompt = @"
You are an expert content writer skilled in SEO and audience tailoring.
Write a $length, $tone article based on the following brief:

$brief

Return the response as plain text. Include a one-line meta description separated by the line: ===META===
"@

    # Prefer the codex CLI if available
    $codexCmd = Get-Command codex -ErrorAction SilentlyContinue
    if ($codexCmd) {
        try {
            $args = @('--prompt', $prompt, '--max-tokens', '1200', '--temperature', '0.2')
            # Run the codex CLI and capture output
            $output = & $codexCmd.Source @args 2>&1
            if ($LASTEXITCODE -ne 0) {
                Write-Warning "codex CLI exited with code $LASTEXITCODE. Falling back to REST API. Output: $($output -join '\n')"
            } else {
                $text = $output -join "`n"
                return $text
            }
        } catch {
            Write-Warning "Error running codex CLI: $_. Falling back to REST API."
        }
    } else {
        Write-Warning "codex CLI not found in PATH — using REST fallback."
    }

    # Fallback: use OpenAI-style REST completion (requires CODEX_API_KEY to be a valid OpenAI key)
    try {
        $url = 'https://api.openai.com/v1/chat/completions'
        $headers = @{ Authorization = "Bearer $apiKey"; 'Content-Type' = 'application/json' }
        $messages = @(
            @{ role = 'system'; content = 'You are a helpful SEO content generator.' },
            @{ role = 'user'; content = $prompt }
        )
        $body = @{ model = 'gpt-4o-mini'; messages = $messages; max_tokens = 1200; temperature = 0.2 } | ConvertTo-Json -Depth 10
        $resp = Invoke-RestMethod -Uri $url -Method Post -Headers $headers -Body $body -ErrorAction Stop
        $text = $resp.choices[0].message.content
        return $text
    } catch {
        Write-Error "Failed to generate content via REST API: $_"
        return $null
    }
}

function Publish-ToCMS {
    param(
        $slug,
        $content,
        $title = $null,
        $status = "draft"
    )
    if (-not $title) { $title = ($slug -replace '-',' ') }

    $user = $env:CMS_API_USER
    $pass = $env:CMS_API_PASSWORD
    $base = $env:CMS_BASE_URL

    if (-not $user -or -not $pass -or -not $base) {
        Write-Error "CMS_API_USER, CMS_API_PASSWORD, and CMS_BASE_URL must be set as environment variables."
        return $null
    }

    $pair = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes("$user:$pass"))
    $headers = @{ Authorization = "Basic $pair"; "Content-Type" = "application/json" }

    $body = @{ title = $title; content = $content; status = $status; slug = $slug } | ConvertTo-Json -Depth 5

    try {
        $url = "$base/posts"
        $resp = Invoke-RestMethod -Uri $url -Method Post -Headers $headers -Body $body -ErrorAction Stop
        Write-Host "Created post (status: $status): ID=$($resp.id) Link=$($resp.link)"
        return $resp
    } catch {
        Write-Error "Failed to publish to CMS: $_"
        return $null
    }
}

function Push-Ads {
    param($audience,$creative)
    Write-Host "[Placeholder] Creating ad for audience $audience via Ads API (ADS_API_KEY)."
}

function Start-Pipeline {
    param($seedKeywords)
    Write-Host "Starting pipeline for: $seedKeywords"
    $topics = Get-TopicsFromClaude -seedKeywords $seedKeywords
    foreach ($t in $topics) {
        $content = GenerateContentWithCodex -brief $t
        Publish-ToCMS -slug ($t -replace ' ','-') -content $content
    }
    Push-Ads -audience "seed-audience" -creative "auto-generated"
}

# Entrypoint
if (-not (Ensure-PowerShellVersion)) { exit 1 }
Install-CodexCli
Configure-Claude

if ($StartPipeline) {
    Start-Pipeline -seedKeywords @('off-grid','battery storage')
}

Write-Host "Scaffold ready. Edit functions to integrate real API calls and secrets. See README_AUTOMATION.txt for next steps."