package service

import (
	"claude-code-relay/common"
	"claude-code-relay/model"
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

// RpmTpmService RPM/TPM统计服务
type RpmTpmService struct {
	redisClient *redis.Client
}

// NewRpmTpmService 创建RPM/TPM统计服务实例
func NewRpmTpmService() *RpmTpmService {
	return &RpmTpmService{
		redisClient: common.RDB,
	}
}

// RpmTpmResult 统计结果结构
type RpmTpmResult struct {
	Rpm                int        `json:"rpm"`
	Tpm                int        `json:"tpm"`
	RpmUsagePercentage float64    `json:"rpm_usage_percentage"`
	TpmUsagePercentage float64    `json:"tpm_usage_percentage"`
	IsRpmLimited       bool       `json:"is_rpm_limited"`
	IsTpmLimited       bool       `json:"is_tpm_limited"`
	RateLimitEndTime   *time.Time `json:"rate_limit_end_time"`
}

// UpdateBothLevelsStats 同时更新API Key、Account和系统级别的RPM/TPM统计
func (s *RpmTpmService) UpdateBothLevelsStats(apiKeyID, accountID uint, usage *common.TokenUsage) error {
	now := time.Now()
	// 使用5秒精度的时间键
	second := now.Second()
	roundedSecond := (second / 5) * 5
	timeKey := fmt.Sprintf("%s:%02d", now.Format("2006-01-02-15:04"), roundedSecond)

	// 计算总token数
	totalTokens := usage.InputTokens + usage.OutputTokens + usage.CacheReadInputTokens + usage.CacheCreationInputTokens

	// 并发更新三个级别的统计
	err1 := s.updateApiKeyStats(apiKeyID, timeKey, totalTokens)
	err2 := s.updateAccountStats(accountID, timeKey, totalTokens)
	err3 := s.updateSystemStats(timeKey, totalTokens)

	if err1 != nil {
		return fmt.Errorf("更新API Key统计失败: %v", err1)
	}
	if err2 != nil {
		return fmt.Errorf("更新Account统计失败: %v", err2)
	}
	if err3 != nil {
		return fmt.Errorf("更新系统统计失败: %v", err3)
	}

	return nil
}

// updateApiKeyStats 更新API Key级别统计
func (s *RpmTpmService) updateApiKeyStats(apiKeyID uint, timeKey string, tokens int) error {
	ctx := context.Background()

	// Redis键名 - 使用5秒精度
	requestsKey := fmt.Sprintf("api_key:%d:requests:%s", apiKeyID, timeKey)
	tokensKey := fmt.Sprintf("api_key:%d:tokens:%s", apiKeyID, timeKey)

	// 使用Pipeline提高性能
	pipe := s.redisClient.Pipeline()

	// 增加请求计数和token计数
	pipe.Incr(ctx, requestsKey)
	pipe.Expire(ctx, requestsKey, 90*time.Second) // TTL 90秒，保证60秒窗口数据完整
	pipe.IncrBy(ctx, tokensKey, int64(tokens))
	pipe.Expire(ctx, tokensKey, 90*time.Second)

	_, err := pipe.Exec(ctx)
	return err
}

// updateAccountStats 更新Account级别统计
func (s *RpmTpmService) updateAccountStats(accountID uint, timeKey string, tokens int) error {
	ctx := context.Background()

	// Redis键名 - 使用5秒精度
	requestsKey := fmt.Sprintf("account:%d:requests:%s", accountID, timeKey)
	tokensKey := fmt.Sprintf("account:%d:tokens:%s", accountID, timeKey)

	// 使用Pipeline提高性能
	pipe := s.redisClient.Pipeline()

	// 增加请求计数和token计数
	pipe.Incr(ctx, requestsKey)
	pipe.Expire(ctx, requestsKey, 90*time.Second) // TTL 90秒
	pipe.IncrBy(ctx, tokensKey, int64(tokens))
	pipe.Expire(ctx, tokensKey, 90*time.Second)

	_, err := pipe.Exec(ctx)
	return err
}

// updateSystemStats 更新系统级别统计
func (s *RpmTpmService) updateSystemStats(timeKey string, tokens int) error {
	ctx := context.Background()

	// Redis键名 - 系统级别统计
	requestsKey := fmt.Sprintf("system:requests:%s", timeKey)
	tokensKey := fmt.Sprintf("system:tokens:%s", timeKey)

	// 使用Pipeline提高性能
	pipe := s.redisClient.Pipeline()

	// 增加请求计数和token计数
	pipe.Incr(ctx, requestsKey)
	pipe.Expire(ctx, requestsKey, 90*time.Second) // TTL 90秒
	pipe.IncrBy(ctx, tokensKey, int64(tokens))
	pipe.Expire(ctx, tokensKey, 90*time.Second)

	_, err := pipe.Exec(ctx)
	return err
}

// GetApiKeyCurrentStats 获取API Key当前RPM/TPM统计
func (s *RpmTpmService) GetApiKeyCurrentStats(apiKeyID uint) (*RpmTpmResult, error) {
	// 获取API Key配置
	apiKey, err := model.GetApiKeyById(apiKeyID, 0) // 0表示不验证用户权限
	if err != nil {
		return nil, fmt.Errorf("获取API Key失败: %v", err)
	}

	// 计算当前RPM/TPM
	rpm, tpm, err := s.calculateApiKeyCurrentRpmTpm(apiKeyID)
	if err != nil {
		return nil, err
	}

	// 计算使用率和限流状态
	result := &RpmTpmResult{
		Rpm: rpm,
		Tpm: tpm,
	}

	// 计算RPM使用率
	if apiKey.RpmLimit > 0 {
		result.RpmUsagePercentage = float64(rpm) / float64(apiKey.RpmLimit) * 100
		result.IsRpmLimited = rpm >= apiKey.RpmLimit
	}

	// 计算TPM使用率
	if apiKey.TpmLimit > 0 {
		result.TpmUsagePercentage = float64(tpm) / float64(apiKey.TpmLimit) * 100
		result.IsTpmLimited = tpm >= apiKey.TpmLimit
	}

	// 检查限流结束时间
	if apiKey.RateLimitEndTime != nil {
		endTime := time.Time(*apiKey.RateLimitEndTime)
		if endTime.After(time.Now()) {
			result.RateLimitEndTime = &endTime
		}
	}

	return result, nil
}

// GetAccountCurrentStats 获取Account当前RPM/TPM统计
func (s *RpmTpmService) GetAccountCurrentStats(accountID uint) (*RpmTpmResult, error) {
	// 获取Account配置
	account, err := model.GetAccountByID(accountID)
	if err != nil {
		return nil, fmt.Errorf("获取Account失败: %v", err)
	}

	// 计算当前RPM/TPM
	rpm, tpm, err := s.calculateAccountCurrentRpmTpm(accountID)
	if err != nil {
		return nil, err
	}

	// 计算使用率和限流状态
	result := &RpmTpmResult{
		Rpm: rpm,
		Tpm: tpm,
	}

	// 计算RPM使用率
	if account.RpmLimit > 0 {
		result.RpmUsagePercentage = float64(rpm) / float64(account.RpmLimit) * 100
		result.IsRpmLimited = rpm >= account.RpmLimit
	}

	// 计算TPM使用率
	if account.TpmLimit > 0 {
		result.TpmUsagePercentage = float64(tpm) / float64(account.TpmLimit) * 100
		result.IsTpmLimited = tpm >= account.TpmLimit
	}

	// 检查限流结束时间
	if account.RateLimitEndTime != nil {
		endTime := time.Time(*account.RateLimitEndTime)
		if endTime.After(time.Now()) {
			result.RateLimitEndTime = &endTime
		}
	}

	return result, nil
}

// calculateApiKeyCurrentRpmTpm 计算API Key当前RPM/TPM（60秒滑动窗口，5秒精度）
func (s *RpmTpmService) calculateApiKeyCurrentRpmTpm(apiKeyID uint) (int, int, error) {
	ctx := context.Background()
	now := time.Now()

	// 生成12个5秒时间片的key（60秒窗口）
	var requestKeys, tokenKeys []string
	for i := 0; i < 12; i++ {
		t := now.Add(-time.Duration(i*5) * time.Second)
		second := (t.Second() / 5) * 5
		timeKey := fmt.Sprintf("%s:%02d", t.Format("2006-01-02-15:04"), second)
		requestKeys = append(requestKeys, fmt.Sprintf("api_key:%d:requests:%s", apiKeyID, timeKey))
		tokenKeys = append(tokenKeys, fmt.Sprintf("api_key:%d:tokens:%s", apiKeyID, timeKey))
	}

	// 并发获取所有键的值
	pipe := s.redisClient.Pipeline()

	// 添加所有GET命令
	requestCmds := make([]*redis.StringCmd, len(requestKeys))
	tokenCmds := make([]*redis.StringCmd, len(tokenKeys))

	for i, key := range requestKeys {
		requestCmds[i] = pipe.Get(ctx, key)
	}
	for i, key := range tokenKeys {
		tokenCmds[i] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return 0, 0, err
	}

	// 累计结果
	totalRequests := 0
	totalTokens := 0

	for _, cmd := range requestCmds {
		val, err := cmd.Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			log.Printf("获取请求计数失败: %v", err)
			continue
		}
		count, _ := strconv.Atoi(val)
		totalRequests += count
	}

	for _, cmd := range tokenCmds {
		val, err := cmd.Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			log.Printf("获取token计数失败: %v", err)
			continue
		}
		count, _ := strconv.Atoi(val)
		totalTokens += count
	}

	return totalRequests, totalTokens, nil
}

