<template>
  <div class="history-chart">
    <div class="chart-header">
      <div class="chart-title">
        <t-icon name="chart-line" />
        <span>{{ title }}</span>
      </div>
      <div class="chart-controls">
        <!-- 时间范围选择 -->
        <t-select v-model="timeRange" size="small" style="width: 120px" @change="handleTimeRangeChange">
          <t-option value="1h" label="最近1小时" />
          <t-option value="6h" label="最近6小时" />
          <t-option value="24h" label="最近24小时" />
          <t-option value="7d" label="最近7天" />
          <t-option value="custom" label="自定义" />
        </t-select>

        <!-- 自定义时间范围 -->
        <t-date-range-picker
          v-if="timeRange === 'custom'"
          v-model="customDateRange"
          size="small"
          enable-time-picker
          format="YYYY-MM-DD HH:mm:ss"
          @change="handleCustomRangeChange"
        />

        <!-- 刷新按钮 -->
        <t-button variant="outline" size="small" :loading="loading" @click="handleRefresh">
          <template #icon>
            <t-icon name="refresh" />
          </template>
        </t-button>
      </div>
    </div>

    <!-- 图表容器 -->
    <div class="chart-container">
      <div v-if="loading" class="chart-loading">
        <t-loading size="large" text="加载中..." />
      </div>
      <div v-show="!loading" ref="chartRef" class="chart-content" :class="{ 'chart-empty': !hasData }"></div>
      <div v-if="!loading && !hasData" class="chart-empty-state">
        <t-icon name="chart-line" size="48px" />
        <div class="empty-text">暂无历史数据</div>
        <div class="empty-hint">请稍后重试或调整时间范围</div>
      </div>
    </div>

    <!-- 数据概览 -->
    <div v-if="!loading && hasData && summary" class="chart-summary">
      <div class="summary-item">
        <span class="summary-label">平均 RPM:</span>
        <span class="summary-value">{{ formatRpmTpmValue(summary.avg_rpm) }}</span>
      </div>
      <div class="summary-item">
        <span class="summary-label">峰值 RPM:</span>
        <span class="summary-value">{{ formatRpmTpmValue(summary.max_rpm) }}</span>
      </div>
      <div class="summary-item">
        <span class="summary-label">平均 TPM:</span>
        <span class="summary-value">{{ formatRpmTpmValue(summary.avg_tpm) }}</span>
      </div>
      <div class="summary-item">
        <span class="summary-label">峰值 TPM:</span>
        <span class="summary-value">{{ formatRpmTpmValue(summary.max_tpm) }}</span>
      </div>
    </div>
  </div>
</template>
<script setup lang="ts">
import { LineChart } from 'echarts/charts';
import {
  DataZoomComponent,
  GridComponent,
  LegendComponent,
  MarkLineComponent,
  TooltipComponent,
} from 'echarts/components';
import * as echarts from 'echarts/core';
import { CanvasRenderer } from 'echarts/renderers';
import { computed, nextTick, onMounted, onUnmounted, ref, watch } from 'vue';

import type { RpmTpmHistoryItem } from '@/api/rpm-tpm';
import { formatRpmTpmValue } from '@/api/rpm-tpm';

const props = withDefaults(defineProps<Props>(), {
  title: 'RPM/TPM 历史趋势',
  loading: false,
  rpmLimit: 0,
  tpmLimit: 0,
  rpmWarningThreshold: 0,
  tpmWarningThreshold: 0,
});

const emit = defineEmits<Emits>();

// 注册 ECharts 组件
echarts.use([
  LineChart,
  GridComponent,
  TooltipComponent,
  LegendComponent,
  DataZoomComponent,
  MarkLineComponent,
  CanvasRenderer,
]);

interface Props {
  title?: string;
  data?: RpmTpmHistoryItem[];
  loading?: boolean;
  rpmLimit?: number;
  tpmLimit?: number;
  rpmWarningThreshold?: number;
  tpmWarningThreshold?: number;
}

interface Emits {
  (event: 'time-range-change', params: { start_time: string; end_time: string }): void;
  (event: 'refresh'): void;
}

// 图表相关
const chartRef = ref<HTMLElement>();
let chartInstance: echarts.ECharts | null = null;

// 时间范围
const timeRange = ref('24h');
const customDateRange = ref<[string, string]>(['', '']);

