<template>
  <div class="api-keys-rpm-tpm">
    <div class="page-header">
      <div class="header-content">
        <t-breadcrumb>
          <t-breadcrumb-item to="/dashboard">首页</t-breadcrumb-item>
          <t-breadcrumb-item to="/rpm-tpm/dashboard">RPM/TPM 监控</t-breadcrumb-item>
          <t-breadcrumb-item>API Keys 管理</t-breadcrumb-item>
        </t-breadcrumb>
        <div class="header-title">
          <t-icon name="key" class="title-icon" />
          <h1>API Keys RPM/TPM 管理</h1>
        </div>
        <p class="header-subtitle">管理和监控所有 API Key 的 RPM/TPM 限制和使用情况</p>
      </div>
      <div class="header-actions">
        <t-button variant="outline" size="small" :loading="refreshing" @click="handleRefresh">
          <template #icon>
            <t-icon name="refresh" />
          </template>
          刷新
        </t-button>
        <t-button theme="primary" size="small" @click="showBatchSettingsDialog">
          <template #icon>
            <t-icon name="setting" />
          </template>
          批量设置
        </t-button>
      </div>
    </div>

    <!-- 统计概览 -->
    <div class="stats-overview">
      <div class="stats-cards">
        <div class="stat-card">
          <div class="stat-icon">
            <t-icon name="key" />
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ totalStats.total_api_keys }}</div>
            <div class="stat-label">API Keys 总数</div>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon warning">
            <t-icon name="error-circle" />
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ totalStats.limited_api_keys }}</div>
            <div class="stat-label">限流中</div>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon success">
            <t-icon name="check-circle" />
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ totalStats.active_api_keys }}</div>
            <div class="stat-label">正常运行</div>
          </div>
        </div>
        <div class="stat-card">
          <div class="stat-icon info">
            <t-icon name="chart-line" />
          </div>
          <div class="stat-content">
            <div class="stat-value">{{ formatRpmTpmValue(totalStats.total_rpm) }}</div>
            <div class="stat-label">总 RPM</div>
          </div>
        </div>
      </div>
    </div>

    <!-- 筛选和搜索 -->
    <div class="filter-section">
      <div class="filter-left">
        <t-input
          v-model="searchKeyword"
          placeholder="搜索 API Key ID 或名称"
          style="width: 280px"
          clearable
          @change="handleSearch"
        >
          <template #prefix-icon>
            <t-icon name="search" />
          </template>
        </t-input>
        <t-select v-model="statusFilter" placeholder="状态筛选" style="width: 150px" clearable @change="handleFilter">
          <t-option value="normal" label="正常" />
          <t-option value="warning" label="告警" />
          <t-option value="limited" label="限流" />
          <t-option value="unlimited" label="无限制" />
        </t-select>
      </div>
      <div class="filter-right">
        <t-radio-group v-model="viewMode" variant="default-filled" @change="handleViewModeChange">
          <t-radio-button value="table">列表视图</t-radio-button>
          <t-radio-button value="grid">卡片视图</t-radio-button>
        </t-radio-group>
      </div>
    </div>

    <!-- 表格视图 -->
    <div v-if="viewMode === 'table'" class="table-view">
      <t-table
        :columns="tableColumns"
        :data="filteredApiKeys"
        :loading="loading"
        :pagination="paginationConfig"
        row-key="api_key_id"
        stripe
        hover
        @page-change="handlePageChange"
        @page-size-change="handlePageSizeChange"
        @sort-change="handleSortChange"
      >
        <!-- API Key 列 -->
        <template #api_key_id="{ row }">
          <div class="api-key-cell">
            <div class="api-key-id">{{ row.api_key_id }}</div>
            <div v-if="row.api_key_name" class="api-key-name">{{ row.api_key_name }}</div>
          </div>
        </template>

        <!-- RPM 状态列 -->
        <template #rpm_status="{ row }">
          <div class="status-cell">
            <div class="usage-info">
              <span class="current-value" :class="{ 'text-danger': row.is_rpm_limited }">
                {{ formatRpmTpmValue(row.current_rpm) }}
              </span>
              <span v-if="row.rpm_limit > 0" class="limit-value"> / {{ formatRpmTpmValue(row.rpm_limit) }} </span>
              <span v-else class="unlimited-text">无限制</span>
            </div>
            <usage-progress
              title="RPM"
              :current="row.current_rpm"
              :limit="row.rpm_limit"
              :warning-threshold="row.rpm_warning_threshold"
              :is-limited="row.is_rpm_limited"
              compact
            />
          </div>
        </template>

        <!-- TPM 状态列 -->
        <template #tpm_status="{ row }">
          <div class="status-cell">
            <div class="usage-info">
              <span class="current-value" :class="{ 'text-danger': row.is_tpm_limited }">
                {{ formatRpmTpmValue(row.current_tpm) }}
              </span>
              <span v-if="row.tpm_limit > 0" class="limit-value"> / {{ formatRpmTpmValue(row.tpm_limit) }} </span>
              <span v-else class="unlimited-text">无限制</span>
            </div>
            <usage-progress
              title="TPM"
              :current="row.current_tpm"
              :limit="row.tpm_limit"
              :warning-threshold="row.tpm_warning_threshold"
              :is-limited="row.is_tpm_limited"
              compact
            />
          </div>
        </template>

        <!-- 告警状态列 -->
        <template #alerts="{ row }">
          <alert-badge
            :rpm-current="row.current_rpm"
            :rpm-limit="row.rpm_limit"
            :rpm-warning-threshold="row.rpm_warning_threshold"
            :is-rpm-limited="row.is_rpm_limited"
            :tpm-current="row.current_tpm"
            :tpm-limit="row.tpm_limit"
            :tpm-warning-threshold="row.tpm_warning_threshold"
            :is-tpm-limited="row.is_tpm_limited"
            :is-rate-limited="!!row.rate_limit_end_time"
            :rate-limit-end-time="row.rate_limit_end_time"
          />
        </template>

        <!-- 操作列 -->
        <template #actions="{ row }">
          <div class="action-buttons">
            <t-button variant="text" size="small" @click="handleViewDetail(row)">
              <template #icon>
                <t-icon name="view" />
              </template>
            </t-button>
            <t-button variant="text" size="small" @click="handleManageLimits(row)">
              <template #icon>
                <t-icon name="setting" />
              </template>
            </t-button>
            <t-button variant="text" size="small" @click="handleViewHistory(row)">
              <template #icon>
                <t-icon name="chart-line" />
              </template>
            </t-button>
          </div>
        </template>
      </t-table>
    </div>

    <!-- 卡片视图 -->
    <div v-if="viewMode === 'grid'" class="grid-view">
      <div v-if="loading" class="grid-loading">
        <t-loading size="large" text="加载中..." />
      </div>
      <div v-else class="grid-container">
        <stats-card
          v-for="apiKey in paginatedGridData"
          :key="apiKey.api_key_id"
          :title="`API Key ${apiKey.api_key_id}`"
          :subtitle="apiKey.api_key_name"
          icon="key"
          :stats="apiKey"
          @refresh="() => handleRefreshApiKey(apiKey.api_key_id)"
          @manage="() => handleManageLimits(apiKey)"
          @view-history="() => handleViewHistory(apiKey)"
        />
      </div>
      <!-- 卡片视图分页 -->
      <div v-if="!loading" class="grid-pagination">
        <t-pagination
          v-model:current="gridPagination.current"
          v-model:page-size="gridPagination.pageSize"
          :total="filteredApiKeys.length"
          :page-size-options="[12, 24, 36, 48]"
          show-page-size
          show-jumper
        />
      </div>
    </div>

    <!-- 详情对话框 -->
    <t-dialog
      v-model:visible="detailVisible"
      :header="`API Key ${selectedApiKey?.api_key_id} 详细信息`"
      width="800px"
      @cancel="handleCloseDetail"
    >
      <div v-if="selectedApiKey" class="detail-content">
        <div class="detail-stats">
          <real-time-stats
            :stats="selectedApiKey"
            :auto-refresh="true"
            :refresh-interval="30000"
            @refresh="() => handleRefreshApiKey(selectedApiKey.api_key_id)"
          />
        </div>
        <div class="detail-history">
          <history-chart
            :title="`API Key ${selectedApiKey.api_key_id} 历史趋势`"
            :data="detailHistoryData"
            :loading="detailHistoryLoading"
            :rpm-limit="selectedApiKey.rpm_limit"
            :tpm-limit="selectedApiKey.tpm_limit"
            :rpm-warning-threshold="selectedApiKey.rpm_warning_threshold"
            :tpm-warning-threshold="selectedApiKey.tpm_warning_threshold"
            @time-range-change="handleDetailHistoryTimeRange"
            @refresh="handleRefreshDetailHistory"
          />
        </div>
      </div>
      <template #footer>
        <t-button @click="handleCloseDetail">关闭</t-button>
        <t-button theme="primary" @click="() => handleManageLimits(selectedApiKey)"> 管理限制 </t-button>
      </template>
    </t-dialog>

    <!-- 限制管理对话框 -->
    <t-dialog
      v-model:visible="limitsVisible"
      :header="`管理 API Key ${selectedApiKey?.api_key_id} 限制`"
      width="700px"
      @cancel="handleCloseLimits"
    >
      <div v-if="selectedApiKey" class="limits-content">
        <limits-form
          :initial-data="selectedApiKey"
          :current-stats="selectedApiKey"
          :loading="limitsSaving"
          submit-text="保存限制设置"
          @submit="handleSaveLimits"
          @cancel="handleCloseLimits"
        />
      </div>
    </t-dialog>

    <!-- 批量设置对话框 -->
    <t-dialog
      v-model:visible="batchSettingsVisible"
      header="批量设置 API Keys 限制"
      width="700px"
      @cancel="handleCloseBatchSettings"
    >
      <div class="batch-settings-content">
        <div class="batch-notice">
          <t-alert theme="warning" size="small">
            批量设置将应用到所有选中的 API Keys，请谨慎操作。当前选中 {{ selectedApiKeys.length }} 个 API Key。
          </t-alert>
        </div>
        <limits-form
          :loading="batchSettingsSaving"
          submit-text="批量应用设置"
          @submit="handleSaveBatchSettings"
          @cancel="handleCloseBatchSettings"
        />
      </div>
    </t-dialog>

    <!-- 历史趋势对话框 -->
    <t-dialog
      v-model:visible="historyVisible"
      :header="`API Key ${selectedApiKey?.api_key_id} 历史趋势`"
      width="1000px"
      @cancel="handleCloseHistory"
    >
      <div v-if="selectedApiKey" class="history-content">
        <history-chart
          :title="`API Key ${selectedApiKey.api_key_id} 历史趋势分析`"
          :data="historyData"
          :loading="historyLoading"
          :rpm-limit="selectedApiKey.rpm_limit"
          :tpm-limit="selectedApiKey.tpm_limit"
          :rpm-warning-threshold="selectedApiKey.rpm_warning_threshold"
          :tpm-warning-threshold="selectedApiKey.tpm_warning_threshold"
          @time-range-change="handleHistoryTimeRange"
          @refresh="handleRefreshHistory"
        />
      </div>
      <template #footer>
        <t-button @click="handleCloseHistory">关闭</t-button>
      </template>
    </t-dialog>
  </div>
