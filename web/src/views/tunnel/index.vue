<script setup lang="ts">
import { onMounted, ref } from "vue";
import { useRouter } from "vue-router";
import { Icon } from "@iconify/vue";
import { getAllTunnels, createClientTunnel, getOnlineClients } from "@/api/client";
import { formatDateTime } from "@/utils";
import type { ClientTunnelResponse } from "@/types";

const router = useRouter();
const loading = ref(false);
const tunnels = ref<ClientTunnelResponse[]>([]);

// --- Create dialog ---
const createLoading = ref(false);
const createDialog = ref<HTMLDialogElement>();
const clients = ref<{ id: number; hostname: string }[]>([]);

const createForm = ref({
  client_id: null as number | null,
  tunnel_type: 1,
  local_port: null as number | null,
  remote_addr: "",
});

const fetchClients = async () => {
  try {
    clients.value = await getOnlineClients();
  } catch { /* ignore */ }
};

const openCreateDialog = async () => {
  await fetchClients();
  createForm.value = { client_id: null, tunnel_type: 1, local_port: null, remote_addr: "" };
  createDialog.value?.showModal();
};

const handleCreate = async () => {
  if (!createForm.value.client_id) return;
  createLoading.value = true;
  try {
    await createClientTunnel({
      client_id: createForm.value.client_id,
      tunnel_type: createForm.value.tunnel_type,
      local_port: createForm.value.local_port,
      remote_addr: createForm.value.remote_addr,
    });
    createDialog.value?.close();
    fetchTunnels();
  } finally {
    createLoading.value = false;
  }
};

// --- List ---
const fetchTunnels = async () => {
  loading.value = true;
  try {
    tunnels.value = await getAllTunnels();
  } finally {
    loading.value = false;
  }
};

const tunnelTypeLabel = (t: number) => t === 1 ? "TCP" : "UDP";
const statusLabel = (s: number) => {
  switch (s) {
    case 1: return { text: "活跃", cls: "badge-success" };
    case 2: return { text: "停用", cls: "badge-ghost" };
    default: return { text: "未知", cls: "badge-ghost" };
  }
};

const viewClient = (clientId: number) => {
  router.push({ name: "ClientConsole", params: { id: String(clientId) } });
};

onMounted(fetchTunnels);
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-2xl font-bold text-base-content">隧道管理</h2>
      <div class="flex items-center gap-2">
        <button class="btn btn-primary btn-sm" @click="openCreateDialog">
          <Icon icon="mdi:plus" class="w-4 h-4" /> 新建隧道
        </button>
        <button class="btn btn-ghost btn-sm" @click="fetchTunnels">
          <Icon icon="mdi:refresh" class="w-4 h-4" /> 刷新
        </button>
      </div>
    </div>

    <div class="bg-base-100 rounded-lg shadow-sm border border-base-200 overflow-hidden">
      <div class="overflow-x-auto">
        <table class="table table-zebra table-sm">
          <thead>
            <tr>
              <th>ID</th>
              <th>客户端</th>
              <th>类型</th>
              <th>本地端口</th>
              <th>远程地址</th>
              <th>状态</th>
              <th>创建时间</th>
              <th class="text-center">操作</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="8" class="text-center py-8">
                <span class="loading loading-spinner loading-md text-primary" />
              </td>
            </tr>
            <tr v-else-if="!tunnels.length">
              <td colspan="8" class="text-center py-8 text-base-content/50">暂无隧道</td>
            </tr>
            <tr v-for="t in tunnels" :key="t.id">
              <td class="font-mono text-xs">{{ t.id }}</td>
              <td>
                <a class="link link-primary text-sm" @click="viewClient(t.client_id)">
                  #{{ t.client_id }}
                </a>
              </td>
              <td>
                <span class="badge badge-ghost badge-sm">{{ tunnelTypeLabel(t.tunnel_type) }}</span>
              </td>
              <td class="font-mono text-xs">{{ t.local_port }}</td>
              <td class="font-mono text-xs">{{ t.remote_addr }}</td>
              <td>
                <span :class="['badge', 'badge-sm', statusLabel(t.status).cls]">
                  {{ statusLabel(t.status).text }}
                </span>
              </td>
              <td class="text-xs">{{ formatDateTime(t.created_at) }}</td>
              <td class="text-center">
                <button class="btn btn-ghost btn-xs" @click="viewClient(t.client_id)" title="查看客户端">
                  <Icon icon="mdi:eye" class="w-3.5 h-3.5" />
                </button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Create Tunnel Dialog -->
    <dialog ref="createDialog" class="modal">
      <div class="modal-box">
        <h3 class="text-lg font-bold mb-4 flex items-center gap-2">
          <Icon icon="mdi:plus-circle-outline" class="w-5 h-5 text-primary" />
          新建隧道
        </h3>

        <form class="fieldset" @submit.prevent="handleCreate">
          <fieldset class="fieldset">
            <label class="label"><span>选择客户端</span></label>
            <select class="select select-sm" v-model.number="createForm.client_id" required>
              <option disabled :value="null">请选择客户端</option>
              <option v-for="c in clients" :key="c.id" :value="c.id">
                #{{ c.id }} {{ c.hostname }}
              </option>
            </select>
          </fieldset>
          <fieldset class="fieldset">
            <label class="label"><span>隧道类型</span></label>
            <select class="select select-sm" v-model.number="createForm.tunnel_type" required>
              <option :value="1">TCP</option>
              <option :value="2">UDP</option>
            </select>
          </fieldset>
          <fieldset class="fieldset">
            <label class="label"><span>本地端口</span></label>
            <input v-model.number="createForm.local_port" type="number" class="input input-sm validator" required
              placeholder="请输入本地端口" min="1" max="65535" />
          </fieldset>
          <fieldset class="fieldset">
            <label class="label"><span>远程地址</span></label>
            <input v-model="createForm.remote_addr" type="text" class="input input-sm validator" required
              placeholder="IP:PORT" />
          </fieldset>
          <div class="flex justify-end gap-2 mt-4">
            <button class="btn btn-ghost btn-sm" type="button" @click="createDialog?.close()">取消</button>
            <button class="btn btn-primary btn-sm" type="submit" :disabled="createLoading">
              {{ createLoading ? '创建中...' : '创建' }}
            </button>
          </div>
        </form>
      </div>
      <form method="dialog" class="modal-backdrop">
        <button>close</button>
      </form>
    </dialog>
  </div>
</template>
