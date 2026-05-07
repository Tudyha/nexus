<template>
  <div class="space-y-4">
    <!-- header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
      <div>
        <h1 class="text-2xl font-bold text-base-content">{{ greeting }}{{ dashboard.sys_info?.hostname ? `，${dashboard.sys_info.hostname}` : '' }}</h1>
        <p class="text-sm text-base-content/50 mt-1">
          <span>在线 {{ dashboard.client_stats?.online ?? 0 }} / 共 {{ totalClients }} 台主机</span>
          <span v-if="totalClients > 0" class="ml-2">· 在线率 {{ onlineRate }}%</span>
        </p>
      </div>
      <p class="text-sm text-base-content/50">更新时间: {{ lastUpdateTime }}</p>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
      <div class="space-y-4 lg:col-span-2">
        <!-- 核心统计指标 -->
        <div>
          <stats :data="dashboard.client_stats" />
        </div>
        <div>
          <system-stat :data="dashboard.sys_stats" :sys-info="dashboard.sys_info" />
        </div>
        <div>
          <monitor :data="dashboard.sys_stats" />
        </div>
      </div>
      <!-- 系统信息 -->
      <div class="lg:col-span-1 space-y-4">
        <system-info :system-info="dashboard.sys_info" />
        <div>
          <quick-access />
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { getDashboard } from '@/api/dashboard'
import dayjs from 'dayjs'
import type { DashboardResponse } from '@/types'
import SystemInfo from './components/system.vue'
import Stats from './components/stat.vue'
import SystemStat from './components/system-stat.vue'
import QuickAccess from './components/quick-access.vue'
import Monitor from './components/monitor.vue'

const lastUpdateTime = ref(dayjs().format('HH:mm:ss'))
const dashboard = ref<DashboardResponse>({})
const fetchData = async () => {
  try {
    const res = await getDashboard()
    dashboard.value = { ...res }
    lastUpdateTime.value = dayjs().format('HH:mm:ss')
  } catch (error) {
    console.error('Failed to fetch dashboard data:', error)
  }
}

const greeting = computed(() => {
  const hour = dayjs().hour()
  if (hour < 6) return '夜深了'
  if (hour < 9) return '早上好'
  if (hour < 12) return '上午好'
  if (hour < 14) return '中午好'
  if (hour < 18) return '下午好'
  return '晚上好'
})

const totalClients = computed(() => {
  const s = dashboard.value.client_stats
  if (!s) return 0
  return s.online + s.offline
})

const onlineRate = computed(() => {
  const total = totalClients.value
  if (total === 0) return 0
  return ((dashboard.value.client_stats?.online ?? 0) / total * 100).toFixed(1)
})

onMounted(() => {
  fetchData()
})

// 定时刷新 (30s)
useIntervalFn(fetchData, 30000)
</script>
