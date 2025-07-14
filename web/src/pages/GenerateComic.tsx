import React, { useState } from 'react';
import { Input, Button, Spin, Row, Col, Card, message } from 'antd';
import axios from 'axios';

export default function GenerateComic() {
  const [theme, setTheme] = useState('');
  const [panelCount, setPanelCount] = useState(4);
  const [loading, setLoading] = useState(false);
  const [script, setScript] = useState('');
  const [images, setImages] = useState<string[]>([]);

  const handleGenerate = async () => {
    if (!theme) return message.error('请输入主题');
    setLoading(true);
    setScript('');
    setImages([]);
    try {
      const res = await axios.post('/api/v1/ai/generate-comic', {
        theme,
        panel_count: panelCount,
      });
      setScript(res.data.script);
      setImages(res.data.images);
    } catch {
      message.error('生成失败');
    }
    setLoading(false);
  };

  return (
    <div style={{ maxWidth: 800, margin: '0 auto', padding: 24 }}>
      <h2>AI漫画生成</h2>
      <Input
        placeholder="请输入漫画主题"
        value={theme}
        onChange={e => setTheme(e.target.value)}
        style={{ marginBottom: 12 }}
      />
      <Input
        type="number"
        min={1}
        max={10}
        value={panelCount}
        onChange={e => setPanelCount(Number(e.target.value))}
        style={{ width: 120, marginRight: 12 }}
      />
      <Button type="primary" onClick={handleGenerate} loading={loading}>
        生成漫画
      </Button>
      <div style={{ marginTop: 24 }}>
        {loading && <Spin />}
        {script && (
          <Card title="分镜脚本" style={{ marginBottom: 24 }}>
            <pre>{script}</pre>
          </Card>
        )}
        <Row gutter={16}>
          {images.map((img, idx) =>
            img ? (
              <Col span={6} key={idx} style={{ marginBottom: 16 }}>
                <img
                  src={`data:image/png;base64,${img}`}
                  alt={`panel-${idx + 1}`}
                  style={{ width: '100%', border: '1px solid #eee' }}
                />
              </Col>
            ) : null
          )}
        </Row>
      </div>
    </div>
  );
} 