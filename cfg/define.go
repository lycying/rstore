package cfg

const (
	DBGroup_ReplicateMode_Writeone = "writeone"
	DBGroup_ReplicateMode_Writeall = "writeall"
	DBGroup_ReplicateMode_Discard  = "discard"

	Shard_RefType_DBGroup = "dbgroup"
	Shard_RefType_Shard   = "shard"

	Shard_ShardType_Hash        = "hash"
	Shard_ShardType_Mod         = "mod"
	Shard_ShardType_Range       = "range"
	Shard_ShardType_Ketama_Hash = "ketama_hash"
)

type DBer interface {
}

type CfgBase struct {
	DBer
	//auto_eject_hosts
	//server_retry_timeout
	//server_failure_limit
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
	Name     string
	IsMaster bool

	ReadWeight int //default master is zero
}

type CfgDBGroup struct {
	Type   string
	Name   string //UNIQUE
	Enable bool
	Mark   string //you can write some usefull infomation here
	State  int    //the flag show if the unit is used or not

	ReplicateMode string
	Items         []*CfgDBGroupItem
}

type ShardItem struct {
	RefType  string //if it's a dbgroup or a sub-router
	RefName  string //the real dbgroup or the sub-router
	ShardStr string
}

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
	SaverType     string
	SaveDirPrefix string
	ConnStr       string
}
