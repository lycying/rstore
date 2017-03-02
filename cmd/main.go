package main

import (
	"github.com/lycying/rstore/server"
	"time"
)

func main() {

	s := server.NewProxyServer()
	s.StartProxyServer()
	s.StartApiServer()

	time.Sleep(time.Hour * 24 * 365)

}
