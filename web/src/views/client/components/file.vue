<script setup lang="ts">
import { ref, computed } from 'vue'

// --- Types ---
interface FileItem {
  id: string
  name: string
  type: 'file' | 'dir'
  permissions: string
  user: string
  group: string
  size: number | null // null for directory initially
  updatedAt: string
}

// --- State ---
const currentPath = ref('/')
const diskUsage = '158.2 GB'
const totalItems = ref(27)
const pageSize = ref(100)
const currentPage = ref(1)
const searchQuery = ref('')
const searchInSubdir = ref(false)

// Mock Data
const fileList = ref<FileItem[]>([
  { id: '1', name: '.1panel_clash', type: 'dir', permissions: '0755', user: 'root (0)', group: 'root (0)', size: null, updatedAt: '2025-06-20 17:28:40' },
  { id: '2', name: '1panel-agent.service', type: 'dir', permissions: '0755', user: 'root (0)', group: 'root (0)', size: null, updatedAt: '2025-12-04 13:59:07' },
  { id: '3', name: '1panel-core.service', type: 'dir', permissions: '0755', user: 'root (0)', group: 'root (0)', size: null, updatedAt: '2025-12-04 13:59:07' },
  { id: '4', name: 'boot', type: 'dir', permissions: '0755', user: 'root (0)', group: 'root (0)', size: null, updatedAt: '2026-01-15 06:25:48' },
  { id: '5', name: 'dev', type: 'dir', permissions: '0755', user: 'root (0)', group: 'root (0)', size: null, updatedAt: '2025-10-29 12:30:01' },
  { id: '6', name: 'etc', type: 'dir', permissions: '0755', user: 'root (0)', group: 'root (0)', size: null, updatedAt: '2026-01-29 06:26:24' },
  { id: '7', name: 'home', type: 'dir', permissions: '0755', user: 'root (0)', group: 'root (0)', size: null, updatedAt: '2024-08-10 15:02:09' },
  { id: '8', name: 'lost+found', type: 'dir', permissions: '0700', user: 'root (0)', group: 'root (0)', size: null, updatedAt: '2024-01-30 09:55:10' },
  { id: '9', name: 'media', type: 'dir', permissions: '0755', user: 'root (0)', group: 'root (0)', size: null, updatedAt: '2022-04-21 08:57:51' },
  { id: '10', name: 'mnt', type: 'dir', permissions: '0755', user: 'root (0)', group: 'root (0)', size: null, updatedAt: '2022-04-21 08:57:51' },
  { id: '11', name: 'opt', type: 'dir', permissions: '0755', user: 'root (0)', group: 'root (0)', size: null, updatedAt: '2025-06-20 17:28:29' },
])

const selectedItems = ref<string[]>([])
const allSelected = computed({
  get: () => fileList.value.length > 0 && selectedItems.value.length === fileList.value.length,
  set: (val) => {
    if (val) {
      selectedItems.value = fileList.value.map(item => item.id)
    } else {
      selectedItems.value = []
    }
  }
})

// --- Methods ---
const handleRefresh = () => {
  console.log('Refresh')
}

const toggleSelection = (id: string) => {
  const index = selectedItems.value.indexOf(id)
  if (index === -1) {
    selectedItems.value.push(id)
  } else {
    selectedItems.value.splice(index, 1)
  }
}

const formatSize = (size: number | null) => {
  if (size === null) return '计算'
  return size + ' B' // Simplified
}

</script>

