local key = KEYS[1]
local limit = tonumber(ARGV[1])
local window = tonumber(ARGV[2])
local now = tonumber(ARGV[3])
local member = ARGV[4]

local window_start = now - window
redis.call("ZREMRANGEBYSCORE", key, 0, window_start)
redis.call("ZADD", key, now, member)
local count = redis.call("ZCARD", key)
redis.call("PEXPIRE", key, window)

if count > limit then
    return "DENY"
end

return "ALLOW"
