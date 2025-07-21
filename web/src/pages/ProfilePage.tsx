import React, { useState } from 'react';
import { Card, Row, Col, Typography, Form, Input, Button, Avatar, Upload, Switch, Divider, Space, message } from 'antd';
import { motion } from 'framer-motion';
import {
  UserOutlined,
  MailOutlined,
  PhoneOutlined,
  CameraOutlined,
  SaveOutlined,
  LockOutlined,
  BellOutlined,
  EyeInvisibleOutlined,
  EyeTwoTone,
} from '@ant-design/icons';
import { useUser, useAuth } from '../store';
import { authApi } from '../services/api';

const { Title, Text } = Typography;

interface ProfileForm {
  username: string;
  email: string;
  nickname: string;
  phone?: string;
  bio?: string;
}

interface PasswordForm {
  currentPassword: string;
  newPassword: string;
  confirmPassword: string;
}

interface NotificationSettings {
  emailNotifications: boolean;
  taskNotifications: boolean;
  marketingEmails: boolean;
  securityAlerts: boolean;
}

const ProfilePage: React.FC = () => {
  const user = useUser();
  const { logout } = useAuth();
  const [profileForm] = Form.useForm();
  const [passwordForm] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [passwordLoading, setPasswordLoading] = useState(false);
  const [notifications, setNotifications] = useState<NotificationSettings>({
    emailNotifications: true,
    taskNotifications: true,
    marketingEmails: false,
    securityAlerts: true,
  });

  // 初始化表单数据
  React.useEffect(() => {
    if (user) {
      profileForm.setFieldsValue({
        username: user.username,
        email: user.email,
        nickname: user.nickname,
        phone: user.phone,
        bio: user.bio,
      });
    }
  }, [user, profileForm]);

  const handleProfileSubmit = async (values: ProfileForm) => {
    setLoading(true);
    try {
      await authApi.updateProfile(values);
      message.success('个人信息更新成功');
    } catch (error) {
      message.error('更新失败，请稍后重试');
    } finally {
      setLoading(false);
    }
  };

  const handlePasswordSubmit = async (values: PasswordForm) => {
    setPasswordLoading(true);
    try {
      // TODO: 实现密码修改API
      console.log('修改密码:', values);
      message.success('密码修改成功，请重新登录');
      passwordForm.resetFields();
      // 可选择是否自动登出
      // logout();
    } catch (error) {
      message.error('密码修改失败，请稍后重试');
    } finally {
      setPasswordLoading(false);
    }
  };

  const handleAvatarChange = (info: any) => {
    if (info.file.status === 'done') {
      message.success('头像上传成功');
    } else if (info.file.status === 'error') {
      message.error('头像上传失败');
    }
  };

  const handleNotificationChange = (key: keyof NotificationSettings, value: boolean) => {
    setNotifications(prev => ({ ...prev, [key]: value }));
    // TODO: 保存通知设置到后端
    message.success('通知设置已更新');
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
          个人资料
        </Title>
        <Text type="secondary" style={{ fontSize: 16 }}>
          管理您的个人信息、安全设置和通知偏好。
        </Text>
      </motion.div>

      <Row gutter={[24, 24]}>
        {/* 基本信息 */}
        <Col xs={24} lg={16}>
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.1 }}
          >
            <Card title="基本信息" style={{ marginBottom: 24 }}>
              <Form
                form={profileForm}
                layout="vertical"
                onFinish={handleProfileSubmit}
                size="large"
              >
                <Row gutter={[16, 0]}>
                  <Col xs={24} sm={12}>
                    <Form.Item
                      name="username"
                      label="用户名"
                      rules={[
                        { required: true, message: '请输入用户名' },
                        { min: 3, message: '用户名至少3个字符' },
                      ]}
                    >
                      <Input
                        prefix={<UserOutlined style={{ color: '#94a3b8' }} />}
                        placeholder="请输入用户名"
                      />
                    </Form.Item>
                  </Col>
                  <Col xs={24} sm={12}>
                    <Form.Item
                      name="nickname"
                      label="昵称"
                      rules={[{ max: 20, message: '昵称最多20个字符' }]}
                    >
                      <Input
                        prefix={<UserOutlined style={{ color: '#94a3b8' }} />}
                        placeholder="请输入昵称"
                      />
                    </Form.Item>
                  </Col>
                </Row>

                <Row gutter={[16, 0]}>
                  <Col xs={24} sm={12}>
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
                      />
                    </Form.Item>
                  </Col>
                  <Col xs={24} sm={12}>
                    <Form.Item
                      name="phone"
                      label="手机号码"
                      rules={[
                        { pattern: /^1[3-9]\d{9}$/, message: '请输入有效的手机号码' },
                      ]}
                    >
                      <Input
                        prefix={<PhoneOutlined style={{ color: '#94a3b8' }} />}
                        placeholder="请输入手机号码"
                      />
                    </Form.Item>
                  </Col>
                </Row>

                <Form.Item
                  name="bio"
                  label="个人简介"
                  rules={[{ max: 200, message: '个人简介最多200个字符' }]}
                >
                  <Input.TextArea
                    rows={4}
                    placeholder="介绍一下自己吧..."
                    showCount
                    maxLength={200}
                  />
                </Form.Item>

                <Form.Item>
                  <Button
                    type="primary"
                    htmlType="submit"
                    loading={loading}
                    icon={<SaveOutlined />}
                    size="large"
                  >
                    保存更改
                  </Button>
                </Form.Item>
              </Form>
            </Card>
          </motion.div>

          {/* 密码修改 */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.2 }}
          >
            <Card title="修改密码">
              <Form
                form={passwordForm}
                layout="vertical"
                onFinish={handlePasswordSubmit}
                size="large"
              >
                <Form.Item
                  name="currentPassword"
                  label="当前密码"
                  rules={[{ required: true, message: '请输入当前密码' }]}
                >
                  <Input.Password
                    prefix={<LockOutlined style={{ color: '#94a3b8' }} />}
                    placeholder="请输入当前密码"
                    iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
                  />
                </Form.Item>

                <Form.Item
                  name="newPassword"
                  label="新密码"
                  rules={[
                    { required: true, message: '请输入新密码' },
                    { min: 6, message: '密码至少6个字符' },
                  ]}
                >
                  <Input.Password
                    prefix={<LockOutlined style={{ color: '#94a3b8' }} />}
                    placeholder="请输入新密码"
                    iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
                  />
                </Form.Item>

                <Form.Item
                  name="confirmPassword"
                  label="确认新密码"
                  dependencies={['newPassword']}
                  rules={[
                    { required: true, message: '请确认新密码' },
                    ({ getFieldValue }) => ({
                      validator(_, value) {
                        if (!value || getFieldValue('newPassword') === value) {
                          return Promise.resolve();
                        }
                        return Promise.reject(new Error('两次输入的密码不一致'));
                      },
                    }),
                  ]}
                >
                  <Input.Password
                    prefix={<LockOutlined style={{ color: '#94a3b8' }} />}
                    placeholder="请再次输入新密码"
                    iconRender={(visible) => (visible ? <EyeTwoTone /> : <EyeInvisibleOutlined />)}
                  />
                </Form.Item>

                <Form.Item>
                  <Button
                    type="primary"
                    htmlType="submit"
                    loading={passwordLoading}
                    icon={<SaveOutlined />}
                    size="large"
                  >
                    修改密码
                  </Button>
                </Form.Item>
              </Form>
            </Card>
          </motion.div>
        </Col>

        {/* 侧边栏 */}
        <Col xs={24} lg={8}>
          {/* 头像设置 */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.3 }}
          >
            <Card title="头像设置" style={{ marginBottom: 24 }}>
              <div style={{ textAlign: 'center' }}>
                <Avatar
                  size={120}
                  src={user?.avatar}
                  icon={<UserOutlined />}
                  style={{ 
                    backgroundColor: '#0ea5e9',
                    marginBottom: 16,
                  }}
                />
                <div>
                  <Upload
                    name="avatar"
                    showUploadList={false}
                    action="/api/v1/upload/avatar"
                    onChange={handleAvatarChange}
                    accept="image/*"
                  >
                    <Button icon={<CameraOutlined />}>
                      更换头像
                    </Button>
                  </Upload>
                </div>
                <Text type="secondary" style={{ fontSize: 12, display: 'block', marginTop: 8 }}>
                  支持 JPG、PNG 格式，文件大小不超过 2MB
                </Text>
              </div>
            </Card>
          </motion.div>

          {/* 通知设置 */}
          <motion.div
            initial={{ opacity: 0, y: 20 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 0.5, delay: 0.4 }}
          >
            <Card title={<Space><BellOutlined />通知设置</Space>}>
              <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <div>
                    <div style={{ fontWeight: 500 }}>邮件通知</div>
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      接收重要的邮件通知
                    </Text>
                  </div>
                  <Switch
                    checked={notifications.emailNotifications}
                    onChange={(checked) => handleNotificationChange('emailNotifications', checked)}
                  />
                </div>

                <Divider style={{ margin: 0 }} />

                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <div>
                    <div style={{ fontWeight: 500 }}>任务通知</div>
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      任务完成时发送通知
                    </Text>
                  </div>
                  <Switch
                    checked={notifications.taskNotifications}
                    onChange={(checked) => handleNotificationChange('taskNotifications', checked)}
                  />
                </div>

                <Divider style={{ margin: 0 }} />

                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <div>
                    <div style={{ fontWeight: 500 }}>营销邮件</div>
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      接收产品更新和优惠信息
                    </Text>
                  </div>
                  <Switch
                    checked={notifications.marketingEmails}
                    onChange={(checked) => handleNotificationChange('marketingEmails', checked)}
                  />
                </div>

                <Divider style={{ margin: 0 }} />

                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <div>
                    <div style={{ fontWeight: 500 }}>安全提醒</div>
                    <Text type="secondary" style={{ fontSize: 12 }}>
                      账户安全相关通知
                    </Text>
                  </div>
                  <Switch
                    checked={notifications.securityAlerts}
                    onChange={(checked) => handleNotificationChange('securityAlerts', checked)}
                  />
                </div>
              </div>
            </Card>
          </motion.div>
        </Col>
      </Row>
    </div>
  );
};

export default ProfilePage;
