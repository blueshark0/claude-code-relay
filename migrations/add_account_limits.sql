-- 添加账号限额功能相关字段
-- 执行时间: 2025-01-18

-- 1. 添加新字段到 accounts 表
ALTER TABLE accounts
ADD COLUMN daily_limit DECIMAL(10,4) DEFAULT 0 COMMENT '每日限额(美元),0表示不限制',
ADD COLUMN total_limit DECIMAL(10,4) DEFAULT 0 COMMENT '总限额(美元),0表示不限制',
ADD COLUMN total_cost DECIMAL(10,4) DEFAULT 0 COMMENT '累计总费用(USD)';

-- 2. 创建索引优化查询性能
CREATE INDEX idx_accounts_daily_limit ON accounts(daily_limit);
CREATE INDEX idx_accounts_total_cost ON accounts(total_cost);

-- 验证数据迁移结果
SELECT
    COUNT(*) as total_accounts,
    COUNT(CASE WHEN daily_limit > 0 THEN 1 END) as with_daily_limit,
    COUNT(CASE WHEN total_limit > 0 THEN 1 END) as with_total_limit
FROM accounts;