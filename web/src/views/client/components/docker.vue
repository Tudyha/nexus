<script setup lang="ts">
import { getAppTemplates, checkClientApps, installClientApp } from "@/api/ops";
import type { AppTemplate, AppEntry } from "@/api/ops";
import { useUIStore } from "@/stores/ui";

const props = defineProps<{ id: string }>();
const ui = useUIStore();

const templates = ref<AppTemplate[]>([]);
const installed = ref<Map<string, AppEntry>>(new Map());
const loading = ref(false);
const installing = ref<string | null>(null);

// 按分类分组
const categories = computed(() => {
  const map = new Map<string, { label: string; apps: AppTemplate[] }>();
  for (const t of templates.value) {
    const cat = t.category;
    if (!map.has(cat)) {
      const label = cat === "runtime" ? "运行环境" : cat === "database" ? "数据库" : cat === "web" ? "Web 服务" : cat;
      map.set(cat, { label, apps: [] });
    }
    map.get(cat)!.apps.push(t);
  }
  return Array.from(map.values());
});

async function refresh() {
  loading.value = true;
  try {
    const [tmpl, apps] = await Promise.all([
      getAppTemplates(),
      checkClientApps(Number(props.id)),
    ]);
    templates.value = tmpl;
    const m = new Map<string, AppEntry>();
    for (const a of apps) {
      m.set(a.app_id, a);
    }
    installed.value = m;
  } catch {
    // ignore
  } finally {
    loading.value = false;
  }
}

async function handleInstall(appId: string) {
  installing.value = appId;
  try {
    await installClientApp(Number(props.id), appId);
    ui.showToast("安装任务已下发，请稍后刷新查看状态", "success");
  } catch {
    // toast handled by interceptor
  } finally {
    installing.value = null;
  }
}

onMounted(refresh);
</script>

<template>
  <div class="h-full flex flex-col space-y-4">
    <div class="flex items-center justify-between shrink-0">
      <h3 class="text-sm font-semibold text-base-content/70">应用商店</h3>
      <button class="btn btn-ghost btn-xs" :disabled="loading" @click="refresh">
        <Icon icon="mdi:refresh" :class="{ 'animate-spin': loading }" class="w-3.5 h-3.5" />
        刷新
      </button>
    </div>

    <div v-if="loading" class="flex-1 flex items-center justify-center">
      <span class="loading loading-spinner loading-md text-primary" />
    </div>

    <div v-else-if="!templates.length" class="flex-1 flex items-center justify-center text-base-content/40 text-sm">
      暂无可用应用
    </div>

    <div v-else class="flex-1 overflow-y-auto space-y-6 pr-1">
      <div v-for="cat in categories" :key="cat.label">
        <h4 class="text-xs font-medium text-base-content/50 mb-2 uppercase tracking-wider">{{ cat.label }}</h4>
        <div class="grid grid-cols-1 sm:grid-cols-2 gap-3">
          <div v-for="t in cat.apps" :key="t.id"
            class="card card-compact bg-base-100 border border-base-200 shadow-sm hover:shadow-md transition-shadow">
            <div class="card-body p-4">
              <div class="flex items-start justify-between">
                <div class="flex items-center gap-3">
                  <Icon :icon="t.icon" class="w-8 h-8 text-base-content/60" />
                  <div>
                    <h5 class="font-medium text-sm">{{ t.name }}</h5>
                    <p class="text-xs text-base-content/40 mt-0.5">{{ t.description }}</p>
                  </div>
                </div>
                <div class="shrink-0">
                  <!-- 已安装 -->
                  <div v-if="installed.get(t.id)?.version" class="flex items-center gap-2">
                    <span class="badge badge-success badge-xs gap-1">
                      <Icon icon="mdi:check" class="w-3 h-3" />
                      {{ installed.get(t.id)!.version }}
                    </span>
                    <button class="btn btn-ghost btn-xs" title="管理">
                      <Icon icon="mdi:console" class="w-3.5 h-3.5" />
                    </button>
                  </div>
                  <!-- 未安装 -->
                  <button v-else class="btn btn-primary btn-xs" :disabled="installing === t.id"
                    @click="handleInstall(t.id)">
                    <span v-if="installing === t.id" class="loading loading-spinner loading-xs" />
                    <span v-else>安装</span>
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
