package controller

import (
	"claude-code-relay/common"
	"claude-code-relay/constant"
	"claude-code-relay/model"
	"claude-code-relay/relay"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"io"
	"net/http"
	"strconv"
	"strings"
)

type ExchangeRequest struct {
	AuthorizationCode string `json:"authorization_code" binding:"required"`
	CallbackUrl       string `json:"callback_url" binding:"required"`
	ProxyURI          string `json:"proxy_uri" binding:"omitempty,url"`
	CodeVerifier      string `json:"code_verifier" binding:"required"`
	State             string `json:"state" binding:"required"`
}

// TestAccountRequest 测试账号请求参数
type TestAccountRequest struct {
	AccountID uint `json:"account_id" binding:"required" form:"account_id"`
}

// TestAccountResponse 测试账号响应结构
type TestAccountResponse struct {
	Success      bool   `json:"success"`
	Message      string `json:"message"`
	StatusCode   int    `json:"status_code,omitempty"`
	PlatformType string `json:"platform_type"`
}

// RequestContext 请求上下文信息
type RequestContext struct {
	APIKey           *model.ApiKey
	Body             []byte
	ModelName        string
	FilteredAccounts []model.Account
}

// GetOAuthURL 获取OAuth授权URL
func GetOAuthURL(c *gin.Context) {
	oauthHelper := common.NewOAuthHelper(nil)
	// 生成OAuth参数
	params, err := oauthHelper.GenerateOAuthParams()
	if err != nil {
		// 处理错误
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "操作成功",
		"code":    constant.Success,
		"data":    params,
	})
}

// ExchangeCode 验证授权码并返回token
func ExchangeCode(c *gin.Context) {
	var req ExchangeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误",
			"code":  constant.InvalidParams,
		})
		return
	}

	oauthHelper := common.NewOAuthHelper(nil)
	finalAuthCode, err := oauthHelper.ParseCallbackURL(req.AuthorizationCode)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的授权码",
			"code":  constant.InvalidParams,
		})
		return
	}

	// 生成访问令牌
	tokenResult, err := oauthHelper.ExchangeCodeForTokens(finalAuthCode, req.CodeVerifier, req.State, req.ProxyURI)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "生成访问令牌事变",
			"code":  constant.InternalServerError,
		})
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "操作成功",
		"code":    constant.Success,
		"data":    tokenResult,
	})
}

// prepareRequestContext 预处理请求上下文
func prepareRequestContext(c *gin.Context) (*RequestContext, bool) {
	// 从上下文中获取API Key的详细信息
	apiKey, _ := c.Get("api_key")
	keyInfo := apiKey.(*model.ApiKey)

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "请求参数异常",
			"code":    constant.InternalServerError,
		})
		return nil, false
	}

	modelName := gjson.GetBytes(body, "model").String()
	if modelName == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"message": "missing model",
			"code":    constant.InternalServerError,
		})
		return nil, false
	}

	// 根据API Key的分组ID查询可用账号列表
	accounts, err := model.GetAvailableAccountsByGroupID(keyInfo.GroupID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"message": "查询账号列表失败",
			"code":    constant.InternalServerError,
		})
		return nil, false
	}

	// 根据模型权限过滤账号
	filteredAccounts := filterAccountsByModelPermission(accounts, keyInfo, modelName)

	// 根据限额过滤账号
	filteredAccounts = filterAccountsByLimit(filteredAccounts)

	if len(filteredAccounts) == 0 {
		if len(accounts) == 0 {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "没有可用的账号",
				"code":    constant.NotFound,
			})
		} else {
			c.JSON(http.StatusForbidden, gin.H{
				"message": "没有权限访问模型: " + modelName,
				"code":    constant.Forbidden,
			})
		}
		return nil, false
	}

	return &RequestContext{
		APIKey:           keyInfo,
		Body:             body,
		ModelName:        modelName,
		FilteredAccounts: filteredAccounts,
	}, true
}

// GetMessages 获取对话消息
func GetMessages(c *gin.Context) {
	ctx, ok := prepareRequestContext(c)
	if !ok {
		return
	}

	// 选择第一个账号（已按优先级和使用次数排序）
	selectedAccount := ctx.FilteredAccounts[0]

	// 根据平台类型路由到不同的处理器
	switch selectedAccount.PlatformType {
	case constant.PlatformClaude:
		relay.HandleClaudeRequest(c, &selectedAccount, ctx.Body)
	case constant.PlatformClaudeConsole:
		relay.HandleClaudeConsoleRequest(c, &selectedAccount, ctx.Body)
	case constant.PlatformOpenAI:
		relay.HandleOpenAIRequest(c, &selectedAccount, ctx.Body)
	default:
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "不支持的平台类型: " + selectedAccount.PlatformType,
			"code":    constant.InvalidParams,
		})
	}
}

// TestGetMessages 测试账号连接
func TestGetMessages(c *gin.Context) {
	// 解析账号ID
	accountID, err := parseAccountID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": err.Error(),
			"code":    constant.InvalidParams,
		})
		return
	}

	// 获取账号信息
	account, err := model.GetAccountByID(accountID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "账号不存在",
			"code":    constant.NotFound,
		})
		return
	}

	// 执行账号测试
	testResult := executeAccountTest(account)

	// 处理不支持的平台类型
	if !testResult.Success && testResult.Message == "不支持的平台类型: "+account.PlatformType {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": testResult.Message,
			"code":    constant.InvalidParams,
		})
		return
	}

	// 返回测试结果
	c.JSON(http.StatusOK, gin.H{
		"message": "账号测试完成",
		"code":    constant.Success,
		"data":    testResult,
	})
}

