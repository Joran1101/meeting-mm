package main

import (
	"log"

	"meeting-mm/config"
	"meeting-mm/server"
)

func main() {
	// 加载配置
	if err := config.LoadConfig(""); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	cfg := config.GetConfig()

	// 创建RESTful API服务器
	restServer := server.NewRESTServer(cfg.Port)

	// 启动服务器
	log.Fatal(restServer.Start())
}
