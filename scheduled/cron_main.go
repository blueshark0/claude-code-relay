package scheduled

import (
	"claude-code-relay/common"
	"claude-code-relay/constant"
	"claude-code-relay/model"
	"claude-code-relay/relay"
	"claude-code-relay/service"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/robfig/cron/v3"
)

// 全局定时任务服务实例
var GlobalCronService *CronService

// CronService 定时任务服务
type CronService struct {
	cron *cron.Cron
}

// NewCronService 创建定时任务服务实例
func NewCronService() *CronService {
	// 使用带秒的cron解析器
	c := cron.New(cron.WithSeconds())
	return &CronService{cron: c}
}

// Start 启动定时任务
func (s *CronService) Start() {
	// 每天凌晨0点清理统计数据
	_, err := s.cron.AddFunc("0 0 0 * * *", s.resetDailyStats)
	if err != nil {
		log.Printf("Failed to add daily reset cron job: %v", err)
		return
	}

	// 每天凌晨1点清理过期日志
	_, err = s.cron.AddFunc("0 0 1 * * *", s.cleanExpiredLogs)
	if err != nil {
		log.Printf("Failed to add log cleanup cron job: %v", err)
		return
	}

	// 每30分钟执行一次账号异常恢复测试
	_, err = s.cron.AddFunc("0 */30 * * * *", s.recoverAbnormalAccounts)
	if err != nil {
		log.Printf("Failed to add account recovery cron job: %v", err)
		return
	}

	// 每10分钟检查限流过期账号
	_, err = s.cron.AddFunc("0 */10 * * * *", s.checkRateLimitExpiredAccounts)
	if err != nil {
		log.Printf("Failed to add rate limit check cron job: %v", err)
		return
	}

	// RPM/TPM 统计相关定时任务
	s.addRpmTpmCronJobs()

	// 启动定时任务
	s.cron.Start()
	common.SysLog("Cron service started successfully")
}

// Stop 停止定时任务
func (s *CronService) Stop() {
	if s.cron != nil {
		ctx := s.cron.Stop()
		<-ctx.Done()
		common.SysLog("Cron service stopped")
	}
}

// resetDailyStats 重置每日统计数据
func (s *CronService) resetDailyStats() {
	startTime := time.Now()
	common.SysLog("Starting daily stats reset task")

	// 重置Account表的今日统计数据
	err := s.resetAccountStats()
	if err != nil {
		common.SysError("Failed to reset account daily stats: " + err.Error())
	} else {
		common.SysLog("Account daily stats reset successfully")
	}

	// 重置ApiKey表的今日统计数据
	err = s.resetApiKeyStats()
	if err != nil {
		common.SysError("Failed to reset api key daily stats: " + err.Error())
	} else {
		common.SysLog("API Key daily stats reset successfully")
	}

	duration := time.Since(startTime)
	common.SysLog("Daily stats reset task completed in " + duration.String())
}

// resetAccountStats 重置账户今日统计数据
func (s *CronService) resetAccountStats() error {
	result := model.DB.Model(&model.Account{}).Where("1 = 1").Updates(map[string]interface{}{
		"today_usage_count":                 0,
		"today_input_tokens":                0,
		"today_output_tokens":               0,
		"today_cache_read_input_tokens":     0,
		"today_cache_creation_input_tokens": 0,
		"today_total_cost":                  0,
	})

	if result.Error != nil {
		return result.Error
	}

	log.Printf("Reset daily stats for %d accounts", result.RowsAffected)
	return nil
}

// resetApiKeyStats 重置API Key今日统计数据
func (s *CronService) resetApiKeyStats() error {
	result := model.DB.Model(&model.ApiKey{}).Where("1 = 1").Updates(map[string]interface{}{
		"today_usage_count":                 0,
		"today_input_tokens":                0,
		"today_output_tokens":               0,
		"today_cache_read_input_tokens":     0,
		"today_cache_creation_input_tokens": 0,
		"today_total_cost":                  0,
	})

	if result.Error != nil {
		return result.Error
	}

	log.Printf("Reset daily stats for %d api keys", result.RowsAffected)
	return nil
}

// ManualResetStats 手动重置统计数据（用于测试或管理员操作）
func (s *CronService) ManualResetStats() error {
	common.SysLog("Manual daily stats reset triggered")

	err := s.resetAccountStats()
	if err != nil {
		return err
	}

	err = s.resetApiKeyStats()
	if err != nil {
		return err
	}

	common.SysLog("Manual daily stats reset completed")
	return nil
}

// InitCronService 初始化全局定时任务服务
func InitCronService() {
	GlobalCronService = NewCronService()
	GlobalCronService.Start()
}

