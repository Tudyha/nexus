<script setup lang="ts">
import type { TaskExecutionResponse } from '@/types';

defineProps<{
  currentTask: TaskExecutionResponse | null;
  hasRunningTask: boolean | null;
  showCompleted: boolean;
  completedSuccess: boolean;
}>();
</script>

<template>
  <div v-if="hasRunningTask || showCompleted" class="px-3 pb-2">
    <div v-if="showCompleted" class="flex items-center gap-2 text-xs"
      :class="completedSuccess ? 'text-success' : 'text-error'">
      <Icon :icon="completedSuccess ? 'mdi:check-circle' : 'mdi:alert-circle'" class="w-4 h-4" />
      <span>{{ completedSuccess ? '升级完成' : (currentTask?.error || '升级失败') }}</span>
    </div>
    <div v-else class="space-y-1">
      <div class="flex justify-between text-xs">
        <span class="text-base-content/70">{{ currentTask?.message || '升级中' }}</span>
        <span class="text-base-content/50">{{ currentTask?.progress || 0 }}%</span>
      </div>
      <progress class="progress progress-primary w-full h-2" :value="currentTask?.progress || 0" max="100"></progress>
    </div>
  </div>
</template>