<template>
  <div class="h-full flex flex-col bg-base-100 gap-4">
    <!-- Top Navigation Bar -->
    <div class="flex items-center gap-2 bg-base-100 p-2 rounded-lg border border-base-200 shadow-sm">
      <div class="flex gap-1">
        <!-- <button class="btn btn-circle btn-ghost btn-sm">
          <Icon icon="mdi:arrow-left" class="w-5 h-5 text-base-content/70" />
        </button>
        <button class="btn btn-circle btn-ghost btn-sm">
          <Icon icon="mdi:arrow-right" class="w-5 h-5 text-base-content/70" />
        </button> -->
        <button class="btn btn-circle btn-ghost btn-sm">
          <Icon icon="mdi:arrow-up" class="w-5 h-5 text-base-content/70" />
        </button>
        <button class="btn btn-circle btn-ghost btn-sm" @click="handleRefresh">
          <Icon icon="mdi:refresh" class="w-5 h-5 text-base-content/70" />
        </button>
        <button class="btn btn-circle btn-ghost btn-sm">
          <Icon icon="mdi:eye-outline" class="w-5 h-5 text-base-content/70" />
        </button>
      </div>

      <div class="divider divider-horizontal m-0"></div>

      <!-- Breadcrumb / Path -->
      <div class="flex-1 flex items-center bg-base-200/50 rounded-md px-3 py-1.5 text-sm gap-2">
        <Icon icon="mdi:home" class="w-5 h-5 text-base-content/70" />
        <span class="font-medium">/ (根目录) {{ diskUsage }}</span>
      </div>
    </div>

    <!-- Toolbar -->
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div class="flex items-center gap-2 flex-wrap">
        <!-- Create Button -->
        <div class="dropdown">
          <div tabindex="0" role="button" class="btn btn-primary btn-sm gap-1">
            创建
            <Icon icon="mdi:chevron-down" class="w-4 h-4" />
          </div>
          <ul tabindex="0" class="dropdown-content z-[1] menu p-2 shadow bg-base-100 rounded-box w-52">
            <li><a>文件夹</a></li>
            <li><a>文件</a></li>
          </ul>
        </div>

        <!-- Upload -->
        <button class="btn btn-primary btn-sm">
          上传文件
        </button>

        <!-- File Operations -->
        <button class="btn btn-primary btn-sm" :disabled="selectedItems.length === 0">下载</button>
        <button class="btn btn-primary btn-sm" :disabled="selectedItems.length === 0">复制</button>
        <button class="btn btn-primary btn-sm" :disabled="selectedItems.length === 0">移动</button>
        <button class="btn btn-primary btn-sm text-error" :disabled="selectedItems.length === 0">删除</button>
        <button class="btn btn-primary btn-sm" :disabled="selectedItems.length === 0">权限</button>

        <!-- <div class="text-sm text-base-content/60 ml-2">
          / (根目录) {{ diskUsage }}
        </div> -->
      </div>

      <div class="flex items-center gap-2">
        <div class="join">
          <input
            type="text"
            v-model="searchQuery"
            placeholder="在当前目录下查找"
            class="input input-sm input-bordered join-item w-48"
          />
          <button class="btn btn-sm btn-square join-item">
            <Icon icon="mdi:magnify" class="w-5 h-5" />
          </button>
        </div>
      </div>
    </div>

    <!-- Notice Bar -->
    <div class="alert alert-info bg-blue-50 text-blue-800 border-none py-2 rounded-none text-xs flex items-center gap-2">
       <span>注意：1. 搜索结果不支持排序功能 2. 文件夹无法按大小排序。</span>
    </div>

    <!-- File Table -->
    <div class="flex-1 overflow-hidden flex flex-col bg-base-100 rounded-lg border border-base-200 shadow-sm">
      <div class="overflow-x-auto flex-1">
        <table class="table table-sm table-pin-rows">
          <thead>
            <tr class="bg-base-100 border-b border-base-200">
              <th class="w-10">
                <label>
                  <input type="checkbox" class="checkbox checkbox-xs" v-model="allSelected" />
                </label>
              </th>
              <th class="cursor-pointer hover:bg-base-200">
                <div class="flex items-center gap-1">
                  名称
                  <Icon icon="mdi:unfold-more-horizontal" class="w-4 h-4 text-base-content/40" />
                </div>
              </th>
              <th>权限</th>
              <th>用户</th>
              <th>用户组</th>
              <th class="cursor-pointer hover:bg-base-200">
                <div class="flex items-center gap-1">
                  大小
                  <Icon icon="mdi:unfold-more-horizontal" class="w-4 h-4 text-base-content/40" />
                </div>
              </th>
              <th class="cursor-pointer hover:bg-base-200">
                <div class="flex items-center gap-1">
                  修改时间
                  <Icon icon="mdi:unfold-more-horizontal" class="w-4 h-4 text-base-content/40" />
                </div>
              </th>
              <th class="text-right pr-6">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="file in fileList" :key="file.id" class="hover group">
              <td>
                <label>
                  <input
                    type="checkbox"
                    class="checkbox checkbox-xs"
                    :checked="selectedItems.includes(file.id)"
                    @change="toggleSelection(file.id)"
                  />
                </label>
              </td>
              <td>
                <div class="flex items-center gap-2">
                  <Icon
                    v-if="file.type === 'dir'"
                    icon="mdi:folder"
                    class="w-5 h-5 text-yellow-400"
                  />
                  <Icon
                    v-else
                    icon="mdi:file-document-outline"
                    class="w-5 h-5 text-base-content/60"
                  />
                  <span class="font-medium text-blue-600 cursor-pointer hover:underline">{{ file.name }}</span>
                </div>
              </td>
              <td class="font-mono text-xs">{{ file.permissions }}</td>
              <td>{{ file.user }}</td>
              <td>{{ file.group }}</td>
              <td>
                <span v-if="file.type === 'dir'" class="text-blue-600 cursor-pointer hover:underline text-xs">计算</span>
                <span v-else>{{ formatSize(file.size) }}</span>
              </td>
              <td class="text-base-content/70 text-xs">{{ file.updatedAt }}</td>
              <td class="text-right">
                <div class="flex items-center justify-end gap-2 opacity-0 group-hover:opacity-100 transition-opacity">
                  <button class="text-blue-600 hover:text-blue-800 text-xs font-medium">打开</button>
                  <button class="text-blue-600 hover:text-blue-800 text-xs font-medium">下载</button>
                  <button class="text-blue-600 hover:text-blue-800 text-xs font-medium">更多</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Footer -->
      <div class="border-t border-base-200 p-2 flex flex-col sm:flex-row items-center justify-between text-xs text-base-content/70 gap-2">
        <div class="flex items-center gap-1">
          <span>共 {{ totalItems }} 个目录, 0 个文件, 当前目录大小</span>
          <span class="text-blue-600 cursor-pointer hover:underline">计算</span>
        </div>

        <div class="flex items-center gap-2">
          <span>共 {{ totalItems }} 条</span>
          <select v-model="pageSize" class="select select-bordered select-xs h-8">
            <option :value="100">100条/页</option>
            <option :value="50">50条/页</option>
            <option :value="20">20条/页</option>
          </select>

          <div class="join">
            <button class="join-item btn btn-xs h-8" :disabled="currentPage === 1">
              <Icon icon="mdi:chevron-left" class="w-4 h-4" />
            </button>
            <button class="join-item btn btn-xs btn-active h-8">{{ currentPage }}</button>
            <button class="join-item btn btn-xs h-8">
              <Icon icon="mdi:chevron-right" class="w-4 h-4" />
            </button>
          </div>

          <div class="flex items-center gap-1">
            <span>前往</span>
            <input type="number" class="input input-bordered input-xs w-12 h-8 text-center" value="1" />
            <span>页</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
/* Custom styles if needed to tweak daisyUI defaults */
:deep(.table-pin-rows th) {
  z-index: 10;
}
</style>
