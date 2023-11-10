local zsKey = KEYS[1]
-- 对应 SortedSet Key
local zsMember = ARGV[1]
-- +1 或者 -1
local delta = tonumber(ARGV[2])
local exists = redis.call("EXISTS", zsKey)
if exists == 1 then
    redis.call("ZINCRBY", zsKey, delta, zsMember)
    return 1
else
    -- 自增不成功
    return 0
end