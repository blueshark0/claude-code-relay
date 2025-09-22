<template>
  <div class="limits-form">
    <t-form ref="formRef" :model="formData" :rules="formRules" label-align="top" @submit="handleSubmit">
      <div class="form-grid">
        <!-- RPM 限制设置 -->
        <div class="form-section">
          <div class="section-title">
            <t-icon name="speed" />
            <span>RPM (每分钟请求数) 限制</span>
          </div>

          <t-form-item label="RPM 限制值" name="rpm_limit">
            <t-input-number
              v-model="formData.rpm_limit"
              :min="0"
              :max="999999"
              placeholder="0表示无限制"
              style="width: 100%"
            />
            <template #tips>
              <div class="form-tips">设置每分钟最大请求数，0表示无限制。建议根据账户类型合理设置。</div>
            </template>
          </t-form-item>

          <t-form-item label="RPM 告警阈值" name="rpm_warning_threshold">
            <t-input-number
              v-model="formData.rpm_warning_threshold"
              :min="0"
              :max="formData.rpm_limit || 999999"
              :disabled="!formData.rpm_limit"
              placeholder="0表示不告警"
              style="width: 100%"
            />
            <template #tips>
              <div class="form-tips">当RPM达到此阈值时发出告警提醒，通常设置为限制值的80-90%。</div>
            </template>
          </t-form-item>
        </div>

        <!-- TPM 限制设置 -->
        <div class="form-section">
          <div class="section-title">
            <t-icon name="data-base" />
            <span>TPM (每分钟Token数) 限制</span>
          </div>

          <t-form-item label="TPM 限制值" name="tpm_limit">
            <t-input-number
              v-model="formData.tpm_limit"
              :min="0"
              :max="9999999"
              placeholder="0表示无限制"
              style="width: 100%"
            />
            <template #tips>
              <div class="form-tips">设置每分钟最大Token数，0表示无限制。Token包括输入和输出Token总和。</div>
            </template>
          </t-form-item>

          <t-form-item label="TPM 告警阈值" name="tpm_warning_threshold">
            <t-input-number
              v-model="formData.tpm_warning_threshold"
              :min="0"
              :max="formData.tpm_limit || 9999999"
              :disabled="!formData.tpm_limit"
              placeholder="0表示不告警"
              style="width: 100%"
            />
            <template #tips>
              <div class="form-tips">当TPM达到此阈值时发出告警提醒，通常设置为限制值的80-90%。</div>
            </template>
          </t-form-item>
        </div>
      </div>

      <!-- 快捷设置 -->
      <div class="quick-settings">
        <div class="quick-title">快捷设置模板</div>
        <div class="quick-buttons">
          <t-button
            v-for="template in quickTemplates"
            :key="template.name"
            variant="outline"
            size="small"
            @click="applyTemplate(template)"
          >
            {{ template.name }}
          </t-button>
        </div>
      </div>

      <!-- 当前状态对比 -->
      <div v-if="currentStats" class="current-comparison">
        <div class="comparison-title">当前状态对比</div>
        <div class="comparison-content">
          <div class="comparison-item">
            <div class="comparison-label">当前 RPM:</div>
            <div class="comparison-value">
              {{ formatRpmTpmValue(currentStats.current_rpm) }}
              <span class="comparison-percentage" :class="getRpmComparisonClass()"> ({{ getRpmUsageText() }}) </span>
            </div>
          </div>
          <div class="comparison-item">
            <div class="comparison-label">当前 TPM:</div>
            <div class="comparison-value">
              {{ formatRpmTpmValue(currentStats.current_tpm) }}
              <span class="comparison-percentage" :class="getTpmComparisonClass()"> ({{ getTpmUsageText() }}) </span>
            </div>
          </div>
        </div>
      </div>

      <!-- 表单操作按钮 -->
      <div class="form-actions">
        <t-button @click="handleReset">重置</t-button>
        <t-button theme="primary" type="submit" :loading="loading">
          {{ submitText }}
        </t-button>
      </div>
    </t-form>
  </div>
</template>
<script setup lang="ts">
import type { FormInstanceFunctions, FormRule } from 'tdesign-vue-next';
import { reactive, ref, watch } from 'vue';

