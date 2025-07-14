import React, { useState } from 'react';
import { Input, Button, Card, message } from 'antd';
import axios from 'axios';

export default function GenerateTweet() {
  const [topic, setTopic] = useState('');
  const [tweet, setTweet] = useState('');
  const [loading, setLoading] = useState(false);

  const handleGenerate = async () => {
    if (!topic) return message.error('请输入推文主题');
    setLoading(true);
    setTweet('');
    try {
      const res = await axios.post('/api/v1/ai/generate-tweet', { topic });
      setTweet(res.data.tweet);
    } catch {
      setTweet('生成失败');
    }
    setLoading(false);
  };

  return (
    <Card title="AI推文生成" style={{ maxWidth: 600, margin: '0 auto', marginTop: 32 }}>
      <Input
        placeholder="请输入推文主题"
        value={topic}
        onChange={e => setTopic(e.target.value)}
        style={{ marginBottom: 12 }}
      />
      <Button type="primary" onClick={handleGenerate} loading={loading}>
        生成推文
      </Button>
      <div style={{ marginTop: 24, minHeight: 60 }}>{tweet}</div>
    </Card>
  );
} 