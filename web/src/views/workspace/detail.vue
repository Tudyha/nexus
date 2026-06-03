<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  getWorkspace,
  getApps,
  createApp,
  updateApp,
  deleteApp,
  updateAppConfig,
  getWorkspaceUsers,
  addWorkspaceUser,
  removeWorkspaceUser,
} from "@/api/workspace";
import { useUIStore } from "@/stores/ui";
import { formatDateTime } from "@/utils";
import type { Workspace, WorkspaceUser, App } from "@/types";

const route = useRoute();
const router = useRouter();
const ui = useUIStore();
const id = Number(route.params.id);

const workspace = ref<Workspace | null>(null);
const apps = ref<App[]>([]);
const users = ref<WorkspaceUser[]>([]);
const loading = ref(false);
const appLoading = ref(false);
const userLoading = ref(false);

// App modal
const appModal = ref<HTMLDialogElement>();
const appEditMode = ref(false);
const editingAppId = ref<number | null>(null);
const appForm = ref({ name: "", description: "" });

// App config modal
const configModal = ref<HTMLDialogElement>();
const configApp = ref<App | null>(null);
const configForm = ref("");

// Confirmation modals
const confirmModal = ref<HTMLDialogElement>();
const confirmAction = ref<() => Promise<void>>();
const confirmMessage = ref("");

const showConfirm = (msg: string, action: () => Promise<void>) => {
  confirmMessage.value = msg;
  confirmAction.value = action;
  confirmModal.value?.showModal();
};

// Add user modal
const userModal = ref<HTMLDialogElement>();
const addUserForm = ref({ user_id: 0, role: 2 });

// Role edit modal
const roleModal = ref<HTMLDialogElement>();
const editingUser = ref<{ userId: number; currentRole: number } | null>(null);
const editRoleForm = ref({ role: 2 });

const openEditRole = (u: WorkspaceUser) => {
  editingUser.value = { userId: u.user_id, currentRole: u.role };
  editRoleForm.value = { role: u.role };
  roleModal.value?.showModal();
};

const handleUpdateRole = async () => {
  if (!editingUser.value) return;
  try {
    const { default: http } = await import("@/api/index");
    await http.put(`/v1/workspace/${id}/users/${editingUser.value.userId}/role`, { role: editRoleForm.value.role });
    ui.showToast("角色更新成功", "success");
    roleModal.value?.close();
    await fetchUsers();
  } catch { /* toast handled by interceptor */ }
};

const fetchWorkspace = async () => {
  try {
    workspace.value = await getWorkspace(id);
  } catch { /* toast handled by interceptor */ }
};

const fetchApps = async () => {
  appLoading.value = true;
  try {
    apps.value = await getApps(id);
  } finally {
    appLoading.value = false;
  }
};

const fetchUsers = async () => {
  userLoading.value = true;
  try {
    users.value = await getWorkspaceUsers(id);
  } finally {
    userLoading.value = false;
  }
};

// App CRUD
const openCreateApp = () => {
  appEditMode.value = false;
  editingAppId.value = null;
  appForm.value = { name: "", description: "" };
  appModal.value?.showModal();
};

const openEditApp = (app: App) => {
  appEditMode.value = true;
  editingAppId.value = app.id;
  appForm.value = { name: app.name, description: app.description || "" };
  appModal.value?.showModal();
};

const handleSaveApp = async () => {
  if (!appForm.value.name.trim()) {
    ui.showToast("请输入应用名称", "error");
    return;
  }
  try {
    if (appEditMode.value && editingAppId.value) {
      await updateApp(editingAppId.value, appForm.value);
      ui.showToast("更新成功", "success");
    } else {
      await createApp(id, appForm.value);
      ui.showToast("创建成功", "success");
    }
    appModal.value?.close();
    await fetchApps();
  } catch { /* toast handled by interceptor */ }
};

