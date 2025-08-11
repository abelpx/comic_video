import React, { useState } from 'react';
import { Card, Button, Space, Typography, List, Tag, Progress, message } from 'antd';
import { ReloadOutlined, PlusOutlined, DeleteOutlined } from '@ant-design/icons';
import { useTasks } from '../store';
import useTaskManager from '../hooks/useTaskManager';
import { aiApi } from '../services/api';

const { Title, Text } = Typography;

const TaskTestPage: React.FC = () => {
  const { tasks, activeTasks, addTask, removeTask } = useTasks();
  const { refreshTasks, isPolling } = useTaskManager({
    autoLoad: true,
    enablePolling: true,
    pollInterval: 2000,
  });

  const [loading, setLoading] = useState(false);

  // 生成UUID格式的ID
  const generateUUID = () => {
    return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
      const r = Math.random() * 16 | 0;
      const v = c === 'x' ? r : (r & 0x3 | 0x8);
      return v.toString(16);
    });
  };

  // 创建测试任务
  const createTestTask = () => {
    const testTask = {
      id: generateUUID(),
      title: `测试任务 ${new Date().toLocaleTimeString()}`,
      type: 'video' as const,
      status: 'processing' as const,
      progress: Math.floor(Math.random() * 80) + 10,
      steps: [
        { name: 'script_generation', status: 'completed', progress: 100 },
        { name: 'image_generation', status: 'processing', progress: 60 },
        { name: 'voice_generation', status: 'pending', progress: 0 },
        { name: 'video_composition', status: 'pending', progress: 0 },
      ],
      result: null,
      error: null,
      createdAt: new Date().toISOString(),
      updatedAt: new Date().toISOString(),
    };

    addTask(testTask);
    message.success('测试任务已创建');
  };

  // 创建真实任务
  const createRealTask = async () => {
    setLoading(true);
    try {
      const response = await aiApi.novelToAll('这是一个测试小说内容，用于验证任务恢复功能。');
      if (response && response.task_id) {
        addTask({
          id: response.task_id,
          title: '真实任务 - 小说转视频',
          type: 'video',
          status: 'processing',
          progress: 0,
          steps: [],
          result: null,
          error: null,
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        });
        message.success('真实任务已创建');
      }
    } catch (error) {
      message.error('创建真实任务失败');
    } finally {
      setLoading(false);
    }
  };

  // 获取状态颜色
  const getStatusColor = (status: string) => {
    switch (status) {
      case 'completed':
        return 'success';
      case 'processing':
        return 'processing';
      case 'failed':
        return 'error';
      default:
        return 'default';
    }
  };

  // 获取状态文本
  const getStatusText = (status: string) => {
    switch (status) {
      case 'completed':
        return '已完成';
      case 'processing':
        return '进行中';
      case 'failed':
        return '失败';
      case 'pending':
        return '等待中';
      default:
        return '未知';
    }
  };

  return (
    <div style={{ padding: 24 }}>
      <Title level={2}>任务恢复测试页面</Title>
      <Text type="secondary">
        此页面用于测试任务的持久化和恢复功能。刷新页面后，任务应该能够自动恢复并继续轮询。
      </Text>

      <Card style={{ marginTop: 24 }}>
        <Space style={{ marginBottom: 16 }}>
          <Button 
            type="primary" 
            icon={<PlusOutlined />} 
            onClick={createTestTask}
          >
            创建测试任务
          </Button>
          <Button 
            icon={<PlusOutlined />} 
            onClick={createRealTask}
            loading={loading}
          >
            创建真实任务
          </Button>
          <Button 
            icon={<ReloadOutlined />} 
            onClick={refreshTasks}
          >
            刷新任务
          </Button>
        </Space>

        <div style={{ marginBottom: 16 }}>
          <Text strong>轮询状态: </Text>
          <Tag color={isPolling ? 'green' : 'red'}>
            {isPolling ? '正在轮询' : '未轮询'}
          </Tag>
          <Text style={{ marginLeft: 16 }}>
            总任务数: {tasks.length} | 活跃任务数: {activeTasks.length}
          </Text>
        </div>

        <Title level={4}>所有任务</Title>
        <List
          dataSource={tasks}
          renderItem={(task) => (
            <List.Item
              actions={[
                <Button 
                  type="text" 
                  danger 
                  icon={<DeleteOutlined />}
                  onClick={() => removeTask(task.id)}
                >
                  删除
                </Button>
              ]}
            >
              <List.Item.Meta
                title={
                  <Space>
                    <Text strong>{task.title}</Text>
                    <Tag color={getStatusColor(task.status)}>
                      {getStatusText(task.status)}
                    </Tag>
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      {task.id}
                    </Text>
                  </Space>
                }
                description={
                  <div>
                    <Progress 
                      percent={task.progress || 0} 
                      size="small" 
                      style={{ marginBottom: 8 }}
                    />
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      创建时间: {new Date(task.createdAt).toLocaleString()}
                    </Text>
                    {task.updatedAt && (
                      <Text type="secondary" style={{ fontSize: 12, marginLeft: 16 }}>
                        更新时间: {new Date(task.updatedAt).toLocaleString()}
                      </Text>
                    )}
                  </div>
                }
              />
            </List.Item>
          )}
          locale={{ emptyText: '暂无任务' }}
        />

        <Title level={4} style={{ marginTop: 24 }}>活跃任务 (正在轮询)</Title>
        <List
          dataSource={activeTasks}
          renderItem={(task) => (
            <List.Item>
              <List.Item.Meta
                title={
                  <Space>
                    <Text strong>{task.title}</Text>
                    <Tag color="processing">正在处理</Tag>
                  </Space>
                }
                description={
                  <div>
                    <Progress 
                      percent={task.progress || 0} 
                      size="small" 
                      status="active"
                    />
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      步骤: {task.steps?.length || 0} | 
                      最后更新: {task.updatedAt ? new Date(task.updatedAt).toLocaleString() : '未知'}
                    </Text>
                  </div>
                }
              />
            </List.Item>
          )}
          locale={{ emptyText: '暂无活跃任务' }}
        />
      </Card>

      <Card style={{ marginTop: 24 }} title="使用说明">
        <ol>
          <li>点击"创建测试任务"创建一个模拟任务</li>
          <li>点击"创建真实任务"创建一个真实的API任务</li>
          <li>观察任务列表和活跃任务列表</li>
          <li><strong>刷新页面</strong>，观察任务是否能够恢复</li>
          <li>活跃任务应该继续自动轮询更新状态</li>
        </ol>
      </Card>
    </div>
  );
};

export default TaskTestPage;
