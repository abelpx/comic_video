package character

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// VisualConsistencyValidator 视觉一致性验证器
type VisualConsistencyValidator struct {
	featureExtractor *FeatureExtractor
	similarityEngine *SimilarityEngine
	qualityAnalyzer  *ImageQualityAnalyzer
}

// NewVisualConsistencyValidator 创建视觉一致性验证器
func NewVisualConsistencyValidator() *VisualConsistencyValidator {
	return &VisualConsistencyValidator{
		featureExtractor: NewFeatureExtractor(),
		similarityEngine: NewSimilarityEngine(),
		qualityAnalyzer:  NewImageQualityAnalyzer(),
	}
}

// ValidateVisualConsistency 验证视觉一致性
func (vcv *VisualConsistencyValidator) ValidateVisualConsistency(targetImagePath string, referenceImages []string) (float64, error) {
	log.Printf("[VisualValidator] 验证视觉一致性: %s", targetImagePath)

	// 检查目标图像是否存在
	if _, err := os.Stat(targetImagePath); os.IsNotExist(err) {
		return 0, fmt.Errorf("目标图像不存在: %s", targetImagePath)
	}

	// 如果没有参考图像，返回基础分数
	if len(referenceImages) == 0 {
		log.Printf("[VisualValidator] 没有参考图像，返回基础分数")
		return 0.6, nil
	}

	// 1. 提取目标图像特征
	targetFeatures, err := vcv.featureExtractor.ExtractFeatures(targetImagePath)
	if err != nil {
		return 0, fmt.Errorf("提取目标图像特征失败: %w", err)
	}

	// 2. 计算与参考图像的相似度
	var totalSimilarity float64
	validReferenceCount := 0

	for _, refImagePath := range referenceImages {
		// 检查参考图像是否存在
		if _, err := os.Stat(refImagePath); os.IsNotExist(err) {
			log.Printf("[VisualValidator] 参考图像不存在，跳过: %s", refImagePath)
			continue
		}

		// 提取参考图像特征
		refFeatures, err := vcv.featureExtractor.ExtractFeatures(refImagePath)
		if err != nil {
			log.Printf("[VisualValidator] 提取参考图像特征失败，跳过: %s, 错误: %v", refImagePath, err)
			continue
		}

		// 计算相似度
		similarity := vcv.similarityEngine.CalculateSimilarity(targetFeatures, refFeatures)
		totalSimilarity += similarity
		validReferenceCount++

		log.Printf("[VisualValidator] 与参考图像 %s 的相似度: %.3f", filepath.Base(refImagePath), similarity)
	}

	// 3. 计算平均相似度
	var averageSimilarity float64
	if validReferenceCount > 0 {
		averageSimilarity = totalSimilarity / float64(validReferenceCount)
	} else {
		// 如果没有有效的参考图像，返回基础分数
		averageSimilarity = 0.6
	}

	// 4. 图像质量评估
	qualityScore, err := vcv.qualityAnalyzer.AnalyzeImageQuality(targetImagePath)
	if err != nil {
		log.Printf("[VisualValidator] 图像质量分析失败: %v", err)
		qualityScore = 0.7 // 默认质量分数
	}

	// 5. 综合评分 (相似度权重0.7，质量权重0.3)
	finalScore := averageSimilarity*0.7 + qualityScore*0.3

	log.Printf("[VisualValidator] 视觉一致性验证完成: 相似度=%.3f, 质量=%.3f, 综合=%.3f", 
		averageSimilarity, qualityScore, finalScore)

	return finalScore, nil
}

// FeatureExtractor 特征提取器
type FeatureExtractor struct {
	// 可以集成深度学习模型进行特征提取
}

// NewFeatureExtractor 创建特征提取器
func NewFeatureExtractor() *FeatureExtractor {
	return &FeatureExtractor{}
}

// ImageFeatures 图像特征
type ImageFeatures struct {
	ColorHistogram    []float64 `json:"color_histogram"`    // 颜色直方图
	TextureFeatures   []float64 `json:"texture_features"`   // 纹理特征
	ShapeFeatures     []float64 `json:"shape_features"`     // 形状特征
	FacialFeatures    []float64 `json:"facial_features"`    // 面部特征
	StyleFeatures     []float64 `json:"style_features"`     // 风格特征
	SemanticFeatures  []float64 `json:"semantic_features"`  // 语义特征
}

// ExtractFeatures 提取图像特征
func (fe *FeatureExtractor) ExtractFeatures(imagePath string) (*ImageFeatures, error) {
	log.Printf("[FeatureExtractor] 提取图像特征: %s", imagePath)

	// 这里应该使用真实的计算机视觉库（如OpenCV、PIL等）
	// 现在使用模拟数据
	features := &ImageFeatures{
		ColorHistogram:   fe.generateMockColorHistogram(),
		TextureFeatures:  fe.generateMockTextureFeatures(),
		ShapeFeatures:    fe.generateMockShapeFeatures(),
		FacialFeatures:   fe.generateMockFacialFeatures(),
		StyleFeatures:    fe.generateMockStyleFeatures(),
		SemanticFeatures: fe.generateMockSemanticFeatures(),
	}

	return features, nil
}

