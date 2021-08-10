package xhttp

import (
	"log"
	"net/http"
	"sync"
	"time"
)

type HttpServer struct {
	server *http.Server
	addr   string
	lock   sync.Mutex
	routes map[string]func(w http.ResponseWriter, r *http.Request)
}

func NewServer(addr string) *HttpServer {
	return &HttpServer{
		addr:   addr,
		routes: make(map[string]func(w http.ResponseWriter, r *http.Request), 0),
	}
}
func (hs *HttpServer) AddRoute(route string, fn func(w http.ResponseWriter, r *http.Request)) {
	defer hs.lock.Unlock()
	hs.lock.Lock()
	hs.routes[route] = fn
}
func (hs *HttpServer) Start() {

	mux := http.NewServeMux()
	for k := range hs.routes {
		route := hs.routes[k]
		mux.HandleFunc(k, route)
	}
	hs.server = &http.Server{Addr: hs.addr, WriteTimeout: time.Second * 4, Handler: mux}

	log.Printf("启动 xhttp")
	err := hs.server.ListenAndServe()
	if err != nil {
		// 正常退出
		if err == http.ErrServerClosed {
			log.Fatal("Server closed under request")
		} else {
			log.Fatal("Server closed unexpected", err)
		}
	}
	log.Fatal("关闭 xhttp")
}
func (hs *HttpServer) Stop() {
	err := hs.server.Shutdown(nil)
	if err != nil {
		log.Printf("shutdown the server err")
	}
}