</template>
<script setup lang="ts">
import type { TableColumns } from 'tdesign-vue-next';
import { computed, onMounted, onUnmounted, reactive, ref } from 'vue';

import type { GetRpmTpmHistoryParams, RpmTpmHistoryItem, RpmTpmStats, UpdateRpmTpmLimitsParams } from '@/api/rpm-tpm';
import {
  formatRpmTpmValue,
  getApiKeyRpmTpmHistory,
  getApiKeyRpmTpmRanking,
  getApiKeyRpmTpmStats,
  updateApiKeyRpmTpmLimits,
} from '@/api/rpm-tpm';
import AlertBadge from '@/components/rpm-tpm/AlertBadge.vue';
import RealTimeStats from '@/components/rpm-tpm/RealTimeStats.vue';
import UsageProgress from '@/components/rpm-tpm/UsageProgress.vue';

import HistoryChart from '../components/HistoryChart.vue';
import LimitsForm from '../components/LimitsForm.vue';
import StatsCard from '../components/StatsCard.vue';

interface TotalStats {
  total_api_keys: number;
  active_api_keys: number;
  limited_api_keys: number;
  total_rpm: number;
  total_tpm: number;
}

// 数据状态
const apiKeys = ref<RpmTpmStats[]>([]);
const loading = ref(false);
const refreshing = ref(false);