// 生成模拟特征数据的方法
func (fe *FeatureExtractor) generateMockColorHistogram() []float64 {
	histogram := make([]float64, 256)
	for i := range histogram {
		histogram[i] = rand.Float64()
	}
	return histogram
}

func (fe *FeatureExtractor) generateMockTextureFeatures() []float64 {
	features := make([]float64, 64)
	for i := range features {
		features[i] = rand.Float64()
	}
	return features
}

func (fe *FeatureExtractor) generateMockShapeFeatures() []float64 {
	features := make([]float64, 32)
	for i := range features {
		features[i] = rand.Float64()
	}
	return features
}

func (fe *FeatureExtractor) generateMockFacialFeatures() []float64 {
	features := make([]float64, 128)
	for i := range features {
		features[i] = rand.Float64()
	}
	return features
}

func (fe *FeatureExtractor) generateMockStyleFeatures() []float64 {
	features := make([]float64, 64)
	for i := range features {
		features[i] = rand.Float64()
	}
	return features
}

func (fe *FeatureExtractor) generateMockSemanticFeatures() []float64 {
	features := make([]float64, 256)
	for i := range features {
		features[i] = rand.Float64()
	}
	return features
}

// SimilarityEngine 相似度计算引擎
type SimilarityEngine struct {
	weights map[string]float64 // 不同特征的权重
}

// NewSimilarityEngine 创建相似度计算引擎
func NewSimilarityEngine() *SimilarityEngine {
	return &SimilarityEngine{
		weights: map[string]float64{
			"color":    0.15, // 颜色权重
			"texture":  0.10, // 纹理权重
			"shape":    0.15, // 形状权重
			"facial":   0.35, // 面部特征权重（最重要）
			"style":    0.15, // 风格权重
			"semantic": 0.10, // 语义权重
		},
	}
}

// CalculateSimilarity 计算两个图像特征的相似度
func (se *SimilarityEngine) CalculateSimilarity(features1, features2 *ImageFeatures) float64 {
	// 计算各个特征维度的相似度
	colorSim := se.calculateVectorSimilarity(features1.ColorHistogram, features2.ColorHistogram)
	textureSim := se.calculateVectorSimilarity(features1.TextureFeatures, features2.TextureFeatures)
	shapeSim := se.calculateVectorSimilarity(features1.ShapeFeatures, features2.ShapeFeatures)
	facialSim := se.calculateVectorSimilarity(features1.FacialFeatures, features2.FacialFeatures)
	styleSim := se.calculateVectorSimilarity(features1.StyleFeatures, features2.StyleFeatures)
	semanticSim := se.calculateVectorSimilarity(features1.SemanticFeatures, features2.SemanticFeatures)

	// 加权平均
	totalSimilarity := colorSim*se.weights["color"] +
		textureSim*se.weights["texture"] +
		shapeSim*se.weights["shape"] +
		facialSim*se.weights["facial"] +
		styleSim*se.weights["style"] +
		semanticSim*se.weights["semantic"]

	return totalSimilarity
}

// calculateVectorSimilarity 计算向量相似度（余弦相似度）
func (se *SimilarityEngine) calculateVectorSimilarity(vec1, vec2 []float64) float64 {
	if len(vec1) != len(vec2) {
		return 0
	}

	var dotProduct, norm1, norm2 float64
	for i := 0; i < len(vec1); i++ {
		dotProduct += vec1[i] * vec2[i]
		norm1 += vec1[i] * vec1[i]
		norm2 += vec2[i] * vec2[i]
	}

	if norm1 == 0 || norm2 == 0 {
		return 0
	}

	return dotProduct / (math.Sqrt(norm1) * math.Sqrt(norm2))
}

// ImageQualityAnalyzer 图像质量分析器
type ImageQualityAnalyzer struct {
	// 可以集成图像质量评估模型
}

// NewImageQualityAnalyzer 创建图像质量分析器
func NewImageQualityAnalyzer() *ImageQualityAnalyzer {
	return &ImageQualityAnalyzer{}
}

// AnalyzeImageQuality 分析图像质量
func (iqa *ImageQualityAnalyzer) AnalyzeImageQuality(imagePath string) (float64, error) {
	log.Printf("[ImageQualityAnalyzer] 分析图像质量: %s", imagePath)

	// 这里应该使用真实的图像质量评估算法
	// 现在使用基于文件名和随机因子的模拟评估

	// 基础质量分数
	baseScore := 0.7

	// 根据文件名调整分数（模拟不同质量的图像）
	filename := strings.ToLower(filepath.Base(imagePath))
	
	if strings.Contains(filename, "high") || strings.Contains(filename, "hd") {
		baseScore += 0.2
	}
	if strings.Contains(filename, "low") || strings.Contains(filename, "blur") {
		baseScore -= 0.3
	}
	if strings.Contains(filename, "noise") {
		baseScore -= 0.2
	}

	// 添加随机变化
	randomFactor := (rand.Float64() - 0.5) * 0.2 // -0.1 到 0.1 的随机变化
	finalScore := baseScore + randomFactor

	// 确保分数在合理范围内
	if finalScore > 1.0 {
		finalScore = 1.0
	}
	if finalScore < 0.0 {
		finalScore = 0.0
	}

	return finalScore, nil
}

