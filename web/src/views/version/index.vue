<script setup lang="ts">
import { getVersionPage, uploadVersion, deleteVersion } from "@/api/version";
import { usePageList } from "@/composables/usePageList";
import { useUIStore } from "@/stores/ui";
import { formatBytes, formatDateTime } from "@/utils";
import { clientOsIconMap, osOptions, archOptions } from "@/map/client"

const ui = useUIStore();

const uploadModal = ref<HTMLDialogElement>();
const uploadForm = ref({
  version: "",
  version_name: "",
  os: "linux",
  arch: "amd64",
  changelog: "",
});
const uploadFile = ref<File>();
const uploading = ref(false);

const { data, loading, currentPage, pageSize, search, changePage } = usePageList(getVersionPage);

const handleSearch = () => search({});

const handlePageChange = ({ page, pageSize: size }: { page: number; pageSize: number }) => {
  changePage(page, size);
};

const handleFileChange = (e: Event) => {
  const target = e.target as HTMLInputElement;
  if (target.files?.length) {
    uploadFile.value = target.files[0];
  }
};

const handleUpload = async () => {
  if (!uploadFile.value) {
    ui.showToast("请选择要上传的二进制文件", "error");
    return;
  }
  if (!uploadForm.value.version || !uploadForm.value.version_name) {
    ui.showToast("请填写版本号和版本名称", "error");
    return;
  }

  uploading.value = true;
  try {
    const fd = new FormData();
    fd.append("file", uploadFile.value);
    fd.append("version", uploadForm.value.version);
    fd.append("version_name", uploadForm.value.version_name);
    fd.append("os", uploadForm.value.os);
    fd.append("arch", uploadForm.value.arch);
    fd.append("changelog", uploadForm.value.changelog);
    await uploadVersion(fd);
    ui.showToast("上传成功", "success");
    uploadModal.value?.close();
    resetForm();
    handleSearch();
  } catch {
    // toast handled by interceptor
  } finally {
    uploading.value = false;
  }
};

const resetForm = () => {
  uploadForm.value = { version: "", version_name: "", os: "linux", arch: "amd64", changelog: "" };
  uploadFile.value = undefined;
};

const deleteId = ref<number | null>(null);
const deleteModal = ref<HTMLDialogElement>();

const confirmDelete = (id: number) => {
  deleteId.value = id;
  deleteModal.value?.showModal();
};

const handleDelete = async () => {
  if (deleteId.value === null) return;
  await deleteVersion(deleteId.value);
  ui.showToast("删除成功", "success");
  deleteModal.value?.close();
  deleteId.value = null;
  handleSearch();
};

const osIconMap: Record<string, string> = clientOsIconMap;
</script>