// 筛选和搜索
const searchKeyword = ref('');
const statusFilter = ref('');
const viewMode = ref<'table' | 'grid'>('table');

// 分页配置
const paginationConfig = reactive({
  current: 1,
  pageSize: 20,
  total: 0,
  showJumper: true,
  showPageSize: true,
  pageSizeOptions: [20, 50, 100],
});

// 卡片视图分页
const gridPagination = reactive({
  current: 1,
  pageSize: 12,
});

// 对话框状态
const detailVisible = ref(false);
const limitsVisible = ref(false);
const batchSettingsVisible = ref(false);
const historyVisible = ref(false);

// 选中的API Key
const selectedApiKey = ref<RpmTpmStats | null>(null);
const selectedApiKeys = ref<RpmTpmStats[]>([]);

// 历史数据
const detailHistoryData = ref<RpmTpmHistoryItem[]>([]);
const detailHistoryLoading = ref(false);
const historyData = ref<RpmTpmHistoryItem[]>([]);
const historyLoading = ref(false);

// 保存状态
const limitsSaving = ref(false);
const batchSettingsSaving = ref(false);

// 自动刷新定时器
let refreshTimer: NodeJS.Timeout | null = null;

// 表格列配置
const tableColumns: TableColumns = [
  {
    colKey: 'api_key_id',
    title: 'API Key',
    width: 200,
    fixed: 'left',
    sorter: true,
  },
  {
    colKey: 'rpm_status',
    title: 'RPM 状态',
    width: 200,
  },
  {
    colKey: 'tpm_status',
    title: 'TPM 状态',
    width: 200,
  },
  {
    colKey: 'alerts',
    title: '告警状态',
    width: 120,
    align: 'center',
  },
  {
    colKey: 'actions',
    title: '操作',
    width: 150,
    align: 'center',
    fixed: 'right',
  },
];