// StopCronService 停止全局定时任务服务
func StopCronService() {
	if GlobalCronService != nil {
		GlobalCronService.Stop()
	}
}

// cleanExpiredLogs 清理过期日志
func (s *CronService) cleanExpiredLogs() {
	startTime := time.Now()
	common.SysLog("Starting expired logs cleanup task")

	// 从环境变量获取日志保留月数，默认为3个月
	retentionMonths := getLogRetentionMonths()

	logService := service.NewLogService()
	deletedCount, err := logService.DeleteExpiredLogs(retentionMonths)
	if err != nil {
		common.SysError("Failed to clean expired logs: " + err.Error())
	} else {
		common.SysLog("Cleaned expired logs successfully, deleted " + strconv.FormatInt(deletedCount, 10) + " records (older than " + strconv.Itoa(retentionMonths) + " months)")
	}

	duration := time.Since(startTime)
	common.SysLog("Expired logs cleanup task completed in " + duration.String())
}

// getLogRetentionMonths 从环境变量获取日志保留月数
func getLogRetentionMonths() int {
	monthsStr := os.Getenv("LOG_RETENTION_MONTHS")
	if monthsStr == "" {
		return 3 // 默认保留3个月
	}

	months, err := strconv.Atoi(monthsStr)
	if err != nil || months <= 0 {
		log.Printf("Invalid LOG_RETENTION_MONTHS value: %s, using default value 3", monthsStr)
		return 3
	}

	return months
}

// recoverAbnormalAccounts 恢复异常账号测试
func (s *CronService) recoverAbnormalAccounts() {
	startTime := time.Now()
	common.SysLog("Starting abnormal accounts recovery task")

	// 筛选current_status==2且active_status==1的账号
	var abnormalAccounts []model.Account
	err := model.DB.Where("current_status = ? AND active_status = ?", 2, 1).Find(&abnormalAccounts).Error
	if err != nil {
		common.SysError("Failed to query abnormal accounts: " + err.Error())
		return
	}

	if len(abnormalAccounts) == 0 {
		return
	}

	common.SysLog(fmt.Sprintf("Found %d abnormal accounts to test", len(abnormalAccounts)))

	recoveredCount := 0
	failedCount := 0

	// 逐个测试异常账号
	for _, account := range abnormalAccounts {
		if s.testAndRecoverAccount(&account) {
			recoveredCount++
			common.SysLog(fmt.Sprintf("Account %s (ID: %d) recovered successfully", account.Name, account.ID))
		} else {
			failedCount++
		}
	}

	duration := time.Since(startTime)
	common.SysLog(fmt.Sprintf("Abnormal accounts recovery task completed in %s. Recovered: %d, Failed: %d", duration.String(), recoveredCount, failedCount))
}

// testAndRecoverAccount 测试并恢复单个账号
func (s *CronService) testAndRecoverAccount(account *model.Account) bool {
	var statusCode int
	var err string

	// 根据平台类型调用不同的测试函数
	switch account.PlatformType {
	case constant.PlatformClaude:
		statusCode, err = relay.TestsHandleClaudeRequest(account)
	case constant.PlatformClaudeConsole:
		statusCode, err = relay.TestHandleClaudeConsoleRequest(account)
	case constant.PlatformOpenAI:
		statusCode, err = relay.TestHandleOpenAIRequest(account)
	default:
		common.SysError(fmt.Sprintf("Unsupported platform type for account %s (ID: %d): %s", account.Name, account.ID, account.PlatformType))
		return false
	}

	// 检查测试结果：状态码在200-300之间且无错误视为成功
	if err == "" && statusCode >= 200 && statusCode < 300 {
		// 测试成功，恢复账号状态为正常
		updateErr := model.DB.Model(account).Update("current_status", 1).Error
		if updateErr != nil {
			common.SysError(fmt.Sprintf("Failed to recover account %s (ID: %d): %v", account.Name, account.ID, updateErr))
			return false
		}
		return true
	}

	return false
}

// ManualCleanExpiredLogs 手动清理过期日志（用于测试或管理员操作）
func (s *CronService) ManualCleanExpiredLogs() (int64, error) {
	common.SysLog("Manual expired logs cleanup triggered")

	retentionMonths := getLogRetentionMonths()
	logService := service.NewLogService()
	deletedCount, err := logService.DeleteExpiredLogs(retentionMonths)
	if err != nil {
		return 0, err
	}

	common.SysLog("Manual expired logs cleanup completed, deleted " + strconv.FormatInt(deletedCount, 10) + " records")
	return deletedCount, nil
}

