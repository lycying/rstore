package redisx

type Redis interface {
	Close() error

	EXISTS(key string) (string ,bool, error) //key_type , bool ,error
	GET(key string) (string, error)
	SET(key string, value string) error
	DEL(key []string) (int, error)
	INCRBY(key string, increment int64) (int64, error)
	EXPIRE(key string, expireSeconds int) (int64, error)
	EXPIREAT(key string, expireAt int) (int64, error)
	TTL(key string) (int64, error)
	TYPE(key string) (string, error)

	HSET(key string, field string, value string) (int, error)
	HMGET(key string, fields []string) (map[string]string, error)
	HMSET(key string, imap map[string]string) (string, error)
	HGET(key string, field string) (string, error)
	HGETALL(key string) (map[string]string, error)
	HDEL(key string, field string) (int, error)
	HLEN(key string) (int, error)
	HEXISTS(key string, field string) (bool, error)
	HKEYS(key string) ([]string, error)
	HVALS(key string) ([]string, error)
	HINCRBY(key string, field string, increment int64) (int64, error)

	ZADD(rkey string, score float64, member string) (int, error)
	ZSCORE(rkey string, member string) (string, error)
	ZREM(rkey string, member string) (int, error)
	ZREMRANGEBYSCORE(rkey string, min float64, max float64) (int, error)
	ZCARD(rkey string) (int, error)
	ZCOUNT(rkey string, min float64, max float64) (int, error)
	ZRANK(key string, member string) (int, error)
	ZRANGE(rkey string, begin int, end int, withscore bool) ([]string, error)
	ZRANGEBYSCORE(rkey string, min float64, max float64, withscore bool) ([]string, error)
	ZREVRANGEWITHSCORE(rkey string, start int, end int) ([]string, error)

	SADD(rkey string, members []string) (int, error)
	SCARD(rkey string) (int, error)
	SISMEMBER(rkey string, member string) (bool, error)
	SMEMBERS(rkey string) ([]string, error)
	SREM(rkey string, members []string) (int, error)
}
