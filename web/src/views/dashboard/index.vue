<template>
  <div class="space-y-4">
    <!-- header -->
    <div>
      <div class="flex items-center justify-between">
        <h1 class="text-2xl font-bold text-base-content">仪表盘</h1>
        <p class="text-sm text-base-content/50">更新时间: {{ lastUpdateTime }}</p>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
      <div class="space-y-4 lg:col-span-2">
        <!-- 核心统计指标 -->
        <div>
          <stats :data="dashboard.client_stats" />
        </div>
        <div>
          <system-stat :data="dashboard.sys_stats" />
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

onMounted(() => {
  fetchData()
})

// 定时刷新 (30s)
useIntervalFn(fetchData, 30000)
</script>
