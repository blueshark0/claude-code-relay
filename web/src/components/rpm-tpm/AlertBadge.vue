<template>
  <div class="alert-badge">
    <!-- 简单徽章模式 -->
    <t-badge v-if="simple" :count="alertLevel" :color="badgeColor" :dot="dot" :max-count="3">
      <slot />
    </t-badge>

    <!-- 详细状态模式 -->
    <div v-else class="alert-status">
      <div class="alert-indicators">
        <!-- RPM 状态指示器 -->
        <div v-if="rpmStatus !== 'normal'" class="status-indicator" :class="`status-${rpmStatus}`">
          <t-icon :name="getStatusIcon(rpmStatus)" />
          <span class="status-text">RPM {{ getStatusText(rpmStatus) }}</span>
        </div>

        <!-- TPM 状态指示器 -->
        <div v-if="tpmStatus !== 'normal'" class="status-indicator" :class="`status-${tpmStatus}`">
          <t-icon :name="getStatusIcon(tpmStatus)" />
          <span class="status-text">TPM {{ getStatusText(tpmStatus) }}</span>
        </div>

        <!-- 限流状态指示器 -->
        <div v-if="isRateLimited" class="status-indicator status-limited">
          <t-icon name="time" />
          <span class="status-text">限流中</span>
        </div>
      </div>

      <!-- 详细信息弹窗触发器 -->
      <t-popup
        v-if="showDetail && (alertLevel > 0 || isRateLimited)"
        :content="detailContent"
        placement="top"
        trigger="hover"
      >
        <div class="detail-trigger">
          <t-icon name="info-circle" />
        </div>
      </t-popup>
    </div>
  </div>
</template>
<script setup lang="ts">
import { computed } from 'vue';

interface Props {
  // RPM 相关
  rpmCurrent?: number;
  rpmLimit?: number;
  rpmWarningThreshold?: number;
  isRpmLimited?: boolean;

  // TPM 相关
  tpmCurrent?: number;
  tpmLimit?: number;
  tpmWarningThreshold?: number;
  isTpmLimited?: boolean;

  // 限流状态
  isRateLimited?: boolean;
  rateLimitEndTime?: string;

  // 显示模式
  simple?: boolean;
  dot?: boolean;
  showDetail?: boolean;
}

const props = withDefaults(defineProps<Props>(), {
  rpmCurrent: 0,
  rpmLimit: 0,
  rpmWarningThreshold: 0,
  isRpmLimited: false,
  tpmCurrent: 0,
  tpmLimit: 0,
  tpmWarningThreshold: 0,
  isTpmLimited: false,
  isRateLimited: false,
  rateLimitEndTime: '',
  simple: false,
  dot: false,
  showDetail: true,
});

// 计算 RPM 状态
const rpmStatus = computed(() => {
  if (props.isRpmLimited || props.isRateLimited) return 'error';
  if (props.rpmLimit > 0) {
    const percentage = (props.rpmCurrent / props.rpmLimit) * 100;
    if (percentage >= 95) return 'error';
    if (percentage >= 80) return 'warning';
  }
  if (props.rpmWarningThreshold > 0 && props.rpmCurrent >= props.rpmWarningThreshold) {
    return 'warning';
  }
  return 'normal';
});

// 计算 TPM 状态
const tpmStatus = computed(() => {
  if (props.isTpmLimited || props.isRateLimited) return 'error';
  if (props.tpmLimit > 0) {
    const percentage = (props.tpmCurrent / props.tpmLimit) * 100;
    if (percentage >= 95) return 'error';
    if (percentage >= 80) return 'warning';
  }
  if (props.tpmWarningThreshold > 0 && props.tpmCurrent >= props.tpmWarningThreshold) {
    return 'warning';
  }
  return 'normal';
});

// 计算总体警告级别
const alertLevel = computed(() => {
  let level = 0;
  if (rpmStatus.value === 'warning' || tpmStatus.value === 'warning') level = Math.max(level, 1);
  if (rpmStatus.value === 'error' || tpmStatus.value === 'error' || props.isRateLimited) level = Math.max(level, 2);
  return level;
});

