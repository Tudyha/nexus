<template>
  <div ref="terminalEl" class="h-full bg-black rounded-b-box" />
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

// xterm 配置
const options = {
  cursorBlink: true,
  theme: {
    foreground: '#ECECEC',
    background: '#000000',
  },
  rows: 40
};

const initXterm = () => {
  xterm = new Terminal(options);
  xterm.loadAddon(fitAddon);

  // 挂载到 DOM
  xterm.open(terminalEl.value);
  fitAddon.fit();

  const u = useUserStore();

  // 建立 WebSocket 连接
  const wsUrl = `${VITE_WS_API_BASE_URL}/v1/client/${props.id}/pty?token=${u.token}`;

  ws = new WebSocket(wsUrl);
  ws.binaryType = 'arraybuffer'
  ws.onopen = () => xterm?.focus()

  // 接收消息并显示在终端上
  ws.onmessage = event => {
    if (!xterm) return;
    xterm.write(new Uint8Array(event.data));
  };

  ws.onclose = () => xterm?.write('\r\n\x1b[31m[disconnected]\x1b[0m\r\n')

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
  console.log("close terminal");
  if (ws) {
    ws.close();
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
