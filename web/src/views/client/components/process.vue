<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { Icon } from '@iconify/vue'
import { getProcessList, killProcess } from '@/api/ops'
import type { ProcessEntry } from '@/types'

const props = defineProps<{ id: string }>()

// --- State ---
const processList = ref<ProcessEntry[]>([])
const loading = ref(false)
const filterStatus = ref('')
const filterPid = ref('')
const filterName = ref('')
const filterUser = ref('')
const killing = ref<number | null>(null)

const statusOptions = ['运行', '睡眠', '空闲', '停止', '僵死']

// --- Load ---
async function load() {
  loading.value = true
  try {
    processList.value = await getProcessList(Number(props.id))
  } catch {
    processList.value = []
  } finally {
    loading.value = false
  }
}

onMounted(load)

// --- Kill ---
const killTarget = ref<number | null>(null)
const killModal = ref<HTMLDialogElement>()

function handleKill(pid: number) {
  killTarget.value = pid
  killModal.value?.showModal()
}

async function confirmKill() {
  const pid = killTarget.value
  if (!pid) return
  killing.value = pid
  try {
    await killProcess(Number(props.id), pid)
    processList.value = processList.value.filter(p => p.pid !== pid)
    killModal.value?.close()
    killTarget.value = null
  } finally {
    killing.value = null
  }
}

// --- Helpers ---
const fmtNum = (val: number | undefined | null, fallback = 0) => val ?? fallback

const formatMemory = (bytes: number | undefined | null) => {
  const b = fmtNum(bytes)
  if (b === 0) return '0B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(b) / Math.log(k))
  return parseFloat((b / Math.pow(k, i)).toFixed(2)) + sizes[i]
}

const formatCpu = (val: number | undefined | null) => fmtNum(val).toFixed(2) + '%'

const statusColor = (status: string) => {
  if (status === '运行') return 'text-success'
  if (status === '睡眠') return 'text-info'
  if (status === '空闲') return 'text-base-content/50'
  if (status === '停止' || status === '僵死') return 'text-error'
  return 'text-base-content/70'
}

// --- Filters ---
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
    <div class="flex justify-start gap-2 flex-wrap items-center">
      <select v-model="filterStatus" class="select select-bordered select-sm w-32">
        <option value="">全部状态</option>
        <option v-for="s in statusOptions" :key="s" :value="s">{{ s }}</option>
      </select>

      <div class="relative">
        <input type="text" v-model="filterPid" placeholder="进程ID" class="input input-bordered input-sm w-32 pr-8" />
        <Icon icon="mdi:magnify" class="w-4 h-4 absolute right-2 top-1/2 -translate-y-1/2 text-base-content/40" />
      </div>

      <div class="relative">
        <input type="text" v-model="filterName" placeholder="名称" class="input input-bordered input-sm w-48 pr-8" />
        <Icon icon="mdi:magnify" class="w-4 h-4 absolute right-2 top-1/2 -translate-y-1/2 text-base-content/40" />
      </div>

      <div class="relative">
        <input type="text" v-model="filterUser" placeholder="用户" class="input input-bordered input-sm w-32 pr-8" />
        <Icon icon="mdi:magnify" class="w-4 h-4 absolute right-2 top-1/2 -translate-y-1/2 text-base-content/40" />
      </div>

      <button class="btn btn-ghost btn-sm ml-auto" @click="load" :disabled="loading">
        <Icon icon="mdi:refresh" class="w-4 h-4" />
        {{ loading ? '刷新中...' : '刷新' }}
      </button>
    </div>

    <!-- Process Table -->
    <div class="flex-1 overflow-hidden flex flex-col bg-base-100 rounded-lg border-t border-base-200">
      <div class="overflow-x-auto h-full">
        <table class="table table-sm table-pin-rows">
          <thead>
            <tr class="border-b border-base-200">
              <th class="bg-base-100 text-base-content/60 font-medium">PID</th>
              <th class="bg-base-100 text-base-content/60 font-medium">名称</th>
              <th class="bg-base-100 text-base-content/60 font-medium">用户</th>
              <th class="bg-base-100 text-base-content/60 font-medium">CPU</th>
              <th class="bg-base-100 text-base-content/60 font-medium">内存</th>
              <th class="bg-base-100 text-base-content/60 font-medium">线程</th>
              <th class="bg-base-100 text-base-content/60 font-medium">状态</th>
              <th class="bg-base-100 text-base-content/60 font-medium">启动时间</th>
              <th class="bg-base-100 text-base-content/60 font-medium">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="proc in filteredList" :key="proc.pid" class="hover border-b border-base-100">
              <td class="font-mono">{{ proc.pid }}</td>
              <td>{{ proc.name }}</td>
              <td>{{ proc.user }}</td>
              <td>{{ formatCpu(proc.cpu_percent) }}</td>
              <td>{{ formatMemory(proc.mem_rss) }} ({{ fmtNum(proc.mem_percent).toFixed(1) }}%)</td>
              <td>{{ fmtNum(proc.threads) }}</td>
              <td :class="statusColor(proc.status)">{{ proc.status }}</td>
              <td class="text-base-content/70 text-xs">{{ proc.start_time }}</td>
              <td>
                <button
                  class="text-error hover:text-red-700 text-xs font-medium disabled:opacity-50"
                  :disabled="killing === proc.pid"
                  @click="handleKill(proc.pid)"
                >
                  {{ killing === proc.pid ? '结束中...' : '结束' }}
                </button>
              </td>
            </tr>
          </tbody>
          <tbody v-if="!loading && filteredList.length === 0">
            <tr>
              <td colspan="9" class="text-center py-12 text-base-content/40">暂无进程数据</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>

  <!-- Kill Process Confirm Modal -->
  <dialog ref="killModal" class="modal">
    <div class="modal-box">
      <h3 class="font-bold text-lg flex items-center gap-2">
        <Icon icon="mdi:alert-circle-outline" class="w-6 h-6 text-error" />
        确认结束进程
      </h3>
      <p class="py-4 text-base-content/70">
        确定结束进程 <span class="font-bold text-base-content">PID {{ killTarget }}</span>？
      </p>
      <div class="modal-action">
        <form method="dialog">
          <button class="btn btn-ghost" @click="killTarget = null">取消</button>
        </form>
        <button class="btn btn-error" :disabled="killing === killTarget" @click="confirmKill">
          <Icon icon="mdi:close-circle" class="w-4 h-4" />
          {{ killing === killTarget ? '结束中...' : '结束进程' }}
        </button>
      </div>
    </div>
    <form method="dialog" class="modal-backdrop">
      <button @click="killTarget = null">close</button>
    </form>
  </dialog>
</template>

<style scoped>
.table :where(thead, tbody) :where(tr:not(:last-child)), .table :where(thead, tbody) :where(tr:first-child:last-child) {
  border-bottom-color: transparent;
}
tbody tr {
  border-bottom: 1px solid var(--fallback-b2,oklch(var(--b2)/var(--tw-border-opacity))) !important;
}
</style>
