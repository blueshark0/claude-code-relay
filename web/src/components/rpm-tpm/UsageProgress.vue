<template>
  <div class="usage-progress">
    <!-- 简洁模式 -->
    <div v-if="compact" class="progress-compact">
      <div class="progress-info">
        <span class="progress-current">{{ formatRpmTpmValue(current) }}</span>
        <span v-if="limit > 0" class="progress-limit">/ {{ formatRpmTpmValue(limit) }}</span>
        <span v-if="limit > 0" class="progress-percentage">({{ percentage.toFixed(1) }}%)</span>
      </div>
      <t-progress
        v-if="limit > 0"
        :percentage="percentage"
        :status="progressStatus"
        size="small"
        :stroke-width="4"
        :show-info="false"
      />
    </div>

    <!-- 完整模式 -->
    <div v-else class="progress-full">
      <div class="progress-header">
        <div class="progress-title">
          <span>{{ title }}</span>
          <t-tag v-if="isLimited" theme="danger" size="small">{{ limitedText }}</t-tag>
        </div>
        <div class="progress-values">
          <span class="current-value" :class="{ 'text-danger': isLimited }">
            {{ formatRpmTpmValue(current) }}
          </span>
          <span v-if="limit > 0" class="limit-value">/ {{ formatRpmTpmValue(limit) }}</span>
        </div>
      </div>

      <div v-if="limit > 0" class="progress-bar-container">
        <t-progress :percentage="percentage" :status="progressStatus" :stroke-width="8" :show-info="false" />
        <div class="progress-details">
          <span class="progress-percentage">{{ percentage.toFixed(1) }}%</span>
          <span v-if="warningThreshold > 0" class="warning-threshold">
            告警阈值: {{ formatRpmTpmValue(warningThreshold) }}
          </span>
        </div>
      </div>

      <!-- 无限制提示 -->
      <div v-else class="no-limit-hint">
        <t-icon name="infinite" />
        <span>无限制</span>
      </div>

      <!-- 警告信息 -->
      <div v-if="showWarning && warningMessage" class="warning-message">
        <t-alert :theme="warningTheme" size="small" :message="warningMessage" />
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { computed } from 'vue';

import { formatRpmTpmValue, getUsageTheme } from '@/api/rpm-tpm';

interface Props {
  title?: string;
  current: number;
  limit: number;
  warningThreshold?: number;
  isLimited?: boolean;
  limitedText?: string;
  compact?: boolean;
  showWarning?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  title: '',
  warningThreshold: 0,
  isLimited: false,
  limitedText: '限流中',
  compact: false,
  showWarning: true,
});

// 计算使用百分比
const percentage = computed(() => {
  if (props.limit <= 0) return 0;
  return Math.min((props.current / props.limit) * 100, 100);
});

// 进度条状态
const progressStatus = computed(() => {
  if (props.isLimited) return 'error';
  return getUsageTheme(percentage.value);
});

// 警告消息
const warningMessage = computed(() => {
  if (props.limit <= 0) return '';

  if (props.isLimited) {
    return `当前${props.title}已达到限制值，请求被限流`;
  }

  if (props.warningThreshold > 0 && props.current >= props.warningThreshold) {
    return `当前${props.title}已达到告警阈值 ${formatRpmTpmValue(props.warningThreshold)}`;
  }

  if (percentage.value >= 95) {
    return `当前${props.title}使用率已达到 ${percentage.value.toFixed(1)}%，接近限制值`;
  }

  if (percentage.value >= 80) {
    return `当前${props.title}使用率为 ${percentage.value.toFixed(1)}%，请注意监控`;
  }

  return '';
});

// 警告主题
const warningTheme = computed(() => {
  if (props.isLimited || percentage.value >= 95) return 'error';
  if (percentage.value >= 80 || (props.warningThreshold > 0 && props.current >= props.warningThreshold)) {
    return 'warning';
  }
  return 'info';
});
</script>
<style lang="less" scoped>
.usage-progress {
  width: 100%;

  // 简洁模式样式
  .progress-compact {
    .progress-info {
      display: flex;
      align-items: center;
      gap: 4px;
      margin-bottom: 4px;
      font-size: 12px;

      .progress-current {
        font-weight: 600;
        color: var(--td-text-color-primary);
      }

      .progress-limit {
        color: var(--td-text-color-secondary);
      }

      .progress-percentage {
        color: var(--td-text-color-placeholder);
        font-size: 11px;
      }
    }
  }

  // 完整模式样式
  .progress-full {
    .progress-header {
      display: flex;
      justify-content: space-between;
      align-items: center;
      margin-bottom: 8px;

      .progress-title {
        display: flex;
        align-items: center;
        gap: 8px;
        font-size: 14px;
        font-weight: 500;
        color: var(--td-text-color-primary);
      }

      .progress-values {
        display: flex;
        align-items: center;
        gap: 4px;
        font-size: 16px;

        .current-value {
          font-weight: 600;
          color: var(--td-text-color-primary);

          &.text-danger {
            color: var(--td-error-color);
          }
        }

        .limit-value {
          color: var(--td-text-color-secondary);
        }
      }
    }

    .progress-bar-container {
      margin-bottom: 8px;

      .progress-details {
        display: flex;
        justify-content: space-between;
        align-items: center;
        margin-top: 4px;
        font-size: 12px;

        .progress-percentage {
          color: var(--td-text-color-primary);
          font-weight: 500;
        }

        .warning-threshold {
          color: var(--td-text-color-secondary);
        }
      }
    }

    .no-limit-hint {
      display: flex;
      align-items: center;
      gap: 4px;
      color: var(--td-text-color-secondary);
      font-size: 14px;
      padding: 8px 0;

      .t-icon {
        font-size: 16px;
      }
    }

    .warning-message {
      margin-top: 8px;

      :deep(.t-alert) {
        .t-alert__message {
          font-size: 12px;
          line-height: 1.4;
        }
      }
    }
  }

  // 进度条全局样式调整
  :deep(.t-progress) {
    .t-progress__bar {
      transition: all 0.3s ease;
    }

    &.t-is-error .t-progress__bar {
      background: var(--td-error-color);
    }

    &.t-is-warning .t-progress__bar {
      background: var(--td-warning-color);
    }

    &.t-is-success .t-progress__bar {
      background: var(--td-success-color);
    }
  }
}

// 响应式设计
@media (max-width: 768px) {
  .usage-progress {
    .progress-full {
      .progress-header {
        flex-direction: column;
        align-items: flex-start;
        gap: 4px;

        .progress-values {
          font-size: 14px;
        }
      }

      .progress-bar-container {
        .progress-details {
          flex-direction: column;
          align-items: flex-start;
          gap: 2px;
        }
      }
    }
  }
}
</style>