// checkRateLimitExpiredAccounts 检查限流过期账号
func (s *CronService) checkRateLimitExpiredAccounts() {
	startTime := time.Now()
	common.SysLog("Starting rate limit expired accounts check task")

	// 筛选current_status==3且active_status==1的账号
	var rateLimitedAccounts []model.Account
	err := model.DB.Where("current_status = ? AND active_status = ?", 3, 1).Find(&rateLimitedAccounts).Error
	if err != nil {
		common.SysError("Failed to query rate limited accounts: " + err.Error())
		return
	}

	if len(rateLimitedAccounts) == 0 {
		return
	}

	common.SysLog(fmt.Sprintf("Found %d rate limited accounts to check", len(rateLimitedAccounts)))

	recoveredCount := 0
	now := time.Now()

	// 检查每个限流账号的限流结束时间
	for _, account := range rateLimitedAccounts {
		// 检查限流结束时间是否已过期
		if account.RateLimitEndTime != nil && now.After(time.Time(*account.RateLimitEndTime)) {
			// 限流时间已过，将账号状态恢复为正常并清空限流结束时间
			err := model.DB.Model(&account).Updates(map[string]interface{}{
				"current_status":      1,
				"rate_limit_end_time": nil,
			}).Error
			if err != nil {
				common.SysError(fmt.Sprintf("Failed to recover rate limited account %s (ID: %d): %v", account.Name, account.ID, err))
				continue
			}
			recoveredCount++
			common.SysLog(fmt.Sprintf("Rate limited account %s (ID: %d) recovered successfully, limit expired at %v", account.Name, account.ID, account.RateLimitEndTime))
		} else if account.RateLimitEndTime != nil {
			common.SysLog(fmt.Sprintf("Account %s (ID: %d) still rate limited until %v", account.Name, account.ID, account.RateLimitEndTime))
		} else {
			common.SysLog(fmt.Sprintf("Account %s (ID: %d) has no rate limit end time, skipping", account.Name, account.ID))
		}
	}

	duration := time.Since(startTime)
	common.SysLog(fmt.Sprintf("Rate limit expired accounts check task completed in %s. Recovered: %d", duration.String(), recoveredCount))
}

// addRpmTpmCronJobs 添加RPM/TPM相关定时任务
func (s *CronService) addRpmTpmCronJobs() {
	// 每分钟更新数据库中的RPM/TPM统计
	_, err := s.cron.AddFunc("0 * * * * *", s.updateRpmTpmDatabaseStats)
	if err != nil {
		log.Printf("Failed to add RPM/TPM database stats update job: %v", err)
		return
	}

	// 每分钟保存历史统计数据
	_, err = s.cron.AddFunc("30 * * * * *", s.saveRpmTpmHistoryStats)
	if err != nil {
		log.Printf("Failed to add RPM/TPM history stats save job: %v", err)
		return
	}

	// 每5分钟检查告警阈值
	_, err = s.cron.AddFunc("0 */5 * * * *", s.checkRpmTpmAlerts)
	if err != nil {
		log.Printf("Failed to add RPM/TPM alerts check job: %v", err)
		return
	}

	// 每天凌晨3点清理过期的Redis缓存
	_, err = s.cron.AddFunc("0 0 3 * * *", s.cleanRpmTpmExpiredCache)
	if err != nil {
		log.Printf("Failed to add RPM/TPM cache cleanup job: %v", err)
		return
	}

	common.SysLog("RPM/TPM cron jobs added successfully")
}

// updateRpmTpmDatabaseStats 更新数据库中的RPM/TPM统计
func (s *CronService) updateRpmTpmDatabaseStats() {
	rpmTpmService := service.NewRpmTpmService()
	err := rpmTpmService.UpdateDatabaseStats()
	if err != nil {
		log.Printf("Failed to update RPM/TPM database stats: %v", err)
	}
}

// saveRpmTpmHistoryStats 保存RPM/TPM历史统计数据
func (s *CronService) saveRpmTpmHistoryStats() {
	rpmTpmService := service.NewRpmTpmService()
	err := rpmTpmService.SaveHistoryStats()
	if err != nil {
		log.Printf("Failed to save RPM/TPM history stats: %v", err)
	}
}

// checkRpmTpmAlerts 检查RPM/TPM告警阈值
func (s *CronService) checkRpmTpmAlerts() {
	rpmTpmService := service.NewRpmTpmService()
	err := rpmTpmService.CheckAlerts()
	if err != nil {
		log.Printf("Failed to check RPM/TPM alerts: %v", err)
	}
}

// cleanRpmTpmExpiredCache 清理过期的RPM/TPM Redis缓存
func (s *CronService) cleanRpmTpmExpiredCache() {
	rpmTpmService := service.NewRpmTpmService()
	err := rpmTpmService.CleanExpiredCache()
	if err != nil {
		log.Printf("Failed to clean RPM/TPM expired cache: %v", err)
	}
}
