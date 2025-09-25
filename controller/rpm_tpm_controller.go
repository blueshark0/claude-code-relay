package controller

import (
	"claude-code-relay/constant"
	"claude-code-relay/model"
	"claude-code-relay/service"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// GetApiKeyRpmTpmStats 获取API Key的RPM/TPM统计信息
func GetApiKeyRpmTpmStats(c *gin.Context) {
	// 获取API Key ID
	idStr := c.Param("id")
	apiKeyID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的API Key ID",
			"code":  constant.InvalidParams,
		})
		return
	}

	// 获取当前用户信息
	user := c.MustGet("user").(*model.User)

	// 验证API Key权限
	apiKey, err := model.GetApiKeyById(uint(apiKeyID), user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "API Key不存在或无权访问",
			"code":  constant.ResourceNotFound,
		})
		return
	}

	// 获取RPM/TPM统计
	rpmTpmService := service.NewRpmTpmService()
	stats, err := rpmTpmService.GetApiKeyCurrentStats(uint(apiKeyID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取统计信息失败: " + err.Error(),
			"code":  constant.InternalServerError,
		})
		return
	}

	// 构建响应数据
	response := &model.ApiKeyRpmTpmStats{
		ApiKeyID:            apiKey.ID,
		CurrentRpm:          stats.Rpm,
		CurrentTpm:          stats.Tpm,
		MaxRpm:              apiKey.MaxRpm,
		MaxTpm:              apiKey.MaxTpm,
		RpmLimit:            apiKey.RpmLimit,
		TpmLimit:            apiKey.TpmLimit,
		RpmWarningThreshold: apiKey.RpmWarningThreshold,
		TpmWarningThreshold: apiKey.TpmWarningThreshold,
		RpmUsagePercentage:  stats.RpmUsagePercentage,
		TpmUsagePercentage:  stats.TpmUsagePercentage,
		IsRpmLimited:        stats.IsRpmLimited,
		IsTpmLimited:        stats.IsTpmLimited,
		RateLimitEndTime:    apiKey.RateLimitEndTime,
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
		"code": constant.Success,
	})
}

// GetAccountRpmTpmStats 获取Account的RPM/TPM统计信息
func GetAccountRpmTpmStats(c *gin.Context) {
	// 获取Account ID
	idStr := c.Param("id")
	accountID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的Account ID",
			"code":  constant.InvalidParams,
		})
		return
	}

	// 获取当前用户信息
	user := c.MustGet("user").(*model.User)
	var userID *uint
	if user.Role != "admin" {
		userID = &user.ID
	}

	// 验证Account权限
	accountService := service.NewAccountService()
	account, err := accountService.GetAccountByID(uint(accountID), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Account不存在或无权访问",
			"code":  constant.ResourceNotFound,
		})
		return
	}

	// 获取RPM/TPM统计
	rpmTpmService := service.NewRpmTpmService()
	stats, err := rpmTpmService.GetAccountCurrentStats(uint(accountID))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取统计信息失败: " + err.Error(),
			"code":  constant.InternalServerError,
		})
		return
	}

	// 构建响应数据
	response := &model.AccountRpmTpmStats{
		AccountID:           account.ID,
		CurrentRpm:          stats.Rpm,
		CurrentTpm:          stats.Tpm,
		MaxRpm:              account.MaxRpm,
		MaxTpm:              account.MaxTpm,
		RpmLimit:            account.RpmLimit,
		TpmLimit:            account.TpmLimit,
		RpmWarningThreshold: account.RpmWarningThreshold,
		TpmWarningThreshold: account.TpmWarningThreshold,
		RpmUsagePercentage:  stats.RpmUsagePercentage,
		TpmUsagePercentage:  stats.TpmUsagePercentage,
		IsRpmLimited:        stats.IsRpmLimited,
		IsTpmLimited:        stats.IsTpmLimited,
		RateLimitEndTime:    account.RateLimitEndTime,
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
		"code": constant.Success,
	})
}

