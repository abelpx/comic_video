package imageprocessing

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
)

// Service 图像后处理服务
type Service struct {
	outputDir string
}

// NewService 创建图像后处理服务
func NewService(outputDir string) *Service {
	return &Service{
		outputDir: outputDir,
	}
}

// ProcessingOptions 图像处理选项
type ProcessingOptions struct {
	TargetWidth    int     `json:"target_width"`    // 目标宽度
	TargetHeight   int     `json:"target_height"`   // 目标高度
	Quality        int     `json:"quality"`         // 质量（1-100）
	Brightness     float64 `json:"brightness"`      // 亮度调整（-100到100）
	Contrast       float64 `json:"contrast"`        // 对比度调整（-100到100）
	Saturation     float64 `json:"saturation"`      // 饱和度调整（-100到100）
	Sharpen        float64 `json:"sharpen"`         // 锐化强度（0-10）
	Blur           float64 `json:"blur"`            // 模糊强度（0-10）
	AddBorder      bool    `json:"add_border"`      // 是否添加边框
	BorderColor    string  `json:"border_color"`    // 边框颜色
	BorderWidth    int     `json:"border_width"`    // 边框宽度
	Watermark      string  `json:"watermark"`       // 水印文字
	WatermarkPos   string  `json:"watermark_pos"`   // 水印位置
	Format         string  `json:"format"`          // 输出格式（jpg/png）
}

// ProcessingResult 处理结果
type ProcessingResult struct {
	OriginalPath string `json:"original_path"`
	ProcessedPath string `json:"processed_path"`
	OriginalSize  int64  `json:"original_size"`
	ProcessedSize int64  `json:"processed_size"`
	Width         int    `json:"width"`
	Height        int    `json:"height"`
	Format        string `json:"format"`
}

// ProcessImage 处理单张图像
func (s *Service) ProcessImage(ctx context.Context, imagePath string, options *ProcessingOptions) (*ProcessingResult, error) {
	log.Printf("[ImageProcessing] 开始处理图像: %s", imagePath)

	// 1. 加载图像
	img, err := s.loadImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("加载图像失败: %w", err)
	}

	originalBounds := img.Bounds()
	log.Printf("[ImageProcessing] 原始尺寸: %dx%d", originalBounds.Dx(), originalBounds.Dy())

	// 2. 应用处理选项
	processedImg := s.applyProcessing(img, options)

	// 3. 保存处理后的图像
	outputPath, err := s.saveProcessedImage(processedImg, imagePath, options)
	if err != nil {
		return nil, fmt.Errorf("保存图像失败: %w", err)
	}

	// 4. 获取文件信息
	originalInfo, _ := os.Stat(imagePath)
	processedInfo, _ := os.Stat(outputPath)

	result := &ProcessingResult{
		OriginalPath:  imagePath,
		ProcessedPath: outputPath,
		OriginalSize:  originalInfo.Size(),
		ProcessedSize: processedInfo.Size(),
		Width:         processedImg.Bounds().Dx(),
		Height:        processedImg.Bounds().Dy(),
		Format:        options.Format,
	}

	log.Printf("[ImageProcessing] 图像处理完成: %s -> %s", imagePath, outputPath)
	return result, nil
}

// ProcessBatch 批量处理图像
func (s *Service) ProcessBatch(ctx context.Context, imagePaths []string, options *ProcessingOptions) ([]*ProcessingResult, error) {
	log.Printf("[ImageProcessing] 开始批量处理 %d 张图像", len(imagePaths))

	var results []*ProcessingResult
	for i, imagePath := range imagePaths {
		select {
		case <-ctx.Done():
			return results, ctx.Err()
		default:
		}

		log.Printf("[ImageProcessing] 处理进度: %d/%d", i+1, len(imagePaths))
		
		result, err := s.ProcessImage(ctx, imagePath, options)
		if err != nil {
			log.Printf("[ImageProcessing] 处理图像失败: %s, error: %v", imagePath, err)
			continue
		}

		results = append(results, result)
	}

	log.Printf("[ImageProcessing] 批量处理完成: %d/%d 成功", len(results), len(imagePaths))
	return results, nil
}

// StandardizeForVideo 为视频制作标准化图像
func (s *Service) StandardizeForVideo(ctx context.Context, imagePaths []string, videoFormat string) ([]*ProcessingResult, error) {
	// 根据视频格式设置标准选项
	options := s.getVideoStandardOptions(videoFormat)
	
	log.Printf("[ImageProcessing] 为视频格式 %s 标准化 %d 张图像", videoFormat, len(imagePaths))
	
	return s.ProcessBatch(ctx, imagePaths, options)
}

