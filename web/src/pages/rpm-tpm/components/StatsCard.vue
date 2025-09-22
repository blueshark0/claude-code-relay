<template>
  <t-card class="stats-card" :loading="loading">
    <template #header>
      <div class="card-header">
        <div class="header-content">
          <t-icon :name="icon" class="header-icon" />
          <div class="header-text">
            <div class="header-title">{{ title }}</div>
            <div v-if="subtitle" class="header-subtitle">{{ subtitle }}</div>
          </div>
        </div>
        <div class="header-actions">
          <alert-badge
            :rpm-current="stats.current_rpm"
            :rpm-limit="stats.rpm_limit"
            :rpm-warning-threshold="stats.rpm_warning_threshold"
            :is-rpm-limited="stats.is_rpm_limited"
            :tpm-current="stats.current_tpm"
            :tpm-limit="stats.tpm_limit"
            :tpm-warning-threshold="stats.tpm_warning_threshold"
            :is-tpm-limited="stats.is_tpm_limited"
            :is-rate-limited="!!stats.rate_limit_end_time"
            :rate-limit-end-time="stats.rate_limit_end_time"
            simple
          />
          <t-button variant="text" size="small" :loading="loading" @click="$emit('refresh')">
            <template #icon>
              <t-icon name="refresh" />
            </template>
          </t-button>
        </div>
      </div>
    </template>

    <div class="card-content">
      <!-- 主要统计数据 -->
      <div class="main-stats">
        <div class="stat-group">
          <div class="stat-item">
            <div class="stat-label">当前 RPM</div>
            <div class="stat-value" :class="{ 'text-danger': stats.is_rpm_limited }">
              {{ formatRpmTpmValue(stats.current_rpm) }}
              <span v-if="stats.rpm_limit > 0" class="stat-limit"> / {{ formatRpmTpmValue(stats.rpm_limit) }} </span>
            </div>
            <usage-progress
              title="RPM"
              :current="stats.current_rpm"
              :limit="stats.rpm_limit"
              :warning-threshold="stats.rpm_warning_threshold"
              :is-limited="stats.is_rpm_limited"
              compact
            />
          </div>

          <div class="stat-item">
            <div class="stat-label">当前 TPM</div>
            <div class="stat-value" :class="{ 'text-danger': stats.is_tpm_limited }">
              {{ formatRpmTpmValue(stats.current_tpm) }}
              <span v-if="stats.tpm_limit > 0" class="stat-limit"> / {{ formatRpmTpmValue(stats.tpm_limit) }} </span>
            </div>
            <usage-progress
              title="TPM"
              :current="stats.current_tpm"
              :limit="stats.tpm_limit"
              :warning-threshold="stats.tpm_warning_threshold"
              :is-limited="stats.is_tpm_limited"
              compact
            />
          </div>
        </div>

        <!-- 峰值数据 -->
        <div class="peak-stats">
          <div class="peak-item">
            <span class="peak-label">历史峰值:</span>
            <span class="peak-values">
              RPM {{ formatRpmTpmValue(stats.max_rpm) }} / TPM {{ formatRpmTpmValue(stats.max_tpm) }}
            </span>
          </div>
        </div>
      </div>

      <!-- 限流状态提示 -->
      <div v-if="stats.rate_limit_end_time" class="rate-limit-notice">
        <t-alert theme="error" size="small">
          <template #icon>
            <t-icon name="time" />
          </template>
          <div class="limit-info">
            <div class="limit-title">当前处于限流状态</div>
            <div class="limit-time">限流至: {{ formatDateTime(stats.rate_limit_end_time) }}</div>
          </div>
        </t-alert>
      </div>

      <!-- 操作按钮 -->
      <div class="card-actions">
        <t-button v-if="showManageButton" variant="outline" size="small" @click="$emit('manage')">
          <template #icon>
            <t-icon name="setting" />
          </template>
          管理限制
        </t-button>
        <t-button v-if="showHistoryButton" variant="text" size="small" @click="$emit('view-history')">
          <template #icon>
            <t-icon name="chart-line" />
          </template>
          查看历史
        </t-button>
      </div>
    </div>
  </t-card>
