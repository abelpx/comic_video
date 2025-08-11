package scene

import (
	"context"
	"fmt"
	"log"
	"math"
	"regexp"
	"strings"
)

// SemanticAnalyzer 语义分析器
type SemanticAnalyzer struct {
	aiService AIService
}

// NewSemanticAnalyzer 创建语义分析器
func NewSemanticAnalyzer(aiService AIService) *SemanticAnalyzer {
	return &SemanticAnalyzer{
		aiService: aiService,
	}
}

// AnalyzeSemantics 分析语义
func (sa *SemanticAnalyzer) AnalyzeSemantics(ctx context.Context, text string) (*SemanticContext, error) {
	log.Printf("[SemanticAnalyzer] 分析语义")

	// 使用AI进行语义分析
	prompt := fmt.Sprintf(`
请分析以下文本的语义内容，提取：
1. 主要动作
2. 次要动作
3. 存在的物体
4. 空间关系
5. 时间标记
6. 因果关系
7. 主题
8. 象征意义

文本：%s

请以结构化格式返回分析结果。`, text)

	response, err := sa.aiService.GenerateText(ctx, prompt)
	if err != nil {
		log.Printf("[SemanticAnalyzer] AI语义分析失败: %v", err)
		// 降级到基于规则的分析
		return sa.ruleBasedSemanticAnalysis(text), nil
	}

	// 解析AI响应
	semanticContext := sa.parseSemanticResponse(response, text)
	return semanticContext, nil
}

// ruleBasedSemanticAnalysis 基于规则的语义分析
func (sa *SemanticAnalyzer) ruleBasedSemanticAnalysis(text string) *SemanticContext {
	context := &SemanticContext{
		MainAction:       sa.extractMainAction(text),
		SubActions:       sa.extractSubActions(text),
		ObjectsPresent:   sa.extractObjects(text),
		SpatialRelations: sa.extractSpatialRelations(text),
		TemporalMarkers:  sa.extractTemporalMarkers(text),
		CausalLinks:      sa.extractCausalLinks(text),
		Themes:           sa.extractThemes(text),
		Symbolism:        sa.extractSymbolism(text),
	}
	return context
}

// extractMainAction 提取主要动作
func (sa *SemanticAnalyzer) extractMainAction(text string) string {
	// 动作词汇模式
	actionPatterns := []string{
		`(\w+)着`, `(\w+)了`, `(\w+)起来`, `(\w+)下去`,
		`正在(\w+)`, `开始(\w+)`, `继续(\w+)`, `停止(\w+)`,
	}
	
	for _, pattern := range actionPatterns {
		re := regexp.MustCompile(pattern)
		matches := re.FindStringSubmatch(text)
		if len(matches) > 1 {
			return matches[1]
		}
	}
	
	return "未识别"
}

// extractSubActions 提取次要动作
func (sa *SemanticAnalyzer) extractSubActions(text string) []string {
	actions := make([]string, 0)
	
	// 常见动作词汇
	actionWords := []string{
		"走", "跑", "坐", "站", "看", "听", "说", "想", "笑", "哭",
		"拿", "放", "开", "关", "推", "拉", "举", "放下", "转身", "回头",
	}
	
	lowerText := strings.ToLower(text)
	for _, action := range actionWords {
		if strings.Contains(lowerText, action) {
			actions = append(actions, action)
		}
	}
	
	return actions
}

// extractObjects 提取物体
func (sa *SemanticAnalyzer) extractObjects(text string) []string {
	objects := make([]string, 0)
	
	// 常见物体词汇
	objectWords := []string{
		"桌子", "椅子", "门", "窗户", "书", "笔", "杯子", "手机", "电脑",
		"车", "房子", "树", "花", "山", "水", "天空", "太阳", "月亮",
		"衣服", "鞋子", "帽子", "包", "钱", "钥匙", "眼镜", "手表",
	}
	
	for _, obj := range objectWords {
		if strings.Contains(text, obj) {
			objects = append(objects, obj)
		}
	}
	
	return objects
}

// extractSpatialRelations 提取空间关系
func (sa *SemanticAnalyzer) extractSpatialRelations(text string) []string {
	relations := make([]string, 0)
	
	// 空间关系词汇
	spatialWords := []string{
		"在...上面", "在...下面", "在...旁边", "在...前面", "在...后面",
		"在...里面", "在...外面", "在...中间", "在...左边", "在...右边",
		"靠近", "远离", "围绕", "穿过", "越过", "到达", "离开",
	}
	
	for _, spatial := range spatialWords {
		if strings.Contains(text, spatial) {
			relations = append(relations, spatial)
		}
	}
	
	return relations
}