// UpdateApiKeyRpmTpmLimits 更新API Key的RPM/TPM限制
func UpdateApiKeyRpmTpmLimits(c *gin.Context) {
	// 获取API Key ID
	idStr := c.Param("id")
	apiKeyID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的API Key ID",
			"code":  constant.InvalidParams,
		})
		return
	}

	// 解析请求体
	var req model.UpdateApiKeyRpmTpmLimitsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
			"code":  constant.InvalidParams,
		})
		return
	}

	// 获取当前用户信息
	user := c.MustGet("user").(*model.User)

	// 验证API Key权限
	apiKey, err := model.GetApiKeyById(uint(apiKeyID), user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "API Key不存在或无权访问",
			"code":  constant.ResourceNotFound,
		})
		return
	}

	// 更新限制参数
	updates := make(map[string]interface{})
	if req.RpmLimit != nil {
		updates["rpm_limit"] = *req.RpmLimit
	}
	if req.TpmLimit != nil {
		updates["tpm_limit"] = *req.TpmLimit
	}
	if req.RpmWarningThreshold != nil {
		updates["rpm_warning_threshold"] = *req.RpmWarningThreshold
	}
	if req.TpmWarningThreshold != nil {
		updates["tpm_warning_threshold"] = *req.TpmWarningThreshold
	}

	// 执行更新
	err = model.DB.Model(&model.ApiKey{}).
		Where("id = ? AND user_id = ?", apiKeyID, user.ID).
		Updates(updates).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "更新失败: " + err.Error(),
			"code":  constant.InternalServerError,
		})
		return
	}

	// 清理API Key缓存
	model.ClearApiKeyCache(apiKey.Key)

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"code":    constant.Success,
	})
}

// UpdateAccountRpmTpmLimits 更新Account的RPM/TPM限制
func UpdateAccountRpmTpmLimits(c *gin.Context) {
	// 获取Account ID
	idStr := c.Param("id")
	accountID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的Account ID",
			"code":  constant.InvalidParams,
		})
		return
	}

	// 解析请求体
	var req model.UpdateAccountRpmTpmLimitsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
			"code":  constant.InvalidParams,
		})
		return
	}

	// 获取当前用户信息
	user := c.MustGet("user").(*model.User)
	var userID *uint
	if user.Role != "admin" {
		userID = &user.ID
	}

	// 验证Account权限
	accountService := service.NewAccountService()
	_, err = accountService.GetAccountByID(uint(accountID), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Account不存在或无权访问",
			"code":  constant.ResourceNotFound,
		})
		return
	}

	// 更新限制参数
	updates := make(map[string]interface{})
	if req.RpmLimit != nil {
		updates["rpm_limit"] = *req.RpmLimit
	}
	if req.TpmLimit != nil {
		updates["tpm_limit"] = *req.TpmLimit
	}
	if req.RpmWarningThreshold != nil {
		updates["rpm_warning_threshold"] = *req.RpmWarningThreshold
	}
	if req.TpmWarningThreshold != nil {
		updates["tpm_warning_threshold"] = *req.TpmWarningThreshold
	}

	// 构建查询条件
	query := model.DB.Model(&model.Account{}).Where("id = ?", accountID)
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}

	// 执行更新
	err = query.Updates(updates).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "更新失败: " + err.Error(),
			"code":  constant.InternalServerError,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "更新成功",
		"code":    constant.Success,
	})
}

// GetApiKeyRpmTpmHistory 获取API Key的RPM/TPM历史数据
func GetApiKeyRpmTpmHistory(c *gin.Context) {
	// 获取API Key ID
	idStr := c.Param("id")
	apiKeyID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的API Key ID",
			"code":  constant.InvalidParams,
		})
		return
	}

	// 解析查询参数
	var req model.ApiKeyRpmTpmHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "查询参数错误: " + err.Error(),
			"code":  constant.InvalidParams,
		})
		return
	}

	// 设置默认值
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 20
	}

	// 默认查询最近24小时的数据
	if req.StartTime == nil && req.EndTime == nil {
		now := time.Now()
		endTime := model.Time(now)
		startTime := model.Time(now.Add(-24 * time.Hour))
		req.StartTime = &startTime
		req.EndTime = &endTime
	}

	// 获取当前用户信息
	user := c.MustGet("user").(*model.User)

	// 验证API Key权限
	_, err = model.GetApiKeyById(uint(apiKeyID), user.ID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "API Key不存在或无权访问",
			"code":  constant.ResourceNotFound,
		})
		return
	}

	// 查询历史数据
	var history []model.ApiKeyRpmTpmHistory
	var total int64

	query := model.DB.Model(&model.ApiKeyRpmTpmHistory{}).
		Where("api_key_id = ?", apiKeyID)

	if req.StartTime != nil {
		query = query.Where("minute_timestamp >= ?", req.StartTime)
	}
	if req.EndTime != nil {
		query = query.Where("minute_timestamp <= ?", req.EndTime)
	}

	// 获取总数
	err = query.Count(&total).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "查询失败: " + err.Error(),
			"code":  constant.InternalServerError,
		})
		return
	}

	// 分页查询
	offset := (req.Page - 1) * req.Limit
	err = query.Offset(offset).Limit(req.Limit).
		Order("minute_timestamp DESC").
		Find(&history).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "查询失败: " + err.Error(),
			"code":  constant.InternalServerError,
		})
		return
	}

	// 构建响应
	response := &model.ApiKeyRpmTpmHistoryResponse{
		History: history,
		Total:   total,
		Page:    req.Page,
		Limit:   req.Limit,
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
		"code": constant.Success,
	})
}

