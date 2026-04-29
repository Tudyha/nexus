<script setup lang="ts">
import type { SystemInfo } from '@/types';
import { formatDateTime, formatUptime } from '@/utils';
const props = defineProps<{
    systemInfo?: SystemInfo;
}>();
const systemInfo = computed(() => {
    return [
        { label: '主机名称', value: props.systemInfo?.hostname || '未知' },
        { label: '发行版本', value: (props.systemInfo?.platform_family || '未知') + " " + (props.systemInfo?.platform_version || '') },
        { label: '内核版本', value: props.systemInfo?.kernel_version || '未知' },
        { label: '系统类型', value: props.systemInfo?.arch || '未知' },
        { label: '启动时间', value: formatDateTime((props.systemInfo?.boot_time ?? 0) * 1000) || '未知' },
        { label: '运行时间', value: formatUptime(props.systemInfo?.uptime) || '未知' },
    ]
});

</script>

<template>
    <div class="card bg-base-100 border border-base-200 shadow-sm">
        <div class="card-body p-6">
            <div class="flex items-center justify-between">
                <h2 class="card-title text-lg font-bold flex items-center gap-2">
                    <Icon icon="mdi:info" class="w-5 h-5 text-primary" />
                    系统信息
                </h2>
            </div>
            <div class="space-y-4 p-4">
                <div v-for="(item, index) in systemInfo" :key="index" class="space-x-4">
                    <span>{{ item.label }}</span>
                    <span>{{ item.value }}</span>
                </div>
            </div>
        </div>
    </div>
</template>