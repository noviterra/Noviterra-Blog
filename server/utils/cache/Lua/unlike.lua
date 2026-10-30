-- KEYS[1]: article:like:set:{articleID}
-- KEYS[2]: article:like:cnt:{articleID}
-- KEYS[3]: user:like:arts:{userID}
-- ARGV[1]: userID
-- ARGV[2]: articleID
-- 返回: 1=点赞成功, 0=已点过赞（幂等拒绝）

if redis.call('SREM', KEYS[1], ARGV[1]) == 1 then
	redis.call('DECR', KEYS[2])
	redis.call('SREM', KEYS[3], ARGV[2])
	return 1
end
return 0