// GetAccountRpmTpmHistory 获取Account的RPM/TPM历史数据
func GetAccountRpmTpmHistory(c *gin.Context) {
	// 获取Account ID
	idStr := c.Param("id")
	accountID, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的Account ID",
			"code":  constant.InvalidParams,
		})
		return
	}

	// 解析查询参数
	var req model.AccountRpmTpmHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "查询参数错误: " + err.Error(),
			"code":  constant.InvalidParams,
		})
		return
	}

	// 设置默认值
	if req.Page < 1 {
		req.Page = 1
	}
	if req.Limit < 1 || req.Limit > 100 {
		req.Limit = 20
	}

	// 默认查询最近24小时的数据
	if req.StartTime == nil && req.EndTime == nil {
		now := time.Now()
		endTime := model.Time(now)
		startTime := model.Time(now.Add(-24 * time.Hour))
		req.StartTime = &startTime
		req.EndTime = &endTime
	}

	// 获取当前用户信息
	user := c.MustGet("user").(*model.User)
	var userID *uint
	if user.Role != "admin" {
		userID = &user.ID
	}

	// 验证Account权限
	accountService := service.NewAccountService()
	_, err = accountService.GetAccountByID(uint(accountID), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "Account不存在或无权访问",
			"code":  constant.ResourceNotFound,
		})
		return
	}

	// 查询历史数据
	var history []model.AccountRpmTpmHistory
	var total int64

	query := model.DB.Model(&model.AccountRpmTpmHistory{}).
		Where("account_id = ?", accountID)

	if req.StartTime != nil {
		query = query.Where("minute_timestamp >= ?", req.StartTime)
	}
	if req.EndTime != nil {
		query = query.Where("minute_timestamp <= ?", req.EndTime)
	}

	// 获取总数
	err = query.Count(&total).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "查询失败: " + err.Error(),
			"code":  constant.InternalServerError,
		})
		return
	}

	// 分页查询
	offset := (req.Page - 1) * req.Limit
	err = query.Offset(offset).Limit(req.Limit).
		Order("minute_timestamp DESC").
		Find(&history).Error
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "查询失败: " + err.Error(),
			"code":  constant.InternalServerError,
		})
		return
	}

	// 构建响应
	response := &model.AccountRpmTpmHistoryResponse{
		History: history,
		Total:   total,
		Page:    req.Page,
		Limit:   req.Limit,
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
		"code": constant.Success,
	})
}

