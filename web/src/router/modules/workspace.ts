const workspace = [
  {
    path: "/workspace",
    name: "Workspace",
    redirect: "/workspace/list",
    meta: {
      title: "Workspace",
      requiresAuth: true,
      hideSidebar: false,
    },
    children: [
      {
        path: "list",
        name: "WorkspaceList",
        component: () => import("@/views/workspace/index.vue"),
        meta: {
          title: "工作空间",
          requiresAuth: true,
          hideSidebar: false,
        },
      },
      {
        path: ":id",
        name: "WorkspaceDetail",
        component: () => import("@/views/workspace/detail.vue"),
        meta: {
          title: "工作空间详情",
          requiresAuth: true,
          hideSidebar: true,
        },
      },
    ],
  },
];

export default workspace;