// BatchValidateConsistency 批量验证一致性
func (vcv *VisualConsistencyValidator) BatchValidateConsistency(ctx context.Context, imageGroup []string) (*BatchValidationResult, error) {
	log.Printf("[VisualValidator] 批量验证一致性: %d 张图像", len(imageGroup))

	if len(imageGroup) < 2 {
		return &BatchValidationResult{
			OverallConsistency: 1.0,
			PairwiseScores:     make(map[string]float64),
			Recommendations:    []string{"图像数量不足，无法进行一致性验证"},
		}, nil
	}

	result := &BatchValidationResult{
		PairwiseScores:  make(map[string]float64),
		Recommendations: make([]string, 0),
	}

	var totalScore float64
	pairCount := 0

	// 计算所有图像对的相似度
	for i := 0; i < len(imageGroup); i++ {
		for j := i + 1; j < len(imageGroup); j++ {
			score, err := vcv.ValidateVisualConsistency(imageGroup[i], []string{imageGroup[j]})
			if err != nil {
				log.Printf("[VisualValidator] 验证图像对失败: %s <-> %s, 错误: %v", 
					filepath.Base(imageGroup[i]), filepath.Base(imageGroup[j]), err)
				continue
			}

			pairKey := fmt.Sprintf("%s-%s", filepath.Base(imageGroup[i]), filepath.Base(imageGroup[j]))
			result.PairwiseScores[pairKey] = score
			totalScore += score
			pairCount++
		}
	}

	// 计算整体一致性
	if pairCount > 0 {
		result.OverallConsistency = totalScore / float64(pairCount)
	}

	// 生成建议
	if result.OverallConsistency < 0.6 {
		result.Recommendations = append(result.Recommendations, "整体一致性较低，建议重新生成部分图像")
	}
	if result.OverallConsistency < 0.8 {
		result.Recommendations = append(result.Recommendations, "建议优化角色描述和生成参数")
	}

	log.Printf("[VisualValidator] 批量验证完成: 整体一致性=%.3f", result.OverallConsistency)

	return result, nil
}

// BatchValidationResult 批量验证结果
type BatchValidationResult struct {
	OverallConsistency float64            `json:"overall_consistency"`
	PairwiseScores     map[string]float64 `json:"pairwise_scores"`
	Recommendations    []string           `json:"recommendations"`
}

// SaveFeaturesToCache 保存特征到缓存
func (fe *FeatureExtractor) SaveFeaturesToCache(imagePath string, features *ImageFeatures) error {
	cacheDir := filepath.Join(filepath.Dir(imagePath), ".features_cache")
	if err := os.MkdirAll(cacheDir, 0755); err != nil {
		return err
	}

	filename := filepath.Base(imagePath)
	cacheFile := filepath.Join(cacheDir, filename+".features.json")

	data, err := json.Marshal(features)
	if err != nil {
		return err
	}

	return os.WriteFile(cacheFile, data, 0644)
}

// LoadFeaturesFromCache 从缓存加载特征
func (fe *FeatureExtractor) LoadFeaturesFromCache(imagePath string) (*ImageFeatures, error) {
	cacheDir := filepath.Join(filepath.Dir(imagePath), ".features_cache")
	filename := filepath.Base(imagePath)
	cacheFile := filepath.Join(cacheDir, filename+".features.json")

	data, err := os.ReadFile(cacheFile)
	if err != nil {
		return nil, err
	}

	var features ImageFeatures
	err = json.Unmarshal(data, &features)
	if err != nil {
		return nil, err
	}

	return &features, nil
}

// ExtractFeaturesWithCache 带缓存的特征提取
func (fe *FeatureExtractor) ExtractFeaturesWithCache(imagePath string) (*ImageFeatures, error) {
	// 尝试从缓存加载
	if features, err := fe.LoadFeaturesFromCache(imagePath); err == nil {
		log.Printf("[FeatureExtractor] 从缓存加载特征: %s", imagePath)
		return features, nil
	}

	// 提取新特征
	features, err := fe.ExtractFeatures(imagePath)
	if err != nil {
		return nil, err
	}

	// 保存到缓存
	if err := fe.SaveFeaturesToCache(imagePath, features); err != nil {
		log.Printf("[FeatureExtractor] 保存特征到缓存失败: %v", err)
	}

	return features, nil
}
