import React from 'react';
import { Card, Row, Col, Button, Typography, List, Tag, Space, Layout } from 'antd';
import { Link, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Helmet } from 'react-helmet-async';
import {
  CheckOutlined,
  StarOutlined,
  CrownOutlined,
  RocketOutlined,
  ArrowLeftOutlined,
} from '@ant-design/icons';
import { useAuth } from '../store';

const { Title, Text, Paragraph } = Typography;
const { Header, Content } = Layout;

interface PricingPlan {
  id: string;
  name: string;
  price: number;
  period: string;
  description: string;
  features: string[];
  popular?: boolean;
  icon: React.ReactNode;
  color: string;
  buttonText: string;
  quotas: {
    video: number;
    image: number;
    storage: number;
    api: number;
  };
}

const PricingPage: React.FC = () => {
  const navigate = useNavigate();
  const { isAuthenticated } = useAuth();

  const plans: PricingPlan[] = [
    {
      id: 'free',
      name: '免费版',
      price: 0,
      period: '永久免费',
      description: '适合个人用户体验和轻度使用',
      icon: <StarOutlined />,
      color: '#64748b',
      buttonText: '免费注册',
      quotas: {
        video: 10,
        image: 100,
        storage: 1,
        api: 1000,
      },
      features: [
        '每月10个视频生成',
        '每月100张图片生成',
        '1GB存储空间',
        '基础AI模型',
        '标准画质输出',
        '社区支持',
      ],
    },
    {
      id: 'pro',
      name: '专业版',
      price: 29,
      period: '每月',
      description: '适合内容创作者和小团队',
      icon: <RocketOutlined />,
      color: '#0ea5e9',
      buttonText: '立即订阅',
      popular: true,
      quotas: {
        video: 100,
        image: 1000,
        storage: 10,
        api: 10000,
      },
      features: [
        '每月100个视频生成',
        '每月1000张图片生成',
        '10GB存储空间',
        '高级AI模型',
        '高清画质输出',
        '角色一致性保证',
        '场景连贯性优化',
        '优先处理队列',
        '邮件技术支持',
      ],
    },
    {
      id: 'enterprise',
      name: '企业版',
      price: 99,
      period: '每月',
      description: '适合企业和大型团队',
      icon: <CrownOutlined />,
      color: '#a855f7',
      buttonText: '联系销售',
      quotas: {
        video: 500,
        image: 5000,
        storage: 100,
        api: 100000,
      },
      features: [
        '每月500个视频生成',
        '每月5000张图片生成',
        '100GB存储空间',
        '顶级AI模型',
        '4K超高清输出',
        '自定义角色模板',
        '批量处理功能',
        'API接口访问',
        '专属客户经理',
        '24/7技术支持',
        '定制化服务',
      ],
    },
  ];

  const handleSubscribe = (planId: string) => {
    if (!isAuthenticated) {
      navigate('/auth/register');
      return;
    }

    if (planId === 'free') {
      navigate('/app/dashboard');
    } else {
      // TODO: 集成支付流程
      console.log('订阅计划:', planId);
    }
  };

  return (
    <>
      <Helmet>
        <title>定价方案 - Comic AI</title>
        <meta name="description" content="选择适合您的Comic AI订阅计划，获得更多创作配额和高级功能" />
      </Helmet>

      <Layout style={{ minHeight: '100vh', background: '#f8fafc' }}>
        {/* 简化的导航栏 */}
        <Header style={{
          background: 'rgba(255, 255, 255, 0.95)',
          backdropFilter: 'blur(10px)',
          borderBottom: '1px solid #f0f0f0',
          padding: '0 24px',
        }}>
          <div style={{
            maxWidth: 1200,
            margin: '0 auto',
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'space-between',
            height: 64,
          }}>
            <Button
              type="text"
              icon={<ArrowLeftOutlined />}
              onClick={() => navigate(-1)}
              style={{ fontSize: 16 }}
            >
              返回
            </Button>
            
            <Link to="/" style={{ textDecoration: 'none' }}>
              <span style={{
                fontSize: 20,
                fontWeight: 600,
                background: 'linear-gradient(135deg, #0ea5e9 0%, #0284c7 100%)',
                WebkitBackgroundClip: 'text',
                WebkitTextFillColor: 'transparent',
              }}>
                Comic AI
              </span>
            </Link>

            <div style={{ width: 80 }} />
          </div>
        </Header>

        <Content style={{ padding: '80px 24px' }}>
          <div style={{ maxWidth: 1200, margin: '0 auto' }}>
            {/* 页面标题 */}
            <motion.div
              initial={{ opacity: 0, y: 30 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.8 }}
              style={{ textAlign: 'center', marginBottom: 64 }}
            >
              <Title level={1} style={{ fontSize: 48, marginBottom: 16 }}>
                选择适合您的方案
              </Title>
              <Paragraph style={{ fontSize: 18, color: '#64748b', maxWidth: 600, margin: '0 auto' }}>
                从免费体验到企业级解决方案，我们为每个创作者提供合适的选择
              </Paragraph>
            </motion.div>

            {/* 定价卡片 */}
            <Row gutter={[32, 32]} justify="center">
              {plans.map((plan, index) => (
                <Col xs={24} lg={8} key={plan.id}>
                  <motion.div
                    initial={{ opacity: 0, y: 20 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.5, delay: index * 0.1 }}
                  >
                    <Card
                      style={{
                        height: '100%',
                        borderRadius: 16,
                        border: plan.popular ? `2px solid ${plan.color}` : '1px solid #e2e8f0',
                        boxShadow: plan.popular ? 
                          `0 20px 40px ${plan.color}20` : 
                          '0 4px 6px rgba(0, 0, 0, 0.05)',
                        position: 'relative',
                        overflow: 'hidden',
                      }}
                      bodyStyle={{ padding: 32 }}
                    >
                      {/* 热门标签 */}
                      {plan.popular && (
                        <div style={{
                          position: 'absolute',
                          top: 0,
                          right: 0,
                          background: plan.color,
                          color: 'white',
                          padding: '8px 24px',
                          fontSize: 12,
                          fontWeight: 600,
                          borderBottomLeftRadius: 8,
                        }}>
                          最受欢迎
                        </div>
                      )}

                      {/* 计划图标和名称 */}
                      <div style={{ textAlign: 'center', marginBottom: 24 }}>
                        <div style={{
                          width: 64,
                          height: 64,
                          borderRadius: 16,
                          background: `${plan.color}10`,
                          display: 'flex',
                          alignItems: 'center',
                          justifyContent: 'center',
                          margin: '0 auto 16px',
                          color: plan.color,
                          fontSize: 24,
                        }}>
                          {plan.icon}
                        </div>
                        <Title level={3} style={{ margin: 0, color: plan.color }}>
                          {plan.name}
                        </Title>
                        <Text type="secondary" style={{ fontSize: 14 }}>
                          {plan.description}
                        </Text>
                      </div>

                      {/* 价格 */}
                      <div style={{ textAlign: 'center', marginBottom: 32 }}>
                        <div style={{ display: 'flex', alignItems: 'baseline', justifyContent: 'center', gap: 4 }}>
                          <Text style={{ fontSize: 16, color: '#64748b' }}>¥</Text>
                          <Title level={1} style={{ 
                            margin: 0, 
                            fontSize: 48, 
                            fontWeight: 700,
                            color: plan.color,
                          }}>
                            {plan.price}
                          </Title>
                        </div>
                        <Text type="secondary" style={{ fontSize: 14 }}>
                          {plan.period}
                        </Text>
                      </div>

                      {/* 配额信息 */}
                      <div style={{ marginBottom: 24 }}>
                        <Row gutter={[8, 8]}>
                          <Col span={12}>
                            <div style={{ textAlign: 'center', padding: 8 }}>
                              <Text strong style={{ color: plan.color }}>{plan.quotas.video}</Text>
                              <br />
                              <Text type="secondary" style={{ fontSize: 12 }}>视频/月</Text>
                            </div>
                          </Col>
                          <Col span={12}>
                            <div style={{ textAlign: 'center', padding: 8 }}>
                              <Text strong style={{ color: plan.color }}>{plan.quotas.image}</Text>
                              <br />
                              <Text type="secondary" style={{ fontSize: 12 }}>图片/月</Text>
                            </div>
                          </Col>
                        </Row>
                      </div>

                      {/* 功能列表 */}
                      <List
                        dataSource={plan.features}
                        renderItem={(feature) => (
                          <List.Item style={{ padding: '8px 0', border: 'none' }}>
                            <Space>
                              <CheckOutlined style={{ color: plan.color }} />
                              <Text style={{ fontSize: 14 }}>{feature}</Text>
                            </Space>
                          </List.Item>
                        )}
                        style={{ marginBottom: 32 }}
                      />

                      {/* 订阅按钮 */}
                      <Button
                        type={plan.popular ? 'primary' : 'default'}
                        size="large"
                        block
                        onClick={() => handleSubscribe(plan.id)}
                        style={{
                          height: 48,
                          borderRadius: 8,
                          fontSize: 16,
                          fontWeight: 500,
                          ...(plan.popular ? {} : {
                            borderColor: plan.color,
                            color: plan.color,
                          }),
                        }}
                      >
                        {plan.buttonText}
                      </Button>
                    </Card>
                  </motion.div>
                </Col>
              ))}
            </Row>

            {/* 常见问题 */}
            <motion.div
              initial={{ opacity: 0, y: 20 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.5, delay: 0.5 }}
              style={{ marginTop: 80, textAlign: 'center' }}
            >
              <Title level={2} style={{ marginBottom: 32 }}>
                常见问题
              </Title>
              <Row gutter={[32, 32]}>
                <Col xs={24} md={8}>
                  <Card style={{ height: '100%', textAlign: 'left' }}>
                    <Title level={4}>可以随时取消订阅吗？</Title>
                    <Text type="secondary">
                      是的，您可以随时取消订阅。取消后，您仍可以使用剩余的配额直到当前计费周期结束。
                    </Text>
                  </Card>
                </Col>
                <Col xs={24} md={8}>
                  <Card style={{ height: '100%', textAlign: 'left' }}>
                    <Title level={4}>配额用完了怎么办？</Title>
                    <Text type="secondary">
                      配额用完后，您可以升级到更高级的套餐，或者等待下个月配额重置。
                    </Text>
                  </Card>
                </Col>
                <Col xs={24} md={8}>
                  <Card style={{ height: '100%', textAlign: 'left' }}>
                    <Title level={4}>支持哪些支付方式？</Title>
                    <Text type="secondary">
                      我们支持支付宝、微信支付、银行卡等多种支付方式，安全便捷。
                    </Text>
                  </Card>
                </Col>
              </Row>
            </motion.div>
          </div>
        </Content>
      </Layout>
    </>
  );
};

export default PricingPage;