// calculateAccountCurrentRpmTpm 计算Account当前RPM/TPM（60秒滑动窗口，5秒精度）
func (s *RpmTpmService) calculateAccountCurrentRpmTpm(accountID uint) (int, int, error) {
	ctx := context.Background()
	now := time.Now()

	// 生成12个5秒时间片的key（60秒窗口）
	var requestKeys, tokenKeys []string
	for i := 0; i < 12; i++ {
		t := now.Add(-time.Duration(i*5) * time.Second)
		second := (t.Second() / 5) * 5
		timeKey := fmt.Sprintf("%s:%02d", t.Format("2006-01-02-15:04"), second)
		requestKeys = append(requestKeys, fmt.Sprintf("account:%d:requests:%s", accountID, timeKey))
		tokenKeys = append(tokenKeys, fmt.Sprintf("account:%d:tokens:%s", accountID, timeKey))
	}

	// 并发获取所有键的值
	pipe := s.redisClient.Pipeline()

	// 添加所有GET命令
	requestCmds := make([]*redis.StringCmd, len(requestKeys))
	tokenCmds := make([]*redis.StringCmd, len(tokenKeys))

	for i, key := range requestKeys {
		requestCmds[i] = pipe.Get(ctx, key)
	}
	for i, key := range tokenKeys {
		tokenCmds[i] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return 0, 0, err
	}

	// 累计结果
	totalRequests := 0
	totalTokens := 0

	for _, cmd := range requestCmds {
		val, err := cmd.Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			log.Printf("获取请求计数失败: %v", err)
			continue
		}
		count, _ := strconv.Atoi(val)
		totalRequests += count
	}

	for _, cmd := range tokenCmds {
		val, err := cmd.Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			log.Printf("获取token计数失败: %v", err)
			continue
		}
		count, _ := strconv.Atoi(val)
		totalTokens += count
	}

	return totalRequests, totalTokens, nil
}

