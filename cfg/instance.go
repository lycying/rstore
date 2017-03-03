package cfg

import (
	"errors"
	"fmt"
	"github.com/lycying/rstore/redisx"
	"math/rand"
	"regexp"
)

type DB_Instance struct {
	Cfg     DBer
	Backend redisx.Redis
	State   int
}

type DBExt_Instance struct {
	DB         *DB_Instance
	IsMaster   bool
	ReadWeight int
}

type DBGroup_Instance struct {
	Cfg             *CfgDBGroup
	MasterSlaves    []*DBExt_Instance
	TotalReadWeight int //change this  when dbs changed
}

type Shard_Instance struct {
	Cfg        *CfgShard
	ShardParts []interface{}
}

type Rule_Instance struct {
	Cfg    *CfgRule
	Regexp *regexp.Regexp
}

type Instance struct {
	saver Saver //fetch data etc
	fly   Saver //make effect

	DBMap      map[string]map[string]*DB_Instance //type:name:instance
	DBGroupMap map[string]*DBGroup_Instance
	ShardMap   map[string]*Shard_Instance
	RuleMap    map[string]*Rule_Instance
}

func NewDBInstance(cfg DBer) *DB_Instance {
	db := &DB_Instance{}
	db.Cfg = cfg
	return db
}
func NewDBGroupInstance(cfg *CfgDBGroup) *DBGroup_Instance {
	db := &DBGroup_Instance{}
	db.Cfg = cfg
	db.MasterSlaves = make([]*DBExt_Instance, 0)
	return db
}
func NewShardInstance(cfg *CfgShard) *Shard_Instance {
	db := &Shard_Instance{}
	db.Cfg = cfg
	return db
}
func NewRuleInstance(cfg *CfgRule) *Rule_Instance {
	db := &Rule_Instance{}
	db.Cfg = cfg
	return db
}

func (shard *Shard_Instance) GetDBGroupInstance(hashKey string) *DBGroup_Instance {
	//TODO
	ise := GetInstance()
	name := shard.Cfg.ShardMap[0].RefName
	return ise.DBGroupMap[name]
}

func (inst *Instance) GetReadDB(cmd string, key string) (redisx.Redis, error) {
	var db *DBExt_Instance
	for _, v := range inst.RuleMap {
		sm := v.Regexp.FindSubmatch([]byte(key))
		if sm != nil {
			//yes , match
			cfg := v.Cfg
			shardName := cfg.ShardName
			hashKey := string(sm[cfg.HashSlot])
			dbShardInstance := inst.ShardMap[shardName]
			dbGroupInstance := dbShardInstance.GetDBGroupInstance(hashKey)

			var weight int
			rnd := rand.Intn(dbGroupInstance.TotalReadWeight)

			for _, tmp := range dbGroupInstance.MasterSlaves {
				if tmp.IsMaster && tmp.ReadWeight == 0 {
					continue
				}
				weight += tmp.ReadWeight
				if weight >= rnd {
					db = tmp
					break
				}
			}
			if db == nil {
				db = dbGroupInstance.MasterSlaves[rand.Intn(len(dbGroupInstance.MasterSlaves))]
			}
			break
		}
	}
	if db != nil {
		return db.DB.Backend, nil
	}
	return nil, errors.New(fmt.Sprintf("no router found for %v => %v", cmd, key))
}

func NewInstance(saver Saver, fly Saver) *Instance {
	ise := &Instance{}
	ise.DBMap = make(map[string]map[string]*DB_Instance)
	ise.DBGroupMap = make(map[string]*DBGroup_Instance)
	ise.ShardMap = make(map[string]*Shard_Instance)
	ise.RuleMap = make(map[string]*Rule_Instance)
	ise.saver = saver
	ise.fly = fly
	return ise
}

//init it from saver
func (ise *Instance) Init() {
	all_pg, err := ise.saver.GetAllPostgres()
	if err != nil {
		logger.Err(err,"error")
	} else {
		for _, v := range all_pg {
			ise.fly.SaveOrUpdatePostgres(v)
		}
	}

	all_mysql, err := ise.saver.GetAllMysql()
	if err != nil {
		logger.Err(err,"error")
	} else {
		for _, v := range all_mysql {
			ise.fly.SaveOrUpdateMySql(v)
		}
	}

	all_redis, err := ise.saver.GetAllRedis()
	if err != nil {
		logger.Err(err,"error")
	} else {
		for _, v := range all_redis {
			ise.fly.SaveOrUpdateRedis(v)
		}
	}

	all_dbgroup, err := ise.saver.GetAllDBGroup()
	if err != nil {
		logger.Err(err,"error")
	} else {
		for _, v := range all_dbgroup {
			ise.fly.SaveOrUpdateDBGroup(v)
		}
	}

	all_shard, err := ise.saver.GetAllShard()
	if err != nil {
		logger.Err(err,"error")
	} else {
		for _, v := range all_shard {
			ise.fly.SaveOrUpdateShard(v)
		}
	}

	all_rule, err := ise.saver.GetAllRule()
	if err != nil {
		logger.Err(err,"error")
	} else {
		for _, v := range all_rule {
			ise.fly.SaveOrUpdateRule(v)
		}
	}
}