// 徽章颜色
const badgeColor = computed(() => {
  if (alertLevel.value >= 2) return '#f5222d';
  if (alertLevel.value >= 1) return '#faad14';
  return '#52c41a';
});

// 获取状态图标
const getStatusIcon = (status: string): string => {
  switch (status) {
    case 'error':
      return 'close-circle-filled';
    case 'warning':
      return 'error-circle-filled';
    default:
      return 'check-circle-filled';
  }
};

// 获取状态文本
const getStatusText = (status: string): string => {
  switch (status) {
    case 'error':
      return '异常';
    case 'warning':
      return '告警';
    default:
      return '正常';
  }
};

// 详细信息内容
const detailContent = computed(() => {
  const details: string[] = [];

  if (rpmStatus.value !== 'normal') {
    const rpmPercentage = props.rpmLimit > 0 ? (props.rpmCurrent / props.rpmLimit) * 100 : 0;
    details.push(
      `RPM: ${props.rpmCurrent}${props.rpmLimit > 0 ? `/${props.rpmLimit}` : ''} (${rpmPercentage.toFixed(1)}%)`,
    );
  }

  if (tpmStatus.value !== 'normal') {
    const tpmPercentage = props.tpmLimit > 0 ? (props.tpmCurrent / props.tpmLimit) * 100 : 0;
    details.push(
      `TPM: ${props.tpmCurrent}${props.tpmLimit > 0 ? `/${props.tpmLimit}` : ''} (${tpmPercentage.toFixed(1)}%)`,
    );
  }

  if (props.isRateLimited && props.rateLimitEndTime) {
    const endTime = new Date(props.rateLimitEndTime);
    details.push(`限流至: ${endTime.toLocaleString('zh-CN')}`);
  }

  return details.join('\n') || '状态正常';
});
</script>
<style lang="less" scoped>
.alert-badge {
  display: inline-flex;
  align-items: center;

  .alert-status {
    display: flex;
    align-items: center;
    gap: 8px;

    .alert-indicators {
      display: flex;
      align-items: center;
      gap: 4px;

      .status-indicator {
        display: flex;
        align-items: center;
        gap: 2px;
        padding: 2px 6px;
        border-radius: 4px;
        font-size: 11px;
        font-weight: 500;
        line-height: 1;

        .t-icon {
          font-size: 12px;
        }

        &.status-warning {
          background: var(--td-warning-color-1);
          color: var(--td-warning-color);
          border: 1px solid var(--td-warning-color-3);
        }

        &.status-error {
          background: var(--td-error-color-1);
          color: var(--td-error-color);
          border: 1px solid var(--td-error-color-3);
        }

        &.status-limited {
          background: var(--td-error-color-1);
          color: var(--td-error-color);
          border: 1px solid var(--td-error-color-3);
        }

        .status-text {
          white-space: nowrap;
        }
      }
    }

    .detail-trigger {
      display: flex;
      align-items: center;
      justify-content: center;
      width: 16px;
      height: 16px;
      border-radius: 50%;
      background: var(--td-bg-color-component);
      border: 1px solid var(--td-border-level-1-color);
      cursor: pointer;
      transition: all 0.2s ease;

      .t-icon {
        font-size: 10px;
        color: var(--td-text-color-secondary);
      }

      &:hover {
        border-color: var(--td-brand-color);
        background: var(--td-brand-color-1);

        .t-icon {
          color: var(--td-brand-color);
        }
      }
    }
  }

  // 徽章模式的全局样式调整
  :deep(.t-badge) {
    .t-badge__count {
      font-size: 10px;
      min-width: 16px;
      height: 16px;
      line-height: 16px;
      padding: 0 4px;
    }

    .t-badge__dot {
      width: 6px;
      height: 6px;
    }
  }
}

// 响应式设计
@media (max-width: 768px) {
  .alert-badge {
    .alert-status {
      .alert-indicators {
        .status-indicator {
          padding: 1px 4px;
          font-size: 10px;

          .status-text {
            display: none; // 移动端隐藏文本，只显示图标
          }
        }
      }
    }
  }
}
</style>
