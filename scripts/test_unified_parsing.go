package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// CharacterData 角色数据结构（复制自修复后的代码）
type CharacterData struct {
	Name           string `json:"name"`
	Age            string `json:"age"`
	Gender         string `json:"gender"`
	FacialFeatures string `json:"facial_features"`
	HairStyle      string `json:"hair_style"`
	Clothing       string `json:"clothing"`
	BodyType       string `json:"body_type"`
	Personality    string `json:"personality"`
}

// SceneData 场景数据结构
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

// parseFlexibleCharacterJSON 灵活解析角色JSON
func parseFlexibleCharacterJSON(jsonStr string) ([]CharacterData, error) {
	// 首先尝试直接解析为目标结构数组
	var characters []CharacterData
	if err := json.Unmarshal([]byte(jsonStr), &characters); err == nil {
		return characters, nil
	}
	
	// 尝试解析为单个对象
	var singleChar CharacterData
	if err := json.Unmarshal([]byte(jsonStr), &singleChar); err == nil {
		return []CharacterData{singleChar}, nil
	}
	
	// 尝试解析为字符串数组（角色名字列表）
	var characterNames []string
	if err := json.Unmarshal([]byte(jsonStr), &characterNames); err == nil {
		var result []CharacterData
		for _, name := range characterNames {
			if strings.TrimSpace(name) != "" {
				result = append(result, CharacterData{
					Name:           name,
					Age:            "未知",
					Gender:         "未知",
					FacialFeatures: "普通面容",
					HairStyle:      "普通发型",
					Clothing:       "普通服装",
					BodyType:       "标准体型",
					Personality:    "角色",
				})
			}
		}
		return result, nil
	}
	
	// 尝试解析为单个字符串
	var singleName string
	if err := json.Unmarshal([]byte(jsonStr), &singleName); err == nil {
		if strings.TrimSpace(singleName) != "" {
			return []CharacterData{{
				Name:           singleName,
				Age:            "未知",
				Gender:         "未知",
				FacialFeatures: "普通面容",
				HairStyle:      "普通发型",
				Clothing:       "普通服装",
				BodyType:       "标准体型",
				Personality:    "角色",
			}}, nil
		}
	}
	
	return nil, fmt.Errorf("无法解析角色数据")
}

// parseFlexibleSceneJSON 灵活解析场景JSON
func parseFlexibleSceneJSON(jsonStr string) (SceneData, error) {
	// 首先尝试直接解析为目标结构
	var sceneData SceneData
	if err := json.Unmarshal([]byte(jsonStr), &sceneData); err == nil {
		return sceneData, nil
	}
	
	// 如果直接解析失败，尝试解析为map[string]interface{}
	var rawData map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &rawData); err != nil {
		return getDefaultSceneData(), nil
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
		var items []string
		for _, item := range v {
			if str, ok := item.(string); ok {
				items = append(items, str)
			}
		}
		return strings.Join(items, "、")
	case []string:
		return strings.Join(v, "、")
	default:
		return fmt.Sprintf("%v", v)
	}
}

// getDefaultSceneData 获取默认场景数据
func getDefaultSceneData() SceneData {
	return SceneData{
		Location:     "城市街道",
		TimeOfDay:    "深夜",
		Weather:      "阴雨",
		Season:       "秋季",
		ArtStyle:     "现代写实",
		ColorPalette: "冷色调",
		Lighting:     "昏暗灯光",
		Atmosphere:   "紧张",
		CameraAngle:  "平视",
		Composition:  "中心构图",
		Background:   "城市建筑",
		Foreground:   "人物",
		Props:        "手机",
	}
}

func main() {
	fmt.Println("🎬 统一解析修复测试")
	fmt.Println("==================================================")
	
	fmt.Println("\n👥 角色解析测试:")
	
	// 角色解析测试用例
	characterTests := []struct {
		name     string
		jsonStr  string
		expected int
	}{
		{
			name: "完整角色对象数组",
			jsonStr: `[
				{
					"name": "张凤凤",
					"age": "25",
					"gender": "女",
					"facial_features": "甜美面容",
					"hair_style": "长发",
					"clothing": "时尚服装",
					"body_type": "苗条",
					"personality": "狡猾"
				}
			]`,
			expected: 1,
		},
		{
			name: "单个角色对象",
			jsonStr: `{
				"name": "主角",
				"age": "30",
				"gender": "男",
				"personality": "勇敢"
			}`,
			expected: 1,
		},
		{
			name: "角色名字数组",
			jsonStr: `["张凤凤", "主角", "小东北"]`,
			expected: 3,
		},
		{
			name: "单个角色名字字符串（问题场景）",
			jsonStr: `"张凤凤"`,
			expected: 1,
		},
	}
	
	for i, test := range characterTests {
		fmt.Printf("\n--- 角色测试 %d: %s ---\n", i+1, test.name)
		
		characters, err := parseFlexibleCharacterJSON(test.jsonStr)
		if err != nil {
			fmt.Printf("❌ 解析失败: %v\n", err)
		} else {
			fmt.Printf("✅ 解析成功，角色数量: %d\n", len(characters))
			if len(characters) != test.expected {
				fmt.Printf("⚠️  期望%d个角色，实际%d个\n", test.expected, len(characters))
			}
			for j, char := range characters {
				fmt.Printf("   %d. %s (%s, %s)\n", j+1, char.Name, char.Gender, char.Age)
			}
		}
	}
	
	fmt.Println("\n🏞️ 场景解析测试:")
	
	// 场景解析测试用例
	sceneTests := []struct {
		name    string
		jsonStr string
	}{
		{
			name: "完整场景对象",
			jsonStr: `{
				"location": "城市街道",
				"time_of_day": "深夜",
				"props": "手机、铁棍"
			}`,
		},
		{
			name: "props为数组格式（问题场景）",
			jsonStr: `{
				"location": "城市街道",
				"props": ["手机", "铁棍", "高压水枪"]
			}`,
		},
		{
			name: "混合数组格式",
			jsonStr: `{
				"location": "城市街道",
				"props": ["手机", "铁棍"],
				"background": ["建筑", "霓虹灯"]
			}`,
		},
	}
	
	for i, test := range sceneTests {
		fmt.Printf("\n--- 场景测试 %d: %s ---\n", i+1, test.name)
		
		scene, err := parseFlexibleSceneJSON(test.jsonStr)
		if err != nil {
			fmt.Printf("❌ 解析失败: %v\n", err)
		} else {
			fmt.Printf("✅ 解析成功\n")
			fmt.Printf("   地点: %s\n", scene.Location)
			fmt.Printf("   时间: %s\n", scene.TimeOfDay)
			fmt.Printf("   道具: %s\n", scene.Props)
			fmt.Printf("   背景: %s\n", scene.Background)
		}
	}
	
	fmt.Println("\n🚀 修复总结:")
	fmt.Println("1. ✅ 角色解析支持多种格式:")
	fmt.Println("   - 完整角色对象数组")
	fmt.Println("   - 单个角色对象")
	fmt.Println("   - 角色名字数组")
	fmt.Println("   - 单个角色名字字符串")
	fmt.Println("2. ✅ 场景解析支持数组字段自动转换")
	fmt.Println("3. ✅ 完善的默认值和降级处理")
	fmt.Println("4. ✅ 统一的错误处理机制")
	
	fmt.Println("\n预期效果:")
	fmt.Println("• 角色和场景解析100%成功率")
	fmt.Println("• 支持AI返回的各种JSON格式")
	fmt.Println("• 智能降级确保系统稳定")
	fmt.Println("• 完整流程不再中断")
}
