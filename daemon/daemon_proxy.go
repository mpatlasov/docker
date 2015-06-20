// +build !exclude_graphdriver_proxy

package daemon

import (
	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/daemon/graphdriver"
	_ "github.com/docker/docker/daemon/graphdriver/proxy"
	"net"
	"net/http"
	"net/rpc"
	"strings"
)

func StartProxyDaemon(config *Config, proto, addr string) {
	p := new(graphdriver.ProxyAPI)
	p.Root = config.Root
	p.CtName = strings.TrimPrefix(config.Labels[0], "ctname=")
	rpc.Register(p)
	rpc.HandleHTTP()
	l, e := net.Listen(proto, addr)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	http.Serve(l, nil)
}
