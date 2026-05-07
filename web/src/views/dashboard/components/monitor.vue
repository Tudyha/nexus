<script setup lang="ts">
import type { SysStats } from "@/types";
import { Line } from "vue-chartjs";
import {
  Chart as ChartJS,
  CategoryScale,
  LinearScale,
  PointElement,
  LineElement,
  Filler,
  Title,
  Tooltip,
  Legend,
} from "chart.js";

ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Filler, Title, Tooltip, Legend);

const props = defineProps<{
  data?: SysStats;
}>();

const MAX_POINTS = 20;
const timestamps = ref<string[]>([]);
const cpuHistory = ref<number[]>([]);
const memHistory = ref<number[]>([]);

// Defensive copies so Chart.js mutations don't trigger reactive loops
const chartData = computed(() => ({
  labels: [...timestamps.value],
  datasets: [
    {
      label: "CPU",
      data: [...cpuHistory.value],
      borderColor: "#3b82f6",
      backgroundColor: "rgba(59, 130, 246, 0.1)",
      fill: true,
      tension: 0.3,
      pointRadius: 2,
    },
    {
      label: "Memory",
      data: [...memHistory.value],
      borderColor: "#10b981",
      backgroundColor: "rgba(16, 185, 129, 0.1)",
      fill: true,
      tension: 0.3,
      pointRadius: 2,
    },
  ],
}));

const chartOptions = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: {
      display: true,
      position: "top" as const,
      labels: { boxWidth: 12, padding: 16 },
    },
    tooltip: {
      callbacks: {
        label: (ctx: any) => `${ctx.dataset.label}: ${ctx.parsed.y}%`,
      },
    },
  },
  scales: {
    x: {
      display: true,
      grid: { display: false },
      ticks: { maxTicksLimit: 8 },
    },
    y: {
      min: 0,
      max: 100,
      ticks: { callback: (v: any) => `${v}%` },
    },
  },
};

function addDataPoint(val: SysStats) {
  const now = new Date();
  const time = `${String(now.getHours()).padStart(2, "0")}:${String(now.getMinutes()).padStart(2, "0")}:${String(now.getSeconds()).padStart(2, "0")}`;

  timestamps.value.push(time);
  cpuHistory.value.push(+val.cpu_usage.toFixed(1));
  memHistory.value.push(+val.mem_usage.toFixed(1));

  if (timestamps.value.length > MAX_POINTS) {
    timestamps.value.shift();
    cpuHistory.value.shift();
    memHistory.value.shift();
  }
}

onMounted(() => {
  if (props.data) {
    addDataPoint(props.data);
  }
});

watch(() => props.data, (val) => {
  if (val) addDataPoint(val);
});
</script>

<template>
  <div class="card bg-base-100 border border-base-200 shadow-sm">
    <div class="card-body p-6">
      <div class="flex items-center justify-between mb-4">
        <h2 class="card-title text-lg font-bold flex items-center gap-2">
          <Icon icon="mdi:monitor" class="w-5 h-5 text-primary" />
          系统监控
        </h2>
        <span class="text-xs text-base-content/40">实时更新 (30s)</span>
      </div>
      <div class="h-64" v-if="cpuHistory.length > 1">
        <Line :data="chartData" :options="chartOptions" />
      </div>
      <div v-else class="h-64 flex items-center justify-center text-base-content/40">
        等待数据...
      </div>
    </div>
  </div>
</template>
