<template>
  <div class="login-wrapper">
    <login-header />

    <div class="login-container">
      <div class="title-container">
        <h1 class="title margin-no">{{ t('pages.login.loginTitle') }}</h1>
        <h1 class="title">Claude Code Relay</h1>
        <div v-if="registrationEnabled" class="sub-title">
          <p class="tip">{{ type === 'register' ? t('pages.login.existAccount') : t('pages.login.noAccount') }}</p>
          <p class="tip" @click="switchType(type === 'register' ? 'login' : 'register')">
            {{ type === 'register' ? t('pages.login.signIn') : t('pages.login.createAccount') }}
          </p>
        </div>
      </div>

      <login v-if="type === 'login' || !registrationEnabled" />
      <register v-else-if="registrationEnabled" @register-success="switchType('login')" />
      <tdesign-setting />
    </div>
  </div>
</template>
<script setup lang="ts">
import { onMounted, ref } from 'vue';

import { getSystemStatus } from '@/api/system';
import TdesignSetting from '@/layouts/setting.vue';
import { t } from '@/locales';

import LoginHeader from './components/Header.vue';
import Login from './components/Login.vue';
import Register from './components/Register.vue';

defineOptions({
  name: 'LoginIndex',
});

const type = ref('login');
const registrationEnabled = ref(true); // 默认启用注册，加载后更新

const switchType = (val: string) => {
  type.value = val;
};

// 获取系统配置
const fetchSystemConfig = async () => {
  try {
    const response = await getSystemStatus();
    registrationEnabled.value = response.registration_enabled;

    // 如果注册被禁用且当前在注册页面，切换到登录页面
    if (!response.registration_enabled && type.value === 'register') {
      type.value = 'login';
    }
  } catch (error) {
    console.error('获取系统配置失败:', error);
    // 发生错误时默认启用注册功能
    registrationEnabled.value = true;
  }
};

onMounted(() => {
  fetchSystemConfig();
});
</script>
<style lang="less" scoped>
@import './index.less';
</style>