// getVideoStandardOptions 获取视频标准化选项
func (s *Service) getVideoStandardOptions(videoFormat string) *ProcessingOptions {
	switch strings.ToLower(videoFormat) {
	case "4k", "uhd":
		return &ProcessingOptions{
			TargetWidth:  3840,
			TargetHeight: 2160,
			Quality:      95,
			Brightness:   0,
			Contrast:     10,
			Saturation:   5,
			Sharpen:      1.0,
			Format:       "jpg",
		}
	case "1080p", "fhd":
		return &ProcessingOptions{
			TargetWidth:  1920,
			TargetHeight: 1080,
			Quality:      90,
			Brightness:   0,
			Contrast:     10,
			Saturation:   5,
			Sharpen:      1.0,
			Format:       "jpg",
		}
	case "720p", "hd":
		return &ProcessingOptions{
			TargetWidth:  1280,
			TargetHeight: 720,
			Quality:      85,
			Brightness:   0,
			Contrast:     5,
			Saturation:   0,
			Sharpen:      0.5,
			Format:       "jpg",
		}
	case "vertical", "short": // 短视频格式
		return &ProcessingOptions{
			TargetWidth:  1080,
			TargetHeight: 1920,
			Quality:      90,
			Brightness:   5,
			Contrast:     15,
			Saturation:   10,
			Sharpen:      1.5,
			Format:       "jpg",
		}
	default:
		return &ProcessingOptions{
			TargetWidth:  1920,
			TargetHeight: 1080,
			Quality:      90,
			Format:       "jpg",
		}
	}
}

// loadImage 加载图像
func (s *Service) loadImage(imagePath string) (image.Image, error) {
	file, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	return img, nil
}

// applyProcessing 应用图像处理
func (s *Service) applyProcessing(img image.Image, options *ProcessingOptions) image.Image {
	processed := img

	// 1. 调整尺寸
	if options.TargetWidth > 0 && options.TargetHeight > 0 {
		processed = imaging.Resize(processed, options.TargetWidth, options.TargetHeight, imaging.Lanczos)
	}

	// 2. 亮度调整
	if options.Brightness != 0 {
		processed = imaging.AdjustBrightness(processed, options.Brightness)
	}

	// 3. 对比度调整
	if options.Contrast != 0 {
		processed = imaging.AdjustContrast(processed, options.Contrast)
	}

	// 4. 饱和度调整
	if options.Saturation != 0 {
		processed = imaging.AdjustSaturation(processed, options.Saturation)
	}

	// 5. 锐化
	if options.Sharpen > 0 {
		processed = imaging.Sharpen(processed, options.Sharpen)
	}

	// 6. 模糊
	if options.Blur > 0 {
		processed = imaging.Blur(processed, options.Blur)
	}

	// 7. 添加边框
	if options.AddBorder && options.BorderWidth > 0 {
		processed = s.addBorder(processed, options.BorderWidth, options.BorderColor)
	}

	// 8. 添加水印
	if options.Watermark != "" {
		processed = s.addWatermark(processed, options.Watermark, options.WatermarkPos)
	}

	return processed
}

// addBorder 添加边框
func (s *Service) addBorder(img image.Image, width int, colorStr string) image.Image {
	bounds := img.Bounds()
	newBounds := image.Rect(0, 0, bounds.Dx()+2*width, bounds.Dy()+2*width)
	
	bordered := image.NewRGBA(newBounds)
	
	// 设置边框颜色
	borderColor := color.RGBA{0, 0, 0, 255} // 默认黑色
	if colorStr != "" {
		borderColor = s.parseColor(colorStr)
	}
	
	// 填充边框
	draw.Draw(bordered, newBounds, &image.Uniform{borderColor}, image.Point{}, draw.Src)
	
	// 绘制原图
	draw.Draw(bordered, bounds.Add(image.Pt(width, width)), img, bounds.Min, draw.Over)
	
	return bordered
}

// addWatermark 添加水印（简化实现）
func (s *Service) addWatermark(img image.Image, watermark, position string) image.Image {
	// 这里是简化实现，实际应该使用字体渲染库
	// 暂时返回原图
	return img
}

// parseColor 解析颜色字符串
func (s *Service) parseColor(colorStr string) color.RGBA {
	switch strings.ToLower(colorStr) {
	case "white":
		return color.RGBA{255, 255, 255, 255}
	case "black":
		return color.RGBA{0, 0, 0, 255}
	case "red":
		return color.RGBA{255, 0, 0, 255}
	case "green":
		return color.RGBA{0, 255, 0, 255}
	case "blue":
		return color.RGBA{0, 0, 255, 255}
	default:
		return color.RGBA{0, 0, 0, 255}
	}
}

// saveProcessedImage 保存处理后的图像
func (s *Service) saveProcessedImage(img image.Image, originalPath string, options *ProcessingOptions) (string, error) {
	// 生成输出文件名
	ext := filepath.Ext(originalPath)
	baseName := strings.TrimSuffix(filepath.Base(originalPath), ext)
	
	format := options.Format
	if format == "" {
		format = "jpg"
	}
	
	outputFileName := fmt.Sprintf("%s_processed.%s", baseName, format)
	outputPath := filepath.Join(s.outputDir, outputFileName)

	// 确保输出目录存在
	if err := os.MkdirAll(s.outputDir, 0755); err != nil {
		return "", err
	}

	// 创建输出文件
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return "", err
	}
	defer outputFile.Close()

	// 根据格式保存
	switch strings.ToLower(format) {
	case "png":
		err = png.Encode(outputFile, img)
	case "jpg", "jpeg":
		quality := options.Quality
		if quality == 0 {
			quality = 90
		}
		err = jpeg.Encode(outputFile, img, &jpeg.Options{Quality: quality})
	default:
		err = jpeg.Encode(outputFile, img, &jpeg.Options{Quality: 90})
	}

	if err != nil {
		return "", err
	}

	return outputPath, nil
}