// 计算属性
const totalStats = computed<TotalStats>(() => {
  const total = apiKeys.value.length;
  const active = apiKeys.value.filter((item) => !item.is_rpm_limited && !item.is_tpm_limited).length;
  const limited = apiKeys.value.filter((item) => item.is_rpm_limited || item.is_tpm_limited).length;
  const totalRpm = apiKeys.value.reduce((sum, item) => sum + item.current_rpm, 0);
  const totalTpm = apiKeys.value.reduce((sum, item) => sum + item.current_tpm, 0);

  return {
    total_api_keys: total,
    active_api_keys: active,
    limited_api_keys: limited,
    total_rpm: totalRpm,
    total_tpm: totalTpm,
  };
});

// 过滤后的数据
const filteredApiKeys = computed(() => {
  let filtered = [...apiKeys.value];

  // 关键词搜索
  if (searchKeyword.value) {
    const keyword = searchKeyword.value.toLowerCase();
    filtered = filtered.filter(
      (item) => item.api_key_id?.toString().includes(keyword) || item.api_key_name?.toLowerCase().includes(keyword),
    );
  }

  // 状态筛选
  if (statusFilter.value) {
    switch (statusFilter.value) {
      case 'normal':
        filtered = filtered.filter((item) => !item.is_rpm_limited && !item.is_tpm_limited);
        break;
      case 'warning':
        filtered = filtered.filter(
          (item) =>
            (item.rpm_limit > 0 && item.current_rpm >= item.rpm_warning_threshold) ||
            (item.tpm_limit > 0 && item.current_tpm >= item.tpm_warning_threshold),
        );
        break;
      case 'limited':
        filtered = filtered.filter((item) => item.is_rpm_limited || item.is_tpm_limited);
        break;
      case 'unlimited':
        filtered = filtered.filter((item) => item.rpm_limit === 0 && item.tpm_limit === 0);
        break;
    }
  }

  return filtered;
});

