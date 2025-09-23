package relay

import (
	"bytes"
	"claude-code-relay/common"
	"claude-code-relay/model"
	"claude-code-relay/service"
	"crypto/tls"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"
)

// CodexTargetConfig Codex目标配置
type CodexTargetConfig struct {
	BaseURL   string
	ModelName string
}

// HandleCodexRequest 处理 Codex 请求的中转
func HandleCodexRequest(c *gin.Context, account *model.Account, requestBody []byte) {
	// 记录请求开始时间用于计算耗时
	startTime := time.Now()

	// 从上下文中获取API Key信息
	var apiKey *model.ApiKey
	if keyInfo, exists := c.Get("api_key"); exists {
		apiKey = keyInfo.(*model.ApiKey)
	}
	ctx := c.Request.Context()

	// 解析Claude请求
	var claudeReq ClaudeRequest
	if err := json.Unmarshal(requestBody, &claudeReq); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"type":    "json_parse_error",
				"message": "Failed to parse request JSON: " + err.Error(),
			},
		})
		return
	}

	// 直接使用账号配置的请求地址和默认模型
	if account.RequestURL == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": map[string]interface{}{
				"type":    "configuration_error",
				"message": "账号未配置请求地址",
			},
		})
		return
	}

	targetConfig := &CodexTargetConfig{
		BaseURL:   account.RequestURL, // 使用账号配置的请求地址
		ModelName: "code-davinci-002", // Codex默认模型
	}

	// 应用模型映射，复用OpenAI的映射逻辑
	mappedModelName := applyModelMapping(claudeReq.Model, account.ModelMapping, targetConfig.ModelName)

	// 转换Claude请求为OpenAI格式（Codex使用OpenAI格式）
	openaiReq := convertClaudeToOpenAI(claudeReq, mappedModelName)

	// 序列化OpenAI请求
	openaiBody, err := json.Marshal(openaiReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"type":    "json_marshal_error",
				"message": "Failed to marshal Codex request: " + err.Error(),
			},
		})
		return
	}

	// 创建Codex API请求
	codexURL := targetConfig.BaseURL + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, "POST", codexURL, bytes.NewBuffer(openaiBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"type":    "internal_server_error",
				"message": "Failed to create request: " + err.Error(),
			},
		})
		return
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+account.SecretKey)

	// 创建HTTP客户端
	httpClientTimeout, _ := time.ParseDuration(os.Getenv("HTTP_CLIENT_TIMEOUT") + "s")
	if httpClientTimeout == 0 {
		httpClientTimeout = 120 * time.Second
	}

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// 配置代理
	if account.ProxyURI != "" {
		proxyURL, err := url.Parse(account.ProxyURI)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": map[string]interface{}{
					"type":    "proxy_configuration_error",
					"message": "Invalid proxy URI: " + err.Error(),
				},
			})
			return
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	client := &http.Client{
		Timeout:   httpClientTimeout,
		Transport: transport,
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Codex API request failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": map[string]interface{}{
				"type":    "network_error",
				"message": "Failed to execute request: " + err.Error(),
			},
		})
		return
	}
	defer common.CloseIO(resp.Body)

	// 检查响应状态
	accountService := service.NewAccountService()
	if resp.StatusCode >= 400 {
		accountService.UpdateAccountStatus(account, resp.StatusCode, nil)
		bodyBytes, _ := io.ReadAll(resp.Body)
		c.Data(resp.StatusCode, "application/json", bodyBytes)
		return
	}

	// 统一使用流式响应处理（传递原始Claude模型名称，用于日志记录）
	handleStreamingResponse(c, resp, claudeReq.Model, claudeReq.Stream, account, apiKey, startTime)
}

// TestHandleCodexRequest 仅用于单元测试，返回状态码和响应内容
func TestHandleCodexRequest(account *model.Account) (int, string) {
	requestBody := common.GetTestRequestBody(100)

	// 解析Claude请求
	var claudeReq ClaudeRequest
	if err := json.Unmarshal([]byte(requestBody), &claudeReq); err != nil {
		return http.StatusBadRequest, "Failed to parse request JSON: " + err.Error()
	}

	// 检查账号配置
	if account.RequestURL == "" {
		return http.StatusBadRequest, "账号未配置请求地址"
	}

	targetConfig := &CodexTargetConfig{
		BaseURL:   account.RequestURL,
		ModelName: "code-davinci-002",
	}

	// 应用模型映射
	mappedModelName := applyModelMapping(claudeReq.Model, account.ModelMapping, targetConfig.ModelName)

	// 转换Claude请求为OpenAI格式
	openaiReq := convertClaudeToOpenAI(claudeReq, mappedModelName)

	// 序列化OpenAI请求
	openaiBody, err := json.Marshal(openaiReq)
	if err != nil {
		return http.StatusInternalServerError, "Failed to marshal Codex request: " + err.Error()
	}

	// 创建Codex API请求
	codexURL := targetConfig.BaseURL + "/chat/completions"
	req, err := http.NewRequest("POST", codexURL, bytes.NewBuffer(openaiBody))
	if err != nil {
		return http.StatusInternalServerError, "Failed to create request: " + err.Error()
	}

	// 设置请求头
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+account.SecretKey)

	// 创建HTTP客户端
	httpClientTimeout := 30 * time.Second
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	// 配置代理
	if account.ProxyURI != "" {
		proxyURL, err := url.Parse(account.ProxyURI)
		if err != nil {
			return http.StatusInternalServerError, "Invalid proxy URI: " + err.Error()
		}
		transport.Proxy = http.ProxyURL(proxyURL)
	}

	client := &http.Client{
		Timeout:   httpClientTimeout,
		Transport: transport,
	}

	// 发送请求
	resp, err := client.Do(req)
	if err != nil {
		return http.StatusInternalServerError, "Request failed: " + err.Error()
	}
	defer common.CloseIO(resp.Body)

	// 读取错误响应内容
	if resp.StatusCode >= 400 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return resp.StatusCode, string(bodyBytes)
	}

	return resp.StatusCode, ""
}