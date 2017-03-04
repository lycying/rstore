package cfg

import (
	"container/list"
	"errors"
	"fmt"
	"github.com/lycying/rstore/redisx"
	"hash/crc32"
	"math/rand"
	"regexp"
	"strconv"
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

type ShardType_Range_Instance struct {
	Start int64
	End   int64
}

type ShardType_ModHash_Instance struct {
	HashSeq int
}

type ShardItem_Instance struct {
	Cfg    *ShardItem
	Holder interface{}
	ShardPartInstance interface{}
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
	Rule    *Rule_Instance
	Shard   []*Shard_Instance
	Group   *DBGroup_Instance
	DBs     []*DBExt_Instance
	HashKey string
}

func NewPath() *Path {
	path := &Path{
		Shard: make([]*Shard_Instance, 0),
		DBs:   make([]*DBExt_Instance, 0),
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
func NewShardItemInstance(cfg *ShardItem) *ShardItem_Instance {
	db := &ShardItem_Instance{}
	db.Cfg = cfg
	return db
}
func NewRuleInstance(cfg *CfgRule) *Rule_Instance {
	db := &Rule_Instance{}
	db.Cfg = cfg
	return db
}

func (shard *Shard_Instance) GetDBGroupInstance(hashKey string, path *Path) (*DBGroup_Instance, error) {
	path.Shard = append(path.Shard, shard)

	if len(shard.Cfg.ShardMap) <= 0 {
		return nil, errors.New(fmt.Sprintf("shard (%v) has no any shard information", shard.Cfg.Name))
	}

	var aim *ShardItem_Instance = nil

	if shard.Cfg.ShardType == Shard_ShardType_Mod || shard.Cfg.ShardType == Shard_ShardType_Hash {
		var err error
		int64hash := int64(0)
		//mod assume the hashkey is a number , if not , return error
		if shard.Cfg.ShardType == Shard_ShardType_Mod {
			int64hash, err = strconv.ParseInt(hashKey, 10, 64)
		} else {
			//but the hash method use the sum32 to 	compute a number first ,it showed no error
			ie := crc32.NewIEEE()
			ie.Write([]byte(hashKey))
			int64hash = int64(ie.Sum32())
		}
		if err != nil {
			return nil, errors.New(fmt.Sprintf("can not parse hashkey %v to int ",hashKey))
		}

		length := int64(len(shard.Cfg.ShardMap))
		slot := int(int64hash % length)

		//loop to find it
		for _, v := range shard.ShardParts {
			if v.ShardPartInstance.(*ShardType_ModHash_Instance).HashSeq == slot {
				aim = v
				break
			}
		}
	} else if shard.Cfg.ShardType == Shard_ShardType_Range {
		int64hash, err := strconv.ParseInt(hashKey, 10, 64)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("can not parse hashkey %v to int ",hashKey))
		}
		for _, v := range shard.ShardParts {
			part := v.ShardPartInstance.(*ShardType_Range_Instance)
			if int64hash > part.Start && int64hash <= part.End {
				aim = v
				break
			}
		}
	}

	if aim == nil {
		return nil, errors.New(fmt.Sprintf("not match any mod . hashkey = %v,shardName = %v", hashKey, shard.Cfg.Name))
	}

	if aim.Cfg.RefType == Shard_RefType_Shard {
		//loop to find it
		return aim.Holder.(*Shard_Instance).GetDBGroupInstance(hashKey, path)
	} else if aim.Cfg.RefType == Shard_RefType_DBGroup {
		return aim.Holder.(*DBGroup_Instance), nil
	}

	return nil, errors.New("no match dbgroup founded")
}

func (ise *Instance) GetReadDB(isReadCmd bool, key string) (*Path, error) {
	var p *Path = NewPath()
	for _, rule := range ise.RuleMap {
		sm := rule.Regexp.FindStringSubmatch(key)
		if sm != nil {
			p.Rule = rule
			//yes , match
			if rule.Cfg.HashSlot >= len(sm) {
				return p, errors.New(fmt.Sprintf("HashSlot is %v , but we only has %v submatch ", rule.Cfg.HashSlot, len(sm)))
			}
			hashKey := sm[rule.Cfg.HashSlot]
			p.HashKey = hashKey
			shardName := rule.Cfg.ShardName
			// can find the shard
			if dbShardInstance, ok := ise.ShardMap[shardName]; ok {
				dbGroupInstance, err := dbShardInstance.GetDBGroupInstance(hashKey, p)
				//the error has translate
				if err != nil {
					return p, err
				}
				p.Group = dbGroupInstance
				if dbGroupInstance.TotalReadWeight <= 0 {
					return p, errors.New("no readable db found , you may change the readweight of the dbgroup")
				}
				if isReadCmd {
					msLen := len(dbGroupInstance.MasterSlaves)
					//only has one
					if msLen == 1 {
						p.DBs = append(p.DBs, dbGroupInstance.MasterSlaves[0])
						break
					}

					var weight int
					rnd := rand.Intn(dbGroupInstance.TotalReadWeight)
					for _, tmp := range dbGroupInstance.MasterSlaves {
						//do not use any unit if the read weight is zero
						if tmp.ReadWeight <= 0 {
							continue
						}
						weight += tmp.ReadWeight
						if weight >= rnd {
							p.DBs = append(p.DBs, tmp)
							break
						}
					}
				} else {
					masters := make([]*DBExt_Instance, 0)
					for _, tmp := range dbGroupInstance.MasterSlaves {
						if tmp.IsMaster {
							masters = append(masters, tmp)
						}
					}
					if len(masters) <= 0 {
						return p, errors.New(fmt.Sprintf("Error : no master found in dbgroup %v", dbGroupInstance.Cfg.Name))
					}
					if dbGroupInstance.Cfg.ReplicateMode == DBGroup_ReplicateMode_Writeone {
						//TODO should make raft and 2pc
						p.DBs = append(p.DBs, masters[0])
					} else if dbGroupInstance.Cfg.ReplicateMode == DBGroup_ReplicateMode_Writeall {
						p.DBs = append(p.DBs, masters[0])
						//TODO
					} else if dbGroupInstance.Cfg.ReplicateMode == DBGroup_ReplicateMode_Discard {
						p.DBs = append(p.DBs, masters[0])
						//TODO
					}
				}
			} else { //can not find the shard
				return p, errors.New(fmt.Sprintf("not found %v in shardmap", shardName))
			}
			//match once then break
			break
		}
	}
	if len(p.DBs) > 0 {
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

	loopList := list.New()
	all_shard, err := ise.saver.GetAllShard()
	if err != nil {
		logger.Err(err, "error")
	} else {
		for _, v := range all_shard {
			err := ise.fly.SaveOrUpdateShard(v)
			if err != nil {
				loopList.PushBack(v)
			}
		}
	}
	for loopList.Len() > 0 {
		for e := loopList.Front(); e != nil; e = e.Next() {
			try := e.Value.(*CfgShard)
			err := ise.fly.SaveOrUpdateShard(try)
			if err != nil {
				logger.Debug("%v", try)
			} else {
				loopList.Remove(e)
			}
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
