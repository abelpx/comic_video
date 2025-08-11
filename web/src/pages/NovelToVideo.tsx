import React, { useState, useRef } from 'react';
import { Button, Input, Typography, Progress, message, Card, Spin, Tabs, Row, Col } from 'antd';
import { RobotOutlined, MonitorOutlined, HistoryOutlined } from '@ant-design/icons';
import { aiApi } from '../services/api';
import { useTasks } from '../store';
import TaskProgress from '../components/TaskProgress';
import SmartCreationAssistant from '../components/SmartCreationAssistant';
import RealTimeMonitor from '../components/RealTimeMonitor';
import useTaskManager from '../hooks/useTaskManager';

const { Title, Paragraph } = Typography;
const { TabPane } = Tabs;

export default function NovelToVideo() {
  const [novel, setNovel] = useState('');
  const [taskId, setTaskId] = useState('');
  const [progress, setProgress] = useState(0);
  const [status, setStatus] = useState('');
  const [result, setResult] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [taskSteps, setTaskSteps] = useState<any[]>([]);
  const [taskError, setTaskError] = useState<string>('');
  const timerRef = useRef<any>(null);

  // 使用任务管理器
  const { addTask } = useTasks();
  const { } = useTaskManager({
    autoLoad: false,
    enablePolling: true,
  });

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
      console.log('API Response:', response); // 调试日志
      if (response && response.task_id) {
        setTaskId(response.task_id);

        // 添加任务到全局状态管理
        addTask({
          id: response.task_id,
          title: `小说转视频 - ${novel.substring(0, 20)}...`,
          type: 'video',
          status: 'processing',
          progress: 0,
          steps: [],
          result: null,
          error: null,
          createdAt: new Date().toISOString(),
          updatedAt: new Date().toISOString(),
        });

        pollStatus(response.task_id);
        message.success('🚀 任务已提交，AI正在为您创作精彩视频...');
      } else {
        setLoading(false);
        message.error('任务提交失败: 未返回任务ID');
      }
    } catch (error: any) {
      console.error('API Error:', error); // 调试日志
      setLoading(false);
      const errorMessage = error.response?.data?.message || error.message || '任务提交异常';
      message.error(`任务提交失败: ${errorMessage}`);
    }
  };

  const handleSmartSubmit = async (content: string, enhancements: any) => {
    setNovel(content);
    console.log('智能增强配置:', enhancements);

    // 使用增强配置提交任务
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
        console.log('Task Status:', taskStatus); // 调试日志
        if (taskStatus) {
          setProgress(taskStatus.progress || 0);
          setStatus(taskStatus.status);

          // 解析步骤信息
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
            message.success('任务完成！');
          } else if (taskStatus.status === 'failed') {
            setLoading(false);
            clearInterval(timerRef.current!);
            setTaskError(taskStatus.error || '生成失败');
            message.error(taskStatus.error || '生成失败');
          }
        }
      } catch (error: any) {
        console.error('Poll Status Error:', error); // 调试日志
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
    <Card style={{ maxWidth: 800, margin: '32px auto' }}>
      <Title level={3}>小说一键生成漫画、推文、动漫视频</Title>
      <Paragraph>粘贴你的小说内容，点击“一键生成”，AI将自动生成分镜、漫画图片、推文、配音和动漫视频。</Paragraph>
      <Input.TextArea
        rows={8}
        value={novel}
        onChange={e => setNovel(e.target.value)}
        placeholder="请输入小说内容..."
        disabled={loading}
        style={{ marginBottom: 16 }}
      />
      <Button type="primary" onClick={handleSubmit} loading={loading} style={{ marginBottom: 16 }}>
        一键生成
      </Button>
      {(loading || taskId) && (
        <TaskProgress
          taskId={taskId}
          status={status}
          progress={progress}
          steps={taskSteps}
          result={result}
          error={taskError}
        />
      )}
      {result && (
        <div style={{ marginTop: 24 }}>
          <Title level={4}>生成结果</Title>
          {result.images && result.images.length > 0 && (
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8, marginBottom: 16 }}>
              {result.images.map((img: string, idx: number) => (
                <img key={idx} src={`data:image/png;base64,${img}`} alt={`panel-${idx+1}`} style={{ width: 120, borderRadius: 4 }} />
              ))}
            </div>
          )}
          {result.panels && (
            <div style={{ marginBottom: 16 }}>
              <Title level={5}>分镜脚本</Title>
              <ol>
                {result.panels.map((p: string, idx: number) => <li key={idx}>{p}</li>)}
              </ol>
            </div>
          )}
          {result.url && (
            <div style={{ marginBottom: 16 }}>
              <Title level={5}>动漫视频</Title>
              <video src={result.url} controls style={{ width: 400, borderRadius: 4 }} />
              <div><a href={result.url} download>下载视频</a></div>
            </div>
          )}
        </div>
      )}
    </Card>
  );
} 