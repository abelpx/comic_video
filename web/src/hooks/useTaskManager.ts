import { useEffect, useRef, useCallback } from 'react';
import { useTasks } from '../store';
import { taskApi } from '../services/api';
import toast from 'react-hot-toast';

interface UseTaskManagerOptions {
  autoLoad?: boolean;
  pollInterval?: number;
  enablePolling?: boolean;
}

export const useTaskManager = (options: UseTaskManagerOptions = {}) => {
  const {
    autoLoad = true,
    pollInterval = 3000, // 3秒轮询一次
    enablePolling = true,
  } = options;

  const {
    tasks,
    activeTasks,
    addTask,
    updateTask,
    removeTask,
    loadUserTasks,
    setTasks,
  } = useTasks();

  const pollIntervalRef = useRef<NodeJS.Timeout | null>(null);
  const isPollingRef = useRef(false);

  // 加载用户任务
  const loadTasks = useCallback(async () => {
    try {
      await loadUserTasks();
    } catch (error) {
      console.error('加载任务失败:', error);
      toast.error('加载任务失败');
    }
  }, [loadUserTasks]);

  // 轮询活跃任务状态
  const pollActiveTasks = useCallback(async () => {
    if (isPollingRef.current || activeTasks.length === 0) {
      return;
    }

    isPollingRef.current = true;

    try {
      // 并发查询所有活跃任务的状态
      const statusPromises = activeTasks.map(task =>
        taskApi.getTaskStatus(task.id).then(response => {
          // 处理API响应格式
          if (response && response.data) {
            return response.data;
          }
          return response;
        }).catch(error => {
          console.error(`查询任务 ${task.id} 状态失败:`, error);
          return null;
        })
      );

      const statusResults = await Promise.all(statusPromises);

      // 更新任务状态
      statusResults.forEach((result, index) => {
        if (result) {
          const task = activeTasks[index];

          // 解析steps和result字段（如果是字符串）
          let steps = result.steps;
          let resultData = result.result;

          if (typeof steps === 'string' && steps) {
            try {
              steps = JSON.parse(steps);
            } catch (e) {
              console.warn('解析steps失败:', e);
              steps = [];
            }
          }

          if (typeof resultData === 'string' && resultData) {
            try {
              resultData = JSON.parse(resultData);
            } catch (e) {
              console.warn('解析result失败:', e);
            }
          }

          const hasChanges =
            result.status !== task.status ||
            result.progress !== task.progress ||
            JSON.stringify(steps) !== JSON.stringify(task.steps) ||
            JSON.stringify(resultData) !== JSON.stringify(task.result);

          if (hasChanges) {
            updateTask(task.id, {
              status: result.status,
              progress: result.progress,
              steps: steps,
              result: resultData,
              error: result.error,
              updatedAt: new Date().toISOString(),
            });

            // 任务完成或失败时显示通知
            if (result.status === 'completed' && task.status !== 'completed') {
              toast.success(`任务"${task.title || task.type}"已完成！`);
            } else if (result.status === 'failed' && task.status !== 'failed') {
              toast.error(`任务"${task.title || task.type}"执行失败`);
            }
          }
        }
      });
    } catch (error) {
      console.error('轮询任务状态失败:', error);
    } finally {
      isPollingRef.current = false;
    }
  }, [activeTasks, updateTask]);

  // 开始轮询
  const startPolling = useCallback(() => {
    if (pollIntervalRef.current || !enablePolling) {
      return;
    }

    pollIntervalRef.current = setInterval(() => {
      pollActiveTasks();
    }, pollInterval);
  }, [pollActiveTasks, pollInterval, enablePolling]);

  // 停止轮询
  const stopPolling = useCallback(() => {
    if (pollIntervalRef.current) {
      clearInterval(pollIntervalRef.current);
      pollIntervalRef.current = null;
    }
  }, []);

  // 重新开始轮询
  const restartPolling = useCallback(() => {
    stopPolling();
    startPolling();
  }, [stopPolling, startPolling]);

  // 手动刷新任务
  const refreshTasks = useCallback(async () => {
    await loadTasks();
    // 立即轮询一次活跃任务
    if (enablePolling) {
      pollActiveTasks();
    }
  }, [loadTasks, pollActiveTasks, enablePolling]);

  // 初始化和清理
  useEffect(() => {
    // 自动加载任务
    if (autoLoad) {
      loadTasks();
    }

    // 开始轮询
    if (enablePolling) {
      startPolling();
    }

    // 清理函数
    return () => {
      stopPolling();
    };
  }, [autoLoad, enablePolling, loadTasks, startPolling, stopPolling]);

  // 当活跃任务变化时，重新开始轮询
  useEffect(() => {
    if (enablePolling) {
      restartPolling();
    }
  }, [activeTasks.length, enablePolling, restartPolling]);

  // 页面可见性变化时的处理
  useEffect(() => {
    const handleVisibilityChange = () => {
      if (document.hidden) {
        // 页面隐藏时停止轮询
        stopPolling();
      } else {
        // 页面显示时恢复轮询并刷新任务
        if (enablePolling) {
          refreshTasks();
          startPolling();
        }
      }
    };

    document.addEventListener('visibilitychange', handleVisibilityChange);
    return () => {
      document.removeEventListener('visibilitychange', handleVisibilityChange);
    };
  }, [enablePolling, refreshTasks, startPolling, stopPolling]);

  return {
    tasks,
    activeTasks,
    addTask,
    updateTask,
    removeTask,
    loadTasks,
    refreshTasks,
    startPolling,
    stopPolling,
    restartPolling,
    isPolling: !!pollIntervalRef.current,
  };
};

export default useTaskManager;
