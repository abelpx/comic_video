import React, { useState, useEffect } from 'react';
import { Card, Row, Col, Progress, Typography, Space, Tag, Statistic, Alert, Badge, Tooltip } from 'antd';
import { 
  ThunderboltOutlined, DatabaseOutlined, FireOutlined, 
  EyeOutlined, WarningOutlined, CheckCircleOutlined 
} from '@ant-design/icons';

const { Text, Title } = Typography;

interface GPUMetrics {
  usage: number;
  memory: number;
  memory_total: number;
  memory_used: number;
  temperature: number;
  power: number;
  name: string;
}

interface MemoryMetrics {
  usage: number;
  total: number;
  used: number;
  free: number;
}

interface AIMetrics {
  processing_speed: number;
  queue_length: number;
  success_rate: number;
  avg_time: number;
}

interface SystemMetrics {
  gpu: GPUMetrics;
  memory: MemoryMetrics;
  ai: AIMetrics;
}

const RealTimeMonitor: React.FC = () => {
  const [metrics, setMetrics] = useState<SystemMetrics | null>(null);
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // 建立EventSource连接进行实时监控
    const eventSource = new EventSource('/api/v1/system/stream');
    
    eventSource.onopen = () => {
      setConnected(true);
      setError(null);
    };

    eventSource.onmessage = (event) => {
      if (event.type === 'metrics') {
        try {
          const data = JSON.parse(event.data);
          setMetrics(data);
        } catch (err) {
          console.error('解析监控数据失败:', err);
        }
      }
    };

    eventSource.onerror = () => {
      setConnected(false);
      setError('监控连接断开，尝试重连中...');
      
      // 降级到轮询模式
      setTimeout(() => {
        fetchMetrics();
      }, 5000);
    };

    // 清理函数
    return () => {
      eventSource.close();
    };
  }, []);

  // 降级轮询获取数据
  const fetchMetrics = async () => {
    try {
      const response = await fetch('/api/v1/system/metrics');
      const result = await response.json();
      if (result.code === 200) {
        setMetrics(result.data);
        setConnected(true);
        setError(null);
      }
    } catch (err) {
      setError('获取监控数据失败');
    }
  };

  const getStatusColor = (value: number, thresholds: [number, number]) => {
    if (value < thresholds[0]) return 'success';
    if (value < thresholds[1]) return 'warning';
    return 'exception';
  };

  const getStatusIcon = (value: number, thresholds: [number, number]) => {
    if (value < thresholds[0]) return <CheckCircleOutlined style={{ color: '#52c41a' }} />;
    if (value < thresholds[1]) return <WarningOutlined style={{ color: '#faad14' }} />;
    return <WarningOutlined style={{ color: '#ff4d4f' }} />;
  };

  if (!metrics) {
    return (
      <Card title="系统监控" loading>
        <div style={{ textAlign: 'center', padding: 40 }}>
          <Text type="secondary">正在加载监控数据...</Text>
        </div>
      </Card>
    );
  }

  return (
    <div>
      {/* 连接状态 */}
      <Card 
        size="small" 
        style={{ marginBottom: 16 }}
        title={
          <Space>
            <Badge status={connected ? 'processing' : 'error'} />
            <span>实时监控</span>
            {connected ? (
              <Tag color="green">已连接</Tag>
            ) : (
              <Tag color="red">连接断开</Tag>
            )}
          </Space>
        }
      >
        {error && (
          <Alert message={error} type="warning" showIcon style={{ marginBottom: 16 }} />
        )}
      </Card>

      {/* GPU监控 */}
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={12}>
          <Card 
            title={
              <Space>
                <ThunderboltOutlined style={{ color: '#1890ff' }} />
                <span>GPU状态</span>
                {getStatusIcon(Math.max(metrics.gpu.usage, metrics.gpu.memory), [70, 85])}
              </Space>
            }
            extra={<Text type="secondary">{metrics.gpu.name}</Text>}
          >
            <Row gutter={[16, 16]}>
              <Col span={12}>
                <div style={{ marginBottom: 16 }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                    <Text strong>GPU使用率</Text>
                    <Text>{metrics.gpu.usage.toFixed(1)}%</Text>
                  </div>
                  <Progress 
                    percent={metrics.gpu.usage} 
                    status={getStatusColor(metrics.gpu.usage, [70, 85])}
                    strokeWidth={8}
                  />
                </div>
                
                <div style={{ marginBottom: 16 }}>
                  <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                    <Text strong>显存使用</Text>
                    <Text>{metrics.gpu.memory.toFixed(1)}%</Text>
                  </div>
                  <Progress 
                    percent={metrics.gpu.memory} 
                    status={getStatusColor(metrics.gpu.memory, [75, 90])}
                    strokeWidth={8}
                  />
                  <Text type="secondary" style={{ fontSize: 12 }}>
                    {(metrics.gpu.memory_used / 1024).toFixed(1)}GB / {(metrics.gpu.memory_total / 1024).toFixed(1)}GB
                  </Text>
                </div>
              </Col>
              
              <Col span={12}>
                <Row gutter={[8, 8]}>
                  <Col span={12}>
                    <Statistic 
                      title="温度" 
                      value={metrics.gpu.temperature.toFixed(0)} 
                      suffix="°C"
                      valueStyle={{ 
                        fontSize: 18,
                        color: metrics.gpu.temperature > 80 ? '#ff4d4f' : '#52c41a'
                      }}
                    />
                  </Col>
                  <Col span={12}>
                    <Statistic 
                      title="功耗" 
                      value={metrics.gpu.power.toFixed(0)} 
                      suffix="W"
                      valueStyle={{ fontSize: 18 }}
                    />
                  </Col>
                </Row>
              </Col>
            </Row>
          </Card>
        </Col>

        {/* 系统内存 */}
        <Col xs={24} lg={12}>
          <Card 
            title={
              <Space>
                <DatabaseOutlined style={{ color: '#52c41a' }} />
                <span>系统内存</span>
                {getStatusIcon(metrics.memory.usage, [70, 85])}
              </Space>
            }
            extra={<Text type="secondary">96GB 高性能内存</Text>}
          >
            <div style={{ marginBottom: 16 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', marginBottom: 8 }}>
                <Text strong>内存使用率</Text>
                <Text>{metrics.memory.usage.toFixed(1)}%</Text>
              </div>
              <Progress 
                percent={metrics.memory.usage} 
                status={getStatusColor(metrics.memory.usage, [70, 85])}
                strokeWidth={8}
              />
              <Text type="secondary" style={{ fontSize: 12 }}>
                已用: {(metrics.memory.used / 1024).toFixed(1)}GB | 
                可用: {(metrics.memory.free / 1024).toFixed(1)}GB | 
                总计: {(metrics.memory.total / 1024).toFixed(1)}GB
              </Text>
            </div>

            <Row gutter={[16, 16]}>
              <Col span={8}>
                <Statistic 
                  title="已用内存" 
                  value={(metrics.memory.used / 1024).toFixed(1)} 
                  suffix="GB"
                  valueStyle={{ fontSize: 16, color: '#1890ff' }}
                />
              </Col>
              <Col span={8}>
                <Statistic 
                  title="可用内存" 
                  value={(metrics.memory.free / 1024).toFixed(1)} 
                  suffix="GB"
                  valueStyle={{ fontSize: 16, color: '#52c41a' }}
                />
              </Col>
              <Col span={8}>
                <Statistic 
                  title="使用率" 
                  value={metrics.memory.usage.toFixed(1)} 
                  suffix="%"
                  valueStyle={{ fontSize: 16 }}
                />
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>

      {/* AI处理性能 */}
      <Row gutter={[16, 16]} style={{ marginTop: 16 }}>
        <Col span={24}>
          <Card 
            title={
              <Space>
                <FireOutlined style={{ color: '#fa8c16' }} />
                <span>AI处理性能</span>
                <Tag color="blue">实时统计</Tag>
              </Space>
            }
          >
            <Row gutter={[24, 16]}>
              <Col xs={12} sm={6}>
                <Statistic 
                  title="处理速度" 
                  value={metrics.ai.processing_speed.toFixed(1)} 
                  suffix="张/分钟"
                  prefix={<ThunderboltOutlined />}
                  valueStyle={{ color: '#52c41a', fontSize: 20 }}
                />
              </Col>
              <Col xs={12} sm={6}>
                <Statistic 
                  title="队列长度" 
                  value={metrics.ai.queue_length} 
                  suffix="个任务"
                  prefix={<DatabaseOutlined />}
                  valueStyle={{ color: '#1890ff', fontSize: 20 }}
                />
              </Col>
              <Col xs={12} sm={6}>
                <Statistic 
                  title="成功率" 
                  value={metrics.ai.success_rate.toFixed(1)} 
                  suffix="%"
                  prefix={<CheckCircleOutlined />}
                  valueStyle={{ color: '#fa8c16', fontSize: 20 }}
                />
              </Col>
              <Col xs={12} sm={6}>
                <Statistic 
                  title="平均耗时" 
                  value={metrics.ai.avg_time.toFixed(0)} 
                  suffix="秒"
                  prefix={<EyeOutlined />}
                  valueStyle={{ fontSize: 20 }}
                />
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>

      {/* 性能建议 */}
      {(metrics.gpu.usage > 85 || metrics.gpu.memory > 85 || metrics.memory.usage > 80) && (
        <Alert
          style={{ marginTop: 16 }}
          message="性能优化建议"
          description={
            <ul style={{ margin: 0, paddingLeft: 20 }}>
              {metrics.gpu.usage > 85 && <li>GPU使用率过高，建议降低图片分辨率或减少并发任务</li>}
              {metrics.gpu.memory > 85 && <li>GPU显存不足，建议减少批次大小或使用更小的模型</li>}
              {metrics.memory.usage > 80 && <li>系统内存使用率较高，建议关闭不必要的程序</li>}
            </ul>
          }
          type="warning"
          showIcon
        />
      )}
    </div>
  );
};

export default RealTimeMonitor;
