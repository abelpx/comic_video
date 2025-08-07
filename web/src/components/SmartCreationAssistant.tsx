import React, { useState, useEffect } from 'react';
import { 
  Card, Input, Button, Space, Typography, Row, Col, Tag, Tooltip, 
  Progress, Alert, Divider, Avatar, Badge, Statistic 
} from 'antd';
import { 
  RobotOutlined, BulbOutlined, ThunderboltOutlined, StarOutlined,
  FireOutlined, CrownOutlined, MagicOutlined, TrophyOutlined
} from '@ant-design/icons';

const { TextArea } = Input;
const { Text, Title } = Typography;

interface SmartCreationAssistantProps {
  onSubmit?: (content: string, enhancements: any) => void;
  loading?: boolean;
}

const SmartCreationAssistant: React.FC<SmartCreationAssistantProps> = ({ onSubmit, loading = false }) => {
  const [content, setContent] = useState('');
  const [analysis, setAnalysis] = useState<any>(null);
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [qualityScore, setQualityScore] = useState(0);
  const [enhancements, setEnhancements] = useState<any>({});

  // 实时内容分析
  useEffect(() => {
    if (content.length > 50) {
      analyzeContent(content);
    } else {
      setAnalysis(null);
      setSuggestions([]);
      setQualityScore(0);
    }
  }, [content]);

  const analyzeContent = (text: string) => {
    // 模拟AI分析
    const wordCount = text.length;
    const sentences = text.split(/[。！？.!?]/).filter(s => s.trim().length > 0);
    const characters = extractCharacters(text);
    const emotions = analyzeEmotions(text);
    const scenes = analyzeScenes(text);
    
    const score = Math.min(100, Math.max(0, 
      (wordCount / 10) + 
      (sentences.length * 2) + 
      (characters.length * 5) + 
      (emotions.length * 3) +
      (scenes.length * 4)
    ));

    setQualityScore(score);
    setAnalysis({
      wordCount,
      sentences: sentences.length,
      characters,
      emotions,
      scenes,
      complexity: getComplexityLevel(score)
    });

    generateSuggestions(text, score);
  };

  const extractCharacters = (text: string) => {
    const characterKeywords = ['主角', '男主', '女主', '老人', '孩子', '朋友', '敌人', '师父', '同学'];
    return characterKeywords.filter(char => text.includes(char));
  };

  const analyzeEmotions = (text: string) => {
    const emotionKeywords = {
      '快乐': ['开心', '高兴', '快乐', '兴奋', '愉快'],
      '悲伤': ['难过', '悲伤', '哭泣', '痛苦', '失落'],
      '愤怒': ['愤怒', '生气', '愤慨', '暴怒', '恼火'],
      '恐惧': ['害怕', '恐惧', '担心', '紧张', '焦虑'],
      '惊讶': ['惊讶', '震惊', '意外', '吃惊', '诧异']
    };

    const detectedEmotions = [];
    for (const [emotion, keywords] of Object.entries(emotionKeywords)) {
      if (keywords.some(keyword => text.includes(keyword))) {
        detectedEmotions.push(emotion);
      }
    }
    return detectedEmotions;
  };

  const analyzeScenes = (text: string) => {
    const sceneKeywords = ['房间', '街道', '学校', '公园', '海边', '山上', '森林', '城市', '乡村', '办公室'];
    return sceneKeywords.filter(scene => text.includes(scene));
  };

  const getComplexityLevel = (score: number) => {
    if (score >= 80) return { level: '大师级', color: '#722ed1', icon: <CrownOutlined /> };
    if (score >= 60) return { level: '专业级', color: '#1890ff', icon: <StarOutlined /> };
    if (score >= 40) return { level: '进阶级', color: '#52c41a', icon: <ThunderboltOutlined /> };
    return { level: '入门级', color: '#faad14', icon: <BulbOutlined /> };
  };

  const generateSuggestions = (text: string, score: number) => {
    const suggestions = [];
    
    if (score < 40) {
      suggestions.push('💡 建议增加更多细节描述，让故事更生动');
      suggestions.push('🎭 可以添加更多人物对话，增强戏剧性');
    }
    
    if (text.length < 200) {
      suggestions.push('📝 内容较短，建议扩展情节发展');
    }
    
    if (!text.includes('。') && !text.includes('！') && !text.includes('？')) {
      suggestions.push('✏️ 建议使用标点符号，让内容更清晰');
    }
    
    const characters = extractCharacters(text);
    if (characters.length === 0) {
      suggestions.push('👥 建议明确主要角色，让故事更有焦点');
    }
    
    const emotions = analyzeEmotions(text);
    if (emotions.length === 0) {
      suggestions.push('💝 可以添加情感元素，让故事更感人');
    }
    
    if (suggestions.length === 0) {
      suggestions.push('🎉 内容质量很好，可以直接生成视频！');
      suggestions.push('🚀 建议尝试高质量模式获得更好效果');
    }
    
    setSuggestions(suggestions);
  };

  const handleSubmit = () => {
    const enhancedContent = {
      original: content,
      analysis,
      qualityScore,
      suggestions,
      enhancements: {
        autoEnhance: true,
        qualityMode: qualityScore >= 60 ? 'professional' : 'balanced',
        stylePreset: getStylePreset(),
        optimizations: getOptimizations()
      }
    };
    
    onSubmit?.(content, enhancedContent);
  };

  const getStylePreset = () => {
    if (analysis?.emotions.includes('快乐')) return 'cheerful';
    if (analysis?.emotions.includes('悲伤')) return 'dramatic';
    if (analysis?.emotions.includes('恐惧')) return 'mysterious';
    return 'balanced';
  };

  const getOptimizations = () => {
    const opts = [];
    if (qualityScore >= 70) opts.push('high_resolution');
    if (analysis?.characters.length > 2) opts.push('character_consistency');
    if (analysis?.scenes.length > 1) opts.push('scene_transitions');
    if (analysis?.emotions.length > 0) opts.push('emotion_enhancement');
    return opts;
  };

  return (
    <div style={{ maxWidth: 1000, margin: '0 auto' }}>
      {/* 智能助手标题 */}
      <Card style={{ marginBottom: 24, background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)', border: 'none' }}>
        <Row align="middle" justify="center">
          <Col>
            <Space direction="vertical" align="center">
              <Avatar size={64} style={{ background: 'rgba(255,255,255,0.2)' }}>
                <RobotOutlined style={{ fontSize: 32, color: 'white' }} />
              </Avatar>
              <Title level={2} style={{ color: 'white', margin: 0 }}>
                🤖 AI智能创作助手
              </Title>
              <Text style={{ color: 'rgba(255,255,255,0.8)', fontSize: 16 }}>
                实时分析 · 智能优化 · 专业建议
              </Text>
            </Space>
          </Col>
        </Row>
      </Card>

      <Row gutter={24}>
        {/* 内容输入区 */}
        <Col xs={24} lg={16}>
          <Card title="📝 创作内容" style={{ height: '100%' }}>
            <TextArea
              value={content}
              onChange={(e) => setContent(e.target.value)}
              placeholder="在这里输入您的创意内容...

例如：
• 小说片段：描述一个引人入胜的故事情节
• 剧本对话：展现人物之间的精彩互动  
• 场景描述：营造生动的环境氛围
• 情感表达：传达深刻的内心感受

AI助手将实时分析您的内容，提供专业的优化建议！"
              rows={12}
              showCount
              maxLength={5000}
              style={{ fontSize: 16, lineHeight: 1.6 }}
            />
            
            <Divider />
            
            <Row justify="space-between" align="middle">
              <Col>
                <Space>
                  <Text type="secondary">字数: {content.length}</Text>
                  {analysis && (
                    <>
                      <Divider type="vertical" />
                      <Text type="secondary">句子: {analysis.sentences}</Text>
                      <Divider type="vertical" />
                      <Text type="secondary">角色: {analysis.characters.length}</Text>
                    </>
                  )}
                </Space>
              </Col>
              <Col>
                <Button 
                  type="primary" 
                  size="large"
                  icon={<MagicOutlined />}
                  onClick={handleSubmit}
                  disabled={content.length < 20}
                  loading={loading}
                  style={{ 
                    background: 'linear-gradient(45deg, #667eea, #764ba2)',
                    border: 'none',
                    borderRadius: 8
                  }}
                >
                  🚀 智能生成视频
                </Button>
              </Col>
            </Row>
          </Card>
        </Col>

        {/* 智能分析区 */}
        <Col xs={24} lg={8}>
          <Space direction="vertical" style={{ width: '100%' }} size="middle">
            {/* 质量评分 */}
            <Card size="small" title="📊 内容质量评分">
              {qualityScore > 0 ? (
                <Space direction="vertical" style={{ width: '100%' }}>
                  <div style={{ textAlign: 'center' }}>
                    <Progress
                      type="circle"
                      percent={qualityScore}
                      format={() => (
                        <div>
                          <div style={{ fontSize: 24, fontWeight: 'bold' }}>{qualityScore}</div>
                          <div style={{ fontSize: 12 }}>分</div>
                        </div>
                      )}
                      strokeColor={{
                        '0%': '#108ee9',
                        '100%': '#87d068',
                      }}
                    />
                  </div>
                  {analysis && (
                    <div style={{ textAlign: 'center' }}>
                      <Tag color={analysis.complexity.color} icon={analysis.complexity.icon}>
                        {analysis.complexity.level}
                      </Tag>
                    </div>
                  )}
                </Space>
              ) : (
                <div style={{ textAlign: 'center', padding: 20 }}>
                  <Text type="secondary">输入内容后开始分析...</Text>
                </div>
              )}
            </Card>

            {/* 内容分析 */}
            {analysis && (
              <Card size="small" title="🔍 智能分析">
                <Space direction="vertical" style={{ width: '100%' }} size="small">
                  {analysis.characters.length > 0 && (
                    <div>
                      <Text strong>检测到的角色:</Text>
                      <div style={{ marginTop: 4 }}>
                        {analysis.characters.map((char: string, index: number) => (
                          <Tag key={index} color="blue">{char}</Tag>
                        ))}
                      </div>
                    </div>
                  )}
                  
                  {analysis.emotions.length > 0 && (
                    <div>
                      <Text strong>情感色彩:</Text>
                      <div style={{ marginTop: 4 }}>
                        {analysis.emotions.map((emotion: string, index: number) => (
                          <Tag key={index} color="orange">{emotion}</Tag>
                        ))}
                      </div>
                    </div>
                  )}
                  
                  {analysis.scenes.length > 0 && (
                    <div>
                      <Text strong>场景设定:</Text>
                      <div style={{ marginTop: 4 }}>
                        {analysis.scenes.map((scene: string, index: number) => (
                          <Tag key={index} color="green">{scene}</Tag>
                        ))}
                      </div>
                    </div>
                  )}
                </Space>
              </Card>
            )}

            {/* 优化建议 */}
            {suggestions.length > 0 && (
              <Card size="small" title="💡 智能建议">
                <Space direction="vertical" style={{ width: '100%' }} size="small">
                  {suggestions.map((suggestion, index) => (
                    <Alert
                      key={index}
                      message={suggestion}
                      type="info"
                      showIcon
                      style={{ fontSize: 12 }}
                    />
                  ))}
                </Space>
              </Card>
            )}

            {/* 性能预测 */}
            {qualityScore > 0 && (
              <Card size="small" title="⚡ 生成预测">
                <Row gutter={[8, 8]}>
                  <Col span={12}>
                    <Statistic 
                      title="预计时间" 
                      value={Math.max(2, Math.ceil(qualityScore / 20))} 
                      suffix="分钟"
                      valueStyle={{ fontSize: 16 }}
                    />
                  </Col>
                  <Col span={12}>
                    <Statistic 
                      title="质量等级" 
                      value={qualityScore >= 70 ? "专业" : qualityScore >= 40 ? "标准" : "基础"} 
                      valueStyle={{ fontSize: 16 }}
                    />
                  </Col>
                </Row>
              </Card>
            )}
          </Space>
        </Col>
      </Row>
    </div>
  );
};

export default SmartCreationAssistant;
