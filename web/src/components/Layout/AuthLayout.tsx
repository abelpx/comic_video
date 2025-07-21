import React from 'react';
import { Layout, Card } from 'antd';
import { Link } from 'react-router-dom';
import { motion } from 'framer-motion';

const { Content } = Layout;

interface AuthLayoutProps {
  children: React.ReactNode;
}

const AuthLayout: React.FC<AuthLayoutProps> = ({ children }) => {
  return (
    <Layout style={{
      minHeight: '100vh',
      background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      position: 'relative',
      overflow: 'hidden',
    }}>
      {/* 背景装饰 */}
      <div style={{
        position: 'absolute',
        top: 0,
        left: 0,
        right: 0,
        bottom: 0,
        background: `
          radial-gradient(circle at 20% 80%, rgba(120, 119, 198, 0.3) 0%, transparent 50%),
          radial-gradient(circle at 80% 20%, rgba(255, 119, 198, 0.3) 0%, transparent 50%),
          radial-gradient(circle at 40% 40%, rgba(120, 219, 255, 0.3) 0%, transparent 50%)
        `,
      }} />

      <Content style={{
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        padding: 24,
        position: 'relative',
        zIndex: 1,
      }}>
        <motion.div
          initial={{ opacity: 0, scale: 0.9 }}
          animate={{ opacity: 1, scale: 1 }}
          transition={{ duration: 0.5 }}
          style={{ width: '100%', maxWidth: 400 }}
        >
          <Card
            style={{
              background: 'rgba(255, 255, 255, 0.95)',
              backdropFilter: 'blur(20px)',
              border: '1px solid rgba(255, 255, 255, 0.2)',
              borderRadius: 16,
              boxShadow: '0 20px 40px rgba(0, 0, 0, 0.1)',
            }}
            bodyStyle={{ padding: 40 }}
          >
            {/* Logo */}
            <div style={{
              textAlign: 'center',
              marginBottom: 32,
            }}>
              <Link to="/" style={{ textDecoration: 'none' }}>
                <motion.div
                  whileHover={{ scale: 1.05 }}
                  transition={{ duration: 0.2 }}
                  style={{
                    display: 'inline-flex',
                    alignItems: 'center',
                    gap: 12,
                  }}
                >
                  <div style={{
                    width: 48,
                    height: 48,
                    background: 'linear-gradient(135deg, #0ea5e9 0%, #0284c7 100%)',
                    borderRadius: 12,
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'center',
                    color: 'white',
                    fontWeight: 'bold',
                    fontSize: 24,
                  }}>
                    C
                  </div>
                  <div>
                    <div style={{
                      fontSize: 24,
                      fontWeight: 700,
                      background: 'linear-gradient(135deg, #0ea5e9 0%, #0284c7 100%)',
                      WebkitBackgroundClip: 'text',
                      WebkitTextFillColor: 'transparent',
                      lineHeight: 1,
                    }}>
                      Comic AI
                    </div>
                    <div style={{
                      fontSize: 12,
                      color: '#64748b',
                      marginTop: 4,
                    }}>
                      AI驱动的创作平台
                    </div>
                  </div>
                </motion.div>
              </Link>
            </div>

            {/* 表单内容 */}
            {children}
          </Card>

          {/* 底部链接 */}
          <div style={{
            textAlign: 'center',
            marginTop: 24,
            color: 'rgba(255, 255, 255, 0.8)',
            fontSize: 14,
          }}>
            <Link 
              to="/" 
              style={{ 
                color: 'rgba(255, 255, 255, 0.9)',
                textDecoration: 'none',
                borderBottom: '1px solid rgba(255, 255, 255, 0.3)',
                paddingBottom: 2,
              }}
            >
              返回首页
            </Link>
          </div>
        </motion.div>
      </Content>

      {/* 浮动元素装饰 */}
      <motion.div
        animate={{
          y: [0, -20, 0],
          rotate: [0, 5, 0],
        }}
        transition={{
          duration: 6,
          repeat: Infinity,
          ease: "easeInOut",
        }}
        style={{
          position: 'absolute',
          top: '10%',
          left: '10%',
          width: 60,
          height: 60,
          background: 'rgba(255, 255, 255, 0.1)',
          borderRadius: '50%',
          backdropFilter: 'blur(10px)',
        }}
      />

      <motion.div
        animate={{
          y: [0, 30, 0],
          rotate: [0, -5, 0],
        }}
        transition={{
          duration: 8,
          repeat: Infinity,
          ease: "easeInOut",
          delay: 1,
        }}
        style={{
          position: 'absolute',
          top: '20%',
          right: '15%',
          width: 40,
          height: 40,
          background: 'rgba(255, 255, 255, 0.1)',
          borderRadius: '50%',
          backdropFilter: 'blur(10px)',
        }}
      />

      <motion.div
        animate={{
          y: [0, -15, 0],
          x: [0, 10, 0],
        }}
        transition={{
          duration: 7,
          repeat: Infinity,
          ease: "easeInOut",
          delay: 2,
        }}
        style={{
          position: 'absolute',
          bottom: '20%',
          left: '20%',
          width: 80,
          height: 80,
          background: 'rgba(255, 255, 255, 0.05)',
          borderRadius: '50%',
          backdropFilter: 'blur(10px)',
        }}
      />
    </Layout>
  );
};

export default AuthLayout;
