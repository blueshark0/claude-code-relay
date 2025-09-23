package router

import (
	"claude-code-relay/controller"
	"claude-code-relay/middleware"
	"embed"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func SetAPIRouter(server *gin.Engine, embeddedFS embed.FS, staticFileSystem http.FileSystem, enableStatic bool) {
	// 健康检查
	server.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Unix(),
		})
	})

	// Claude Code 路由
	claude := server.Group("/claude-code")
	claude.Use(middleware.ClaudeCodeAuth())
	{
		// 对话接口
		claude.POST("/v1/messages", controller.GetMessages)
		// 使用量统计接口
		claude.POST("/v1/messages/count_tokens", controller.GetCountTokens)
	}

	// API路由组
	api := server.Group("/api/v1")
	api.Use(middleware.RateLimit(300, time.Minute))
	{
		// 公开接口
		auth := api.Group("/auth")
		{
			auth.POST("/login", controller.Login)
			auth.POST("/register", controller.Register)
			auth.POST("/send-verification-code", controller.SendVerificationCode)
			auth.GET("/api-key", controller.GetApiKeyInfo)          // 根据API Key查询统计信息（公开接口）
			auth.GET("/api-key/:api_key", controller.GetApiKeyInfo) // 支持URL路径参数方式
		}

		// 系统状态
		api.GET("/status", controller.GetStatus)

		// 需要认证的接口
		authenticated := api.Group("")
		authenticated.Use(middleware.Auth())
		{
			// 用户相关
			user := authenticated.Group("/user")
			{
				user.GET("/profile", controller.GetProfile)
				user.PUT("/profile", controller.UpdateProfile)
				user.PUT("/change-email", controller.ChangeEmail)
				user.PUT("/change-password", controller.ChangePassword)
			}

			// 菜单相关
			authenticated.GET("/menu-list", controller.GetMenuList)

			// 分组相关
			group := authenticated.Group("/groups")
			{
				group.GET("/list", controller.GetGroups)            // 获取分组列表
				group.GET("/all", controller.GetAllGroups)          // 获取所有分组（用于下拉选择）
				group.POST("/create", controller.CreateGroup)       // 创建分组
				group.GET("/detail/:id", controller.GetGroup)       // 获取分组详情
				group.PUT("/update/:id", controller.UpdateGroup)    // 更新分组
				group.DELETE("/delete/:id", controller.DeleteGroup) // 删除分组
			}

			// 账号管理相关
			account := authenticated.Group("/accounts")
			{
				account.GET("/list", controller.GetAccountList)                                  // 获取账号列表
				account.POST("/create", controller.CreateAccount)                                // 创建账号
				account.GET("/detail/:id", controller.GetAccount)                                // 获取账号详情
				account.PUT("/update/:id", controller.UpdateAccount)                             // 更新账号
				account.DELETE("/delete/:id", controller.DeleteAccount)                          // 删除账号
				account.PUT("/update-active-status/:id", controller.UpdateAccountActiveStatus)   // 更新账号激活状态
				account.PUT("/update-current-status/:id", controller.UpdateAccountCurrentStatus) // 更新账号当前状态
				account.POST("/test/:id", controller.TestGetMessages)                            // 测试账号连通性
			}

			// Claude OAuth 相关
			oauth := authenticated.Group("/oauth")
			{
				oauth.GET("/generate-auth-url", controller.GetOAuthURL) // 获取OAuth授权URL
				oauth.POST("/exchange-code", controller.ExchangeCode)   // 验证授权码并获取token
			}

			// API Key 相关
			apikey := authenticated.Group("/api-keys")
			{
				apikey.GET("/list", controller.GetApiKeys)                      // 获取API Key列表
				apikey.POST("/create", controller.CreateApiKey)                 // 创建API Key
				apikey.POST("/autocreate", controller.AutoCreateApiKey)         // 自动创建API Key（自动生成名称）
				apikey.GET("/detail/:id", controller.GetApiKey)                 // 获取API Key详情
				apikey.PUT("/update/:id", controller.UpdateApiKey)              // 更新API Key
				apikey.PUT("/update-status/:id", controller.UpdateApiKeyStatus) // 更新API Key状态
				apikey.DELETE("/delete/:id", controller.DeleteApiKey)           // 删除API Key
			}

			// 日志相关（用户接口）
			logs := authenticated.Group("/logs")
			{
				logs.GET("/my", controller.GetMyLogs)                   // 获取当前用户的日志记录
				logs.GET("/stats/my", controller.GetMyLogStats)         // 获取当前用户的日志统计
				logs.GET("/usage-stats/my", controller.GetMyUsageStats) // 获取当前用户的使用统计
				logs.GET("/detail/:id", controller.GetLogById)          // 获取日志详情
			}

			// 仪表盘数据接口
			authenticated.GET("/dashboard/stats", controller.GetDashboardStats) // 获取仪表盘统计数据

			// 管理员接口
			admin := authenticated.Group("/admin")
			admin.Use(middleware.AdminAuth())
			{
				admin.GET("/users", controller.GetUsers)
				admin.POST("/users", controller.AdminCreateUser)
				admin.PUT("/users/:id/status", controller.AdminUpdateUserStatus)
				admin.GET("/logs", controller.GetApiLogs)
				admin.GET("/dashboard", controller.GetDashboard)

				// 日志管理（管理员专用）
				adminLogs := admin.Group("/logs")
				{
					adminLogs.GET("/list", controller.GetLogs)                 // 获取所有日志列表（支持筛选）
					adminLogs.GET("/stats", controller.GetLogStats)            // 获取日志统计（支持指定用户）
					adminLogs.GET("/usage-stats", controller.GetUsageStats)    // 获取使用统计（管理员可查看所有用户）
					adminLogs.GET("/detail/:id", controller.GetLogById)        // 获取日志详情
					adminLogs.DELETE("/delete/:id", controller.DeleteLogById)  // 删除指定日志
					adminLogs.DELETE("/cleanup", controller.DeleteExpiredLogs) // 删除过期日志
				}

				// 定时任务测试接口（管理员专用）
				admin.POST("/test/reset-stats", controller.ManualResetStats) // 手动重置统计数据
				admin.POST("/test/clean-logs", controller.ManualCleanLogs)   // 手动清理过期日志
			}
		}
	}

	// 前端路由处理 - 只在启用静态文件服务时处理
	if enableStatic {
		server.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path

			// 如果是 API 请求，返回404
			if strings.HasPrefix(path, "/api/") || strings.HasPrefix(path, "/health") {
				c.JSON(http.StatusNotFound, gin.H{"error": "API not found"})
				return
			}

			// 如果是 claude-code 路径但不是已定义的路由，返回空白200响应
			if strings.HasPrefix(path, "/claude-code/") {
				c.Status(200)
				return
			}

			// 尝试从嵌入文件系统中获取静态文件
			if staticFileSystem != nil {
				file, err := staticFileSystem.Open(strings.TrimPrefix(path, "/"))
				if err == nil {
					if closeErr := file.Close(); closeErr != nil {
						// 记录关闭文件错误，但不影响继续处理
					}
					// 成功找到文件，使用文件服务器处理
					http.FileServer(staticFileSystem).ServeHTTP(c.Writer, c.Request)
					return
				}
			}

			// 如果找不到静态文件，返回 index.html (SPA路由)
			indexContent, err := embeddedFS.ReadFile("web/dist/index.html")
			if err != nil {
				c.String(http.StatusInternalServerError, "Failed to load index.html: "+err.Error())
				return
			}
			c.Data(http.StatusOK, "text/html; charset=utf-8", indexContent)
		})
	} else {
		// 纯API模式，对非API路径返回404
		server.NoRoute(func(c *gin.Context) {
			path := c.Request.URL.Path
			// 如果是 claude-code 路径但不是已定义的路由，返回空白200响应
			if strings.HasPrefix(path, "/claude-code/") {
				c.Status(200)
				return
			}
			c.JSON(http.StatusNotFound, gin.H{"error": "Not found - API only mode"})
		})
	}
}
