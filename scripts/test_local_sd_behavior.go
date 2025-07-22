package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// SDTestRequest SD测试请求结构
type SDTestRequest struct {
	Prompt       string `json:"prompt"`
	Width        int    `json:"width"`
	Height       int    `json:"height"`
	Steps        int    `json:"steps"`
	CfgScale     int    `json:"cfg_scale"`
	SamplerName  string `json:"sampler_name"`
	Seed         int    `json:"seed"`
}

// SDTestResponse SD测试响应结构
type SDTestResponse struct {
	Images []string `json:"images"`
	Info   string   `json:"info"`
}

// testSDPrompt 测试SD提示词
func testSDPrompt(endpoint, prompt string) (bool, string, error) {
	// 构建请求
	reqBody := SDTestRequest{
		Prompt:      prompt,
		Width:       512,
		Height:      512,
		Steps:       20,
		CfgScale:    7,
		SamplerName: "DPM++ 2M Karras",
		Seed:        -1,
	}
	
	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return false, "", fmt.Errorf("JSON编码失败: %v", err)
	}
	
	// 发送请求
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Post(endpoint+"/sdapi/v1/txt2img", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return false, "", fmt.Errorf("请求失败: %v", err)
	}
	defer resp.Body.Close()
	
	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, "", fmt.Errorf("读取响应失败: %v", err)
	}
	
	if resp.StatusCode != 200 {
		return false, string(body), fmt.Errorf("SD API错误: %s", resp.Status)
	}
	
	// 解析响应
	var sdResp SDTestResponse
	if err := json.Unmarshal(body, &sdResp); err != nil {
		return false, string(body), fmt.Errorf("响应解析失败: %v", err)
	}
	
	// 检查是否生成了图像
	success := len(sdResp.Images) > 0 && sdResp.Images[0] != ""
	return success, sdResp.Info, nil
}

func main() {
	fmt.Println("🔍 本地SD行为测试")
	fmt.Println("==================================================")
	
	// SD服务端点 - 请根据您的实际配置修改
	sdEndpoint := "http://127.0.0.1:7860"
	
	fmt.Printf("测试SD端点: %s\n", sdEndpoint)
	
	// 测试用例
	testCases := []struct {
		name        string
		prompt      string
		expectIssue bool
		description string
	}{
		{
			name:        "您的原始问题提示词",
			prompt:      "Northern Myanmar people armed with guns surrounded the group. Zhang Fengfeng smirked, turned, and entered the small building, guns aimed at us., high quality, detailed",
			expectIssue: true,
			description: "复杂叙事 + 敏感词汇",
		},
		{
			name:        "简化的敏感词汇测试",
			prompt:      "people with guns, high quality, detailed",
			expectIssue: false,
			description: "仅包含敏感词汇",
		},
		{
			name:        "地理位置测试",
			prompt:      "Northern Myanmar landscape, high quality, detailed",
			expectIssue: false,
			description: "仅包含地理位置",
		},
		{
			name:        "复杂叙事测试",
			prompt:      "A person walked, turned, entered building, looked around, smiled, then sat down, high quality, detailed",
			expectIssue: true,
			description: "仅包含复杂叙事",
		},
		{
			name:        "简单清晰测试",
			prompt:      "a young woman with long hair, anime style, high quality, detailed",
			expectIssue: false,
			description: "简单清晰的描述",
		},
		{
			name:        "优化后的提示词",
			prompt:      "people from remote area holding objects gathered around group, Zhang Fengfeng smiling, anime style, illustration, high quality, detailed",
			expectIssue: false,
			description: "经过优化的版本",
		},
	}
	
	fmt.Println("\n📝 开始测试各种提示词:")
	
	successCount := 0
	totalCount := len(testCases)
	
	for i, tc := range testCases {
		fmt.Printf("\n--- 测试 %d: %s ---\n", i+1, tc.name)
		fmt.Printf("描述: %s\n", tc.description)
		fmt.Printf("提示词: %s\n", tc.prompt)
		
		success, info, err := testSDPrompt(sdEndpoint, tc.prompt)
		
		if err != nil {
			fmt.Printf("❌ 测试失败: %v\n", err)
			if tc.expectIssue {
				fmt.Printf("✅ 符合预期 - 确实存在问题\n")
			} else {
				fmt.Printf("⚠️  意外失败 - 预期应该成功\n")
			}
		} else if success {
			fmt.Printf("✅ 生成成功\n")
			successCount++
			if tc.expectIssue {
				fmt.Printf("⚠️  意外成功 - 预期可能有问题\n")
			} else {
				fmt.Printf("✅ 符合预期 - 正常生成\n")
			}
		} else {
			fmt.Printf("❌ 生成失败\n")
			if tc.expectIssue {
				fmt.Printf("✅ 符合预期 - 确实存在问题\n")
			} else {
				fmt.Printf("⚠️  意外失败 - 预期应该成功\n")
			}
		}
		
		if info != "" {
			fmt.Printf("SD信息: %s\n", info[:min(100, len(info))])
		}
	}
	
	fmt.Printf("\n📊 测试结果统计:\n")
	fmt.Printf("成功率: %d/%d (%.1f%%)\n", successCount, totalCount, float64(successCount)/float64(totalCount)*100)
	
	fmt.Println("\n🔍 分析结论:")
	
	if successCount == totalCount {
		fmt.Println("✅ 所有测试都成功 - 本地SD没有内容限制")
		fmt.Println("   问题可能在于:")
		fmt.Println("   1. 提示词过于复杂")
		fmt.Println("   2. SD模型理解能力限制")
		fmt.Println("   3. 参数设置问题")
	} else if successCount == 0 {
		fmt.Println("❌ 所有测试都失败 - 可能的原因:")
		fmt.Println("   1. SD服务未启动或端点错误")
		fmt.Println("   2. SD模型加载失败")
		fmt.Println("   3. 系统资源不足")
	} else {
		fmt.Println("⚠️  部分测试失败 - 分析:")
		fmt.Println("   1. 复杂提示词确实有问题")
		fmt.Println("   2. 简单提示词应该能成功")
		fmt.Println("   3. 需要优化提示词质量")
	}
	
	fmt.Println("\n💡 建议:")
	fmt.Println("1. 如果简单提示词能成功，说明问题在于提示词复杂度")
	fmt.Println("2. 如果敏感词汇能成功，说明本地SD没有内容过滤")
	fmt.Println("3. 如果优化后的提示词成功率更高，说明优化有效")
	fmt.Println("4. 建议使用简洁、清晰的提示词描述")
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
