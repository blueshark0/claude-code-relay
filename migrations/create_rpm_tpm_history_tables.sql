-- 创建 RPM/TPM 历史统计表
-- 执行时间: 2025-01-19

-- 1. 创建 API Key RPM/TPM 历史统计表
CREATE TABLE api_key_rpm_tpm_stats (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    api_key_id INT NOT NULL COMMENT 'API Key ID',
    minute_timestamp DATETIME NOT NULL COMMENT '分钟级时间戳',
    rpm INT DEFAULT 0 COMMENT '该分钟的请求数',
    tpm INT DEFAULT 0 COMMENT '该分钟的Token数',
    input_tokens INT DEFAULT 0 COMMENT '输入Token数',
    output_tokens INT DEFAULT 0 COMMENT '输出Token数',
    cache_read_tokens INT DEFAULT 0 COMMENT '缓存读取Token数',
    cache_creation_tokens INT DEFAULT 0 COMMENT '缓存创建Token数',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_api_key_minute (api_key_id, minute_timestamp),
    INDEX idx_minute_timestamp (minute_timestamp),
    INDEX idx_api_key_id (api_key_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='API Key RPM/TPM分钟级统计历史';

-- 2. 创建 Account RPM/TPM 历史统计表
CREATE TABLE account_rpm_tpm_stats (
    id BIGINT AUTO_INCREMENT PRIMARY KEY,
    account_id INT NOT NULL COMMENT 'Account ID',
    minute_timestamp DATETIME NOT NULL COMMENT '分钟级时间戳',
    rpm INT DEFAULT 0 COMMENT '该分钟的请求数',
    tpm INT DEFAULT 0 COMMENT '该分钟的Token数',
    input_tokens INT DEFAULT 0 COMMENT '输入Token数',
    output_tokens INT DEFAULT 0 COMMENT '输出Token数',
    cache_read_tokens INT DEFAULT 0 COMMENT '缓存读取Token数',
    cache_creation_tokens INT DEFAULT 0 COMMENT '缓存创建Token数',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    INDEX idx_account_minute (account_id, minute_timestamp),
    INDEX idx_minute_timestamp (minute_timestamp),
    INDEX idx_account_id (account_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='Account RPM/TPM分钟级统计历史';

-- 3. 创建复合索引优化查询性能
ALTER TABLE api_key_rpm_tpm_stats ADD INDEX idx_api_key_time_range (api_key_id, minute_timestamp, rpm, tpm);
ALTER TABLE account_rpm_tpm_stats ADD INDEX idx_account_time_range (account_id, minute_timestamp, rpm, tpm);

-- 4. 验证表创建结果
SELECT
    'api_key_rpm_tpm_stats' as table_name,
    COUNT(*) as record_count,
    MAX(minute_timestamp) as latest_timestamp
FROM api_key_rpm_tpm_stats

UNION ALL

SELECT
    'account_rpm_tpm_stats' as table_name,
    COUNT(*) as record_count,
    MAX(minute_timestamp) as latest_timestamp
FROM account_rpm_tpm_stats;

-- 5. 显示表结构
DESCRIBE api_key_rpm_tpm_stats;
DESCRIBE account_rpm_tpm_stats;