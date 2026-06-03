<script setup lang="ts">
import { getAppConfig, updateAppConfig } from "@/api/ops";
import { useUIStore } from "@/stores/ui";

const ui = useUIStore();
const saving = ref(false);
const loading = ref(true);

// 表单字段
const heartbeatInterval = ref(30)
const connectTimeout = ref(10)
const reconnectInterval = ref(5)
const extraJson = ref("{}")

// 从 API 加载配置
function parseConfig(raw: string) {
  try {
    const obj = JSON.parse(raw)
    heartbeatInterval.value = obj.heartbeat_interval ?? 30
    connectTimeout.value = obj.connect_timeout ?? 10
    reconnectInterval.value = obj.reconnect_interval ?? 5
    // 剩余未知字段保留为 JSON
    const { heartbeat_interval, connect_timeout, reconnect_interval, ...rest } = obj
    extraJson.value = JSON.stringify(rest, null, 2)
  } catch {
    // 解析失败则全部作为 extra
    extraJson.value = raw
  }
}

// 合并为 JSON 字符串
function buildConfig(): string {
  try {
    const extra = JSON.parse(extraJson.value || "{}")
    return JSON.stringify({
      heartbeat_interval: heartbeatInterval.value,
      connect_timeout: connectTimeout.value,
      reconnect_interval: reconnectInterval.value,
      ...extra,
    }, null, 2)
  } catch {
    return ""
  }
}

onMounted(async () => {
  try {
    const res = await getAppConfig()
    parseConfig(res.config || "{}")
  } catch { /* ignore */ }
  loading.value = false
})

const handleSave = async () => {
  const json = buildConfig()
  if (!json) {
    ui.showToast("extra JSON 格式不正确", "error")
    return
  }
  saving.value = true
  try {
    await updateAppConfig({ config: json })
    ui.showToast("配置已保存", "success")
  } catch { /* toast handled by interceptor */ }
  finally { saving.value = false }
}
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <h2 class="text-2xl font-bold text-base-content">应用配置</h2>
      <button class="btn btn-primary btn-sm" :disabled="saving || loading" @click="handleSave">
        <Icon icon="mdi:content-save" class="w-4 h-4" /> {{ saving ? "保存中..." : "保存" }}
      </button>
    </div>

    <div v-if="loading" class="flex justify-center py-12">
      <span class="loading loading-spinner loading-md text-primary" />
    </div>

    <template v-else>
      <!-- 说明 -->
      <div class="alert bg-base-200 border border-base-300 text-sm text-base-content/70">
        <Icon icon="mdi:information-outline" class="w-5 h-5 shrink-0" />
        <span>此配置将在客户端握手时下发，覆盖客户端的本地配置。留空则使用客户端本地默认值。</span>
      </div>

      <!-- 常用配置表单 -->
      <div class="card bg-base-100 border border-base-200 shadow-sm">
        <div class="card-body p-6 space-y-4">
          <h3 class="card-title text-base font-bold flex items-center gap-2">
            <Icon icon="mdi:tune" class="w-5 h-5 text-primary" />
            常用配置
          </h3>

          <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
            <label class="form-control w-full">
              <span class="label-text">心跳间隔（秒）</span>
              <div class="join mt-1">
                <input v-model.number="heartbeatInterval" type="number" class="input input-bordered input-sm join-item w-full" min="5" max="300" />
                <span class="join-item px-3 bg-base-200 text-xs flex items-center">秒</span>
              </div>
              <span class="text-xs text-base-content/40 mt-1">默认 30s，客户端心跳发送频率</span>
            </label>

            <label class="form-control w-full">
              <span class="label-text">连接超时（秒）</span>
              <div class="join mt-1">
                <input v-model.number="connectTimeout" type="number" class="input input-bordered input-sm join-item w-full" min="1" max="60" />
                <span class="join-item px-3 bg-base-200 text-xs flex items-center">秒</span>
              </div>
              <span class="text-xs text-base-content/40 mt-1">默认 10s，与服务端建立连接超时</span>
            </label>

            <label class="form-control w-full">
              <span class="label-text">重连间隔（秒）</span>
              <div class="join mt-1">
                <input v-model.number="reconnectInterval" type="number" class="input input-bordered input-sm join-item w-full" min="1" max="120" />
                <span class="join-item px-3 bg-base-200 text-xs flex items-center">秒</span>
              </div>
              <span class="text-xs text-base-content/40 mt-1">默认 5s，断开后自动重连等待时间</span>
            </label>
          </div>
        </div>
      </div>

      <!-- 高级配置 -->
      <div class="card bg-base-100 border border-base-200 shadow-sm">
        <div class="card-body p-6">
          <div class="flex items-center justify-between mb-2">
            <h3 class="card-title text-base font-bold flex items-center gap-2">
              <Icon icon="mdi:code-json" class="w-5 h-5 text-primary" />
              高级配置 (JSON)
            </h3>
            <button class="btn btn-ghost btn-xs" @click="extraJson = JSON.stringify(JSON.parse(extraJson || '{}'), null, 2)">
              <Icon icon="mdi:code-json" class="w-3.5 h-3.5" /> 格式化
            </button>
          </div>
          <p class="text-xs text-base-content/40 mb-3">以上表单未覆盖的配置项，可直接编辑 JSON。与表单字段合并后下发。</p>
          <textarea v-model="extraJson" class="textarea textarea-bordered font-mono text-sm w-full h-48" spellcheck="false" />
        </div>
      </div>
    </template>
  </div>
</template>
