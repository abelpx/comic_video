package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// SceneData 场景数据结构（复制自修复后的代码）
type SceneData struct {
	Location     string `json:"location"`
	TimeOfDay    string `json:"time_of_day"`
	Weather      string `json:"weather"`
	Season       string `json:"season"`
	ArtStyle     string `json:"art_style"`
	ColorPalette string `json:"color_palette"`
	Lighting     string `json:"lighting"`
	Atmosphere   string `json:"atmosphere"`
	CameraAngle  string `json:"camera_angle"`
	Composition  string `json:"composition"`
	Background   string `json:"background"`
	Foreground   string `json:"foreground"`
	Props        string `json:"props"`
}

// parseFlexibleSceneJSON 灵活解析场景JSON，处理数组字段
func parseFlexibleSceneJSON(jsonStr string) (SceneData, error) {
	// 首先尝试直接解析为目标结构
	var sceneData SceneData
	if err := json.Unmarshal([]byte(jsonStr), &sceneData); err == nil {
		return sceneData, nil
	}
	
	// 如果直接解析失败，尝试解析为map[string]interface{}
	var rawData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &rawData); err != nil {
		return SceneData{}, fmt.Errorf("无法解析JSON: %v", err)
	}
	
	// 手动转换每个字段，处理可能的数组
	sceneData = SceneData{
		Location:     getStringValue(rawData, "location"),
		TimeOfDay:    getStringValue(rawData, "time_of_day"),
		Weather:      getStringValue(rawData, "weather"),
		Season:       getStringValue(rawData, "season"),
		ArtStyle:     getStringValue(rawData, "art_style"),
		ColorPalette: getStringValue(rawData, "color_palette"),
		Lighting:     getStringValue(rawData, "lighting"),
		Atmosphere:   getStringValue(rawData, "atmosphere"),
		CameraAngle:  getStringValue(rawData, "camera_angle"),
		Composition:  getStringValue(rawData, "composition"),
		Background:   getStringValue(rawData, "background"),
		Foreground:   getStringValue(rawData, "foreground"),
		Props:        getStringValue(rawData, "props"),
	}
	
	return sceneData, nil
}

// getStringValue 从map中获取字符串值，处理数组情况
func getStringValue(data map[string]interface{}, key string) string {
	value, exists := data[key]
	if !exists {
		return ""
	}
	
	switch v := value.(type) {
	case string:
		return v
	case []interface{}:
		// 如果是数组，转换为逗号分隔的字符串
		var items []string
		for _, item := range v {
			if str, ok := item.(string); ok {
				items = append(items, str)
			}
		}
		return strings.Join(items, "、")
	case []string:
		// 如果是字符串数组
		return strings.Join(v, "、")
	default:
		// 其他类型转换为字符串
		return fmt.Sprintf("%v", v)
	}
}

func main() {
	fmt.Println("🔧 JSON解析修复测试")
	fmt.Println("==================================================")
	
	// 测试不同的JSON格式
	testCases := []struct {
		name     string
		jsonStr  string
		expected bool
	}{
		{
			name: "正常的字符串格式",
			jsonStr: `{
				"location": "城市街道",
				"time_of_day": "深夜",
				"weather": "阴雨",
				"props": "手机、铁棍"
			}`,
			expected: true,
		},
		{
			name: "props为数组格式（问题场景）",
			jsonStr: `{
				"location": "城市街道",
				"time_of_day": "深夜", 
				"weather": "阴雨",
				"props": ["手机", "铁棍", "高压水枪"]
			}`,
			expected: true,
		},
		{
			name: "混合格式",
			jsonStr: `{
				"location": "城市街道",
				"time_of_day": "深夜",
				"weather": "阴雨", 
				"props": ["手机", "铁棍"],
				"background": "城市建筑",
				"foreground": ["人物", "车辆"]
			}`,
			expected: true,
		},
		{
			name: "不完整的JSON",
			jsonStr: `{
				"location": "城市街道",
				"time_of_day": "深夜"
			}`,
			expected: true,
		},
		{
			name: "无效的JSON",
			jsonStr: `{invalid json}`,
			expected: false,
		},
	}
	
	fmt.Println("\n📝 测试JSON解析:")
	
	for i, tc := range testCases {
		fmt.Printf("\n--- 测试 %d: %s ---\n", i+1, tc.name)
		fmt.Printf("输入JSON: %s\n", tc.jsonStr)
		
		sceneData, err := parseFlexibleSceneJSON(tc.jsonStr)
		
		if err != nil {
			if tc.expected {
				fmt.Printf("❌ 解析失败（预期成功）: %v\n", err)
			} else {
				fmt.Printf("✅ 解析失败（符合预期）: %v\n", err)
			}
		} else {
			if tc.expected {
				fmt.Printf("✅ 解析成功\n")
				fmt.Printf("   地点: %s\n", sceneData.Location)
				fmt.Printf("   时间: %s\n", sceneData.TimeOfDay)
				fmt.Printf("   天气: %s\n", sceneData.Weather)
				fmt.Printf("   道具: %s\n", sceneData.Props)
				fmt.Printf("   背景: %s\n", sceneData.Background)
				fmt.Printf("   前景: %s\n", sceneData.Foreground)
			} else {
				fmt.Printf("⚠️  解析成功（预期失败）\n")
			}
		}
	}
	
	fmt.Println("\n🎯 特殊测试：数组转字符串")
	
	// 测试数组转字符串的功能
	testData := map[string]interface{}{
		"props":      []interface{}{"手机", "铁棍", "高压水枪"},
		"background": []string{"城市建筑", "霓虹灯"},
		"single":     "单个字符串",
		"number":     123,
	}
	
	for key, value := range testData {
		result := getStringValue(testData, key)
		fmt.Printf("  %s: %v -> %s\n", key, value, result)
	}
	
	fmt.Println("\n🚀 修复总结:")
	fmt.Println("1. ✅ 创建了灵活的JSON解析函数")
	fmt.Println("2. ✅ 支持数组字段自动转换为字符串")
	fmt.Println("3. ✅ 处理不完整或格式错误的JSON")
	fmt.Println("4. ✅ 保持向后兼容性")
	
	fmt.Println("\n预期效果:")
	fmt.Println("• 场景分析JSON解析错误完全解决")
	fmt.Println("• 支持AI返回的各种JSON格式")
	fmt.Println("• 数组字段自动转换为逗号分隔字符串")
	fmt.Println("• 系统稳定性进一步提升")
}
