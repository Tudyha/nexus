import { defineConfig, loadEnv } from "vite";
import vue from "@vitejs/plugin-vue";
import tailwindcss from "@tailwindcss/vite";
import AutoImport from "unplugin-auto-import/vite";
import Component from "unplugin-vue-components/vite";
import Icons from "unplugin-icons/vite";
import { resolve } from "path";
import { viteMockServe } from "vite-plugin-mock";
import RadixVueResolver from 'radix-vue/resolver'

// https://vite.dev/config/
export default defineConfig(({ mode }) => {
  const env = loadEnv(mode, process.cwd());
  return {
    plugins: [
      vue(),
      tailwindcss(),
      AutoImport({
        imports: ["vue", "vue-router", "vue-i18n", "@vueuse/core"],
        dts: "src/auto-import.d.ts",
        eslintrc: {
          enabled: true,
        },
        resolvers: [RadixVueResolver()],
      }),
      Component({
        dts: "src/components.d.ts",
        dirs: ["src/components"],
        extensions: ["vue"],
      }),
      Icons({
        compiler: "vue3",
        autoInstall: true,
      }),
      viteMockServe({
        mockPath: "./src/mocks",
        enable: env.VITE_USE_MOCK === "true",
      }),
    ],
    resolve: {
      alias: {
        "@": resolve(__dirname, "src"),
      },
    },
    build: {
      rollupOptions: {
        output: {
          manualChunks: {
            // Vue 核心运行时
            "vue-vendor": ["vue", "vue-router", "pinia", "pinia-plugin-persistedstate", "vue-i18n"],
            // 图表库 echarts（~1MB），仅在监控页使用
            "echarts": ["echarts", "vue-echarts"],
            // 轻量图表 chart.js，在 dashboard 使用
            "chart": ["chart.js", "vue-chartjs"],
            // 终端组件（~800KB），仅在客户端详情页使用
            "xterm": ["@xterm/xterm", "@xterm/addon-fit"],
          },
        },
      },
      // 启用 CSS 代码分割
      cssCodeSplit: true,
      // 生成 sourcemap 便于调试（生产环境可关闭）
      sourcemap: false,
    },
    server: {
      proxy: {
        "/api": {
          target: "http://127.0.0.1:8080",
          changeOrigin: true,
        },
        '/ws-api': {
          target: 'ws://127.0.0.1:8080',
          changeOrigin: true,
          ws: true,
          rewrite: (path) => path.replace(/^\/ws-api/, '/api'),
        },
      },
    },
  };
});
