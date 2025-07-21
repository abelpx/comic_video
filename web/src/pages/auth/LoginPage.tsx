import React, { useState } from 'react';
import { Form, Input, Button, Typography, Divider, Space, Alert } from 'antd';
import { Link, useNavigate, useLocation } from 'react-router-dom';
import { motion } from 'framer-motion';
import { UserOutlined, LockOutlined, EyeInvisibleOutlined, EyeTwoTone } from '@ant-design/icons';
import { useAuth } from '../../store';
import { authApi } from '../../services/api';
import toast from 'react-hot-toast';

const { Title, Text } = Typography;

interface LoginForm {
  username: string;
  password: string;
}

const LoginPage: React.FC = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { login } = useAuth();
  const navigate = useNavigate();
  const location = useLocation();

  // 获取重定向路径
  const from = (location.state as any)?.from?.pathname || '/app/dashboard';

  const handleSubmit = async (values: LoginForm) => {
    setLoading(true);
    setError(null);

    try {
      const response = await authApi.login({
        username: values.username,
        password: values.password,
      });

      // 登录成功
      login(response.user, response.token);
      toast.success('登录成功！');
      
      // 重定向到目标页面
      navigate(from, { replace: true });
    } catch (err: any) {
      const errorMessage = err.response?.data?.message || '登录失败，请检查用户名和密码';
      setError(errorMessage);
      toast.error(errorMessage);
    } finally {
      setLoading(false);
    }
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.5 }}
    >
      <div style={{ textAlign: 'center', marginBottom: 32 }}>
        <Title level={2} style={{ marginBottom: 8 }}>
          欢迎回来
        </Title>
        <Text type="secondary">
          登录您的账户以继续使用 Comic AI
        </Text>
      </div>

      {error && (
        <Alert
          message={error}
          type="error"
          showIcon
          closable
          onClose={() => setError(null)}
          style={{ marginBottom: 24 }}
        />
      )}

      <Form
        form={form}
        layout="vertical"
        onFinish={handleSubmit}
        size="large"
        requiredMark={false}
      >
        <Form.Item
          name="username"
          label="用户名或邮箱"
          rules={[
            { required: true, message: '请输入用户名或邮箱' },
            { min: 3, message: '用户名至少3个字符' },
          ]}
        >
          <Input
            prefix={<UserOutlined style={{ color: '#94a3b8' }} />}
            placeholder="请输入用户名或邮箱"
            style={{ borderRadius: 8 }}
          />
        </Form.Item>

        <Form.Item
          name="password"
          label="密码"
          rules={[
            { required: true, message: '请输入密码' },
            { min: 6, message: '密码至少6个字符' },
          ]}
        >
          <Input.Password
            prefix={<LockOutlined style={{ color: '#94a3b8' }} />}
            placeholder="请输入密码"
            iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
            style={{ borderRadius: 8 }}
          />
        </Form.Item>

        <Form.Item style={{ marginBottom: 16 }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <div></div>
            <Link 
              to="/auth/forgot-password" 
              style={{ 
                color: '#0ea5e9',
                fontSize: 14,
              }}
            >
              忘记密码？
            </Link>
          </div>
        </Form.Item>

        <Form.Item style={{ marginBottom: 24 }}>
          <Button
            type="primary"
            htmlType="submit"
            loading={loading}
            block
            style={{
              height: 48,
              borderRadius: 8,
              fontSize: 16,
              fontWeight: 500,
            }}
          >
            {loading ? '登录中...' : '登录'}
          </Button>
        </Form.Item>

        <Divider style={{ margin: '24px 0' }}>
          <Text type="secondary" style={{ fontSize: 14 }}>
            或者
          </Text>
        </Divider>

        <div style={{ textAlign: 'center' }}>
          <Text type="secondary" style={{ fontSize: 14 }}>
            还没有账户？{' '}
            <Link 
              to="/auth/register" 
              style={{ 
                color: '#0ea5e9',
                fontWeight: 500,
              }}
            >
              立即注册
            </Link>
          </Text>
        </div>
      </Form>

      {/* 演示账户提示 */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.5, duration: 0.5 }}
        style={{
          marginTop: 32,
          padding: 16,
          background: '#f0f9ff',
          border: '1px solid #bae6fd',
          borderRadius: 8,
          textAlign: 'center',
        }}
      >
        <Text style={{ fontSize: 12, color: '#0369a1' }}>
          💡 演示账户：demo / 123456
        </Text>
      </motion.div>
    </motion.div>
  );
};

export default LoginPage;
