<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { getFileList, downloadFile, uploadFile, fetchFile, deleteFile, createDir, renameFile } from '@/api/ops'
import type { FileEntry } from '@/types'

const props = defineProps<{ id: string }>()

// --- State ---
const currentPath = ref('/')
const fileList = ref<FileEntry[]>([])
const loading = ref(false)
const selectedItems = ref<string[]>([])

const allSelected = computed({
  get: () => fileList.value.length > 0 && selectedItems.value.length === fileList.value.length,
  set: (val) => {
    if (val) selectedItems.value = fileList.value.map(item => item.name)
    else selectedItems.value = []
  }
})
const toggleSelection = (name: string) => {
  const idx = selectedItems.value.indexOf(name)
  if (idx === -1) selectedItems.value.push(name)
  else selectedItems.value.splice(idx, 1)
}

// --- Search ---
const searchQuery = ref('')
const filteredList = computed(() => {
  if (!searchQuery.value) return fileList.value
  const q = searchQuery.value.toLowerCase()
  return fileList.value.filter(f => f.name.toLowerCase().includes(q))
})

// --- Sort ---
type SortKey = 'name' | 'size' | 'mod_time'
const sortKey = ref<SortKey>('name')
const sortAsc = ref(true)
const sortedList = computed(() => {
  const list = [...filteredList.value]
  const dirs = list.filter(f => f.is_dir)
  const files = list.filter(f => !f.is_dir)
  const cmp = (a: FileEntry, b: FileEntry, k: SortKey) => {
    if (k === 'name') return a.name.localeCompare(b.name)
    if (k === 'size') return a.size - b.size
    return a.mod_time - b.mod_time
  }
  const s = sortAsc.value ? 1 : -1
  dirs.sort((a, b) => cmp(a, b, sortKey.value) * s)
  files.sort((a, b) => cmp(a, b, sortKey.value) * s)
  return [...dirs, ...files]
})
function toggleSort(key: SortKey) {
  if (sortKey.value === key) sortAsc.value = !sortAsc.value
  else { sortKey.value = key; sortAsc.value = true }
}
function sortIcon(key: SortKey) {
  if (sortKey.value !== key) return 'mdi:unfold-more-horizontal'
  return sortAsc.value ? 'mdi:arrow-up' : 'mdi:arrow-down'
}

// --- Path ---
const pathSegments = computed(() => {
  const parts = currentPath.value.split('/').filter(Boolean)
  const segs: { name: string; path: string }[] = []
  let acc = ''
  for (const p of parts) { acc += '/' + p; segs.push({ name: p, path: acc }) }
  return segs
})

function navigateTo(path: string) { currentPath.value = path }
function navigateUp() {
  if (currentPath.value === '/') return
  currentPath.value = currentPath.value.substring(0, currentPath.value.lastIndexOf('/')) || '/'
}

// --- Load ---
async function loadFileList() {
  loading.value = true
  try { fileList.value = await getFileList(Number(props.id), currentPath.value) }
  catch { fileList.value = [] }
  finally { loading.value = false }
}
watch(currentPath, loadFileList)
loadFileList()

function buildPath(name: string) {
  return currentPath.value === '/' ? '/' + name : currentPath.value + '/' + name
}

// --- Download ---
function handleDownload(item: FileEntry) {
  if (item.is_dir) return
  downloadFile(Number(props.id), buildPath(item.name))
}
function handleDownloadSelected() {
  for (const name of selectedItems.value) {
    const item = fileList.value.find(f => f.name === name)
    if (item && !item.is_dir) handleDownload(item)
  }
}

// --- Upload (multi + drag) ---
const uploadInput = ref<HTMLInputElement>()
const isDragOver = ref(false)

function triggerUpload() { uploadInput.value?.click() }

async function doUpload(file: File) {
  try {
    await uploadFile(Number(props.id), buildPath(file.name), file)
  } catch { /* toast */ }
}

async function handleFileUpload(e: Event) {
  const input = e.target as HTMLInputElement
  if (!input.files?.length) return
  for (const file of Array.from(input.files)) await doUpload(file)
  input.value = ''
  loadFileList()
}

