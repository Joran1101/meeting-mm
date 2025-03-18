package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"meeting-mm/api"
	"meeting-mm/config"
	"meeting-mm/services"
)

// 设置测试环境
func setupTestEnv() (*gin.Engine, error) {
	// 设置测试模式
	gin.SetMode(gin.TestMode)

	// 加载测试配置
	cfg, err := config.LoadConfig("../.env.test")
	if err != nil {
		// 如果测试配置文件不存在，使用默认配置
		cfg = &config.Config{
			Port:             "8080",
			Env:              "test",
			DeepSeekAPIKey:   "test_key",
			DeepSeekBaseURL:  "https://api.deepseek.com/v1",
			NotionAPIKey:     "test_key",
			NotionDatabaseID: "test_db",
			WhisperModelPath: "../whisper/models/ggml-base.bin",
			UseLocalWhisper:  true,
		}
	}

	// 创建服务
	deepseekService := services.NewDeepSeekService(cfg)
	notionService := services.NewNotionService(cfg)
	whisperService := services.NewWhisperService(cfg)

	// 创建路由
	router := gin.New()
	router.Use(gin.Recovery())

	// 注册API路由
	handler := api.NewHandler(deepseekService, notionService, whisperService)
	api.RegisterRoutes(router, handler)

	return router, nil
}

// 测试健康检查API
func TestHealthCheck(t *testing.T) {
	router, err := setupTestEnv()
	if err != nil {
		t.Fatalf("无法设置测试环境: %v", err)
	}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/health", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	err = json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	assert.Equal(t, "ok", response["status"])
	assert.Contains(t, response, "time")
}

// 测试转录分析API
func TestAnalyzeTranscript(t *testing.T) {
	router, err := setupTestEnv()
	if err != nil {
		t.Fatalf("无法设置测试环境: %v", err)
	}

	// 准备测试数据
	testData := map[string]interface{}{
		"title": "测试会议",
		"transcript": `
			张三：大家好，今天我们讨论项目进度。
			李四：我已经完成了前端开发，需要王五测试一下。
			王五：好的，我明天会进行测试。
			张三：那我们决定下周一发布第一个版本。
		`,
	}

	jsonData, err := json.Marshal(testData)
	assert.NoError(t, err)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/meetings/analyze", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	// 由于DeepSeek服务在测试环境中是模拟的，我们只检查状态码
	// 在实际集成测试中，可以检查更多的响应内容
	assert.Equal(t, http.StatusOK, w.Code)
}

// 测试音频上传API
func TestUploadAudio(t *testing.T) {
	// 跳过此测试，除非明确指定要运行
	if os.Getenv("RUN_UPLOAD_TEST") != "true" {
		t.Skip("跳过音频上传测试。设置 RUN_UPLOAD_TEST=true 环境变量以启用此测试。")
	}

	router, err := setupTestEnv()
	if err != nil {
		t.Fatalf("无法设置测试环境: %v", err)
	}

	// 准备测试音频文件
	testAudioPath := "../test/testdata/test_audio.wav"

	// 检查测试音频文件是否存在
	if _, err := os.Stat(testAudioPath); os.IsNotExist(err) {
		t.Skipf("测试音频文件不存在: %s", testAudioPath)
	}

	// 创建multipart表单
	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	// 添加标题字段
	err = w.WriteField("title", "测试音频上传")
	assert.NoError(t, err)

	// 添加音频文件
	f, err := os.Open(testAudioPath)
	if err != nil {
		t.Fatalf("无法打开测试音频文件: %v", err)
	}
	defer f.Close()

	fw, err := w.CreateFormFile("audio", filepath.Base(testAudioPath))
	assert.NoError(t, err)

	_, err = io.Copy(fw, f)
	assert.NoError(t, err)

	// 关闭multipart writer
	w.Close()

	// 发送请求
	req, err := http.NewRequest("POST", "/api/audio/upload", &b)
	assert.NoError(t, err)
	req.Header.Set("Content-Type", w.FormDataContentType())

	// 执行请求
	resp := httptest.NewRecorder()
	router.ServeHTTP(resp, req)

	// 检查响应
	assert.Equal(t, http.StatusOK, resp.Code)
}

// 创建测试目录
func TestMain(m *testing.M) {
	// 创建测试数据目录
	testDataDir := "../test/testdata"
	if _, err := os.Stat(testDataDir); os.IsNotExist(err) {
		err = os.MkdirAll(testDataDir, 0755)
		if err != nil {
			fmt.Printf("无法创建测试数据目录: %v\n", err)
			os.Exit(1)
		}
	}

	// 创建测试环境文件
	testEnvPath := "../.env.test"
	if _, err := os.Stat(testEnvPath); os.IsNotExist(err) {
		testEnvContent := strings.Join([]string{
			"PORT=8080",
			"ENV=test",
			"DEEPSEEK_API_KEY=test_key",
			"DEEPSEEK_BASE_URL=https://api.deepseek.com/v1",
			"NOTION_API_KEY=test_key",
			"NOTION_DATABASE_ID=test_db",
			"WHISPER_MODEL_PATH=../whisper/models/ggml-base.bin",
			"USE_LOCAL_WHISPER=true",
		}, "\n")

		err = os.WriteFile(testEnvPath, []byte(testEnvContent), 0644)
		if err != nil {
			fmt.Printf("无法创建测试环境文件: %v\n", err)
			os.Exit(1)
		}
	}

	// 运行测试
	code := m.Run()

	// 清理（可选）
	// os.Remove(testEnvPath)

	os.Exit(code)
}
