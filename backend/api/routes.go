package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

// RegisterRoutes 注册API路由
func RegisterRoutes(app *fiber.App, handler *Handler) {
	// 添加中间件
	app.Use(cors.New())
	app.Use(logger.New())
	app.Use(recover.New())

	// API路由组
	api := app.Group("/api")

	// 健康检查
	api.Get("/health", handler.HealthCheck)

	// 音频相关路由
	api.Post("/audio/upload", handler.UploadAudio)
	api.Post("/audio/stream", handler.StreamAudio)

	// 会议相关路由
	api.Post("/meetings/analyze", handler.AnalyzeTranscript)
	api.Post("/meetings/sync-notion", handler.SyncToNotionHandler)
}
