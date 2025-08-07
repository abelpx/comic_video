import React, { useState, useEffect } from 'react';
import { Card, Steps, Progress, Typography, Space, Tag, Spin, Image, Row, Col, Avatar, Button } from 'antd';
import { CheckCircleOutlined, LoadingOutlined, ExclamationCircleOutlined, ClockCircleOutlined, UserOutlined } from '@ant-design/icons';

const { Text, Title } = Typography;

interface TaskStep {
  name: string;
  status: string;
  progress: number;
  start_time?: string;
  end_time?: string;
  result?: string;
  error?: string;
  description?: string;
}

interface TaskProgressProps {
  taskId: string;
  status: string;
  progress: number;
  steps?: TaskStep[];
  result?: any;
  error?: string;
}

const TaskProgress: React.FC<TaskProgressProps> = ({
  taskId,
  status,
  progress,
  steps = [],
  result,
  error
}) => {
  // 步骤名称映射
  const stepNameMap: Record<string, string> = {
    script_generation: '分镜脚本生成',
    image_generation: '图片生成',
    voice_generation: '配音生成',
    video_composition: '视频合成'
  };

  // 获取步骤状态图标
  const getStepIcon = (stepStatus: string) => {
    switch (stepStatus) {
      case 'completed':
        return <CheckCircleOutlined style={{ color: '#52c41a' }} />;
      case 'processing':
        return <LoadingOutlined style={{ color: '#1890ff' }} />;
      case 'failed':
        return <ExclamationCircleOutlined style={{ color: '#ff4d4f' }} />;
      default:
        return <ClockCircleOutlined style={{ color: '#d9d9d9' }} />;
    }
  };

  // 获取步骤状态标签
  const getStepStatusTag = (stepStatus: string) => {
    const statusConfig = {
      pending: { color: 'default', text: '等待中' },
      processing: { color: 'processing', text: '进行中' },
      completed: { color: 'success', text: '已完成' },
      failed: { color: 'error', text: '失败' }
    };
    const config = statusConfig[stepStatus as keyof typeof statusConfig] || statusConfig.pending;
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  // 计算当前步骤
  const getCurrentStep = () => {
    const processingIndex = steps.findIndex(step => step.status === 'processing');
    if (processingIndex !== -1) return processingIndex;
    
    const completedCount = steps.filter(step => step.status === 'completed').length;
    return completedCount;
  };

  // 格式化时间
  const formatTime = (timeStr?: string) => {
    if (!timeStr) return '';
    return new Date(timeStr).toLocaleTimeString();
  };

  // 计算步骤耗时
  const getStepDuration = (step: TaskStep) => {
    if (!step.start_time || !step.end_time) return '';
    const start = new Date(step.start_time);
    const end = new Date(step.end_time);
    const duration = Math.round((end.getTime() - start.getTime()) / 1000);
    return `${duration}秒`;
  };

  // 解析步骤结果
  const parseStepResult = (step: TaskStep) => {
    if (!step.result) return null;
    try {
      return JSON.parse(step.result);
    } catch {
      return step.result;
    }
  };

  // 渲染分镜内容
  const renderScriptContent = (scriptStep: TaskStep) => {
    const scriptResult = parseStepResult(scriptStep);
    if (!scriptResult) return null;

    const panels = Array.isArray(scriptResult) ? scriptResult : scriptResult.panels;
    const characters = scriptResult.characters || [];

    return (
      <Space direction="vertical" style={{ width: '100%', marginTop: 8 }} size="middle">
        {/* 人物信息 */}
        {characters.length > 0 && (
          <Card size="small" title={`提取的人物 (${characters.length}个)`}>
            <Row gutter={[16, 8]}>
              {characters.map((character: any, index: number) => (
                <Col span={8} key={index}>
                  <Card size="small" style={{ textAlign: 'center' }}>
                    <Avatar
                      size={40}
                      src={character.avatar}
                      icon={<UserOutlined />}
                      style={{ marginBottom: 8 }}
                    />
                    <div>
                      <Text strong style={{ display: 'block' }}>{character.name}</Text>
                      <Text type="secondary" style={{ fontSize: 12 }}>
                        出现 {character.appearances} 次
                      </Text>
                      {character.description && (
                        <Paragraph
                          ellipsis={{ rows: 2 }}
                          style={{ fontSize: 11, margin: '4px 0 0 0' }}
                        >
                          {character.description}
                        </Paragraph>
                      )}
                    </div>
                  </Card>
                </Col>
              ))}
            </Row>
          </Card>
        )}

        {/* 分镜脚本 */}
        {panels && (
          <Card size="small" title={`分镜脚本 (${panels.length}个镜头)`}>
            <Row gutter={[16, 8]}>
              {panels.map((panel: string, index: number) => (
                <Col span={24} key={index}>
                  <Card size="small" style={{ backgroundColor: '#fafafa' }}>
                    <Space>
                      <Avatar size="small" style={{ backgroundColor: '#1890ff' }}>
                        {index + 1}
                      </Avatar>
                      <Text>{panel}</Text>
                    </Space>
                  </Card>
                </Col>
              ))}
            </Row>
          </Card>
        )}
      </Space>
    );
  };

  // 渲染图片内容 - 增强版
  const renderImageContent = () => {
    if (!result?.images) return null;

    return (
      <Card size="small" title={`生成图片 (${result.images.length}张)`} style={{ marginTop: 8 }}>
        <Row gutter={[16, 16]}>
          {result.images.map((img: string, index: number) => (
            <Col xs={24} sm={12} md={8} lg={6} key={index}>
              <Card
                size="small"
                hoverable
                cover={
                  <div style={{ position: 'relative' }}>
                    <Image
                      src={img.startsWith('data:') ? img : `data:image/png;base64,${img}`}
                      alt={`分镜 ${index + 1}`}
                      style={{ height: 160, objectFit: 'cover', width: '100%' }}
                      preview={{
                        mask: <div style={{ color: 'white' }}>🔍 预览</div>
                      }}
                    />
                    <div style={{
                      position: 'absolute',
                      top: 8,
                      left: 8,
                      background: 'rgba(0,0,0,0.7)',
                      color: 'white',
                      padding: '4px 8px',
                      borderRadius: 4,
                      fontSize: 12
                    }}>
                      #{index + 1}
                    </div>
                  </div>
                }
                style={{ borderRadius: 8 }}
              >
                <Card.Meta
                  title={`第${index + 1}幕`}
                  description={
                    result.panels?.[index] ? (
                      <Text ellipsis={{ rows: 2 }} style={{ fontSize: 12, color: '#666' }}>
                        {result.panels[index]}
                      </Text>
                    ) : (
                      <Text type="secondary" style={{ fontSize: 12 }}>
                        AI生成的精美图片
                      </Text>
                    )
                  }
                />
              </Card>
            </Col>
          ))}
        </Row>
      </Card>
    );
  };

  // 渲染音频内容
  const renderAudioContent = () => {
    if (!result?.audio_url) return null;

    return (
      <Card size="small" title="AI配音" style={{ marginTop: 8 }}>
        <Space direction="vertical" style={{ width: '100%' }}>
          <audio controls style={{ width: '100%' }}>
            <source src={result.audio_url} type="audio/mpeg" />
            您的浏览器不支持音频播放
          </audio>
          <Text type="secondary" style={{ fontSize: 12 }}>
            🎵 专业级AI语音合成，自然流畅的旁白效果
          </Text>
        </Space>
      </Card>
    );
  };

  // 渲染视频内容
  const renderVideoContent = () => {
    if (!result?.video_url) return null;

    return (
      <Card size="small" title="最终视频" style={{ marginTop: 8 }}>
        <Space direction="vertical" style={{ width: '100%' }}>
          <video
            controls
            style={{ width: '100%', maxHeight: 400, borderRadius: 8 }}
            poster={result.images?.[0] ? `data:image/png;base64,${result.images[0]}` : undefined}
          >
            <source src={result.video_url} type="video/mp4" />
            您的浏览器不支持视频播放
          </video>
          <Row gutter={16}>
            <Col span={12}>
              <Button type="primary" block icon={<DownloadOutlined />}>
                下载视频
              </Button>
            </Col>
            <Col span={12}>
              <Button block icon={<ShareAltOutlined />}>
                分享作品
              </Button>
            </Col>
          </Row>
        </Space>
      </Card>
    );
  };

  return (
    <Card title="任务进度" style={{ marginTop: 16 }}>
      <Space direction="vertical" style={{ width: '100%' }} size="large">
        {/* 总体进度 */}
        <div>
          <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
            <Text strong>总体进度</Text>
            <Text>{progress}%</Text>
          </div>
          <Progress 
            percent={progress} 
            status={status === 'failed' ? 'exception' : status === 'completed' ? 'success' : 'active'}
            strokeColor={{
              '0%': '#108ee9',
              '100%': '#87d068',
            }}
          />
        </div>

        {/* 步骤详情 */}
        {steps.length > 0 && (
          <div>
            <Title level={5}>执行步骤</Title>
            <Steps
              current={getCurrentStep()}
              status={status === 'failed' ? 'error' : 'process'}
              direction="vertical"
              size="small"
            >
              {steps.map((step, index) => (
                <Steps.Step
                  key={step.name}
                  title={stepNameMap[step.name] || step.name}
                  description={
                    <Space direction="vertical" size={8} style={{ width: '100%' }}>
                      <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                        {getStepIcon(step.status)}
                        {getStepStatusTag(step.status)}
                        {step.status === 'processing' && (
                          <Progress
                            percent={step.progress}
                            size="small"
                            style={{ width: 150 }}
                            showInfo={true}
                          />
                        )}
                      </div>
                      {step.description && (
                        <Text type="secondary" style={{ fontSize: 12 }}>
                          {step.description}
                        </Text>
                      )}
                      {step.start_time && (
                        <Text type="secondary" style={{ fontSize: 12 }}>
                          开始: {formatTime(step.start_time)}
                          {step.end_time && ` | 完成: ${formatTime(step.end_time)}`}
                          {getStepDuration(step) && ` | 耗时: ${getStepDuration(step)}`}
                        </Text>
                      )}
                      {step.error && (
                        <Text type="danger" style={{ fontSize: 12 }}>
                          错误: {step.error}
                        </Text>
                      )}

                      {/* 步骤结果展示 */}
                      {step.status === 'completed' && step.name === 'script_generation' && renderScriptContent(step)}
                      {step.status === 'completed' && step.name === 'image_generation' && renderImageContent()}
                      {step.status === 'completed' && step.name === 'voice_generation' && renderAudioContent()}
                      {step.status === 'completed' && step.name === 'video_composition' && renderVideoContent()}
                    </Space>
                  }
                  icon={step.status === 'processing' ? <Spin size="small" /> : undefined}
                />
              ))}
            </Steps>
          </div>
        )}

        {/* 任务结果 */}
        {status === 'completed' && result && (
          <div>
            <Title level={5}>任务结果</Title>
            <Card size="small" style={{ backgroundColor: '#f6ffed', border: '1px solid #b7eb8f' }}>
              {result.url && (
                <div style={{ marginBottom: 8 }}>
                  <Text strong>视频地址: </Text>
                  <a href={result.url} target="_blank" rel="noopener noreferrer">
                    查看视频
                  </a>
                </div>
              )}
              {result.images && (
                <div style={{ marginBottom: 8 }}>
                  <Text strong>生成图片: </Text>
                  <Text>{result.images.length} 张</Text>
                </div>
              )}
              {result.panels && (
                <div>
                  <Text strong>分镜数量: </Text>
                  <Text>{result.panels.length} 个</Text>
                </div>
              )}
            </Card>
          </div>
        )}

        {/* 错误信息 */}
        {status === 'failed' && error && (
          <div>
            <Title level={5}>错误信息</Title>
            <Card size="small" style={{ backgroundColor: '#fff2f0', border: '1px solid #ffccc7' }}>
              <Text type="danger">{error}</Text>
            </Card>
          </div>
        )}
      </Space>
    </Card>
  );
};

export default TaskProgress;
