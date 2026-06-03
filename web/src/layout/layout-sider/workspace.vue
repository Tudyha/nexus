<template>
  <div class="mt-auto pt-4 border-t border-base-300">
    <div class="dropdown dropdown-top w-full">
      <div tabindex="0" role="button" class="btn btn-ghost w-full justify-between hover:bg-base-200">
        <div class="flex items-center gap-3">
          <div class="avatar placeholder">
            <Icon icon="material-symbols:group-outline" />
          </div>
          <div>
            <span class="text-gray-500">{{ currentWorkspaceName }}</span>
          </div>
        </div>
      </div>
      <ul tabindex="0"
        class="dropdown-content z-1 menu p-2 shadow-lg bg-base-100 rounded-box w-50 mb-2 border border-base-200">
        <li class="menu-title">全部空间</li>
        <li v-for="space in workspaceList" :class="{ 'text-primary font-semibold': String(space.id) === String(currentWorkspace) }">
          <a @click="selectWorkspace(space.id)">{{ space.name }}</a>
        </li>
      </ul>
    </div>

    <div class="dropdown dropdown-top w-full">
      <div tabindex="0" role="button" class="btn btn-ghost w-full justify-between hover:bg-base-300">
        <div class="flex items-center gap-3">
          <div class="avatar placeholder">
            <Icon icon="majesticons:applications" />
          </div>
          <div>
            <span class="text-gray-500">{{ currentAppName }}</span>
          </div>
        </div>
      </div>
      <ul tabindex="0"
        class="dropdown-content z-1 menu p-2 shadow-lg bg-base-100 rounded-box w-64 mb-2 border border-base-200">
        <li class="menu-title">全部应用</li>
        <li v-for="app in appList || []" :class="{ 'text-primary font-semibold': String(app.id) === String(currentApp) }">
          <a @click="currentApp = app.id">{{ app.name }}</a>
        </li>
      </ul>
    </div>
  </div>
</template>

<script setup lang="ts">
import { useUserStore } from "@/stores/auth";
import { storeToRefs } from 'pinia';

const userStore = useUserStore();
const { currentWorkspace, currentApp, workspaceList, appList } = storeToRefs(userStore);

const currentWorkspaceName = computed(() => {
  const ws = workspaceList.value.find(space => String(space.id) === String(currentWorkspace.value))
  return ws?.name || '选择空间'
})

const currentAppName = computed(() => {
  const list = appList.value || []
  const app = list.find(a => String(a.id) === String(currentApp.value))
  return app?.name || '选择应用'
})

function selectWorkspace(id: any) {
  currentWorkspace.value = id
  // 切换到该空间下的第一个应用
  const ws = workspaceList.value.find(s => String(s.id) === String(id))
  currentApp.value = ws?.app_list?.[0]?.id || null
}
</script>