// parseAccountID 解析账号ID（从URL参数或请求体）
func parseAccountID(c *gin.Context) (uint, error) {
	// 从URL参数中获取账号ID
	if accountIDStr := c.Param("id"); accountIDStr != "" {
		accountID, err := strconv.ParseUint(accountIDStr, 10, 32)
		if err != nil {
			return 0, fmt.Errorf("无效的账号ID")
		}
		return uint(accountID), nil
	}

	// 尝试从查询参数或表单中获取
	var req TestAccountRequest
	if err := c.ShouldBind(&req); err != nil {
		return 0, fmt.Errorf("参数错误: %s", err.Error())
	}
	return req.AccountID, nil
}

// executeAccountTest 执行账号测试并返回结果
func executeAccountTest(account *model.Account) TestAccountResponse {
	testResult := TestAccountResponse{
		PlatformType: account.PlatformType,
	}

	var statusCode int
	var errorMsg string

	// 根据平台类型调用不同的测试函数
	switch account.PlatformType {
	case constant.PlatformClaude:
		statusCode, errorMsg = relay.TestsHandleClaudeRequest(account)
	case constant.PlatformClaudeConsole:
		statusCode, errorMsg = relay.TestHandleClaudeConsoleRequest(account)
	case constant.PlatformOpenAI:
		statusCode, errorMsg = relay.TestHandleOpenAIRequest(account)
	default:
		return TestAccountResponse{
			Success:      false,
			Message:      "不支持的平台类型: " + account.PlatformType,
			PlatformType: account.PlatformType,
		}
	}

	// 设置测试结果
	testResult.StatusCode = statusCode
	if errorMsg != "" {
		testResult.Success = false
		testResult.Message = fmt.Sprintf("测试失败: %v", errorMsg)
	} else if statusCode == http.StatusOK {
		testResult.Success = true
		testResult.Message = "测试成功，账号连接正常"
	} else {
		testResult.Success = false
		testResult.Message = fmt.Sprintf("测试失败，HTTP状态码: %d", statusCode)
	}

	return testResult
}

// filterAccountsByModelPermission 根据模型权限过滤账号列表
func filterAccountsByModelPermission(accounts []model.Account, apiKey *model.ApiKey, modelName string) []model.Account {
	// 首先检查API Key的模型限制（优先级最高）
	if apiKey.ModelRestriction != "" {
		// 解析API Key允许的模型列表
		allowedModels := strings.Split(apiKey.ModelRestriction, ",")

		// 检查当前模型是否在API Key允许列表中
		isModelAllowed := false
		for _, allowedModel := range allowedModels {
			if strings.EqualFold(strings.TrimSpace(allowedModel), modelName) {
				isModelAllowed = true
				break
			}
		}

		// 如果API Key不允许此模型，直接返回空列表
		if !isModelAllowed {
			return []model.Account{}
		}
	}

	// API Key允许此模型或没有限制，继续检查账号级别的模型限制
	var filteredAccounts []model.Account
	for _, account := range accounts {
		// 如果账号没有模型限制，直接加入列表
		if account.ModelRestriction == "" {
			filteredAccounts = append(filteredAccounts, account)
			continue
		}

		// 解析账号允许的模型列表
		accountAllowedModels := strings.Split(account.ModelRestriction, ",")

		// 检查当前模型是否在账号允许列表中
		isAccountModelAllowed := false
		for _, allowedModel := range accountAllowedModels {
			if strings.EqualFold(strings.TrimSpace(allowedModel), modelName) {
				isAccountModelAllowed = true
				break
			}
		}

		// 如果账号允许此模型，加入列表
		if isAccountModelAllowed {
			filteredAccounts = append(filteredAccounts, account)
		}
	}

	return filteredAccounts
}

// filterAccountsByLimit 根据账号限额过滤账号列表
func filterAccountsByLimit(accounts []model.Account) []model.Account {
	var result []model.Account
	for _, account := range accounts {
		// 检查每日限额
		if account.DailyLimit > 0 && account.TodayTotalCost >= account.DailyLimit {
			continue
		}
		// 检查总限额
		if account.TotalLimit > 0 && account.TotalCost >= account.TotalLimit {
			continue
		}
		result = append(result, account)
	}
	return result
}

// GetCountTokens 获取token计数数据
func GetCountTokens(c *gin.Context) {
	ctx, ok := prepareRequestContext(c)
	if !ok {
		return
	}

	// 查找 PlatformType 为 constant.PlatformClaude 的账号
	var selectedAccount *model.Account
	for _, account := range ctx.FilteredAccounts {
		if account.PlatformType == constant.PlatformClaude {
			selectedAccount = &account
			break
		}
	}

	// 检查是否找到符合条件的账号
	if selectedAccount == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "没有可用的Claude平台账号",
			"code":    constant.NotFound,
		})
		return
	}

	// 调用Claude平台的GetCountTokens处理器
	relay.GetCountTokens(c, selectedAccount, ctx.Body)
}
