## alpha 版
<img src="https://rawgithub.com/lycying/psd/master/rstore/header.png">

`rstore`是一个使用**golang**实现的快速、轻量级的数据库中间件，可使用`redis`协议与其交互，根据路由规则，某类key可同时存放在`mysql`，`postgresql`，`redis`中（更多的后端如`hbase`也可以很容易支持）。为了方便配置管理，`restore`以插件形式提供了名为`eyes`的后台管理工具，访问任意节点的相应端口，即可进行web交互。

`rstore`主要解决使用`redis`作为存储但容量有限的场景，通过将数据转移到传统db中，节省成本，你可以一半key存放在`postgresql`中，另一半存放在`mysql`中。`rstore`尽量让此过程平滑过渡，业务方无须关心数据的具体组织方式。

### 特性

- 使用`redis`协议访问交互，可使用`redis-cli`，`redis-benchmark`等现有工具和客户端
- 解决用`redis`做存储的公司的成本，从`reids`切换到db只需要几分钟(数据迁移工具开发中)
- 保持长连接，通过代理的方式减少缓存服务器的连接数
- 所有配置可**在线生效**，不需要重启代理服务器
- 使用`key`正则匹配的方式解决`redis key`的随意存储，没有备案的`key`将无法入库
- 正则的`group`可作为`hashkey`，这和`Twemproxy`的`hashtag`类似但更强大
- 可使用**多层路由**，比如`1000-2000`范围使用`hash`方式存放在`redis`，`0-1000`范围的使用`mod`存放在`mysql`，路由支持`range`、`mod`、`hash`、`ketama_hash`等
- 如`db`有效率问题，db可前置`redis`做缓存
- 为提高数据一致性，提供延迟读从的功能 (todo)
- 支持单写、双写、多写，通过权重进行读请求分配
- 配置管理可选`etcd`，`zookeeper (todo)`，这是抽象层
- 有支持工具来进行问题排查，比如一个key从路由到后端的`path` (todo)
- 管理接口通过提供`rest api`实现，你可以很容易集成到你自己的管理后台中
- 监控通过提供`rest api`，你可以很容集成到`grafana`、 `zabbix`、`nagios`中
- 插件方式，很容易做其他后端存储的功能 (调整中)

### 缺点

- 不支持针对多个值的操作，比如取`sets`的子交并补等
- 不支持Redis的事务操作
- `redis`的不少特性不支持，不过以笔者的经验，现有的支持足够了
- 增删节点数据还没有`Reblance`，你还是需要自己去写迁移工具 (暂时的)
- 自动`failover`需要自己实现

### 性能

测试中。

与你的db配置有关。

### 扩容

对于扩容最简单的办法是：

- 创建新的集群；
- 双写两个集群；
- 把数据从老集群迁移到新集群（不存在才设置值，防止覆盖新的值）；
- 复制速度要根据实际情况调整，不能影响老集群的性能；
- 切换到新集群即可，如果使用rstore代理层的话，可以做到迁移对读的应用透明。

### 设计
<img src="https://rawgithub.com/lycying/rstore/master/doc/design.svg">

<img src="https://rawgithub.com/lycying/rstore/master/doc/tables.svg">
### 后台管理工具概览
数据库节点列表，可设置连接池等

<img src="https://rawgithub.com/lycying/psd/master/rstore/dbunit_list.png">

组合数据库节点，组成数据库集群，如箭头所指为一主一丛节点
<img src="https://rawgithub.com/lycying/psd/master/rstore/dbgroup_list.png">

设置分片规则，规则条目可选择数据库集群，也可选择子分片规则
<img src="https://rawgithub.com/lycying/psd/master/rstore/shard_list.png">

正则匹配key列表，`hashslot`指明哪个group作为分区键，注意`0`表示整个的key
<img src="https://rawgithub.com/lycying/psd/master/rstore/rule_list.png">

