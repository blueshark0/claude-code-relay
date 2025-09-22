<template>
  <div class="real-time-stats">
    <div class="stats-grid">
      <!-- RPM 统计 -->
      <div class="stat-item">
        <div class="stat-header">
          <span class="stat-label">当前 RPM</span>
          <t-tag v-if="stats.is_rpm_limited" theme="danger" size="small">限流中</t-tag>
        </div>
        <div class="stat-content">
          <div class="stat-value" :class="{ 'text-danger': stats.is_rpm_limited }">
            {{ formatRpmTpmValue(stats.current_rpm) }}
          </div>
          <div v-if="stats.rpm_limit > 0" class="stat-limit">/ {{ formatRpmTpmValue(stats.rpm_limit) }}</div>
        </div>
        <div v-if="stats.rpm_limit > 0" class="stat-progress">
          <t-progress
            :percentage="stats.rpm_usage_percentage"
            :status="getUsageTheme(stats.rpm_usage_percentage)"
            size="small"
            :stroke-width="6"
          />
          <span class="progress-text">{{ stats.rpm_usage_percentage.toFixed(1) }}%</span>
        </div>
      </div>

      <!-- TPM 统计 -->
      <div class="stat-item">
        <div class="stat-header">
          <span class="stat-label">当前 TPM</span>
          <t-tag v-if="stats.is_tpm_limited" theme="danger" size="small">限流中</t-tag>
        </div>
        <div class="stat-content">
          <div class="stat-value" :class="{ 'text-danger': stats.is_tpm_limited }">
            {{ formatRpmTpmValue(stats.current_tpm) }}
          </div>
          <div v-if="stats.tpm_limit > 0" class="stat-limit">/ {{ formatRpmTpmValue(stats.tpm_limit) }}</div>
        </div>
        <div v-if="stats.tpm_limit > 0" class="stat-progress">
          <t-progress
            :percentage="stats.tpm_usage_percentage"
            :status="getUsageTheme(stats.tpm_usage_percentage)"
            size="small"
            :stroke-width="6"
          />
          <span class="progress-text">{{ stats.tpm_usage_percentage.toFixed(1) }}%</span>
        </div>
      </div>

      <!-- 历史峰值 -->
      <div class="stat-item">
        <div class="stat-header">
          <span class="stat-label">历史峰值</span>
        </div>
        <div class="stat-content">
          <div class="peak-stats">
            <div class="peak-item">
              <span class="peak-label">RPM:</span>
              <span class="peak-value">{{ formatRpmTpmValue(stats.max_rpm) }}</span>
            </div>
            <div class="peak-item">
              <span class="peak-label">TPM:</span>
              <span class="peak-value">{{ formatRpmTpmValue(stats.max_tpm) }}</span>
            </div>
          </div>
        </div>
      </div>

      <!-- 限流状态 -->
      <div v-if="stats.rate_limit_end_time" class="stat-item rate-limit-info">
        <div class="stat-header">
          <span class="stat-label">限流状态</span>
        </div>
        <div class="stat-content">
          <div class="limit-status">
            <t-icon name="time" class="limit-icon" />
            <div class="limit-text">
              <div class="limit-until">限流至</div>
              <div class="limit-time">{{ formatDateTime(stats.rate_limit_end_time) }}</div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- 刷新按钮 -->
    <div class="stats-actions">
      <t-button variant="text" size="small" :loading="loading" @click="$emit('refresh')">
        <template #icon>
          <t-icon name="refresh" />
        </template>
        刷新
      </t-button>
      <span class="last-update"> 最后更新: {{ lastUpdateTime }} </span>
    </div>
  </div>
</template>
<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue';

import type { RpmTpmStats } from '@/api/rpm-tpm';
import { formatRpmTpmValue, getUsageTheme } from '@/api/rpm-tpm';

interface Props {
  stats: RpmTpmStats;
  loading?: boolean;
  autoRefresh?: boolean;
  refreshInterval?: number;
}

interface Emits {
  (event: 'refresh'): void;
}

const props = withDefaults(defineProps<Props>(), {
  loading: false,
  autoRefresh: true,
  refreshInterval: 30000, // 30秒
});

