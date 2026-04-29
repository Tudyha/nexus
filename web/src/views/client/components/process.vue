<script setup lang="ts">
import { ref, computed } from 'vue'
import { Icon } from '@iconify/vue'

// --- Types ---
interface ProcessItem {
  pid: number
  name: string
  ppid: number
  threads: number
  user: string
  cpu: number
  memory: number // in bytes
  connections: number
  status: string
  startTime: string
}

// --- State ---
const filterStatus = ref('')
const filterPid = ref('')
const filterName = ref('')
const filterUser = ref('')

// Mock Data
const processList = ref<ProcessItem[]>([
  { pid: 1, name: 'systemd', ppid: 0, threads: 1, user: 'root', cpu: 0.02, memory: 13286195, connections: 20, status: '睡眠', startTime: '2024-4-28 16:08:14' },
  { pid: 2, name: 'kthreadd', ppid: 0, threads: 1, user: 'root', cpu: 0.00, memory: 0, connections: 0, status: '睡眠', startTime: '2024-4-28 16:08:14' },
  { pid: 3, name: 'rcu_gp', ppid: 2, threads: 1, user: 'root', cpu: 0.00, memory: 0, connections: 0, status: '空闲', startTime: '2024-4-28 16:08:14' },
  { pid: 4, name: 'rcu_par_gp', ppid: 2, threads: 1, user: 'root', cpu: 0.00, memory: 0, connections: 0, status: '空闲', startTime: '2024-4-28 16:08:14' },
  { pid: 5, name: 'slub_flushwq', ppid: 2, threads: 1, user: 'root', cpu: 0.00, memory: 0, connections: 0, status: '空闲', startTime: '2024-4-28 16:08:14' },
  { pid: 6, name: 'netns', ppid: 2, threads: 1, user: 'root', cpu: 0.00, memory: 0, connections: 0, status: '空闲', startTime: '2024-4-28 16:08:14' },
  { pid: 8, name: 'kworker/0:0H-events_highpri', ppid: 2, threads: 1, user: 'root', cpu: 0.00, memory: 0, connections: 0, status: '空闲', startTime: '2024-4-28 16:08:14' },
  { pid: 10, name: 'mm_percpu_wq', ppid: 2, threads: 1, user: 'root', cpu: 0.00, memory: 0, connections: 0, status: '空闲', startTime: '2024-4-28 16:08:14' },
  { pid: 11, name: 'rcu_tasks_rude_', ppid: 2, threads: 1, user: 'root', cpu: 0.00, memory: 0, connections: 0, status: '睡眠', startTime: '2024-4-28 16:08:14' },
  { pid: 12, name: 'rcu_tasks_trace', ppid: 2, threads: 1, user: 'root', cpu: 0.00, memory: 0, connections: 0, status: '睡眠', startTime: '2024-4-28 16:08:14' },
  { pid: 13, name: 'ksoftirqd/0', ppid: 2, threads: 1, user: 'root', cpu: 0.00, memory: 0, connections: 0, status: '睡眠', startTime: '2024-4-28 16:08:14' },
  { pid: 14, name: 'rcu_sched', ppid: 2, threads: 1, user: 'root', cpu: 0.08, memory: 0, connections: 0, status: '空闲', startTime: '2024-4-28 16:08:14' },
  { pid: 15, name: 'migration/0', ppid: 2, threads: 1, user: 'root', cpu: 0.00, memory: 0, connections: 0, status: '睡眠', startTime: '2024-4-28 16:08:14' },
])

// --- Helpers ---
const formatMemory = (bytes: number) => {
  if (bytes === 0) return '0B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + sizes[i]
}

const formatCpu = (val: number) => {
  return val.toFixed(2) + '%'
}

// --- Methods ---
const handleDetail = (pid: number) => {
  console.log('Detail', pid)
}

const handleKill = (pid: number) => {
  console.log('Kill', pid)
}

// Filtered Data (Simple Implementation)
const filteredList = computed(() => {
  return processList.value.filter(item => {
    const matchStatus = !filterStatus.value || item.status === filterStatus.value
    const matchPid = !filterPid.value || item.pid.toString().includes(filterPid.value)
    const matchName = !filterName.value || item.name.toLowerCase().includes(filterName.value.toLowerCase())
    const matchUser = !filterUser.value || item.user.toLowerCase().includes(filterUser.value.toLowerCase())
    return matchStatus && matchPid && matchName && matchUser
  })
})

