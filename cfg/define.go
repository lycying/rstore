package cfg

type DBer interface {
}

type CfgBase struct {
	DBer
	Type   string
	Name   string //UNIQUE
	Enable bool
	Mark   string //you can write some usefull infomation here
	State  int    //the flag show if the unit is used or not
}

type CfgDBPostgres struct {
	CfgBase
	Url string

	//for postgres
	MaxIdle     int
	MaxOpen     int
	MaxLifetime int //second
}

type CfgDBMysql struct {
	CfgBase
	Url string

	MaxIdle     int
	MaxOpen     int
	MaxLifetime int
}

type CfgDBRedis struct {
	CfgBase
	Server string
	Port   int
}

type CfgDBGroupItem struct {
	Name          string
	IsMaster      bool

	ReadWeight int //default master is zero
}

type CfgDBGroup struct {
	CfgBase

	ReplicateMode string
	Items []*CfgDBGroupItem
}

type ShardItem struct {
	RefType  string //if it's a dbgroup or a sub-router
	RefName  string //the real dbgroup or the sub-router
	ShardStr string
}

// example
// [0-1000]   :hash-router-magic-v2
// [1000-2000]:redis-group-v30.cluster.com:6379

// example
// 0:redis-group-ok.com:9001
// 1:pg-db-group-ok.com:9002
// 2:redis-group-ok.com:9003
// 3:redis-group-ok.com:9003
// 4:redis-group-ok.com:9003
// 5:redis-group-ok.com:9003
// 6:redis-group-ok.com:9003
// 7:redis-group-ok.com:9003
// 8:redis-group-ok.com:9003
// 9:redis-group-ok.com:9004

// another example
// foo: redis-cluster-001.com:6379
// bar: redis-cluster-002.com:6379

type CfgShard struct {
	Name      string
	ShardType string
	ShardMap  []*ShardItem
}

type CfgRule struct {
	Name      string
	Order     int
	Regexp    string
	HashSlot  int
	ShardName string
	Example   string
	Mark      string
}

type CfgPublic struct {
	SaverType string
	ConnStr   string
}
