package server

import (
	"testing"
	"time"
)

func TestProxyServer_Start(t *testing.T) {
	s := NewProxyServer()
	s.StartProxyServer()
	s.StartApiServer()

	time.Sleep(time.Hour)
}
