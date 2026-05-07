<script setup lang="ts">
import gauge from '@/components/chart/gauge.vue';
import type { SysStats } from "@/types";
import type { SystemInfo } from "@/types";

const props = defineProps<{
  data?: SysStats;
  sysInfo?: SystemInfo;
}>();

const memoryUsed = computed(() => {
  if (!props.data || !props.sysInfo?.mem_total) return null;
  const usedGB = ((props.sysInfo.mem_total * (props.data.mem_usage / 100)) / 1024 / 1024 / 1024).toFixed(1);
  const totalGB = (props.sysInfo.mem_total / 1024 / 1024 / 1024).toFixed(1);
  return `${usedGB}GB / ${totalGB}GB`;
});
</script>

<template>
  <div class="card bg-base-100 border border-base-200 shadow-sm">
    <div class="card-body p-6">
      <div class="flex items-center justify-between">
        <h2 class="card-title text-lg font-bold flex items-center gap-2">
          <Icon icon="ic:sharp-pie-chart" class="w-5 h-5 text-primary" />
          系统状态
        </h2>
        <span v-if="memoryUsed" class="text-xs text-base-content/50">{{ memoryUsed }}</span>
      </div>
      <div class="grid grid-cols-2 gap-4">
        <gauge title="CPU使用率" :value="data?.cpu_usage ?? 0"/>
        <gauge title="内存使用" :value="data?.mem_usage ?? 0"/>
      </div>
    </div>
  </div>
</template>