import type { RpmTpmStats, UpdateRpmTpmLimitsParams } from '@/api/rpm-tpm';
import { formatRpmTpmValue } from '@/api/rpm-tpm';

interface Props {
  initialData?: Partial<UpdateRpmTpmLimitsParams>;
  currentStats?: RpmTpmStats;
  loading?: boolean;
  submitText?: string;
}

interface Emits {
  (event: 'submit', data: UpdateRpmTpmLimitsParams): void;
  (event: 'cancel'): void;
}

const props = withDefaults(defineProps<Props>(), {
  initialData: () => ({}),
  loading: false,
  submitText: '保存设置',
});

const emit = defineEmits<Emits>();

// 表单引用
const formRef = ref<FormInstanceFunctions>();

// 表单数据
const formData = reactive<UpdateRpmTpmLimitsParams>({
  rpm_limit: 0,
  tpm_limit: 0,
  rpm_warning_threshold: 0,
  tpm_warning_threshold: 0,
});

// 表单验证规则
const formRules: Record<string, FormRule[]> = {
  rpm_limit: [{ validator: (val) => val >= 0, message: 'RPM限制值不能为负数' }],
  tpm_limit: [{ validator: (val) => val >= 0, message: 'TPM限制值不能为负数' }],
  rpm_warning_threshold: [
    {
      validator: (val) => {
        if (val > 0 && formData.rpm_limit > 0) {
          return val <= formData.rpm_limit;
        }
        return true;
      },
      message: 'RPM告警阈值不能超过RPM限制值',
    },
  ],
  tpm_warning_threshold: [
    {
      validator: (val) => {
        if (val > 0 && formData.tpm_limit > 0) {
          return val <= formData.tpm_limit;
        }
        return true;
      },
      message: 'TPM告警阈值不能超过TPM限制值',
    },
  ],
};

// 快捷设置模板
const quickTemplates = [
  {
    name: '无限制',
    rpm_limit: 0,
    tpm_limit: 0,
    rpm_warning_threshold: 0,
    tpm_warning_threshold: 0,
  },
  {
    name: '轻量使用',
    rpm_limit: 60,
    tpm_limit: 50000,
    rpm_warning_threshold: 50,
    tpm_warning_threshold: 40000,
  },
  {
    name: '中等使用',
    rpm_limit: 200,
    tpm_limit: 150000,
    rpm_warning_threshold: 160,
    tpm_warning_threshold: 120000,
  },
  {
    name: '重度使用',
    rpm_limit: 500,
    tpm_limit: 400000,
    rpm_warning_threshold: 400,
    tpm_warning_threshold: 320000,
  },
];

// 监听 initialData 变化
watch(
  () => props.initialData,
  (newData) => {
    Object.assign(formData, {
      rpm_limit: 0,
      tpm_limit: 0,
      rpm_warning_threshold: 0,
      tpm_warning_threshold: 0,
      ...newData,
    });
  },
  { immediate: true, deep: true },
);

// RPM使用率文本和样式
const getRpmUsageText = () => {
  if (!formData.rpm_limit || formData.rpm_limit === 0) return '无限制';
  if (!props.currentStats) return '';
  const percentage = (props.currentStats.current_rpm / formData.rpm_limit) * 100;
  return `${percentage.toFixed(1)}%`;
};

const getRpmComparisonClass = () => {
  if (!formData.rpm_limit || formData.rpm_limit === 0) return 'normal';
  if (!props.currentStats) return 'normal';
  const percentage = (props.currentStats.current_rpm / formData.rpm_limit) * 100;
  if (percentage >= 95) return 'danger';
  if (percentage >= 80) return 'warning';
  return 'safe';
};

// TPM使用率文本和样式
const getTpmUsageText = () => {
  if (!formData.tpm_limit || formData.tpm_limit === 0) return '无限制';
  if (!props.currentStats) return '';
  const percentage = (props.currentStats.current_tpm / formData.tpm_limit) * 100;
  return `${percentage.toFixed(1)}%`;
};

