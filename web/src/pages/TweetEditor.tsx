import React, { useState, useEffect } from 'react';
import { 
  Card, 
  Input, 
  Button, 
  Select, 
  Space, 
  message, 
  Row, 
  Col, 
  Divider,
  Tag,
  Progress,
  Tooltip,
  Modal
} from 'antd';
import { 
  SaveOutlined, 
  SendOutlined, 
  EyeOutlined, 
  CopyOutlined,
  DeleteOutlined,
  EditOutlined
} from '@ant-design/icons';
import axios from 'axios';

const { TextArea } = Input;
const { Option } = Select;

interface Tweet {
  id: string;
  content: string;
  title: string;
  platform: string;
  style: string;
  theme: string;
  hashtags: string[];
  length: number;
  quality: number;
  status: string;
  created_at: string;
}

export default function TweetEditor() {
  const [tweet, setTweet] = useState<Partial<Tweet>>({
    content: '',
    title: '',
    platform: 'twitter',
    style: '吸引人',
    theme: '',
    status: 'draft'
  });
  
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [previewVisible, setPreviewVisible] = useState(false);
  const [tweetList, setTweetList] = useState<Tweet[]>([]);
  const [editingId, setEditingId] = useState<string | null>(null);

  // 平台配置
  const platformConfig = {
    twitter: { maxLength: 280, name: 'Twitter' },
    weibo: { maxLength: 140, name: '微博' },
    douyin: { maxLength: 55, name: '抖音' }
  };

  // 风格选项
  const styleOptions = [
    '吸引人', '正式', '幽默', '营销', '情感', '专业', '轻松', '激励'
  ];

  useEffect(() => {
    loadTweetList();
  }, []);

  // 加载推文列表
  const loadTweetList = async () => {
    try {
      const response = await axios.get('/api/v1/tweets?limit=10');
      if (response.data.code === 200) {
        setTweetList(response.data.data.tweets || []);
      }
    } catch (error) {
      console.error('加载推文列表失败:', error);
    }
  };

  // 保存推文
  const saveTweet = async () => {
    if (!tweet.content?.trim()) {
      message.warning('请输入推文内容');
      return;
    }

    setSaving(true);
    try {
      const payload = {
        content: tweet.content,
        title: tweet.title || '',
        platform: tweet.platform,
        style: tweet.style,
        theme: tweet.theme || '',
        source_type: 'manual'
      };

      let response;
      if (editingId) {
        // 更新现有推文
        response = await axios.put(`/api/v1/tweets/${editingId}`, payload);
      } else {
        // 创建新推文
        response = await axios.post('/api/v1/tweets', payload);
      }

      if (response.data.code === 200) {
        message.success(editingId ? '推文更新成功' : '推文保存成功');
        setEditingId(null);
        resetForm();
        loadTweetList();
      }
    } catch (error) {
      message.error('保存失败');
      console.error('保存推文失败:', error);
    } finally {
      setSaving(false);
    }
  };

  // 发布推文
  const publishTweet = async (id: string) => {
    try {
      const response = await axios.post(`/api/v1/tweets/${id}/publish`);
      if (response.data.code === 200) {
        message.success('推文发布成功');
        loadTweetList();
      }
    } catch (error) {
      message.error('发布失败');
      console.error('发布推文失败:', error);
    }
  };

  // 删除推文
  const deleteTweet = async (id: string) => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除这条推文吗？',
      onOk: async () => {
        try {
          const response = await axios.delete(`/api/v1/tweets/${id}`);
          if (response.data.code === 200) {
            message.success('推文删除成功');
            loadTweetList();
          }
        } catch (error) {
          message.error('删除失败');
          console.error('删除推文失败:', error);
        }
      }
    });
  };

  // 编辑推文
  const editTweet = (tweetData: Tweet) => {
    setTweet({
      content: tweetData.content,
      title: tweetData.title,
      platform: tweetData.platform,
      style: tweetData.style,
      theme: tweetData.theme
    });
    setEditingId(tweetData.id);
  };

  // 复制推文内容
  const copyTweet = (content: string) => {
    navigator.clipboard.writeText(content);
    message.success('内容已复制到剪贴板');
  };

  // 重置表单
  const resetForm = () => {
    setTweet({
      content: '',
      title: '',
      platform: 'twitter',
      style: '吸引人',
      theme: '',
      status: 'draft'
    });
  };

  // 计算推文质量
  const calculateQuality = (content: string) => {
    let score = 0;
    if (content.length >= 50 && content.length <= 200) score += 30;
    if (content.includes('#')) score += 20;
    if (content.includes('？') || content.includes('！')) score += 20;
    if (/[\u{1F600}-\u{1F64F}]|[\u{1F300}-\u{1F5FF}]/u.test(content)) score += 15;
    if (content.length > 20) score += 15;
    return Math.min(score, 100);
  };

  const currentPlatform = platformConfig[tweet.platform as keyof typeof platformConfig];
  const contentLength = tweet.content?.length || 0;
  const quality = calculateQuality(tweet.content || '');
  const isOverLimit = contentLength > currentPlatform.maxLength;

  return (
    <div style={{ padding: '24px', maxWidth: '1200px', margin: '0 auto' }}>
      <Row gutter={24}>
        {/* 编辑器区域 */}
        <Col span={14}>
          <Card 
            title={editingId ? "编辑推文" : "创建推文"} 
            extra={
              <Space>
                <Button onClick={resetForm}>重置</Button>
                <Button 
                  type="primary" 
                  icon={<SaveOutlined />}
                  loading={saving}
                  onClick={saveTweet}
                >
                  {editingId ? '更新' : '保存'}
                </Button>
              </Space>
            }
          >
            <Space direction="vertical" style={{ width: '100%' }} size="large">
              {/* 基本信息 */}
              <Row gutter={16}>
                <Col span={12}>
                  <label>推文标题（可选）</label>
                  <Input
                    placeholder="为推文添加一个标题"
                    value={tweet.title}
                    onChange={(e) => setTweet({ ...tweet, title: e.target.value })}
                  />
                </Col>
                <Col span={12}>
                  <label>主题/角度</label>
                  <Input
                    placeholder="推文主题或角度"
                    value={tweet.theme}
                    onChange={(e) => setTweet({ ...tweet, theme: e.target.value })}
                  />
                </Col>
              </Row>

              {/* 平台和风格 */}
              <Row gutter={16}>
                <Col span={12}>
                  <label>目标平台</label>
                  <Select
                    style={{ width: '100%' }}
                    value={tweet.platform}
                    onChange={(value) => setTweet({ ...tweet, platform: value })}
                  >
                    {Object.entries(platformConfig).map(([key, config]) => (
                      <Option key={key} value={key}>
                        {config.name} (最多{config.maxLength}字符)
                      </Option>
                    ))}
                  </Select>
                </Col>
                <Col span={12}>
                  <label>推文风格</label>
                  <Select
                    style={{ width: '100%' }}
                    value={tweet.style}
                    onChange={(value) => setTweet({ ...tweet, style: value })}
                  >
                    {styleOptions.map(style => (
                      <Option key={style} value={style}>{style}</Option>
                    ))}
                  </Select>
                </Col>
              </Row>

              {/* 推文内容 */}
              <div>
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: 8 }}>
                  <label>推文内容</label>
                  <Space>
                    <span style={{ color: isOverLimit ? '#ff4d4f' : '#666' }}>
                      {contentLength}/{currentPlatform.maxLength}
                    </span>
                    <Tooltip title="推文质量评分">
                      <Progress 
                        type="circle" 
                        size={24} 
                        percent={quality}
                        format={() => quality}
                        strokeColor={quality >= 70 ? '#52c41a' : quality >= 40 ? '#faad14' : '#ff4d4f'}
                      />
                    </Tooltip>
                  </Space>
                </div>
                <TextArea
                  rows={6}
                  placeholder="输入推文内容..."
                  value={tweet.content}
                  onChange={(e) => setTweet({ ...tweet, content: e.target.value })}
                  status={isOverLimit ? 'error' : ''}
                />
                {isOverLimit && (
                  <div style={{ color: '#ff4d4f', fontSize: '12px', marginTop: 4 }}>
                    内容超出{currentPlatform.name}字符限制
                  </div>
                )}
              </div>

              {/* 预览按钮 */}
              <Button 
                icon={<EyeOutlined />}
                onClick={() => setPreviewVisible(true)}
                disabled={!tweet.content}
              >
                预览效果
              </Button>
            </Space>
          </Card>
        </Col>

        {/* 推文列表区域 */}
        <Col span={10}>
          <Card title="我的推文" extra={<Button onClick={loadTweetList}>刷新</Button>}>
            <div style={{ maxHeight: '600px', overflowY: 'auto' }}>
              {tweetList.map((item) => (
                <Card 
                  key={item.id}
                  size="small" 
                  style={{ marginBottom: 12 }}
                  actions={[
                    <Tooltip title="编辑">
                      <EditOutlined onClick={() => editTweet(item)} />
                    </Tooltip>,
                    <Tooltip title="复制">
                      <CopyOutlined onClick={() => copyTweet(item.content)} />
                    </Tooltip>,
                    item.status === 'draft' && (
                      <Tooltip title="发布">
                        <SendOutlined onClick={() => publishTweet(item.id)} />
                      </Tooltip>
                    ),
                    <Tooltip title="删除">
                      <DeleteOutlined onClick={() => deleteTweet(item.id)} />
                    </Tooltip>
                  ].filter(Boolean)}
                >
                  <div style={{ marginBottom: 8 }}>
                    <Space>
                      <Tag color={item.status === 'published' ? 'green' : 'blue'}>
                        {item.status === 'published' ? '已发布' : '草稿'}
                      </Tag>
                      <Tag>{platformConfig[item.platform as keyof typeof platformConfig]?.name}</Tag>
                      <span style={{ fontSize: '12px', color: '#999' }}>
                        {item.length}字符
                      </span>
                    </Space>
                  </div>
                  <div style={{ 
                    fontSize: '14px', 
                    lineHeight: '1.4',
                    maxHeight: '60px',
                    overflow: 'hidden',
                    textOverflow: 'ellipsis'
                  }}>
                    {item.content}
                  </div>
                  {item.title && (
                    <div style={{ fontSize: '12px', color: '#666', marginTop: 4 }}>
                      标题: {item.title}
                    </div>
                  )}
                </Card>
              ))}
            </div>
          </Card>
        </Col>
      </Row>

      {/* 预览模态框 */}
      <Modal
        title="推文预览"
        open={previewVisible}
        onCancel={() => setPreviewVisible(false)}
        footer={[
          <Button key="close" onClick={() => setPreviewVisible(false)}>
            关闭
          </Button>,
          <Button key="copy" onClick={() => copyTweet(tweet.content || '')}>
            复制内容
          </Button>
        ]}
      >
        <div style={{ 
          border: '1px solid #d9d9d9', 
          borderRadius: '8px', 
          padding: '16px',
          backgroundColor: '#fafafa'
        }}>
          <div style={{ marginBottom: 8 }}>
            <Space>
              <Tag>{currentPlatform.name}</Tag>
              <Tag color="blue">{tweet.style}</Tag>
              {tweet.theme && <Tag color="green">{tweet.theme}</Tag>}
            </Space>
          </div>
          <div style={{ 
            fontSize: '16px', 
            lineHeight: '1.5',
            whiteSpace: 'pre-wrap'
          }}>
            {tweet.content}
          </div>
          <Divider />
          <div style={{ fontSize: '12px', color: '#666' }}>
            <Space>
              <span>字符数: {contentLength}/{currentPlatform.maxLength}</span>
              <span>质量评分: {quality}/100</span>
            </Space>
          </div>
        </div>
      </Modal>
    </div>
  );
}
