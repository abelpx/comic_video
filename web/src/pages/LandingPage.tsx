import React from 'react';
import { Button, Layout, Typography, Row, Col, Card, Space, Statistic } from 'antd';
import { Link, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { Helmet } from 'react-helmet-async';
import {
  VideoCameraOutlined,
  PictureOutlined,
  BookOutlined,
  RocketOutlined,
  CheckCircleOutlined,
  StarOutlined,
  UserOutlined,
  PlayCircleOutlined,
} from '@ant-design/icons';
import { useAuth } from '../store';

const { Header, Content, Footer } = Layout;
const { Title, Paragraph, Text } = Typography;

const LandingPage: React.FC = () => {
  const navigate = useNavigate();
  const { isAuthenticated } = useAuth();

  // 功能特性
  const features = [
    {
      icon: <VideoCameraOutlined style={{ fontSize: 32, color: '#0ea5e9' }} />,
      title: '小说转视频',
      description: '将文字小说智能转换为精美的动漫视频，支持角色一致性和场景连贯性',
    },
    {
      icon: <PictureOutlined style={{ fontSize: 32, color: '#a855f7' }} />,
      title: '漫画生成',
      description: '基于文本描述生成高质量漫画分镜，支持多种画风和风格定制',
    },
    {
      icon: <BookOutlined style={{ fontSize: 32, color: '#06b6d4' }} />,
      title: '小说创作',
      description: 'AI辅助小说创作，提供情节构思、角色设定和文本润色功能',
    },
    {
      icon: <PlayCircleOutlined style={{ fontSize: 32, color: '#10b981' }} />,
      title: '视频转动漫',
      description: '将真实视频转换为动漫风格，保持动作流畅性和画面美感',
    },
  ];

  // 优势特点
  const advantages = [
    '🎨 专业级AI模型，确保生成质量',
    '⚡ 快速生成，节省创作时间',
    '🎯 角色一致性，场景连贯性',
    '🔧 多样化工具，满足不同需求',
    '💡 简单易用，无需专业技能',
    '🌟 持续更新，功能不断完善',
  ];

  return (
    <>
      <Helmet>
        <title>Comic AI - AI驱动的创作平台</title>
        <meta name="description" content="使用AI技术将小说转换为动漫视频，生成漫画和创作内容的专业平台" />
        <meta name="keywords" content="AI,动漫,视频生成,漫画,小说,创作" />
      </Helmet>

      <Layout style={{ minHeight: '100vh' }}>
        {/* 导航栏 */}
        <Header style={{
          background: 'rgba(255, 255, 255, 0.95)',
          backdropFilter: 'blur(10px)',
          borderBottom: '1px solid #f0f0f0',
          position: 'fixed',
          top: 0,
          width: '100%',
          zIndex: 1000,
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
            {/* Logo */}
            <Link to="/" style={{ textDecoration: 'none' }}>
              <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                <div style={{
                  width: 40,
                  height: 40,
                  background: 'linear-gradient(135deg, #0ea5e9 0%, #0284c7 100%)',
                  borderRadius: 10,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  color: 'white',
                  fontWeight: 'bold',
                  fontSize: 20,
                }}>
                  C
                </div>
                <span style={{
                  fontSize: 20,
                  fontWeight: 600,
                  background: 'linear-gradient(135deg, #0ea5e9 0%, #0284c7 100%)',
                  WebkitBackgroundClip: 'text',
                  WebkitTextFillColor: 'transparent',
                }}>
                  Comic AI
                </span>
              </div>
            </Link>

            {/* 导航菜单 */}
            <Space size="large">
              <Link to="/pricing" style={{ color: '#64748b', fontWeight: 500 }}>定价</Link>
              {isAuthenticated ? (
                <Button type="primary" onClick={() => navigate('/app/dashboard')}>
                  进入应用
                </Button>
              ) : (
                <Space>
                  <Button type="text" onClick={() => navigate('/auth/login')}>
                    登录
                  </Button>
                  <Button type="primary" onClick={() => navigate('/auth/register')}>
                    免费注册
                  </Button>
                </Space>
              )}
            </Space>
          </div>
        </Header>

        <Content style={{ paddingTop: 64 }}>
          {/* Hero Section */}
          <section style={{
            background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
            padding: '120px 24px',
            textAlign: 'center',
            color: 'white',
            position: 'relative',
            overflow: 'hidden',
          }}>
            <div style={{ maxWidth: 800, margin: '0 auto', position: 'relative', zIndex: 1 }}>
              <motion.div
                initial={{ opacity: 0, y: 30 }}
                animate={{ opacity: 1, y: 0 }}
                transition={{ duration: 0.8 }}
              >
                <Title level={1} style={{ 
                  color: 'white', 
                  fontSize: 48, 
                  marginBottom: 24,
                  fontWeight: 700,
                }}>
                  AI驱动的创作平台
                </Title>
                <Paragraph style={{ 
                  fontSize: 20, 
                  color: 'rgba(255,255,255,0.9)', 
                  marginBottom: 40,
                  lineHeight: 1.6,
                }}>
                  将您的创意转化为精美的动漫视频和漫画作品
                  <br />
                  专业AI技术，简单易用，让创作变得更加轻松
                </Paragraph>
                <Space size="large">
                  <Button 
                    type="primary" 
                    size="large" 
                    icon={<RocketOutlined />}
                    onClick={() => navigate(isAuthenticated ? '/app/dashboard' : '/auth/register')}
                    style={{
                      height: 48,
                      padding: '0 32px',
                      fontSize: 16,
                      borderRadius: 24,
                    }}
                  >
                    {isAuthenticated ? '进入应用' : '免费开始'}
                  </Button>
                  <Button 
                    size="large"
                    style={{
                      height: 48,
                      padding: '0 32px',
                      fontSize: 16,
                      borderRadius: 24,
                      background: 'rgba(255,255,255,0.1)',
                      border: '1px solid rgba(255,255,255,0.3)',
                      color: 'white',
                    }}
                  >
                    观看演示
                  </Button>
                </Space>
              </motion.div>
            </div>

            {/* 背景装饰 */}
            <div style={{
              position: 'absolute',
              top: 0,
              left: 0,
              right: 0,
              bottom: 0,
              background: `
                radial-gradient(circle at 20% 80%, rgba(120, 119, 198, 0.3) 0%, transparent 50%),
                radial-gradient(circle at 80% 20%, rgba(255, 119, 198, 0.3) 0%, transparent 50%)
              `,
            }} />
          </section>

          {/* 统计数据 */}
          <section style={{ padding: '80px 24px', background: '#ffffff' }}>
            <div style={{ maxWidth: 1200, margin: '0 auto' }}>
              <Row gutter={[32, 32]} justify="center">
                <Col xs={12} sm={6}>
                  <div style={{ textAlign: 'center' }}>
                    <Statistic
                      title="用户数量"
                      value={10000}
                      suffix="+"
                      valueStyle={{ color: '#0ea5e9', fontSize: 32, fontWeight: 700 }}
                    />
                  </div>
                </Col>
                <Col xs={12} sm={6}>
                  <div style={{ textAlign: 'center' }}>
                    <Statistic
                      title="生成作品"
                      value={50000}
                      suffix="+"
                      valueStyle={{ color: '#a855f7', fontSize: 32, fontWeight: 700 }}
                    />
                  </div>
                </Col>
                <Col xs={12} sm={6}>
                  <div style={{ textAlign: 'center' }}>
                    <Statistic
                      title="满意度"
                      value={98}
                      suffix="%"
                      valueStyle={{ color: '#10b981', fontSize: 32, fontWeight: 700 }}
                    />
                  </div>
                </Col>
                <Col xs={12} sm={6}>
                  <div style={{ textAlign: 'center' }}>
                    <Statistic
                      title="处理时间"
                      value={30}
                      suffix="秒"
                      valueStyle={{ color: '#f59e0b', fontSize: 32, fontWeight: 700 }}
                    />
                  </div>
                </Col>
              </Row>
            </div>
          </section>

          {/* 功能特性 */}
          <section style={{ padding: '80px 24px', background: '#f8fafc' }}>
            <div style={{ maxWidth: 1200, margin: '0 auto' }}>
              <div style={{ textAlign: 'center', marginBottom: 64 }}>
                <Title level={2} style={{ marginBottom: 16 }}>强大的AI创作工具</Title>
                <Paragraph style={{ fontSize: 16, color: '#64748b', maxWidth: 600, margin: '0 auto' }}>
                  我们提供多种AI驱动的创作工具，帮助您轻松创建专业级的动漫内容
                </Paragraph>
              </div>

              <Row gutter={[32, 32]}>
                {features.map((feature, index) => (
                  <Col xs={24} sm={12} lg={6} key={index}>
                    <motion.div
                      initial={{ opacity: 0, y: 20 }}
                      whileInView={{ opacity: 1, y: 0 }}
                      transition={{ duration: 0.5, delay: index * 0.1 }}
                      viewport={{ once: true }}
                    >
                      <Card
                        hoverable
                        style={{
                          height: '100%',
                          borderRadius: 12,
                          border: '1px solid #e2e8f0',
                          boxShadow: '0 1px 3px rgba(0,0,0,0.1)',
                        }}
                        bodyStyle={{ padding: 24, textAlign: 'center' }}
                      >
                        <div style={{ marginBottom: 16 }}>
                          {feature.icon}
                        </div>
                        <Title level={4} style={{ marginBottom: 12 }}>
                          {feature.title}
                        </Title>
                        <Paragraph style={{ color: '#64748b', margin: 0 }}>
                          {feature.description}
                        </Paragraph>
                      </Card>
                    </motion.div>
                  </Col>
                ))}
              </Row>
            </div>
          </section>

          {/* 优势特点 */}
          <section style={{ padding: '80px 24px', background: '#ffffff' }}>
            <div style={{ maxWidth: 1200, margin: '0 auto' }}>
              <Row gutter={[64, 32]} align="middle">
                <Col xs={24} lg={12}>
                  <Title level={2} style={{ marginBottom: 24 }}>
                    为什么选择 Comic AI？
                  </Title>
                  <div style={{ marginBottom: 32 }}>
                    {advantages.map((advantage, index) => (
                      <motion.div
                        key={index}
                        initial={{ opacity: 0, x: -20 }}
                        whileInView={{ opacity: 1, x: 0 }}
                        transition={{ duration: 0.5, delay: index * 0.1 }}
                        viewport={{ once: true }}
                        style={{
                          display: 'flex',
                          alignItems: 'center',
                          marginBottom: 16,
                          fontSize: 16,
                        }}
                      >
                        <CheckCircleOutlined style={{ color: '#10b981', marginRight: 12 }} />
                        {advantage}
                      </motion.div>
                    ))}
                  </div>
                  <Button 
                    type="primary" 
                    size="large"
                    onClick={() => navigate(isAuthenticated ? '/app/dashboard' : '/auth/register')}
                  >
                    立即体验
                  </Button>
                </Col>
                <Col xs={24} lg={12}>
                  <div style={{
                    background: 'linear-gradient(135deg, #f0f9ff 0%, #e0f2fe 100%)',
                    borderRadius: 16,
                    padding: 40,
                    textAlign: 'center',
                  }}>
                    <div style={{
                      width: 120,
                      height: 120,
                      background: 'linear-gradient(135deg, #0ea5e9 0%, #0284c7 100%)',
                      borderRadius: '50%',
                      display: 'flex',
                      alignItems: 'center',
                      justifyContent: 'center',
                      margin: '0 auto 24px',
                      color: 'white',
                      fontSize: 48,
                    }}>
                      <StarOutlined />
                    </div>
                    <Title level={3} style={{ marginBottom: 16 }}>
                      专业品质保证
                    </Title>
                    <Paragraph style={{ color: '#64748b' }}>
                      我们使用最先进的AI技术，确保每一个作品都达到专业级别的质量标准
                    </Paragraph>
                  </div>
                </Col>
              </Row>
            </div>
          </section>
        </Content>

        {/* 页脚 */}
        <Footer style={{ background: '#1e293b', color: '#94a3b8', textAlign: 'center', padding: '40px 24px' }}>
          <div style={{ maxWidth: 1200, margin: '0 auto' }}>
            <div style={{ marginBottom: 24 }}>
              <Link to="/" style={{ textDecoration: 'none' }}>
                <span style={{
                  fontSize: 20,
                  fontWeight: 600,
                  color: '#0ea5e9',
                }}>
                  Comic AI
                </span>
              </Link>
            </div>
            <Paragraph style={{ color: '#94a3b8', margin: 0 }}>
              © 2024 Comic AI. All rights reserved. | 
              <Link to="/privacy" style={{ color: '#94a3b8', marginLeft: 8 }}>隐私政策</Link> | 
              <Link to="/terms" style={{ color: '#94a3b8', marginLeft: 8 }}>服务条款</Link>
            </Paragraph>
          </div>
        </Footer>
      </Layout>
    </>
  );
};

export default LandingPage;
