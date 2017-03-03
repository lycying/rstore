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

type ShardItem_Instance struct {
	Cfg 	*ShardItem
	Holder interface{}
}

type Shard_Instance struct {
	Cfg        *CfgShard
	ShardParts []*ShardItem_Instance
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

type Path struct {
	Rule *Rule_Instance
	Shard []*Shard_Instance
	Group *DBGroup_Instance
	DB *DBExt_Instance
	HashKey string
}

func NewPath() *Path{
	path := &Path{
		Shard: make([]*Shard_Instance,0),
	}
	return path
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
func NewShardItemInstance(cfg *ShardItem) *ShardItem_Instance{
	db := &ShardItem_Instance{}
	db.Cfg = cfg
	return db
}
func NewRuleInstance(cfg *CfgRule) *Rule_Instance {
	db := &Rule_Instance{}
	db.Cfg = cfg
	return db
}

func (shard *ShardItem_Instance) MatchItem(hashKey string) bool{
	//TODO
	return true
}

func (shard *Shard_Instance) GetDBGroupInstance(hashKey string,path *Path) (*DBGroup_Instance, error) {
	path.Shard = append(path.Shard,shard)

	for _,v := range shard.ShardParts{
		if v.MatchItem(hashKey){
			if v.Cfg.RefType == "shard"{
				return v.Holder.(*Shard_Instance).GetDBGroupInstance(hashKey,path)
			}else{
				return v.Holder.(*DBGroup_Instance),nil
			}
			break
		}
	}
	return nil,errors.New("no match dbgroup founded")
}


func (ise *Instance) GetReadDB(isReadCmd bool, key string) (*Path, error) {
	var p *Path = NewPath()
	var db *DBExt_Instance
	for _, rule := range ise.RuleMap {
		sm := rule.Regexp.FindStringSubmatch(key)
		if sm != nil {
			p.Rule = rule
			//yes , match
			if rule.Cfg.HashSlot >= len(sm) {
				return nil, errors.New(fmt.Sprintf("HashSlot is %v , but we only has %v submatch ", rule.Cfg.HashSlot, len(sm)))
			}
			hashKey := sm[rule.Cfg.HashSlot]
			p.HashKey = hashKey
			shardName := rule.Cfg.ShardName
			if dbShardInstance, ok := ise.ShardMap[shardName]; ok {
				dbGroupInstance, err := dbShardInstance.GetDBGroupInstance(hashKey,p)
				if err != nil {
					return p, err
				}
				p.Group = dbGroupInstance
				if isReadCmd {
					var weight int
					rnd := rand.Intn(dbGroupInstance.TotalReadWeight)
					for _, tmp := range dbGroupInstance.MasterSlaves {
						if tmp.ReadWeight == 0 {
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
				}else{
					masters := make([]*DBExt_Instance,0)
					for _, tmp := range dbGroupInstance.MasterSlaves {
						if tmp.IsMaster{
							masters = append(masters,tmp)
						}
					}
					if len(masters) <= 0 {
						return p,errors.New(fmt.Sprintf("Error : no master found in dbgroup %v", dbGroupInstance.Cfg.Name))
					}
					if dbGroupInstance.Cfg.ReplicateMode == "none"{
						//TODO should make raft and 2pc
						db = masters[0]
					}
				}
			}else{
				return p,errors.New(fmt.Sprintf("not found %v in shardmap",shardName))
			}
			//match once then break
			break
		}
	}
	if db != nil {
		p.DB = db
		return p, nil
	}
	return p, errors.New(fmt.Sprintf("no router found for  %v", key))
}

func NewInstance(saver Saver, fly Saver) *Instance {
	ise := &Instance{
		DBMap:      make(map[string]map[string]*DB_Instance),
		DBGroupMap: make(map[string]*DBGroup_Instance),
		ShardMap:   make(map[string]*Shard_Instance),
		RuleMap:    make(map[string]*Rule_Instance),
		saver:      saver,
		fly:        fly,
	}
	return ise
}

//init it from saver
func (ise *Instance) Init() {
	all_pg, err := ise.saver.GetAllPostgres()
	if err != nil {
		logger.Err(err, "error")
	} else {
		for _, v := range all_pg {
			ise.fly.SaveOrUpdatePostgres(v)
		}
	}

	all_mysql, err := ise.saver.GetAllMysql()
	if err != nil {
		logger.Err(err, "error")
	} else {
		for _, v := range all_mysql {
			ise.fly.SaveOrUpdateMySql(v)
		}
	}

	all_redis, err := ise.saver.GetAllRedis()
	if err != nil {
		logger.Err(err, "error")
	} else {
		for _, v := range all_redis {
			ise.fly.SaveOrUpdateRedis(v)
		}
	}

	all_dbgroup, err := ise.saver.GetAllDBGroup()
	if err != nil {
		logger.Err(err, "error")
	} else {
		for _, v := range all_dbgroup {
			ise.fly.SaveOrUpdateDBGroup(v)
		}
	}

	all_shard, err := ise.saver.GetAllShard()
	if err != nil {
		logger.Err(err, "error")
	} else {
		for _, v := range all_shard {
			ise.fly.SaveOrUpdateShard(v)
		}
	}

	all_rule, err := ise.saver.GetAllRule()
	if err != nil {
		logger.Err(err, "error")
	} else {
		for _, v := range all_rule {
			ise.fly.SaveOrUpdateRule(v)
		}
	}
}