// 卡片视图分页数据
const paginatedGridData = computed(() => {
  const start = (gridPagination.current - 1) * gridPagination.pageSize;
  const end = start + gridPagination.pageSize;
  return filteredApiKeys.value.slice(start, end);
});

// 加载API Keys数据
const loadApiKeys = async () => {
  try {
    loading.value = true;
    const response = await getApiKeyRpmTpmRanking({
      limit: 1000, // 加载所有数据
    });
    apiKeys.value = response.data;
    paginationConfig.total = response.data.length;
  } catch (error) {
    console.error('Failed to load API keys:', error);
  } finally {
    loading.value = false;
  }
};

// 刷新数据
const handleRefresh = async () => {
  refreshing.value = true;
  try {
    await loadApiKeys();
  } finally {
    refreshing.value = false;
  }
};

// 刷新单个API Key
const handleRefreshApiKey = async (apiKeyId: number) => {
  try {
    const response = await getApiKeyRpmTpmStats(apiKeyId);
    const index = apiKeys.value.findIndex((item) => item.api_key_id === apiKeyId);
    if (index !== -1) {
      apiKeys.value[index] = response.data;
    }

    // 如果是当前选中的API Key，也更新选中项
    if (selectedApiKey.value?.api_key_id === apiKeyId) {
      selectedApiKey.value = response.data;
    }
  } catch (error) {
    console.error('Failed to refresh API key:', error);
  }
};

// 搜索处理
const handleSearch = () => {
  paginationConfig.current = 1;
  gridPagination.current = 1;
};

// 筛选处理
const handleFilter = () => {
  paginationConfig.current = 1;
  gridPagination.current = 1;
};

// 视图模式切换
const handleViewModeChange = () => {
  // 重置分页
  paginationConfig.current = 1;
  gridPagination.current = 1;
};

// 分页处理
const handlePageChange = (current: number) => {
  paginationConfig.current = current;
};

const handlePageSizeChange = (pageSize: number) => {
  paginationConfig.pageSize = pageSize;
  paginationConfig.current = 1;
};

// 排序处理
const handleSortChange = (sortInfo: any) => {
  console.log('Sort change:', sortInfo);
  // TODO: 实现排序逻辑
};

// 查看详情
const handleViewDetail = async (apiKey: RpmTpmStats) => {
  selectedApiKey.value = apiKey;
  detailVisible.value = true;

  // 加载详情历史数据
  await loadDetailHistory();
};

// 关闭详情
const handleCloseDetail = () => {
  detailVisible.value = false;
  selectedApiKey.value = null;
  detailHistoryData.value = [];
};

// 管理限制
const handleManageLimits = (apiKey: RpmTpmStats) => {
  selectedApiKey.value = apiKey;
  limitsVisible.value = true;
};

// 关闭限制管理
const handleCloseLimits = () => {
  limitsVisible.value = false;
  selectedApiKey.value = null;
};

// 保存限制设置
const handleSaveLimits = async (data: UpdateRpmTpmLimitsParams) => {
  if (!selectedApiKey.value?.api_key_id) return;

  try {
    limitsSaving.value = true;
    await updateApiKeyRpmTpmLimits(selectedApiKey.value.api_key_id, data);

    // 刷新数据
    await handleRefreshApiKey(selectedApiKey.value.api_key_id);

    // 关闭对话框
    handleCloseLimits();
  } catch (error) {
    console.error('Failed to save limits:', error);
  } finally {
    limitsSaving.value = false;
  }
};

// 查看历史
const handleViewHistory = async (apiKey: RpmTpmStats) => {
  selectedApiKey.value = apiKey;
  historyVisible.value = true;

  // 加载历史数据
  await loadHistory();
};

// 关闭历史
const handleCloseHistory = () => {
  historyVisible.value = false;
  selectedApiKey.value = null;
  historyData.value = [];
};

