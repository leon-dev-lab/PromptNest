package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

const (
	CommonCategory   = "通用分类"
	FallbackCategory = "未分类"
)

// CategoryPalette is deliberately finite: PromptNest exposes these 20 curated
// colours in the UI so category colour choices stay consistent across Windows
// and macOS and remain readable in both light and dark themes.
var CategoryPalette = []string{
	"#D96C4B", "#6F8B5B", "#7F6FB2", "#4D8C9C", "#C29342",
	"#5F78B5", "#B76479", "#4F9A7B", "#D9895B", "#839750",
	"#5D91B8", "#B77B45", "#8B7665", "#6F9AA4", "#A66C9B",
	"#67A18D", "#E08A72", "#9A8B52", "#6A87A9", "#AA745C",
}

type storeSettings struct {
	DataDirectory string `json:"dataDirectory"`
}

type DataStore struct {
	DefaultDataDirectory string
	ConfigDirectory      string
	ConfigFilePath       string
	DataDirectory        string
	DataFilePath         string
}

func NewDataStore() *DataStore {
	defaultDir := defaultPromptNestDirectory()
	store := &DataStore{
		DefaultDataDirectory: defaultDir,
		ConfigDirectory:      defaultDir,
		ConfigFilePath:       filepath.Join(defaultDir, "settings.json"),
		DataDirectory:        defaultDir,
		DataFilePath:         filepath.Join(defaultDir, "prompts.json"),
	}
	store.loadSettings()
	return store
}

func defaultPromptNestDirectory() string {
	if runtime.GOOS == "windows" {
		if base := strings.TrimSpace(os.Getenv("LOCALAPPDATA")); base != "" {
			return filepath.Join(base, "PromptNest")
		}
	}
	if base, err := os.UserConfigDir(); err == nil && strings.TrimSpace(base) != "" {
		return filepath.Join(base, "PromptNest")
	}
	if base, err := os.UserHomeDir(); err == nil && strings.TrimSpace(base) != "" {
		return filepath.Join(base, ".promptnest")
	}
	return filepath.Join(".", "PromptNestData")
}

func (s *DataStore) loadSettings() {
	_ = os.MkdirAll(s.ConfigDirectory, 0o755)
	data, err := os.ReadFile(s.ConfigFilePath)
	if err != nil {
		return
	}
	var settings storeSettings
	if json.Unmarshal(data, &settings) != nil {
		return
	}
	dir := strings.TrimSpace(settings.DataDirectory)
	if dir == "" {
		return
	}
	if abs, err := filepath.Abs(dir); err == nil {
		dir = abs
	}
	s.DataDirectory = filepath.Clean(dir)
	s.DataFilePath = filepath.Join(s.DataDirectory, "prompts.json")
}

func (s *DataStore) saveSettings() error {
	if err := os.MkdirAll(s.ConfigDirectory, 0o755); err != nil {
		return err
	}
	settings := storeSettings{}
	if !samePath(s.DataDirectory, s.DefaultDataDirectory) {
		settings.DataDirectory = s.DataDirectory
	}
	data, err := json.MarshalIndent(settings, "", "  ")
	if err != nil {
		return err
	}
	tmp := s.ConfigFilePath + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	_ = os.Remove(s.ConfigFilePath)
	return os.Rename(tmp, s.ConfigFilePath)
}

func (s *DataStore) SetDataDirectory(dir string) error {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return errors.New("数据目录不能为空")
	}
	abs, err := filepath.Abs(dir)
	if err == nil {
		dir = abs
	}
	dir = filepath.Clean(dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("无法创建数据目录: %w", err)
	}
	oldDir, oldFile := s.DataDirectory, s.DataFilePath
	s.DataDirectory = dir
	s.DataFilePath = filepath.Join(dir, "prompts.json")
	if err := s.saveSettings(); err != nil {
		s.DataDirectory, s.DataFilePath = oldDir, oldFile
		return err
	}
	return nil
}

func (s *DataStore) ResetDataDirectory() error {
	return s.SetDataDirectory(s.DefaultDataDirectory)
}

func (s *DataStore) IsCustomDataDirectory() bool {
	return !samePath(s.DataDirectory, s.DefaultDataDirectory)
}

func (s *DataStore) DataFileExistsInDirectory(dir string) bool {
	_, err := os.Stat(filepath.Join(filepath.Clean(dir), "prompts.json"))
	return err == nil
}

func (s *DataStore) LoadFromDirectory(dir string) (AppState, error) {
	path := filepath.Join(filepath.Clean(dir), "prompts.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return AppState{}, err
	}
	var state AppState
	if err := json.Unmarshal(data, &state); err != nil {
		return AppState{}, fmt.Errorf("无法解析目标目录中的 prompts.json: %w", err)
	}
	normalizeState(&state)
	return state, nil
}

