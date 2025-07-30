import axios, { AxiosInstance, AxiosRequestConfig, AxiosResponse } from 'axios';
import toast from 'react-hot-toast';
import { useAppStore } from '../store';

// API响应接口
interface ApiResponse<T = any> {
  code: number;
  message: string;
  data: T;
  timestamp?: string;
}

// 获取API基础URL
const getApiBaseUrl = (): string => {
  // 在开发环境和生产环境都使用相对路径
  // 通过Vite代理或nginx代理来访问后端
  return '/api/v1';
};

// 创建axios实例
const createApiClient = (): AxiosInstance => {
  const client = axios.create({
    baseURL: getApiBaseUrl(),
    timeout: 30000,
    headers: {
      'Content-Type': 'application/json',
    },
  });

  // 请求拦截器
  client.interceptors.request.use(
    (config) => {
      const token = useAppStore.getState().token;
      if (token) {
        config.headers.Authorization = `Bearer ${token}`;
      }

      // 添加用户ID到header（临时方案）
      const user = useAppStore.getState().user;
      if (user) {
        config.headers['X-User-ID'] = user.id;
      } else {
        // 为匿名用户添加默认ID
        config.headers['X-User-ID'] = 'anonymous';
      }

      return config;
    },
    (error) => {
      return Promise.reject(error);
    }
  );

  // 响应拦截器
  client.interceptors.response.use(
    (response: AxiosResponse<ApiResponse>) => {
      return response;
    },
    (error) => {
      const { response } = error;
      
      if (response?.status === 401) {
        // 未授权，清除登录状态
        useAppStore.getState().logout();
        toast.error('登录已过期，请重新登录');
        window.location.href = '/login';
      } else if (response?.status === 403) {
        toast.error('权限不足');
      } else if (response?.status === 429) {
        toast.error('请求过于频繁，请稍后再试');
      } else if (response?.status >= 500) {
        toast.error('服务器错误，请稍后再试');
      } else if (response?.data?.message) {
        toast.error(response.data.message);
      } else {
        toast.error('网络错误，请检查网络连接');
      }
      
      return Promise.reject(error);
    }
  );

  return client;
};

// API客户端实例
export const apiClient = createApiClient();

// 通用API方法
export const api = {
  // GET请求
  get: <T = any>(url: string, config?: AxiosRequestConfig): Promise<T> =>
    apiClient.get<ApiResponse<T>>(url, config).then(res => res.data.data),

  // POST请求
  post: <T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> =>
    apiClient.post<ApiResponse<T>>(url, data, config).then(res => res.data.data),

  // PUT请求
  put: <T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> =>
    apiClient.put<ApiResponse<T>>(url, data, config).then(res => res.data.data),

  // DELETE请求
  delete: <T = any>(url: string, config?: AxiosRequestConfig): Promise<T> =>
    apiClient.delete<ApiResponse<T>>(url, config).then(res => res.data.data),

  // PATCH请求
  patch: <T = any>(url: string, data?: any, config?: AxiosRequestConfig): Promise<T> =>
    apiClient.patch<ApiResponse<T>>(url, data, config).then(res => res.data.data),
};

// 认证相关API
export const authApi = {
  // 登录
  login: (credentials: { username: string; password: string }) =>
    api.post<{ user: any; token: string }>('/auth/login', credentials),

  // 注册
  register: (userData: { username: string; email: string; password: string; nickname?: string }) =>
    api.post<{ user: any; token: string }>('/auth/register', userData),

  // 获取用户信息
  getProfile: () =>
    api.get<any>('/auth/profile'),

  // 更新用户信息
  updateProfile: (data: any) =>
    api.put<any>('/auth/profile', data),

  // 登出
  logout: () =>
    api.post('/auth/logout'),
};

// AI相关API
export const aiApi = {
  // 小说转视频
  novelToVideo: (novel: string) =>
    api.post<{ task_id: string }>('/ai/novel-to-video', { novel }),

  // 小说转全部
  novelToAll: (novel_prompt: string, title?: string) =>
    api.post<{ task_id: string }>('/ai/novel-to-all', { novel_prompt, title }),

  // 生成小说
  generateNovel: (novel_prompt: string, title?: string) =>
    api.post<{ task_id: string }>('/ai/generate-novel', { novel_prompt, title }),

  // 获取用户配额
  getUserQuota: () =>
    api.get<Record<string, any>>('/ai/quota'),

  // 获取使用统计
  getUsageStats: () =>
    api.get<any>('/ai/usage-stats'),
};

// 任务相关API
export const taskApi = {
  // 获取任务状态
  getTaskStatus: (taskId: string) =>
    api.get<any>(`/task/${taskId}/status`),

  // 获取用户任务列表
  getUserTasks: (page = 1, limit = 20) =>
    api.get<{ tasks: any[]; total: number; page: number; limit: number }>(`/tasks?page=${page}&limit=${limit}`),
};

// 项目相关API
export const projectApi = {
  // 获取项目列表
  getProjects: (page = 1, limit = 20) =>
    api.get<{ projects: any[]; total: number }>(`/projects?page=${page}&limit=${limit}`),

  // 创建项目
  createProject: (data: any) =>
    api.post<any>('/projects', data),

  // 获取项目详情
  getProject: (id: string) =>
    api.get<any>(`/projects/${id}`),

  // 更新项目
  updateProject: (id: string, data: any) =>
    api.put<any>(`/projects/${id}`, data),

  // 删除项目
  deleteProject: (id: string) =>
    api.delete(`/projects/${id}`),

  // 分享项目
  shareProject: (id: string, data: any) =>
    api.post<any>(`/projects/${id}/share`, data),
};

// 用户相关API
export const userApi = {
  // 获取用户列表（管理员）
  getUsers: (page = 1, limit = 20) =>
    api.get<{ users: any[]; total: number }>(`/users?page=${page}&limit=${limit}`),

  // 获取用户详情
  getUser: (id: string) =>
    api.get<any>(`/users/${id}`),

  // 更新用户
  updateUser: (id: string, data: any) =>
    api.put<any>(`/users/${id}`, data),

  // 删除用户
  deleteUser: (id: string) =>
    api.delete(`/users/${id}`),
};

export default api;