// 计算属性
const hasData = computed(() => props.data && props.data.length > 0);

const summary = computed(() => {
  if (!hasData.value) return null;

  const rpms = props.data.map((item) => item.rpm);
  const tpms = props.data.map((item) => item.tpm);

  return {
    avg_rpm: Math.round(rpms.reduce((sum, val) => sum + val, 0) / rpms.length),
    max_rpm: Math.max(...rpms),
    avg_tpm: Math.round(tpms.reduce((sum, val) => sum + val, 0) / tpms.length),
    max_tpm: Math.max(...tpms),
  };
});

// 初始化图表
const initChart = () => {
  if (!chartRef.value) return;

  chartInstance = echarts.init(chartRef.value);

  // 监听窗口大小变化
  const resizeHandler = () => {
    if (chartInstance) {
      chartInstance.resize();
    }
  };

  window.addEventListener('resize', resizeHandler);

  // 在组件卸载时清理
  return () => {
    window.removeEventListener('resize', resizeHandler);
  };
};

// 更新图表
const updateChart = () => {
  if (!chartInstance || !hasData.value) return;

  const xAxisData = props.data.map((item) =>
    new Date(item.minute_timestamp).toLocaleString('zh-CN', {
      month: '2-digit',
      day: '2-digit',
      hour: '2-digit',
      minute: '2-digit',
    }),
  );

  const rpmData = props.data.map((item) => item.rpm);
  const tpmData = props.data.map((item) => item.tpm);

  // 构建标记线数据
  const markLines: any[] = [];

  if (props.rpmLimit > 0) {
    markLines.push({
      name: 'RPM限制',
      yAxis: props.rpmLimit,
      lineStyle: { color: '#f5222d', type: 'dashed' },
      label: { formatter: `RPM限制: ${formatRpmTpmValue(props.rpmLimit)}` },
    });
  }

  if (props.rpmWarningThreshold > 0) {
    markLines.push({
      name: 'RPM告警',
      yAxis: props.rpmWarningThreshold,
      lineStyle: { color: '#faad14', type: 'dashed' },
      label: { formatter: `RPM告警: ${formatRpmTpmValue(props.rpmWarningThreshold)}` },
    });
  }

  const option = {
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'cross',
        crossStyle: {
          color: '#999',
        },
      },
      formatter: (params: any) => {
        let result = `${params[0].axisValue}<br/>`;
        params.forEach((param: any) => {
          const value = formatRpmTpmValue(param.value);
          result += `${param.marker} ${param.seriesName}: ${value}<br/>`;
        });
        return result;
      },
    },
    legend: {
      data: ['RPM', 'TPM'],
      top: 10,
    },
    grid: {
      left: '3%',
      right: '4%',
      bottom: '10%',
      top: '15%',
      containLabel: true,
    },
    xAxis: {
      type: 'category',
      data: xAxisData,
      axisPointer: {
        type: 'shadow',
      },
    },
    yAxis: [
      {
        type: 'value',
        name: 'RPM',
        position: 'left',
        axisLabel: {
          formatter: (value: number) => formatRpmTpmValue(value),
        },
      },
      {
        type: 'value',
        name: 'TPM',
        position: 'right',
        axisLabel: {
          formatter: (value: number) => formatRpmTpmValue(value),
        },
      },
    ],
    dataZoom: [
      {
        type: 'slider',
        show: true,
        xAxisIndex: [0],
        start: 0,
        end: 100,
      },
    ],
    series: [
      {
        name: 'RPM',
        type: 'line',
        yAxisIndex: 0,
        data: rpmData,
        smooth: true,
        symbol: 'circle',
        symbolSize: 4,
        lineStyle: {
          color: '#1890ff',
          width: 2,
        },
        itemStyle: {
          color: '#1890ff',
        },
        markLine: markLines.length > 0 ? { data: markLines } : undefined,
      },
      {
        name: 'TPM',
        type: 'line',
        yAxisIndex: 1,
        data: tpmData,
        smooth: true,
        symbol: 'circle',
        symbolSize: 4,
        lineStyle: {
          color: '#52c41a',
          width: 2,
        },
        itemStyle: {
          color: '#52c41a',
        },
      },
    ],
  };

  chartInstance.setOption(option, true);
};

