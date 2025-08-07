import React, { useState, useEffect } from 'react';
import { Card, Radio, Typography, Space, Tag, Statistic, Row, Col, Button, Alert, Tooltip, Progress } from 'antd';
import { ThunderboltOutlined, RocketOutlined, SettingOutlined, CheckCircleOutlined } from '@ant-design/icons';

const { Title, Text } = Typography;

interface EngineInfo {
  name: string;
  type: string;
  speed: string;
  memory_usage: number;
  quality: string;
  description: string;
  advantages: string[];
  status: 'ready' | 'loading' | 'error';
}

interface EngineSelectorProps {
  onEngineChange?: (engine: string) => void;
  currentEngine?: string;
}

const EngineSelector: React.FC<EngineSelectorProps> = ({ onEngineChange, currentEngine = 'builtin' }) => {
  const [selectedEngine, setSelectedEngine] = useState(currentEngine);
  const [engineStatus, setEngineStatus] = useState<any>(null);
  const [switching, setSwitching] = useState(false);

  const engines: Record<string, EngineInfo> = {
    builtin: {
      name: '内置轻量级引擎',
      type: 'builtin',
      speed: '2-5秒/张',
      memory_usage: 4.2,
      quality: '8.5/10',
      description: '基于SDXL-Turbo/Lightning的轻量级引擎，显存占用低，生成速度快',
      advantages: [
        '显存占用减少50%',
        '生成速度提升3-5倍',
        '支持多种轻量级模型',
        '智能批处理优化',
        '内置LoRA支持'
      ],
      status: 'ready'
    },
    sd_webui: {
      name: 'Stable Diffusion WebUI',
      type: 'sd_webui',
      speed: '8-15秒/张',
      memory_usage: 8.0,
      quality: '9.0/10',
      description: '传统SD WebUI，质量最高，功能最全面',
      advantages: [
        '最高图片质量',
        '完整功能支持',
        '丰富的插件生态',
        '成熟稳定',
        '社区支持好'
      ],
      status: 'ready'
    }
  };

  useEffect(() => {
    fetchEngineStatus();
  }, []);

  const fetchEngineStatus = async () => {
    try {
      const response = await fetch('/api/v1/system/engine');
      const result = await response.json();
      if (result.code === 200) {
        setEngineStatus(result.data);
      }
    } catch (error) {
      console.error('获取引擎状态失败:', error);
    }
  };

  const handleEngineChange = async (engine: string) => {
    if (engine === selectedEngine) return;

    setSwitching(true);
    try {
      const response = await fetch('/api/v1/system/engine/switch', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ engine }),
      });

      const result = await response.json();
      if (result.code === 200) {
        setSelectedEngine(engine);
        onEngineChange?.(engine);
        await fetchEngineStatus();
      }
    } catch (error) {
      console.error('切换引擎失败:', error);
    } finally {
      setSwitching(false);
    }
  };

  const getEngineIcon = (type: string) => {
    switch (type) {
      case 'builtin':
        return <ThunderboltOutlined style={{ color: '#52c41a' }} />;
      case 'sd_webui':
        return <RocketOutlined style={{ color: '#1890ff' }} />;
      default:
        return <SettingOutlined />;
    }
  };

  const getPerformanceColor = (value: number, type: 'memory' | 'speed') => {
    if (type === 'memory') {
      return value < 5 ? '#52c41a' : value < 8 ? '#faad14' : '#ff4d4f';
    }
    return '#1890ff';
  };

  return (
    <div>
      <Card title={
        <Space>
          <SettingOutlined />
          <span>AI引擎选择</span>
          <Tag color="blue">智能切换</Tag>
        </Space>
      }>
        <Alert
          message="💡 引擎选择建议"
          description="内置引擎适合快速生成和批量处理，SD WebUI适合高质量单张生成。您可以随时切换。"
          type="info"
          showIcon
          style={{ marginBottom: 24 }}
        />

        <Radio.Group 
          value={selectedEngine} 
          onChange={(e) => handleEngineChange(e.target.value)}
          style={{ width: '100%' }}
          disabled={switching}
        >
          <Row gutter={[16, 16]}>
            {Object.entries(engines).map(([key, engine]) => (
              <Col xs={24} lg={12} key={key}>
                <Card
                  hoverable
                  style={{
                    border: selectedEngine === key ? '2px solid #1890ff' : '1px solid #d9d9d9',
                    position: 'relative'
                  }}
                  bodyStyle={{ padding: 16 }}
                >
                  <Radio value={key} style={{ position: 'absolute', top: 16, right: 16 }} />
                  
                  <Space direction="vertical" style={{ width: '100%' }} size="middle">
                    {/* 引擎标题 */}
                    <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                      {getEngineIcon(engine.type)}
                      <div>
                        <Title level={5} style={{ margin: 0 }}>
                          {engine.name}
                        </Title>
                        <Text type="secondary" style={{ fontSize: 12 }}>
                          {engine.description}
                        </Text>
                      </div>
                    </div>

                    {/* 性能指标 */}
                    <Row gutter={16}>
                      <Col span={8}>
                        <Statistic
                          title="生成速度"
                          value={engine.speed}
                          valueStyle={{ 
                            fontSize: 14, 
                            color: getPerformanceColor(0, 'speed') 
                          }}
                        />
                      </Col>
                      <Col span={8}>
                        <Statistic
                          title="显存占用"
                          value={engine.memory_usage}
                          suffix="GB"
                          valueStyle={{ 
                            fontSize: 14, 
                            color: getPerformanceColor(engine.memory_usage, 'memory') 
                          }}
                        />
                      </Col>
                      <Col span={8}>
                        <Statistic
                          title="质量评分"
                          value={engine.quality}
                          valueStyle={{ fontSize: 14 }}
                        />
                      </Col>
                    </Row>

                    {/* 显存使用进度条 */}
                    <div>
                      <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 4 }}>
                        <Text style={{ fontSize: 12 }}>显存使用率</Text>
                        <Text style={{ fontSize: 12 }}>
                          {engine.memory_usage}GB / 16GB
                        </Text>
                      </div>
                      <Progress
                        percent={(engine.memory_usage / 16) * 100}
                        strokeColor={getPerformanceColor(engine.memory_usage, 'memory')}
                        size="small"
                        showInfo={false}
                      />
                    </div>

                    {/* 优势列表 */}
                    <div>
                      <Text strong style={{ fontSize: 12, color: '#666' }}>核心优势:</Text>
                      <div style={{ marginTop: 8 }}>
                        {engine.advantages.slice(0, 3).map((advantage, index) => (
                          <Tag key={index} color="blue" style={{ fontSize: 11, margin: '2px 4px 2px 0' }}>
                            {advantage}
                          </Tag>
                        ))}
                      </div>
                    </div>

                    {/* 状态指示 */}
                    <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                      <Space>
                        <CheckCircleOutlined style={{ color: '#52c41a' }} />
                        <Text style={{ fontSize: 12, color: '#52c41a' }}>就绪</Text>
                      </Space>
                      {selectedEngine === key && (
                        <Tag color="green">当前使用</Tag>
                      )}
                    </div>
                  </Space>
                </Card>
              </Col>
            ))}
          </Row>
        </Radio.Group>

        {/* 性能对比 */}
        {engineStatus && (
          <Card 
            title="性能对比" 
            size="small" 
            style={{ marginTop: 16 }}
          >
            <Row gutter={16}>
              <Col span={8}>
                <Statistic
                  title="速度提升"
                  value="3-5"
                  suffix="倍"
                  prefix="🚀"
                  valueStyle={{ color: '#52c41a' }}
                />
                <Text type="secondary" style={{ fontSize: 11 }}>
                  内置引擎 vs SD WebUI
                </Text>
              </Col>
              <Col span={8}>
                <Statistic
                  title="显存节省"
                  value="50"
                  suffix="%"
                  prefix="💾"
                  valueStyle={{ color: '#1890ff' }}
                />
                <Text type="secondary" style={{ fontSize: 11 }}>
                  相比传统方案
                </Text>
              </Col>
              <Col span={8}>
                <Statistic
                  title="并发能力"
                  value="3-6"
                  suffix="张"
                  prefix="⚡"
                  valueStyle={{ color: '#faad14' }}
                />
                <Text type="secondary" style={{ fontSize: 11 }}>
                  同时处理图片数
                </Text>
              </Col>
            </Row>
          </Card>
        )}

        {/* 切换状态 */}
        {switching && (
          <Alert
            message="正在切换引擎..."
            description="请稍候，系统正在切换到新的AI引擎。"
            type="info"
            showIcon
            style={{ marginTop: 16 }}
          />
        )}
      </Card>
    </div>
  );
};

export default EngineSelector;