</script>

<template>
  <div class="h-full flex flex-col bg-base-100 p-4 gap-4">
    <!-- Top Filters -->
    <div class="flex justify-start gap-2 flex-wrap">
      <!-- Status -->
      <select v-model="filterStatus" class="select select-bordered select-sm w-32">
        <option value="">状态</option>
        <option value="睡眠">睡眠</option>
        <option value="空闲">空闲</option>
        <option value="运行">运行</option>
      </select>

      <!-- PID Search -->
      <div class="relative">
        <input
          type="text"
          v-model="filterPid"
          placeholder="进程ID"
          class="input input-bordered input-sm w-32 pr-8"
        />
        <Icon icon="mdi:magnify" class="w-4 h-4 absolute right-2 top-1/2 -translate-y-1/2 text-base-content/40" />
      </div>

      <!-- Name Search -->
      <div class="relative">
        <input
          type="text"
          v-model="filterName"
          placeholder="名称"
          class="input input-bordered input-sm w-48 pr-8"
        />
        <Icon icon="mdi:magnify" class="w-4 h-4 absolute right-2 top-1/2 -translate-y-1/2 text-base-content/40" />
      </div>

      <!-- User Search -->
      <div class="relative">
        <input
          type="text"
          v-model="filterUser"
          placeholder="用户"
          class="input input-bordered input-sm w-32 pr-8"
        />
        <Icon icon="mdi:magnify" class="w-4 h-4 absolute right-2 top-1/2 -translate-y-1/2 text-base-content/40" />
      </div>
    </div>

    <!-- Process Table -->
    <div class="flex-1 overflow-hidden flex flex-col bg-base-100 rounded-lg border-t border-base-200">
      <div class="overflow-x-auto h-full">
        <table class="table table-sm table-pin-rows">
          <thead>
            <tr class="border-b border-base-200">
              <th class="bg-base-100 text-base-content/60 font-medium">
                <div class="flex items-center gap-1 cursor-pointer hover:text-base-content">
                  PID
                  <Icon icon="mdi:arrow-up" class="w-3 h-3" />
                </div>
              </th>
              <th class="bg-base-100 text-base-content/60 font-medium">名称</th>
              <th class="bg-base-100 text-base-content/60 font-medium">父进程ID</th>
              <th class="bg-base-100 text-base-content/60 font-medium">线程</th>
              <th class="bg-base-100 text-base-content/60 font-medium">用户</th>
              <th class="bg-base-100 text-base-content/60 font-medium">CPU</th>
              <th class="bg-base-100 text-base-content/60 font-medium">内存</th>
              <th class="bg-base-100 text-base-content/60 font-medium">连接</th>
              <th class="bg-base-100 text-base-content/60 font-medium">状态</th>
              <th class="bg-base-100 text-base-content/60 font-medium">启动时间</th>
              <th class="bg-base-100 text-base-content/60 font-medium">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="proc in filteredList" :key="proc.pid" class="hover border-b border-base-100">
              <td class="font-mono">{{ proc.pid }}</td>
              <td>{{ proc.name }}</td>
              <td>{{ proc.ppid }}</td>
              <td>{{ proc.threads }}</td>
              <td>{{ proc.user }}</td>
              <td>{{ formatCpu(proc.cpu) }}</td>
              <td>{{ formatMemory(proc.memory) }}</td>
              <td>{{ proc.connections }}</td>
              <td>{{ proc.status }}</td>
              <td class="text-base-content/70 text-xs">{{ proc.startTime }}</td>
              <td>
                <div class="flex items-center gap-3">
                  <button
                    class="text-blue-600 hover:text-blue-800 text-xs font-medium"
                    @click="handleDetail(proc.pid)"
                  >
                    查看详情
                  </button>
                  <button
                    class="text-blue-600 hover:text-blue-800 text-xs font-medium"
                    @click="handleKill(proc.pid)"
                  >
                    结束
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Remove table inner borders to match screenshot style more closely if needed */
.table :where(thead, tbody) :where(tr:not(:last-child)), .table :where(thead, tbody) :where(tr:first-child:last-child) {
  border-bottom-color: transparent;
}
/* Add explicit border bottom to rows to mimic the screenshot's light separator */
tbody tr {
  border-bottom: 1px solid var(--fallback-b2,oklch(var(--b2)/var(--tw-border-opacity))) !important;
}
</style>
