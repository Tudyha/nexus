<script setup lang="ts">
import type { SystemInfo } from '@/types';
import { formatDateTime, formatUptime, formatBytesToGB } from '@/utils';

const props = defineProps<{
    systemInfo?: SystemInfo;
}>();
const systemInfo = computed(() => {
    const info = props.systemInfo;
    if (!info) return [];
    return [
        { label: '主机名称', value: info.hostname || '未知' },
        { label: '发行版本', value: (info.platform_family || '未知') + ' ' + (info.platform_version || '') },
        { label: '内核版本', value: info.kernel_version || '未知' },
        { label: '系统类型', value: info.arch || '未知' },
        { label: 'CPU', value: info.cpu_info ? `${info.cpu_num}核 ${info.cpu_info}` : `${info.cpu_num || '?'}核` },
        { label: '内存总量', value: info.mem_total ? formatBytesToGB(info.mem_total) + 'GB' : '未知' },
        { label: '磁盘总量', value: info.disk_total ? formatBytesToGB(info.disk_total) + 'GB' : '未知' },
        { label: '启动时间', value: formatDateTime((info.boot_time ?? 0) * 1000) || '未知' },
        { label: '运行时间', value: formatUptime(info.uptime) || '未知' },
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
            <div class="space-y-3">
                <div v-for="(item, index) in systemInfo" :key="index"
                     class="flex justify-between items-center py-1 border-b border-base-200/50 last:border-b-0">
                    <span class="text-sm text-base-content/60">{{ item.label }}</span>
                    <span class="text-sm font-medium text-right max-w-[55%] truncate" :title="item.value">{{ item.value }}</span>
                </div>
            </div>
        </div>
    </div>
</template>
