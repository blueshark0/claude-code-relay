package main

import (
	"claude-code-relay/common"
	"claude-code-relay/middleware"
	"claude-code-relay/model"
	"claude-code-relay/router"
	"claude-code-relay/scheduled"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

// 嵌入前端静态文件
//go:embed web/dist
var staticFS embed.FS

func main() {
	// 加载环境变量
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// 设置全局时间格式
	time.Local, _ = time.LoadLocation("Asia/Shanghai")

	// 设置日志
	common.SetupLogger()
	common.SysLog("Claude Code Relay started")

	// 设置Gin模式
	if os.Getenv("GIN_MODE") != "debug" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化数据库
	err = model.InitDB()
	if err != nil {
		common.FatalLog("failed to initialize database: " + err.Error())
	}
	defer func() {
		if err := model.CloseDB(); err != nil {
			common.FatalLog("failed to close database: " + err.Error())
		}
	}()

	// 初始化Redis
	err = common.InitRedisClient()
	if err != nil {
		common.FatalLog("failed to initialize Redis: " + err.Error())
	}

	// 初始化定时任务服务
	scheduled.InitCronService()
	defer scheduled.StopCronService()

	// 初始化HTTP服务器
	server := gin.New()
	server.Use(gin.CustomRecovery(func(c *gin.Context, err any) {
		common.SysError(fmt.Sprintf("panic detected: %v", err))
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"message": fmt.Sprintf("系统异常: %v", err),
				"type":    "system_error",
			},
		})
	}))

	// 请求ID中间件
	server.Use(middleware.RequestId())

	// 设置日志中间件
	middleware.SetUpLogger(server)

	// 设置跨域中间件
	server.Use(middleware.CORS())

	// 检查是否启用静态文件服务
	serveStatic := os.Getenv("SERVE_STATIC")
	if serveStatic == "" {
		serveStatic = "true" // 默认启用，保持向后兼容
	}

	// 准备静态文件系统
	var staticFileSystem http.FileSystem
	if serveStatic == "true" {
		if sub, err := fs.Sub(staticFS, "web/dist"); err == nil {
			staticFileSystem = http.FS(sub)
		}
		common.SysLog("Static file serving enabled")
	} else {
		common.SysLog("Static file serving disabled - API only mode")
	}

	// 设置API路由
	router.SetAPIRouter(server, staticFS, staticFileSystem, serveStatic == "true")

	// 启动服务器
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// 设置信号处理，优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		common.SysLog("Server starting on port " + port)
		err = server.Run(":" + port)
		if err != nil {
			common.FatalLog("failed to start HTTP server: " + err.Error())
		}
	}()

	// 等待退出信号
	<-quit
	common.SysLog("Shutting down server...")

	// 停止定时任务服务
	scheduled.StopCronService()

	common.SysLog("Server stopped gracefully")
}
