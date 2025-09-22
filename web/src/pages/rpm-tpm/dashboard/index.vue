<template>
  <div class="rpm-tpm-dashboard">
    <div class="dashboard-header">
      <div class="header-content">
        <t-icon name="dashboard" class="header-icon" />
        <div class="header-text">
          <h1 class="dashboard-title">RPM/TPM 监控仪表盘</h1>
          <p class="dashboard-subtitle">实时监控系统请求和Token使用情况</p>
        </div>
      </div>
      <div class="header-actions">
        <t-button variant="outline" size="small" :loading="refreshing" @click="handleRefreshAll">
          <template #icon>
            <t-icon name="refresh" />
          </template>
          刷新数据
        </t-button>
        <t-button theme="primary" size="small" @click="showSettingsDialog">
          <template #icon>
            <t-icon name="setting" />
          </template>
          全局设置
        </t-button>
      </div>
    </div>

    <!-- 系统概况卡片 -->
    <div class="overview-section">
      <div class="section-title">系统概况</div>
      <div class="overview-grid">
        <stats-card
          title="系统总体状态"
          subtitle="所有账户聚合数据"
          icon="layers"
          :stats="systemOverview"
          :loading="systemLoading"
          :show-manage-button="false"
          :show-history-button="false"
          @refresh="handleRefreshSystem"
        />
        <div class="overview-alerts">
          <div class="alerts-title">
            <t-icon name="notification" />
            <span>实时告警</span>
          </div>
          <div v-if="alerts.length === 0" class="no-alerts">
            <t-icon name="check-circle" />
            <span>系统运行正常</span>
          </div>
          <div v-else class="alerts-list">
            <t-alert v-for="alert in alerts" :key="alert.id" :theme="alert.theme" size="small" class="alert-item">
              <template #icon>
                <t-icon :name="alert.icon" />
              </template>
              <div class="alert-content">
                <div class="alert-title">{{ alert.title }}</div>
                <div class="alert-message">{{ alert.message }}</div>
              </div>
            </t-alert>
          </div>
        </div>
      </div>
    </div>

    <!-- 排行榜部分 -->
    <div class="rankings-section">
      <div class="section-title">使用率排行榜</div>
      <div class="rankings-grid">
        <!-- API Keys 排行榜 -->
        <div class="ranking-card">
          <div class="ranking-header">
            <t-icon name="key" />
            <span>API Keys Top 10</span>
            <t-button variant="text" size="small" @click="$router.push('/rpm-tpm/api-keys')"> 查看全部 </t-button>
          </div>
          <div v-if="apiKeysLoading" class="ranking-loading">
            <t-loading size="small" text="加载中..." />
          </div>
          <div v-else-if="topApiKeys.length === 0" class="ranking-empty">
            <t-icon name="inbox" />
            <span>暂无数据</span>
          </div>
          <div v-else class="ranking-list">
            <div
              v-for="(item, index) in topApiKeys"
              :key="item.api_key_id"
              class="ranking-item"
              :class="{ 'ranking-warning': item.rpm_usage_percentage >= 80 || item.tpm_usage_percentage >= 80 }"
            >
              <div class="ranking-index">
                <t-icon v-if="index < 3" name="medal" :class="`medal-${index + 1}`" />
                <span v-else class="rank-number">{{ index + 1 }}</span>
              </div>
              <div class="ranking-info">
                <div class="ranking-name">API Key {{ item.api_key_id }}</div>
                <div class="ranking-usage">
                  <usage-progress
                    title="RPM"
                    :current="item.current_rpm"
                    :limit="item.rpm_limit"
                    :is-limited="item.is_rpm_limited"
                    compact
                  />
                  <usage-progress
                    title="TPM"
                    :current="item.current_tpm"
                    :limit="item.tpm_limit"
                    :is-limited="item.is_tpm_limited"
                    compact
                  />
                </div>
              </div>
              <alert-badge
                :rpm-current="item.current_rpm"
                :rpm-limit="item.rpm_limit"
                :is-rpm-limited="item.is_rpm_limited"
                :tpm-current="item.current_tpm"
                :tpm-limit="item.tpm_limit"
                :is-tpm-limited="item.is_tpm_limited"
                :is-rate-limited="!!item.rate_limit_end_time"
                :rate-limit-end-time="item.rate_limit_end_time"
                simple
                dot
              />
            </div>
          </div>
        </div>

        <!-- Accounts 排行榜 -->
        <div class="ranking-card">
          <div class="ranking-header">
            <t-icon name="user-circle" />
            <span>Accounts Top 10</span>
            <t-button variant="text" size="small" @click="$router.push('/rpm-tpm/accounts')"> 查看全部 </t-button>
          </div>
          <div v-if="accountsLoading" class="ranking-loading">
            <t-loading size="small" text="加载中..." />
          </div>
          <div v-else-if="topAccounts.length === 0" class="ranking-empty">
            <t-icon name="inbox" />
            <span>暂无数据</span>
          </div>
          <div v-else class="ranking-list">
            <div
              v-for="(item, index) in topAccounts"
              :key="item.account_id"
              class="ranking-item"
              :class="{ 'ranking-warning': item.rpm_usage_percentage >= 80 || item.tpm_usage_percentage >= 80 }"
            >
              <div class="ranking-index">
                <t-icon v-if="index < 3" name="medal" :class="`medal-${index + 1}`" />
                <span v-else class="rank-number">{{ index + 1 }}</span>
              </div>
              <div class="ranking-info">
                <div class="ranking-name">Account {{ item.account_id }}</div>
                <div class="ranking-usage">
                  <usage-progress
                    title="RPM"
                    :current="item.current_rpm"
                    :limit="item.rpm_limit"
                    :is-limited="item.is_rpm_limited"
                    compact
                  />
                  <usage-progress
                    title="TPM"
                    :current="item.current_tpm"
                    :limit="item.tpm_limit"
                    :is-limited="item.is_tpm_limited"
                    compact
                  />
                </div>
              </div>
              <alert-badge
                :rpm-current="item.current_rpm"
                :rpm-limit="item.rpm_limit"
                :is-rpm-limited="item.is_rpm_limited"
                :tpm-current="item.current_tpm"
                :tpm-limit="item.tpm_limit"
                :is-tpm-limited="item.is_tpm_limited"
                :is-rate-limited="!!item.rate_limit_end_time"
                :rate-limit-end-time="item.rate_limit_end_time"
                simple
                dot
              />
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 历史趋势图表 -->
    <div class="charts-section">
      <div class="section-title">历史趋势分析</div>
      <div class="charts-grid">
        <div class="chart-container">
          <history-chart
            title="系统 RPM/TPM 历史趋势"
            :data="systemHistoryData"
            :loading="systemHistoryLoading"
            :rpm-limit="systemOverview.rpm_limit"
            :tpm-limit="systemOverview.tpm_limit"
            :rpm-warning-threshold="systemOverview.rpm_warning_threshold"
            :tpm-warning-threshold="systemOverview.tpm_warning_threshold"
            @time-range-change="handleSystemHistoryTimeRange"
            @refresh="handleRefreshSystemHistory"
          />
        </div>
      </div>
    </div>

    <!-- 全局设置对话框 -->
    <t-dialog
      v-model:visible="settingsVisible"
      header="全局 RPM/TPM 设置"
      width="600px"
      @confirm="handleSaveGlobalSettings"
      @cancel="handleCancelGlobalSettings"
    >
      <div class="global-settings">
        <t-form ref="globalSettingsFormRef" :model="globalSettingsForm" :rules="globalSettingsRules" label-align="top">
          <div class="settings-notice">
            <t-alert theme="info" size="small">
              全局设置将作为新创建账户和API Key的默认限制值，不会影响现有配置。
            </t-alert>
          </div>
          <limits-form
            :initial-data="globalSettingsForm"
            :loading="globalSettingsSaving"
            submit-text="保存全局设置"
            @submit="handleSaveGlobalSettings"
            @cancel="handleCancelGlobalSettings"
          />
        </t-form>
      </div>
    </t-dialog>
  </div>
