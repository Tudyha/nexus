<script setup lang="ts">
import { Icon } from '@iconify/vue';
import CommonTable from '@/components/common/Table.vue';
import type { ClientTunnelResponse } from '@/types';
import { getClientTunnel, createClientTunnel, deleteClientTunnel } from '@/api/client';

const props = defineProps<{
  id: string
}>();

const data = ref<ClientTunnelResponse[]>([]);
const deleteTarget = ref<ClientTunnelResponse | null>(null);

const addTunnelDialog = ref<HTMLDialogElement | null>(null);
const deleteTunnelDialog = ref<HTMLDialogElement | null>(null);
const loading = ref(false);

const columns = [
  {
    key: 'id',
    label: 'ID',
  },
  {
    key: 'tunnel_type',
    label: '隧道类型',
    render: (_column: any, row: any) => {
      const text = row.tunnel_type === 1 ? 'TCP' : 'UDP';
      return h('span', text);
    }
  },
  {
    key: 'local_port',
    label: '本地端口',
  },
  {
    key: 'remote_addr',
    label: '远程地址',
  },
  {
    key: 'action',
    label: '操作',
    render: (_column: any, row: any) => {
      return h('button', {
        class: 'btn btn-sm btn-ghost text-error',
        onClick: () => {
          deleteTarget.value = row;
          deleteTunnelDialog.value?.showModal();
        }
      }, [
        h(Icon, { icon: 'mdi:trash-can-outline', class: 'w-4 h-4' })
      ]);
    }
  },
]

const form = ref({
  tunnel_type: 1,
  local_port: null,
  remote_addr: '',
});

onMounted(async () => {
  handleSearch();
});

const handleSearch = async () => {
  const res = await getClientTunnel(Number(props.id));
  data.value = res;
};

const handleAdd = async () => {
  loading.value = true;
  try {
    await createClientTunnel(Number(props.id), form.value);
    await handleSearch();
    form.value = {
      tunnel_type: 1,
      local_port: null,
      remote_addr: '',
    };
    addTunnelDialog.value?.close();
  } finally {
    loading.value = false;
  }
}

const handleDelete = async () => {
  if (!deleteTarget.value) return;
  loading.value = true;
  try {
    await deleteClientTunnel(Number(props.id), deleteTarget.value.id);
    deleteTarget.value = null;
    deleteTunnelDialog.value?.close();
    await handleSearch();
  } finally {
    loading.value = false;
  }
}
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-end space-x-1">
      <button class="btn btn-sm btn-primary" @click="addTunnelDialog?.showModal()">
        <icon icon="mdi:plus-thick" class="w-4 h-4" />
      </button>
      <button class="btn btn-sm btn-primary" @click="handleSearch">
        <icon icon="mdi:refresh" class="w-4 h-4" />
      </button>
    </div>
    <CommonTable :data="data" :columns="columns" />

    <!-- 新增隧道弹窗 -->
    <dialog ref="addTunnelDialog" id="add_tunnel_dialog" class="modal">
      <div class="modal-box">
        <h3 class="text-lg font-bold mb-4">新增隧道代理</h3>

        <form class="fieldset" @submit.prevent="handleAdd">
          <fieldset class="fieldset">
            <label class="label">
              <span>隧道类型</span>
            </label>
            <select class="select select-sm" v-model="form.tunnel_type" required>
              <option value="1">TCP</option>
              <option value="2">UDP</option>
            </select>
            <p class="validator-hint">请选择隧道类型</p>
          </fieldset>
          <fieldset class="fieldset">
            <label class="label">
              <span>本地端口</span>
            </label>
            <input v-model.number="form.local_port" type="number" class="input input-sm validator" required
              placeholder="请输入本地端口" />
            <p class="validator-hint">请输入本地端口</p>
          </fieldset>
          <fieldset class="fieldset">
            <label class="label">
              <span>远程地址</span>
            </label>
            <input v-model="form.remote_addr" type="text" class="input input-sm validator" required
              placeholder="请输入远程地址，格式：IP:PORT" />
            <p class="validator-hint">请输入远程地址</p>
          </fieldset>
          <div class="flex justify-end">
            <button class="btn" type="submit" :disabled="loading">{{ loading ? '提交中...' : '确定' }}</button>
          </div>
        </form>
      </div>
    </dialog>

    <!-- 删除隧道确认弹窗 -->
    <dialog ref="deleteTunnelDialog" id="delete_tunnel_dialog" class="modal">
      <div class="modal-box">
        <h3 class="text-lg font-bold mb-2">确认删除</h3>
        <p class="text-base-content/70">确认删除隧道？此操作将关闭该隧道并断开所有连接。</p>
        <div class="modal-action">
          <button class="btn btn-ghost" @click="deleteTunnelDialog?.close()">取消</button>
          <button class="btn btn-error" :disabled="loading" @click="handleDelete">{{ loading ? '删除中...' : '确认删除' }}</button>
        </div>
      </div>
    </dialog>
  </div>
</template>
