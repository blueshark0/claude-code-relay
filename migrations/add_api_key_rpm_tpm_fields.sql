-- 添加 API Key RPM/TPM 功能相关字段
-- 执行时间: 2025-01-19

-- 1. 添加 RPM/TPM 相关字段到 api_keys 表
ALTER TABLE api_keys
ADD COLUMN current_rpm INT DEFAULT 0 COMMENT '当前RPM(每分钟请求数)',
ADD COLUMN current_tpm INT DEFAULT 0 COMMENT '当前TPM(每分钟Token数)',
ADD COLUMN max_rpm INT DEFAULT 0 COMMENT '历史最大RPM',
ADD COLUMN max_tpm INT DEFAULT 0 COMMENT '历史最大TPM',
ADD COLUMN rpm_limit INT DEFAULT 0 COMMENT 'RPM限制(0=无限制)',
ADD COLUMN tpm_limit INT DEFAULT 0 COMMENT 'TPM限制(0=无限制)',
ADD COLUMN rpm_warning_threshold INT DEFAULT 0 COMMENT 'RPM告警阈值',
ADD COLUMN tpm_warning_threshold INT DEFAULT 0 COMMENT 'TPM告警阈值',
ADD COLUMN rate_limit_end_time DATETIME NULL COMMENT '限流结束时间';

-- 2. 创建索引优化查询性能
CREATE INDEX idx_api_keys_current_rpm ON api_keys(current_rpm);
CREATE INDEX idx_api_keys_current_tpm ON api_keys(current_tpm);
CREATE INDEX idx_api_keys_rpm_limit ON api_keys(rpm_limit);
CREATE INDEX idx_api_keys_tpm_limit ON api_keys(tpm_limit);
CREATE INDEX idx_api_keys_rate_limit_end_time ON api_keys(rate_limit_end_time);

-- 3. 验证数据迁移结果
SELECT
    COUNT(*) as total_api_keys,
    COUNT(CASE WHEN rpm_limit > 0 THEN 1 END) as with_rpm_limit,
    COUNT(CASE WHEN tpm_limit > 0 THEN 1 END) as with_tpm_limit,
    COUNT(CASE WHEN current_rpm > 0 THEN 1 END) as with_current_rpm,
    COUNT(CASE WHEN current_tpm > 0 THEN 1 END) as with_current_tpm
FROM api_keys;