// extractTemporalMarkers 提取时间标记
func (sa *SemanticAnalyzer) extractTemporalMarkers(text string) []string {
	markers := make([]string, 0)
	
	// 时间标记词汇
	timeWords := []string{
		"早上", "中午", "下午", "晚上", "夜里", "深夜",
		"昨天", "今天", "明天", "前天", "后天",
		"现在", "刚才", "一会儿", "马上", "立刻", "突然",
		"之前", "之后", "同时", "接着", "然后", "最后",
	}
	
	for _, timeWord := range timeWords {
		if strings.Contains(text, timeWord) {
			markers = append(markers, timeWord)
		}
	}
	
	return markers
}

// extractCausalLinks 提取因果关系
func (sa *SemanticAnalyzer) extractCausalLinks(text string) []string {
	links := make([]string, 0)
	
	// 因果关系词汇
	causalWords := []string{
		"因为", "由于", "所以", "因此", "导致", "造成", "引起",
		"结果", "后果", "原因", "缘故", "为了", "目的是",
	}
	
	for _, causal := range causalWords {
		if strings.Contains(text, causal) {
			links = append(links, causal)
		}
	}
	
	return links
}

// extractThemes 提取主题
func (sa *SemanticAnalyzer) extractThemes(text string) []string {
	themes := make([]string, 0)
	
	// 主题词汇
	themeWords := []string{
		"爱情", "友情", "亲情", "成长", "梦想", "希望", "绝望",
		"勇气", "恐惧", "孤独", "团结", "背叛", "忠诚", "自由",
		"正义", "邪恶", "善良", "复仇", "宽恕", "牺牲", "拯救",
	}
	
	for _, theme := range themeWords {
		if strings.Contains(text, theme) {
			themes = append(themes, theme)
		}
	}
	
	return themes
}

// extractSymbolism 提取象征意义
func (sa *SemanticAnalyzer) extractSymbolism(text string) []string {
	symbolism := make([]string, 0)
	
	// 象征性词汇
	symbolWords := map[string]string{
		"鸽子":   "和平",
		"玫瑰":   "爱情",
		"十字架": "信仰",
		"彩虹":   "希望",
		"黑暗":   "绝望",
		"光明":   "希望",
		"桥":    "连接",
		"墙":    "阻隔",
	}
	
	for symbol, meaning := range symbolWords {
		if strings.Contains(text, symbol) {
			symbolism = append(symbolism, fmt.Sprintf("%s象征%s", symbol, meaning))
		}
	}
	
	return symbolism
}

// parseSemanticResponse 解析语义响应
func (sa *SemanticAnalyzer) parseSemanticResponse(response, originalText string) *SemanticContext {
	// 简化的响应解析，实际应该更复杂
	context := sa.ruleBasedSemanticAnalysis(originalText)
	
	// 尝试从AI响应中提取更多信息
	lines := strings.Split(response, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "主要动作") {
			// 提取主要动作
			parts := strings.Split(line, ":")
			if len(parts) > 1 {
				context.MainAction = strings.TrimSpace(parts[1])
			}
		}
	}
	
	return context
}

// ContextTracker 上下文跟踪器
type ContextTracker struct {
	// 可以添加状态跟踪
}

// NewContextTracker 创建上下文跟踪器
func NewContextTracker() *ContextTracker {
	return &ContextTracker{}
}

// ContextFlow 上下文流
type ContextFlow struct {
	Scenes           []*EnhancedScene      `json:"scenes"`
	Transitions      []*ContextTransition  `json:"transitions"`
	CharacterArcs    []*CharacterArc       `json:"character_arcs"`
	ThematicFlow     *ThematicFlow         `json:"thematic_flow"`
	EmotionalCurve   *EmotionalCurve       `json:"emotional_curve"`
	NarrativeStructure *NarrativeStructure `json:"narrative_structure"`
}

// ContextTransition 上下文过渡
type ContextTransition struct {
	FromScene    int                    `json:"from_scene"`
	ToScene      int                    `json:"to_scene"`
	TransitionType string               `json:"transition_type"`
	Continuity   map[string]interface{} `json:"continuity"`
	Contrast     map[string]interface{} `json:"contrast"`
}

// CharacterArc 角色弧线
type CharacterArc struct {
	CharacterName string                 `json:"character_name"`
	Appearances   []int                  `json:"appearances"`
	Development   map[string]interface{} `json:"development"`
	EmotionalJourney []string            `json:"emotional_journey"`
}

// ThematicFlow 主题流
type ThematicFlow struct {
	MainThemes    []string               `json:"main_themes"`
	ThemeEvolution map[string][]float64  `json:"theme_evolution"`
	Symbolism     map[string]string      `json:"symbolism"`
}