</template>
<script setup lang="ts">
import type { FormInstanceFunctions } from 'tdesign-vue-next';
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue';

import type { GetRpmTpmHistoryParams, RpmTpmHistoryItem, RpmTpmStats, UpdateRpmTpmLimitsParams } from '@/api/rpm-tpm';
import {
  getAccountRpmTpmRanking,
  getApiKeyRpmTpmRanking,
  getGlobalRpmTpmLimits,
  getSystemRpmTpmHistory,
  getSystemRpmTpmStats,
  updateGlobalRpmTpmLimits,
} from '@/api/rpm-tpm';
import AlertBadge from '@/components/rpm-tpm/AlertBadge.vue';
import UsageProgress from '@/components/rpm-tpm/UsageProgress.vue';

import HistoryChart from '../components/HistoryChart.vue';
import LimitsForm from '../components/LimitsForm.vue';
import StatsCard from '../components/StatsCard.vue';

interface AlertItem {
  id: string;
  title: string;
  message: string;
  theme: 'error' | 'warning' | 'info';
  icon: string;
}

// 系统概况数据
const systemOverview = ref<RpmTpmStats>({
  current_rpm: 0,
  current_tpm: 0,
  max_rpm: 0,
  max_tpm: 0,
  rpm_limit: 0,
  tpm_limit: 0,
  rpm_usage_percentage: 0,
  tpm_usage_percentage: 0,
  is_rpm_limited: false,
  is_tpm_limited: false,
  rpm_warning_threshold: 0,
  tpm_warning_threshold: 0,
});

