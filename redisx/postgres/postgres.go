package postgres

import (
	"database/sql"
	_ "github.com/lib/pq"
	"github.com/lycying/rstore"
	"github.com/lycying/rstore/redisx"
	"math"
	"strconv"
	"strings"
	"time"
)

var (
	KV_UPDATE_SQL      = "insert into rstore_kv(\"rkey\",\"val\",\"lastTime\") values($1,$2,$3) on conflict(rkey) do update set \"val\"=$2,\"lastTime\"=$3"
	KV_QUERY_VALUE_SQL = "select val from rstore_kv where \"rkey\" = $1"

	HASH_UPDATE_SQL      = "insert into rstore_hash(\"rkey\",\"hkey\",\"val\",\"lastTime\") values($1,$2,$3,$4) on conflict(rkey,hkey) do update set \"val\"=$3,\"lastTime\"=$4"
	HASH_QUERY_VALUE_SQL = "select val from rstore_hash where \"rkey\" = $1 and \"hkey\" = $2 "
	HASH_QUERY_ALL_SQL   = "select hkey,val from rstore_hash where \"rkey\" = $1"
	HASH_DEL_HKEY_SQL    = "delete from rstore_hash where \"rkey\" = $1 and \"hkey\" = $2 "
	HASH_COUNT_SQL       = "select count(*) as num from rstore_hash where \"rkey\" = $1"

	ZSET_UPDATE_SQL           = "insert into rstore_zset(\"rkey\",\"member\",\"score\",\"lastTime\") values($1,$2,$3,$4) on conflict(rkey,member) do update set \"score\"=$3,\"lastTime\"=$4"
	ZSET_QUERY_ONE_SQL        = "select score from rstore_zset where \"rkey\" = $1 and \"member\" = $2 "
	ZSET_DEL_ALL_SQL          = "delete from rstore_zset where \"rkey\" = $1"
	ZSET_DEL_ONE_SQL          = "delete from rstore_zset where \"rkey\" = $1 and \"member\" = $2"
	ZSET_COUNT_SQL            = "select count(*) as num from rstore_zset where \"rkey\" = $1"
	ZSET_ZCOUNT_SQL           = "select count(*) as num from rstore_zset where \"rkey\" = $1 and score>=$2 and score<=$3"
	ZSET_REMZRANGEBYSCORE_SQL = "delete from rstore_zset where \"rkey\" = $1 and score>=$2 and score<=$3"
	ZSET_ZRANGEBYSCORE_SQL    = "select member,score from rstore_zset where \"rkey\" = $1 and score>=$2 and score<=$3 order by score asc,member asc"
	ZSET_ZRANGE_SQL           = "select member,score from rstore_zset where \"rkey\" = $1 order by score asc offset $2 limit $3 "
	ZSET_ZRANK_SQL            = "select rank from (select member,rank() over (order by \"score\" asc, \"lastTime\" asc) as rank from rstore_zset where \"rkey\" = $1 ) m where m.\"member\"= $2;"
)

type Postgres struct {
	redisx.Redis

	db *sql.DB
}

func NewPostgres(url string) (*Postgres, error) {
	pg := &Postgres{}
	err := pg.init(url)
	return pg, err
}

func (pg *Postgres) Close() error{
	return pg.db.Close()
}

func (pg *Postgres) GetReal() *sql.DB{
	return pg.db
}

func (pg *Postgres) init(url string) error {
	db, err := sql.Open("postgres", url)
	if err != nil {
		logger.Err(err, "error while create db")
		return err
	}
	pg.db = db

	//var tablename string
	//rows, err := db.Query("SELECT tablename from pg_tables ")
	//defer rows.Close()
	//
	//if err != nil {
	//	logger.Err(err, "error")
	//} else {
	//	for rows.Next() {
	//		rows.Scan(&tablename)
	//		logger.Debug("%+v", tablename)
	//	}
	//}
	return nil
}

