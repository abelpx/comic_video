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
  Modal,
  Form,
  Tag,
  Tooltip,
  Divider,
  List,
  Avatar
} from 'antd';
import { 
  PlusOutlined, 
  EditOutlined, 
  DeleteOutlined, 
  CopyOutlined,
  ThunderboltOutlined,
  StarOutlined,
  EyeOutlined
} from '@ant-design/icons';
import axios from 'axios';

const { TextArea } = Input;
const { Option } = Select;

interface TweetTemplate {
  id: string;
  name: string;
  description: string;
  category: string;
  template: string;
  variables: string[];
  platform: string;
  style: string;
  max_length: number;
  use_count: number;
  rating: number;
  is_public: boolean;
  is_premium: boolean;
  created_at: string;
}

export default function TweetTemplates() {
  const [templates, setTemplates] = useState<TweetTemplate[]>([]);
  const [categories, setCategories] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [modalVisible, setModalVisible] = useState(false);
  const [useModalVisible, setUseModalVisible] = useState(false);
  const [editingTemplate, setEditingTemplate] = useState<TweetTemplate | null>(null);
  const [selectedTemplate, setSelectedTemplate] = useState<TweetTemplate | null>(null);
  const [form] = Form.useForm();
  const [useForm] = Form.useForm();
  
  // 筛选条件
  const [filters, setFilters] = useState({
    category: '',
    platform: '',
    keyword: ''
  });

  // 平台配置
  const platformConfig = {
    twitter: { maxLength: 280, name: 'Twitter' },
    weibo: { maxLength: 140, name: '微博' },
    douyin: { maxLength: 55, name: '抖音' }
  };

  // 分类选项
  const categoryOptions = [
    '小说推广', '产品营销', '个人品牌', '新闻资讯', '娱乐内容', 
    '教育培训', '生活分享', '技术分享', '其他'
  ];

  useEffect(() => {
    loadTemplates();
    loadCategories();
  }, [filters]);

  // 加载模板列表
  const loadTemplates = async () => {
    setLoading(true);
    try {
      const params = new URLSearchParams();
      if (filters.category) params.append('category', filters.category);
      if (filters.platform) params.append('platform', filters.platform);
      if (filters.keyword) params.append('keyword', filters.keyword);
      
      const url = filters.keyword 
        ? `/api/v1/tweet-templates/search?${params.toString()}`
        : `/api/v1/tweet-templates?${params.toString()}`;
        
      const response = await axios.get(url);
      if (response.data.code === 200) {
        setTemplates(response.data.data.templates || []);
      }
    } catch (error) {
      message.error('加载模板失败');
      console.error('加载模板失败:', error);
    } finally {
      setLoading(false);
    }
  };

  // 加载分类列表
  const loadCategories = async () => {
    try {
      const response = await axios.get('/api/v1/tweet-templates/categories');
      if (response.data.code === 200) {
        setCategories(response.data.data || []);
      }
    } catch (error) {
      console.error('加载分类失败:', error);
    }
  };

  // 创建/更新模板
  const saveTemplate = async (values: any) => {
    try {
      const payload = {
        ...values,
        is_public: values.is_public || false,
        is_premium: values.is_premium || false
      };

      let response;
      if (editingTemplate) {
        response = await axios.put(`/api/v1/tweet-templates/${editingTemplate.id}`, payload);
      } else {
        response = await axios.post('/api/v1/tweet-templates', payload);
      }

      if (response.data.code === 200) {
        message.success(editingTemplate ? '模板更新成功' : '模板创建成功');
        setModalVisible(false);
        setEditingTemplate(null);
        form.resetFields();
        loadTemplates();
      }
    } catch (error) {
      message.error('保存模板失败');
      console.error('保存模板失败:', error);
    }
  };

  // 删除模板
  const deleteTemplate = async (id: string) => {
    Modal.confirm({
      title: '确认删除',
      content: '确定要删除这个模板吗？',
      onOk: async () => {
        try {
          const response = await axios.delete(`/api/v1/tweet-templates/${id}`);
          if (response.data.code === 200) {
            message.success('模板删除成功');
            loadTemplates();
          }
        } catch (error) {
          message.error('删除模板失败');
          console.error('删除模板失败:', error);
        }
      }
    });
  };

  // 使用模板
  const useTemplate = async (values: any) => {
    try {
      const response = await axios.post(`/api/v1/tweet-templates/${selectedTemplate?.id}/use`, {
        variables: values
      });
      
      if (response.data.code === 200) {
        const content = response.data.data.content;
        
        // 复制到剪贴板
        navigator.clipboard.writeText(content);
        message.success('推文内容已生成并复制到剪贴板');
        
        setUseModalVisible(false);
        useForm.resetFields();
        
        // 可以选择跳转到推文编辑器
        Modal.confirm({
          title: '跳转到编辑器',
          content: '是否跳转到推文编辑器进行进一步编辑？',
          onOk: () => {
            // 这里可以跳转到推文编辑器并传递内容
            window.location.href = `/app/tweet-editor?content=${encodeURIComponent(content)}`;
          }
        });
      }
    } catch (error) {
      message.error('使用模板失败');
      console.error('使用模板失败:', error);
    }
  };

  // 编辑模板
  const editTemplate = (template: TweetTemplate) => {
    setEditingTemplate(template);
    form.setFieldsValue({
      name: template.name,
      description: template.description,
      category: template.category,
      template: template.template,
      platform: template.platform,
      style: template.style,
      max_length: template.max_length,
      is_public: template.is_public,
      is_premium: template.is_premium
    });
    setModalVisible(true);
  };

  // 打开使用模板对话框
  const openUseModal = (template: TweetTemplate) => {
    setSelectedTemplate(template);
    setUseModalVisible(true);
    
    // 设置表单初始值
    const initialValues: any = {};
    template.variables.forEach(variable => {
      initialValues[variable] = '';
    });
    useForm.setFieldsValue(initialValues);
  };

  return (
    <div style={{ padding: '24px' }}>
      {/* 页面标题和操作 */}
      <div style={{ marginBottom: 24, display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <h2>推文模板</h2>
        <Button 
          type="primary" 
          icon={<PlusOutlined />}
          onClick={() => {
            setEditingTemplate(null);
            form.resetFields();
            setModalVisible(true);
          }}
        >
          创建模板
        </Button>
      </div>

      {/* 筛选条件 */}
      <Card style={{ marginBottom: 24 }}>
        <Row gutter={16}>
          <Col span={6}>
            <Input.Search
              placeholder="搜索模板..."
              value={filters.keyword}
              onChange={(e) => setFilters({ ...filters, keyword: e.target.value })}
              onSearch={loadTemplates}
            />
          </Col>
          <Col span={6}>
            <Select
              style={{ width: '100%' }}
              placeholder="选择分类"
              value={filters.category}
              onChange={(value) => setFilters({ ...filters, category: value })}
              allowClear
            >
              {categoryOptions.map(category => (
                <Option key={category} value={category}>{category}</Option>
              ))}
            </Select>
          </Col>
          <Col span={6}>
            <Select
              style={{ width: '100%' }}
              placeholder="选择平台"
              value={filters.platform}
              onChange={(value) => setFilters({ ...filters, platform: value })}
              allowClear
            >
              {Object.entries(platformConfig).map(([key, config]) => (
                <Option key={key} value={key}>{config.name}</Option>
              ))}
            </Select>
          </Col>
          <Col span={6}>
            <Button onClick={loadTemplates}>重置筛选</Button>
          </Col>
        </Row>
      </Card>

      {/* 模板列表 */}
      <List
        loading={loading}
        grid={{ gutter: 16, xs: 1, sm: 2, md: 2, lg: 3, xl: 3, xxl: 4 }}
        dataSource={templates}
        renderItem={(template) => (
          <List.Item>
            <Card
              size="small"
              title={
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <span style={{ fontSize: '14px' }}>{template.name}</span>
                  <Space>
                    {template.is_premium && <StarOutlined style={{ color: '#faad14' }} />}
                    <span style={{ fontSize: '12px', color: '#999' }}>
                      {template.use_count}次使用
                    </span>
                  </Space>
                </div>
              }
              actions={[
                <Tooltip title="使用模板">
                  <ThunderboltOutlined onClick={() => openUseModal(template)} />
                </Tooltip>,
                <Tooltip title="编辑">
                  <EditOutlined onClick={() => editTemplate(template)} />
                </Tooltip>,
                <Tooltip title="删除">
                  <DeleteOutlined onClick={() => deleteTemplate(template.id)} />
                </Tooltip>
              ]}
            >
              <div style={{ marginBottom: 8 }}>
                <Space wrap>
                  <Tag color="blue">{template.category}</Tag>
                  <Tag>{platformConfig[template.platform as keyof typeof platformConfig]?.name}</Tag>
                  <Tag color="green">{template.style}</Tag>
                </Space>
              </div>
              
              <div style={{ 
                fontSize: '12px', 
                color: '#666', 
                marginBottom: 8,
                height: '32px',
                overflow: 'hidden',
                textOverflow: 'ellipsis'
              }}>
                {template.description}
              </div>
              
              <div style={{ 
                fontSize: '13px',
                backgroundColor: '#f5f5f5',
                padding: '8px',
                borderRadius: '4px',
                maxHeight: '60px',
                overflow: 'hidden',
                textOverflow: 'ellipsis'
              }}>
                {template.template}
              </div>
              
              {template.variables.length > 0 && (
                <div style={{ marginTop: 8 }}>
                  <div style={{ fontSize: '12px', color: '#999', marginBottom: 4 }}>
                    变量: {template.variables.join(', ')}
                  </div>
                </div>
              )}
            </Card>
          </List.Item>
        )}
      />

      {/* 创建/编辑模板对话框 */}
      <Modal
        title={editingTemplate ? "编辑模板" : "创建模板"}
        open={modalVisible}
        onCancel={() => {
          setModalVisible(false);
          setEditingTemplate(null);
          form.resetFields();
        }}
        footer={null}
        width={800}
      >
        <Form
          form={form}
          layout="vertical"
          onFinish={saveTemplate}
        >
          <Row gutter={16}>
            <Col span={12}>
              <Form.Item
                name="name"
                label="模板名称"
                rules={[{ required: true, message: '请输入模板名称' }]}
              >
                <Input placeholder="输入模板名称" />
              </Form.Item>
            </Col>
            <Col span={12}>
              <Form.Item
                name="category"
                label="分类"
                rules={[{ required: true, message: '请选择分类' }]}
              >
                <Select placeholder="选择分类">
                  {categoryOptions.map(category => (
                    <Option key={category} value={category}>{category}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
          </Row>

          <Form.Item
            name="description"
            label="模板描述"
          >
            <Input.TextArea rows={2} placeholder="描述模板的用途和特点" />
          </Form.Item>

          <Form.Item
            name="template"
            label="模板内容"
            rules={[{ required: true, message: '请输入模板内容' }]}
            extra="使用 {{变量名}} 的格式定义变量，如：{{书名}}、{{作者}}等"
          >
            <TextArea rows={4} placeholder="输入模板内容，使用{{变量名}}定义变量" />
          </Form.Item>

          <Row gutter={16}>
            <Col span={8}>
              <Form.Item
                name="platform"
                label="目标平台"
                rules={[{ required: true, message: '请选择平台' }]}
              >
                <Select placeholder="选择平台">
                  {Object.entries(platformConfig).map(([key, config]) => (
                    <Option key={key} value={key}>{config.name}</Option>
                  ))}
                </Select>
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                name="style"
                label="推文风格"
              >
                <Input placeholder="如：吸引人、正式、幽默等" />
              </Form.Item>
            </Col>
            <Col span={8}>
              <Form.Item
                name="max_length"
                label="最大长度"
              >
                <Input type="number" placeholder="字符数限制" />
              </Form.Item>
            </Col>
          </Row>

          <div style={{ textAlign: 'right' }}>
            <Space>
              <Button onClick={() => setModalVisible(false)}>取消</Button>
              <Button type="primary" htmlType="submit">
                {editingTemplate ? '更新' : '创建'}
              </Button>
            </Space>
          </div>
        </Form>
      </Modal>

      {/* 使用模板对话框 */}
      <Modal
        title={`使用模板: ${selectedTemplate?.name}`}
        open={useModalVisible}
        onCancel={() => {
          setUseModalVisible(false);
          useForm.resetFields();
        }}
        footer={null}
      >
        {selectedTemplate && (
          <>
            <div style={{ marginBottom: 16, padding: 12, backgroundColor: '#f5f5f5', borderRadius: 4 }}>
              <div style={{ fontSize: '13px', marginBottom: 8 }}>模板内容:</div>
              <div style={{ fontSize: '14px' }}>{selectedTemplate.template}</div>
            </div>
            
            <Form
              form={useForm}
              layout="vertical"
              onFinish={useTemplate}
            >
              {selectedTemplate.variables.map(variable => (
                <Form.Item
                  key={variable}
                  name={variable}
                  label={`${variable}`}
                  rules={[{ required: true, message: `请输入${variable}` }]}
                >
                  <Input placeholder={`输入${variable}`} />
                </Form.Item>
              ))}
              
              <div style={{ textAlign: 'right' }}>
                <Space>
                  <Button onClick={() => setUseModalVisible(false)}>取消</Button>
                  <Button type="primary" htmlType="submit">
                    生成推文
                  </Button>
                </Space>
              </div>
            </Form>
          </>
        )}
      </Modal>
    </div>
  );
}
