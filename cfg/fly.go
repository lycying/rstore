package cfg

import ()
import (
	"errors"
	"fmt"
	"github.com/lycying/rstore/redisx/postgres"
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
		go orgi.Backend.(*postgres.Postgres).GetReal().Close()
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
	t := cfg.Type
	if _, ok := fly.ise.DBGroupMap[t]; !ok {
		fly.ise.DBGroupMap[t] = make(map[string]*DBGroup_Instance, 0)
	}
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
	}
	dbgroup := NewDBGroupInstance(cfg)
	dbgroup.TotalReadWeight = totalReadWeight
	fly.ise.DBGroupMap[t][cfg.Name] = dbgroup
	return nil
}
func (fly *Fly) SaveOrUpdateShard(cfg *CfgShard) error {
	return nil
}
func (fly *Fly) RemovePostgres(name string) error {
	return nil
}
func (fly *Fly) RemoveMysql(name string) error {
	return nil
}
func (fly *Fly) RemoveRedis(name string) error {
	return nil
}
func (fly *Fly) RemoveShard(name string) error {
	return nil
}
func (fly *Fly) RemoveDBGroup(name string) error {
	return nil
}
func (fly *Fly) SaveOrUpdateRule(cfg *CfgRule) error {
	return nil
}
func (fly *Fly) RemoveRule(name string) error {
	return nil
}

//no need implement it
func (fly *Fly) GetAllPostgres() ([]*CfgDBPostgres, error) { return nil, nil }
func (fly *Fly) GetAllMysql() ([]*CfgDBMysql, error)       { return nil, nil }
func (fly *Fly) GetAllRedis() ([]*CfgDBRedis, error)       { return nil, nil }
func (fly *Fly) GetAllDBGroup() ([]*CfgDBGroup, error)     { return nil, nil }
func (fly *Fly) GetAllShard() ([]*CfgShard, error)         { return nil, nil }
func (fly *Fly) GetAllRule() ([]*CfgRule, error)           { return nil, nil }