</template>
<script setup lang="ts">
import type { RpmTpmStats } from '@/api/rpm-tpm';
import { formatRpmTpmValue } from '@/api/rpm-tpm';
import AlertBadge from '@/components/rpm-tpm/AlertBadge.vue';
import UsageProgress from '@/components/rpm-tpm/UsageProgress.vue';

interface Props {
  title: string;
  subtitle?: string;
  icon: string;
  stats: RpmTpmStats;
  loading?: boolean;
  showManageButton?: boolean;
  showHistoryButton?: boolean;
}

interface Emits {
  (event: 'refresh'): void;
  (event: 'manage'): void;
  (event: 'view-history'): void;
}

withDefaults(defineProps<Props>(), {
  subtitle: '',
  loading: false,
  showManageButton: true,
  showHistoryButton: true,
});

defineEmits<Emits>();

// 格式化日期时间
const formatDateTime = (dateStr: string): string => {
  if (!dateStr) return '';
  const date = new Date(dateStr);
  if (Number.isNaN(date.getTime())) return '';
  return date.toLocaleString('zh-CN');
};
</script>
<style lang="less" scoped>
.stats-card {
  height: 100%;

  .card-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    width: 100%;

    .header-content {
      display: flex;
      align-items: center;
      gap: 12px;

      .header-icon {
        font-size: 20px;
        color: var(--td-brand-color);
      }

      .header-text {
        .header-title {
          font-size: 16px;
          font-weight: 600;
          color: var(--td-text-color-primary);
          line-height: 1.2;
        }

        .header-subtitle {
          font-size: 12px;
          color: var(--td-text-color-secondary);
          margin-top: 2px;
        }
      }
    }

    .header-actions {
      display: flex;
      align-items: center;
      gap: 8px;
    }
  }

  .card-content {
    .main-stats {
      .stat-group {
        display: grid;
        grid-template-columns: 1fr 1fr;
        gap: 20px;
        margin-bottom: 16px;

        .stat-item {
          .stat-label {
            font-size: 12px;
            color: var(--td-text-color-secondary);
            margin-bottom: 4px;
          }

          .stat-value {
            font-size: 20px;
            font-weight: 600;
            color: var(--td-text-color-primary);
            margin-bottom: 8px;
            display: flex;
            align-items: baseline;
            gap: 4px;

            &.text-danger {
              color: var(--td-error-color);
            }

            .stat-limit {
              font-size: 14px;
              color: var(--td-text-color-secondary);
              font-weight: normal;
            }
          }
        }
      }

      .peak-stats {
        padding: 12px;
        background: var(--td-bg-color-container-hover);
        border-radius: 6px;
        margin-bottom: 16px;

        .peak-item {
          display: flex;
          justify-content: space-between;
          align-items: center;
          font-size: 12px;

          .peak-label {
            color: var(--td-text-color-secondary);
          }

          .peak-values {
            color: var(--td-text-color-primary);
            font-weight: 500;
          }
        }
      }
    }

    .rate-limit-notice {
      margin-bottom: 16px;

      .limit-info {
        .limit-title {
          font-weight: 500;
          margin-bottom: 2px;
        }

        .limit-time {
          font-size: 12px;
          opacity: 0.8;
        }
      }
    }

    .card-actions {
      display: flex;
      gap: 8px;
      justify-content: flex-end;
      padding-top: 12px;
      border-top: 1px solid var(--td-border-level-1-color);
    }
  }

  // 加载状态样式
  &.t-loading {
    .card-content {
      opacity: 0.6;
    }
  }
}

// 响应式设计
@media (max-width: 768px) {
  .stats-card {
    .card-header {
      .header-content {
        gap: 8px;

        .header-icon {
          font-size: 18px;
        }

        .header-text {
          .header-title {
            font-size: 14px;
          }
        }
      }
    }

    .card-content {
      .main-stats {
        .stat-group {
          grid-template-columns: 1fr;
          gap: 16px;

          .stat-item {
            .stat-value {
              font-size: 18px;
            }
          }
        }
      }

      .card-actions {
        flex-direction: column;
        align-items: stretch;

        .t-button {
          width: 100%;
        }
      }
    }
  }
}
</style>