const getTpmComparisonClass = () => {
  if (!formData.tpm_limit || formData.tpm_limit === 0) return 'normal';
  if (!props.currentStats) return 'normal';
  const percentage = (props.currentStats.current_tpm / formData.tpm_limit) * 100;
  if (percentage >= 95) return 'danger';
  if (percentage >= 80) return 'warning';
  return 'safe';
};

// 应用模板
const applyTemplate = (template: (typeof quickTemplates)[0]) => {
  Object.assign(formData, template);
};

// 处理表单提交
const handleSubmit = async () => {
  const valid = await formRef.value?.validate();
  if (valid !== true) return;

  emit('submit', { ...formData });
};

// 重置表单
const handleReset = () => {
  Object.assign(formData, {
    rpm_limit: 0,
    tpm_limit: 0,
    rpm_warning_threshold: 0,
    tpm_warning_threshold: 0,
    ...props.initialData,
  });
};
</script>
<style lang="less" scoped>
.limits-form {
  .form-grid {
    display: grid;
    grid-template-columns: 1fr 1fr;
    gap: 24px;
    margin-bottom: 24px;

    .form-section {
      .section-title {
        display: flex;
        align-items: center;
        gap: 8px;
        font-size: 16px;
        font-weight: 600;
        color: var(--td-text-color-primary);
        margin-bottom: 16px;
        padding-bottom: 8px;
        border-bottom: 1px solid var(--td-border-level-1-color);

        .t-icon {
          color: var(--td-brand-color);
        }
      }

      :deep(.t-form-item) {
        margin-bottom: 20px;

        .t-form__label {
          font-weight: 500;
          color: var(--td-text-color-primary);
        }
      }
    }
  }

  .form-tips {
    font-size: 12px;
    color: var(--td-text-color-secondary);
    line-height: 1.4;
    margin-top: 4px;
  }

  .quick-settings {
    margin-bottom: 24px;
    padding: 16px;
    background: var(--td-bg-color-container-hover);
    border-radius: 8px;

    .quick-title {
      font-size: 14px;
      font-weight: 500;
      color: var(--td-text-color-primary);
      margin-bottom: 12px;
    }

    .quick-buttons {
      display: flex;
      gap: 8px;
      flex-wrap: wrap;
    }
  }

  .current-comparison {
    margin-bottom: 24px;
    padding: 16px;
    border: 1px solid var(--td-border-level-1-color);
    border-radius: 8px;

    .comparison-title {
      font-size: 14px;
      font-weight: 500;
      color: var(--td-text-color-primary);
      margin-bottom: 12px;
    }

    .comparison-content {
      display: grid;
      grid-template-columns: 1fr 1fr;
      gap: 16px;

      .comparison-item {
        display: flex;
        justify-content: space-between;
        align-items: center;

        .comparison-label {
          font-size: 13px;
          color: var(--td-text-color-secondary);
        }

        .comparison-value {
          font-weight: 500;
          color: var(--td-text-color-primary);

          .comparison-percentage {
            font-size: 12px;
            margin-left: 4px;

            &.safe {
              color: var(--td-success-color);
            }

            &.warning {
              color: var(--td-warning-color);
            }

            &.danger {
              color: var(--td-error-color);
            }

            &.normal {
              color: var(--td-text-color-secondary);
            }
          }
        }
      }
    }
  }

  .form-actions {
    display: flex;
    justify-content: flex-end;
    gap: 12px;
    padding-top: 20px;
    border-top: 1px solid var(--td-border-level-1-color);
  }
}

// 响应式设计
@media (max-width: 768px) {
  .limits-form {
    .form-grid {
      grid-template-columns: 1fr;
      gap: 20px;
    }

    .quick-settings {
      .quick-buttons {
        grid-template-columns: repeat(2, 1fr);
      }
    }

    .current-comparison {
      .comparison-content {
        grid-template-columns: 1fr;
        gap: 12px;

        .comparison-item {
          flex-direction: column;
          align-items: flex-start;
          gap: 4px;
        }
      }
    }

    .form-actions {
      flex-direction: column-reverse;

      .t-button {
        width: 100%;
      }
    }
  }
}
</style>
