-- 添加总限额功能相关字段
-- 执行时间: 2024-01-01

-- 1. 添加新字段到 api_keys 表
ALTER TABLE api_keys
ADD COLUMN total_limit DECIMAL(10,4) DEFAULT 0 COMMENT '总限额(美元),0表示不限制',
ADD COLUMN total_cost DECIMAL(10,4) DEFAULT 0 COMMENT '累计总费用(USD)';

-- 2. 为现有API密钥初始化累计总费用
-- 基于历史日志计算每个API密钥的累计费用
UPDATE api_keys
SET total_cost = (
    SELECT COALESCE(SUM(total_cost), 0)
    FROM logs
    WHERE logs.api_key_id = api_keys.id
);

-- 3. 创建索引优化查询性能
CREATE INDEX idx_api_keys_total_limit ON api_keys(total_limit);
CREATE INDEX idx_api_keys_total_cost ON api_keys(total_cost);

-- 验证数据迁移结果
SELECT
    COUNT(*) as total_api_keys,
    COUNT(CASE WHEN total_limit > 0 THEN 1 END) as with_total_limit,
    COUNT(CASE WHEN total_cost > 0 THEN 1 END) as with_total_cost,
    ROUND(AVG(total_cost), 4) as avg_total_cost,
    ROUND(MAX(total_cost), 4) as max_total_cost
FROM api_keys;