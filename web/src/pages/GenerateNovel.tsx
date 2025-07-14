import React, { useState } from 'react';
import { Input, Button, Card, message } from 'antd';
import axios from 'axios';

export default function GenerateNovel() {
  const [topic, setTopic] = useState('');
  const [length, setLength] = useState(300);
  const [novel, setNovel] = useState('');
  const [loading, setLoading] = useState(false);

  const handleGenerate = async () => {
    if (!topic) return message.error('请输入小说主题');
    setLoading(true);
    setNovel('');
    try {
      const res = await axios.post('/api/v1/ai/generate-novel', { topic, length });
      setNovel(res.data.novel);
    } catch {
      setNovel('生成失败');
    }
    setLoading(false);
  };

  return (
    <Card title="AI小说生成" style={{ maxWidth: 700, margin: '0 auto', marginTop: 32 }}>
      <Input
        placeholder="请输入小说主题"
        value={topic}
        onChange={e => setTopic(e.target.value)}
        style={{ marginBottom: 12 }}
      />
      <Input
        type="number"
        min={100}
        max={2000}
        value={length}
        onChange={e => setLength(Number(e.target.value))}
        style={{ width: 120, marginRight: 12 }}
      />
      <Button type="primary" onClick={handleGenerate} loading={loading}>
        生成小说
      </Button>
      <div style={{ marginTop: 24, minHeight: 120, whiteSpace: 'pre-wrap' }}>{novel}</div>
    </Card>
  );
} 