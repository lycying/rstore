package cfg

import ()
import (
	"errors"
	"fmt"
	"github.com/lycying/rstore/redisx/postgres"
	"regexp"
	"time"
)

type Fly struct {
	ise *Instance
}

func NewFly() *Fly {
	fly := &Fly{}
	fly.ise = NewInstance()
	return fly
}

func (fly *Fly) SaveOrUpdatePostgres(cfg *CfgDBPostgres) error {
	t := cfg.Type
	if _, ok := fly.ise.DBMap[t]; !ok {
		fly.ise.DBMap[t] = make(map[string]*DB_Instance, 0)
	}
	var needDestory bool = false
	if db, ok := fly.ise.DBMap[t][cfg.Name]; ok {
		//check if replace it
		orgi := db.Cfg.(*CfgDBPostgres)

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
	}
	//mark it if it needed !
	orgi := fly.ise.DBMap[t][cfg.Name]

	db := NewDBInstance(cfg)
	pg, err := postgres.NewPostgres(cfg.Url)
	if err != nil {
		return err
	}
	pg.GetReal().SetMaxIdleConns(cfg.MaxIdle)
	pg.GetReal().SetMaxOpenConns(cfg.MaxOpen)
	pg.GetReal().SetConnMaxLifetime(time.Duration(cfg.MaxLifetime) * time.Second)
	db.Backend = pg

	fly.ise.DBMap[t][cfg.Name] = db

	if needDestory {
		orgi.Backend.Close()
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
	dbgroup := NewDBGroupInstance(cfg)
	t := cfg.Type
	//rebuild
	totalReadWeight := 0
	for _, item := range cfg.Items {
		totalReadWeight += item.ReadWeight
		new := &DBExt_Instance{}
		new.IsMaster = item.IsMaster
		new.ReadWeight = item.ReadWeight

		if db, ok := fly.ise.DBMap[t][item.Name]; ok {
			new.DB = db
		} else {
			return errors.New(fmt.Sprintf("no db named '%v' found ! can make instance. ", item.Name))
		}
		dbgroup.MasterSlaves = append(dbgroup.MasterSlaves, new)
	}
	dbgroup.TotalReadWeight = totalReadWeight
	fly.ise.DBGroupMap[cfg.Name] = dbgroup
	return nil
}
func (fly *Fly) SaveOrUpdateShard(cfg *CfgShard) error {
	shard := NewShardInstance(cfg)
	for _, item := range cfg.ShardMap {
		if item.RefType == "shard" {
			if subShard, ok := fly.ise.ShardMap[item.RefName]; ok {
				print(subShard)
			} else {
				return errors.New(fmt.Sprintf("no shard named '%v' found ! can make instance. ", item.RefName))
			}
		} else {
			if subDBGroup, ok := fly.ise.DBGroupMap[item.RefName]; ok {
				print(subDBGroup)
			} else {
				return errors.New(fmt.Sprintf("no dbgroup named '%v' found ! can make instance. ", item.RefName))
			}
		}
	}
	fly.ise.ShardMap[cfg.Name] = shard
	return nil
}
func (fly *Fly) SaveOrUpdateRule(cfg *CfgRule) error {
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
	fly.ise.RuleMap[cfg.Name] = rule
	return nil
}
func (fly *Fly) RemovePostgres(name string) error {
	t := "postgres"
	//check use
	for _, v := range fly.ise.DBGroupMap {
		if v.Cfg.Type == t {
			for _, gv := range v.Cfg.Items {
				if gv.Name == name {
					return errors.New(fmt.Sprintf(" %v is used by dbgroup %v", name, v.Cfg.Name))
				}
			}
		}
	}
	//delete it
	if v, ok := fly.ise.DBMap[t][name]; ok {
		v.Backend.Close()
		delete(fly.ise.DBMap[t], name)
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
	for _, v := range fly.ise.ShardMap {
		for _, sv := range v.Cfg.ShardMap {
			if sv.RefType == "dbgroup" {
				if sv.RefName == name {

					return errors.New(fmt.Sprintf(" %v is used by shard %v", name, v.Cfg.Name))
				}
			}
		}
	}
	if _, ok := fly.ise.DBGroupMap[name]; ok {
		delete(fly.ise.DBGroupMap, name)
	}
	return nil
}
func (fly *Fly) RemoveShard(name string) error {
	for _, v := range fly.ise.ShardMap {
		for _, sv := range v.Cfg.ShardMap {
			if sv.RefType == "shard" {
				if sv.RefName == name {
					return errors.New(fmt.Sprintf(" %v is used by shard %v", name, v.Cfg.Name))
				}
			}
		}
	}
	for _, v := range fly.ise.RuleMap {
		if v.Cfg.ShardName == name {
			return errors.New(fmt.Sprintf(" %v is used by rule %v", name, v.Cfg.Name))
		}
	}
	if _, ok := fly.ise.ShardMap[name]; ok {
		delete(fly.ise.ShardMap, name)
	}
	return nil
}
func (fly *Fly) RemoveRule(name string) error {
	if _, ok := fly.ise.RuleMap[name]; ok {
		delete(fly.ise.RuleMap, name)
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
