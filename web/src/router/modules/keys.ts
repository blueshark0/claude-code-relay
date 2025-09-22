import { KeyIcon } from 'tdesign-icons-vue-next';
import { shallowRef } from 'vue';

import Layout from '@/layouts/index.vue';

export default [
  {
    path: '/keys',
    component: Layout,
    redirect: '/keys/list',
    name: 'Keys',
    meta: {
      title: 'API 密钥管理',
      icon: shallowRef(KeyIcon),
      orderNo: 3,
    },
    children: [
      {
        path: 'list',
        name: 'KeysList',
        component: () => import('@/pages/keys/list/index.vue'),
        meta: {
          title: '密钥列表',
        },
      },
    ],
  },
];
