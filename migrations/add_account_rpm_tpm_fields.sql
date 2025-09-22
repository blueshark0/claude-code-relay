-- 添加 Account RPM/TPM 功能相关字段
-- 执行时间: 2025-01-19

-- 1. 添加 RPM/TPM 相关字段到 accounts 表
ALTER TABLE accounts
ADD COLUMN current_rpm INT DEFAULT 0 COMMENT '当前RPM(每分钟请求数)',
ADD COLUMN current_tpm INT DEFAULT 0 COMMENT '当前TPM(每分钟Token数)',
ADD COLUMN max_rpm INT DEFAULT 0 COMMENT '历史最大RPM',
ADD COLUMN max_tpm INT DEFAULT 0 COMMENT '历史最大TPM',
ADD COLUMN rpm_limit INT DEFAULT 0 COMMENT 'RPM限制(0=无限制)',
ADD COLUMN tpm_limit INT DEFAULT 0 COMMENT 'TPM限制(0=无限制)',
ADD COLUMN rpm_warning_threshold INT DEFAULT 0 COMMENT 'RPM告警阈值',
ADD COLUMN tpm_warning_threshold INT DEFAULT 0 COMMENT 'TPM告警阈值';

-- 2. 创建索引优化查询性能
CREATE INDEX idx_accounts_current_rpm ON accounts(current_rpm);
CREATE INDEX idx_accounts_current_tpm ON accounts(current_tpm);
CREATE INDEX idx_accounts_rpm_limit ON accounts(rpm_limit);
CREATE INDEX idx_accounts_tpm_limit ON accounts(tpm_limit);

-- 3. 验证数据迁移结果
SELECT
    COUNT(*) as total_accounts,
    COUNT(CASE WHEN rpm_limit > 0 THEN 1 END) as with_rpm_limit,
    COUNT(CASE WHEN tpm_limit > 0 THEN 1 END) as with_tpm_limit,
    COUNT(CASE WHEN current_rpm > 0 THEN 1 END) as with_current_rpm,
    COUNT(CASE WHEN current_tpm > 0 THEN 1 END) as with_current_tpm
FROM accounts;