func (pg *Postgres) timestamp() int64 {
	return time.Now().UnixNano() / 1e6
}

func (pg *Postgres) EXISTS(key string) (bool, error) {
	return false, nil
}
func (pg *Postgres) DEL(key string) error {
	return nil
}

func (pg *Postgres) EXPIRE(key string, expireSeconds int) (int64, error) {
	return 0, nil
}

func (pg *Postgres) EXPIREAT(key string, expireAt int) (int64, error) {
	return 0, nil
}

func (pg *Postgres) TTL(key string) (int64, error) {
	return 0, nil
}

func (pg *Postgres) TYPE(key string) (string, error) {
	return "", nil
}

func (pg *Postgres) GET(key string) (string, error) {
	var v string
	err := pg.db.QueryRow(KV_QUERY_VALUE_SQL, key).Scan(&v)
	if err != nil {
		if err == sql.ErrNoRows {
			v = ""
			err = rstore.KeyIsNilError
		}
	}
	return v, err
}

func (pg *Postgres) SET(key string, value string) error {
	_, err := pg.db.Exec(KV_UPDATE_SQL, key, value, pg.timestamp())

	if err != nil {
		return err
	}
	return nil
}

func (pg *Postgres) INCRBY(key string, increment int64) (int64, error) {
	tx, err := pg.db.Begin()
	if err != nil {
		return 0, err
	}
	var oldValue string
	err = tx.QueryRow(KV_QUERY_VALUE_SQL, key).Scan(&oldValue)
	if err != nil {
		if err == sql.ErrNoRows {
			//make a new one
			oldValue = "0"
		} else {
			return 0, err
		}
	}

	v, err := strconv.ParseInt(oldValue, 10, 64)
	if err != nil {
		return 0, rstore.KeyIsNotInteger
	}
	newValue := v + increment

	_, err = tx.Exec(KV_UPDATE_SQL, key, strconv.Itoa(int(newValue)), pg.timestamp())
	if err != nil {
		return 0, err
	}
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return newValue, nil
}

func (pg *Postgres) HSET(key string, field string, value string) (int, error) {
	rs, err := pg.db.Exec(HASH_UPDATE_SQL, key, field, value, pg.timestamp())

	if err != nil {
		return 0, err
	}

	af, err := rs.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(af), nil
}

func (pg *Postgres) HMSET(key string, rMap map[string]string) (string, error) {
	tx, err := pg.db.Begin()
	if err != nil {
		return "", err
	}
	lastTime := pg.timestamp()
	stmt, err := tx.Prepare(HASH_UPDATE_SQL)
	defer stmt.Close()
	for k, v := range rMap {
		stmt.Exec(key, k, v, lastTime)
	}
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return "", err
	}
	return "OK", nil
}

func (pg *Postgres) HGET(key string, field string) (string, error) {
	var v string
	err := pg.db.QueryRow(HASH_QUERY_VALUE_SQL, key, field).Scan(&v)
	if err != nil {
		if err == sql.ErrNoRows {
			v = ""
			err = rstore.KeyIsNilError
		} else {
			return "", err
		}
	}
	return v, err
}

func (pg *Postgres) HGETALL(key string) (map[string]string, error) {
	var hk string
	var hv string
	ret := make(map[string]string)

	rows, err := pg.db.Query(HASH_QUERY_ALL_SQL, key)
	if err != nil {
		if err == sql.ErrNoRows {
			//ignore
			err = nil
		} else {
			return ret, err
		}
	}
	for rows.Next() {
		err = rows.Scan(&hk, &hv)
		ret[hk] = hv
		if err != nil {
			return ret, err
		}
	}
	return ret, err
}

