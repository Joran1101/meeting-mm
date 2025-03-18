package main

import (
	"fmt"
	"log"

	"meeting-mm/api"
	"meeting-mm/config"
	"meeting-mm/services"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func main() {
	// 加载配置
	if err := config.LoadConfig(""); err != nil {
		log.Fatalf("加载配置失败: %v", err)
	}

	cfg := config.GetConfig()

	// 初始化服务
	deepseekService := services.NewDeepSeekService(cfg)
	notionService := services.NewNotionService(cfg)
	whisperService := services.NewWhisperService(cfg)

	// 创建API处理器
	handler := api.NewHandler(deepseekService, notionService, whisperService)

	// 创建Fiber应用
	app := fiber.New(fiber.Config{
		BodyLimit: 50 * 1024 * 1024, // 50MB
	})

	// 添加中间件
	app.Use(logger.New())
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET, POST, PUT, DELETE",
	}))

	// 注册路由
	api.RegisterRoutes(app, handler)

	// 启动服务器
	port := cfg.Port
	log.Printf("服务器启动在 http://localhost:%s", port)
	if err := app.Listen(fmt.Sprintf(":%s", port)); err != nil {
		log.Fatalf("启动服务器失败: %v", err)
	}
}

// 自定义错误处理程序
func customErrorHandler(c *fiber.Ctx, err error) error {
	// 默认状态码为500
	code := fiber.StatusInternalServerError

	// 检查是否为Fiber错误
	if e, ok := err.(*fiber.Error); ok {
		code = e.Code
	}

	// 返回JSON错误响应
	return c.Status(code).JSON(fiber.Map{
		"error": err.Error(),
	})
}