const handleDeleteApp = (appId: number, name: string) => {
  showConfirm(`确定删除应用「${name}」？此操作不可恢复。`, async () => {
    await deleteApp(appId);
    ui.showToast("删除成功", "success");
    confirmModal.value?.close();
    await fetchApps();
  });
};

// App config
const openConfig = (app: App) => {
  configApp.value = app;
  configForm.value = app.config || "";
  configModal.value?.showModal();
};

const handleSaveConfig = async () => {
  if (!configApp.value) return;
  try {
    await updateAppConfig(configApp.value.id, { config: configForm.value });
    ui.showToast("配置更新成功", "success");
    configModal.value?.close();
    await fetchApps();
  } catch { /* toast handled by interceptor */ }
};

// User management
const openAddUser = () => {
  addUserForm.value = { user_id: 0, role: 1 };
  userModal.value?.showModal();
};

const handleAddUser = async () => {
  if (!addUserForm.value.user_id) {
    ui.showToast("请输入用户ID", "error");
    return;
  }
  try {
    await addWorkspaceUser(id, addUserForm.value);
    ui.showToast("添加成功", "success");
    userModal.value?.close();
    await fetchUsers();
  } catch { /* toast handled by interceptor */ }
};

const handleRemoveUser = (userId: number) => {
  showConfirm("确定将该用户移出工作空间？", async () => {
    await removeWorkspaceUser(id, userId);
    ui.showToast("移除成功", "success");
    confirmModal.value?.close();
    await fetchUsers();
  });
};

const roleLabel = (role: number) => {
  switch (role) {
    case 1: return "管理员";
    case 2: return "成员";
    default: return "未知";
  }
};

onMounted(async () => {
  loading.value = true;
  try {
    await Promise.all([fetchWorkspace(), fetchApps(), fetchUsers()]);
  } finally {
    loading.value = false;
  }
});
</script>

