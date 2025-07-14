import React, { useState } from 'react';
import { Upload, Button, Card, Spin, message } from 'antd';
import { UploadOutlined } from '@ant-design/icons';
import axios from 'axios';

export default function VideoToAnime() {
  const [file, setFile] = useState<any>(null);
  const [loading, setLoading] = useState(false);
  const [subtitle, setSubtitle] = useState('');
  const [script, setScript] = useState('');
  const [images, setImages] = useState<string[]>([]);
  const [audio, setAudio] = useState('');
  const [videoUrl, setVideoUrl] = useState('');

  const handleUpload = async () => {
    if (!file) return message.error('请先选择视频文件');
    setLoading(true);
    setSubtitle('');
    setScript('');
    setImages([]);
    setAudio('');
    setVideoUrl('');
    const formData = new FormData();
    formData.append('video', file);
    try {
      const res = await axios.post('/api/v1/ai/video-to-anime', formData, {
        headers: { 'Content-Type': 'multipart/form-data' },
        timeout: 10 * 60 * 1000,
      });
      setSubtitle(res.data.subtitle);
      setScript(res.data.script);
      setImages(res.data.images);
      setAudio(res.data.audio);
      setVideoUrl(res.data.video_url);
    } catch {
      message.error('生成失败');
    }
    setLoading(false);
  };

  return (
    <Card title="视频转动漫" style={{ maxWidth: 900, margin: '0 auto', marginTop: 32 }}>
      <Upload
        beforeUpload={file => {
          setFile(file);
          return false;
        }}
        showUploadList={false}
        accept="video/*"
      >
        <Button icon={<UploadOutlined />}>选择视频文件</Button>
      </Upload>
      <Button type="primary" onClick={handleUpload} loading={loading} style={{ marginLeft: 16 }}>
        一键生成动漫
      </Button>
      <div style={{ marginTop: 24 }}>
        {loading && <Spin />}
        {subtitle && (
          <Card title="识别字幕" style={{ marginBottom: 16 }}>
            <pre>{subtitle}</pre>
          </Card>
        )}
        {script && (
          <Card title="分镜脚本" style={{ marginBottom: 16 }}>
            <pre>{script}</pre>
          </Card>
        )}
        {images.length > 0 && (
          <Card title="动漫画面" style={{ marginBottom: 16 }}>
            <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8 }}>
              {images.map((img, idx) =>
                img ? (
                  <img
                    key={idx}
                    src={`data:image/png;base64,${img}`}
                    alt={`panel-${idx + 1}`}
                    style={{ width: 180, border: '1px solid #eee' }}
                  />
                ) : null
              )}
            </div>
          </Card>
        )}
        {audio && (
          <Card title="AI配音">
            <audio controls src={`data:audio/wav;base64,${audio}`} />
          </Card>
        )}
        {videoUrl && (
          <Card title="动漫视频下载">
            <video controls src={videoUrl} style={{ width: 400 }} />
            <div>
              <a href={videoUrl} download target="_blank" rel="noopener noreferrer">
                点击下载动漫视频
              </a>
            </div>
          </Card>
        )}
      </div>
    </Card>
  );
} 