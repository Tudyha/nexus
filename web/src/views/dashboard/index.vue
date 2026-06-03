<template>
  <div class="space-y-4">
    <!-- header -->
    <div class="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
      <div>
        <h1 class="text-2xl font-bold text-base-content">{{ greeting }}{{ dashboard.sys_info?.hostname ? `，${dashboard.sys_info.hostname}` : '' }}</h1>
        <p class="text-sm text-base-content/50 mt-1">
          <span>在线 {{ clientsOnline }} / 共 {{ clientsTotal }} 台主机</span>
          <span v-if="clientsTotal > 0" class="ml-2">· 在线率 {{ onlineRatePct }}%</span>
        </p>
      </div>
      <p class="text-sm text-base-content/50">更新时间: {{ lastUpdateTime }}</p>
    </div>

    <!-- Prometheus 实时指标 -->
    <div class="card bg-base-100 border border-base-200 shadow-sm">
      <div class="card-body p-6">
        <div class="flex items-center justify-between mb-1">
          <h2 class="card-title text-lg font-bold flex items-center gap-2">
            <Icon icon="mdi:chart-line" class="w-5 h-5 text-primary" />
            实时指标
          </h2>
          <div v-if="metricsError" class="text-xs text-error">{{ metricsError }}</div>
        </div>
        <div class="stats stats-vertical lg:stats-horizontal w-full">
          <div class="stat">
            <div class="stat-figure text-base-content/40">
              <Icon icon="mdi:server" class="w-6 h-6" />
            </div>
            <div class="stat-title text-base-content/50 text-sm">注册客户端</div>
            <div class="stat-value text-2xl text-base-content/70">{{ clientsTotal }}</div>
          </div>
          <div class="stat">
            <div class="stat-figure text-primary">
              <Icon icon="mdi:server" class="w-6 h-6" />
            </div>
            <div class="stat-title text-base-content/50 text-sm">在线客户端</div>
            <div class="stat-value text-2xl text-primary">{{ clientsOnline }}</div>
          </div>
          <div class="stat">
            <div class="stat-figure text-base-content/30">
              <Icon icon="mdi:server-off" class="w-6 h-6" />
            </div>
            <div class="stat-title text-base-content/50 text-sm">离线客户端</div>
            <div class="stat-value text-2xl text-base-content/40">{{ clientsOffline }}</div>
          </div>
          <div class="stat">
            <div class="stat-figure text-secondary">
              <Icon icon="mdi:connection" class="w-6 h-6" />
            </div>
            <div class="stat-title text-base-content/50 text-sm">活跃 Session</div>
            <div class="stat-value text-2xl text-secondary">{{ sessionsActive }}</div>
          </div>
          <div class="stat">
            <div class="stat-figure text-accent">
              <Icon icon="mdi:pipe-disconnected" class="w-6 h-6" />
            </div>
            <div class="stat-title text-base-content/50 text-sm">活跃隧道</div>
            <div class="stat-value text-2xl text-accent">{{ tunnelsActive }}</div>
          </div>
        </div>
      </div>
    </div>

    <!-- Tunnel Traffic -->
    <div class="card bg-base-100 border border-base-200 shadow-sm">
      <div class="card-body p-6">
        <div class="flex items-center justify-between mb-4">
          <h2 class="card-title text-lg font-bold flex items-center gap-2">
            <Icon icon="mdi:chart-bar" class="w-5 h-5 text-primary" />
            隧道流量统计
          </h2>
        </div>
        <div v-if="!hasTraffic" class="text-center py-4 text-base-content/40 text-sm">
          暂无隧道流量数据
        </div>
        <div v-else class="overflow-x-auto -mx-2">
          <table class="table table-sm">
            <thead>
              <tr>
                <th>客户端 ID</th>
                <th>隧道类型</th>
                <th>发送</th>
                <th>接收</th>
                <th>总计</th>
              </tr>
            </thead>
            <tbody>
              <tr v-for="t in tunnelTraffic" :key="t.key">
                <td class="font-mono text-xs">#{{ t.clientId }}</td>
                <td>
                  <span class="badge badge-ghost badge-xs">{{ t.tunnelType === '1' ? 'TCP' : 'UDP' }}</span>
                </td>
                <td class="font-mono text-xs">{{ formatBytes(t.sent) }}</td>
                <td class="font-mono text-xs">{{ formatBytes(t.recv) }}</td>
                <td class="font-mono text-xs font-semibold">{{ formatBytes(t.sent + t.recv) }}</td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </div>

    <div class="grid grid-cols-1 lg:grid-cols-3 gap-4">
      <div class="space-y-4 lg:col-span-2">
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
import { Icon } from '@iconify/vue'
import http from '@/api/index'
import type { DashboardResponse } from '@/types'
import SystemInfo from './components/system.vue'
import SystemStat from './components/system-stat.vue'
import QuickAccess from './components/quick-access.vue'
import Monitor from './components/monitor.vue'