// 处理时间范围变化
const handleTimeRangeChange = (value: string) => {
  if (value === 'custom') return;

  const now = new Date();
  let startTime: Date;

  switch (value) {
    case '1h':
      startTime = new Date(now.getTime() - 60 * 60 * 1000);
      break;
    case '6h':
      startTime = new Date(now.getTime() - 6 * 60 * 60 * 1000);
      break;
    case '24h':
      startTime = new Date(now.getTime() - 24 * 60 * 60 * 1000);
      break;
    case '7d':
      startTime = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
      break;
    default:
      startTime = new Date(now.getTime() - 24 * 60 * 60 * 1000);
  }

  emit('time-range-change', {
    start_time: startTime.toISOString(),
    end_time: now.toISOString(),
  });
};

// 处理自定义时间范围变化
const handleCustomRangeChange = (value: [string, string]) => {
  if (value && value[0] && value[1]) {
    emit('time-range-change', {
      start_time: new Date(value[0]).toISOString(),
      end_time: new Date(value[1]).toISOString(),
    });
  }
};

// 处理刷新
const handleRefresh = () => {
  emit('refresh');
};

// 监听数据变化
watch(
  () => [props.data, props.loading],
  () => {
    if (!props.loading) {
      nextTick(() => {
        updateChart();
      });
    }
  },
  { deep: true },
);

// 监听限制值变化
watch(
  () => [props.rpmLimit, props.tpmLimit, props.rpmWarningThreshold, props.tpmWarningThreshold],
  () => {
    if (!props.loading && hasData.value) {
      nextTick(() => {
        updateChart();
      });
    }
  },
);

// 生命周期
onMounted(() => {
  const cleanup = initChart();

  onUnmounted(() => {
    cleanup?.();
    if (chartInstance) {
      chartInstance.dispose();
      chartInstance = null;
    }
  });

  // 初始加载数据
  handleTimeRangeChange(timeRange.value);
});
</script>
<style lang="less" scoped>
.history-chart {
  height: 100%;
  display: flex;
  flex-direction: column;

  .chart-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    padding-bottom: 12px;
    border-bottom: 1px solid var(--td-border-level-1-color);

    .chart-title {
      display: flex;
      align-items: center;
      gap: 8px;
      font-size: 16px;
      font-weight: 600;
      color: var(--td-text-color-primary);

      .t-icon {
        color: var(--td-brand-color);
      }
    }

    .chart-controls {
      display: flex;
      align-items: center;
      gap: 12px;
    }
  }

  .chart-container {
    flex: 1;
    position: relative;
    min-height: 400px;

    .chart-loading {
      position: absolute;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
      z-index: 10;
    }

    .chart-content {
      width: 100%;
      height: 100%;

      &.chart-empty {
        display: none;
      }
    }

    .chart-empty-state {
      position: absolute;
      top: 50%;
      left: 50%;
      transform: translate(-50%, -50%);
      text-align: center;
      color: var(--td-text-color-placeholder);

      .t-icon {
        margin-bottom: 16px;
        opacity: 0.5;
      }

      .empty-text {
        font-size: 16px;
        margin-bottom: 8px;
      }

      .empty-hint {
        font-size: 14px;
        opacity: 0.8;
      }
    }
  }

  .chart-summary {
    display: grid;
    grid-template-columns: repeat(4, 1fr);
    gap: 16px;
    margin-top: 16px;
    padding: 16px;
    background: var(--td-bg-color-container-hover);
    border-radius: 8px;

    .summary-item {
      display: flex;
      justify-content: space-between;
      align-items: center;
      font-size: 13px;

      .summary-label {
        color: var(--td-text-color-secondary);
      }

      .summary-value {
        color: var(--td-text-color-primary);
        font-weight: 500;
      }
    }
  }
}

// 响应式设计
@media (max-width: 768px) {
  .history-chart {
    .chart-header {
      flex-direction: column;
      align-items: flex-start;
      gap: 12px;

      .chart-controls {
        width: 100%;
        justify-content: space-between;
      }
    }

    .chart-container {
      min-height: 300px;
    }

    .chart-summary {
      grid-template-columns: repeat(2, 1fr);
      gap: 12px;

      .summary-item {
        font-size: 12px;
      }
    }
  }
}
</style>