func (pg *Postgres) HMGET(key string, fields []string) (map[string]string, error) {
	inQ := strings.Join(fields, "','")
	SQL := "select hkey,val from rstore_hash where \"rkey\" = $1 and \"hkey\" in ('" + inQ + "' )"
	var hk string
	var hv string
	ret := make(map[string]string)

	rows, err := pg.db.Query(SQL, key)
	if err != nil {
		if err == sql.ErrNoRows {
			//ignore
			err = nil
		} else {
			return ret, err
		}
	}
	for rows.Next() {
		err = rows.Scan(&hk, &hv)
		ret[hk] = hv
		if err != nil {
			return ret, err
		}
	}
	return ret, err
}

func (pg *Postgres) HDEL(key string, field string) (int, error) {
	rs, err := pg.db.Exec(HASH_DEL_HKEY_SQL, key, field)
	if err != nil {
		return 0, err
	}
	af, err := rs.RowsAffected()
	return int(af), err
}

func (pg *Postgres) HLEN(key string) (int, error) {
	var n string
	rows, err := pg.db.Query(HASH_COUNT_SQL, key)
	if err != nil {
		if err == sql.ErrNoRows {
			//ignore
			err = nil
		} else {
			return 0, err
		}
	}
	for rows.Next() {
		err = rows.Scan(&n)
		if err != nil {
			return 0, err
		}
	}
	num, err := strconv.ParseInt(n, 10, 32)
	return int(num), err
}

func (pg *Postgres) HEXISTS(key string, field string) (bool, error) {
	_, err := pg.HGET(key, field)
	if err != nil {
		if err == rstore.KeyIsNilError {
			return false, nil
		} else {
			return false, err
		}
	}
	return true, nil
}

func (pg *Postgres) HKEYS(key string) ([]string, error) {
	kvs, err := pg.HGETALL(key)
	if err != nil {
		return []string{}, err
	}

	keys := make([]string, len(kvs))
	i := 0
	for k, _ := range kvs {
		keys[i] = k
		i++
	}
	return keys, nil
}

func (pg *Postgres) HVALS(key string) ([]string, error) {
	kvs, err := pg.HGETALL(key)
	if err != nil {
		return []string{}, err
	}

	values := make([]string, len(kvs))
	i := 0
	for _, v := range kvs {
		values[i] = v
		i++
	}
	return values, nil
}

func (pg *Postgres) HINCRBY(key string, field string, increment int64) (int64, error) {
	tx, err := pg.db.Begin()
	if err != nil {
		return 0, err
	}
	var oldValue string
	err = tx.QueryRow(HASH_QUERY_VALUE_SQL, key, field).Scan(&oldValue)
	if err != nil {
		if err == sql.ErrNoRows {
			//make a new one
			oldValue = "0"
		} else {
			return 0, err
		}
	}
	v, err := strconv.ParseInt(oldValue, 10, 64)
	newValue := v + increment
	if err != nil {
		return 0, rstore.KeyIsNotInteger
	}
	_, err = tx.Exec(HASH_UPDATE_SQL, key, field, strconv.Itoa(int(newValue)), pg.timestamp())
	if err != nil {
		return 0, err
	}
	err = tx.Commit()
	if err != nil {
		return 0, err
	}
	return newValue, nil
}

func (pg *Postgres) ZADD(rkey string, score float64, member string) (int, error) {
	rs, err := pg.db.Exec(ZSET_UPDATE_SQL, rkey, member, score, pg.timestamp())
	if err != nil {
		return 0, err
	}
	af, err := rs.RowsAffected()
	return int(af), nil
}

func (pg *Postgres) ZSCORE(rkey string, member string) (string, error) {
	var v string
	err := pg.db.QueryRow(ZSET_QUERY_ONE_SQL, rkey, member).Scan(&v)
	if err != nil {
		if err == sql.ErrNoRows {
			v = ""
			err = rstore.KeyIsNilError
		}
	}
	return v, err
}

func (pg *Postgres) ZREM(rkey string, member string) (int, error) {
	rs, err := pg.db.Exec(ZSET_DEL_ONE_SQL, rkey, member)
	if err != nil {
		return 0, err
	}
	af, err := rs.RowsAffected()
	return int(af), err
}