// EmotionalCurve 情感曲线
type EmotionalCurve struct {
	IntensityPoints []float64 `json:"intensity_points"`
	EmotionTypes    []string  `json:"emotion_types"`
	TensionCurve    []float64 `json:"tension_curve"`
	ClimaxPoints    []int     `json:"climax_points"`
}

// NarrativeStructure 叙事结构
type NarrativeStructure struct {
	Act1Scenes []int    `json:"act1_scenes"` // 第一幕场景
	Act2Scenes []int    `json:"act2_scenes"` // 第二幕场景
	Act3Scenes []int    `json:"act3_scenes"` // 第三幕场景
	Climax     int      `json:"climax"`      // 高潮场景
	Resolution int      `json:"resolution"`  // 结局场景
	PlotPoints []string `json:"plot_points"` // 情节点
}

// TrackContextFlow 跟踪上下文流
func (ct *ContextTracker) TrackContextFlow(ctx context.Context, scenes []*EnhancedScene) (*ContextFlow, error) {
	log.Printf("[ContextTracker] 跟踪上下文流: %d个场景", len(scenes))

	flow := &ContextFlow{
		Scenes:      scenes,
		Transitions: ct.analyzeTransitions(scenes),
		CharacterArcs: ct.analyzeCharacterArcs(scenes),
		ThematicFlow: ct.analyzeThematicFlow(scenes),
		EmotionalCurve: ct.analyzeEmotionalCurve(scenes),
		NarrativeStructure: ct.analyzeNarrativeStructure(scenes),
	}

	return flow, nil
}

// analyzeTransitions 分析过渡
func (ct *ContextTracker) analyzeTransitions(scenes []*EnhancedScene) []*ContextTransition {
	transitions := make([]*ContextTransition, 0)

	for i := 0; i < len(scenes)-1; i++ {
		transition := &ContextTransition{
			FromScene:      i,
			ToScene:        i + 1,
			TransitionType: ct.determineTransitionType(scenes[i], scenes[i+1]),
			Continuity:     ct.analyzeContinuity(scenes[i], scenes[i+1]),
			Contrast:       ct.analyzeContrast(scenes[i], scenes[i+1]),
		}
		transitions = append(transitions, transition)
	}

	return transitions
}

// determineTransitionType 确定过渡类型
func (ct *ContextTracker) determineTransitionType(scene1, scene2 *EnhancedScene) string {
	// 基于位置变化
	if scene1.Location != scene2.Location {
		return "location_change"
	}
	
	// 基于时间变化
	if scene1.TimeOfDay != scene2.TimeOfDay {
		return "time_change"
	}
	
	// 基于情感变化
	if scene1.EmotionalTone.PrimaryEmotion != scene2.EmotionalTone.PrimaryEmotion {
		return "emotional_shift"
	}
	
	return "continuous"
}

// analyzeContinuity 分析连续性
func (ct *ContextTracker) analyzeContinuity(scene1, scene2 *EnhancedScene) map[string]interface{} {
	continuity := make(map[string]interface{})
	
	// 角色连续性（暂时跳过，因为Scene实体没有Characters字段）
	continuity["characters"] = []string{}
	
	// 位置连续性
	continuity["location_similarity"] = ct.calculateLocationSimilarity(scene1.Location, scene2.Location)
	
	// 主题连续性
	continuity["theme_overlap"] = ct.calculateThemeOverlap(scene1.SemanticContext.Themes, scene2.SemanticContext.Themes)
	
	return continuity
}

// analyzeContrast 分析对比
func (ct *ContextTracker) analyzeContrast(scene1, scene2 *EnhancedScene) map[string]interface{} {
	contrast := make(map[string]interface{})
	
	// 情感对比
	contrast["emotional_contrast"] = math.Abs(scene1.EmotionalTone.Intensity - scene2.EmotionalTone.Intensity)
	
	// 节奏对比
	contrast["pace_contrast"] = ct.calculatePaceContrast(scene1, scene2)
	
	// 视觉对比
	contrast["visual_contrast"] = ct.calculateVisualContrast(scene1.VisualElements, scene2.VisualElements)
	
	return contrast
}

// analyzeCharacterArcs 分析角色弧线
func (ct *ContextTracker) analyzeCharacterArcs(scenes []*EnhancedScene) []*CharacterArc {
	characterMap := make(map[string]*CharacterArc)
	
	for i, scene := range scenes {
		// 暂时跳过角色分析，因为Scene实体没有Characters字段
		_ = scene
		_ = i
	}

	arcs := make([]*CharacterArc, 0, len(characterMap))
	for _, arc := range characterMap {
		arcs = append(arcs, arc)
	}
	
	return arcs
}

