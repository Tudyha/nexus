<script setup lang="ts">
import gauge from '@/components/chart/gauge.vue';
import type { SysStats } from "@/types";

const props = defineProps<{
  data?: SysStats;
}>();

</script>

<template>
  <div class="card bg-base-100 border border-base-200 shadow-sm">
    <div class="card-body p-6">
      <div class="flex items-center justify-between">
        <h2 class="card-title text-lg font-bold flex items-center gap-2">
          <Icon icon="ic:sharp-pie-chart" class="w-5 h-5 text-primary" />
          系统状态
        </h2>
      </div>
      <div class="grid grid-cols-3 gap-4">
        <gauge title="cpu使用率" :value="props.data?.cpu_usage ?? 0"/>
        <gauge title="内存使用" :value="props.data?.mem_usage ?? 0"/>
        <gauge title="磁盘使用" :value="props.data?.cpu_usage ?? 0"/>
      </div>
    </div>
  </div>
</template>