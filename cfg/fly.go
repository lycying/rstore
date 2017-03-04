package cfg

import (
	"container/list"
	"errors"
	"fmt"
	"github.com/lycying/rstore/redisx/postgres"
	"regexp"
	"time"
)

type Fly struct {
}

func NewFly() *Fly {
	fly := &Fly{}
	return fly
}

func (fly *Fly) SaveOrUpdatePostgres(cfg *CfgDBPostgres) error {
	ise := GetInstance()
	t := cfg.Type
	//make new type map if not exist
	if _, ok := ise.DBMap[t]; !ok {
		ise.DBMap[t] = make(map[string]*DB_Instance, 0)
	}
	//if need destory the orgi connections
	var needDestory bool = false
	if db, ok := ise.DBMap[t][cfg.Name]; ok {
		//check if replace it
		orgi := db.Cfg.(*CfgDBPostgres)

		//the connection url is changed
		if orgi.Url != cfg.Url {
			needDestory = true
		}
		//same
		if orgi.MaxIdle == cfg.MaxIdle && orgi.MaxLifetime == cfg.MaxLifetime && orgi.MaxOpen == orgi.MaxOpen {
			//pass
		} else {
			pg := db.Backend.(*postgres.Postgres)
			pg.GetReal().SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)
			pg.GetReal().SetMaxOpenConns(cfg.MaxOpen)
			pg.GetReal().SetMaxIdleConns(cfg.MaxIdle)
		}
	} else {
		//mark it if it needed !
		orgi := ise.DBMap[t][cfg.Name]
		db := NewDBInstance(cfg)
		pg, err := postgres.NewPostgres(cfg.Url)
		if err != nil {
			return err
		}
		pg.GetReal().SetMaxIdleConns(cfg.MaxIdle)
		pg.GetReal().SetMaxOpenConns(cfg.MaxOpen)
		pg.GetReal().SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)
		db.Backend = pg
		ise.DBMap[t][cfg.Name] = db
		if needDestory {
			orgi.Backend.Close()
		}
	}

	return nil
}

func (fly *Fly) SaveOrUpdateMySql(cfg *CfgDBMysql) error {
	return nil
}
func (fly *Fly) SaveOrUpdateRedis(cfg *CfgDBRedis) error {
	return nil
}

func (fly *Fly) SaveOrUpdateDBGroup(cfg *CfgDBGroup) error {
	ise := GetInstance()
	t := cfg.Type
	dbgroup := NewDBGroupInstance(cfg)
	//rebuild
	totalReadWeight := 0
	for _, item := range cfg.Items {
		totalReadWeight += item.ReadWeight
		new := &DBExt_Instance{}
		new.IsMaster = item.IsMaster
		new.ReadWeight = item.ReadWeight

		if db, ok := ise.DBMap[t][item.Name]; ok {
			new.DB = db
		} else {
			return errors.New(fmt.Sprintf("no db(%v) named '%v' found ! can not make instance. ", t, item.Name))
		}
		dbgroup.MasterSlaves = append(dbgroup.MasterSlaves, new)
	}
	dbgroup.TotalReadWeight = totalReadWeight
	ise.DBGroupMap[cfg.Name] = dbgroup
	return nil
}

func shardDeps(shard *CfgShard, allShard []*CfgShard) *list.List {
	shardList := list.New()
	for _, v := range shard.ShardMap {
		if v.RefType == Shard_RefType_Shard {
			shardList.PushBack(v.RefName)
			var subShard *CfgShard
			for _, v2 := range allShard {
				if v2.Name == v.RefName {
					subShard = v2
					break
				}
			}
			if subShard != nil {
				subList := shardDeps(subShard, allShard)
				if subList.Len() > 0 {
					for e := subList.Front(); e != nil; e = e.Next() {
						shardList.PushBack(e.Value)
					}
				}
			}
		}
	}
	return shardList
}

