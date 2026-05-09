const version = [
  {
    path: "/version",
    name: "Version",
    redirect: "/version/list",
    meta: {
      title: "Version",
      requiresAuth: true,
      hideSidebar: false,
    },
    children: [
      {
        path: "list",
        name: "VersionList",
        component: () => import("@/views/version/index.vue"),
        meta: {
          title: "版本管理",
          requiresAuth: true,
          hideSidebar: false,
        },
      },
    ],
  },
];

export default version;