// calculateSystemCurrentRpmTpm 计算系统当前RPM/TPM（60秒滑动窗口，5秒精度）
func (s *RpmTpmService) calculateSystemCurrentRpmTpm() (int, int, error) {
	ctx := context.Background()
	now := time.Now()

	// 生成12个5秒时间片的key（60秒窗口）
	var requestKeys, tokenKeys []string
	for i := 0; i < 12; i++ {
		t := now.Add(-time.Duration(i*5) * time.Second)
		second := (t.Second() / 5) * 5
		timeKey := fmt.Sprintf("%s:%02d", t.Format("2006-01-02-15:04"), second)
		requestKeys = append(requestKeys, fmt.Sprintf("system:requests:%s", timeKey))
		tokenKeys = append(tokenKeys, fmt.Sprintf("system:tokens:%s", timeKey))
	}

	// 并发获取所有键的值
	pipe := s.redisClient.Pipeline()

	// 添加所有GET命令
	requestCmds := make([]*redis.StringCmd, len(requestKeys))
	tokenCmds := make([]*redis.StringCmd, len(tokenKeys))

	for i, key := range requestKeys {
		requestCmds[i] = pipe.Get(ctx, key)
	}
	for i, key := range tokenKeys {
		tokenCmds[i] = pipe.Get(ctx, key)
	}

	_, err := pipe.Exec(ctx)
	if err != nil && err != redis.Nil {
		return 0, 0, err
	}

	// 累计结果
	totalRequests := 0
	totalTokens := 0

	for _, cmd := range requestCmds {
		val, err := cmd.Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			log.Printf("获取请求计数失败: %v", err)
			continue
		}
		count, _ := strconv.Atoi(val)
		totalRequests += count
	}

	for _, cmd := range tokenCmds {
		val, err := cmd.Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			log.Printf("获取token计数失败: %v", err)
			continue
		}
		count, _ := strconv.Atoi(val)
		totalTokens += count
	}

	return totalRequests, totalTokens, nil
}

