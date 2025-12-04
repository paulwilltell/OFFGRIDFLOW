## GitHub Copilot Chat

- Extension: 0.33.3 (prod)
- VS Code: 1.106.3 (bf9252a2fb45be6893dd8870c0bf37e2e1766d61)
- OS: win32 10.0.27975 x64
- GitHub Account: paulwilltell

## Network

User Settings:
```json
  "github.copilot.advanced.debug.useElectronFetcher": true,
  "github.copilot.advanced.debug.useNodeFetcher": false,
  "github.copilot.advanced.debug.useNodeFetchFetcher": true
```

Connecting to https://api.github.com:
- DNS ipv4 Lookup: 140.82.116.5 (11 ms)
- DNS ipv6 Lookup: Error (23 ms): getaddrinfo ENOTFOUND api.github.com
- Proxy URL: None (2 ms)
- Electron fetch (configured): HTTP 200 (137 ms)
- Node.js https: HTTP 200 (135 ms)
- Node.js fetch: HTTP 200 (250 ms)

Connecting to https://api.individual.githubcopilot.com/_ping:
- DNS ipv4 Lookup: 140.82.114.21 (5 ms)
- DNS ipv6 Lookup: Error (22 ms): getaddrinfo ENOTFOUND api.individual.githubcopilot.com
- Proxy URL: None (1 ms)
- Electron fetch (configured): HTTP 200 (93 ms)
- Node.js https: HTTP 200 (282 ms)
- Node.js fetch: HTTP 200 (295 ms)

Connecting to https://proxy.individual.githubcopilot.com/_ping:
- DNS ipv4 Lookup: 138.91.182.224 (44 ms)
- DNS ipv6 Lookup: Error (22 ms): getaddrinfo ENOTFOUND proxy.individual.githubcopilot.com
- Proxy URL: None (2 ms)
- Electron fetch (configured): HTTP 200 (194 ms)
- Node.js https: HTTP 200 (172 ms)
- Node.js fetch: HTTP 200 (195 ms)

Connecting to https://github.com: HTTP 200 (166 ms)
Connecting to https://telemetry.individual.githubcopilot.com/_ping: HTTP 200 (297 ms)

Number of system certificates: 32

## Documentation

In corporate networks: [Troubleshooting firewall settings for GitHub Copilot](https://docs.github.com/en/copilot/troubleshooting-github-copilot/troubleshooting-firewall-settings-for-github-copilot).