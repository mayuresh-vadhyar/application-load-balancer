# Application Load Balancer

Lightweight HTTP load balancer implemented in Go. Routes incoming requests to a pool of backend servers using multiple balancing algorithms, performs health checks, exposes server metrics and a dashboard, and supports optional Redis-backed rate limiting and server-pool persistence.

**Features**
- **Load Balancing Algorithms:** Round Robin, Weighted Round Robin, IP-hash, URL-hash.
- **Health Checks:** Periodic upstream health checks with cooldowns and restart limits.
- **Server Management API:** Register, list and remove backend servers via `/server` (GET, POST, DELETE).
- **Monitoring:** Per-server request and health statistics via `/server/stats` and an HTML dashboard at `/server/dashboard`.
- **Proxying:** Proxies requests on `/` to chosen backend and injects `tracking-id` and `X-Forwarded-Server` headers.
- **Upstream Failure Handling:** Marks backends unhealthy when a proxy connection fails or the backend responds with `502 Bad Gateway`.
- **Request Logging:** Detailed logging with response time (TAT), status codes, response body size, and target server tracking.
- **Request Caching:** Redis-backed GET response caching is implemented; see the limitation below before relying on it.
- **Server Pool Persistence:** Saves server state to Redis on an interval and restores it at startup using the load balancer ID.
- **Rate Limiting (optional):** Redis-backed strategies like Fixed Window, Token Bucket, Leaky Bucket and Sliding Window (Lua script) with configurable request identification.
- **Config-driven:** Behavior controlled by `config.json`.
- **Graceful Shutdown:** Handles interrupt and termination signals, stops health checks, and closes HTTP and Redis resources.

**Quick Links**
- Config: [config.json](config.json)
- Rate limiter Lua scripts: [rateLimiter/token_bucket.lua](rateLimiter/token_bucket.lua), [rateLimiter/sliding_window.lua](rateLimiter/sliding_window.lua)

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
  - **`strategy`**: Rate limiting strategy. Valid values: `FW` (Fixed Window), `TB` (Token Bucket), `LB` (Leaky Bucket) `SW` (Sliding Window).
  - **`identifier`**: Identifier type for rate limiting: `IP`, `ApiKey`, `UserID`, `ApiPath`, or `Resource`.
  - **`limit`**: Request limit per window (e.g. `10`).
  - **`window`**: Window duration for rate limit (e.g. `1m`).
  - **`rate`**: Token generation rate for Token Bucket strategy (tokens per window).
- **`healthCheck`**: Health check configuration with sub-fields:
  - **`interval`**: Interval between health checks (e.g. `2s`).
  - **`maxUnhealthyChecks`**: Number of failed checks before marking server unhealthy.
  - **`cooldown`**: Duration to wait before retrying a failed health check (e.g. `5s`).
  - **`maxRestart`**: Maximum number of health check restart attempts before removing server.

Example `config.json` (minimal; use a Redis address to enable Redis-backed features):

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
  "redis": "localhost:6379",
  "rateLimit": { "enable": false }
}
```

**API**
- `GET /server` : List registered servers. Supports query params `isHealthy` and `urlParam`.
- `POST /server` : Register a new server. Body: `{ "url": "http://<host>:<port>", "weight": <int> }`.
- `DELETE /server` : Remove a server. Body: `{ "url": "http://<host>:<port>" }`.
- `GET /server/stats` : Return per-server ID, URL, health status, total request count, and active request count as JSON.
- `GET /server/dashboard` : Render an HTML table of per-server health and request statistics.
- `GET /` : Main proxy endpoint — forwards requests to upstream hosts.

Server management endpoints accept only their documented HTTP methods. `POST /server` expects a JSON body containing a `url`; weights are configured for initial backends in the `weights` array alongside `servers`.

Headers:
- `tracking-id`: added to both request and response to correlate proxied requests.
- `X-Forwarded-Server`: indicates the chosen upstream host.
- `X-Cache-Hit: true`: set on responses served from the Redis cache.

**Request Logging**
- Enabled by default; disable with `disableLogs` in config.
- Each request logs: HTTP method, path, target server, response status code, response time (TAT), and response body size.
- Can be disabled to reduce overhead in high-traffic scenarios.

**Request Caching**
- The code includes an optional Redis-backed GET response cache keyed by URL path and query string; only `200 OK` upstream response bodies are stored.
- Requests with `Cache-Control: no-cache` bypass the cache and are forwarded to the upstream. Cache hits include `X-Cache-Hit: true`.
- **Known limitation:** Request-cache initialization currently does not enable caching when `requestCacheExpiry` is a valid duration. The cache will remain inactive until this is corrected in the application code.

**Server Pool Persistence**
- When Redis is available and `serverPoolInterval` is positive, the current server state is periodically saved under a key associated with the load balancer `Id`.
- The saved pool expires according to `serverPoolExpiry` and is restored at startup when a matching pool exists.
- Persisted state includes backend URLs, health, weights, and request counters.

**Rate Limiting**
- Enabled when `config.json` provides a Redis URL and `rateLimit.enable` is `true`.
- **Available strategies:**
  - `FixedWindow` (`FW`): Simple counter + expiry in a fixed time window. Resets counter at the end of each window.
  - `TokenBucket` (`TB`): Token-based algorithm via Lua script in `rateLimiter/token_bucket.lua`. Allows burst traffic within token budget.
  - `LeakyBucket` (`LB`): Leaky bucket algorithm. Smooths out bursty traffic by processing requests at a constant drain rate.
  - `SlidingWindow` (`SW`): Precise sliding window counter via Lua script in `rateLimiter/sliding_window.lua`. Avoids boundary spikes of fixed windows.
- **Identifier support:** Controls what value is used as the rate limit key. Supported values: `IP` (client IP), `ApiKey` (`X-API-Key` header), `UserID` (`X-User-ID` header), `ApiPath` (request URL path), `Resource` (first path segment). Defaults to `IP` if not specified or invalid.
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
- Background routines perform periodic `HEAD` requests to upstream servers at the configured `healthCheck.interval`.
- Servers failing checks are marked unhealthy and skipped by load-balancing strategies; proxy connection failures and upstream `502` responses also mark a server unhealthy.
- After a server is marked unhealthy, it waits for `healthCheck.cooldown` before attempting recovery.
- Each failed recovery attempt is tracked; if `healthCheck.maxRestart` restarts are exhausted, the server is permanently removed from the pool.
- A server is marked unhealthy after `healthCheck.maxUnhealthyChecks` consecutive failed checks.

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
- `main.go` — HTTP server, routing, initialization, request caching, and graceful shutdown.
- `dashboard.go` — HTML dashboard for per-server health and request metrics.
- `LogResponseWriter.go` — Custom response writer for capturing and logging response metadata (status, size, timing).
- `loadBalancerStrategy/` — Load balancing algorithm implementations (Round Robin, Weighted, IP Hash, URL Hash).
- `server/` — Server registration, health checks, reverse proxy handling.
- `rateLimiter/` — Redis-based rate limiting strategies (Fixed Window, Token Bucket, Leaky Bucket, Sliding Window) and Lua scripts.
- `config/` — Configuration parsing from `config.json`.
- `Response/` — Helper functions for JSON response formatting (success/error responses).
- `Redis/` — Redis client initialization and connection management.
- `constants/` — Algorithm and strategy constants.

**Next steps & Notes**
- To enable rate limiting, set `rateLimit.enable` to `true` and provide a reachable `redis` address in `config.json`.
- Logs can be disabled with `disableLogs` in the config.
- Contributions welcome — open an issue or PR.

---