func (s *DataStore) SaveToDirectory(state AppState, dir string) error {
	dir = filepath.Clean(dir)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	normalizeState(&state)
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	path := filepath.Join(dir, "prompts.json")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	_ = os.Remove(path)
	return os.Rename(tmp, path)
}

func samePath(a, b string) bool {
	aa, errA := filepath.Abs(filepath.Clean(a))
	bb, errB := filepath.Abs(filepath.Clean(b))
	if errA == nil {
		a = aa
	}
	if errB == nil {
		b = bb
	}
	if runtime.GOOS == "windows" {
		return strings.EqualFold(filepath.Clean(a), filepath.Clean(b))
	}
	return filepath.Clean(a) == filepath.Clean(b)
}

func (s *DataStore) Load() AppState {
	if err := os.MkdirAll(s.DataDirectory, 0o755); err != nil {
		// If a custom path is temporarily unavailable, fall back to the standard
		// app data directory instead of silently losing the user's library.
		if s.IsCustomDataDirectory() {
			s.DataDirectory = s.DefaultDataDirectory
			s.DataFilePath = filepath.Join(s.DefaultDataDirectory, "prompts.json")
			_ = s.saveSettings()
			_ = os.MkdirAll(s.DataDirectory, 0o755)
		}
	}
	data, err := os.ReadFile(s.DataFilePath)
	if errors.Is(err, os.ErrNotExist) {
		state := createDefaultState()
		_ = s.Save(&state)
		return state
	}
	if err != nil {
		return createDefaultState()
	}

	var state AppState
	if err := json.Unmarshal(data, &state); err != nil {
		backup := filepath.Join(s.DataDirectory, fmt.Sprintf("prompts.corrupt.%s.json", time.Now().Format("20060102150405")))
		_ = os.WriteFile(backup, data, 0o644)
		return createDefaultState()
	}
	normalizeState(&state)
	return state
}

func (s *DataStore) Save(state *AppState) error {
	normalizeState(state)
	return s.SaveToDirectory(*state, s.DataDirectory)
}

func (s *DataStore) Export(state AppState, path string) error {
	normalizeState(&state)
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o644)
}

func (s *DataStore) Import(path string) (AppState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return AppState{}, err
	}
	var state AppState
	if err := json.Unmarshal(data, &state); err != nil {
		return AppState{}, fmt.Errorf("无法解析该提示词库文件: %w", err)
	}
	normalizeState(&state)
	return state, nil
}

func normalizeState(state *AppState) {
	if state.Categories == nil {
		state.Categories = []string{}
	}
	if state.Tags == nil {
		state.Tags = []string{}
	}
	if state.CategoryColors == nil {
		state.CategoryColors = map[string]string{}
	}
	if state.Prompts == nil {
		state.Prompts = []PromptItem{}
	}

	legacyOrder := len(state.Prompts) > 1
	if legacyOrder {
		for _, p := range state.Prompts {
			if p.SortOrder != 0 {
				legacyOrder = false
				break
			}
		}
	}

	for i := range state.Prompts {
		p := &state.Prompts[i]
		if strings.TrimSpace(p.ID) == "" {
			p.ID = newID()
		}
		p.Title = strings.TrimSpace(p.Title)
		if p.Title == "" {
			p.Title = "未命名提示词"
		}
		p.Category = strings.TrimSpace(p.Category)
		if p.Category == "" {
			p.Category = FallbackCategory
		}
		p.Tags = strings.TrimSpace(p.Tags)
		if p.CreatedAt.IsZero() {
			p.CreatedAt = time.Now()
		}
		if p.UpdatedAt.IsZero() {
			p.UpdatedAt = p.CreatedAt
		}
		if p.SortOrder < 0 || legacyOrder {
			p.SortOrder = i
		}
	}

	sort.SliceStable(state.Prompts, func(i, j int) bool {
		return state.Prompts[i].SortOrder < state.Prompts[j].SortOrder
	})
	for i := range state.Prompts {
		state.Prompts[i].SortOrder = i
	}

	cats := make([]string, 0, len(state.Categories)+len(state.Prompts))
	cats = append(cats, state.Categories...)
	for _, p := range state.Prompts {
		cats = append(cats, p.Category)
	}
	state.Categories = distinctPreserveOrder(cats)
	normalizeCategoryColors(state)

	tags := append([]string{}, state.Tags...)
	for _, p := range state.Prompts {
		tags = append(tags, splitTags(p.Tags)...)
	}
	state.Tags = distinctPreserveOrder(tags)
}

