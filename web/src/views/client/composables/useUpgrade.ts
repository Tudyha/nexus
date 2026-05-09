import { ref, computed, onUnmounted } from 'vue';
import type { TaskExecutionResponse } from '@/types';
import { triggerUpgrade, getTaskProgress } from '@/api/version';
import { useUIStore } from '@/stores/ui';

export function useUpgrade(clientId: number, getInitialTask: () => TaskExecutionResponse | null) {
  const ui = useUIStore();

  const upgrading = ref(false);
  const pollingTaskId = ref<number | null>(null);
  const pollingProgress = ref<TaskExecutionResponse | null>(null);
  const showCompleted = ref(false);
  const completedSuccess = ref(false);
  let pollingTimer: ReturnType<typeof setInterval> | null = null;

  const currentTask = computed<TaskExecutionResponse | null>(() => {
    if (pollingTaskId.value !== null && pollingProgress.value) {
      return pollingProgress.value;
    }
    return getInitialTask();
  });

  const hasRunningTask = computed(() => {
    const task = currentTask.value;
    return task && task.status === 1;
  });

  const stopPolling = () => {
    if (pollingTimer) {
      clearInterval(pollingTimer);
      pollingTimer = null;
    }
    pollingTaskId.value = null;
    pollingProgress.value = null;
  };

  const handleUpgrade = async (force: boolean = false) => {
    const elem = document.activeElement as HTMLElement;
    if (elem) elem.blur();

    upgrading.value = true;
    try {
      const res = await triggerUpgrade(clientId, force);
      ui.showToast(`已触发${force ? '强制' : ''}拉取升级`, 'success');

      stopPolling();
      pollingTaskId.value = res.task_id;
      pollingProgress.value = {
        id: 0,
        task_id: res.task_id,
        client_id: clientId,
        task_type: 1,
        status: 1,
        progress: 0,
        message: '任务已创建',
        error: '',
        created_at: Date.now(),
        updated_at: Date.now(),
      };

      pollingTimer = setInterval(async () => {
        try {
          const progress = await getTaskProgress(res.task_id, clientId);
          if (progress) {
            pollingProgress.value = progress;
            if (progress.status === 3) {
              stopPolling();
              showCompleted.value = true;
              completedSuccess.value = true;
            } else if (progress.status === 2 || progress.status === 4) {
              stopPolling();
              showCompleted.value = true;
              completedSuccess.value = false;
            }
          }
        } catch {
          stopPolling();
        }
      }, 2000);
    } catch {
      // toast handled by interceptor
    } finally {
      upgrading.value = false;
    }
  };

  onUnmounted(() => {
    stopPolling();
  });

  return {
    upgrading,
    currentTask,
    hasRunningTask,
    showCompleted,
    completedSuccess,
    handleUpgrade,
    stopPolling,
  };
}
