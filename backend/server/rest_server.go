package server

import (
	"fmt"
	"log"
	"net/http"

	"meeting-mm/api"
)

// RESTServer 表示RESTful API服务器
type RESTServer struct {
	port    string
	handler *api.RestHandler
}

// NewRESTServer 创建一个新的RESTful API服务器
func NewRESTServer(port string) *RESTServer {
	return &RESTServer{
		port:    port,
		handler: api.NewRestHandler(),
	}
}

// Start 启动RESTful API服务器
func (s *RESTServer) Start() error {
	// 创建多路复用器
	mux := http.NewServeMux()

	// 注册RESTful API路由
	mux.HandleFunc("/api/health", s.handler.HealthCheckHandler)
	mux.HandleFunc("/api/meetings/sync-notion", s.handler.SyncMeetingToNotionHandler)

	// 启动服务器
	addr := fmt.Sprintf(":%s", s.port)
	log.Printf("RESTful API服务器启动在 http://localhost%s", addr)
	return http.ListenAndServe(addr, mux)
}
