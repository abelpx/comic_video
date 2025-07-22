import React, { useState, useEffect, useRef } from 'react';
import { 
  Card, 
  Steps, 
  Button, 
  Form, 
  Input, 
  Select, 
  Switch, 
  Progress, 
  message, 
  Spin,
  Typography,
  Space,
  Divider
} from 'antd';
import { 
  PlayCircleOutlined, 
  FileTextOutlined, 
  UserOutlined,
  PictureOutlined,
  VideoCameraOutlined,
  SoundOutlined,
  EditOutlined,
  ShareAltOutlined
} from '@ant-design/icons';
import { motion } from 'framer-motion';
import { Helmet } from 'react-helmet-async';
import { workflowService, NovelToVideoWorkflowRequest, WorkflowResponse, WorkflowTask } from '../services/workflow';

const { Title, Paragraph, Text } = Typography;
const { TextArea } = Input;
const { Option } = Select;

interface WorkflowStep {
  key: string;
  title: string;
  description: string;
  icon: React.ReactNode;
  status: 'wait' | 'process' | 'finish' | 'error';
}

const WorkflowPage: React.FC = () => {
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [workflowId, setWorkflowId] = useState<string | null>(null);
  const [currentStep, setCurrentStep] = useState(0);
  const [progress, setProgress] = useState(0);
  const [workflow, setWorkflow] = useState<WorkflowResponse | null>(null);
  const [tasks, setTasks] = useState<WorkflowTask[]>([]);
  const stopPollingRef = useRef<(() => void) | null>(null);

  const workflowSteps: WorkflowStep[] = [
    {
      key: 'script_adapt',
      title: '剧本改编',
      description: 'AI分析小说结构，改编为标准剧本格式',
      icon: <FileTextOutlined />,
      status: 'wait'
    },
    {
      key: 'character_gen',
      title: '角色生成',
      description: '提取角色信息，生成一致性角色图像',
      icon: <UserOutlined />,
      status: 'wait'
    },
    {
      key: 'scene_gen',
      title: '场景生成',
      description: '分析场景信息，生成概念场景图',
      icon: <PictureOutlined />,
      status: 'wait'
    },
    {
      key: 'storyboard',
      title: '分镜制作',
      description: '生成详细分镜脚本和分镜图',
      icon: <VideoCameraOutlined />,
      status: 'wait'
    },
    {
      key: 'voice_gen',
      title: '语音合成',
      description: '为角色对话生成AI语音',
      icon: <SoundOutlined />,
      status: 'wait'
    },
    {
      key: 'music_gen',
      title: '音乐生成',
      description: '生成背景音乐和音效',
      icon: <SoundOutlined />,
      status: 'wait'
    },
    {
      key: 'video_edit',
      title: '视频剪辑',
      description: 'AI自动剪辑合成最终视频',
      icon: <EditOutlined />,
      status: 'wait'
    },
    {
      key: 'publish',
      title: '发布宣传',
      description: '生成宣传素材，多平台发布',
      icon: <ShareAltOutlined />,
      status: 'wait'
    }
  ];

  const [steps, setSteps] = useState(workflowSteps);

  const handleSubmit = async (values: any) => {
    setLoading(true);
    try {
      const request: NovelToVideoWorkflowRequest = {
        project_id: values.projectId || '00000000-0000-0000-0000-000000000000',
        novel_text: values.novelText,
        title: values.title,
        description: values.description,
        settings: {
          video_theme: values.videoTheme,
          target_audience: values.targetAudience,
          content_type: values.contentType,
          auto_publish: values.autoPublish
        }
      };

      const result = await workflowService.createNovelToVideoWorkflow(request);
      setWorkflowId(result.workflow_id);
      message.success('工作流启动成功！');

      // 开始轮询工作流状态
      startPolling(result.workflow_id);
    } catch (error) {
      message.error('启动工作流失败，请重试');
      console.error('Workflow error:', error);
    } finally {
      setLoading(false);
    }
  };

  const startPolling = (workflowId: string) => {
    const stopPolling = workflowService.pollWorkflowStatus(
      workflowId,
      ({ workflow, tasks }) => {
        setWorkflow(workflow);
        setTasks(tasks);
        setProgress(workflow.progress);

        // 更新步骤状态
        const updatedSteps = steps.map((step, index) => {
          const task = tasks.find(t => t.step === step.key);
          if (task) {
            return {
              ...step,
              status: task.status === 'completed' ? 'finish' :
                     task.status === 'running' ? 'process' :
                     task.status === 'failed' ? 'error' : 'wait'
            };
          }
          return step;
        });
        setSteps(updatedSteps);

        // 找到当前步骤
        const currentTaskIndex = tasks.findIndex(t => t.status === 'running');
        if (currentTaskIndex !== -1) {
          setCurrentStep(currentTaskIndex);
        } else if (workflow.status === 'completed') {
          setCurrentStep(tasks.length);
          message.success('工作流执行完成！');
        } else if (workflow.status === 'failed') {
          message.error('工作流执行失败！');
        }
      }
    );

    stopPollingRef.current = stopPolling;
  };

  // 组件卸载时停止轮询
  useEffect(() => {
    return () => {
      if (stopPollingRef.current) {
        stopPollingRef.current();
      }
    };
  }, []);

  return (
    <>
      <Helmet>
        <title>AI工作流 - VidCraft Studio</title>
        <meta name="description" content="一键式AI视频制作工作流，从小说到视频的完整自动化流程" />
      </Helmet>

      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        style={{ padding: '24px', maxWidth: '1200px', margin: '0 auto' }}
      >
        <div style={{ textAlign: 'center', marginBottom: '32px' }}>
          <Title level={2}>🎬 AI视频制作工作流</Title>
          <Paragraph type="secondary" style={{ fontSize: '16px' }}>
            从小说文本到完整视频的一站式AI制作流程
          </Paragraph>
        </div>

        <div style={{ display: 'flex', gap: '24px' }}>
          {/* 左侧配置面板 */}
          <Card 
            title="工作流配置" 
            style={{ flex: '0 0 400px' }}
            extra={workflowId && <Text type="success">ID: {workflowId.slice(0, 8)}...</Text>}
          >
            <Form
              form={form}
              layout="vertical"
              onFinish={handleSubmit}
              disabled={loading || !!workflowId}
            >
              <Form.Item
                name="title"
                label="视频标题"
                rules={[{ required: true, message: '请输入视频标题' }]}
              >
                <Input placeholder="输入视频标题" />
              </Form.Item>

              <Form.Item
                name="novelText"
                label="小说文本"
                rules={[{ required: true, message: '请输入小说文本' }]}
              >
                <TextArea 
                  rows={6} 
                  placeholder="粘贴您的小说文本内容..."
                  showCount
                  maxLength={10000}
                />
              </Form.Item>

              <Form.Item name="description" label="视频描述">
                <TextArea rows={3} placeholder="描述视频内容（可选）" />
              </Form.Item>

              <Form.Item name="videoTheme" label="视频主题">
                <Select placeholder="选择视频主题">
                  <Option value="fantasy">奇幻冒险</Option>
                  <Option value="romance">浪漫爱情</Option>
                  <Option value="scifi">科幻未来</Option>
                  <Option value="mystery">悬疑推理</Option>
                  <Option value="comedy">轻松喜剧</Option>
                </Select>
              </Form.Item>

              <Form.Item name="targetAudience" label="目标受众">
                <Select placeholder="选择目标受众">
                  <Option value="young">年轻人(18-25)</Option>
                  <Option value="adult">成年人(26-40)</Option>
                  <Option value="family">家庭观众</Option>
                  <Option value="general">大众观众</Option>
                </Select>
              </Form.Item>

              <Form.Item name="contentType" label="内容类型">
                <Select placeholder="选择内容类型">
                  <Option value="short">短视频(1-3分钟)</Option>
                  <Option value="medium">中等长度(3-10分钟)</Option>
                  <Option value="long">长视频(10分钟以上)</Option>
                </Select>
              </Form.Item>

              <Form.Item name="autoPublish" label="自动发布" valuePropName="checked">
                <Switch checkedChildren="开启" unCheckedChildren="关闭" />
              </Form.Item>

              <Form.Item>
                <Button 
                  type="primary" 
                  htmlType="submit" 
                  loading={loading}
                  icon={<PlayCircleOutlined />}
                  size="large"
                  block
                >
                  {loading ? '启动中...' : '启动工作流'}
                </Button>
              </Form.Item>
            </Form>
          </Card>

          {/* 右侧进度面板 */}
          <Card title="执行进度" style={{ flex: 1 }}>
            {workflowId ? (
              <Space direction="vertical" style={{ width: '100%' }}>
                <div style={{ textAlign: 'center', marginBottom: '24px' }}>
                  <Progress 
                    type="circle" 
                    percent={Math.round(progress)} 
                    status={progress === 100 ? 'success' : 'active'}
                    size={120}
                  />
                  <div style={{ marginTop: '16px' }}>
                    <Text strong>当前步骤: {steps[currentStep]?.title || '完成'}</Text>
                  </div>
                </div>

                <Divider />

                <Steps
                  direction="vertical"
                  current={currentStep}
                  items={steps.map(step => ({
                    title: step.title,
                    description: step.description,
                    icon: step.icon,
                    status: step.status
                  }))}
                />
              </Space>
            ) : (
              <div style={{ textAlign: 'center', padding: '60px 0' }}>
                <Text type="secondary" style={{ fontSize: '16px' }}>
                  配置完成后点击"启动工作流"开始制作
                </Text>
              </div>
            )}
          </Card>
        </div>
      </motion.div>
    </>
  );
};

export default WorkflowPage;
