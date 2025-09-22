import { UserCircleIcon } from 'tdesign-icons-vue-next';
import { shallowRef } from 'vue';

import Layout from '@/layouts/index.vue';

export default [
  {
    path: '/accounts',
    component: Layout,
    redirect: '/accounts/list',
    name: 'Accounts',
    meta: {
      title: '账户管理',
      icon: shallowRef(UserCircleIcon),
      orderNo: 2,
    },
    children: [
      {
        path: 'list',
        name: 'AccountsList',
        component: () => import('@/pages/accounts/list/index.vue'),
        meta: {
          title: '账户列表',
        },
      },
    ],
  },
];