<template>
  <div class="space-y-4">
    <!-- header -->
    <div class="flex items-center justify-between">
      <h2 class="text-2xl font-bold text-base-content">版本管理</h2>
      <button class="btn btn-primary btn-sm" @click="uploadModal?.showModal()">
        <Icon icon="mdi:upload" class="w-4 h-4" /> 上传版本
      </button>
    </div>

    <!-- 版本列表 -->
    <div class="bg-base-100 rounded-lg shadow-sm border border-base-200 overflow-hidden">
      <div class="overflow-x-auto">
        <table class="table table-zebra table-sm">
          <thead>
            <tr>
              <th>版本号</th>
              <th>版本名称</th>
              <th>平台</th>
              <th>架构</th>
              <th>文件大小</th>
              <th>更新日志</th>
              <th>上传时间</th>
              <th class="text-center">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="8" class="text-center py-8">
                <span class="loading loading-spinner loading-md text-primary" />
              </td>
            </tr>
            <tr v-else-if="!data?.list.length">
              <td colspan="8" class="text-center py-8 text-base-content/50">暂无版本数据</td>
            </tr>
            <tr v-for="item in data?.list" :key="item.id">
              <td class="font-mono text-sm">{{ item.version }}</td>
              <td>{{ item.version_name }}</td>
              <td>
                <div class="flex items-center gap-1">
                  <Icon :icon="osIconMap[item.os] || 'mdi:code-braces'" class="w-4 h-4" />
                  {{ item.os === "darwin" ? "macOS" : item.os }}
                </div>
              </td>
              <td>{{ item.arch }}</td>
              <td>{{ formatBytes(item.binary_size) }}</td>
              <td class="max-w-40 truncate" :title="item.changelog">{{ item.changelog || "-" }}</td>
              <td class="text-xs">{{ formatDateTime(item.created_at) }}</td>
              <td class="text-center">
                <button class="btn btn-ghost btn-xs text-error" @click="confirmDelete(item.id)">
                  <Icon icon="mdi:delete" class="w-4 h-4" /> 删除
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
      <div class="border-t border-base-200 px-4 py-3">
        <Pagination
          v-if="!loading"
          :total="data?.total || 0"
          :current-page="currentPage"
          :page-size="pageSize"
          @change="handlePageChange"
        />
      </div>
    </div>

    <!-- 删除确认模态框 -->
    <dialog ref="deleteModal" class="modal">
      <div class="modal-box">
        <h3 class="font-bold text-lg mb-2">确认删除</h3>
        <p class="text-base-content/70">确定删除此版本？此操作不可恢复。</p>
        <div class="modal-action">
          <button class="btn btn-ghost btn-sm" @click="deleteModal?.close()">取消</button>
          <button class="btn btn-error btn-sm" @click="handleDelete">确认删除</button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop"><button>close</button></form>
    </dialog>

  <!-- 上传版本模态框 -->
    <dialog ref="uploadModal" class="modal">
      <div class="modal-box max-w-lg">
        <form method="dialog">
          <button class="btn btn-sm btn-circle btn-ghost absolute right-2 top-2" @click="resetForm">✕</button>
        </form>
        <h3 class="font-bold text-lg mb-4 flex items-center gap-2">
          <Icon icon="mdi:upload" class="text-primary w-6 h-6" />
          上传新版本
        </h3>
        <div class="space-y-4">
          <div class="form-control">
            <label class="label"><span class="label-text">二进制文件</span></label>
            <input type="file" class="file-input file-input-bordered file-input-sm w-full" @change="handleFileChange" />
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="form-control">
              <label class="label"><span class="label-text">版本号 <span class="text-error">*</span></span></label>
              <input v-model="uploadForm.version" type="number" class="input input-bordered input-sm" placeholder="例如: 1" />
            </div>
            <div class="form-control">
              <label class="label"><span class="label-text">版本名称 <span class="text-error">*</span></span></label>
              <input v-model="uploadForm.version_name" type="text" class="input input-bordered input-sm" placeholder="例如: v1.0.0" />
            </div>
          </div>
          <div class="grid grid-cols-2 gap-4">
            <div class="form-control">
              <label class="label"><span class="label-text">操作系统</span></label>
              <select v-model="uploadForm.os" class="select select-bordered select-sm">
                <option v-for="opt in osOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
              </select>
            </div>
            <div class="form-control">
              <label class="label"><span class="label-text">架构</span></label>
              <select v-model="uploadForm.arch" class="select select-bordered select-sm">
                <option v-for="opt in archOptions" :key="opt.value" :value="opt.value">{{ opt.label }}</option>
              </select>
            </div>
          </div>
          <div class="form-control">
            <label class="label"><span class="label-text">更新日志</span></label>
            <textarea v-model="uploadForm.changelog" class="textarea textarea-bordered h-20 text-sm" placeholder="本次更新的内容..."></textarea>
          </div>
        </div>
        <div class="modal-action">
          <form method="dialog">
            <button class="btn btn-ghost" @click="resetForm">取消</button>
          </form>
          <button class="btn btn-primary" :disabled="uploading" @click="handleUpload">
            <span v-if="uploading" class="loading loading-spinner loading-xs" />
            {{ uploading ? "上传中..." : "上传" }}
          </button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button @click="resetForm">close</button>
      </form>
    </dialog>
  </div>
</template>
