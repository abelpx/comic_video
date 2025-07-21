import React from 'react';
import { Layout, Menu, Avatar, Dropdown, Badge, Button, Tooltip } from 'antd';
import { Link, useLocation, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import {
  DashboardOutlined,
  VideoCameraOutlined,
  PictureOutlined,
  TwitterOutlined,
  BookOutlined,
  PlayCircleOutlined,
  FolderOutlined,
  BarChartOutlined,
  UserOutlined,
  SettingOutlined,
  LogoutOutlined,
  BellOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  SunOutlined,
  MoonOutlined,
} from '@ant-design/icons';
import { useAuth, useUser, useUI, useTasks } from '../../store';

const { Header, Sider, Content } = Layout;

interface AppLayoutProps {
  children: React.ReactNode;
}

const AppLayout: React.FC<AppLayoutProps> = ({ children }) => {
  const location = useLocation();
  const navigate = useNavigate();
  const { logout } = useAuth();
  const user = useUser();
  const { sidebarCollapsed, setSidebarCollapsed, theme, setTheme } = useUI();
  const { activeTasks } = useTasks();

  // 菜单项配置
  const menuItems = [
    {
      key: '/app/dashboard',
      icon: <DashboardOutlined />,
      label: <Link to="/app/dashboard">仪表板</Link>,
    },
    {
      key: 'ai-tools',
      icon: <VideoCameraOutlined />,
      label: 'AI创作工具',
      children: [
        {
          key: '/app/novel-to-video',
          icon: <VideoCameraOutlined />,
          label: <Link to="/app/novel-to-video">小说转视频</Link>,
        },
        {
          key: '/app/generate-comic',
          icon: <PictureOutlined />,
          label: <Link to="/app/generate-comic">漫画生成</Link>,
        },
        {
          key: '/app/generate-tweet',
          icon: <TwitterOutlined />,
          label: <Link to="/app/generate-tweet">推文生成</Link>,
        },
        {
          key: '/app/generate-novel',
          icon: <BookOutlined />,
          label: <Link to="/app/generate-novel">小说生成</Link>,
        },
        {
          key: '/app/video-to-anime',
          icon: <PlayCircleOutlined />,
          label: <Link to="/app/video-to-anime">视频转动漫</Link>,
        },
      ],
    },
    {
      key: '/app/works',
      icon: <FolderOutlined />,
      label: <Link to="/app/works">我的作品</Link>,
    },
    {
      key: '/app/quota',
      icon: <BarChartOutlined />,
      label: <Link to="/app/quota">配额管理</Link>,
    },
  ];

  // 用户菜单
  const userMenuItems = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人资料',
      onClick: () => navigate('/app/profile'),
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '设置',
    },
    {
      type: 'divider' as const,
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      onClick: () => {
        logout();
        navigate('/');
      },
    },
  ];

  // 获取当前选中的菜单项
  const getSelectedKeys = () => {
    const path = location.pathname;
    // 如果是子菜单项，返回具体路径
    if (path.startsWith('/app/')) {
      return [path];
    }
    return ['/app/dashboard'];
  };

  // 获取展开的菜单项
  const getOpenKeys = () => {
    const path = location.pathname;
    if (path.includes('novel-to-video') || 
        path.includes('generate-comic') || 
        path.includes('generate-tweet') || 
        path.includes('generate-novel') || 
        path.includes('video-to-anime')) {
      return ['ai-tools'];
    }
    return [];
  };

  return (
    <Layout style={{ minHeight: '100vh' }}>
      {/* 侧边栏 */}
      <Sider
        trigger={null}
        collapsible
        collapsed={sidebarCollapsed}
        width={240}
        style={{
          background: '#ffffff',
          borderRight: '1px solid #f0f0f0',
          boxShadow: '2px 0 8px rgba(0,0,0,0.02)',
        }}
      >
        {/* Logo */}
        <div style={{
          height: 64,
          display: 'flex',
          alignItems: 'center',
          justifyContent: sidebarCollapsed ? 'center' : 'flex-start',
          padding: sidebarCollapsed ? 0 : '0 24px',
          borderBottom: '1px solid #f0f0f0',
        }}>
          <motion.div
            initial={false}
            animate={{ scale: sidebarCollapsed ? 0.8 : 1 }}
            transition={{ duration: 0.2 }}
          >
            {sidebarCollapsed ? (
              <div style={{
                width: 32,
                height: 32,
                background: 'linear-gradient(135deg, #0ea5e9 0%, #0284c7 100%)',
                borderRadius: 8,
                display: 'flex',
                alignItems: 'center',
                justifyContent: 'center',
                color: 'white',
                fontWeight: 'bold',
                fontSize: 16,
              }}>
                C
              </div>
            ) : (
              <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
                <div style={{
                  width: 32,
                  height: 32,
                  background: 'linear-gradient(135deg, #0ea5e9 0%, #0284c7 100%)',
                  borderRadius: 8,
                  display: 'flex',
                  alignItems: 'center',
                  justifyContent: 'center',
                  color: 'white',
                  fontWeight: 'bold',
                  fontSize: 16,
                }}>
                  C
                </div>
                <span style={{
                  fontSize: 18,
                  fontWeight: 600,
                  background: 'linear-gradient(135deg, #0ea5e9 0%, #0284c7 100%)',
                  WebkitBackgroundClip: 'text',
                  WebkitTextFillColor: 'transparent',
                }}>
                  Comic AI
                </span>
              </div>
            )}
          </motion.div>
        </div>

        {/* 菜单 */}
        <Menu
          mode="inline"
          selectedKeys={getSelectedKeys()}
          defaultOpenKeys={getOpenKeys()}
          items={menuItems}
          style={{
            border: 'none',
            height: 'calc(100vh - 64px)',
            overflow: 'auto',
          }}
        />
      </Sider>

      {/* 主内容区 */}
      <Layout>
        {/* 顶部导航 */}
        <Header style={{
          background: '#ffffff',
          padding: '0 24px',
          borderBottom: '1px solid #f0f0f0',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          boxShadow: '0 1px 4px rgba(0,0,0,0.02)',
        }}>
          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            {/* 折叠按钮 */}
            <Button
              type="text"
              icon={sidebarCollapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />}
              onClick={() => setSidebarCollapsed(!sidebarCollapsed)}
              style={{ fontSize: 16 }}
            />
          </div>

          <div style={{ display: 'flex', alignItems: 'center', gap: 16 }}>
            {/* 活跃任务提示 */}
            {activeTasks.length > 0 && (
              <Tooltip title={`${activeTasks.length} 个任务正在处理中`}>
                <Badge count={activeTasks.length} size="small">
                  <Button type="text" icon={<BellOutlined />} />
                </Badge>
              </Tooltip>
            )}

            {/* 主题切换 */}
            <Button
              type="text"
              icon={theme === 'dark' ? <SunOutlined /> : <MoonOutlined />}
              onClick={() => setTheme(theme === 'dark' ? 'light' : 'dark')}
            />

            {/* 用户菜单 */}
            <Dropdown menu={{ items: userMenuItems }} placement="bottomRight">
              <div style={{
                display: 'flex',
                alignItems: 'center',
                gap: 8,
                cursor: 'pointer',
                padding: '4px 8px',
                borderRadius: 8,
                transition: 'background-color 0.2s',
              }}>
                <Avatar
                  size="small"
                  src={user?.avatar}
                  icon={<UserOutlined />}
                  style={{ backgroundColor: '#0ea5e9' }}
                />
                <span style={{ fontSize: 14, fontWeight: 500 }}>
                  {user?.nickname || user?.username}
                </span>
              </div>
            </Dropdown>
          </div>
        </Header>

        {/* 内容区域 */}
        <Content style={{
          margin: 24,
          padding: 0,
          minHeight: 'calc(100vh - 112px)',
          overflow: 'auto',
        }}>
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -20 }}
            transition={{ duration: 0.3 }}
          >
            {children}
          </motion.div>
        </Content>
      </Layout>
    </Layout>
  );
};

export default AppLayout;
