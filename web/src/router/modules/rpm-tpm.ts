import { ChartLineIcon } from 'tdesign-icons-vue-next';
import { shallowRef } from 'vue';

import Layout from '@/layouts/index.vue';

export default [
  {
    path: '/rpm-tpm',
    component: Layout,
    redirect: '/rpm-tpm/dashboard',
    name: 'RpmTpm',
    meta: {
      title: 'RPM/TPM 监控',
      icon: shallowRef(ChartLineIcon),
      orderNo: 5,
    },
    children: [
      {
        path: 'dashboard',
        name: 'RpmTpmDashboard',
        component: () => import('@/pages/rpm-tpm/dashboard/index.vue'),
        meta: {
          title: '监控仪表盘',
        },
      },
      {
        path: 'api-keys',
        name: 'RpmTpmApiKeys',
        component: () => import('@/pages/rpm-tpm/api-keys/index.vue'),
        meta: {
          title: 'API Keys 管理',
        },
      },
      {
        path: 'accounts',
        name: 'RpmTpmAccounts',
        component: () => import('@/pages/rpm-tpm/accounts/index.vue'),
        meta: {
          title: 'Accounts 管理',
        },
      },
    ],
  },
];