function onDragOver(e: DragEvent) {
  e.preventDefault()
  isDragOver.value = true
}
function onDragLeave() { isDragOver.value = false }
async function onDrop(e: DragEvent) {
  e.preventDefault()
  isDragOver.value = false
  if (!e.dataTransfer?.files.length) return
  for (const file of Array.from(e.dataTransfer.files)) await doUpload(file)
  loadFileList()
}

// --- Delete ---
const deleteTarget = ref<{ name: string; path: string; is_dir: boolean } | null>(null)
const deleteModal = ref<HTMLDialogElement>()

function handleDelete(item: FileEntry) {
  deleteTarget.value = { name: item.name, path: buildPath(item.name), is_dir: item.is_dir }
  deleteModal.value?.showModal()
}
async function confirmDelete() {
  if (!deleteTarget.value) return
  await deleteFile(Number(props.id), deleteTarget.value.path)
  deleteModal.value?.close()
  deleteTarget.value = null
  loadFileList()
}

// --- Mkdir ---
const mkdirName = ref('')
const mkdirModal = ref<HTMLDialogElement>()

function showMkdir() { mkdirName.value = ''; mkdirModal.value?.showModal() }

async function confirmMkdir() {
  if (!mkdirName.value) return
  await createDir(Number(props.id), buildPath(mkdirName.value))
  mkdirModal.value?.close()
  loadFileList()
}

// --- Rename ---
const renameTarget = ref<FileEntry | null>(null)
const newName = ref('')
const renameModal = ref<HTMLDialogElement>()

function handleRename(item: FileEntry) {
  renameTarget.value = item
  newName.value = item.name
  renameModal.value?.showModal()
}
async function confirmRename() {
  if (!renameTarget.value || !newName.value) return
  await renameFile(Number(props.id), buildPath(renameTarget.value.name), buildPath(newName.value))
  renameModal.value?.close()
  renameTarget.value = null
  loadFileList()
}

// --- Preview ---
const previewItem = ref<FileEntry | null>(null)
const previewContent = ref<string>('')
const previewImageUrl = ref<string>('')
const previewType = ref<'text' | 'image' | 'binary'>('text')
const previewLoading = ref(false)
const previewModal = ref<HTMLDialogElement>()

function closePreview() {
  if (previewImageUrl.value) { window.URL.revokeObjectURL(previewImageUrl.value); previewImageUrl.value = '' }
  previewItem.value = null; previewContent.value = ''
}
async function handlePreview(item: FileEntry) {
  if (item.is_dir) return
  previewItem.value = item; previewLoading.value = true; previewContent.value = ''
  previewModal.value?.showModal()
  try {
    const result = await fetchFile(Number(props.id), buildPath(item.name))
    previewType.value = result.type
    if (result.type === 'text') previewContent.value = result.content
    else if (result.type === 'image') previewImageUrl.value = result.blobUrl
  } catch { previewContent.value = '加载失败'; previewType.value = 'text' }
  finally { previewLoading.value = false }
}
function handleRowClick(item: FileEntry) {
  if (item.is_dir) { currentPath.value = buildPath(item.name) }
  else { handlePreview(item) }
}

// --- File icon ---
function fileIcon(name: string): string {
  const ext = name.split('.').pop()?.toLowerCase() || ''
  const map: Record<string, string> = {
    txt: 'mdi:file-text-outline', md: 'mdi:language-markdown',
    json: 'mdi:code-json', xml: 'mdi:xml', yaml: 'mdi:code-braces', yml: 'mdi:code-braces',
    js: 'mdi:language-javascript', ts: 'mdi:language-typescript',
    go: 'mdi:language-go', py: 'mdi:language-python',
    java: 'mdi:language-java', rs: 'mdi:language-rust',
    sh: 'mdi:terminal', bash: 'mdi:terminal', zsh: 'mdi:terminal',
    conf: 'mdi:file-cog', ini: 'mdi:file-cog', cfg: 'mdi:file-cog',
    log: 'mdi:text-box-search',
    zip: 'mdi:zip-box', tar: 'mdi:archive', gz: 'mdi:archive', bz2: 'mdi:archive',
    png: 'mdi:file-image', jpg: 'mdi:file-image', jpeg: 'mdi:file-image', gif: 'mdi:file-image', webp: 'mdi:file-image',
    pdf: 'mdi:file-pdf-box',
    doc: 'mdi:file-word', docx: 'mdi:file-word',
    xls: 'mdi:file-excel', xlsx: 'mdi:file-excel',
    csv: 'mdi:file-delimited',
    svg: 'mdi:file-image-plus',
  }
  return map[ext] || 'mdi:file-document-outline'
}

