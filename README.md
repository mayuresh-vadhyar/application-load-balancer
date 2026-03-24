# Application Load Balancer

Lightweight HTTP load balancer implemented in Go. Routes incoming requests to a pool of backend servers using multiple balancing algorithms, performs health checks, and optionally enforces Redis-backed rate limiting.

**Features**
- **Load Balancing Algorithms:** Round Robin, Weighted Round Robin, IP-hash, URL-hash.
- **Health Checks:** Periodic upstream health checks with cooldowns and restart limits.
- **Server Management API:** Register, list and remove backend servers via `/server` (GET, POST, DELETE).
- **Proxying:** Proxies requests on `/` to chosen backend and injects `tracking-id` and `X-Forwarded-Server` headers.
- **Rate Limiting (optional):** Redis-backed strategies like Fixed Window and Token Bucket (Lua script).
- **Config-driven:** Behavior controlled by `config.json`.

**Configuration**
All runtime configuration lives in `config.json`. Key fields:

- **`algorithm`**: Balancing algorithm to use. Valid values: `RR`, `WRR`, `IPHash`, `UrlHash`.
- **`port`**: HTTP listen port (e.g. `":8080"`).
- **`disableLogs`**: Disable request logging when `true`.
- **`servers`**: Initial list of backend server URLs.
- **`weights`**: Parallel array of weights when using weighted round robin.
- **`rateLimit`**: Rate limiter config (enable, strategy, identifier, limit, window).
- **`healthCheck`**: Health check settings (interval, cooldown, maxUnhealthyChecks, maxRestart).
- **`redis`**: Redis address used for rate limiting (e.g. `127.0.0.1:6379`).

Example `config.json` (minimal):

```json
{
  "algorithm": "ROUND_ROBIN",
  "port": ":8080",
  "disableLogs": false,
  "servers": ["http://localhost:9001", "http://localhost:9002"],
  "rateLimit": { "enable": false }
}
```

**API**
- `GET /server` : List registered servers. Supports query params `isHealthy` and `urlParam`.
- `POST /server` : Register a new server. Body: `{ "url": "http://<host>:<port>", "weight": <int> }`.
- `DELETE /server` : Remove a server. Body: `{ "url": "http://<host>:<port>" }`.
- `GET /` : Main proxy endpoint — forwards requests to upstream hosts.

Headers:
- `tracking-id`: added to both request and response to correlate proxied requests.
- `X-Forwarded-Server`: indicates the chosen upstream host.

**Rate Limiting**
- Enabled when `config.json` provides a Redis URL and `rateLimit.enable` is `true`.
- Available strategies: `FixedWindow` (simple counter + expiry) and `TokenBucket` (Lua script in `rateLimiter/token_bucket.lua`).

**Health Checks**
- Background routines perform `HEAD` requests to upstreams at the configured `healthCheck.interval`.
- Servers failing checks are marked unhealthy; configurable cooldown and restart behavior govern removal and retries.

**Build & Run**
Requires Go (>=1.20) and, if using rate limiting, a running Redis instance.

Build:
```powershell
go build -o alb.exe .
```

Run (dev):
```powershell
go run main.go
```

There is also `go-start.bat` included for a quick start on Windows.