func (fly *Fly) SaveOrUpdateShard(cfg *CfgShard) error {
	ise := GetInstance()
	shard := NewShardInstance(cfg)
	for _, item := range cfg.ShardMap {
		shardItemInstance := NewShardItemInstance(item)
		if item.RefType == Shard_RefType_Shard {
			if subShard, ok := ise.ShardMap[item.RefName]; ok {
				//check loop
				if subShard.Cfg.Name == cfg.Name {
					return errors.New(fmt.Sprintf("can not ref itself %v", cfg.Name))
				}
				allShard, _ := ise.saver.GetAllShard()
				depList := shardDeps(shard.Cfg, allShard)
				for e := depList.Front(); e != nil; e = e.Next() {
					if e.Value.(string) == cfg.Name {
						return errors.New(fmt.Sprintf("has detect loop"))
					}
				}

				shardItemInstance.Holder = subShard
			} else {
				return errors.New(fmt.Sprintf("no shard named '%v' found ! can not make instance. ", item.RefName))
			}
		} else if item.RefType == Shard_RefType_DBGroup {
			if subDBGroup, ok := ise.DBGroupMap[item.RefName]; ok {
				shardItemInstance.Holder = subDBGroup
			} else {
				return errors.New(fmt.Sprintf("no dbgroup named '%v' found ! can not make instance. ", item.RefName))
			}
		}
		shard.ShardParts = append(shard.ShardParts, shardItemInstance)
	}
	ise.ShardMap[cfg.Name] = shard
	return nil
}

func (fly *Fly) SaveOrUpdateRule(cfg *CfgRule) error {
	ise := GetInstance()
	rule := NewRuleInstance(cfg)
	regex, err := regexp.Compile(cfg.Regexp)
	if err != nil {
		return err
	}
	if cfg.Example != "" {
		if !regex.MatchString(cfg.Example) {
			return errors.New(fmt.Sprint("( %v ) not match the regex ( %v )", cfg.Example, cfg.Regexp))
		}
	}
	rule.Regexp = regex
	ise.RuleMap[cfg.Name] = rule
	return nil
}

func (fly *Fly) RemovePostgres(name string) error {
	ise := GetInstance()
	t := "postgres"
	//check use
	for _, v := range ise.DBGroupMap {
		if v.Cfg.Type == t {
			for _, gv := range v.Cfg.Items {
				if gv.Name == name {
					return errors.New(fmt.Sprintf(" %v  (%v) is used by dbgroup %v", name, t, v.Cfg.Name))
				}
			}
		}
	}
	//delete it
	if v, ok := ise.DBMap[t][name]; ok {
		v.Backend.Close()
		delete(ise.DBMap[t], name)
	}
	return nil
}

func (fly *Fly) RemoveMysql(name string) error {
	return nil
}
func (fly *Fly) RemoveRedis(name string) error {
	return nil
}

func (fly *Fly) RemoveDBGroup(name string) error {
	ise := GetInstance()
	for _, v := range ise.ShardMap {
		for _, sv := range v.Cfg.ShardMap {
			if sv.RefType == Shard_RefType_DBGroup {
				if sv.RefName == name {
					return errors.New(fmt.Sprintf(" %v is used by shard %v", name, v.Cfg.Name))
				}
			}
		}
	}
	if _, ok := ise.DBGroupMap[name]; ok {
		delete(ise.DBGroupMap, name)
	}
	return nil
}

func (fly *Fly) RemoveShard(name string) error {
	ise := GetInstance()
	for _, v := range ise.ShardMap {
		for _, sv := range v.Cfg.ShardMap {
			if sv.RefType == "shard" {
				if sv.RefName == name {
					return errors.New(fmt.Sprintf(" %v is used by shard %v", name, v.Cfg.Name))
				}
			}
		}
	}
	for _, v := range ise.RuleMap {
		if v.Cfg.ShardName == name {
			return errors.New(fmt.Sprintf(" %v is used by rule %v", name, v.Cfg.Name))
		}
	}
	if _, ok := ise.ShardMap[name]; ok {
		delete(ise.ShardMap, name)
	}
	return nil
}

func (fly *Fly) RemoveRule(name string) error {
	ise := GetInstance()
	if _, ok := ise.RuleMap[name]; ok {
		delete(ise.RuleMap, name)
	}
	return nil
}

//no need implement it
func (fly *Fly) GetAllPostgres() ([]*CfgDBPostgres, error) { return nil, nil }
func (fly *Fly) GetAllMysql() ([]*CfgDBMysql, error)       { return nil, nil }
func (fly *Fly) GetAllRedis() ([]*CfgDBRedis, error)       { return nil, nil }
func (fly *Fly) GetAllDBGroup() ([]*CfgDBGroup, error)     { return nil, nil }
func (fly *Fly) GetAllShard() ([]*CfgShard, error)         { return nil, nil }
func (fly *Fly) GetAllRule() ([]*CfgRule, error)           { return nil, nil }
