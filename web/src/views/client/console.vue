<script setup lang="ts">
import { ref, computed } from 'vue';
import { useRoute } from 'vue-router';
import SystemStat from './components/system-stat.vue'
import Terminal from './components/terminal.vue';
import Tunnel from './components/tunnel.vue';
import SystemInfo from './components/system-info.vue';
import Process from './components/process.vue';
import Docker from './components/docker.vue';
import Database from './components/database.vue';
import File from './components/file.vue';
import Network from './components/network.vue';

const route = useRoute();
const id = route.params.id + '';

const activeTab = ref(0);

const tabs = [
    {
        name: "系统资源",
        icon: "mdi:home",
        component: SystemStat,
    },
    {
        name: "在线终端",
        icon: "mdi:terminal",
        component: Terminal,
    },
    {
        name: "文件管理",
        icon: "mdi:folder",
        component: File,
    },
    {
        name: "进程管理",
        icon: "simple-icons:processingfoundation",
        component: Process,
    },
    {
        name: "网络管理",
        icon: "mdi:network",
        component: Network,
    },
    {
        name: "Docker",
        icon: "mdi:docker",
        component: Docker,
    },
    {
        name: "数据库",
        icon: "mdi:database",
        component: Database,
    },
    {
        name: "隧道管理",
        icon: "mdi:tunnel",
        component: Tunnel,
    },
];
const currentComponent = computed(() => tabs[activeTab.value]?.component);
</script>

<template>
    <div class="flex flex-col h-[calc(100vh-(--spacing(16)))] -m-4 sm:-m-6 overflow-hidden">
        <!-- 极简顶部栏 -->
        <SystemInfo :id="id" />

        <!-- 选项卡导航 (紧凑型) -->
        <div class="bg-base-100 border-b border-base-200 px-2 sm:px-4 shrink-0 flex items-center gap-1 overflow-x-auto [&::-webkit-scrollbar]:hidden">
            <button
                v-for="(item, index) in tabs"
                :key="index"
                class="px-3 sm:px-4 py-2 text-sm font-medium transition-all relative shrink-0"
                :class="activeTab === index ? 'text-primary' : 'text-base-content/50 hover:text-base-content'"
                @click="activeTab = index"
            >
                <div class="flex items-center gap-2">
                    <Icon :icon="item.icon" class="w-5 h-5 sm:w-4 sm:h-4" />
                    <span class="hidden sm:inline">{{ item.name }}</span>
                </div>
                <!-- 激活状态下划线 -->
                <div v-if="activeTab === index" class="absolute bottom-0 left-0 right-0 h-0.5 bg-primary"></div>
            </button>
        </div>

        <!-- 内容区域：自动填充剩余高度 -->
        <div class="flex-1 overflow-hidden p-3 bg-base-200/30">
            <div class="h-full bg-base-100 rounded-lg border border-base-200 shadow-sm overflow-hidden">
                <KeepAlive>
                    <component :is="currentComponent" :id="id" class="h-full p-4" />
                </KeepAlive>
            </div>
        </div>
    </div>
</template>
