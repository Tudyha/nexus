const ops = [
  {
    path: "/ops",
    name: "Ops",
    redirect: "/ops/config",
    meta: {
      title: "Ops",
      requiresAuth: true,
      hideSidebar: false,
    },
    children: [
      {
        path: "config",
        name: "OpsConfig",
        component: () => import("@/views/ops/config/index.vue"),
        meta: {
          title: "应用配置",
          requiresAuth: true,
          hideSidebar: false,
        },
      },
    ],
  },
];

export default ops;
