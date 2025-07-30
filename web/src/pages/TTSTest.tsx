import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Input, 
  Button, 
  Select, 
  Space, 
  message, 
  Row, 
  Col, 
  Slider,
  Typography,
  Divider,
  Tag,
  Alert,
  Spin
} from 'antd';
import { 
  PlayCircleOutlined, 
  PauseCircleOutlined,
  DownloadOutlined,
  SoundOutlined,
  InfoCircleOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined
} from '@ant-design/icons';
import axios from 'axios';

const { TextArea } = Input;
const { Option } = Select;
const { Title, Text } = Typography;

interface TTSConfig {
  endpoint: string;
  supported_languages: string[];
  supported_emotions: string[];
  default_voice: string;
  default_language: string;
  default_speed: number;
  default_pitch: number;
  default_volume: number;
}

interface ServiceInfo {
  name: string;
  version: string;
  status: string;
  models: string[];
  endpoint: string;
  last_checked: string;
}

export default function TTSTest() {
  const [text, setText] = useState('你好，这是一个语音合成测试。欢迎使用VidCraft Studio的TTS服务！');
  const [voiceModel, setVoiceModel] = useState('spark_tts_zh');
  const [language, setLanguage] = useState('zh');
  const [speed, setSpeed] = useState(1.0);
  const [pitch, setPitch] = useState(1.0);
  const [volume, setVolume] = useState(1.0);
  const [emotion, setEmotion] = useState('neutral');
  
  const [loading, setLoading] = useState(false);
  const [playing, setPlaying] = useState(false);
  const [audioUrl, setAudioUrl] = useState<string | null>(null);
  const [audioElement, setAudioElement] = useState<HTMLAudioElement | null>(null);
  
  const [config, setConfig] = useState<TTSConfig | null>(null);
  const [serviceInfo, setServiceInfo] = useState<ServiceInfo | null>(null);
  const [healthStatus, setHealthStatus] = useState<'checking' | 'healthy' | 'unhealthy'>('checking');

  useEffect(() => {
    loadConfig();
    loadServiceInfo();
    checkHealth();
  }, []);

  // 加载TTS配置
  const loadConfig = async () => {
    try {
      const response = await axios.get('/api/v1/tts/config');
      if (response.data.code === 200) {
        const configData = response.data.data;
        setConfig(configData);
        
        // 设置默认值
        setVoiceModel(configData.default_voice);
        setLanguage(configData.default_language);
        setSpeed(configData.default_speed);
        setPitch(configData.default_pitch);
        setVolume(configData.default_volume);
      }
    } catch (error) {
      console.error('加载TTS配置失败:', error);
    }
  };

  // 加载服务信息
  const loadServiceInfo = async () => {
    try {
      const response = await axios.get('/api/v1/tts/info');
      if (response.data.code === 200) {
        setServiceInfo(response.data.data);
      }
    } catch (error) {
      console.error('加载服务信息失败:', error);
    }
  };

  // 检查服务健康状态
  const checkHealth = async () => {
    setHealthStatus('checking');
    try {
      const response = await axios.get('/api/v1/tts/health');
      if (response.data.code === 200) {
        setHealthStatus('healthy');
      } else {
        setHealthStatus('unhealthy');
      }
    } catch (error) {
      setHealthStatus('unhealthy');
      console.error('健康检查失败:', error);
    }
  };

  // 生成语音
  const generateVoice = async () => {
    if (!text.trim()) {
      message.warning('请输入要合成的文本');
      return;
    }

    if (text.length > 1000) {
      message.warning('文本长度不能超过1000个字符');
      return;
    }

    setLoading(true);
    try {
      const response = await axios.post('/api/v1/tts/generate', {
        text: text.trim(),
        voice_model: voiceModel,
        language: language,
        speed: speed,
        pitch: pitch,
        volume: volume,
        emotion: emotion
      }, {
        responseType: 'blob'
      });

      // 创建音频URL
      const audioBlob = new Blob([response.data], { type: 'audio/wav' });
      const url = URL.createObjectURL(audioBlob);
      setAudioUrl(url);

      // 创建音频元素
      const audio = new Audio(url);
      setAudioElement(audio);

      // 获取音频信息
      const duration = response.headers['x-audio-duration'];
      const textLength = response.headers['x-text-length'];
      
      message.success(`语音生成成功！时长: ${duration}秒, 文本长度: ${textLength}字符`);
    } catch (error: any) {
      console.error('语音生成失败:', error);
      if (error.response?.data) {
        const reader = new FileReader();
        reader.onload = () => {
          try {
            const errorData = JSON.parse(reader.result as string);
            message.error(`语音生成失败: ${errorData.message}`);
          } catch {
            message.error('语音生成失败');
          }
        };
        reader.readAsText(error.response.data);
      } else {
        message.error('语音生成失败');
      }
    } finally {
      setLoading(false);
    }
  };

  // 播放/暂停音频
  const togglePlayback = () => {
    if (!audioElement) return;

    if (playing) {
      audioElement.pause();
      setPlaying(false);
    } else {
      audioElement.play();
      setPlaying(true);
      
      audioElement.onended = () => {
        setPlaying(false);
      };
    }
  };

  // 下载音频
  const downloadAudio = () => {
    if (!audioUrl) return;

    const link = document.createElement('a');
    link.href = audioUrl;
    link.download = `tts_${Date.now()}.wav`;
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
  };

  // 测试语音
  const testVoice = async () => {
    setLoading(true);
    try {
      const response = await axios.get(`/api/v1/tts/test?voice_model=${voiceModel}&language=${language}`, {
        responseType: 'blob'
      });

      const audioBlob = new Blob([response.data], { type: 'audio/wav' });
      const url = URL.createObjectURL(audioBlob);
      
      const audio = new Audio(url);
      audio.play();

      const testText = response.headers['x-test-text'];
      message.success(`测试语音播放成功！测试文本: ${testText}`);
    } catch (error) {
      message.error('测试语音失败');
      console.error('测试语音失败:', error);
    } finally {
      setLoading(false);
    }
  };

  const getHealthStatusIcon = () => {
    switch (healthStatus) {
      case 'checking':
        return <Spin size="small" />;
      case 'healthy':
        return <CheckCircleOutlined style={{ color: '#52c41a' }} />;
      case 'unhealthy':
        return <ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />;
    }
  };

  const getHealthStatusText = () => {
    switch (healthStatus) {
      case 'checking':
        return '检查中...';
      case 'healthy':
        return '服务正常';
      case 'unhealthy':
        return '服务异常';
    }
  };

  return (
    <div style={{ padding: '24px', maxWidth: '1200px', margin: '0 auto' }}>
      <Title level={2}>TTS 语音合成测试</Title>
      
      {/* 服务状态 */}
      <Card style={{ marginBottom: 24 }}>
        <Row gutter={24}>
          <Col span={8}>
            <Space>
              {getHealthStatusIcon()}
              <Text strong>服务状态: </Text>
              <Text type={healthStatus === 'healthy' ? 'success' : 'danger'}>
                {getHealthStatusText()}
              </Text>
              <Button size="small" onClick={checkHealth}>刷新</Button>
            </Space>
          </Col>
          <Col span={8}>
            {serviceInfo && (
              <Space>
                <InfoCircleOutlined />
                <Text>服务: {serviceInfo.name} v{serviceInfo.version}</Text>
              </Space>
            )}
          </Col>
          <Col span={8}>
            {config && (
              <Space>
                <SoundOutlined />
                <Text>端点: {config.endpoint}</Text>
              </Space>
            )}
          </Col>
        </Row>
      </Card>

      {healthStatus === 'unhealthy' && (
        <Alert
          message="TTS服务不可用"
          description="请检查TTS服务是否正常运行，或联系管理员。"
          type="error"
          showIcon
          style={{ marginBottom: 24 }}
        />
      )}

      <Row gutter={24}>
        {/* 左侧：参数配置 */}
        <Col span={12}>
          <Card title="语音参数配置">
            <Space direction="vertical" style={{ width: '100%' }} size="large">
              {/* 文本输入 */}
              <div>
                <Text strong>合成文本</Text>
                <TextArea
                  rows={4}
                  value={text}
                  onChange={(e) => setText(e.target.value)}
                  placeholder="输入要合成的文本..."
                  maxLength={1000}
                  showCount
                />
              </div>

              {/* 语音模型 */}
              <Row gutter={16}>
                <Col span={12}>
                  <Text strong>语音模型</Text>
                  <Select
                    style={{ width: '100%' }}
                    value={voiceModel}
                    onChange={setVoiceModel}
                  >
                    {serviceInfo?.models.map(model => (
                      <Option key={model} value={model}>{model}</Option>
                    ))}
                  </Select>
                </Col>
                <Col span={12}>
                  <Text strong>语言</Text>
                  <Select
                    style={{ width: '100%' }}
                    value={language}
                    onChange={setLanguage}
                  >
                    {config?.supported_languages.map(lang => (
                      <Option key={lang} value={lang}>
                        {lang === 'zh' ? '中文' : lang === 'en' ? '英文' : lang === 'zh-en' ? '中英混合' : lang}
                      </Option>
                    ))}
                  </Select>
                </Col>
              </Row>

              {/* 情感 */}
              <div>
                <Text strong>情感</Text>
                <Select
                  style={{ width: '100%' }}
                  value={emotion}
                  onChange={setEmotion}
                >
                  {config?.supported_emotions.map(emo => (
                    <Option key={emo} value={emo}>
                      {emo === 'neutral' ? '中性' : 
                       emo === 'happy' ? '开心' :
                       emo === 'sad' ? '悲伤' :
                       emo === 'angry' ? '愤怒' :
                       emo === 'surprise' ? '惊讶' :
                       emo === 'fear' ? '恐惧' :
                       emo === 'disgust' ? '厌恶' : emo}
                    </Option>
                  ))}
                </Select>
              </div>

              {/* 语音参数滑块 */}
              <div>
                <Text strong>语速: {speed}</Text>
                <Slider
                  min={0.5}
                  max={2.0}
                  step={0.1}
                  value={speed}
                  onChange={setSpeed}
                  marks={{
                    0.5: '0.5x',
                    1.0: '1.0x',
                    1.5: '1.5x',
                    2.0: '2.0x'
                  }}
                />
              </div>

              <div>
                <Text strong>音调: {pitch}</Text>
                <Slider
                  min={0.5}
                  max={2.0}
                  step={0.1}
                  value={pitch}
                  onChange={setPitch}
                  marks={{
                    0.5: '低',
                    1.0: '正常',
                    2.0: '高'
                  }}
                />
              </div>

              <div>
                <Text strong>音量: {volume}</Text>
                <Slider
                  min={0.1}
                  max={2.0}
                  step={0.1}
                  value={volume}
                  onChange={setVolume}
                  marks={{
                    0.1: '小',
                    1.0: '正常',
                    2.0: '大'
                  }}
                />
              </div>
            </Space>
          </Card>
        </Col>

        {/* 右侧：操作和结果 */}
        <Col span={12}>
          <Card title="语音生成">
            <Space direction="vertical" style={{ width: '100%' }} size="large">
              {/* 操作按钮 */}
              <Row gutter={16}>
                <Col span={12}>
                  <Button
                    type="primary"
                    icon={<SoundOutlined />}
                    loading={loading}
                    onClick={generateVoice}
                    disabled={healthStatus !== 'healthy'}
                    block
                  >
                    生成语音
                  </Button>
                </Col>
                <Col span={12}>
                  <Button
                    icon={<PlayCircleOutlined />}
                    loading={loading}
                    onClick={testVoice}
                    disabled={healthStatus !== 'healthy'}
                    block
                  >
                    测试语音
                  </Button>
                </Col>
              </Row>

              {/* 音频控制 */}
              {audioUrl && (
                <Card size="small" title="生成的语音">
                  <Space>
                    <Button
                      type={playing ? 'default' : 'primary'}
                      icon={playing ? <PauseCircleOutlined /> : <PlayCircleOutlined />}
                      onClick={togglePlayback}
                    >
                      {playing ? '暂停' : '播放'}
                    </Button>
                    <Button
                      icon={<DownloadOutlined />}
                      onClick={downloadAudio}
                    >
                      下载
                    </Button>
                  </Space>
                </Card>
              )}

              <Divider />

              {/* 服务信息 */}
              {serviceInfo && (
                <Card size="small" title="服务信息">
                  <Space direction="vertical" style={{ width: '100%' }}>
                    <div>
                      <Text strong>可用模型: </Text>
                      <div style={{ marginTop: 8 }}>
                        {serviceInfo.models.map(model => (
                          <Tag key={model} color="blue" style={{ marginBottom: 4 }}>
                            {model}
                          </Tag>
                        ))}
                      </div>
                    </div>
                    <div>
                      <Text strong>支持语言: </Text>
                      <div style={{ marginTop: 8 }}>
                        {config?.supported_languages.map(lang => (
                          <Tag key={lang} color="green" style={{ marginBottom: 4 }}>
                            {lang}
                          </Tag>
                        ))}
                      </div>
                    </div>
                    <div>
                      <Text strong>支持情感: </Text>
                      <div style={{ marginTop: 8 }}>
                        {config?.supported_emotions.map(emo => (
                          <Tag key={emo} color="orange" style={{ marginBottom: 4 }}>
                            {emo}
                          </Tag>
                        ))}
                      </div>
                    </div>
                  </Space>
                </Card>
              )}
            </Space>
          </Card>
        </Col>
      </Row>
    </div>
  );
}
