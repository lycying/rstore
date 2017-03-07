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

### schema