// analyzeThematicFlow 分析主题流
func (ct *ContextTracker) analyzeThematicFlow(scenes []*EnhancedScene) *ThematicFlow {
	themeMap := make(map[string][]float64)
	allThemes := make(map[string]bool)
	
	// 收集所有主题
	for _, scene := range scenes {
		for _, theme := range scene.SemanticContext.Themes {
			allThemes[theme] = true
		}
	}
	
	// 计算每个主题在各场景中的强度
	for theme := range allThemes {
		intensities := make([]float64, len(scenes))
		for i, scene := range scenes {
			intensity := 0.0
			for _, sceneTheme := range scene.SemanticContext.Themes {
				if sceneTheme == theme {
					intensity = 1.0
					break
				}
			}
			intensities[i] = intensity
		}
		themeMap[theme] = intensities
	}
	
	mainThemes := make([]string, 0, len(allThemes))
	for theme := range allThemes {
		mainThemes = append(mainThemes, theme)
	}
	
	return &ThematicFlow{
		MainThemes:     mainThemes,
		ThemeEvolution: themeMap,
		Symbolism:      make(map[string]string),
	}
}

// analyzeEmotionalCurve 分析情感曲线
func (ct *ContextTracker) analyzeEmotionalCurve(scenes []*EnhancedScene) *EmotionalCurve {
	intensities := make([]float64, len(scenes))
	emotions := make([]string, len(scenes))
	tensions := make([]float64, len(scenes))
	climaxPoints := make([]int, 0)
	
	for i, scene := range scenes {
		intensities[i] = scene.EmotionalTone.Intensity
		emotions[i] = scene.EmotionalTone.PrimaryEmotion
		tensions[i] = scene.EmotionalTone.TensionLevel
		
		// 识别高潮点
		if scene.EmotionalTone.Intensity > 0.8 {
			climaxPoints = append(climaxPoints, i)
		}
	}
	
	return &EmotionalCurve{
		IntensityPoints: intensities,
		EmotionTypes:    emotions,
		TensionCurve:    tensions,
		ClimaxPoints:    climaxPoints,
	}
}

// analyzeNarrativeStructure 分析叙事结构
func (ct *ContextTracker) analyzeNarrativeStructure(scenes []*EnhancedScene) *NarrativeStructure {
	totalScenes := len(scenes)
	
	// 简化的三幕结构划分
	act1End := totalScenes / 4
	act2End := totalScenes * 3 / 4
	
	act1Scenes := make([]int, 0)
	act2Scenes := make([]int, 0)
	act3Scenes := make([]int, 0)
	
	for i := 0; i < totalScenes; i++ {
		if i < act1End {
			act1Scenes = append(act1Scenes, i)
		} else if i < act2End {
			act2Scenes = append(act2Scenes, i)
		} else {
			act3Scenes = append(act3Scenes, i)
		}
	}
	
	// 找到高潮场景（情感强度最高的场景）
	climax := 0
	maxIntensity := 0.0
	for i, scene := range scenes {
		if scene.EmotionalTone.Intensity > maxIntensity {
			maxIntensity = scene.EmotionalTone.Intensity
			climax = i
		}
	}
	
	return &NarrativeStructure{
		Act1Scenes: act1Scenes,
		Act2Scenes: act2Scenes,
		Act3Scenes: act3Scenes,
		Climax:     climax,
		Resolution: totalScenes - 1,
		PlotPoints: []string{"开端", "发展", "高潮", "结局"},
	}
}

// 辅助方法
func (ct *ContextTracker) findCommonCharacters(chars1, chars2 []string) []string {
	common := make([]string, 0)
	for _, char1 := range chars1 {
		for _, char2 := range chars2 {
			if char1 == char2 {
				common = append(common, char1)
				break
			}
		}
	}
	return common
}

func (ct *ContextTracker) calculateLocationSimilarity(loc1, loc2 string) float64 {
	if loc1 == loc2 {
		return 1.0
	}
	// 简化的相似度计算
	return 0.0
}

func (ct *ContextTracker) calculateThemeOverlap(themes1, themes2 []string) float64 {
	if len(themes1) == 0 && len(themes2) == 0 {
		return 1.0
	}
	
	common := 0
	for _, theme1 := range themes1 {
		for _, theme2 := range themes2 {
			if theme1 == theme2 {
				common++
				break
			}
		}
	}
	
	total := len(themes1) + len(themes2) - common
	if total == 0 {
		return 1.0
	}
	
	return float64(common) / float64(total)
}

func (ct *ContextTracker) calculatePaceContrast(scene1, scene2 *EnhancedScene) float64 {
	// 简化的节奏对比计算
	return 0.5
}

func (ct *ContextTracker) calculateVisualContrast(visual1, visual2 *VisualElements) float64 {
	// 简化的视觉对比计算
	return 0.5
}
