import React, { useEffect, useState } from 'react';
import { Card, Row, Col, Progress, Typography, Button, Statistic, Alert, Space, Tag } from 'antd';
import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import {
  VideoCameraOutlined,
  PictureOutlined,
  CloudOutlined,
  ApiOutlined,
  RocketOutlined,
  CalendarOutlined,
  TrophyOutlined,
} from '@ant-design/icons';
import { useQuotas } from '../store';
import { aiApi } from '../services/api';
import dayjs from 'dayjs';

const { Title, Text, Paragraph } = Typography;

interface QuotaInfo {
  type: string;
  name: string;
  icon: React.ReactNode;
  color: string;
  unit: string;
  description: string;
}

const QuotaPage: React.FC = () => {
  const quotas = useQuotas();
  const [loading, setLoading] = useState(true);
  const [usageStats, setUsageStats] = useState<any>(null);

  // 配额类型配置
  const quotaTypes: QuotaInfo[] = [
    {
      type: 'video',
      name: '视频生成',
      icon: <VideoCameraOutlined />,
      color: '#0ea5e9',
      unit: '个',
      description: '每月可生成的视频数量',
    },
    {
      type: 'image',
      name: '图片生成',
      icon: <PictureOutlined />,
      color: '#a855f7',
      unit: '张',
      description: '每月可生成的图片数量',
    },
    {
      type: 'storage',
      name: '存储空间',
      icon: <CloudOutlined />,
      color: '#06b6d4',
      unit: 'GB',
      description: '可使用的存储空间大小',
    },
    {
      type: 'api',
      name: 'API调用',
      icon: <ApiOutlined />,
      color: '#10b981',
      unit: '次',
      description: '每月可调用API的次数',
    },
  ];

  useEffect(() => {
    const loadData = async () => {
      try {
        // 加载配额信息
        await aiApi.getUserQuota();
        
        // 加载使用统计
        const stats = await aiApi.getUsageStats();
        setUsageStats(stats);
      } catch (error) {
        console.error('加载配额数据失败:', error);
      } finally {
        setLoading(false);
      }
    };

    loadData();
  }, []);

  // 计算配额使用百分比
  const getUsagePercentage = (used: number, limit: number) => {
    if (limit === 0) return 0;
    return Math.round((used / limit) * 100);
  };

  // 获取配额状态颜色
  const getQuotaStatusColor = (percentage: number) => {
    if (percentage >= 90) return '#ef4444';
    if (percentage >= 70) return '#f59e0b';
    return '#10b981';
  };

  // 格式化存储大小
  const formatStorageSize = (bytes: number) => {
    const gb = bytes / (1024 * 1024 * 1024);
    return gb.toFixed(2);
  };

  return (
    <div style={{ padding: 24 }}>
      {/* 页面标题 */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        style={{ marginBottom: 32 }}
      >
        <Title level={2} style={{ marginBottom: 8 }}>
          配额管理
        </Title>
        <Paragraph style={{ fontSize: 16, color: '#64748b', margin: 0 }}>
          查看您的使用配额和统计信息，合理规划您的创作计划。
        </Paragraph>
      </motion.div>

      {/* 配额即将用完提醒 */}
      {Object.values(quotas).some((quota: any) => 
        quota && getUsagePercentage(quota.used, quota.limit) >= 80
      ) && (
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.5, delay: 0.1 }}
          style={{ marginBottom: 24 }}
        >
          <Alert
            message="配额使用提醒"
            description="您的部分配额使用量已超过80%，建议升级套餐以获得更多配额。"
            type="warning"
            showIcon
            action={
              <Link to="/pricing">
                <Button size="small" type="primary">
                  升级套餐
                </Button>
              </Link>
            }
            closable
          />
        </motion.div>
      )}

      {/* 配额卡片 */}
      <Row gutter={[24, 24]} style={{ marginBottom: 32 }}>
        {quotaTypes.map((quotaType, index) => {
          const quota = quotas[quotaType.type];
          const used = quota?.used || 0;
          const limit = quota?.limit || 0;
          const percentage = getUsagePercentage(used, limit);
          const statusColor = getQuotaStatusColor(percentage);

          // 特殊处理存储空间显示
          const displayUsed = quotaType.type === 'storage' ? formatStorageSize(used) : used;
          const displayLimit = quotaType.type === 'storage' ? formatStorageSize(limit) : limit;

          return (
            <Col xs={24} sm={12} lg={6} key={quotaType.type}>
              <motion.div
                initial={{ opacity: 0, y: 20 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.5, delay: 0.1 + index * 0.1 }}
              >
                <Card
                  style={{
                    height: '100%',
                    borderRadius: 12,
                    border: `1px solid ${quotaType.color}20`,
                  }}
                  bodyStyle={{ padding: 20 }}
                >
                  <div style={{ marginBottom: 16 }}>
                    <div style={{
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'space-between',
                      marginBottom: 8,
                    }}>
                      <div style={{
                        width: 40,
                        height: 40,
                        borderRadius: 8,
                        background: `${quotaType.color}10`,
                        display: 'flex',
                        alignItems: 'center',
                        justifyContent: 'center',
                        color: quotaType.color,
                        fontSize: 18,
                      }}>
                        {quotaType.icon}
                      </div>
                      <Tag color={percentage >= 80 ? 'red' : 'green'}>
                        {percentage}%
                      </Tag>
                    </div>
                    <Title level={4} style={{ margin: 0, fontSize: 16 }}>
                      {quotaType.name}
                    </Title>
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      {quotaType.description}
                    </Text>
                  </div>

                  <div style={{ marginBottom: 12 }}>
                    <div style={{
                      display: 'flex',
                      justifyContent: 'space-between',
                      alignItems: 'baseline',
                      marginBottom: 8,
                    }}>
                      <Text strong style={{ fontSize: 20, color: statusColor }}>
                        {displayUsed}
                      </Text>
                      <Text type="secondary" style={{ fontSize: 14 }}>
                        / {displayLimit} {quotaType.unit}
                      </Text>
                    </div>
                    <Progress
                      percent={percentage}
                      showInfo={false}
                      strokeColor={statusColor}
                      trailColor="#f1f5f9"
                      strokeWidth={6}
                    />
                  </div>

                  <Text type="secondary" style={{ fontSize: 12 }}>
                    剩余 {quotaType.type === 'storage' ? 
                      formatStorageSize(limit - used) : 
                      (limit - used)
                    } {quotaType.unit}
                  </Text>
                </Card>
              </motion.div>
            </Col>
          );
        })}
      </Row>

      <Row gutter={[24, 24]}>
        {/* 使用统计 */}
        <Col xs={24} lg={16}>
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.5 }}
          >
            <Card
              title={
                <Space>
                  <TrophyOutlined style={{ color: '#f59e0b' }} />
                  使用统计
                </Space>
              }
              style={{ height: '100%' }}
            >
              {usageStats ? (
                <Row gutter={[24, 24]}>
                  <Col xs={24} sm={8}>
                    <Statistic
                      title="本月总使用"
                      value={Object.values(quotas).reduce((total: number, quota: any) => 
                        total + (quota?.used || 0), 0
                      )}
                      suffix="次"
                      valueStyle={{ color: '#0ea5e9' }}
                    />
                  </Col>
                  <Col xs={24} sm={8}>
                    <Statistic
                      title="平均每日使用"
                      value={Math.round(
                        Object.values(quotas).reduce((total: number, quota: any) => 
                          total + (quota?.used || 0), 0
                        ) / dayjs().date()
                      )}
                      suffix="次"
                      valueStyle={{ color: '#10b981' }}
                    />
                  </Col>
                  <Col xs={24} sm={8}>
                    <Statistic
                      title="配额利用率"
                      value={Math.round(
                        Object.values(quotas).reduce((totalUsed: number, quota: any) => 
                          totalUsed + (quota?.used || 0), 0
                        ) / Object.values(quotas).reduce((totalLimit: number, quota: any) => 
                          totalLimit + (quota?.limit || 0), 0
                        ) * 100
                      )}
                      suffix="%"
                      valueStyle={{ color: '#a855f7' }}
                    />
                  </Col>
                </Row>
              ) : (
                <div style={{ textAlign: 'center', padding: '40px 0' }}>
                  <Text type="secondary">暂无使用统计数据</Text>
                </div>
              )}
            </Card>
          </motion.div>
        </Col>

        {/* 配额重置信息 */}
        <Col xs={24} lg={8}>
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.6 }}
          >
            <Card
              title={
                <Space>
                  <CalendarOutlined style={{ color: '#06b6d4' }} />
                  配额重置
                </Space>
              }
              style={{ height: '100%' }}
            >
              <div style={{ textAlign: 'center', padding: '20px 0' }}>
                <div style={{ marginBottom: 16 }}>
                  <Text type="secondary">下次重置时间</Text>
                </div>
                <Title level={3} style={{ margin: 0, color: '#06b6d4' }}>
                  {dayjs().add(1, 'month').startOf('month').format('YYYY年MM月DD日')}
                </Title>
                <div style={{ marginTop: 16 }}>
                  <Text type="secondary">
                    距离重置还有 {dayjs().add(1, 'month').startOf('month').diff(dayjs(), 'day')} 天
                  </Text>
                </div>
              </div>

              <div style={{ marginTop: 24, textAlign: 'center' }}>
                <Link to="/pricing">
                  <Button type="primary" icon={<RocketOutlined />} block>
                    升级套餐获得更多配额
                  </Button>
                </Link>
              </div>
            </Card>
          </motion.div>
        </Col>
      </Row>
    </div>
  );
};

export default QuotaPage;
