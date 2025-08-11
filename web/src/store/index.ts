import { create } from 'zustand';
import { devtools, persist } from 'zustand/middleware';

// 用户状态接口
interface User {
  id: string;
  username: string;
  email: string;
  nickname: string;
  avatar?: string;
  subscription?: {
    plan: string;
    status: string;
    expiresAt: string;
  };
}

// 配额信息接口
interface Quota {
  type: string;
  limit: number;
  used: number;
  remaining: number;
  resetTime: string;
}

// 任务状态接口
interface Task {
  id: string;
  type: string;
  status: 'pending' | 'processing' | 'completed' | 'failed';
  progress: number;
  result?: any;
  error?: string;
  createdAt: string;
}

// 应用状态接口
interface AppState {
  // 用户相关
  user: User | null;
  isAuthenticated: boolean;
  token: string | null;
  
  // 配额相关
  quotas: Record<string, Quota>;
  
  // 任务相关
  tasks: Task[];
  activeTasks: Task[];
  
  // UI状态
  loading: boolean;
  sidebarCollapsed: boolean;
  theme: 'light' | 'dark';
  
  // Actions
  setUser: (user: User | null) => void;
  setToken: (token: string | null) => void;
  login: (user: User, token: string) => void;
  logout: () => void;
  
  setQuotas: (quotas: Record<string, Quota>) => void;
  updateQuota: (type: string, quota: Quota) => void;
  
  addTask: (task: Task) => void;
  updateTask: (taskId: string, updates: Partial<Task>) => void;
  removeTask: (taskId: string) => void;
  loadUserTasks: () => Promise<void>;
  setTasks: (tasks: Task[]) => void;
  
  setLoading: (loading: boolean) => void;
  setSidebarCollapsed: (collapsed: boolean) => void;
  setTheme: (theme: 'light' | 'dark') => void;
}

// 创建状态管理
export const useAppStore = create<AppState>()(
  devtools(
    persist(
      (set, get) => ({
        // 初始状态
        user: null,
        isAuthenticated: false,
        token: null,
        quotas: {},
        tasks: [],
        activeTasks: [],
        loading: false,
        sidebarCollapsed: false,
        theme: 'light',

        // 用户相关actions
        setUser: (user) => set({ user, isAuthenticated: !!user }),
        
        setToken: (token) => set({ token }),
        
        login: (user, token) => set({ 
          user, 
          token, 
          isAuthenticated: true 
        }),
        
        logout: () => set({ 
          user: null, 
          token: null, 
          isAuthenticated: false,
          quotas: {},
          tasks: [],
          activeTasks: []
        }),

        // 配额相关actions
        setQuotas: (quotas) => set({ quotas }),
        
        updateQuota: (type, quota) => set((state) => ({
          quotas: { ...state.quotas, [type]: quota }
        })),

        // 任务相关actions
        addTask: (task) => set((state) => {
          const newTasks = [...state.tasks, task];
          const newActiveTasks = task.status === 'pending' || task.status === 'processing' 
            ? [...state.activeTasks, task] 
            : state.activeTasks;
          return { tasks: newTasks, activeTasks: newActiveTasks };
        }),
        
        updateTask: (taskId, updates) => set((state) => {
          const updatedTasks = state.tasks.map(task => 
            task.id === taskId ? { ...task, ...updates } : task
          );
          const updatedActiveTasks = state.activeTasks.map(task => 
            task.id === taskId ? { ...task, ...updates } : task
          ).filter(task => task.status === 'pending' || task.status === 'processing');
          
          return { tasks: updatedTasks, activeTasks: updatedActiveTasks };
        }),
        
        removeTask: (taskId) => set((state) => ({
          tasks: state.tasks.filter(task => task.id !== taskId),
          activeTasks: state.activeTasks.filter(task => task.id !== taskId)
        })),

        setTasks: (tasks) => set({ tasks }),

        loadUserTasks: async () => {
          try {
            const { taskApi } = await import('../services/api');
            const response = await taskApi.getUserTasks(1, 50);

            // 处理API响应格式
            let tasksData = [];
            if (response && response.data && response.data.tasks) {
              tasksData = response.data.tasks;
            } else if (response && response.tasks) {
              tasksData = response.tasks;
            }

            // 转换任务数据格式
            const tasks = tasksData.map((task: any) => {
              // 解析steps和result字段（如果是字符串）
              let steps = task.steps || [];
              let result = task.result;

              if (typeof steps === 'string' && steps) {
                try {
                  steps = JSON.parse(steps);
                } catch (e) {
                  console.warn('解析steps失败:', e);
                  steps = [];
                }
              }

              if (typeof result === 'string' && result) {
                try {
                  result = JSON.parse(result);
                } catch (e) {
                  console.warn('解析result失败:', e);
                }
              }

              return {
                id: task.id,
                title: task.title || `${task.type}任务`,
                type: task.type,
                status: task.status,
                progress: task.progress || 0,
                steps: steps,
                result: result,
                error: task.error || null,
                createdAt: task.created_at || task.createdAt || new Date().toISOString(),
                updatedAt: task.updated_at || task.updatedAt || new Date().toISOString(),
              };
            });

            // 更新任务列表
            set({ tasks });

            // 找出正在进行的任务
            const activeTasks = tasks.filter((task: Task) =>
              task.status === 'processing' || task.status === 'pending'
            );
            set({ activeTasks });

            return tasks;
          } catch (error) {
            console.error('加载用户任务失败:', error);
            return [];
          }
        },

        // UI状态actions
        setLoading: (loading) => set({ loading }),
        setSidebarCollapsed: (sidebarCollapsed) => set({ sidebarCollapsed }),
        setTheme: (theme) => set({ theme }),
      }),
      {
        name: 'comic-video-store',
        partialize: (state) => ({
          user: state.user,
          token: state.token,
          isAuthenticated: state.isAuthenticated,
          theme: state.theme,
          sidebarCollapsed: state.sidebarCollapsed,
          // 持久化任务数据
          tasks: state.tasks,
          activeTasks: state.activeTasks,
        }),
      }
    ),
    { name: 'comic-video-store' }
  )
);

// 选择器hooks
export const useUser = () => useAppStore((state) => state.user);
export const useAuth = () => useAppStore((state) => ({
  isAuthenticated: state.isAuthenticated,
  token: state.token,
  login: state.login,
  logout: state.logout,
}));
export const useQuotas = () => useAppStore((state) => state.quotas);
export const useTasks = () => useAppStore((state) => ({
  tasks: state.tasks,
  activeTasks: state.activeTasks,
  addTask: state.addTask,
  updateTask: state.updateTask,
  removeTask: state.removeTask,
  loadUserTasks: state.loadUserTasks,
  setTasks: state.setTasks,
}));
export const useUI = () => useAppStore((state) => ({
  loading: state.loading,
  sidebarCollapsed: state.sidebarCollapsed,
  theme: state.theme,
  setLoading: state.setLoading,
  setSidebarCollapsed: state.setSidebarCollapsed,
  setTheme: state.setTheme,
}));
