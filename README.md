# Application Load Balancer

Lightweight HTTP load balancer implemented in Go. Routes incoming requests to a pool of backend servers using multiple balancing algorithms, performs health checks, and optionally enforces Redis-backed rate limiting.

**Features**
- **Load Balancing Algorithms:** Round Robin, Weighted Round Robin, IP-hash, URL-hash.
- **Health Checks:** Periodic upstream health checks with cooldowns and restart limits.
- **Server Management API:** Register, list and remove backend servers via `/server` (GET, POST, DELETE).
- **Proxying:** Proxies requests on `/` to chosen backend and injects `tracking-id` and `X-Forwarded-Server` headers.
- **Rate Limiting (optional):** Redis-backed strategies like Fixed Window and Token Bucket (Lua script).
- **Config-driven:** Behavior controlled by `config.json`.

**Quick Links**
- Config: [config.json](config.json)
- Rate limiter Lua script: [rateLimiter/token_bucket.lua](rateLimiter/token_bucket.lua)

**Configuration**
All runtime configuration lives in `config.json`. Key fields:

- **`Id`**: Unique identifier for the load balancer instance.
- **`algorithm`**: Balancing algorithm to use. Valid values: `RR` (Round Robin), `WRR` (Weighted Round Robin), `IPHash`, `UrlHash`.
- **`port`**: HTTP listen port (e.g. `":8080"`).
- **`disableLogs`**: Disable request logging when `true`.
- **`servers`**: Initial list of backend server URLs.
- **`weights`**: Parallel array of weights when using weighted round robin.
- **`serverPoolInterval`**: Server pool cache refreshing interval (e.g. `5s`).
- **`serverPoolExpiry`**: Server pool cache expiry duration (e.g. `10s`).
- **`requestCacheExpiry`**: API request Response cache expiry duration caching (e.g. `20s`). Requires Redis.
- **`redis`**: Redis address for rate limiting and response caching (e.g. `127.0.0.1:6379`).
- **`rateLimit`**: Rate limiter config with sub-fields:
  - **`enable`**: Enable/disable rate limiting.
  - **`strategy`**: Rate limiting strategy. Valid values: `FW` (Fixed Window), `TB` (Token Bucket).
  - **`identifier`**: Identifier type for rate limiting. Currently supported: `IP` (by client IP).
  - **`limit`**: Request limit per window (e.g. `10`).
  - **`window`**: Window duration for rate limit (e.g. `1m`).
  - **`rate`**: Token generation rate for Token Bucket strategy (tokens per window).
- **`healthCheck`**: Health check configuration with sub-fields:
  - **`interval`**: Interval between health checks (e.g. `2s`).
  - **`maxUnhealthyChecks`**: Number of failed checks before marking server unhealthy.
  - **`cooldown`**: Duration to wait before retrying a failed health check (e.g. `5s`).
  - **`maxRestart`**: Maximum number of health check restart attempts before removing server.

Example `config.json` (minimal):

```json
{
  "Id": "load-balancer-1",
  "algorithm": "RR",
  "port": ":8080",
  "disableLogs": false,
  "servers": ["http://localhost:9001", "http://localhost:9002"],
  "serverPoolInterval": "5s",
  "serverPoolExpiry": "10s",
  "requestCacheExpiry": "0s",
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
- **Available strategies:**
  - `FixedWindow`: Simple counter + expiry in a fixed time window. Resets counter at the end of each window.
  - `TokenBucket`: Token-based algorithm via Lua script in `rateLimiter/token_bucket.lua`. Allows burst traffic within token budget.
- **Identifier support:** Currently supports `IP` (rate limit by client IP address). Defaults to `IP` if not specified or invalid.
- **Configuration example:**
  ```json
  "rateLimit": {
    "enable": true,
    "strategy": "TB",
    "identifier": "IP",
    "limit": 10,
    "window": "1m",
    "rate": 10
  }
  ```

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

**Project Layout (high level)**
- `main.go` — HTTP server, routing, initialization.
- `LoadBalancingStrategy.go`, `RoundRobinStrategy.go`, `WeightedRoundRobinStrategy.go`, `IPHashStrategy.go`, `URLHashStrategy.go` — load balancing code.
- `server/` — server registration, health checks, reverse proxy handling.
- `rateLimiter/` — Redis-based rate limiting strategies and Lua script.
- `config/` — configuration parsing.
- `Response/` — helper responses for success/error.

**Next steps & Notes**
- To enable rate limiting, set `rateLimit.enable` to `true` and provide a reachable `redis` address in `config.json`.
- Logs can be disabled with `disableLogs` in the config.
- Contributions welcome — open an issue or PR.

---