// 显示批量设置对话框
const showBatchSettingsDialog = () => {
  // TODO: 实现多选逻辑
  selectedApiKeys.value = [];
  batchSettingsVisible.value = true;
};

// 关闭批量设置
const handleCloseBatchSettings = () => {
  batchSettingsVisible.value = false;
  selectedApiKeys.value = [];
};

// 保存批量设置
const handleSaveBatchSettings = async (data: UpdateRpmTpmLimitsParams) => {
  try {
    batchSettingsSaving.value = true;

    // TODO: 实现批量更新API
    const promises = selectedApiKeys.value.map((apiKey) => updateApiKeyRpmTpmLimits(apiKey.api_key_id!, data));

    await Promise.all(promises);

    // 刷新数据
    await loadApiKeys();

    // 关闭对话框
    handleCloseBatchSettings();
  } catch (error) {
    console.error('Failed to save batch settings:', error);
  } finally {
    batchSettingsSaving.value = false;
  }
};

// 加载详情历史数据
const loadDetailHistory = async (params: GetRpmTpmHistoryParams = {}) => {
  if (!selectedApiKey.value?.api_key_id) return;

  try {
    detailHistoryLoading.value = true;
    const response = await getApiKeyRpmTpmHistory(selectedApiKey.value.api_key_id, {
      start_time: params.start_time || new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
      end_time: params.end_time || new Date().toISOString(),
      ...params,
    });
    detailHistoryData.value = response.data;
  } catch (error) {
    console.error('Failed to load detail history:', error);
  } finally {
    detailHistoryLoading.value = false;
  }
};

// 加载历史数据
const loadHistory = async (params: GetRpmTpmHistoryParams = {}) => {
  if (!selectedApiKey.value?.api_key_id) return;

  try {
    historyLoading.value = true;
    const response = await getApiKeyRpmTpmHistory(selectedApiKey.value.api_key_id, {
      start_time: params.start_time || new Date(Date.now() - 24 * 60 * 60 * 1000).toISOString(),
      end_time: params.end_time || new Date().toISOString(),
      ...params,
    });
    historyData.value = response.data;
  } catch (error) {
    console.error('Failed to load history:', error);
  } finally {
    historyLoading.value = false;
  }
};

// 处理详情历史时间范围变化
const handleDetailHistoryTimeRange = (params: { start_time: string; end_time: string }) => {
  loadDetailHistory(params);
};

// 刷新详情历史数据
const handleRefreshDetailHistory = () => {
  loadDetailHistory();
};

// 处理历史时间范围变化
const handleHistoryTimeRange = (params: { start_time: string; end_time: string }) => {
  loadHistory(params);
};

// 刷新历史数据
const handleRefreshHistory = () => {
  loadHistory();
};

// 启动自动刷新
const startAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer);
  }
  refreshTimer = setInterval(() => {
    handleRefresh();
  }, 30000);
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
  loadApiKeys();
  startAutoRefresh();
});

