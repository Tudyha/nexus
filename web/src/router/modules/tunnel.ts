const tunnel = [
  {
    path: "/tunnel",
    name: "Tunnel",
    redirect: "/tunnel/list",
    meta: {
      title: "Tunnel",
      requiresAuth: true,
      hideSidebar: false,
    },
    children: [
      {
        path: "list",
        name: "TunnelList",
        component: () => import("@/views/tunnel/index.vue"),
        meta: {
          title: "隧道管理",
          requiresAuth: true,
          hideSidebar: false,
        },
      },
    ],
  },
];

export default tunnel;