// GetRpmTpmDashboard 获取RPM/TPM仪表盘数据
func GetRpmTpmDashboard(c *gin.Context) {
	// 获取当前用户信息
	user := c.MustGet("user").(*model.User)

	// 构建仪表盘数据
	dashboard := gin.H{
		"summary": gin.H{
			"total_api_keys": 0,
			"total_accounts": 0,
			"active_limits":  0,
			"current_alerts": 0,
		},
		"top_api_keys":  []gin.H{},
		"top_accounts":  []gin.H{},
		"recent_alerts": []gin.H{},
	}

	// 如果是管理员，查询系统整体数据
	if user.Role == "admin" {
		// 查询API Key总数
		var apiKeyCount int64
		model.DB.Model(&model.ApiKey{}).Where("status = 1").Count(&apiKeyCount)
		dashboard["summary"].(gin.H)["total_api_keys"] = apiKeyCount

		// 查询Account总数
		var accountCount int64
		model.DB.Model(&model.Account{}).Where("active_status = 1").Count(&accountCount)
		dashboard["summary"].(gin.H)["total_accounts"] = accountCount

		// 查询有限制的数量
		var limitCount int64
		model.DB.Model(&model.ApiKey{}).
			Where("status = 1 AND (rpm_limit > 0 OR tpm_limit > 0)").
			Count(&limitCount)
		var accountLimitCount int64
		model.DB.Model(&model.Account{}).
			Where("active_status = 1 AND (rpm_limit > 0 OR tpm_limit > 0)").
			Count(&accountLimitCount)
		dashboard["summary"].(gin.H)["active_limits"] = limitCount + accountLimitCount

		// 查询Top API Keys（按当前RPM排序）
		var topApiKeys []struct {
			ID         uint   `json:"id"`
			Name       string `json:"name"`
			CurrentRpm int    `json:"current_rpm"`
			CurrentTpm int    `json:"current_tpm"`
			RpmLimit   int    `json:"rpm_limit"`
			TpmLimit   int    `json:"tpm_limit"`
		}
		model.DB.Model(&model.ApiKey{}).
			Select("id, name, current_rpm, current_tpm, rpm_limit, tpm_limit").
			Where("status = 1").
			Order("current_rpm DESC").
			Limit(5).
			Scan(&topApiKeys)
		dashboard["top_api_keys"] = topApiKeys

		// 查询Top Accounts（按当前RPM排序）
		var topAccounts []struct {
			ID         uint   `json:"id"`
			Name       string `json:"name"`
			CurrentRpm int    `json:"current_rpm"`
			CurrentTpm int    `json:"current_tpm"`
			RpmLimit   int    `json:"rpm_limit"`
			TpmLimit   int    `json:"tpm_limit"`
		}
		model.DB.Model(&model.Account{}).
			Select("id, name, current_rpm, current_tpm, rpm_limit, tpm_limit").
			Where("active_status = 1").
			Order("current_rpm DESC").
			Limit(5).
			Scan(&topAccounts)
		dashboard["top_accounts"] = topAccounts
	} else {
		// 普通用户只能查看自己的数据
		var apiKeyCount int64
		model.DB.Model(&model.ApiKey{}).Where("user_id = ? AND status = 1", user.ID).Count(&apiKeyCount)
		dashboard["summary"].(gin.H)["total_api_keys"] = apiKeyCount

		var accountCount int64
		model.DB.Model(&model.Account{}).Where("user_id = ? AND active_status = 1", user.ID).Count(&accountCount)
		dashboard["summary"].(gin.H)["total_accounts"] = accountCount

		// 查询用户的Top API Keys
		var topApiKeys []struct {
			ID         uint   `json:"id"`
			Name       string `json:"name"`
			CurrentRpm int    `json:"current_rpm"`
			CurrentTpm int    `json:"current_tpm"`
			RpmLimit   int    `json:"rpm_limit"`
			TpmLimit   int    `json:"tpm_limit"`
		}
		model.DB.Model(&model.ApiKey{}).
			Select("id, name, current_rpm, current_tpm, rpm_limit, tpm_limit").
			Where("user_id = ? AND status = 1", user.ID).
			Order("current_rpm DESC").
			Limit(5).
			Scan(&topApiKeys)
		dashboard["top_api_keys"] = topApiKeys

		// 查询用户的Top Accounts
		var topAccounts []struct {
			ID         uint   `json:"id"`
			Name       string `json:"name"`
			CurrentRpm int    `json:"current_rpm"`
			CurrentTpm int    `json:"current_tpm"`
			RpmLimit   int    `json:"rpm_limit"`
			TpmLimit   int    `json:"tpm_limit"`
		}
		model.DB.Model(&model.Account{}).
			Select("id, name, current_rpm, current_tpm, rpm_limit, tpm_limit").
			Where("user_id = ? AND active_status = 1", user.ID).
			Order("current_rpm DESC").
			Limit(5).
			Scan(&topAccounts)
		dashboard["top_accounts"] = topAccounts
	}

	c.JSON(http.StatusOK, gin.H{
		"data": dashboard,
		"code": constant.Success,
	})
}

// GetSystemRpmTpmStats 获取系统总体RPM/TPM统计信息
func GetSystemRpmTpmStats(c *gin.Context) {
	// 获取当前用户信息
	user := c.MustGet("user").(*model.User)

	// 仅管理员可查看系统总体统计
	if user.Role != "admin" {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "权限不足",
			"code":  constant.PermissionDenied,
		})
		return
	}

	// 获取系统RPM/TPM统计
	rpmTpmService := service.NewRpmTpmService()
	stats, err := rpmTpmService.GetSystemCurrentStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取系统统计信息失败: " + err.Error(),
			"code":  constant.InternalServerError,
		})
		return
	}

	// 构建响应数据
	response := gin.H{
		"rpm":                  stats.Rpm,
		"tpm":                  stats.Tpm,
		"rpm_usage_percentage": stats.RpmUsagePercentage,
		"tpm_usage_percentage": stats.TpmUsagePercentage,
		"is_rpm_limited":       stats.IsRpmLimited,
		"is_tpm_limited":       stats.IsTpmLimited,
		"timestamp":            time.Now(),
	}

	c.JSON(http.StatusOK, gin.H{
		"data": response,
		"code": constant.Success,
	})
}