// 加载状态
const systemLoading = ref(false);
const apiKeysLoading = ref(false);
const accountsLoading = ref(false);
const systemHistoryLoading = ref(false);
const refreshing = ref(false);

// 排行榜数据
const topApiKeys = ref<RpmTpmStats[]>([]);
const topAccounts = ref<RpmTpmStats[]>([]);

// 历史数据
const systemHistoryData = ref<RpmTpmHistoryItem[]>([]);

// 全局设置相关
const settingsVisible = ref(false);
const globalSettingsFormRef = ref<FormInstanceFunctions>();
const globalSettingsSaving = ref(false);
const globalSettingsForm = reactive<UpdateRpmTpmLimitsParams>({
  rpm_limit: 0,
  tpm_limit: 0,
  rpm_warning_threshold: 0,
  tpm_warning_threshold: 0,
});

const globalSettingsRules = {};

// 自动刷新定时器
let refreshTimer: NodeJS.Timeout | null = null;

// 告警数据
const alerts = computed<AlertItem[]>(() => {
  const alertList: AlertItem[] = [];

  // 检查系统级别告警
  if (systemOverview.value.is_rpm_limited) {
    alertList.push({
      id: 'system-rpm-limited',
      title: '系统RPM限流',
      message: '系统当前RPM已达到限制值，部分请求被限流',
      theme: 'error',
      icon: 'close-circle',
    });
  }

  if (systemOverview.value.is_tpm_limited) {
    alertList.push({
      id: 'system-tpm-limited',
      title: '系统TPM限流',
      message: '系统当前TPM已达到限制值，部分请求被限流',
      theme: 'error',
      icon: 'close-circle',
    });
  }

  // 检查高使用率告警
  if (systemOverview.value.rpm_usage_percentage >= 90) {
    alertList.push({
      id: 'system-rpm-high',
      title: 'RPM使用率过高',
      message: `当前RPM使用率 ${systemOverview.value.rpm_usage_percentage.toFixed(1)}%，接近限制值`,
      theme: 'warning',
      icon: 'error-circle',
    });
  }

  if (systemOverview.value.tpm_usage_percentage >= 90) {
    alertList.push({
      id: 'system-tpm-high',
      title: 'TPM使用率过高',
      message: `当前TPM使用率 ${systemOverview.value.tpm_usage_percentage.toFixed(1)}%，接近限制值`,
      theme: 'warning',
      icon: 'error-circle',
    });
  }

  // 检查排行榜中的高风险项目
  const highRiskApiKeys = topApiKeys.value.filter(
    (item) => item.is_rpm_limited || item.is_tpm_limited || item.rate_limit_end_time,
  );
  if (highRiskApiKeys.length > 0) {
    alertList.push({
      id: 'api-keys-risk',
      title: 'API Keys 异常',
      message: `${highRiskApiKeys.length} 个 API Key 存在限流或告警状态`,
      theme: 'warning',
      icon: 'key',
    });
  }

  const highRiskAccounts = topAccounts.value.filter(
    (item) => item.is_rpm_limited || item.is_tpm_limited || item.rate_limit_end_time,
  );
  if (highRiskAccounts.length > 0) {
    alertList.push({
      id: 'accounts-risk',
      title: 'Accounts 异常',
      message: `${highRiskAccounts.length} 个 Account 存在限流或告警状态`,
      theme: 'warning',
      icon: 'user-circle',
    });
  }

  return alertList.slice(0, 5); // 最多显示5个告警
});

// 加载系统概况
const loadSystemOverview = async () => {
  try {
    systemLoading.value = true;
    const response = await getSystemRpmTpmStats();
    systemOverview.value = response.data;
  } catch (error) {
    console.error('Failed to load system overview:', error);
  } finally {
    systemLoading.value = false;
  }
};

// 加载API Keys排行榜
const loadApiKeysRanking = async () => {
  try {
    apiKeysLoading.value = true;
    const response = await getApiKeyRpmTpmRanking({ limit: 10 });
    topApiKeys.value = response.data;
  } catch (error) {
    console.error('Failed to load API keys ranking:', error);
  } finally {
    apiKeysLoading.value = false;
  }
};

