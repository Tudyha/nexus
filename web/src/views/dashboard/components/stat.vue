<script setup lang="ts">
import type { ClientStats } from "@/types";
import { Icon } from "@iconify/vue";

const props = defineProps<{
  data?: ClientStats;
}>();

const stats = computed(() => [
  {
    title: "在线主机数",
    value: props.data?.online || 0,
    icon: "material-symbols:online-prediction",
    color: "success",
  },
  {
    title: "离线主机数",
    value: props.data?.offline || 0,
    icon: "hugeicons:cellular-network-offline",
    color: "error",
  },
]);

</script>

<template>
  <div class="card bg-base-100 border border-base-200 shadow-sm">
    <div class="card-body p-6">
      <div class="flex items-center justify-between">
        <h2 class="card-title text-lg font-bold flex items-center gap-2">
          <Icon icon="material-symbols:computer-outline" class="w-5 h-5 text-primary" />
          主机概览
        </h2>
      </div>
      <div class="stats">
        <div v-for="item in stats" :key="item.title" class="stat">
          <div class="stat-figure" :class="`text-${item.color}`">
            <Icon :icon="item.icon" class="w-6 h-6" />
          </div>
          <div class="stat-title">{{ item.title }}</div>
          <div :class="['stat-value', 'link', `link-${item.color}`]">
            <router-link to="/client">{{ item.value }}</router-link>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>