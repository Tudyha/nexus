<script setup lang="ts">
import { onMounted, ref } from "vue";
import { formatDateTime } from "@/utils";
import { useUIStore } from "@/stores/ui";

interface UserItem {
  id: number;
  username: string;
  nickname: string;
  phone: string;
  email: string;
  status: number;
  created_at: string;
}

const ui = useUIStore();
const loading = ref(false);
const users = ref<UserItem[]>([]);
const total = ref(0);
const page = ref(1);
const pageSize = ref(20);

const fetchUsers = async () => {
  loading.value = true;
  try {
    const { default: http } = await import("@/api/index");
    const res = await http.get<any>("/v1/user/list", { page: page.value, pageSize: pageSize.value });
    users.value = res.list || [];
    total.value = res.total || 0;
  } finally {
    loading.value = false;
  }
};

const handlePageChange = (p: number) => {
  page.value = p;
  fetchUsers();
};

onMounted(fetchUsers);
</script>

<template>
  <div class="space-y-4">
    <div class="flex items-center justify-between">
      <h2 class="text-2xl font-bold text-base-content">用户列表</h2>
    </div>

    <div class="bg-base-100 rounded-lg shadow-sm border border-base-200 overflow-hidden">
      <div class="overflow-x-auto">
        <table class="table table-zebra table-sm">
          <thead>
            <tr>
              <th>ID</th>
              <th>用户名</th>
              <th>昵称</th>
              <th>手机号</th>
              <th>邮箱</th>
              <th>状态</th>
              <th>创建时间</th>
            </tr>
          </thead>
          <tbody>
            <tr v-if="loading">
              <td colspan="7" class="text-center py-8">
                <span class="loading loading-spinner loading-md text-primary" />
              </td>
            </tr>
            <tr v-else-if="!users.length">
              <td colspan="7" class="text-center py-8 text-base-content/50">暂无用户</td>
            </tr>
            <tr v-for="u in users" :key="u.id">
              <td class="font-mono text-xs">{{ u.id }}</td>
              <td class="font-medium">{{ u.username }}</td>
              <td>{{ u.nickname }}</td>
              <td class="font-mono text-xs">{{ u.phone || "-" }}</td>
              <td class="text-xs">{{ u.email || "-" }}</td>
              <td>
                <span :class="u.status === 1 ? 'badge badge-success badge-sm' : 'badge badge-ghost badge-sm'">
                  {{ u.status === 1 ? "启用" : "禁用" }}
                </span>
              </td>
              <td class="text-xs">{{ u.created_at || "-" }}</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <div v-if="total > pageSize" class="flex justify-center">
      <div class="join">
        <button class="join-item btn btn-sm" :disabled="page <= 1" @click="handlePageChange(page - 1)">上一页</button>
        <button class="join-item btn btn-sm">第 {{ page }} 页 / 共 {{ Math.ceil(total / pageSize) }} 页</button>
        <button class="join-item btn btn-sm" :disabled="page * pageSize >= total" @click="handlePageChange(page + 1)">下一页</button>
      </div>
    </div>
  </div>
</template>
