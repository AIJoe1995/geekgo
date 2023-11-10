local key = KEYS[1]
local topn = tonumber(ARGV[1])
local exists = redis.call("EXISTS", key)
if exists == 1 then
    -- 怎么返回一系列值
    return redis.call("ZRANGE", key, 0, topn-1, "REV")
    -- 返回的一系列bizId 还应该返回score
else
    -- 自增不成功
    return 0
end