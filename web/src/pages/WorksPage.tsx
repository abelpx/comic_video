import React, { useEffect, useState } from 'react';
import { Card, Row, Col, Typography, Button, Table, Tag, Space, Modal, Image, Empty, Input, Select, DatePicker } from 'antd';
import { motion } from 'framer-motion';
import {
  EyeOutlined,
  DownloadOutlined,
  ShareAltOutlined,
  DeleteOutlined,
  SearchOutlined,
  FilterOutlined,
  PlayCircleOutlined,
  PictureOutlined,
  VideoCameraOutlined,
  BookOutlined,
  ReloadOutlined,
} from '@ant-design/icons';
import { useTasks } from '../store';
import { taskApi, projectApi } from '../services/api';
import useTaskManager from '../hooks/useTaskManager';
import TaskDebugger from '../components/TaskDebugger';
import dayjs from 'dayjs';

const { Title, Text } = Typography;
const { Search } = Input;
const { Option } = Select;
const { RangePicker } = DatePicker;

interface Work {
  id: string;
  title: string;
  type: 'video' | 'comic' | 'novel' | 'image';
  status: 'completed' | 'processing' | 'failed';
  createdAt: string;
  thumbnail?: string;
  result?: any;
  progress?: number;
}

const WorksPage: React.FC = () => {
  const { tasks } = useTasks();
  const [works, setWorks] = useState<Work[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedWork, setSelectedWork] = useState<Work | null>(null);
  const [previewVisible, setPreviewVisible] = useState(false);
  const [filters, setFilters] = useState({
    search: '',
    type: 'all',
    status: 'all',
    dateRange: null as any,
  });

  // 使用任务管理器
  const { refreshTasks } = useTaskManager({
    autoLoad: false, // 在WorksPage中手动控制加载
    enablePolling: true,
  });

  // 作品类型配置
  const workTypes = {
    video: { name: '视频', icon: <VideoCameraOutlined />, color: '#0ea5e9' },
    comic: { name: '漫画', icon: <PictureOutlined />, color: '#a855f7' },
    novel: { name: '小说', icon: <BookOutlined />, color: '#06b6d4' },
    image: { name: '图片', icon: <PictureOutlined />, color: '#10b981' },
  };

  useEffect(() => {
    // 使用任务管理器加载任务，而不是直接调用loadWorks
    loadWorks();
  }, []);

  // 当组件挂载时，确保任务管理器已经加载了最新数据
  useEffect(() => {
    if (tasks.length === 0) {
      // 如果没有任务数据，触发一次刷新
      refreshTasks();
    }
  }, [tasks.length, refreshTasks]);

  const loadWorks = async () => {
    try {
      setLoading(true);
      // 使用任务管理器刷新任务
      await refreshTasks();

      // 从store中获取任务并转换为作品数据
      const worksData = tasks.map((task: any) => ({
        id: task.id,
        title: task.title || `${getWorkTypeName(task.type)}作品`,
        type: mapTaskTypeToWorkType(task.type),
        status: task.status,
        createdAt: task.created_at || task.createdAt,
        result: task.result,
        progress: task.progress,
      }));

      setWorks(worksData);
    } catch (error) {
      console.error('加载作品失败:', error);
    } finally {
      setLoading(false);
    }
  };

  // 当tasks变化时更新works
  useEffect(() => {
    const worksData = tasks.map((task: any) => ({
      id: task.id,
      title: task.title || `${getWorkTypeName(task.type)}作品`,
      type: mapTaskTypeToWorkType(task.type),
      status: task.status,
      createdAt: task.created_at || task.createdAt,
      result: task.result,
      progress: task.progress,
    }));
    setWorks(worksData);
  }, [tasks]);

  const getWorkTypeName = (type: string) => {
    const typeMap: Record<string, string> = {
      video: '视频',
      comic: '漫画',
      novel: '小说',
      image: '图片',
    };
    return typeMap[type] || '作品';
  };

  const mapTaskTypeToWorkType = (taskType: string): 'video' | 'comic' | 'novel' | 'image' => {
    if (taskType.includes('video')) return 'video';
    if (taskType.includes('comic')) return 'comic';
    if (taskType.includes('novel')) return 'novel';
    return 'image';
  };

  const getStatusTag = (status: string) => {
    const statusMap = {
      completed: { color: 'success', text: '已完成' },
      processing: { color: 'processing', text: '处理中' },
      failed: { color: 'error', text: '失败' },
      pending: { color: 'default', text: '等待中' },
    };
    const config = statusMap[status as keyof typeof statusMap] || statusMap.pending;
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  const handlePreview = (work: Work) => {
    setSelectedWork(work);
    setPreviewVisible(true);
  };

  const handleDownload = (work: Work) => {
    // TODO: 实现下载功能
    console.log('下载作品:', work.id);
  };

  const handleShare = (work: Work) => {
    // TODO: 实现分享功能
    console.log('分享作品:', work.id);
  };

  const handleDelete = (work: Work) => {
    Modal.confirm({
      title: '确认删除',
      content: `确定要删除作品"${work.title}"吗？此操作不可恢复。`,
      okText: '删除',
      okType: 'danger',
      cancelText: '取消',
      onOk: async () => {
        try {
          // TODO: 实现删除功能
          console.log('删除作品:', work.id);
          await refreshTasks(); // 使用任务管理器刷新
        } catch (error) {
          console.error('删除失败:', error);
        }
      },
    });
  };

  // 过滤作品
  const filteredWorks = works.filter(work => {
    if (filters.search && !work.title.toLowerCase().includes(filters.search.toLowerCase())) {
      return false;
    }
    if (filters.type !== 'all' && work.type !== filters.type) {
      return false;
    }
    if (filters.status !== 'all' && work.status !== filters.status) {
      return false;
    }
    if (filters.dateRange && filters.dateRange.length === 2) {
      const workDate = dayjs(work.createdAt);
      if (!workDate.isBetween(filters.dateRange[0], filters.dateRange[1], 'day', '[]')) {
        return false;
      }
    }
    return true;
  });

  const columns = [
    {
      title: '作品',
      dataIndex: 'title',
      key: 'title',
      render: (title: string, record: Work) => (
        <div style={{ display: 'flex', alignItems: 'center', gap: 12 }}>
          <div style={{
            width: 40,
            height: 40,
            borderRadius: 8,
            background: `${workTypes[record.type].color}10`,
            display: 'flex',
            alignItems: 'center',
            justifyContent: 'center',
            color: workTypes[record.type].color,
          }}>
            {workTypes[record.type].icon}
          </div>
          <div>
            <div style={{ fontWeight: 500 }}>{title}</div>
            <Text type="secondary" style={{ fontSize: 12 }}>
              {workTypes[record.type].name}
            </Text>
          </div>
        </div>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => getStatusTag(status),
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
      render: (date: string) => dayjs(date).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: '操作',
      key: 'actions',
      render: (_: any, record: Work) => (
        <Space>
          <Button
            type="text"
            icon={<EyeOutlined />}
            onClick={() => handlePreview(record)}
            disabled={record.status !== 'completed'}
          >
            预览
          </Button>
          <Button
            type="text"
            icon={<DownloadOutlined />}
            onClick={() => handleDownload(record)}
            disabled={record.status !== 'completed'}
          >
            下载
          </Button>
          <Button
            type="text"
            icon={<ShareAltOutlined />}
            onClick={() => handleShare(record)}
            disabled={record.status !== 'completed'}
          >
            分享
          </Button>
          <Button
            type="text"
            danger
            icon={<DeleteOutlined />}
            onClick={() => handleDelete(record)}
          >
            删除
          </Button>
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      {/* 调试组件 */}
      <TaskDebugger />

      {/* 页面标题 */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5 }}
        style={{ marginBottom: 32 }}
      >
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
          <div>
            <Title level={2} style={{ marginBottom: 8 }}>
              我的作品
            </Title>
            <Text type="secondary" style={{ fontSize: 16 }}>
              管理您创作的所有作品，包括视频、漫画、小说等内容。
            </Text>
          </div>
          <Button
            type="primary"
            icon={<ReloadOutlined />}
            onClick={loadWorks}
            loading={loading}
          >
            刷新
          </Button>
        </div>
      </motion.div>

      {/* 筛选器 */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, delay: 0.1 }}
      >
        <Card style={{ marginBottom: 24 }}>
          <Row gutter={[16, 16]} align="middle">
            <Col xs={24} sm={8} md={6}>
              <Search
                placeholder="搜索作品标题"
                value={filters.search}
                onChange={(e) => setFilters({ ...filters, search: e.target.value })}
                style={{ width: '100%' }}
              />
            </Col>
            <Col xs={12} sm={4} md={3}>
              <Select
                value={filters.type}
                onChange={(value) => setFilters({ ...filters, type: value })}
                style={{ width: '100%' }}
              >
                <Option value="all">全部类型</Option>
                <Option value="video">视频</Option>
                <Option value="comic">漫画</Option>
                <Option value="novel">小说</Option>
                <Option value="image">图片</Option>
              </Select>
            </Col>
            <Col xs={12} sm={4} md={3}>
              <Select
                value={filters.status}
                onChange={(value) => setFilters({ ...filters, status: value })}
                style={{ width: '100%' }}
              >
                <Option value="all">全部状态</Option>
                <Option value="completed">已完成</Option>
                <Option value="processing">处理中</Option>
                <Option value="failed">失败</Option>
              </Select>
            </Col>
            <Col xs={24} sm={8} md={6}>
              <RangePicker
                value={filters.dateRange}
                onChange={(dates) => setFilters({ ...filters, dateRange: dates })}
                style={{ width: '100%' }}
              />
            </Col>
            <Col xs={24} sm={24} md={6}>
              <Button
                icon={<FilterOutlined />}
                onClick={() => setFilters({ search: '', type: 'all', status: 'all', dateRange: null })}
              >
                清除筛选
              </Button>
            </Col>
          </Row>
        </Card>
      </motion.div>

      {/* 作品列表 */}
      <motion.div
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.5, delay: 0.2 }}
      >
        <Card>
          {filteredWorks.length > 0 ? (
            <Table
              columns={columns}
              dataSource={filteredWorks}
              rowKey="id"
              loading={loading}
              pagination={{
                pageSize: 10,
                showSizeChanger: true,
                showQuickJumper: true,
                showTotal: (total) => `共 ${total} 个作品`,
              }}
            />
          ) : (
            <Empty
              description="暂无作品"
              style={{ margin: '60px 0' }}
            >
              <Button type="primary" href="/app/novel-to-video">
                开始创作
              </Button>
            </Empty>
          )}
        </Card>
      </motion.div>

      {/* 预览模态框 */}
      <Modal
        title={selectedWork?.title}
        open={previewVisible}
        onCancel={() => setPreviewVisible(false)}
        footer={null}
        width={800}
      >
        {selectedWork && (
          <div style={{ textAlign: 'center' }}>
            {selectedWork.type === 'video' ? (
              <div>
                <PlayCircleOutlined style={{ fontSize: 64, color: '#0ea5e9', marginBottom: 16 }} />
                <div>视频预览功能开发中...</div>
              </div>
            ) : selectedWork.type === 'image' || selectedWork.type === 'comic' ? (
              <div>
                <PictureOutlined style={{ fontSize: 64, color: '#a855f7', marginBottom: 16 }} />
                <div>图片预览功能开发中...</div>
              </div>
            ) : (
              <div>
                <BookOutlined style={{ fontSize: 64, color: '#06b6d4', marginBottom: 16 }} />
                <div>文本预览功能开发中...</div>
              </div>
            )}
          </div>
        )}
      </Modal>
    </div>
  );
};

export default WorksPage;