// --- Format ---
function formatFileSize(size: number): string {
  if (size === 0) return '0 B'
  const units = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(size) / Math.log(1024))
  return (size / Math.pow(1024, i)).toFixed(i > 0 ? 2 : 0) + ' ' + units[i]
}
function formatTime(unix: number): string {
  if (!unix) return '-'
  return new Date(unix * 1000).toLocaleString()
}
</script>

<template>
  <div class="h-full flex flex-col bg-base-100 gap-4"
    @dragover="onDragOver" @dragleave="onDragLeave" @drop="onDrop">

    <!-- Drop overlay -->
    <div v-if="isDragOver"
      class="absolute inset-0 z-50 bg-primary/10 border-2 border-dashed border-primary rounded-2xl flex items-center justify-center pointer-events-none">
      <div class="text-primary font-bold text-lg flex items-center gap-3">
        <span class="w-8 h-8"><Icon icon="mdi:cloud-upload" /></span> 释放以上传文件
      </div>
    </div>

    <!-- Top Navigation Bar -->
    <div class="flex items-center gap-2 bg-base-100 p-2 rounded-lg border border-base-200 shadow-sm">
      <div class="flex gap-1">
        <button class="btn btn-circle btn-ghost btn-sm" @click="navigateUp" :disabled="currentPath === '/'">
          <span class="w-5 h-5 text-base-content/70"><Icon icon="mdi:arrow-up" /></span>
        </button>
        <button class="btn btn-circle btn-ghost btn-sm" @click="loadFileList">
          <span class="w-5 h-5 text-base-content/70"><Icon icon="mdi:refresh" /></span>
        </button>
      </div>
      <div class="divider divider-horizontal m-0"></div>
      <!-- Breadcrumb -->
      <div class="flex-1 flex items-center bg-base-200/50 rounded-md px-3 py-1.5 text-sm gap-1 overflow-x-auto">
        <span class="w-5 h-5 text-base-content/70 shrink-0"><Icon icon="mdi:home" /></span>
        <button class="font-medium hover:text-primary transition-colors shrink-0" @click="navigateTo('/')">/</button>
        <template v-for="(seg, i) in pathSegments" :key="i">
          <span class="text-base-content/30 shrink-0">/</span>
          <button class="hover:text-primary transition-colors truncate max-w-[10rem] shrink-0"
            :class="i === pathSegments.length - 1 ? 'text-primary font-medium' : ''"
            @click="navigateTo(seg.path)">{{ seg.name }}</button>
        </template>
      </div>
    </div>

    <!-- Toolbar -->
    <div class="flex flex-wrap items-center justify-between gap-3">
      <div class="flex items-center gap-2 flex-wrap">
        <button class="btn btn-primary btn-sm" @click="triggerUpload">
          <span class="w-4 h-4"><Icon icon="mdi:upload" /></span> 上传
        </button>
        <input ref="uploadInput" type="file" class="hidden" multiple @change="handleFileUpload" />
        <button class="btn btn-primary btn-sm" @click="showMkdir">
          <span class="w-4 h-4"><Icon icon="mdi:folder-plus" /></span> 新建目录
        </button>
        <button class="btn btn-primary btn-sm" :disabled="selectedItems.length === 0" @click="handleDownloadSelected">
          <span class="w-4 h-4"><Icon icon="mdi:download" /></span> 下载
        </button>
      </div>
      <div class="flex items-center gap-3">
        <label class="input input-bordered input-sm flex items-center gap-1 w-48">
          <span class="w-4 h-4 text-base-content/40"><Icon icon="mdi:magnify" /></span>
          <input v-model="searchQuery" type="text" class="grow" placeholder="搜索文件名" />
        </label>
        <span class="text-sm text-base-content/60">{{ filteredList.length }} / {{ fileList.length }} 项</span>
        <span v-if="loading" class="loading loading-spinner loading-xs" />
      </div>
    </div>

    <!-- File Table -->
    <div class="flex-1 overflow-hidden flex flex-col bg-base-100 rounded-lg border border-base-200 shadow-sm">
      <div class="overflow-x-auto flex-1">
        <table class="table table-sm table-pin-rows">
          <thead>
            <tr class="bg-base-100 border-b border-base-200">
              <th class="w-10">
                <label><input type="checkbox" class="checkbox checkbox-xs" v-model="allSelected" /></label>
              </th>
              <th class="cursor-pointer hover:bg-base-200 w-16" @click="toggleSort('name')">
                <div class="flex items-center gap-1">名称 <span class="w-3.5 h-3.5"><Icon :icon="sortIcon('name')" /></span></div>
              </th>
              <th class="cursor-pointer hover:bg-base-200 w-20" @click="toggleSort('size')">
                大小 <span class="w-3.5 h-3.5"><Icon :icon="sortIcon('size')" /></span>
              </th>
              <th class="cursor-pointer hover:bg-base-200 w-28" @click="toggleSort('mod_time')">
                修改时间 <span class="w-3.5 h-3.5"><Icon :icon="sortIcon('mod_time')" /></span>
              </th>
              <th class="w-20">用户</th>
              <th class="w-20">用户组</th>
              <th class="text-right w-36">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="file in sortedList" :key="file.name"
              class="hover:bg-base-200 cursor-pointer transition-all duration-150 group"
              @dblclick="handleRowClick(file)">
              <td @click.stop>
                <label><input type="checkbox" class="checkbox checkbox-xs"
                  :checked="selectedItems.includes(file.name)" @change="toggleSelection(file.name)" /></label>
              </td>
              <td>
                <div class="flex items-center gap-2">
                  <span v-if="file.is_dir" class="w-5 h-5 text-yellow-400 shrink-0"><Icon icon="mdi:folder" /></span>
                  <span v-else class="w-5 h-5 text-base-content/60 shrink-0"><Icon :icon="fileIcon(file.name)" /></span>
                  <span class="font-medium cursor-pointer hover:underline truncate max-w-[20rem]"
                    :class="file.is_dir ? 'text-blue-600' : 'text-base-content'"
                    @click="handleRowClick(file)">{{ file.name }}</span>
                </div>
              </td>
              <td class="text-sm">
                <span v-if="file.is_dir" class="text-base-content/40">-</span>
                <span v-else>{{ formatFileSize(file.size) }}</span>
              </td>
              <td class="text-base-content/70 text-xs">{{ formatTime(file.mod_time) }}</td>
              <td class="text-xs text-base-content/60">{{ file.owner || '-' }}</td>
              <td class="text-xs text-base-content/60">{{ file.group || '-' }}</td>
              <td class="text-right" @click.stop>
                <div class="flex items-center justify-end gap-1">
                  <button v-if="file.is_dir" class="btn btn-ghost btn-xs" @click="handleRowClick(file)">打开</button>
                  <template v-else>
                    <button class="btn btn-ghost btn-xs" @click="handlePreview(file)">预览</button>
                    <button class="btn btn-ghost btn-xs" @click="handleDownload(file)">下载</button>
                  </template>
                  <button class="btn btn-ghost btn-xs" @click="handleRename(file)">重命名</button>
                  <button class="btn btn-ghost btn-xs text-error" @click="handleDelete(file)">删除</button>
                </div>
              </td>
            </tr>
          </tbody>
          <tbody v-if="!loading && sortedList.length === 0">
            <tr>
              <td colspan="7" class="text-center py-12 text-base-content/40">
                <template v-if="searchQuery">未找到匹配的文件</template>
                <template v-else>空目录</template>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>

  <!-- Delete Modal -->
  <dialog ref="deleteModal" class="modal">
    <div class="modal-box">
      <h3 class="font-bold text-lg flex items-center gap-2">
        <span class="w-6 h-6 text-error"><Icon icon="mdi:alert-circle-outline" /></span>
        确认删除
      </h3>
      <p class="py-4 text-base-content/70">
        确定删除{{ deleteTarget?.is_dir ? '目录' : '文件' }}
        <span class="font-bold text-base-content">"{{ deleteTarget?.name }}"</span>？
        <span v-if="deleteTarget?.is_dir" class="block mt-1 text-warning">目录内的所有内容将被一并删除。</span>
      </p>
      <div class="modal-action">
        <form method="dialog"><button class="btn btn-ghost" @click="deleteTarget = null">取消</button></form>
        <button class="btn btn-error" @click="confirmDelete"><span class="w-4 h-4"><Icon icon="mdi:delete" /></span> 删除</button>
      </div>
    </div>
    <form method="dialog" class="modal-backdrop"><button @click="deleteTarget = null">close</button></form>
  </dialog>

  <!-- Mkdir Modal -->
  <dialog ref="mkdirModal" class="modal">
    <div class="modal-box">
      <h3 class="font-bold text-lg flex items-center gap-2">
        <span class="w-5 h-5 text-primary"><Icon icon="mdi:folder-plus" /></span>
        新建目录
      </h3>
      <div class="py-4">
        <label class="form-control w-full">
          <span class="label-text">目录名称</span>
          <input v-model="mkdirName" type="text" class="input input-bordered w-full mt-2"
            placeholder="输入目录名" @keyup.enter="confirmMkdir" />
        </label>
      </div>
      <div class="modal-action">
        <form method="dialog"><button class="btn btn-ghost">取消</button></form>
        <button class="btn btn-primary" :disabled="!mkdirName" @click="confirmMkdir">创建</button>
      </div>
    </div>
    <form method="dialog" class="modal-backdrop"><button>close</button></form>
  </dialog>

  <!-- Rename Modal -->
  <dialog ref="renameModal" class="modal">
    <div class="modal-box">
      <h3 class="font-bold text-lg flex items-center gap-2">
        <span class="w-5 h-5 text-primary"><Icon icon="mdi:rename" /></span>
        重命名
      </h3>
      <div class="py-4">
        <label class="form-control w-full">
          <span class="label-text">新名称</span>
          <input v-model="newName" type="text" class="input input-bordered w-full mt-2"
            placeholder="输入新名称" @keyup.enter="confirmRename" />
        </label>
      </div>
      <div class="modal-action">
        <form method="dialog"><button class="btn btn-ghost" @click="renameTarget = null">取消</button></form>
        <button class="btn btn-primary" :disabled="!newName" @click="confirmRename">确认</button>
      </div>
    </div>
    <form method="dialog" class="modal-backdrop"><button @click="renameTarget = null">close</button></form>
  </dialog>

  <!-- Preview Modal -->
  <dialog ref="previewModal" class="modal" @close="closePreview">
    <div class="modal-box max-w-4xl max-h-[85vh] flex flex-col">
      <form method="dialog"><button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2" @click="closePreview">✕</button></form>
      <h3 class="font-bold text-lg mb-3 truncate pr-8">
        <span class="w-5 h-5 inline align-text-bottom mr-1"><Icon icon="mdi:file-eye-outline" /></span>
        {{ previewItem?.name }}
      </h3>
      <div v-if="previewLoading" class="flex-1 flex items-center justify-center py-16">
        <span class="loading loading-spinner loading-lg text-primary" />
      </div>
      <template v-else-if="previewType === 'image' && previewImageUrl">
        <div class="flex-1 overflow-auto flex items-center justify-center bg-base-200/30 rounded-lg p-4 min-h-[300px]">
          <img :src="previewImageUrl" :alt="previewItem?.name" class="max-w-full max-h-[70vh] object-contain shadow-sm rounded" />
        </div>
      </template>
      <template v-else-if="previewType === 'text'">
        <pre class="flex-1 overflow-auto bg-base-200/50 rounded-lg p-4 text-sm font-mono leading-relaxed whitespace-pre-wrap break-all border border-base-200 min-h-[300px] max-h-[70vh]"><code>{{ previewContent }}</code></pre>
      </template>
      <template v-else-if="previewType === 'binary'">
        <div class="flex-1 flex flex-col items-center justify-center gap-4 py-16 text-base-content/50">
          <span class="w-16 h-16"><Icon icon="mdi:file-question-outline" /></span>
          <span>该文件类型暂不支持预览</span>
          <button class="btn btn-primary btn-sm" @click="previewItem && handleDownload(previewItem)">下载查看</button>
        </div>
      </template>
      <div class="modal-action"><form method="dialog"><button class="btn" @click="closePreview">关闭</button></form></div>
    </div>
    <form method="dialog" class="modal-backdrop"><button @click="closePreview">close</button></form>
  </dialog>
</template>

<style scoped>
:deep(.table-pin-rows th) { z-index: 10; }
</style>
