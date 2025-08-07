import React, { useState, useRef } from 'react';
import { Button, Input, Typography, Progress, message, Card, Spin, Tabs, Row, Col } from 'antd';
import { RobotOutlined, MonitorOutlined, HistoryOutlined } from '@ant-design/icons';
import { aiApi, taskApi } from '../services/api';
import TaskProgress from '../components/TaskProgress';
import SmartCreationAssistant from '../components/SmartCreationAssistant';
import RealTimeMonitor from '../components/RealTimeMonitor';

const { Title, Paragraph } = Typography;
const { TabPane } = Tabs;

export default function NovelToVideoNew() {
  const [novel, setNovel] = useState('');
  const [taskId, setTaskId] = useState('');
  const [progress, setProgress] = useState(0);
  const [status, setStatus] = useState('');
  const [result, setResult] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [taskSteps, setTaskSteps] = useState<any[]>([]);
  const [taskError, setTaskError] = useState<string>('');
  const timerRef = useRef<any>(null);

  const handleSubmit = async () => {
    if (!novel.trim()) {
      message.warning('请输入小说内容');
      return;
    }
    setLoading(true);
    setResult(null);
    setProgress(0);
    setStatus('');
    setTaskSteps([]);
    setTaskError('');
    
    try {
      const response = await aiApi.novelToAll(novel);
      console.log('API Response:', response);
      if (response && response.task_id) {
        setTaskId(response.task_id);
        pollStatus(response.task_id);
        message.success('🚀 任务已提交，AI正在为您创作精彩视频...');
      } else {
        setLoading(false);
        message.error('任务提交失败: 未返回任务ID');
      }
    } catch (error: any) {
      console.error('API Error:', error);
      setLoading(false);
      const errorMessage = error.response?.data?.message || error.message || '任务提交异常';
      message.error(`任务提交失败: ${errorMessage}`);
    }
  };

  const handleSmartSubmit = async (content: string, enhancements: any) => {
    setNovel(content);
    console.log('智能增强配置:', enhancements);
    
    setLoading(true);
    setResult(null);
    setProgress(0);
    setStatus('');
    setTaskSteps([]);
    setTaskError('');
    
    try {
      const response = await aiApi.novelToAll(content);
      if (response && response.task_id) {
        setTaskId(response.task_id);
        pollStatus(response.task_id);
        message.success('🎯 智能优化任务已提交，预计生成更高质量的视频！');
      } else {
        setLoading(false);
        message.error('任务提交失败: 未返回任务ID');
      }
    } catch (error: any) {
      setLoading(false);
      const errorMessage = error.response?.data?.message || error.message || '任务提交异常';
      message.error(`任务提交失败: ${errorMessage}`);
    }
  };

  const pollStatus = (id: string) => {
    timerRef.current = setInterval(async () => {
      try {
        const taskStatus = await taskApi.getTaskStatus(id);
        console.log('Task Status:', taskStatus);
        if (taskStatus) {
          setProgress(taskStatus.progress || 0);
          setStatus(taskStatus.status);
          
          if (taskStatus.steps) {
            try {
              const steps = typeof taskStatus.steps === 'string' 
                ? JSON.parse(taskStatus.steps) 
                : taskStatus.steps;
              setTaskSteps(steps);
            } catch (e) {
              console.error('解析步骤信息失败:', e);
            }
          }
          
          if (taskStatus.status === 'completed') {
            setLoading(false);
            clearInterval(timerRef.current!);
            setResult(taskStatus.result ? JSON.parse(taskStatus.result) : null);
            message.success('🎉 任务完成！您的AI视频已生成完毕！');
          } else if (taskStatus.status === 'failed') {
            setLoading(false);
            clearInterval(timerRef.current!);
            setTaskError(taskStatus.error || '生成失败');
            message.error(taskStatus.error || '生成失败');
          }
        }
      } catch (error: any) {
        console.error('Poll Status Error:', error);
        setLoading(false);
        clearInterval(timerRef.current!);
        message.error('进度查询失败');
      }
    }, 2000);
  };

  React.useEffect(() => {
    return () => {
      if (timerRef.current) clearInterval(timerRef.current);
    };
  }, []);

  return (
    <div style={{ padding: '20px', minHeight: '100vh', background: '#f5f5f5' }}>
      <div style={{ maxWidth: 1400, margin: '0 auto' }}>
        {/* 顶部标题 */}
        <Card style={{ marginBottom: 24, textAlign: 'center' }}>
          <Title level={1} style={{ 
            margin: 0, 
            background: 'linear-gradient(45deg, #667eea, #764ba2)', 
            WebkitBackgroundClip: 'text', 
            WebkitTextFillColor: 'transparent' 
          }}>
            🎬 VidCraft Studio Pro
          </Title>
          <Paragraph style={{ fontSize: 16, color: '#666', margin: '8px 0 0 0' }}>
            行业领先的AI视频生成平台 · 智能创作 · 实时监控 · 专业品质
          </Paragraph>
        </Card>

        {/* 主要功能区域 */}
        <Tabs defaultActiveKey="smart" size="large">
          <TabPane 
            tab={
              <span>
                <RobotOutlined />
                智能创作
              </span>
            } 
            key="smart"
          >
            <SmartCreationAssistant 
              onSubmit={handleSmartSubmit}
              loading={loading}
            />
            
            {(loading || taskId) && (
              <div style={{ marginTop: 24 }}>
                <TaskProgress
                  taskId={taskId}
                  status={status}
                  progress={progress}
                  steps={taskSteps}
                  result={result}
                  error={taskError}
                />
              </div>
            )}
          </TabPane>

          <TabPane 
            tab={
              <span>
                <MonitorOutlined />
                实时监控
              </span>
            } 
            key="monitor"
          >
            <RealTimeMonitor />
          </TabPane>

          <TabPane 
            tab={
              <span>
                <HistoryOutlined />
                历史记录
              </span>
            } 
            key="history"
          >
            <Card>
              <div style={{ textAlign: 'center', padding: 40 }}>
                <Title level={3}>历史记录</Title>
                <Paragraph type="secondary">
                  这里将显示您的创作历史和作品管理功能
                </Paragraph>
              </div>
            </Card>
          </TabPane>
        </Tabs>
      </div>
    </div>
  );
}
