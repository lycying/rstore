package cfg

import "github.com/lycying/log"

var logger *log.Logger
var instance *Instance
var saver Saver
var fly Saver

func init() {
	logger, _ = log.New(log.DEBUG, "")

	saver = NewEtcdClient()
	fly = NewFly()
	instance = NewInstance(saver, fly)

	instance.Init()
}

func GetInstance() *Instance {
	return instance
}

func GetSaver() Saver {
	return saver
}

func GetFly() Saver {
	return fly
}
