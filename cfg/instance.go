package cfg

import (
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
	DBMap      map[string]map[string]*DB_Instance //type:name:instance
	DBGroupMap map[string]map[string]*DBGroup_Instance
	ShardMap   map[string]*Shard_Instance
	RuleMap    map[string]*Rule_Instance
}

func NewDBInstance(cfg DBer) *DB_Instance {
	db := &DB_Instance{}
	db.Cfg = cfg
	return db
}
func NewDBGroupInstance(cfg *CfgDBGroup) *DBGroup_Instance{
	db := &DBGroup_Instance{}
	db.Cfg = cfg
	return db
}

func (shard *Shard_Instance) GetDBGroupInstance(hashKey string) *DBGroup_Instance {
	return nil
}

func (inst *Instance) GetReadDB(cmd string, key string) redisx.Redis {
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
				if db.IsMaster && db.ReadWeight == 0 {
					continue
				}
				weight += db.ReadWeight
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
		return db.DB.Backend
	}
	return nil
}

func NewInstance() *Instance {
	ise := &Instance{}
	ise.DBMap = make(map[string]map[string]*DB_Instance, 0)
	ise.DBGroupMap = make(map[string]map[string]*DBGroup_Instance, 0)
	ise.ShardMap = make(map[string]*Shard_Instance, 0)
	ise.RuleMap = make(map[string]*Rule_Instance, 0)
	return ise
}
