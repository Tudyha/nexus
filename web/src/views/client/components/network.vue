<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Icon } from '@iconify/vue'
import { getNetworkList } from '@/api/ops'
import type { NetworkEntry } from '@/types'

const props = defineProps<{ id: string }>()

// --- State ---
const networkList = ref<NetworkEntry[]>([])
const loading = ref(false)
const filterPid = ref('')
const filterName = ref('')
const filterPort = ref('')
const selectedStatuses = ref<string[]>(['LISTEN', 'ESTABLISHED'])
const isStatusDropdownOpen = ref(false)

const availableStatuses = [
  'LISTEN', 'ESTABLISHED', 'TIME_WAIT', 'CLOSE_WAIT',
  'SYN_SENT', 'SYN_RECV', 'FIN_WAIT1', 'FIN_WAIT2'
]

// --- Load ---
async function load() {
  loading.value = true
  try {
    networkList.value = await getNetworkList(Number(props.id))
  } catch {
    networkList.value = []
  } finally {
    loading.value = false
  }
}

onMounted(load)

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
    const port = value.split(':').pop()
    if (port) filterPort.value = port
  }
}

const statusColor = (status: string) => {
  if (status === 'ESTABLISHED') return 'text-success'
  if (status === 'LISTEN') return 'text-base-content/70'
  if (status === 'TIME_WAIT' || status === 'CLOSE_WAIT') return 'text-warning'
  return 'text-base-content/50'
}

// --- Filters ---
const filteredList = computed(() => {
  return networkList.value.filter(item => {
    const matchStatus = selectedStatuses.value.length === 0 || selectedStatuses.value.includes(item.status)
    const matchPid = !filterPid.value || item.pid.toString().includes(filterPid.value)
    const matchName = !filterName.value || item.process_name.toLowerCase().includes(filterName.value.toLowerCase())
    const localPort = item.local_addr.split(':').pop() || ''
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
            <Icon icon="mdi:close" class="w-3 h-3 cursor-pointer hover:text-base-content" @click.stop="removeStatus(status)" />
          </div>
          <Icon icon="mdi:chevron-down" class="w-4 h-4 ml-auto text-base-content/40" />
        </div>
        <ul tabindex="0" class="dropdown-content z-[1] menu p-2 shadow bg-base-100 rounded-box w-52 mt-1">
          <li v-for="status in availableStatuses" :key="status">
            <a @click="toggleStatus(status)" :class="{ 'active': selectedStatuses.includes(status) }">
              {{ status }}
              <Icon v-if="selectedStatuses.includes(status)" icon="mdi:check" class="w-4 h-4 ml-auto" />
            </a>
          </li>
        </ul>
      </div>

      <!-- PID Search -->
      <div class="relative">
        <input type="text" v-model="filterPid" placeholder="进程ID" class="input input-bordered input-sm w-32 pr-8" />
        <Icon icon="mdi:magnify" class="w-4 h-4 absolute right-2 top-1/2 -translate-y-1/2 text-base-content/40" />
      </div>

      <!-- Name Search -->
      <div class="relative">
        <input type="text" v-model="filterName" placeholder="进程名称" class="input input-bordered input-sm w-32 pr-8" />
        <Icon icon="mdi:magnify" class="w-4 h-4 absolute right-2 top-1/2 -translate-y-1/2 text-base-content/40" />
      </div>

      <!-- Port Search -->
      <div class="relative">
        <input type="text" v-model="filterPort" placeholder="端口" class="input input-bordered input-sm w-32 pr-8" />
        <Icon icon="mdi:magnify" class="w-4 h-4 absolute right-2 top-1/2 -translate-y-1/2 text-base-content/40" />
      </div>

      <button class="btn btn-ghost btn-sm ml-auto" @click="load" :disabled="loading">
        <Icon icon="mdi:refresh" class="w-4 h-4" />
        {{ loading ? '刷新中...' : '刷新' }}
      </button>
    </div>

    <!-- Network Table -->
    <div class="flex-1 overflow-hidden flex flex-col bg-base-100 rounded-lg border-t border-base-200">
      <div class="overflow-x-auto h-full">
        <table class="table table-sm table-pin-rows">
          <thead>
            <tr class="border-b border-base-200">
              <th class="bg-base-100 text-base-content/60 font-medium">类型</th>
              <th class="bg-base-100 text-base-content/60 font-medium">PID</th>
              <th class="bg-base-100 text-base-content/60 font-medium">进程名称</th>
              <th class="bg-base-100 text-base-content/60 font-medium">本地地址/端口</th>
              <th class="bg-base-100 text-base-content/60 font-medium">远程地址/端口</th>
              <th class="bg-base-100 text-base-content/60 font-medium">状态</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(item, idx) in filteredList" :key="idx" class="hover border-b border-base-100">
              <td class="text-base-content/70">{{ item.protocol }}</td>
              <td class="font-mono">{{ item.pid }}</td>
              <td>
                <div class="flex items-center gap-1 group">
                  {{ item.process_name }}
                  <button
                    class="opacity-0 group-hover:opacity-100 transition-opacity btn btn-ghost btn-xs btn-square h-5 w-5 min-h-0"
                    title="筛选此名称"
                    @click="quickFilter('processName', item.process_name)"
                  >
                    <Icon icon="mdi:filter-variant" class="w-3 h-3 text-base-content/50" />
                  </button>
                </div>
              </td>
              <td class="font-mono text-xs">
                <div class="flex items-center gap-1 group">
                  {{ item.local_addr }}
                  <button
                    class="opacity-0 group-hover:opacity-100 transition-opacity btn btn-ghost btn-xs btn-square h-5 w-5 min-h-0"
                    title="筛选此端口"
                    @click="quickFilter('port', item.local_addr)"
                  >
                    <Icon icon="mdi:filter-variant" class="w-3 h-3 text-base-content/50" />
                  </button>
                </div>
              </td>
              <td class="font-mono text-xs text-base-content/70">{{ item.remote_addr }}</td>
              <td>
                <span class="text-xs font-medium" :class="statusColor(item.status)">{{ item.status }}</span>
              </td>
            </tr>
          </tbody>
          <tbody v-if="!loading && filteredList.length === 0">
            <tr>
              <td colspan="6" class="text-center py-12 text-base-content/40">暂无网络连接数据</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>

<style scoped>
.table :where(thead, tbody) :where(tr:not(:last-child)), .table :where(thead, tbody) :where(tr:first-child:last-child) {
  border-bottom-color: transparent;
}
tbody tr {
  border-bottom: 1px solid var(--fallback-b2,oklch(var(--b2)/var(--tw-border-opacity))) !important;
}
.dropdown-content {
  max-height: 300px;
  overflow-y: auto;
}
</style>
