local key = KEYS[1]
local limit = tonumber(ARGV[1])
local rate = tonumber(ARGV[2])
local now = tonumber(ARGV[3])

print("BEGINNING", limit)
local data = redis.call("HMGET", key, "token_count", "last_refill")
local tokens = tonumber(data[1])
local last_refill = tonumber(data[2])

print("The token count is:", tokens)
if tokens == nil then
    -- initialize full bucket
    tokens = limit
    last_refill = now
end

local elapsed = (now - last_refill) / 1000
local new_tokens = tokens + (elapsed * rate)
print("New tokens will be:", new_tokens)
if (new_tokens > limit) then
    new_tokens = limit
end

local allowed = 0
if new_tokens >= 1 then
    allowed = 1
    new_tokens = new_tokens - 1
end

print("Before sett")
redis.call("HMSET", key, "token_count", new_tokens, "last_refill", now)
-- twice the time needed to fully refill
local ttl = math.floor((limit / rate) * 2)
if ttl > 0 then
    redis.call("EXPIRE", key, ttl)
end