// 加载Accounts排行榜
const loadAccountsRanking = async () => {
  try {
    accountsLoading.value = true;
    const response = await getAccountRpmTpmRanking({ limit: 10 });
    topAccounts.value = response.data;
  } catch (error) {
    console.error('Failed to load accounts ranking:', error);
  } finally {
    accountsLoading.value = false;
  }
};

// 加载系统历史数据
const loadSystemHistory = async (params: GetRpmTpmHistoryParams = {}) => {
  try {
    systemHistoryLoading.value = true;
    const response = await getSystemRpmTpmHistory({
      start_time: params.start_time || new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
      end_time: params.end_time || new Date().toISOString(),
      ...params,
    });
    systemHistoryData.value = response.data;
  } catch (error) {
    console.error('Failed to load system history:', error);
  } finally {
    systemHistoryLoading.value = false;
  }
};

// 加载全局设置
const loadGlobalSettings = async () => {
  try {
    const response = await getGlobalRpmTpmLimits();
    Object.assign(globalSettingsForm, response.data);
  } catch (error) {
    console.error('Failed to load global settings:', error);
  }
};

// 刷新所有数据
const handleRefreshAll = async () => {
  refreshing.value = true;
  try {
    await Promise.all([loadSystemOverview(), loadApiKeysRanking(), loadAccountsRanking(), loadSystemHistory()]);
  } finally {
    refreshing.value = false;
  }
};

// 刷新系统概况
const handleRefreshSystem = () => {
  loadSystemOverview();
};

// 处理历史数据时间范围变化
const handleSystemHistoryTimeRange = (params: { start_time: string; end_time: string }) => {
  loadSystemHistory(params);
};

// 刷新系统历史数据
const handleRefreshSystemHistory = () => {
  loadSystemHistory();
};

// 显示全局设置对话框
const showSettingsDialog = () => {
  settingsVisible.value = true;
  loadGlobalSettings();
};

// 保存全局设置
const handleSaveGlobalSettings = async (data: UpdateRpmTpmLimitsParams) => {
  try {
    globalSettingsSaving.value = true;
    await updateGlobalRpmTpmLimits(data);
    settingsVisible.value = false;
  } catch (error) {
    console.error('Failed to save global settings:', error);
  } finally {
    globalSettingsSaving.value = false;
  }
};

// 取消全局设置
const handleCancelGlobalSettings = () => {
  settingsVisible.value = false;
};

// 启动自动刷新
const startAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer);
  }
  refreshTimer = setInterval(() => {
    handleRefreshAll();
  }, 30000); // 30秒刷新一次
};

// 停止自动刷新
const stopAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer);
    refreshTimer = null;
  }
};

// 生命周期
onMounted(() => {
  handleRefreshAll();
  startAutoRefresh();
});