const emit = defineEmits<Emits>();

// 最后更新时间
const lastUpdateTime = ref('');
let refreshTimer: NodeJS.Timeout | null = null;

// 格式化日期时间
const formatDateTime = (dateStr: string): string => {
  if (!dateStr) return '';
  const date = new Date(dateStr);
  if (Number.isNaN(date.getTime())) return '';
  return date.toLocaleString('zh-CN');
};

// 更新最后刷新时间
const updateLastRefreshTime = () => {
  lastUpdateTime.value = new Date().toLocaleString('zh-CN');
};

// 自动刷新逻辑
const startAutoRefresh = () => {
  if (!props.autoRefresh) return;

  refreshTimer = setInterval(() => {
    emit('refresh');
  }, props.refreshInterval);
};

const stopAutoRefresh = () => {
  if (refreshTimer) {
    clearInterval(refreshTimer);
    refreshTimer = null;
  }
};

// 生命周期
onMounted(() => {
  updateLastRefreshTime();
  startAutoRefresh();
});

onUnmounted(() => {
  stopAutoRefresh();
});

// 监听 stats 变化，更新时间
computed(() => {
  updateLastRefreshTime();
  return props.stats;
});
</script>
<style lang="less" scoped>
.real-time-stats {
  .stats-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(280px, 1fr));
    gap: 16px;
    margin-bottom: 16px;

    .stat-item {
      padding: 16px;
      background: var(--td-bg-color-container);
      border: 1px solid var(--td-border-level-1-color);
      border-radius: 8px;
      transition: all 0.2s ease;

      &:hover {
        border-color: var(--td-border-level-2-color);
        box-shadow: 0 2px 8px rgba(0, 0, 0, 0.08);
      }

      &.rate-limit-info {
        border-color: var(--td-error-color);
        background: var(--td-error-color-1);
      }
    }

    .stat-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 12px;

      .stat-label {
        font-size: 14px;
        color: var(--td-text-color-secondary);
        font-weight: 500;
      }
    }

    .stat-content {
      display: flex;
      align-items: baseline;
      margin-bottom: 8px;

      .stat-value {
        font-size: 28px;
        font-weight: 600;
        color: var(--td-text-color-primary);
        line-height: 1;

        &.text-danger {
          color: var(--td-error-color);
        }
      }

      .stat-limit {
        font-size: 16px;
        color: var(--td-text-color-secondary);
        margin-left: 8px;
      }
    }

    .stat-progress {
      display: flex;
      align-items: center;
      gap: 8px;

      :deep(.t-progress) {
        flex: 1;
      }

      .progress-text {
        font-size: 12px;
        color: var(--td-text-color-secondary);
        min-width: 40px;
        text-align: right;
      }
    }

    .peak-stats {
      display: flex;
      flex-direction: column;
      gap: 4px;

      .peak-item {
        display: flex;
        justify-content: space-between;
        align-items: center;

        .peak-label {
          font-size: 14px;
          color: var(--td-text-color-secondary);
        }

        .peak-value {
          font-size: 16px;
          font-weight: 600;
          color: var(--td-text-color-primary);
        }
      }
    }

    .limit-status {
      display: flex;
      align-items: center;
      gap: 8px;

      .limit-icon {
        color: var(--td-error-color);
        font-size: 16px;
      }

      .limit-text {
        .limit-until {
          font-size: 12px;
          color: var(--td-text-color-secondary);
        }

        .limit-time {
          font-size: 14px;
          font-weight: 500;
          color: var(--td-error-color);
        }
      }
    }
  }

  .stats-actions {
    display: flex;
    justify-content: space-between;
    align-items: center;
    padding: 12px 0;
    border-top: 1px solid var(--td-border-level-1-color);

    .last-update {
      font-size: 12px;
      color: var(--td-text-color-placeholder);
    }
  }
}

// 响应式设计
@media (max-width: 768px) {
  .real-time-stats {
    .stats-grid {
      grid-template-columns: 1fr;
      gap: 12px;

      .stat-item {
        padding: 12px;
      }

      .stat-content {
        .stat-value {
          font-size: 24px;
        }
      }
    }
  }
}
</style>
