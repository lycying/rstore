package cfg


type Watcher interface {
	OnPostgresChange(*CfgDBPostgres) error
	OnMysqlChange(*CfgDBPostgres) error
	OnRedisChange(*CfgDBRedis) error
}

type RedisSaver interface {
	SaveOrUpdateRedis(*CfgDBRedis) error
	RemoveRedis(string) error
	GetAllRedis() ([]*CfgDBRedis, error)
	GetAllRedisByGroup(*CfgDBGroup) ([]*CfgDBRedis, error)
}
type PostgresSaver interface {
	SaveOrUpdatePostgres(*CfgDBPostgres) error
	RemovePostgres(string) error
	GetAllPostgres() ([]*CfgDBPostgres, error)
	GetAllPostgresByGroup(*CfgDBGroup) ([]*CfgDBPostgres, error)
}
type MysqlSaver interface {
	SaveOrUpdateMySql(*CfgDBMysql) error
	RemoveMysql(string) error
	GetAllMysql() ([]*CfgDBMysql, error)
	GetAllMysqlByGroup(*CfgDBGroup) ([]*CfgDBMysql, error)
}

type Saver interface {
	RedisSaver
	PostgresSaver
	MysqlSaver

	SaveOrUpdateDBGroup(*CfgDBGroup) error
	GetAllDBGroup() ([]*CfgDBGroup, error)
	RemoveDBGroup(string) error

	SaveOrUpdateShard(*CfgShard) error
	GetAllShard() ([]*CfgShard, error)
	RemoveShard(string) error

	SaveOrUpdateRule(*CfgRule) error
	GetAllRule() ([]*CfgRule, error)
	RemoveRule(string) error
}