onUnmounted(() => {
  stopAutoRefresh();
});
</script>
<style lang="less" scoped>
.rpm-tpm-dashboard {
  padding: 24px;
  max-width: 1400px;
  margin: 0 auto;

  .dashboard-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 32px;
    padding-bottom: 20px;
    border-bottom: 2px solid var(--td-border-level-1-color);

    .header-content {
      display: flex;
      align-items: center;
      gap: 16px;

      .header-icon {
        font-size: 32px;
        color: var(--td-brand-color);
      }

      .header-text {
        .dashboard-title {
          font-size: 28px;
          font-weight: 700;
          color: var(--td-text-color-primary);
          margin: 0 0 4px 0;
        }

        .dashboard-subtitle {
          font-size: 14px;
          color: var(--td-text-color-secondary);
          margin: 0;
        }
      }
    }

    .header-actions {
      display: flex;
      gap: 12px;
    }
  }

  .section-title {
    font-size: 20px;
    font-weight: 600;
    color: var(--td-text-color-primary);
    margin-bottom: 16px;
    display: flex;
    align-items: center;
    gap: 8px;

    &::before {
      content: '';
      width: 4px;
      height: 20px;
      background: var(--td-brand-color);
      border-radius: 2px;
    }
  }

  // 系统概况部分
  .overview-section {
    margin-bottom: 32px;

    .overview-grid {
      display: grid;
      grid-template-columns: 2fr 1fr;
      gap: 24px;

      .overview-alerts {
        background: var(--td-bg-color-container);
        border: 1px solid var(--td-border-level-1-color);
        border-radius: 8px;
        padding: 20px;

        .alerts-title {
          display: flex;
          align-items: center;
          gap: 8px;
          font-size: 16px;
          font-weight: 600;
          color: var(--td-text-color-primary);
          margin-bottom: 16px;

          .t-icon {
            color: var(--td-brand-color);
          }
        }

        .no-alerts {
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 8px;
          padding: 20px;
          color: var(--td-success-color);
          font-size: 14px;

          .t-icon {
            font-size: 18px;
          }
        }

        .alerts-list {
          .alert-item {
            margin-bottom: 12px;

            &:last-child {
              margin-bottom: 0;
            }

            .alert-content {
              .alert-title {
                font-weight: 500;
                margin-bottom: 2px;
              }

              .alert-message {
                font-size: 12px;
                opacity: 0.9;
              }
            }
          }
        }
      }
    }
  }

  // 排行榜部分
  .rankings-section {
    margin-bottom: 32px;

    .rankings-grid {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 24px;

      .ranking-card {
        background: var(--td-bg-color-container);
        border: 1px solid var(--td-border-level-1-color);
        border-radius: 8px;
        padding: 20px;

        .ranking-header {
          display: flex;
          align-items: center;
          gap: 8px;
          margin-bottom: 16px;
          font-size: 16px;
          font-weight: 600;
          color: var(--td-text-color-primary);

          .t-icon {
            color: var(--td-brand-color);
          }

          .t-button {
            margin-left: auto;
          }
        }

        .ranking-loading,
        .ranking-empty {
          display: flex;
          align-items: center;
          justify-content: center;
          gap: 8px;
          padding: 40px 20px;
          color: var(--td-text-color-secondary);
        }

        .ranking-list {
          .ranking-item {
            display: flex;
            align-items: center;
            gap: 12px;
            padding: 12px;
            border-radius: 6px;
            margin-bottom: 8px;
            border: 1px solid transparent;
            transition: all 0.2s ease;

            &:hover {
              background: var(--td-bg-color-container-hover);
              border-color: var(--td-border-level-2-color);
            }

            &.ranking-warning {
              border-color: var(--td-warning-color-3);
              background: var(--td-warning-color-1);
            }

            &:last-child {
              margin-bottom: 0;
            }

            .ranking-index {
              display: flex;
              align-items: center;
              justify-content: center;
              width: 24px;
              height: 24px;
              flex-shrink: 0;

              .medal-1 {
                color: #ffd700;
              }

              .medal-2 {
                color: #c0c0c0;
              }

              .medal-3 {
                color: #cd7f32;
              }

              .rank-number {
                font-size: 12px;
                font-weight: 600;
                color: var(--td-text-color-secondary);
              }
            }

            .ranking-info {
              flex: 1;
              min-width: 0;

              .ranking-name {
                font-size: 14px;
                font-weight: 500;
                color: var(--td-text-color-primary);
                margin-bottom: 6px;
              }

              .ranking-usage {
                display: flex;
                flex-direction: column;
                gap: 4px;
              }
            }
          }
        }
      }
    }
  }

  // 图表部分
  .charts-section {
    margin-bottom: 32px;

    .charts-grid {
      .chart-container {
        background: var(--td-bg-color-container);
        border: 1px solid var(--td-border-level-1-color);
        border-radius: 8px;
        padding: 20px;
        height: 500px;
      }
    }
  }

  // 全局设置对话框
  .global-settings {
    .settings-notice {
      margin-bottom: 20px;
    }
  }
}

// 响应式设计
@media (max-width: 1200px) {
  .rpm-tpm-dashboard {
    .overview-section {
      .overview-grid {
        grid-template-columns: 1fr;
      }
    }

    .rankings-section {
      .rankings-grid {
        grid-template-columns: 1fr;
      }
    }
  }
}

@media (max-width: 768px) {
  .rpm-tpm-dashboard {
    padding: 16px;

    .dashboard-header {
      flex-direction: column;
      align-items: flex-start;
      gap: 16px;

      .header-content {
        .header-icon {
          font-size: 24px;
        }

        .header-text {
          .dashboard-title {
            font-size: 20px;
          }
        }
      }

      .header-actions {
        width: 100%;
        justify-content: stretch;

        .t-button {
          flex: 1;
        }
      }
    }

    .section-title {
      font-size: 16px;
    }

    .rankings-section {
      .rankings-grid {
        .ranking-card {
          .ranking-list {
            .ranking-item {
              .ranking-info {
                .ranking-usage {
                  gap: 6px;
                }
              }
            }
          }
        }
      }
    }

    .charts-section {
      .charts-grid {
        .chart-container {
          height: 350px;
          padding: 16px;
        }
      }
    }
  }
}
</style>
