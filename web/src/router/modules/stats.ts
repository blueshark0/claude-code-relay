import { ChartBarIcon } from 'tdesign-icons-vue-next';
import { shallowRef } from 'vue';

import Layout from '@/layouts/index.vue';

export default [
  {
    path: '/stats',
    component: Layout,
    redirect: '/stats/api-key',
    name: 'Stats',
    meta: {
      title: '统计分析',
      icon: shallowRef(ChartBarIcon),
      orderNo: 4,
    },
    children: [
      {
        path: 'api-key',
        name: 'ApiKeyStats',
        component: () => import('@/pages/stats/api-key/index.vue'),
        meta: {
          title: 'API Key 统计',
        },
      },
    ],
  },
];
