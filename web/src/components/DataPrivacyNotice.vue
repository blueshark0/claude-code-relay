<template>
  <div class="data-privacy-notice">
    <t-dialog
      v-model:visible="showDialog"
      header="数据传输声明"
      width="600px"
      :footer="false"
      :close-on-overlay-click="false"
    >
      <div class="privacy-content">
        <t-alert theme="warning" title="重要声明">
          使用本服务时，您的对话内容将被传输至第三方AI服务提供商进行处理。
        </t-alert>
        
        <div class="privacy-details">
          <h4>数据传输详情：</h4>
          <ul>
            <li><strong>服务提供商：</strong>Anthropic (Claude AI)</li>
            <li><strong>传输地址：</strong>api.anthropic.com</li>
            <li><strong>传输内容：</strong>您的对话消息、系统提示等</li>
            <li><strong>用途：</strong>生成AI回复</li>
          </ul>
          
          <h4>安全措施：</h4>
          <ul>
            <li>所有数据传输均通过HTTPS加密</li>
            <li>用户身份信息已匿名化处理</li>
            <li>不会永久存储您的对话内容</li>
          </ul>
          
          <h4>您的权利：</h4>
          <ul>
            <li>您可以随时停止使用本服务</li>
            <li>对话结束后数据不会被保留</li>
            <li>如有疑问，请联系管理员</li>
          </ul>
        </div>
        
        <div class="privacy-actions">
          <t-checkbox v-model="agreed">
            我已阅读并同意上述数据传输条款
          </t-checkbox>
          <div class="action-buttons">
            <t-button @click="disagree" variant="outline">
              不同意
            </t-button>
            <t-button @click="agree" theme="primary" :disabled="!agreed">
              同意并继续
            </t-button>
          </div>
        </div>
      </div>
    </t-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue';

const showDialog = ref(false);
const agreed = ref(false);

const emit = defineEmits<{
  agreed: [];
  disagreed: [];
}>();

onMounted(() => {
  // 检查用户是否已经同意过
  const hasAgreed = localStorage.getItem('dataPrivacyAgreed');
  if (!hasAgreed) {
    showDialog.value = true;
  } else {
    emit('agreed');
  }
});

const agree = () => {
  localStorage.setItem('dataPrivacyAgreed', 'true');
  localStorage.setItem('dataPrivacyAgreedTime', new Date().toISOString());
  showDialog.value = false;
  emit('agreed');
};

const disagree = () => {
  showDialog.value = false;
  emit('disagreed');
  // 可以重定向到其他页面或显示替代说明
};
</script>

<style scoped>
.privacy-content {
  padding: 16px 0;
}

.privacy-details {
  margin: 20px 0;
}

.privacy-details h4 {
  color: var(--td-text-color-primary);
  margin: 16px 0 8px 0;
}

.privacy-details ul {
  margin: 8px 0;
  padding-left: 20px;
}

.privacy-details li {
  margin: 4px 0;
  line-height: 1.5;
}

.privacy-actions {
  margin-top: 24px;
  padding-top: 20px;
  border-top: 1px solid var(--td-border-level-1-color);
}

.action-buttons {
  margin-top: 16px;
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>