// GetSystemCurrentStats 获取系统当前RPM/TPM统计
func (s *RpmTpmService) GetSystemCurrentStats() (*RpmTpmResult, error) {
	// 计算当前RPM/TPM
	rpm, tpm, err := s.calculateSystemCurrentRpmTpm()
	if err != nil {
		return nil, err
	}

	// 构建返回结果
	result := &RpmTpmResult{
		Rpm: rpm,
		Tpm: tpm,
	}

	// 获取系统配置（如果存在系统级别限制）
	systemRpmLimit := s.getSystemRpmLimit()
	systemTpmLimit := s.getSystemTpmLimit()

	// 计算RPM使用率
	if systemRpmLimit > 0 {
		result.RpmUsagePercentage = float64(rpm) / float64(systemRpmLimit) * 100
		result.IsRpmLimited = rpm >= systemRpmLimit
	}

	// 计算TPM使用率
	if systemTpmLimit > 0 {
		result.TpmUsagePercentage = float64(tpm) / float64(systemTpmLimit) * 100
		result.IsTpmLimited = tpm >= systemTpmLimit
	}

	return result, nil
}

// getSystemRpmLimit 获取系统RPM限制（从配置或环境变量）
func (s *RpmTpmService) getSystemRpmLimit() int {
	// TODO: 从配置文件或环境变量读取
	// 暂时返回0表示无限制
	return 0
}

// getSystemTpmLimit 获取系统TPM限制（从配置或环境变量）
func (s *RpmTpmService) getSystemTpmLimit() int {
	// TODO: 从配置文件或环境变量读取
	// 暂时返回0表示无限制
	return 0
}

// CheckApiKeyLimits 检查API Key是否超过RPM/TPM限制
func (s *RpmTpmService) CheckApiKeyLimits(apiKeyID uint) (bool, string, error) {
	// 获取API Key信息
	apiKey, err := model.GetApiKeyById(apiKeyID, 0)
	if err != nil {
		return false, "", err
	}

	// 如果没有设置限制，直接通过
	if apiKey.RpmLimit == 0 && apiKey.TpmLimit == 0 {
		return false, "", nil
	}

	// 检查是否在限流期间
	if apiKey.RateLimitEndTime != nil {
		endTime := time.Time(*apiKey.RateLimitEndTime)
		if endTime.After(time.Now()) {
			return true, "API Key仍在限流期间", nil
		}
	}

	// 获取当前统计
	rpm, tpm, err := s.calculateApiKeyCurrentRpmTpm(apiKeyID)
	if err != nil {
		return false, "", err
	}

	// 检查RPM限制
	if apiKey.RpmLimit > 0 && rpm >= apiKey.RpmLimit {
		s.setApiKeyRateLimit(apiKeyID, 5*time.Minute) // 限流5分钟
		return true, fmt.Sprintf("API Key RPM超限: %d/%d", rpm, apiKey.RpmLimit), nil
	}

	// 检查TPM限制
	if apiKey.TpmLimit > 0 && tpm >= apiKey.TpmLimit {
		s.setApiKeyRateLimit(apiKeyID, 5*time.Minute) // 限流5分钟
		return true, fmt.Sprintf("API Key TPM超限: %d/%d", tpm, apiKey.TpmLimit), nil
	}

	return false, "", nil
}

