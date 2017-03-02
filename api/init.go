package api

import "github.com/lycying/log"

var logger *log.Logger

func init() {
	logger, _ = log.New(log.DEBUG, "")
}
