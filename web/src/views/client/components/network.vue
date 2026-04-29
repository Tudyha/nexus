<script setup lang="ts">
import { ref, computed } from 'vue'
import { Icon } from '@iconify/vue'

// --- Types ---
interface NetworkItem {
  id: string
  type: string
  pid: number
  processName: string
  localAddress: string
  remoteAddress: string
  status: 'LISTEN' | 'ESTABLISHED' | 'TIME_WAIT' | 'CLOSE_WAIT' | 'SYN_SENT' | 'SYN_RECV' | 'FIN_WAIT1' | 'FIN_WAIT2'
}

// --- State ---
const filterPid = ref('')
const filterName = ref('')
const filterPort = ref('')
const selectedStatuses = ref<string[]>(['LISTEN', 'ESTABLISHED'])
const isStatusDropdownOpen = ref(false)

const availableStatuses = [
  'LISTEN', 'ESTABLISHED', 'TIME_WAIT', 'CLOSE_WAIT',
  'SYN_SENT', 'SYN_RECV', 'FIN_WAIT1', 'FIN_WAIT2'
]

// Mock Data
const networkList = ref<NetworkItem[]>([
  { id: '1', type: 'tcp', pid: 129990, processName: 'docker-proxy', localAddress: '0.0.0.0:3002', remoteAddress: '0.0.0.0', status: 'LISTEN' },
  { id: '2', type: 'tcp6', pid: 129997, processName: 'docker-proxy', localAddress: ':::3002', remoteAddress: '::', status: 'LISTEN' },
  { id: '3', type: 'tcp', pid: 1211859, processName: 'AliYunDun', localAddress: '172.16.0.166:47076', remoteAddress: '100.100.30.26:80', status: 'ESTABLISHED' },
  { id: '4', type: 'tcp', pid: 1451559, processName: 'sshd', localAddress: '0.0.0.0:22', remoteAddress: '0.0.0.0', status: 'LISTEN' },
  { id: '5', type: 'tcp', pid: 2556185, processName: 'openresty', localAddress: '0.0.0.0:80', remoteAddress: '0.0.0.0', status: 'LISTEN' },
  { id: '6', type: 'tcp6', pid: 2556185, processName: 'openresty', localAddress: ':::80', remoteAddress: '::', status: 'LISTEN' },
  { id: '7', type: 'tcp', pid: 2556310, processName: 'docker-proxy', localAddress: '127.0.0.1:5432', remoteAddress: '0.0.0.0', status: 'LISTEN' },
  { id: '8', type: 'tcp', pid: 2696896, processName: 'docker-proxy', localAddress: '127.0.0.1:6379', remoteAddress: '0.0.0.0', status: 'LISTEN' },
  { id: '9', type: 'tcp', pid: 2696932, processName: 'docker-proxy', localAddress: '127.0.0.1:3306', remoteAddress: '0.0.0.0', status: 'LISTEN' },
  { id: '10', type: 'tcp', pid: 2699068, processName: 'docker-proxy', localAddress: '127.0.0.1:11434', remoteAddress: '0.0.0.0', status: 'LISTEN' },
  { id: '11', type: 'tcp', pid: 3239833, processName: '1panel-core', localAddress: '0.0.0.0:9999', remoteAddress: '0.0.0.0', status: 'LISTEN' },
  { id: '12', type: 'tcp', pid: 3239833, processName: '1panel-core', localAddress: '172.16.0.166:9999', remoteAddress: '47.92.89.141:41176', status: 'ESTABLISHED' },
  { id: '13', type: 'tcp', pid: 4171312, processName: 'systemd-resolve', localAddress: '127.0.0.53:53', remoteAddress: '0.0.0.0', status: 'LISTEN' },
])

// --- Methods ---
const toggleStatus = (status: string) => {
  const index = selectedStatuses.value.indexOf(status)
  if (index === -1) {
    selectedStatuses.value.push(status)
  } else {
    selectedStatuses.value.splice(index, 1)
  }
}

const removeStatus = (status: string) => {
  const index = selectedStatuses.value.indexOf(status)
  if (index !== -1) {
    selectedStatuses.value.splice(index, 1)
  }
}

const quickFilter = (field: 'processName' | 'port', value: string) => {
  if (field === 'processName') {
    filterName.value = value
  } else if (field === 'port') {
    // Extract port from address (e.g., 0.0.0.0:3002 -> 3002)
    const port = value.split(':').pop()
    if (port) filterPort.value = port
  }
}

// Filtered Data
const filteredList = computed(() => {
  return networkList.value.filter(item => {
    const matchStatus = selectedStatuses.value.length === 0 || selectedStatuses.value.includes(item.status)
    const matchPid = !filterPid.value || item.pid.toString().includes(filterPid.value)
    const matchName = !filterName.value || item.processName.toLowerCase().includes(filterName.value.toLowerCase())

    // Port filtering check on local address
    const localPort = item.localAddress.split(':').pop() || ''
    const matchPort = !filterPort.value || localPort.includes(filterPort.value)

    return matchStatus && matchPid && matchName && matchPort
  })
})

</script>