### schema
```javascript
var redisScm = {
    "type": "object",
    "title": "Redis",
    "properties": {
        "Type":{
            "type":"string",
            "enum": ["redis"],
            "required":true
        },
        "Name":{
            "type":"string",
            "title": "Supply a unique name",
            "default": ""
        },
        "Enable":{
            "type": "boolean",
            "title": "If inused",
            "format": "checkbox"
        },
        "Server": {
            "type": "string",
            "title": "Server",
            "minLength": 2,
            "default": "localhost"
        },
        "Port": {
            "type": "integer",
            "title": "Port",
            "default": 6379,
        },
        "Mark": {
            "type": "string",
            "title": "Mark",
            "format": "textarea",
            "default": "",
        },
    },
};
var postgresScm = {
    "type": "object",
    "title": "Postgres",
    "properties": {
        "Type":{
            "type":"string",
            "enum": ["postgres"],
            "required":true
        },
        "Name":{
            "type":"string",
            "title": "Supply a unique name",
            "default": ""
        },
        "Enable":{
            "type": "boolean",
            "title": "If inused",
            "format": "checkbox"
        },
        "Url": {
            "type": "string",
            "title": "URL",
            "minLength": 12,
            "default": "postgres://postgres:postgres@localhost/postgres?sslmode=disable"
        },
        "MaxIdle": {
            "type": "integer",
            "title": "MaxIdle",
            "default": 10,
        },
        "MaxOpen": {
            "type": "integer",
            "title": "MaxOpen",
            "default": 10,
        },
        "MaxLifetime": {
            "type": "integer",
            "title": "MaxLifetime (Second)",
            "default": 60,
        },
        "Mark": {
            "type": "string",
            "title": "Mark",
            "format": "textarea",
            "default": "",
        },
    },
};
var mysqlScm = {
    "type": "object",
    "title": "Mysql",
    "properties": {
        "Type":{
            "type":"string",
            "enum": ["mysql"],
            "required":true
        },
        "Name":{
            "type":"string",
            "title": "Supply a unique name",
            "default": ""
        },
        "Enable":{
            "type": "boolean",
            "title": "If inused",
            "format": "checkbox"
        },
        "Url": {
            "type": "string",
            "title": "URL",
            "minLength": 12,
            "default": "user:password@tcp(localhost:5555)/dbname?tls=skip-verify&autocommit=true"
        },
        "MaxIdle": {
            "type": "integer",
            "title": "MaxIdle",
            "default": 10,
        },
        "MaxOpen": {
            "type": "integer",
            "title": "MaxOpen",
            "default": 10,
        },
        "MaxLifetime": {
            "type": "integer",
            "title": "MaxLifetime (Second)",
            "default": 60,
        },
        "Mark": {
            "type": "string",
            "title": "Mark",
            "format": "textarea",
            "default": "",
        },
    },
};
////////////////////////////////////////////
//dbgroup
var scm = {
    "type": "object",
    "title": "DB Group",
    "properties": {
        "Type":{
            "type":"string",
            "required":true,
            "enum": ["redis","postgres","mysql"]
        },
        "Name":{
            "type":"string",
            "title": "Supply a unique name",
            "default": ""
        },
        "Enable":{
            "type": "boolean",
            "title": "If inused",
            "format": "checkbox"
        },
        "ReplicateMode":{
            "type": "string",
            "title": "ReplicateMode",
            "required":true,
            "enum": ["writeall","writeone","discard"]
        },
        "Mark": {
            "type": "string",
            "title": "Mark",
            "format": "textarea",
            "default": "",
        },
        "Items":{
            "type": "array",
            "title":"DB Items",
            "items":{
                "type":"object",
                "title":"DBGroup Item",
                "properties": {
                    "Name":{
                        "type": "string",
                        "title": "DBRef",
                        "format": "select",
                        "required":true,
                        "enum": ["redis"]
                    },
                    "IsMaster":{
                        "type": "boolean",
                        "title": "Master",
                        "format": "checkbox",
                        "default":true,
                    },
                   "ReadWeight":{
                        "type": "integer",
                        "title": "ReadWeight",
                        "default":1
                    },

                },
            },
        }
    },
};

////////////////////////////////////////////
//shard
var scm = {
    "type": "object",
    "title": "Shard Defined",
    "properties": {
        "ShardType":{
            "type":"string",
            "ShardType":"string",
            "required":true,
            "enum": ["hash","mod","range","ketama_hash"],
            "default":"hash"
        },
        "Name":{
            "type":"string",
            "title": "Supply a unique name",
            "default": ""
        },
        "ShardMap":{
            "type": "array",
            "uniqueItems": true,
            "title":"Shard Map",
            "items":{
                "title":"Shard Item",
                "oneOf": [
                {
                    "title":"dbgroup",
                    "properties": {
                        "RefName":{
                            "type": "string",
                            "title": "Reference Name",
                            "format": "select",
                            "required":true,
                            "enum": []
                        },
                        "ShardStr":{
                            "type": "string",
                            "title": "ShardStr",
                        },
                        "RefType":{
                            "type": "string",
                            "title": "Reference Type",
                            "format": "select",
                            "required":true,
                            "enum": ["dbgroup"],
                            "minLength": 7,
                        },

                    },
                },
                {
                    "title":"shard",
                    "properties": {
                        "RefName":{
                            "type": "string",
                            "title": "Reference Name",
                            "format": "select",
                            "required":true,
                            "enum": []
                        },
                        "ShardStr":{
                            "type": "string",
                            "title": "ShardStr",
                        },
                        "RefType":{
                            "type": "string",
                            "title": "Reference Type)",
                            "format": "select",
                            "required":true,
                            "enum": ["shard"],
                            "minLength": 5,
                        },

                    },
                },

                ]
            },
        }
    },
};
////////////////////////////////////////////
//rule
var scm = {
    "type": "object",
    "title": "Rule Defined",
    "properties": {
        "Name":{
            "type":"string",
            "title": "Supply a unique name",
            "default": ""
        },
        "Order":{
            "type":"integer",
            "title": "Order",
            "default": 0
        },
        "Regexp":{
            "type":"string",
            "title": "Regexp",
            "minLength": 2,
            "default": "app:(\\d+):age"
        },
        "HashSlot":{
            "type":"integer",
            "title": "HashSlot",
            "default": 0
        },
        "ShardName":{
            "type":"string",
            "title": "ShardName",
            "format": "select",
            "required":true,
            "enum": []
        },
        "Example":{
            "type":"string",
            "title": "Example",
            "default": ""
        },
        "Mark":{
            "type":"string",
            "title": "Mark",
            "format": "textarea",
            "default": ""
        },
    },
};
```
