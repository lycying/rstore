package server

import (
	"github.com/lycying/mut"
	"github.com/lycying/rstore/api"
	"github.com/lycying/rstore/codec"
)

type ProxyServer struct {
	proxy *Proxy
}

func NewProxyServer() *ProxyServer {
	srv := &ProxyServer{}
	srv.proxy = newProxy()
	return srv
}

func (srv *ProxyServer) OnConnect(c *mut.Conn) {
}

func (srv *ProxyServer) OnMessage(c *mut.Conn, p mut.Packet) {
	req := p.(*codec.Request)
	//c.WriteRaw([]byte("$-1\r\n"))
	//resp:=codec.NewResponse()
	//resp.WriteOK()
	//resp.WriteOne()
	//resp.WriteString("pong")
	//resp.WriteBytes([]byte("hello"))
	//resp.WriteInlineString("PONG")
	//resp.WriteBulk([][]byte{{'A'}, nil, {'C', 'D'}})
	//resp.WriteStringBulk([]string{"A", "", "CD"})
	//resp.WriteInt(1000)
	//resp.WriteErrorString("nothing ...... sl")
	//c.WriteAsync(resp)
	//c.Flush()
	resp := srv.proxy.invoke(req)
	c.WriteAsync(resp)
	c.Flush()
}

func (srv *ProxyServer) OnClose(c *mut.Conn) {
}

func (srv *ProxyServer) OnError(c *mut.Conn, err error) {

}

func (srv *ProxyServer) StartProxyServer() {

	cfg := mut.DefaultConfig()
	cfg.SetCallback(srv)
	cfg.SetCodec(codec.NewCodec())

	server := mut.NewServer(":14000", cfg)
	err := server.Servo()
	if err != nil {
		logger.Err(err, "servo error")
		return
	}

}

func (srv *ProxyServer) StartApiServer() {
	go api.Start()
	logger.Debug("Start api server at :8888")
}
