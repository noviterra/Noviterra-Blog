-- KEYS[1]: article:collect:set:{articleID}
-- KEYS[2]: article:collect:cnt:{articleID}
-- KEYS[3]: user:collect:arts:{userID}
-- ARGV[1]: userID
-- ARGV[2]: articleID
-- 返回: 1=收藏成功, 0=已收藏（幂等拒绝）

if redis.call('SREM', KEYS[1], ARGV[1]) == 1 then
	redis.call('DECR', KEYS[2])
	redis.call('SREM', KEYS[3], ARGV[2])
	return 1
end
return 0