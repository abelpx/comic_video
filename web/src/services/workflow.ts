import { api } from './api';

export interface WorkflowStep {
  key: string;
  title: string;
  description: string;
  status: 'wait' | 'process' | 'finish' | 'error';
}

export interface CreateWorkflowRequest {
  project_id: string;
  name: string;
  description: string;
  steps: string[];
  config?: string;
}

export interface NovelToVideoWorkflowRequest {
  project_id: string;
  novel_text: string;
  title: string;
  description?: string;
  settings: {
    video_theme?: string;
    target_audience?: string;
    content_type?: string;
    auto_publish?: boolean;
  };
}

export interface WorkflowResponse {
  id: string;
  project_id: string;
  user_id: string;
  name: string;
  description: string;
  status: 'pending' | 'running' | 'completed' | 'failed' | 'cancelled';
  current_step: string;
  progress: number;
  started_at?: string;
  completed_at?: string;
  created_at: string;
  updated_at: string;
}

export interface WorkflowTask {
  id: string;
  workflow_id: string;
  step: string;
  name: string;
  status: 'pending' | 'running' | 'completed' | 'failed';
  progress: number;
  started_at?: string;
  completed_at?: string;
  error?: string;
}

export interface WorkflowListResponse {
  workflows: WorkflowResponse[];
  total: number;
  page: number;
  page_size: number;
}

class WorkflowService {
  // 创建工作流
  async createWorkflow(data: CreateWorkflowRequest): Promise<WorkflowResponse> {
    const response = await api.post('/workflows', data);
    return response.data.data;
  }

  // 启动工作流
  async startWorkflow(workflowId: string): Promise<void> {
    await api.post(`/workflows/${workflowId}/start`);
  }

  // 获取工作流详情
  async getWorkflow(workflowId: string): Promise<WorkflowResponse> {
    const response = await api.get(`/workflows/${workflowId}`);
    return response.data.data;
  }

  // 获取工作流列表
  async getWorkflows(page = 1, pageSize = 10): Promise<WorkflowListResponse> {
    const response = await api.get('/workflows', {
      params: { page, page_size: pageSize }
    });
    return response.data.data;
  }

  // 小说转视频完整工作流
  async createNovelToVideoWorkflow(data: NovelToVideoWorkflowRequest): Promise<{
    workflow_id: string;
    status: string;
    steps: number;
  }> {
    const response = await api.post('/workflows/novel-to-video', data);
    return response.data.data;
  }

  // 获取工作流任务列表
  async getWorkflowTasks(workflowId: string): Promise<WorkflowTask[]> {
    const response = await api.get(`/workflows/${workflowId}/tasks`);
    return response.data.data;
  }

  // 取消工作流
  async cancelWorkflow(workflowId: string): Promise<void> {
    await api.post(`/workflows/${workflowId}/cancel`);
  }

  // 重试失败的工作流
  async retryWorkflow(workflowId: string): Promise<void> {
    await api.post(`/workflows/${workflowId}/retry`);
  }

  // 获取工作流进度
  async getWorkflowProgress(workflowId: string): Promise<{
    workflow: WorkflowResponse;
    tasks: WorkflowTask[];
    overall_progress: number;
  }> {
    const response = await api.get(`/workflows/${workflowId}/progress`);
    return response.data.data;
  }

  // 轮询工作流状态
  async pollWorkflowStatus(
    workflowId: string,
    onUpdate: (data: { workflow: WorkflowResponse; tasks: WorkflowTask[] }) => void,
    interval = 2000
  ): Promise<() => void> {
    let isPolling = true;
    
    const poll = async () => {
      if (!isPolling) return;
      
      try {
        const data = await this.getWorkflowProgress(workflowId);
        onUpdate({
          workflow: data.workflow,
          tasks: data.tasks
        });
        
        // 如果工作流已完成或失败，停止轮询
        if (data.workflow.status === 'completed' || data.workflow.status === 'failed') {
          isPolling = false;
          return;
        }
        
        setTimeout(poll, interval);
      } catch (error) {
        console.error('轮询工作流状态失败:', error);
        setTimeout(poll, interval * 2); // 错误时延长间隔
      }
    };
    
    poll();
    
    // 返回停止轮询的函数
    return () => {
      isPolling = false;
    };
  }
}

export const workflowService = new WorkflowService();
