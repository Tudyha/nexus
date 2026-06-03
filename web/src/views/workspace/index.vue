<script setup lang="ts">
import { getWorkspaces, createWorkspace, updateWorkspace, deleteWorkspace } from "@/api/workspace";
import { useUIStore } from "@/stores/ui";
import { formatDateTime } from "@/utils";
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";

const ui = useUIStore();
const router = useRouter();

const loading = ref(false);
const list = ref<any[]>([]);
const modal = ref<HTMLDialogElement>();
const editMode = ref(false);
const editingId = ref<number | null>(null);
const form = ref({ name: "", description: "" });

const fetchList = async () => {
  loading.value = true;
  try {
    list.value = await getWorkspaces();
  } finally {
    loading.value = false;
  }
};

const openCreate = () => {
  editMode.value = false;
  editingId.value = null;
  form.value = { name: "", description: "" };
  modal.value?.showModal();
};

const openEdit = (item: any) => {
  editMode.value = true;
  editingId.value = item.id;
  form.value = { name: item.name, description: item.description || "" };
  modal.value?.showModal();
};

const handleSave = async () => {
  if (!form.value.name.trim()) {
    ui.showToast("请输入工作空间名称", "error");
    return;
  }
  try {
    if (editMode.value && editingId.value) {
      await updateWorkspace(editingId.value, form.value);
      ui.showToast("更新成功", "success");
    } else {
      await createWorkspace(form.value);
      ui.showToast("创建成功", "success");
    }
    modal.value?.close();
    await fetchList();
  } catch { /* toast handled by interceptor */ }
};

const deleteTarget = ref<{ id: number; name: string } | null>(null);
const deleteModal = ref<HTMLDialogElement>();

const confirmDelete = (id: number, name: string) => {
  deleteTarget.value = { id, name };
  deleteModal.value?.showModal();
};

const handleDelete = async () => {
  if (!deleteTarget.value) return;
  await deleteWorkspace(deleteTarget.value.id);
  ui.showToast("删除成功", "success");
  deleteModal.value?.close();
  deleteTarget.value = null;
  await fetchList();
};

const viewDetail = (id: number) => {
  router.push({ name: "WorkspaceDetail", params: { id } });
};

onMounted(fetchList);
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-2xl font-bold text-base-content">工作空间管理</h2>
      <button class="btn btn-primary btn-sm" @click="openCreate">
        <Icon icon="mdi:plus" class="w-4 h-4" /> 新建工作空间
      </button>
    </div>

    <div class="bg-base-100 rounded-lg shadow-sm border border-base-200 overflow-hidden">
      <div class="overflow-x-auto">
        <table class="table table-zebra table-sm">
          <thead>
            <tr>
              <th>ID</th>
              <th>名称</th>
              <th>描述</th>
              <th>状态</th>
              <th>创建时间</th>
              <th class="text-center">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="6" class="text-center py-8">
                <span class="loading loading-spinner loading-md text-primary" />
              </td>
            </tr>
            <tr v-else-if="!list.length">
              <td colspan="6" class="text-center py-8 text-base-content/50">暂无工作空间</td>
            </tr>
            <tr v-for="item in list" :key="item.id">
              <td class="font-mono text-xs">{{ item.id }}</td>
              <td>
                <a class="link link-primary" @click="viewDetail(item.id)">{{ item.name }}</a>
              </td>
              <td class="text-sm text-base-content/70">{{ item.description || "-" }}</td>
              <td>
                <span :class="item.status === 1 ? 'badge badge-success badge-sm' : 'badge badge-ghost badge-sm'">
                  {{ item.status === 1 ? "启用" : "禁用" }}
                </span>
              </td>
              <td class="text-xs">{{ formatDateTime(item.created_at) }}</td>
              <td class="text-center">
                <div class="flex items-center justify-center gap-1">
                  <button class="btn btn-ghost btn-xs" @click="viewDetail(item.id)">
                    <Icon icon="mdi:eye" class="w-3.5 h-3.5" />
                  </button>
                  <button class="btn btn-ghost btn-xs" @click="openEdit(item)">
                    <Icon icon="mdi:pencil" class="w-3.5 h-3.5" />
                  </button>
                  <button class="btn btn-ghost btn-xs text-error" @click="confirmDelete(item.id, item.name)">
                    <Icon icon="mdi:delete" class="w-3.5 h-3.5" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Delete Confirm Modal -->
    <dialog ref="deleteModal" class="modal">
      <div class="modal-box">
        <h3 class="font-bold text-lg mb-2">确认删除</h3>
        <p v-if="deleteTarget" class="text-base-content/70">确定删除工作空间「{{ deleteTarget.name }}」？此操作不可恢复。</p>
        <div class="modal-action">
          <button class="btn btn-ghost btn-sm" @click="deleteModal?.close()">取消</button>
          <button class="btn btn-error btn-sm" @click="handleDelete">确认删除</button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop"><button>close</button></form>
    </dialog>

    <!-- Create/Edit Modal -->
    <dialog ref="modal" class="modal">
      <div class="modal-box">
        <h3 class="font-bold text-lg mb-4">{{ editMode ? '编辑工作空间' : '新建工作空间' }}</h3>
        <div class="space-y-3">
          <label class="form-control w-full">
            <span class="label-text">名称</span>
            <input v-model="form.name" type="text" class="input input-bordered input-sm w-full" placeholder="工作空间名称" maxlength="64" />
          </label>
          <label class="form-control w-full">
            <span class="label-text">描述</span>
            <textarea v-model="form.description" class="textarea textarea-bordered textarea-sm w-full" placeholder="可选描述" maxlength="255"></textarea>
          </label>
        </div>
        <div class="modal-action">
          <button class="btn btn-ghost btn-sm" @click="modal?.close()">取消</button>
          <button class="btn btn-primary btn-sm" @click="handleSave">保存</button>
        </div>
      </div>
      <form method="dialog" class="modal-backdrop"><button>close</button></form>
    </dialog>
  </div>
</template>