func normalizeCategoryColors(state *AppState) {
	existing := state.CategoryColors
	normalized := make(map[string]string, len(state.Categories))
	used := map[string]bool{}

	// Preserve valid existing choices, including legacy maps whose key casing no
	// longer exactly matches the category name.
	for _, category := range state.Categories {
		if color := categoryColorFromMap(existing, category); isCategoryPaletteColor(color) {
			normalized[category] = canonicalPaletteColor(color)
			used[strings.ToLower(normalized[category])] = true
		}
	}

	// Old libraries had no explicit colour metadata. Give each category the first
	// unused preset so adjacent/new categories do not unexpectedly share a colour.
	for _, category := range state.Categories {
		if normalized[category] != "" {
			continue
		}
		color := nextUnusedCategoryColor(used, len(normalized))
		normalized[category] = color
		used[strings.ToLower(color)] = true
	}
	state.CategoryColors = normalized
}

func categoryColorFromMap(colors map[string]string, category string) string {
	for key, value := range colors {
		if strings.EqualFold(strings.TrimSpace(key), strings.TrimSpace(category)) {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func canonicalPaletteColor(color string) string {
	for _, candidate := range CategoryPalette {
		if strings.EqualFold(strings.TrimSpace(candidate), strings.TrimSpace(color)) {
			return candidate
		}
	}
	return ""
}

func isCategoryPaletteColor(color string) bool {
	return canonicalPaletteColor(color) != ""
}

func nextUnusedCategoryColor(used map[string]bool, offset int) string {
	for _, color := range CategoryPalette {
		if !used[strings.ToLower(color)] {
			return color
		}
	}
	if len(CategoryPalette) == 0 {
		return "#6F8B5B"
	}
	return CategoryPalette[offset%len(CategoryPalette)]
}

func setCategoryColor(state *AppState, category, color string) {
	category = strings.TrimSpace(category)
	if category == "" {
		return
	}
	if state.CategoryColors == nil {
		state.CategoryColors = map[string]string{}
	}
	for key := range state.CategoryColors {
		if strings.EqualFold(key, category) && key != category {
			delete(state.CategoryColors, key)
		}
	}
	if canonical := canonicalPaletteColor(color); canonical != "" {
		state.CategoryColors[category] = canonical
	}
}

func ensureCategoryColor(state *AppState, category string) {
	if categoryColorFromMap(state.CategoryColors, category) != "" {
		return
	}
	used := map[string]bool{}
	for _, value := range state.CategoryColors {
		if canonical := canonicalPaletteColor(value); canonical != "" {
			used[strings.ToLower(canonical)] = true
		}
	}
	setCategoryColor(state, category, nextUnusedCategoryColor(used, len(state.CategoryColors)))
}

func deleteCategoryColor(state *AppState, category string) {
	for key := range state.CategoryColors {
		if strings.EqualFold(key, category) {
			delete(state.CategoryColors, key)
		}
	}
}

func renameCategoryColor(state *AppState, oldName, newName string) {
	color := categoryColorFromMap(state.CategoryColors, oldName)
	deleteCategoryColor(state, oldName)
	if color != "" {
		setCategoryColor(state, newName, color)
	} else {
		ensureCategoryColor(state, newName)
	}
}

func distinctPreserveOrder(values []string) []string {
	seen := map[string]bool{}
	result := make([]string, 0, len(values))
	for _, raw := range values {
		v := strings.TrimSpace(raw)
		if v == "" {
			continue
		}
		key := strings.ToLower(v)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, v)
	}
	return result
}

func splitTags(value string) []string {
	parts := strings.FieldsFunc(value, func(r rune) bool { return r == ',' || r == '，' })
	return distinctPreserveOrder(parts)
}

func newID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	h := hex.EncodeToString(b[:])
	return h[0:8] + "-" + h[8:12] + "-" + h[12:16] + "-" + h[16:20] + "-" + h[20:32]
}

func createDefaultState() AppState {
	now := time.Now()
	return AppState{
		IsDarkTheme:    false,
		Categories:     []string{CommonCategory, "写作", "开发", "办公", "产品", "分析", "图像"},
		CategoryColors: map[string]string{},
		Tags: []string{
			"结构化", "润色", "中文", "风格", "代码审查", "Go", "Bug", "会议", "总结", "Action Items",
			"PRD", "需求", "验收标准", "数据分析", "指标", "洞察", "图像理解", "标题", "文件命名",
		},
		Prompts: []PromptItem{
			{ID: newID(), Title: "结构化任务助手", Category: CommonCategory, Tags: "结构化", Content: "请先理解我的目标，再把任务整理成清晰、可执行的结果。\n\n要求：\n1. 先提炼目标与约束；\n2. 信息不足时明确指出，不要猜测；\n3. 输出结构清楚，优先给可直接执行的内容；\n4. 如有风险或前置条件，请单独列出。\n\n任务：\n{{task}}", Notes: "通用示例，可用于把零散需求整理成结构化输出。", IsFavorite: true, SortOrder: 0, CreatedAt: now.Add(-7 * 24 * time.Hour), UpdatedAt: now.Add(-2 * 24 * time.Hour), UseCount: 6},
			{ID: newID(), Title: "文章润色", Category: "写作", Tags: "润色, 中文, 风格", Content: "你是一名资深中文编辑。请在不改变原意和事实信息的前提下，润色下面的文本。\n\n要求：\n1. 语言自然、简洁，避免模板化表达；\n2. 优化段落衔接和逻辑；\n3. 保留原有专业术语；\n4. 不要虚构任何信息；\n5. 直接输出润色后的正文。\n\n原文：\n{{text}}", Notes: "适合邮件、说明文、博客初稿。", IsFavorite: true, SortOrder: 1, CreatedAt: now.Add(-6 * 24 * time.Hour), UpdatedAt: now.Add(-2 * 24 * time.Hour), UseCount: 8},
			{ID: newID(), Title: "代码审查助手", Category: "开发", Tags: "代码审查, Go, Bug", Content: "请作为高级软件工程师审查以下代码。重点检查：\n- 潜在 Bug 与边界条件\n- 可维护性与可读性\n- 性能问题\n- 并发与资源释放\n- 安全风险\n\n请按“问题 / 影响 / 修改建议”的结构输出，并给出必要的修正版代码。\n\n代码：\n{{code}}", Notes: "示例以 Go 为标签，也可以按你的技术栈修改。", IsFavorite: true, SortOrder: 2, CreatedAt: now.Add(-5 * 24 * time.Hour), UpdatedAt: now.Add(-11 * time.Hour), UseCount: 15},
			{ID: newID(), Title: "会议纪要整理", Category: "办公", Tags: "会议, 总结, Action Items", Content: "把下面的会议记录整理成清晰的会议纪要，输出：\n1. 会议目标\n2. 关键结论\n3. 已做决策\n4. 待办事项（负责人 / 截止时间 / 状态）\n5. 未解决问题\n\n如果原文没有负责人或截止时间，请明确标记“待确认”，不要猜测。\n\n会议记录：\n{{notes}}", Notes: "适合从语音转写稿中提炼行动项。", SortOrder: 3, CreatedAt: now.Add(-4 * 24 * time.Hour), UpdatedAt: now.Add(-24 * time.Hour), UseCount: 4},
			{ID: newID(), Title: "需求拆解", Category: "产品", Tags: "PRD, 需求, 验收标准", Content: "你是一名产品经理。根据下面的需求描述，拆解为：\n- 背景与目标\n- 用户角色\n- 核心用户故事\n- 功能范围 / 非范围\n- 关键流程\n- 异常与边界情况\n- 验收标准\n- 风险与待确认问题\n\n对于信息不足的部分，不要自行补全，放到“待确认问题”。\n\n需求：\n{{requirement}}", Notes: "用于把零散想法整理成可讨论的需求结构。", SortOrder: 4, CreatedAt: now.Add(-3 * 24 * time.Hour), UpdatedAt: now.Add(-5 * time.Hour), UseCount: 3},
			{ID: newID(), Title: "表格数据洞察", Category: "分析", Tags: "数据分析, 指标, 洞察", Content: "请分析我提供的数据，先检查数据口径与缺失值，再回答：\n1. 最重要的 3-5 个发现\n2. 异常点与可能原因\n3. 指标之间可能存在的关系\n4. 需要进一步验证的假设\n5. 下一步建议\n\n将“数据事实”和“推断”明确区分，不能把相关性写成因果关系。", Notes: "适合 CSV / Excel 分析前的统一分析框架。", SortOrder: 5, CreatedAt: now.Add(-2 * 24 * time.Hour), UpdatedAt: now.Add(-2 * time.Hour), UseCount: 2},
			{ID: newID(), Title: "图片内容命名", Category: "图像", Tags: "图像理解, 标题, 文件命名", Content: "根据提供的图片内容生成一个准确、自然、适合作为文件名的英文标题。\n\n要求：\n1. 只描述图片中真正可见的主体、场景、风格、材质或用途；\n2. 不猜测品牌、人物身份或不可见信息；\n3. 标题控制在 4-14 个英文单词；\n4. 不要引号、编号、解释或文件扩展名；\n5. 只输出最终标题。", Notes: "适合多模态模型做图片理解和批量命名。", SortOrder: 6, CreatedAt: now.Add(-24 * time.Hour), UpdatedAt: now.Add(-90 * time.Minute), UseCount: 1},
		},
	}
}