<template>
  <div class="h-full flex flex-col bg-base-100 p-4 gap-4">
    <!-- Top Filters -->
    <div class="flex justify-start gap-2 flex-wrap items-center">
      <!-- Status Multi-select -->
      <div class="dropdown dropdown-end">
        <div
          tabindex="0"
          role="button"
          class="input input-bordered input-sm min-w-[200px] flex items-center gap-1 flex-wrap py-1 h-auto min-h-[2rem]"
          @click="isStatusDropdownOpen = !isStatusDropdownOpen"
        >
          <span v-if="selectedStatuses.length === 0" class="text-base-content/40 text-xs">选择状态</span>
          <div
            v-for="status in selectedStatuses"
            :key="status"
            class="badge badge-xs gap-1 bg-base-200 border-none text-base-content/70 rounded-sm px-1.5 py-2"
          >
            {{ status }}
            <Icon
              icon="mdi:close"
              class="w-3 h-3 cursor-pointer hover:text-base-content"
              @click.stop="removeStatus(status)"
            />
          </div>
          <Icon icon="mdi:chevron-down" class="w-4 h-4 ml-auto text-base-content/40" />
        </div>
        <ul tabindex="0" class="dropdown-content z-[1] menu p-2 shadow bg-base-100 rounded-box w-52 mt-1">
          <li v-for="status in availableStatuses" :key="status">
            <a
              @click="toggleStatus(status)"
              :class="{ 'active': selectedStatuses.includes(status) }"
            >
              {{ status }}
              <Icon v-if="selectedStatuses.includes(status)" icon="mdi:check" class="w-4 h-4 ml-auto" />
            </a>
          </li>
        </ul>
      </div>

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
          placeholder="进程名称"
          class="input input-bordered input-sm w-32 pr-8"
        />
        <Icon icon="mdi:magnify" class="w-4 h-4 absolute right-2 top-1/2 -translate-y-1/2 text-base-content/40" />
      </div>

      <!-- Port Search -->
      <div class="relative">
        <input
          type="text"
          v-model="filterPort"
          placeholder="端口"
          class="input input-bordered input-sm w-32 pr-8"
        />
        <Icon icon="mdi:magnify" class="w-4 h-4 absolute right-2 top-1/2 -translate-y-1/2 text-base-content/40" />
      </div>
    </div>

    <!-- Network Table -->
    <div class="flex-1 overflow-hidden flex flex-col bg-base-100 rounded-lg border-t border-base-200">
      <div class="overflow-x-auto h-full">
        <table class="table table-sm table-pin-rows">
          <thead>
            <tr class="border-b border-base-200">
              <th class="bg-base-100 text-base-content/60 font-medium">类型</th>
              <th class="bg-base-100 text-base-content/60 font-medium">
                <div class="flex items-center gap-1 cursor-pointer hover:text-base-content">
                  PID
                  <Icon icon="mdi:arrow-up" class="w-3 h-3" />
                </div>
              </th>
              <th class="bg-base-100 text-base-content/60 font-medium">进程名称</th>
              <th class="bg-base-100 text-base-content/60 font-medium">本地地址/端口</th>
              <th class="bg-base-100 text-base-content/60 font-medium">远程地址/端口</th>
              <th class="bg-base-100 text-base-content/60 font-medium">状态</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="item in filteredList" :key="item.id" class="hover border-b border-base-100">
              <td class="text-base-content/70">{{ item.type }}</td>
              <td class="font-mono">{{ item.pid }}</td>
              <td>
                <div class="flex items-center gap-1 group">
                  {{ item.processName }}
                  <button
                    class="opacity-0 group-hover:opacity-100 transition-opacity btn btn-ghost btn-xs btn-square h-5 w-5 min-h-0"
                    title="筛选此名称"
                    @click="quickFilter('processName', item.processName)"
                  >
                    <Icon icon="mdi:filter-variant" class="w-3 h-3 text-base-content/50" />
                  </button>
                </div>
              </td>
              <td class="font-mono text-xs">
                <div class="flex items-center gap-1 group">
                  {{ item.localAddress }}
                  <button
                    class="opacity-0 group-hover:opacity-100 transition-opacity btn btn-ghost btn-xs btn-square h-5 w-5 min-h-0"
                    title="筛选此端口"
                    @click="quickFilter('port', item.localAddress)"
                  >
                    <Icon icon="mdi:filter-variant" class="w-3 h-3 text-base-content/50" />
                  </button>
                </div>
              </td>
              <td class="font-mono text-xs text-base-content/70">{{ item.remoteAddress }}</td>
              <td>
                <span class="text-xs font-medium" :class="{
                  'text-success': item.status === 'ESTABLISHED',
                  'text-base-content/70': item.status === 'LISTEN'
                }">{{ item.status }}</span>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Remove table inner borders to match screenshot style */
.table :where(thead, tbody) :where(tr:not(:last-child)), .table :where(thead, tbody) :where(tr:first-child:last-child) {
  border-bottom-color: transparent;
}
tbody tr {
  border-bottom: 1px solid var(--fallback-b2,oklch(var(--b2)/var(--tw-border-opacity))) !important;
}

/* Custom scrollbar for dropdown if needed */
.dropdown-content {
  max-height: 300px;
  overflow-y: auto;
}
</style>
