package cfg


import (
	"encoding/json"
	"github.com/coreos/etcd/client"
	"golang.org/x/net/context"
	"time"
)

type EtcdClient struct {
	kapi client.KeysAPI
	prefix string
}

func NewEtcdClient() *EtcdClient {
	etcd := &EtcdClient{}
	cfg := client.Config{
		Endpoints:               []string{"http://127.0.0.1:2379"},
		Transport:               client.DefaultTransport,
		HeaderTimeoutPerRequest: time.Second,
	}
	c, err := client.New(cfg)
	if err != nil {
	}
	etcd.kapi = client.NewKeysAPI(c)
	etcd.prefix = "/dir"
	return etcd
}
func (c *EtcdClient) SaveOrUpdatePostgres(cfg *CfgDBPostgres) error {
	b, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	k := c.prefix + "/db/pg/" + cfg.Name + "/"
	v := string(b)
	_, err = c.kapi.Create(context.Background(), k, v)
	if err != nil {
		_, err = c.kapi.Update(context.Background(), k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
func (c *EtcdClient) SaveOrUpdateMySql(cfg *CfgDBMysql) error {
	b, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	k := c.prefix + "/db/mysql/" + cfg.Name + "/"
	v := string(b)
	_, err = c.kapi.Create(context.Background(), k, v)
	if err != nil {
		_, err = c.kapi.Update(context.Background(), k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
func (c *EtcdClient) SaveOrUpdateRedis(cfg *CfgDBRedis) error {
	b, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	k := c.prefix + "/db/redis/" + cfg.Name + "/"
	v := string(b)
	_, err = c.kapi.Create(context.Background(), k, v)
	if err != nil {
		_, err = c.kapi.Update(context.Background(), k, v)
		if err != nil {
			return err
		}
	}
	return nil
}

func nodeWalk(node *client.Node, vars map[string]string) error {
	if node != nil {
		key := node.Key
		if !node.Dir {
			vars[key] = node.Value
		} else {
			for _, node := range node.Nodes {
				nodeWalk(node, vars)
			}
		}
	}
	return nil
}

func (c *EtcdClient) GetAllPostgres() ([]*CfgDBPostgres, error) {
	vars := make(map[string]string)
	k := c.prefix + "/db/pg/"
	resp, err := c.kapi.Get(context.Background(), k, &client.GetOptions{
		Recursive: true,
		Sort:      true,
		Quorum:    true,
	})
	if err != nil {
		return nil, err
	}
	err = nodeWalk(resp.Node, vars)
	if err != nil {
		return nil, err
	}

	ret := make([]*CfgDBPostgres, len(vars))
	index := 0
	for _, v := range vars {
		var item *CfgDBPostgres
		json.Unmarshal([]byte(v), &item)
		ret[index] = item
		index++
	}
	return ret, nil
}
func (c *EtcdClient) GetAllMysql() ([]*CfgDBMysql, error) {
	vars := make(map[string]string)
	k := c.prefix + "/db/mysql/"
	resp, err := c.kapi.Get(context.Background(), k, &client.GetOptions{
		Recursive: true,
		Sort:      true,
		Quorum:    true,
	})
	if err != nil {
		return nil, err
	}
	err = nodeWalk(resp.Node, vars)
	if err != nil {
		return nil, err
	}

	ret := make([]*CfgDBMysql, len(vars))
	index := 0
	for _, v := range vars {
		var item *CfgDBMysql
		json.Unmarshal([]byte(v), &item)
		ret[index] = item
		index++
	}
	return ret, nil
}
func (c *EtcdClient) GetAllRedis() ([]*CfgDBRedis, error) {
	vars := make(map[string]string)
	k := c.prefix + "/db/redis/"
	resp, err := c.kapi.Get(context.Background(), k, &client.GetOptions{
		Recursive: true,
		Sort:      true,
		Quorum:    true,
	})
	if err != nil {
		return nil, err
	}
	err = nodeWalk(resp.Node, vars)
	if err != nil {
		return nil, err
	}

	ret := make([]*CfgDBRedis, len(vars))
	index := 0
	for _, v := range vars {
		var item *CfgDBRedis
		json.Unmarshal([]byte(v), &item)
		ret[index] = item
		index++
	}
	return ret, nil
}
func (c *EtcdClient) SaveOrUpdateDBGroup(cfg *CfgDBGroup) error {
	b, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	k := c.prefix + "/dbgroup/" + cfg.Name + "/"
	v := string(b)
	_, err = c.kapi.Create(context.Background(), k, v)
	if err != nil {
		_, err = c.kapi.Update(context.Background(), k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
func (c *EtcdClient) GetAllDBGroup() ([]*CfgDBGroup, error) {
	vars := make(map[string]string)
	k := c.prefix + "/dbgroup/"
	resp, err := c.kapi.Get(context.Background(), k, &client.GetOptions{
		Recursive: true,
		Sort:      true,
		Quorum:    true,
	})
	if err != nil {
		return nil, err
	}
	err = nodeWalk(resp.Node, vars)
	if err != nil {
		return nil, err
	}

	ret := make([]*CfgDBGroup, len(vars))
	index := 0
	for _, v := range vars {
		var item *CfgDBGroup
		json.Unmarshal([]byte(v), &item)
		ret[index] = item
		index++
	}
	return ret, nil
}
func (c *EtcdClient) SaveOrUpdateShard(cfg *CfgShard) error {
	b, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	k := c.prefix + "/shard/" + cfg.Name + "/"
	v := string(b)
	_, err = c.kapi.Create(context.Background(), k, v)
	if err != nil {
		_, err = c.kapi.Update(context.Background(), k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
func (c *EtcdClient) GetAllShard() ([]*CfgShard, error) {
	vars := make(map[string]string)
	k := c.prefix + "/shard/"
	resp, err := c.kapi.Get(context.Background(), k, &client.GetOptions{
		Recursive: true,
		Sort:      true,
		Quorum:    true,
	})
	if err != nil {
		return nil, err
	}
	err = nodeWalk(resp.Node, vars)
	if err != nil {
		return nil, err
	}

	ret := make([]*CfgShard, len(vars))
	index := 0
	for _, v := range vars {
		var item *CfgShard
		json.Unmarshal([]byte(v), &item)
		ret[index] = item
		index++
	}
	return ret, nil
}

func (c *EtcdClient) RemovePostgres(name string) error {
	k := c.prefix + "/db/pg/" + name
	_, err := c.kapi.Delete(context.Background(), k, &client.DeleteOptions{
		Recursive: true,
		Dir:       false,
	})
	return err
}
func (c *EtcdClient) RemoveMysql(name string) error {
	k := c.prefix + "/db/mysql/" + name
	_, err := c.kapi.Delete(context.Background(), k, &client.DeleteOptions{
		Recursive: true,
		Dir:       false,
	})
	return err
}
func (c *EtcdClient) RemoveRedis(name string) error {
	k := c.prefix + "/db/redis/" + name
	_, err := c.kapi.Delete(context.Background(), k, &client.DeleteOptions{
		Recursive: true,
		Dir:       false,
	})
	return err
}
func (c *EtcdClient) RemoveShard(name string) error {
	k := c.prefix + "/shard/" + name
	_, err := c.kapi.Delete(context.Background(), k, &client.DeleteOptions{
		Recursive: true,
		Dir:       false,
	})
	return err
}
func (c *EtcdClient) RemoveDBGroup(name string) error {
	k := c.prefix + "/dbgroup/" + name
	_, err := c.kapi.Delete(context.Background(), k, &client.DeleteOptions{
		Recursive: true,
		Dir:       false,
	})
	return err
}

func (c *EtcdClient) SaveOrUpdateRule(cfg *CfgRule) error{
	b, err := json.Marshal(cfg)
	if err != nil {
		return err
	}
	k := c.prefix + "/rule/" + cfg.Name + "/"
	v := string(b)
	_, err = c.kapi.Create(context.Background(), k, v)
	if err != nil {
		_, err = c.kapi.Update(context.Background(), k, v)
		if err != nil {
			return err
		}
	}
	return nil
}
func (c *EtcdClient) GetAllRule() ([]*CfgRule, error){
		vars := make(map[string]string)
	k := c.prefix + "/rule/"
	resp, err := c.kapi.Get(context.Background(), k, &client.GetOptions{
		Recursive: true,
		Sort:      true,
		Quorum:    true,
	})
	if err != nil {
		return nil, err
	}
	err = nodeWalk(resp.Node, vars)
	if err != nil {
		return nil, err
	}

	ret := make([]*CfgRule, len(vars))
	index := 0
	for _, v := range vars {
		var item *CfgRule
		json.Unmarshal([]byte(v), &item)
		ret[index] = item
		index++
	}
	return ret, nil
}
func (c *EtcdClient) RemoveRule(name string) error{
	k := c.prefix + "/rule/" + name
	_, err := c.kapi.Delete(context.Background(), k, &client.DeleteOptions{
		Recursive: true,
		Dir:       false,
	})
	return err
}