// CheckAccountLimits 检查Account是否超过RPM/TPM限制
func (s *RpmTpmService) CheckAccountLimits(accountID uint) (bool, string, error) {
	// 获取Account信息
	account, err := model.GetAccountByID(accountID)
	if err != nil {
		return false, "", err
	}

	// 如果没有设置限制，直接通过
	if account.RpmLimit == 0 && account.TpmLimit == 0 {
		return false, "", nil
	}

	// 检查是否在限流期间
	if account.RateLimitEndTime != nil {
		endTime := time.Time(*account.RateLimitEndTime)
		if endTime.After(time.Now()) {
			return true, "Account仍在限流期间", nil
		}
	}

	// 获取当前统计
	rpm, tpm, err := s.calculateAccountCurrentRpmTpm(accountID)
	if err != nil {
		return false, "", err
	}

	// 检查RPM限制
	if account.RpmLimit > 0 && rpm >= account.RpmLimit {
		s.setAccountRateLimit(accountID, 5*time.Minute) // 限流5分钟
		return true, fmt.Sprintf("Account RPM超限: %d/%d", rpm, account.RpmLimit), nil
	}

	// 检查TPM限制
	if account.TpmLimit > 0 && tpm >= account.TpmLimit {
		s.setAccountRateLimit(accountID, 5*time.Minute) // 限流5分钟
		return true, fmt.Sprintf("Account TPM超限: %d/%d", tpm, account.TpmLimit), nil
	}

	return false, "", nil
}

// setApiKeyRateLimit 设置API Key限流
func (s *RpmTpmService) setApiKeyRateLimit(apiKeyID uint, duration time.Duration) error {
	endTime := time.Now().Add(duration)
	modelTime := model.Time(endTime)

	return model.DB.Model(&model.ApiKey{}).
		Where("id = ?", apiKeyID).
		Update("rate_limit_end_time", &modelTime).Error
}

// setAccountRateLimit 设置Account限流
func (s *RpmTpmService) setAccountRateLimit(accountID uint, duration time.Duration) error {
	endTime := time.Now().Add(duration)
	modelTime := model.Time(endTime)

	return model.DB.Model(&model.Account{}).
		Where("id = ?", accountID).
		Update("rate_limit_end_time", &modelTime).Error
}

// UpdateDatabaseStats 更新数据库中的实时统计（定时任务调用）
func (s *RpmTpmService) UpdateDatabaseStats() error {
	// 获取所有活跃的API Keys
	var apiKeys []model.ApiKey
	err := model.DB.Where("status = 1").Find(&apiKeys).Error
	if err != nil {
		return err
	}

	// 获取所有活跃的Accounts
	var accounts []model.Account
	err = model.DB.Where("active_status = 1").Find(&accounts).Error
	if err != nil {
		return err
	}

	// 更新API Key统计
	for _, apiKey := range apiKeys {
		rpm, tpm, err := s.calculateApiKeyCurrentRpmTpm(apiKey.ID)
		if err != nil {
			log.Printf("计算API Key %d RPM/TPM失败: %v", apiKey.ID, err)
			continue
		}

		// 更新当前值和历史最大值
		updates := map[string]interface{}{
			"current_rpm": rpm,
			"current_tpm": tpm,
		}

		if rpm > apiKey.MaxRpm {
			updates["max_rpm"] = rpm
		}
		if tpm > apiKey.MaxTpm {
			updates["max_tpm"] = tpm
		}

		err = model.DB.Model(&model.ApiKey{}).
			Where("id = ?", apiKey.ID).
			Updates(updates).Error
		if err != nil {
			log.Printf("更新API Key %d统计失败: %v", apiKey.ID, err)
		}
	}

	// 更新Account统计
	for _, account := range accounts {
		rpm, tpm, err := s.calculateAccountCurrentRpmTpm(account.ID)
		if err != nil {
			log.Printf("计算Account %d RPM/TPM失败: %v", account.ID, err)
			continue
		}

		// 更新当前值和历史最大值
		updates := map[string]interface{}{
			"current_rpm": rpm,
			"current_tpm": tpm,
		}

		if rpm > account.MaxRpm {
			updates["max_rpm"] = rpm
		}
		if tpm > account.MaxTpm {
			updates["max_tpm"] = tpm
		}

		err = model.DB.Model(&model.Account{}).
			Where("id = ?", account.ID).
			Updates(updates).Error
		if err != nil {
			log.Printf("更新Account %d统计失败: %v", account.ID, err)
		}
	}

	return nil
}

