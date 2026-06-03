<template>
  <div class="relative h-full">
    <div ref="terminalEl" class="h-full bg-black rounded-b-box" />
    <!-- Reconnection indicator -->
    <div v-if="reconnecting"
      class="absolute top-2 right-2 z-10 flex items-center gap-2 px-3 py-1.5 bg-warning text-warning-content rounded-lg shadow-lg text-xs font-medium animate-pulse">
      <span class="loading loading-spinner loading-xs" />
      正在重连...
    </div>
    <div v-if="disconnected && !reconnecting"
      class="absolute top-2 right-2 z-10 flex items-center gap-2 px-3 py-1.5 bg-error text-error-content rounded-lg shadow-lg text-xs font-medium cursor-pointer" @click="reconnect">
      <Icon icon="mdi:connection" class="w-3.5 h-3.5" />
      连接断开，点击重连
    </div>
  </div>
</template>

<script setup lang="ts">
import { Terminal } from "@xterm/xterm";
import { FitAddon } from "@xterm/addon-fit";
import "@xterm/xterm/css/xterm.css";
import { useUserStore } from "@/stores/auth";
const { VITE_WS_API_BASE_URL } = import.meta.env;

const props = defineProps<{
  id: string
}>()

const terminalEl = ref(); // xterm DOM 引用
let ws: WebSocket | null = null; // WebSocket 实例
const fitAddon = new FitAddon(); // xterm fit 插件实例
let xterm: Terminal | null = null; // xterm 实例
let resizeObs: ResizeObserver // 终端尺寸变化观察器
let reconnectTimer: ReturnType<typeof setTimeout> | null = null;
let manualClose = false;
const maxRetries = 10;
let retryCount = 0;

const reconnecting = ref(false);
const disconnected = ref(false);

// xterm 配置
const options = {
  cursorBlink: true,
  theme: {
    foreground: '#ECECEC',
    background: '#000000',
  },
  rows: 40
};

const getWsUrl = () => {
  const u = useUserStore();
  return `${VITE_WS_API_BASE_URL}/v1/client/${props.id}/terminal?token=${u.token}`;
};

const connect = () => {
  const u = useUserStore();
  const wsUrl = getWsUrl();

  if (ws) {
    ws.close();
    ws = null;
  }

  ws = new WebSocket(wsUrl);
  ws.binaryType = 'arraybuffer'

  ws.onopen = () => {
    reconnecting.value = false;
    disconnected.value = false;
    retryCount = 0;
    xterm?.focus();
    if (xterm) {
      xterm.write('\r\n\x1b[32m[connected]\x1b[0m\r\n');
    }
  };

  ws.onmessage = event => {
    if (!xterm) return;
    xterm.write(new Uint8Array(event.data));
  };

  ws.onclose = () => {
    if (manualClose) return;
    disconnected.value = true;
    xterm?.write('\r\n\x1b[33m[disconnected]\x1b[0m\r\n');
    scheduleReconnect();
  };

  ws.onerror = () => {
    // onclose will fire after this
  };
};

const scheduleReconnect = () => {
  if (retryCount >= maxRetries) {
    reconnecting.value = false;
    xterm?.write('\r\n\x1b[31m[reconnect failed, max retries reached]\x1b[0m\r\n');
    return;
  }
  reconnecting.value = true;
  retryCount++;
  const delay = Math.min(1000 * Math.pow(1.5, retryCount - 1), 15000);
  reconnectTimer = setTimeout(() => {
    xterm?.write(`\r\n\x1b[33m[reconnecting... attempt ${retryCount}/${maxRetries}]\x1b[0m\r\n`);
    connect();
  }, delay);
};

const reconnect = () => {
  retryCount = 0;
  reconnecting.value = true;
  connect();
};

const initXterm = () => {
  xterm = new Terminal(options);
  xterm.loadAddon(fitAddon);

  // 挂载到 DOM
  xterm.open(terminalEl.value);
  fitAddon.fit();

  manualClose = false;
  connect();

  // 发送输入内容到服务端
  xterm.onData(input => {
    if (!xterm) return;
    if (!ws) return;
    if (ws.readyState !== WebSocket.OPEN) return;
    const data = new TextEncoder().encode(input);
    const frame = new Uint8Array(1 + data.length);
    frame[0] = 0x00;
    frame.set(data, 1);
    ws.send(frame);
  });

  xterm.onResize(({ rows, cols }) => {
    if (!ws) return;
    if (ws.readyState !== WebSocket.OPEN) return
    const frame = new Uint8Array(5)
    const view = new DataView(frame.buffer)
    frame[0] = 0x01
    view.setUint16(1, rows, false)
    view.setUint16(3, cols, false)
    ws.send(frame)
  })

  resizeObs = new ResizeObserver(() => fitAddon.fit())
  resizeObs.observe(terminalEl.value)
}

const closeTerminal = () => {
  manualClose = true;
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }
  if (ws) {
    ws.close();
    ws = null;
  }
  if (xterm) {
    xterm.dispose();
  }
  if (resizeObs) {
    resizeObs.disconnect();
  }
}

defineExpose({
  close: closeTerminal
});

onMounted(() => {
  initXterm();
});

onUnmounted(() => {
  closeTerminal();
});
</script>