// --- Types ---
interface MetricValue {
  value: number
  labels?: Record<string, string>
}
type MetricsMap = Record<string, MetricValue[]>

// --- Dashboard state ---
const lastUpdateTime = ref(dayjs().format('HH:mm:ss'))
const dashboard = ref<DashboardResponse>({})

// --- Metrics state ---
const metrics = ref<MetricsMap>({})
const metricsError = ref('')

// --- Metrics computed ---
const clientsTotal = computed(() => metrics.value['nexus_clients_total']?.[0]?.value ?? 0)
const clientsOnline = computed(() => metrics.value['nexus_clients_online']?.[0]?.value ?? 0)
const clientsOffline = computed(() => Math.max(0, clientsTotal.value - clientsOnline.value))
const onlineRatePct = computed(() => {
  const total = clientsTotal.value
  if (total === 0) return '0.0'
  return ((clientsOnline.value / total) * 100).toFixed(1)
})
const sessionsActive = computed(() => metrics.value['nexus_sessions_active']?.[0]?.value ?? 0)
const tunnelsActive = computed(() => metrics.value['nexus_tunnels_active']?.[0]?.value ?? 0)
const sessionsTotal = computed(() => metrics.value['nexus_sessions_total']?.[0]?.value ?? 0)

const tunnelTraffic = computed(() => {
  const sent = metrics.value['nexus_tunnel_bytes_sent_total'] || []
  const recv = metrics.value['nexus_tunnel_bytes_received_total'] || []
  const map = new Map<string, { sent: number; recv: number }>()

  for (const m of sent) {
    const key = `${m.labels?.tunnel_type ?? '?'}-${m.labels?.client_id ?? '?'}`
    if (!map.has(key)) map.set(key, { sent: 0, recv: 0 })
    map.get(key)!.sent = m.value
  }
  for (const m of recv) {
    const key = `${m.labels?.tunnel_type ?? '?'}-${m.labels?.client_id ?? '?'}`
    if (!map.has(key)) map.set(key, { sent: 0, recv: 0 })
    map.get(key)!.recv = m.value
  }
  return Array.from(map.entries()).map(([key, v]) => ({
    key,
    tunnelType: key.split('-')[0],
    clientId: key.split('-')[1],
    sent: v.sent,
    recv: v.recv,
  }))
})

const hasTraffic = computed(() => tunnelTraffic.value.length > 0)

function formatBytes(val: number): string {
  if (val === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(val) / Math.log(1024))
  return (val / Math.pow(1024, i)).toFixed(i > 0 ? 2 : 0) + ' ' + units[i]
}

async function fetchMetrics() {
  metricsError.value = ''
  try {
    metrics.value = await http.get('/v1/ops/metrics')
  } catch (e: any) {
    metricsError.value = e.message || '获取指标失败'
  }
}

const fetchData = async () => {
  try {
    const [res] = await Promise.all([
      getDashboard(),
      fetchMetrics(),
    ])
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

onMounted(() => {
  fetchData()
})

// 定时刷新 (30s)
useIntervalFn(fetchData, 30000)
</script>