func (pg *Postgres) ZREMRANGEBYSCORE(rkey string, min float64, max float64) (int, error) {
	rs, err := pg.db.Exec(ZSET_REMZRANGEBYSCORE_SQL, rkey, min, max)
	if err != nil {
		return 0, err
	}
	af, err := rs.RowsAffected()
	return int(af), err
}

func (pg *Postgres) ZCARD(rkey string) (int, error) {
	var n string

	rows, err := pg.db.Query(ZSET_COUNT_SQL, rkey)
	if err != nil {
		if err == sql.ErrNoRows {
			//ignore
			err = nil
		} else {
			return 0, err
		}
	}
	for rows.Next() {
		err = rows.Scan(&n)
		if err != nil {
			return 0, err
		}
	}
	num, err := strconv.ParseInt(n, 10, 32)
	return int(num), err
}

func (pg *Postgres) ZCOUNT(rkey string, min float64, max float64) (int, error) {
	var n string

	rows, err := pg.db.Query(ZSET_ZCOUNT_SQL, rkey, min, max)
	if err != nil {
		if err == sql.ErrNoRows {
			//ignore
			err = nil
		} else {
			return 0, err
		}
	}
	for rows.Next() {
		err = rows.Scan(&n)
		if err != nil {
			return 0, err
		}
	}
	num, err := strconv.ParseInt(n, 10, 32)
	return int(num), err
}

func (pg *Postgres) ZRANK(key string, member string) (int, error) {
	var rank string

	rows, err := pg.db.Query(ZSET_ZRANK_SQL, key, member)
	if err != nil {
		if err == sql.ErrNoRows {
			//ignore
			err = nil
		}
		return -1, err
	}
	for rows.Next() {
		err = rows.Scan(&rank)
		if err != nil {
			return -1, err
		}
	}
	num, err := strconv.ParseInt(rank, 10, 32)
	if err != nil {
		return -1, nil
	}
	return int(num), err
}

func (pg *Postgres) ZRANGEBYSCORE(rkey string, min float64, max float64, withscore bool) ([]string, error) {
	var member string
	var score string
	ret := make([]string, 0)

	rows, err := pg.db.Query(ZSET_ZRANGEBYSCORE_SQL, rkey, min, max)
	if err != nil {
		if err == sql.ErrNoRows {
			//ignore
			err = nil
		} else {
			return ret, err
		}
	}
	for rows.Next() {
		err = rows.Scan(&member, &score)
		if err != nil {
			return ret, err
		}
		ret = append(ret, member)
		if withscore {
			ret = append(ret, score)
		}

	}
	return ret, err
}

func (pg *Postgres) ZRANGE(rkey string, begin int, end int, withscore bool) ([]string, error) {
	var member string
	var score string
	ret := make([]string, 0)

	limit := end - begin + 1
	if end == -1 {
		limit = math.MaxInt32
	}
	rows, err := pg.db.Query(ZSET_ZRANGE_SQL, rkey, begin, limit)
	if err != nil {
		if err == sql.ErrNoRows {
			//ignore
			err = nil
		} else {
			return ret, err
		}
	}
	for rows.Next() {
		err = rows.Scan(&member, &score)
		if err != nil {
			return ret, err
		}
		ret = append(ret, member)
		if withscore {
			ret = append(ret, score)
		}

	}
	return ret, err
}

func (pg *Postgres) ZREVRANGEWITHSCORE(rkey string, start int, end int) ([]string, error) {
	return nil, nil
}

func (pg *Postgres) SADD(rkey string, members []string) (int, error) {
	return 0, nil
}
func (pg *Postgres) SCARD(rkey string) (int, error) {
	return 0, nil
}
func (pg *Postgres) SISMEMBER(rkey string, member string) (bool, error) {
	return true, nil
}
func (pg *Postgres) SMEMBERS(rkey string) ([]string, error) {
	return nil, nil
}
func (pg *Postgres) SREM(rkey string, members []string) (int, error) {
	return 0, nil
}
