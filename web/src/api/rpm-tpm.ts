import { request } from '@/utils/request';

// RPM/TPM API 接口
const Api = {
  // API Key RPM/TPM 相关
  GetApiKeyRpmTpmStats: '/api/v1/api-keys/{id}/rpm-tpm/stats',
  UpdateApiKeyRpmTpmLimits: '/api/v1/api-keys/{id}/rpm-tpm/limits',
  GetApiKeyRpmTpmHistory: '/api/v1/api-keys/{id}/rpm-tpm/history',

  // Account RPM/TPM 相关
  GetAccountRpmTpmStats: '/api/v1/accounts/{id}/rpm-tpm/stats',
  UpdateAccountRpmTpmLimits: '/api/v1/accounts/{id}/rpm-tpm/limits',
  GetAccountRpmTpmHistory: '/api/v1/accounts/{id}/rpm-tpm/history',

  // RPM/TPM 仪表盘
  GetRpmTpmDashboard: '/api/v1/rpm-tpm/dashboard',
};

// ===== 类型定义 =====

// RPM/TPM 统计结果
export interface RpmTpmStats {
  api_key_id?: number;
  account_id?: number;
  current_rpm: number;
  current_tpm: number;
  max_rpm: number;
  max_tpm: number;
  rpm_limit: number;
  tpm_limit: number;
  rpm_warning_threshold: number;
  tpm_warning_threshold: number;
  rpm_usage_percentage: number;
  tpm_usage_percentage: number;
  is_rpm_limited: boolean;
  is_tpm_limited: boolean;
  rate_limit_end_time?: string;
}

// API Key RPM/TPM 统计响应
export interface ApiKeyRpmTpmStatsResponse {
  data: RpmTpmStats;
  code: number;
}

// Account RPM/TPM 统计响应
export interface AccountRpmTpmStatsResponse {
  data: RpmTpmStats;
  code: number;
}

// RPM/TPM 限制更新参数
export interface UpdateRpmTpmLimitsParams {
  rpm_limit?: number;
  tpm_limit?: number;
  rpm_warning_threshold?: number;
  tpm_warning_threshold?: number;
}

// RPM/TPM 历史数据项
export interface RpmTpmHistoryItem {
  id: number;
  api_key_id?: number;
  account_id?: number;
  minute_timestamp: string;
  rpm: number;
  tpm: number;
  input_tokens?: number;
  output_tokens?: number;
  cache_read_tokens?: number;
  cache_creation_tokens?: number;
  created_at: string;
}

// RPM/TPM 历史查询参数
export interface RpmTpmHistoryParams {
  start_time?: string;
  end_time?: string;
  page?: number;
  limit?: number;
}

// RPM/TPM 历史查询响应
export interface RpmTpmHistoryResponse {
  history: RpmTpmHistoryItem[];
  total: number;
  page: number;
  limit: number;
}

// RPM/TPM 仪表盘数据
export interface RpmTpmDashboardData {
  summary: {
    total_api_keys: number;
    total_accounts: number;
    active_limits: number;
    current_alerts: number;
  };
  top_api_keys: Array<{
    id: number;
    name: string;
    current_rpm: number;
    current_tpm: number;
    rpm_limit: number;
    tpm_limit: number;
  }>;
  top_accounts: Array<{
    id: number;
    name: string;
    current_rpm: number;
    current_tpm: number;
    rpm_limit: number;
    tpm_limit: number;
  }>;
  recent_alerts: Array<{
    type: 'api_key' | 'account';
    id: number;
    name: string;
    alert_type: 'rpm' | 'tpm';
    current_value: number;
    limit_value: number;
    timestamp: string;
  }>;
}

// RPM/TPM 仪表盘响应
export interface RpmTpmDashboardResponse {
  data: RpmTpmDashboardData;
  code: number;
}

// ===== API 接口函数 =====

/**
 * 获取API Key的RPM/TPM统计
 */
export function getApiKeyRpmTpmStats(id: number) {
  return request.get<ApiKeyRpmTpmStatsResponse>({
    url: Api.GetApiKeyRpmTpmStats.replace('{id}', String(id)),
  });
}

/**
 * 更新API Key的RPM/TPM限制
 */
export function updateApiKeyRpmTpmLimits(id: number, data: UpdateRpmTpmLimitsParams) {
  return request.put({
    url: Api.UpdateApiKeyRpmTpmLimits.replace('{id}', String(id)),
    data,
  });
}

/**
 * 获取API Key的RPM/TPM历史数据
 */
export function getApiKeyRpmTpmHistory(id: number, params?: RpmTpmHistoryParams) {
  return request.get<RpmTpmHistoryResponse>({
    url: Api.GetApiKeyRpmTpmHistory.replace('{id}', String(id)),
    params,
  });
}

/**
 * 获取Account的RPM/TPM统计
 */
export function getAccountRpmTpmStats(id: number) {
  return request.get<AccountRpmTpmStatsResponse>({
    url: Api.GetAccountRpmTpmStats.replace('{id}', String(id)),
  });
}

/**
 * 更新Account的RPM/TPM限制
 */
export function updateAccountRpmTpmLimits(id: number, data: UpdateRpmTpmLimitsParams) {
  return request.put({
    url: Api.UpdateAccountRpmTpmLimits.replace('{id}', String(id)),
    data,
  });
}

/**
 * 获取Account的RPM/TPM历史数据
 */
export function getAccountRpmTpmHistory(id: number, params?: RpmTpmHistoryParams) {
  return request.get<RpmTpmHistoryResponse>({
    url: Api.GetAccountRpmTpmHistory.replace('{id}', String(id)),
    params,
  });
}

/**
 * 获取RPM/TPM仪表盘数据
 */
export function getRpmTpmDashboard() {
  return request.get<RpmTpmDashboardResponse>({
    url: Api.GetRpmTpmDashboard,
  });
}

// ===== 工具函数 =====

/**
 * 格式化 RPM/TPM 数值显示
 */
export function formatRpmTpmValue(value: number): string {
  if (value >= 1000000) {
    return `${(value / 1000000).toFixed(1)}M`;
  }
  if (value >= 1000) {
    return `${(value / 1000).toFixed(1)}K`;
  }
  return String(value);
}

/**
 * 获取使用率状态主题
 */
export function getUsageTheme(usage: number): 'success' | 'warning' | 'danger' {
  if (usage >= 95) return 'danger';
  if (usage >= 80) return 'warning';
  return 'success';
}

/**
 * 获取限流状态文本
 */
export function getLimitStatusText(isLimited: boolean, endTime?: string): string {
  if (!isLimited) return '正常';
  if (endTime) {
    const end = new Date(endTime);
    const now = new Date();
    if (end > now) {
      return `限流至 ${end.toLocaleString('zh-CN')}`;
    }
  }
  return '限流中';
}

/**
 * 计算使用率百分比
 */
export function calculateUsagePercentage(current: number, limit: number): number {
  if (limit <= 0) return 0;
  return Math.min((current / limit) * 100, 100);
}
