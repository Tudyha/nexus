<script setup lang="ts">
const items = [
  {
    label: "主机列表",
    desc: "查看和管理已连接的客户端",
    icon: "material-symbols:computer-outline",
    route: "/client",
    color: "text-primary",
    bg: "bg-primary/10",
  },
  {
    label: "用户管理",
    desc: "管理系统用户及权限",
    icon: "material-symbols:people-outline",
    route: "/user/list",
    color: "text-secondary",
    bg: "bg-secondary/10",
  },
  {
    label: "仪表盘",
    desc: "返回系统概览",
    icon: "material-symbols:dashboard-outline",
    route: "/dashboard",
    color: "text-success",
    bg: "bg-success/10",
  },
];
</script>

<template>
  <div class="card bg-base-100 border border-base-200 shadow-sm">
    <div class="card-body p-6">
      <div class="flex items-center justify-between mb-4">
        <h2 class="card-title text-lg font-bold flex items-center gap-2">
          <Icon icon="tabler:accessible-filled" class="w-5 h-5 text-primary" />
          快捷入口
        </h2>
      </div>
      <div class="space-y-3">
        <router-link
          v-for="item in items"
          :key="item.label"
          :to="item.route"
          class="flex items-center gap-4 p-3 rounded-lg hover:bg-base-200 transition-colors"
        >
          <div :class="['w-10 h-10 rounded-lg flex items-center justify-center', item.bg]">
            <Icon :icon="item.icon" :class="['w-5 h-5', item.color]" />
          </div>
          <div class="flex-1 min-w-0">
            <div class="font-medium text-sm">{{ item.label }}</div>
            <div class="text-xs text-base-content/50 truncate">{{ item.desc }}</div>
          </div>
          <Icon icon="material-symbols:chevron-right" class="w-4 h-4 text-base-content/30" />
        </router-link>
      </div>
    </div>
  </div>
</template>