<template>
  <div class="space-y-6" v-if="!loading">
    <!-- Workspace Info Header -->
    <div class="bg-base-100 rounded-lg shadow-sm border border-base-200 p-4 flex items-center justify-between">
      <div class="flex items-center gap-3">
        <button class="btn btn-ghost btn-sm btn-square" @click="router.back()">
          <Icon icon="mdi:arrow-left" class="w-4 h-4" />
        </button>
        <div>
          <div class="flex items-center gap-2">
            <h2 class="text-xl font-bold">{{ workspace?.name }}</h2>
            <span :class="workspace?.status === 1 ? 'badge badge-success badge-sm' : 'badge badge-ghost badge-sm'">
              {{ workspace?.status === 1 ? "启用" : "禁用" }}
            </span>
          </div>
          <p class="text-sm text-base-content/60 mt-0.5">{{ workspace?.description || "暂无描述" }}</p>
        </div>
      </div>
    </div>

    <!-- App Management -->
    <div class="bg-base-100 rounded-lg shadow-sm border border-base-200 overflow-hidden">
      <div class="flex items-center justify-between px-4 py-3 border-b border-base-200">
        <h3 class="font-semibold">应用列表</h3>
        <button class="btn btn-primary btn-xs" @click="openCreateApp">
          <Icon icon="mdi:plus" class="w-3.5 h-3.5" /> 新建应用
        </button>
      </div>
      <div class="overflow-x-auto">
        <table class="table table-zebra table-sm">
          <thead>
            <tr>
              <th>ID</th>
              <th>名称</th>
              <th>描述</th>
              <th>App Secret</th>
              <th>状态</th>
              <th>创建时间</th>
              <th class="text-center">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="appLoading">
              <td colspan="7" class="text-center py-8">
                <span class="loading loading-spinner loading-md text-primary" />
              </td>
            </tr>
            <tr v-else-if="!apps.length">
              <td colspan="7" class="text-center py-8 text-base-content/50">暂无应用</td>
            </tr>
            <tr v-for="app in apps" :key="app.id">
              <td class="font-mono text-xs">{{ app.id }}</td>
              <td class="font-medium">{{ app.name }}</td>
              <td class="text-sm text-base-content/70">{{ app.description || "-" }}</td>
              <td>
                <code class="text-xs bg-base-200 px-1 py-0.5 rounded font-mono select-all">{{ app.app_secret }}</code>
              </td>
              <td>
                <span :class="app.status === 1 ? 'badge badge-success badge-sm' : 'badge badge-ghost badge-sm'">
                  {{ app.status === 1 ? "启用" : "禁用" }}
                </span>
              </td>
              <td class="text-xs">{{ formatDateTime(app.created_at) }}</td>
              <td class="text-center">
                <div class="flex items-center justify-center gap-1">
                  <button class="btn btn-ghost btn-xs" @click="openConfig(app)" title="配置">
                    <Icon icon="mdi:cog" class="w-3.5 h-3.5" />
                  </button>
                  <button class="btn btn-ghost btn-xs" @click="openEditApp(app)">
                    <Icon icon="mdi:pencil" class="w-3.5 h-3.5" />
                  </button>
                  <button class="btn btn-ghost btn-xs text-error" @click="handleDeleteApp(app.id, app.name)">
                    <Icon icon="mdi:delete" class="w-3.5 h-3.5" />
                  </button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- User Management -->
    <div class="bg-base-100 rounded-lg shadow-sm border border-base-200 overflow-hidden">
      <div class="flex items-center justify-between px-4 py-3 border-b border-base-200">
        <h3 class="font-semibold">成员管理</h3>
        <button class="btn btn-primary btn-xs" @click="openAddUser">
          <Icon icon="mdi:plus" class="w-3.5 h-3.5" /> 添加成员
        </button>
      </div>
      <div class="overflow-x-auto">
        <table class="table table-zebra table-sm">
          <thead>
            <tr>
              <th>用户</th>
              <th>角色</th>
              <th>加入时间</th>
              <th class="text-center">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="userLoading">
              <td colspan="4" class="text-center py-8">
                <span class="loading loading-spinner loading-md text-primary" />
              </td>
            </tr>
            <tr v-else-if="!users.length">
              <td colspan="4" class="text-center py-8 text-base-content/50">暂无成员</td>
            </tr>
            <tr v-for="u in users" :key="u.id">
              <td>
                <div class="flex items-center gap-2">
                  <div class="avatar placeholder">
                    <div class="w-7 h-7 rounded-full bg-base-200 text-xs">
                      <span v-if="u.avatar"><img :src="u.avatar" class="rounded-full" /></span>
                      <span v-else>{{ u.nickname?.charAt(0) || u.user_id.toString().charAt(0) }}</span>
                    </div>
                  </div>
                  <div>
                    <div class="text-sm font-medium">{{ u.nickname || '用户' + u.user_id }}</div>
                    <div class="text-xs text-base-content/50 font-mono">#{{ u.user_id }}</div>
                  </div>
                </div>
              </td>
              <td>
                <div class="flex items-center gap-2">
                  <span class="badge badge-ghost badge-sm">{{ roleLabel(u.role) }}</span>
                  <button class="btn btn-ghost btn-xs" @click="openEditRole(u)">
                    <Icon icon="mdi:pencil" class="w-3 h-3" />
                  </button>
                </div>
              </td>
              <td class="text-xs">{{ formatDateTime(u.created_at) }}</td>
              <td class="text-center">
                <button class="btn btn-ghost btn-xs text-error" @click="handleRemoveUser(u.user_id)">
                  <Icon icon="mdi:account-remove" class="w-3.5 h-3.5" /> 移除
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Loading state -->
    <div v-if="loading" class="flex justify-center py-20">
      <span class="loading loading-spinner loading-lg text-primary" />
    </div>
  </div>

  <!-- App Create/Edit Modal -->
  <dialog ref="appModal" class="modal">
    <div class="modal-box">
      <h3 class="font-bold text-lg mb-4">{{ appEditMode ? '编辑应用' : '新建应用' }}</h3>
      <div class="space-y-3">
        <label class="form-control w-full">
          <span class="label-text">名称</span>
          <input v-model="appForm.name" type="text" class="input input-bordered input-sm w-full" placeholder="应用名称" maxlength="64" />
        </label>
        <label class="form-control w-full">
          <span class="label-text">描述</span>
          <textarea v-model="appForm.description" class="textarea textarea-bordered textarea-sm w-full" placeholder="可选描述" maxlength="255"></textarea>
        </label>
      </div>
      <div class="modal-action">
        <button class="btn btn-ghost btn-sm" @click="appModal?.close()">取消</button>
        <button class="btn btn-primary btn-sm" @click="handleSaveApp">保存</button>
      </div>
    </div>
    <form method="dialog" class="modal-backdrop"><button>close</button></form>
  </dialog>

  <!-- App Config Modal -->
  <dialog ref="configModal" class="modal">
    <div class="modal-box max-w-2xl">
      <h3 class="font-bold text-lg mb-4">应用配置 — {{ configApp?.name }}</h3>
      <label class="form-control w-full">
        <span class="label-text">配置内容 (JSON)</span>
        <textarea v-model="configForm" class="textarea textarea-bordered w-full mt-1 font-mono text-xs" rows="15" placeholder="{}"></textarea>
      </label>
      <div class="modal-action">
        <button class="btn btn-ghost btn-sm" @click="configModal?.close()">取消</button>
        <button class="btn btn-primary btn-sm" @click="handleSaveConfig">保存配置</button>
      </div>
    </div>
    <form method="dialog" class="modal-backdrop"><button>close</button></form>
  </dialog>

  <!-- Add User Modal -->
  <dialog ref="userModal" class="modal">
    <div class="modal-box">
      <h3 class="font-bold text-lg mb-4">添加成员</h3>
      <div class="space-y-3">
        <label class="form-control w-full">
          <span class="label-text">用户ID</span>
          <input v-model.number="addUserForm.user_id" type="number" class="input input-bordered input-sm w-full" placeholder="输入用户ID" min="1" />
        </label>
        <label class="form-control w-full">
          <span class="label-text">角色</span>
          <select v-model.number="addUserForm.role" class="select select-bordered select-sm w-full">
            <option :value="1">管理员</option>
            <option :value="2">成员</option>
          </select>
        </label>
      </div>
      <div class="modal-action">
        <button class="btn btn-ghost btn-sm" @click="userModal?.close()">取消</button>
        <button class="btn btn-primary btn-sm" @click="handleAddUser">添加</button>
      </div>
    </div>
    <form method="dialog" class="modal-backdrop"><button>close</button></form>
  </dialog>

  <!-- Role Edit Modal -->
  <dialog ref="roleModal" class="modal">
    <div class="modal-box">
      <h3 class="font-bold text-lg mb-4">修改角色</h3>
      <label class="form-control w-full">
        <span class="label-text">角色</span>
        <select v-model.number="editRoleForm.role" class="select select-bordered select-sm w-full mt-1">
          <option :value="1">管理员</option>
          <option :value="2">成员</option>
        </select>
      </label>
      <div class="modal-action">
        <button class="btn btn-ghost btn-sm" @click="roleModal?.close()">取消</button>
        <button class="btn btn-primary btn-sm" @click="handleUpdateRole">保存</button>
      </div>
    </div>
    <form method="dialog" class="modal-backdrop"><button>close</button></form>
  </dialog>

  <!-- Confirm Modal -->
  <dialog ref="confirmModal" class="modal">
    <div class="modal-box">
      <h3 class="font-bold text-lg mb-2">确认操作</h3>
      <p class="text-base-content/70">{{ confirmMessage }}</p>
      <div class="modal-action">
        <button class="btn btn-ghost btn-sm" @click="confirmModal?.close()">取消</button>
        <button class="btn btn-error btn-sm" @click="confirmAction?.()">确认</button>
      </div>
    </div>
    <form method="dialog" class="modal-backdrop"><button>close</button></form>
  </dialog>
</template>
