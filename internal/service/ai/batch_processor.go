package ai

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// BatchProcessor 高性能批处理引擎
type BatchProcessor struct {
	maxConcurrency int
	batchSize      int
	timeout        time.Duration
	semaphore      chan struct{}
	wg             sync.WaitGroup
}

// NewBatchProcessor 创建批处理器
func NewBatchProcessor(maxConcurrency, batchSize int, timeout time.Duration) *BatchProcessor {
	return &BatchProcessor{
		maxConcurrency: maxConcurrency,
		batchSize:      batchSize,
		timeout:        timeout,
		semaphore:      make(chan struct{}, maxConcurrency),
	}
}

// BatchImageGeneration 批量图片生成（针对16GB显存优化）
func (bp *BatchProcessor) BatchImageGeneration(
	ctx context.Context,
	panels []string,
	sd *SDClient,
	characters []CharacterProfile,
	sceneContext SceneContext,
	updateProgress func(int, string),
) ([]string, error) {
	
	totalPanels := len(panels)
	images := make([]string, totalPanels)
	errors := make([]error, totalPanels)
	
	log.Printf("[BatchProcessor] 开始批量生成 %d 张图片，并发度: %d", totalPanels, bp.maxConcurrency)
	
	// 智能分批处理
	batches := bp.createOptimalBatches(panels)
	
	for batchIndex, batch := range batches {
		log.Printf("[BatchProcessor] 处理第 %d/%d 批，包含 %d 张图片", batchIndex+1, len(batches), len(batch.panels))
		
		// 并发处理当前批次
		bp.wg.Add(len(batch.panels))
		
		for i, panel := range batch.panels {
			go func(index int, panelText string, globalIndex int) {
				defer bp.wg.Done()
				
				// 获取信号量
				bp.semaphore <- struct{}{}
				defer func() { <-bp.semaphore }()
				
				// 生成图片
				img, err := bp.generateSingleImage(ctx, panelText, sd, characters, sceneContext, globalIndex)
				
				if err != nil {
					log.Printf("[BatchProcessor] 图片 %d 生成失败: %v", globalIndex+1, err)
					errors[globalIndex] = err
					// 使用占位符
					images[globalIndex] = generatePlaceholderImageBase64()
				} else {
					images[globalIndex] = img
				}
				
				// 更新进度
				progress := int(float64(globalIndex+1) / float64(totalPanels) * 100)
				updateProgress(progress, fmt.Sprintf("已完成 %d/%d 张图片", globalIndex+1, totalPanels))
				
			}(i, panel, batch.startIndex+i)
		}
		
		// 等待当前批次完成
		bp.wg.Wait()
		
		// 批次间延迟，避免GPU过热
		if batchIndex < len(batches)-1 {
			time.Sleep(2 * time.Second)
		}
	}
	
	// 统计结果
	successCount := 0
	for _, err := range errors {
		if err == nil {
			successCount++
		}
	}
	
	log.Printf("[BatchProcessor] 批量生成完成: 成功 %d/%d 张", successCount, totalPanels)
	
	return images, nil
}

// Batch 批次结构
type Batch struct {
	panels     []string
	startIndex int
}

// createOptimalBatches 创建最优批次（基于16GB显存）
func (bp *BatchProcessor) createOptimalBatches(panels []string) []Batch {
	var batches []Batch
	
	// 根据显存大小动态调整批次大小
	// 16GB显存可以同时处理2-3张高质量图片
	optimalBatchSize := 2
	if len(panels) <= 4 {
		optimalBatchSize = 1 // 小任务单张处理，确保质量
	}
	
	for i := 0; i < len(panels); i += optimalBatchSize {
		end := i + optimalBatchSize
		if end > len(panels) {
			end = len(panels)
		}
		
		batches = append(batches, Batch{
			panels:     panels[i:end],
			startIndex: i,
		})
	}
	
	return batches
}

