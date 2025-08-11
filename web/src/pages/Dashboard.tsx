import React, { useEffect, useState } from 'react';
import { Row, Col, Card, Statistic, Button, Typography, Progress, List, Avatar, Tag, Space, Empty } from 'antd';
import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';
import {
  VideoCameraOutlined,
  PictureOutlined,
  BookOutlined,
  PlayCircleOutlined,
  PlusOutlined,
  TrophyOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  ExclamationCircleOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { useUser, useTasks, useQuotas } from '../store';
import { aiApi } from '../services/api';
import useTaskManager from '../hooks/useTaskManager';
import dayjs from 'dayjs';

const { Title, Text, Paragraph } = Typography;

const Dashboard: React.FC = () => {
  const user = useUser();
  const { tasks, activeTasks } = useTasks();
  const quotas = useQuotas();
  const [loading, setLoading] = useState(true);
  const [recentTasks, setRecentTasks] = useState<any[]>([]);

  // 使用任务管理器
  const { refreshTasks } = useTaskManager({
    autoLoad: false, // 在Dashboard中手动控制加载
    enablePolling: true,
  });

  // 快捷工具配置
  const quickTools = [
    {
      title: '小说转视频',
      description: '将文字小说转换为动漫视频',
      icon: <VideoCameraOutlined style={{ fontSize: 24, color: '#0ea5e9' }} />,
      link: '/app/novel-to-video',
      color: '#0ea5e9',
    },
    {
      title: '漫画生成',
      description: '基于文本生成漫画分镜',
      icon: <PictureOutlined style={{ fontSize: 24, color: '#a855f7' }} />,
      link: '/app/generate-comic',
      color: '#a855f7',
    },
    {
      title: '小说创作',
      description: 'AI辅助小说创作',
      icon: <BookOutlined style={{ fontSize: 24, color: '#06b6d4' }} />,
      link: '/app/generate-novel',
      color: '#06b6d4',
    },
    {
      title: '视频转动漫',
      description: '将真实视频转为动漫风格',
      icon: <PlayCircleOutlined style={{ fontSize: 24, color: '#10b981' }} />,
      link: '/app/video-to-anime',
      color: '#10b981',
    },
  ];

  useEffect(() => {
    const loadData = async () => {
      try {
        // 加载用户配额
        await aiApi.getUserQuota();

        // 使用任务管理器刷新任务
        await refreshTasks();

        // 从store中获取最新任务作为最近任务
        setRecentTasks(tasks.slice(0, 10));
      } catch (error) {
        console.error('加载数据失败:', error);
      } finally {
        setLoading(false);
      }
    };

    loadData();
  }, [refreshTasks]);

  // 当tasks变化时更新recentTasks
  useEffect(() => {
    setRecentTasks(tasks.slice(0, 10));
  }, [tasks]);

  // 获取任务状态图标
  const getTaskStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <CheckCircleOutlined style={{ color: '#10b981' }} />;
      case 'processing':
        return <ClockCircleOutlined style={{ color: '#f59e0b' }} />;
      case 'failed':
        return <ExclamationCircleOutlined style={{ color: '#ef4444' }} />;
      default:
        return <ClockCircleOutlined style={{ color: '#94a3b8' }} />;
    }
  };

  // 获取任务状态标签
  const getTaskStatusTag = (status: string) => {
    const statusMap = {
      pending: { color: 'default', text: '等待中' },
      processing: { color: 'processing', text: '处理中' },
      completed: { color: 'success', text: '已完成' },
      failed: { color: 'error', text: '失败' },
    };
    const config = statusMap[status as keyof typeof statusMap] || statusMap.pending;
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  return (
    <div style={{ padding: 24 }}>
      {/* 欢迎区域 */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
      >
        <div style={{ marginBottom: 32 }}>
          <Title level={2} style={{ marginBottom: 8 }}>
            欢迎回来，{user?.nickname || user?.username}！
          </Title>
          <Paragraph style={{ fontSize: 16, color: '#64748b', margin: 0 }}>
            今天想创作什么内容呢？选择下面的工具开始您的创作之旅。
          </Paragraph>
        </div>
      </motion.div>

      {/* 统计卡片 */}
      <Row gutter={[24, 24]} style={{ marginBottom: 32 }}>
        <Col xs={24} sm={12} lg={6}>
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.1 }}
          >
            <Card>
              <Statistic
                title="本月视频"
                value={quotas.video?.used || 0}
                suffix={`/ ${quotas.video?.limit || 10}`}
                prefix={<VideoCameraOutlined style={{ color: '#0ea5e9' }} />}
                valueStyle={{ color: '#0ea5e9' }}
              />
              <Progress
                percent={((quotas.video?.used || 0) / (quotas.video?.limit || 10)) * 100}
                showInfo={false}
                strokeColor="#0ea5e9"
                style={{ marginTop: 8 }}
              />
            </Card>
          </motion.div>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.2 }}
          >
            <Card>
              <Statistic
                title="活跃任务"
                value={activeTasks.length}
                prefix={<ClockCircleOutlined style={{ color: '#f59e0b' }} />}
                valueStyle={{ color: '#f59e0b' }}
              />
            </Card>
          </motion.div>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.3 }}
          >
            <Card>
              <Statistic
                title="总作品数"
                value={recentTasks.filter(task => task.status === 'completed').length}
                prefix={<TrophyOutlined style={{ color: '#10b981' }} />}
                valueStyle={{ color: '#10b981' }}
              />
            </Card>
          </motion.div>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.4 }}
          >
            <Card>
              <Statistic
                title="成功率"
                value={recentTasks.length > 0 ? 
                  Math.round((recentTasks.filter(task => task.status === 'completed').length / recentTasks.length) * 100) : 
                  0
                }
                suffix="%"
                prefix={<CheckCircleOutlined style={{ color: '#a855f7' }} />}
                valueStyle={{ color: '#a855f7' }}
              />
            </Card>
          </motion.div>
        </Col>
      </Row>

      <Row gutter={[24, 24]}>
        {/* 快捷工具 */}
        <Col xs={24} lg={16}>
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.5 }}
          >
            <Card
              title="快捷工具"
              extra={<Link to="/app/works">查看所有作品</Link>}
              style={{ height: '100%' }}
            >
              <Row gutter={[16, 16]}>
                {quickTools.map((tool, index) => (
                  <Col xs={24} sm={12} key={index}>
                    <Link to={tool.link} style={{ textDecoration: 'none' }}>
                      <Card
                        hoverable
                        size="small"
                        style={{
                          borderRadius: 8,
                          border: `1px solid ${tool.color}20`,
                          transition: 'all 0.3s ease',
                        }}
                        bodyStyle={{ padding: 16 }}
                      >
                        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                          <div style={{
                            width: 48,
                            height: 48,
                            borderRadius: 8,
                            background: `${tool.color}10`,
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                          }}>
                            {tool.icon}
                          </div>
                          <div style={{ flex: 1 }}>
                            <div style={{ 
                              fontWeight: 600, 
                              marginBottom: 4,
                              color: '#1e293b',
                            }}>
                              {tool.title}
                            </div>
                            <div style={{ 
                              fontSize: 12, 
                              color: '#64748b',
                              lineHeight: 1.4,
                            }}>
                              {tool.description}
                            </div>
                          </div>
                          <PlusOutlined style={{ color: tool.color }} />
                        </div>
                      </Card>
                    </Link>
                  </Col>
                ))}
              </Row>
            </Card>
          </motion.div>
        </Col>

        {/* 最近任务 */}
        <Col xs={24} lg={8}>
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.6 }}
          >
            <Card
              title="最近任务"
              extra={
                <Space>
                  <Button
                    type="text"
                    size="small"
                    icon={<ReloadOutlined />}
                    onClick={refreshTasks}
                    title="刷新任务"
                  />
                  <Link to="/app/works">查看全部</Link>
                </Space>
              }
              style={{ height: '100%' }}
            >
              {recentTasks.length > 0 ? (
                <List
                  dataSource={recentTasks.slice(0, 5)}
                  renderItem={(task) => (
                    <List.Item style={{ padding: '12px 0', border: 'none' }}>
                      <List.Item.Meta
                        avatar={
                          <Avatar 
                            icon={getTaskStatusIcon(task.status)}
                            style={{ backgroundColor: 'transparent' }}
                          />
                        }
                        title={
                          <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                            <span style={{ fontSize: 14, fontWeight: 500 }}>
                              {task.type === 'video' ? '视频生成' : '其他任务'}
                            </span>
                            {getTaskStatusTag(task.status)}
                          </div>
                        }
                        description={
                          <Space direction="vertical" size={4}>
                            <Text type="secondary" style={{ fontSize: 12 }}>
                              {dayjs(task.created_at).format('MM-DD HH:mm')}
                            </Text>
                            {task.status === 'processing' && (
                              <Progress
                                percent={task.progress || 0}
                                size="small"
                                showInfo={false}
                              />
                            )}
                          </Space>
                        }
                      />
                    </List.Item>
                  )}
                />
              ) : (
                <Empty
                  image={Empty.PRESENTED_IMAGE_SIMPLE}
                  description="暂无任务记录"
                  style={{ margin: '40px 0' }}
                >
                  <Button type="primary" size="small">
                    <Link to="/app/novel-to-video">开始创作</Link>
                  </Button>
                </Empty>
              )}
            </Card>
          </motion.div>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;
