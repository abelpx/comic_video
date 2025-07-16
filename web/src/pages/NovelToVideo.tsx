import React, { useState, useRef } from 'react';
import { Button, Input, Typography, Progress, message, Card, Spin } from 'antd';
import axios from 'axios';

const { Title, Paragraph } = Typography;

export default function NovelToVideo() {
  const [novel, setNovel] = useState('');
  const [taskId, setTaskId] = useState('');
  const [progress, setProgress] = useState(0);
  const [status, setStatus] = useState('');
  const [result, setResult] = useState<any>(null);
  const [loading, setLoading] = useState(false);
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
    try {
      const res = await axios.post('/api/v1/ai/novel-to-all', { novel_prompt: novel });
      if (res.data && res.data.task_id) {
        setTaskId(res.data.task_id);
        pollStatus(res.data.task_id);
        message.success('任务已提交，正在生成...');
      } else {
        setLoading(false);
        message.error('任务提交失败');
      }
    } catch (e) {
      setLoading(false);
      message.error('任务提交异常');
    }
  };

  const pollStatus = (id: string) => {
    timerRef.current = setInterval(async () => {
      try {
        const res = await axios.get(`/api/v1/task/${id}/status`);
        if (res.data && res.data.status) {
          setProgress(res.data.progress || 0);
          setStatus(res.data.status);
          if (res.data.status === 'completed') {
            setLoading(false);
            clearInterval(timerRef.current!);
            setResult(res.data.result ? JSON.parse(res.data.result) : null);
          } else if (res.data.status === 'failed') {
            setLoading(false);
            clearInterval(timerRef.current!);
            message.error(res.data.error || '生成失败');
          }
        }
      } catch {
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
      {loading && (
        <div style={{ margin: '16px 0' }}>
          <Spin />
          <Progress percent={progress} status={status === 'failed' ? 'exception' : 'active'} />
        </div>
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