onUnmounted(() => {
  stopAutoRefresh();
});
</script>
<style lang="less" scoped>
.api-keys-rpm-tpm {
  padding: 24px;
  max-width: 1600px;
  margin: 0 auto;

  .page-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 24px;
    padding-bottom: 20px;
    border-bottom: 1px solid var(--td-border-level-1-color);

    .header-content {
      flex: 1;

      .header-title {
        display: flex;
        align-items: center;
        gap: 12px;
        margin: 12px 0 8px 0;

        .title-icon {
          font-size: 24px;
          color: var(--td-brand-color);
        }

        h1 {
          font-size: 24px;
          font-weight: 600;
          color: var(--td-text-color-primary);
          margin: 0;
        }
      }

      .header-subtitle {
        color: var(--td-text-color-secondary);
        font-size: 14px;
        margin: 0;
      }
    }

    .header-actions {
      display: flex;
      gap: 12px;
    }
  }

  // 统计概览
  .stats-overview {
    margin-bottom: 24px;

    .stats-cards {
      display: grid;
      grid-template-columns: repeat(4, 1fr);
      gap: 16px;

      .stat-card {
        background: var(--td-bg-color-container);
        border: 1px solid var(--td-border-level-1-color);
        border-radius: 8px;
        padding: 20px;
        display: flex;
        align-items: center;
        gap: 16px;

        .stat-icon {
          width: 48px;
          height: 48px;
          border-radius: 8px;
          display: flex;
          align-items: center;
          justify-content: center;
          background: var(--td-brand-color-1);
          color: var(--td-brand-color);

          &.warning {
            background: var(--td-warning-color-1);
            color: var(--td-warning-color);
          }

          &.success {
            background: var(--td-success-color-1);
            color: var(--td-success-color);
          }

          &.info {
            background: var(--td-info-color-1);
            color: var(--td-info-color);
          }

          .t-icon {
            font-size: 24px;
          }
        }

        .stat-content {
          .stat-value {
            font-size: 24px;
            font-weight: 600;
            color: var(--td-text-color-primary);
            margin-bottom: 4px;
          }

          .stat-label {
            font-size: 12px;
            color: var(--td-text-color-secondary);
          }
        }
      }
    }
  }

  // 筛选部分
  .filter-section {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;

    .filter-left {
      display: flex;
      gap: 12px;
    }

    .filter-right {
      display: flex;
      align-items: center;
      gap: 12px;
    }
  }

  // 表格视图
  .table-view {
    .api-key-cell {
      .api-key-id {
        font-weight: 500;
        color: var(--td-text-color-primary);
      }

      .api-key-name {
        font-size: 12px;
        color: var(--td-text-color-secondary);
        margin-top: 2px;
      }
    }

    .status-cell {
      .usage-info {
        display: flex;
        align-items: center;
        gap: 4px;
        margin-bottom: 6px;
        font-size: 13px;

        .current-value {
          font-weight: 500;
          color: var(--td-text-color-primary);

          &.text-danger {
            color: var(--td-error-color);
          }
        }

        .limit-value {
          color: var(--td-text-color-secondary);
        }

        .unlimited-text {
          color: var(--td-text-color-placeholder);
          font-style: italic;
        }
      }
    }

    .action-buttons {
      display: flex;
      gap: 4px;
    }
  }

  // 卡片视图
  .grid-view {
    .grid-loading {
      display: flex;
      justify-content: center;
      align-items: center;
      padding: 60px;
    }

    .grid-container {
      display: grid;
      grid-template-columns: repeat(auto-fill, minmax(320px, 1fr));
      gap: 20px;
      margin-bottom: 24px;
    }

    .grid-pagination {
      display: flex;
      justify-content: center;
      padding: 20px 0;
    }
  }

  // 对话框内容
  .detail-content {
    .detail-stats {
      margin-bottom: 24px;
    }

    .detail-history {
      height: 400px;
    }
  }

  .limits-content {
    padding: 12px 0;
  }

  .batch-settings-content {
    .batch-notice {
      margin-bottom: 20px;
    }
  }

  .history-content {
    height: 500px;
  }
}

// 响应式设计
@media (max-width: 1200px) {
  .api-keys-rpm-tpm {
    .stats-overview {
      .stats-cards {
        grid-template-columns: repeat(2, 1fr);
      }
    }

    .grid-view {
      .grid-container {
        grid-template-columns: repeat(auto-fill, minmax(280px, 1fr));
        gap: 16px;
      }
    }
  }
}

@media (max-width: 768px) {
  .api-keys-rpm-tpm {
    padding: 16px;

    .page-header {
      flex-direction: column;
      align-items: stretch;
      gap: 16px;

      .header-actions {
        width: 100%;

        .t-button {
          flex: 1;
        }
      }
    }

    .stats-overview {
      .stats-cards {
        grid-template-columns: 1fr;
        gap: 12px;

        .stat-card {
          padding: 16px;
        }
      }
    }

    .filter-section {
      flex-direction: column;
      align-items: stretch;
      gap: 12px;

      .filter-left {
        flex-direction: column;
        gap: 8px;

        .t-input,
        .t-select {
          width: 100% !important;
        }
      }

      .filter-right {
        justify-content: center;
      }
    }

    .grid-view {
      .grid-container {
        grid-template-columns: 1fr;
        gap: 12px;
      }
    }
  }
}
</style>