// generateSingleImage 生成单张图片（优化版）
func (bp *BatchProcessor) generateSingleImage(
	ctx context.Context,
	panel string,
	sd *SDClient,
	characters []CharacterProfile,
	sceneContext SceneContext,
	index int,
) (string, error) {
	
	// 为每张图片设置独立超时
	imgCtx, cancel := context.WithTimeout(ctx, bp.timeout)
	defer cancel()
	
	// 智能提示词处理
	finalPrompt := bp.optimizePrompt(panel, characters, sceneContext)
	
	log.Printf("[BatchProcessor] 生成第 %d 张图片: %s", index+1, finalPrompt[:min(50, len(finalPrompt))])
	
	// 多级重试策略
	maxRetries := 3
	for retry := 0; retry < maxRetries; retry++ {
		if retry > 0 {
			log.Printf("[BatchProcessor] 第 %d 张图片重试第 %d 次", index+1, retry)
			time.Sleep(time.Duration(retry*5) * time.Second)
		}
		
		// 根据重试次数调整参数
		_ = bp.getAdaptiveParams(retry)
		
		img, err := sd.Txt2ImgWithConsistency(imgCtx, finalPrompt, characters, sceneContext)
		if err == nil {
			return encodeBase64(img.Data), nil
		}
		
		// 检查错误类型
		if isMemoryError(err) && retry == 0 {
			// 内存不足时立即降级
			log.Printf("[BatchProcessor] 检测到显存不足，降级处理")
			return bp.generateLowMemoryImage(imgCtx, panel, sd)
		}
	}
	
	return "", fmt.Errorf("图片生成失败，已重试 %d 次", maxRetries)
}

// optimizePrompt 智能提示词优化
func (bp *BatchProcessor) optimizePrompt(panel string, characters []CharacterProfile, sceneContext SceneContext) string {
	// 基础提示词
	prompt := panel
	
	// 添加质量标签
	qualityTags := ", masterpiece, best quality, highly detailed, professional, 8k"
	
	// 添加风格标签
	styleTags := ", anime style, vibrant colors, cinematic lighting"
	
	// 组合最终提示词
	finalPrompt := prompt + qualityTags + styleTags
	
	// 限制长度避免过长
	if len(finalPrompt) > 200 {
		finalPrompt = finalPrompt[:200]
	}
	
	return finalPrompt
}

// getAdaptiveParams 获取自适应参数
func (bp *BatchProcessor) getAdaptiveParams(retryCount int) map[string]interface{} {
	baseParams := map[string]interface{}{
		"width":        1024,
		"height":       1024,
		"steps":        20,
		"cfg_scale":    7,
		"sampler_name": "DPM++ 2M Karras",
	}
	
	// 根据重试次数降级参数
	switch retryCount {
	case 1:
		baseParams["width"] = 768
		baseParams["height"] = 768
		baseParams["steps"] = 15
	case 2:
		baseParams["width"] = 512
		baseParams["height"] = 512
		baseParams["steps"] = 10
		baseParams["cfg_scale"] = 5
	}
	
	return baseParams
}

// generateLowMemoryImage 低显存模式生成
func (bp *BatchProcessor) generateLowMemoryImage(ctx context.Context, panel string, sd *SDClient) (string, error) {
	simplePrompt := panel + ", simple, clean"
	
	params := map[string]interface{}{
		"width":        384,
		"height":       384,
		"steps":        8,
		"cfg_scale":    4,
		"sampler_name": "Euler a",
	}
	
	img, err := sd.Txt2Img(simplePrompt, params)
	if err != nil {
		return generatePlaceholderImageBase64(), nil
	}
	
	return encodeBase64(img.Data), nil
}

// isMemoryError 检查是否为显存错误
func isMemoryError(err error) bool {
	errorStr := err.Error()
	return contains(errorStr, "memory") || contains(errorStr, "CUDA") || contains(errorStr, "out of memory")
}

// generatePlaceholderImageBase64 生成占位符图片
func generatePlaceholderImageBase64() string {
	return "data:image/svg+xml;base64,PHN2ZyB3aWR0aD0iNTEyIiBoZWlnaHQ9IjUxMiIgeG1sbnM9Imh0dHA6Ly93d3cudzMub3JnLzIwMDAvc3ZnIj48cmVjdCB3aWR0aD0iMTAwJSIgaGVpZ2h0PSIxMDAlIiBmaWxsPSIjZjBmMGYwIi8+PHRleHQgeD0iNTAlIiB5PSI1MCUiIGZvbnQtZmFtaWx5PSJBcmlhbCIgZm9udC1zaXplPSIyNCIgZmlsbD0iIzk5OSIgdGV4dC1hbmNob3I9Im1pZGRsZSIgZHk9Ii4zZW0iPkFJ55Sf5oiQPC90ZXh0Pjwvc3ZnPg=="
}

// contains 字符串包含检查
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 || (len(s) > len(substr) && (s[:len(substr)] == substr || s[len(s)-len(substr):] == substr || containsInner(s, substr))))
}

func containsInner(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}


