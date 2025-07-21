import React, { useState } from 'react';
import { Form, Input, Button, Typography, Divider, Checkbox, Alert } from 'antd';
import { Link, useNavigate } from 'react-router-dom';
import { motion } from 'framer-motion';
import { UserOutlined, MailOutlined, LockOutlined, EyeInvisibleOutlined, EyeTwoTone } from '@ant-design/icons';
import { useAuth } from '../../store';
import { authApi } from '../../services/api';
import toast from 'react-hot-toast';

const { Title, Text } = Typography;

interface RegisterForm {
  username: string;
  email: string;
  password: string;
  confirmPassword: string;
  nickname?: string;
  agreement: boolean;
}

const RegisterPage: React.FC = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const { login } = useAuth();
  const navigate = useNavigate();

  const handleSubmit = async (values: RegisterForm) => {
    setLoading(true);
    setError(null);

    try {
      const response = await authApi.register({
        username: values.username,
        email: values.email,
        password: values.password,
        nickname: values.nickname,
      });

      // 注册成功，自动登录
      login(response.user, response.token);
      toast.success('注册成功！欢迎使用 Comic AI');
      
      // 跳转到仪表板
      navigate('/app/dashboard', { replace: true });
    } catch (err: any) {
      const errorMessage = err.response?.data?.message || '注册失败，请稍后重试';
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
          创建账户
        </Title>
        <Text type="secondary">
          加入 Comic AI，开启您的创作之旅
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
          label="用户名"
          rules={[
            { required: true, message: '请输入用户名' },
            { min: 3, message: '用户名至少3个字符' },
            { max: 20, message: '用户名最多20个字符' },
            { pattern: /^[a-zA-Z0-9_]+$/, message: '用户名只能包含字母、数字和下划线' },
          ]}
        >
          <Input
            prefix={<UserOutlined style={{ color: '#94a3b8' }} />}
            placeholder="请输入用户名"
            style={{ borderRadius: 8 }}
          />
        </Form.Item>

        <Form.Item
          name="email"
          label="邮箱地址"
          rules={[
            { required: true, message: '请输入邮箱地址' },
            { type: 'email', message: '请输入有效的邮箱地址' },
          ]}
        >
          <Input
            prefix={<MailOutlined style={{ color: '#94a3b8' }} />}
            placeholder="请输入邮箱地址"
            style={{ borderRadius: 8 }}
          />
        </Form.Item>

        <Form.Item
          name="nickname"
          label="昵称（可选）"
          rules={[
            { max: 20, message: '昵称最多20个字符' },
          ]}
        >
          <Input
            prefix={<UserOutlined style={{ color: '#94a3b8' }} />}
            placeholder="请输入昵称"
            style={{ borderRadius: 8 }}
          />
        </Form.Item>

        <Form.Item
          name="password"
          label="密码"
          rules={[
            { required: true, message: '请输入密码' },
            { min: 6, message: '密码至少6个字符' },
            { max: 50, message: '密码最多50个字符' },
          ]}
        >
          <Input.Password
            prefix={<LockOutlined style={{ color: '#94a3b8' }} />}
            placeholder="请输入密码"
            iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
            style={{ borderRadius: 8 }}
          />
        </Form.Item>

        <Form.Item
          name="confirmPassword"
          label="确认密码"
          dependencies={['password']}
          rules={[
            { required: true, message: '请确认密码' },
            ({ getFieldValue }) => ({
              validator(_, value) {
                if (!value || getFieldValue('password') === value) {
                  return Promise.resolve();
                }
                return Promise.reject(new Error('两次输入的密码不一致'));
              },
            }),
          ]}
        >
          <Input.Password
            prefix={<LockOutlined style={{ color: '#94a3b8' }} />}
            placeholder="请再次输入密码"
            iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
            style={{ borderRadius: 8 }}
          />
        </Form.Item>

        <Form.Item
          name="agreement"
          valuePropName="checked"
          rules={[
            { 
              validator: (_, value) =>
                value ? Promise.resolve() : Promise.reject(new Error('请同意服务条款和隐私政策'))
            },
          ]}
          style={{ marginBottom: 24 }}
        >
          <Checkbox>
            我已阅读并同意{' '}
            <Link to="/terms" target="_blank" style={{ color: '#0ea5e9' }}>
              服务条款
            </Link>
            {' '}和{' '}
            <Link to="/privacy" target="_blank" style={{ color: '#0ea5e9' }}>
              隐私政策
            </Link>
          </Checkbox>
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
            {loading ? '注册中...' : '创建账户'}
          </Button>
        </Form.Item>

        <Divider style={{ margin: '24px 0' }}>
          <Text type="secondary" style={{ fontSize: 14 }}>
            或者
          </Text>
        </Divider>

        <div style={{ textAlign: 'center' }}>
          <Text type="secondary" style={{ fontSize: 14 }}>
            已有账户？{' '}
            <Link 
              to="/auth/login" 
              style={{ 
                color: '#0ea5e9',
                fontWeight: 500,
              }}
            >
              立即登录
            </Link>
          </Text>
        </div>
      </Form>

      {/* 注册福利提示 */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        transition={{ delay: 0.5, duration: 0.5 }}
        style={{
          marginTop: 32,
          padding: 16,
          background: '#f0fdf4',
          border: '1px solid #bbf7d0',
          borderRadius: 8,
          textAlign: 'center',
        }}
      >
        <Text style={{ fontSize: 12, color: '#15803d' }}>
          🎉 新用户注册即可获得免费配额，立即开始创作！
        </Text>
      </motion.div>
    </motion.div>
  );
};

export default RegisterPage;
