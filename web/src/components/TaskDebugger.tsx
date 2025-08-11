import React from 'react';
import { Card, Typography, Tag, Space, Button } from 'antd';
import { useTasks, useAuth } from '../store';
import useTaskManager from '../hooks/useTaskManager';

const { Text, Title } = Typography;

const TaskDebugger: React.FC = () => {
  const { isAuthenticated, user, login } = useAuth();
  const { tasks, activeTasks } = useTasks();
  const { refreshTasks, isPolling } = useTaskManager({
    autoLoad: false,
    enablePolling: false,
  });

  // 模拟登录
  const handleTestLogin = () => {
    login(
      {
        id: '1',
        username: 'testuser',
        email: 'test@example.com',
        nickname: 'Test User',
        avatar: '',
        status: 'active',
        createdAt: new Date().toISOString(),
        updatedAt: new Date().toISOString(),
      },
      'test-token-123'
    );
  };

  return (
    <Card 
      title="任务调试信息" 
      size="small" 
      style={{ 
        position: 'fixed', 
        top: 10, 
        right: 10, 
        width: 300, 
        zIndex: 1000,
        fontSize: 12 
      }}
    >
      <Space direction="vertical" size="small" style={{ width: '100%' }}>
        <div>
          <Text strong>认证状态: </Text>
          <Tag color={isAuthenticated ? 'green' : 'red'}>
            {isAuthenticated ? '已登录' : '未登录'}
          </Tag>
        </div>
        
        <div>
          <Text strong>用户: </Text>
          <Text>{user?.username || '无'}</Text>
        </div>
        
        <div>
          <Text strong>任务总数: </Text>
          <Tag color="blue">{tasks.length}</Tag>
        </div>
        
        <div>
          <Text strong>活跃任务: </Text>
          <Tag color="orange">{activeTasks.length}</Tag>
        </div>
        
        <div>
          <Text strong>轮询状态: </Text>
          <Tag color={isPolling ? 'green' : 'red'}>
            {isPolling ? '运行中' : '已停止'}
          </Tag>
        </div>
        
        <Space size="small" style={{ width: '100%' }}>
          <Button size="small" onClick={refreshTasks} style={{ flex: 1 }}>
            刷新任务
          </Button>
          {!isAuthenticated && (
            <Button size="small" onClick={handleTestLogin} type="primary" style={{ flex: 1 }}>
              测试登录
            </Button>
          )}
        </Space>
        
        {tasks.length > 0 && (
          <div>
            <Text strong>任务列表:</Text>
            <div style={{ maxHeight: 100, overflow: 'auto', fontSize: 10 }}>
              {tasks.map((task, index) => (
                <div key={task.id} style={{ marginBottom: 4 }}>
                  <Text>{index + 1}. {task.title}</Text>
                  <br />
                  <Tag size="small" color={
                    task.status === 'completed' ? 'green' :
                    task.status === 'processing' ? 'orange' :
                    task.status === 'failed' ? 'red' : 'default'
                  }>
                    {task.status}
                  </Tag>
                  <Text style={{ fontSize: 10 }}> {task.progress}%</Text>
                </div>
              ))}
            </div>
          </div>
        )}
      </Space>
    </Card>
  );
};

export default TaskDebugger;