// SaveHistoryStats 保存历史统计数据（定时任务调用）
func (s *RpmTpmService) SaveHistoryStats() error {
	now := time.Now()
	minuteTimestamp := time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), now.Minute(), 0, 0, now.Location())

	// 保存API Key历史统计
	err := s.saveApiKeyHistoryStats(minuteTimestamp)
	if err != nil {
		log.Printf("保存API Key历史统计失败: %v", err)
	}

	// 保存Account历史统计
	err = s.saveAccountHistoryStats(minuteTimestamp)
	if err != nil {
		log.Printf("保存Account历史统计失败: %v", err)
	}

	return nil
}

// saveApiKeyHistoryStats 保存API Key历史统计
func (s *RpmTpmService) saveApiKeyHistoryStats(minuteTimestamp time.Time) error {
	minuteKey := minuteTimestamp.Format("2006-01-02-15:04")
	ctx := context.Background()

	// 获取所有活跃的API Keys
	var apiKeys []model.ApiKey
	err := model.DB.Where("status = 1").Find(&apiKeys).Error
	if err != nil {
		return err
	}

	for _, apiKey := range apiKeys {
		requestsKey := fmt.Sprintf("api_key:%d:requests:%s", apiKey.ID, minuteKey)
		tokensKey := fmt.Sprintf("api_key:%d:tokens:%s", apiKey.ID, minuteKey)

		// 获取该分钟的统计数据
		requestsStr, _ := s.redisClient.Get(ctx, requestsKey).Result()
		tokensStr, _ := s.redisClient.Get(ctx, tokensKey).Result()

		requests, _ := strconv.Atoi(requestsStr)
		tokens, _ := strconv.Atoi(tokensStr)

		// 如果有数据才保存
		if requests > 0 || tokens > 0 {
			history := &model.ApiKeyRpmTpmHistory{
				ApiKeyID:        apiKey.ID,
				MinuteTimestamp: model.Time(minuteTimestamp),
				Rpm:             requests,
				Tpm:             tokens,
				// 这里可以从其他地方获取详细的token分类统计
			}

			err = model.DB.Create(history).Error
			if err != nil {
				log.Printf("保存API Key %d历史统计失败: %v", apiKey.ID, err)
			}
		}
	}

	return nil
}

// saveAccountHistoryStats 保存Account历史统计
func (s *RpmTpmService) saveAccountHistoryStats(minuteTimestamp time.Time) error {
	minuteKey := minuteTimestamp.Format("2006-01-02-15:04")
	ctx := context.Background()

	// 获取所有活跃的Accounts
	var accounts []model.Account
	err := model.DB.Where("active_status = 1").Find(&accounts).Error
	if err != nil {
		return err
	}

	for _, account := range accounts {
		requestsKey := fmt.Sprintf("account:%d:requests:%s", account.ID, minuteKey)
		tokensKey := fmt.Sprintf("account:%d:tokens:%s", account.ID, minuteKey)

		// 获取该分钟的统计数据
		requestsStr, _ := s.redisClient.Get(ctx, requestsKey).Result()
		tokensStr, _ := s.redisClient.Get(ctx, tokensKey).Result()

		requests, _ := strconv.Atoi(requestsStr)
		tokens, _ := strconv.Atoi(tokensStr)

		// 如果有数据才保存
		if requests > 0 || tokens > 0 {
			history := &model.AccountRpmTpmHistory{
				AccountID:       account.ID,
				MinuteTimestamp: model.Time(minuteTimestamp),
				Rpm:             requests,
				Tpm:             tokens,
				// 这里可以从其他地方获取详细的token分类统计
			}

			err = model.DB.Create(history).Error
			if err != nil {
				log.Printf("保存Account %d历史统计失败: %v", account.ID, err)
			}
		}
	}

	return nil
}

// CleanExpiredCache 清理过期的Redis缓存（定时任务调用）
func (s *RpmTpmService) CleanExpiredCache() error {
	// 这里可以实现批量清理逻辑，或者依赖Redis的TTL自动清理
	// 由于我们已经设置了TTL，Redis会自动清理过期键

	common.SysLog("清理过期RPM/TPM缓存完成")
	return nil
}

// CheckAlerts 检查告警阈值（定时任务调用）
func (s *RpmTpmService) CheckAlerts() error {
	// 检查API Key告警
	err := s.checkApiKeyAlerts()
	if err != nil {
		log.Printf("检查API Key告警失败: %v", err)
	}

	// 检查Account告警
	err = s.checkAccountAlerts()
	if err != nil {
		log.Printf("检查Account告警失败: %v", err)
	}

	return nil
}

// checkApiKeyAlerts 检查API Key告警阈值
func (s *RpmTpmService) checkApiKeyAlerts() error {
	var apiKeys []model.ApiKey
	err := model.DB.Where("status = 1 AND (rpm_warning_threshold > 0 OR tpm_warning_threshold > 0)").
		Find(&apiKeys).Error
	if err != nil {
		return err
	}

	for _, apiKey := range apiKeys {
		rpm, tpm, err := s.calculateApiKeyCurrentRpmTpm(apiKey.ID)
		if err != nil {
			continue
		}

		// 检查RPM告警
		if apiKey.RpmWarningThreshold > 0 && rpm >= apiKey.RpmWarningThreshold {
			percentage := float64(rpm) / float64(apiKey.RpmWarningThreshold) * 100
			common.SysLog(fmt.Sprintf("API Key %d RPM告警: 当前%d, 阈值%d (%.1f%%)",
				apiKey.ID, rpm, apiKey.RpmWarningThreshold, percentage))
		}

		// 检查TPM告警
		if apiKey.TpmWarningThreshold > 0 && tpm >= apiKey.TpmWarningThreshold {
			percentage := float64(tpm) / float64(apiKey.TpmWarningThreshold) * 100
			common.SysLog(fmt.Sprintf("API Key %d TPM告警: 当前%d, 阈值%d (%.1f%%)",
				apiKey.ID, tpm, apiKey.TpmWarningThreshold, percentage))
		}
	}

	return nil
}

// checkAccountAlerts 检查Account告警阈值
func (s *RpmTpmService) checkAccountAlerts() error {
	var accounts []model.Account
	err := model.DB.Where("active_status = 1 AND (rpm_warning_threshold > 0 OR tpm_warning_threshold > 0)").
		Find(&accounts).Error
	if err != nil {
		return err
	}

	for _, account := range accounts {
		rpm, tpm, err := s.calculateAccountCurrentRpmTpm(account.ID)
		if err != nil {
			continue
		}

		// 检查RPM告警
		if account.RpmWarningThreshold > 0 && rpm >= account.RpmWarningThreshold {
			percentage := float64(rpm) / float64(account.RpmWarningThreshold) * 100
			common.SysLog(fmt.Sprintf("Account %d RPM告警: 当前%d, 阈值%d (%.1f%%)",
				account.ID, rpm, account.RpmWarningThreshold, percentage))
		}

		// 检查TPM告警
		if account.TpmWarningThreshold > 0 && tpm >= account.TpmWarningThreshold {
			percentage := float64(tpm) / float64(account.TpmWarningThreshold) * 100
			common.SysLog(fmt.Sprintf("Account %d TPM告警: 当前%d, 阈值%d (%.1f%%)",
				account.ID, tpm, account.TpmWarningThreshold, percentage))
		}
	}

	return nil
}
