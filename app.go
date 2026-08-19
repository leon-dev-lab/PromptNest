package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	wailsrt "github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx                  context.Context
	store                *DataStore
	mu                   sync.Mutex
	state                AppState
	pendingImport        *AppState
	pendingDataDirectory string
	tray                 *TrayManager
	forceQuit            atomic.Bool
}

func NewApp() *App {
	store := NewDataStore()
	return &App{store: store, state: store.Load()}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.startTray()
}

func (a *App) shutdown(ctx context.Context) {
	a.mu.Lock()
	_ = a.store.Save(&a.state)
	a.mu.Unlock()
	if a.tray != nil {
		a.tray.Stop()
	}
}

func (a *App) beforeClose(ctx context.Context) bool {
	if a.forceQuit.Load() {
		return false
	}
	wailsrt.WindowHide(ctx)
	if a.tray != nil {
		a.tray.Notify("PromptNest 仍在运行", "窗口已隐藏到系统托盘。双击托盘图标即可恢复。")
	}
	return true
}

func (a *App) showWindow() {
	if a.ctx == nil {
		return
	}
	wailsrt.WindowShow(a.ctx)
}

func (a *App) quitFromTray() {
	a.forceQuit.Store(true)
	if a.ctx != nil {
		wailsrt.Quit(a.ctx)
	}
}

func (a *App) HandleMessage(raw string) (string, error) {
	var msg BridgeMessage
	if err := json.Unmarshal([]byte(raw), &msg); err != nil {
		return a.encodeMessages(a.toast("消息格式错误: "+err.Error(), "error"))
	}

	var messages []BridgeEnvelope
	var err error

	switch msg.Type {
	case "ready":
		messages = []BridgeEnvelope{a.stateMessage("ready", "")}
	case "savePrompt":
		var p PromptItem
		p, err = a.savePrompt(msg.Payload)
		if err == nil {
			messages = []BridgeEnvelope{a.stateMessage("saved", p.ID)}
		}
	case "autoSavePrompt":
		var p PromptItem
		p, err = a.savePrompt(msg.Payload)
		if err == nil {
			messages = []BridgeEnvelope{{Type: "autoSaveAck", Payload: map[string]any{"id": p.ID, "updatedAt": p.UpdatedAt}}}
		}
	case "deletePrompt":
		err = a.deletePrompt(msg.Payload)
		if err == nil {
			messages = []BridgeEnvelope{a.stateMessage("deleted", "")}
		}
	case "duplicatePrompt":
		var id string
		id, err = a.duplicatePrompt(msg.Payload)
		if err == nil {
			messages = []BridgeEnvelope{a.stateMessage("duplicated", id)}
		}
	case "toggleFavorite":
		err = a.toggleFavorite(msg.Payload)
		if err == nil {
			messages = []BridgeEnvelope{a.stateMessage("favoriteChanged", "")}
		}
	case "copyPrompt":
		err = a.copyPrompt(msg.Payload)
		if err == nil {
			messages = []BridgeEnvelope{a.stateMessage("copied", "")}
		}
	case "createPrompt":
		var id string
		id, err = a.createPrompt()
		if err == nil {
			messages = []BridgeEnvelope{a.stateMessage("created", id)}
		}
	case "reorderPrompts":
		err = a.reorderPrompts(msg.Payload)
		if err == nil {
			messages = []BridgeEnvelope{a.stateMessage("promptsReordered", "")}
		}
	case "addCategory":
		err = a.addCategory(msg.Payload)
		if err == nil {
			messages = []BridgeEnvelope{a.stateMessage("categoryAdded", "")}
		}
	case "deleteCategory":
		err = a.deleteCategory(msg.Payload)
		if err == nil {
			messages = []BridgeEnvelope{a.stateMessage("categoryDeleted", "")}
		}
	case "renameCategory":
		err = a.renameCategory(msg.Payload)
		if err == nil {
			messages = []BridgeEnvelope{a.stateMessage("categoryRenamed", "")}
		}
	case "reorderCategories":
		err = a.reorderCategories(msg.Payload)
		if err == nil {
			messages = []BridgeEnvelope{a.stateMessage("categoriesReordered", "")}
		}
	case "setCategoryColor":
		err = a.setCategoryColor(msg.Payload)
		if err == nil {
			messages = []BridgeEnvelope{a.stateMessage("categoryColorChanged", "")}
		}
	case "addTag":
		err = a.addTag(msg.Payload)
		if err == nil {
			messages = []BridgeEnvelope{a.stateMessage("tagAdded", "")}
		}
	case "deleteTag":
		err = a.deleteTag(msg.Payload)
		if err == nil {
			messages = []BridgeEnvelope{a.stateMessage("tagDeleted", "")}
		}
	case "renameTag":
		err = a.renameTag(msg.Payload)
		if err == nil {
			messages = []BridgeEnvelope{a.stateMessage("tagRenamed", "")}
		}
	case "reorderTags":
		err = a.reorderTags(msg.Payload)
		if err == nil {
			messages = []BridgeEnvelope{a.stateMessage("tagsReordered", "")}
		}
	case "setTheme":
		err = a.setTheme(msg.Payload)
		if err == nil {
			messages = []BridgeEnvelope{a.stateMessage("themeChanged", "")}
		}
	case "export":
		messages, err = a.exportLibrary()
	case "import":
		messages, err = a.importLibrary()
	case "confirmImport":
		messages, err = a.confirmImportLibrary()
	case "cancelImport":
		a.cancelImportLibrary()
	case "openDataFolder":
		err = a.openDataFolder()
	case "openGitHub":
		if a.ctx != nil {
			wailsrt.BrowserOpenURL(a.ctx, "https://github.com/leon-dev-lab")
		}
	case "chooseDataDirectory":
		messages, err = a.chooseDataDirectory()
	case "setDataDirectory":
		messages, err = a.switchDataDirectory(readString(msg.Payload, "path"))
	case "confirmDataDirectory":
		messages, err = a.confirmDataDirectorySwitch()
	case "cancelDataDirectory":
		a.cancelDataDirectorySwitch()
	case "resetDataDirectory":
		messages, err = a.resetDataDirectory()
	case "hideToTray":
		if a.ctx != nil {
			wailsrt.WindowHide(a.ctx)
		}
	default:
		err = fmt.Errorf("未知操作: %s", msg.Type)
	}

	if err != nil {
		messages = a.toast(err.Error(), "error")
	}
	return a.encodeMessages(messages)
}

func (a *App) encodeMessages(messages ...[]BridgeEnvelope) (string, error) {
	var flat []BridgeEnvelope
	for _, group := range messages {
		flat = append(flat, group...)
	}
	if flat == nil {
		flat = []BridgeEnvelope{}
	}
	data, err := json.Marshal(flat)
	return string(data), err
}

func (a *App) toast(text, kind string) []BridgeEnvelope {
	return []BridgeEnvelope{{Type: "toast", Payload: map[string]any{"text": text, "kind": kind}}}
}

func (a *App) stateMessage(reason, focusID string) BridgeEnvelope {
	a.mu.Lock()
	defer a.mu.Unlock()
	return BridgeEnvelope{Type: "state", Payload: map[string]any{
		"reason":               reason,
		"focusId":              focusID,
		"state":                cloneState(a.state),
		"dataPath":             a.store.DataFilePath,
		"dataDirectory":        a.store.DataDirectory,
		"defaultDataDirectory": a.store.DefaultDataDirectory,
		"customDataDirectory":  a.store.IsCustomDataDirectory(),
		"categoryPalette":      append([]string{}, CategoryPalette...),
	}}
}

func cloneState(state AppState) AppState {
	copyState := state
	copyState.Categories = append([]string{}, state.Categories...)
	copyState.CategoryColors = make(map[string]string, len(state.CategoryColors))
	for key, value := range state.CategoryColors {
		copyState.CategoryColors[key] = value
	}
	copyState.Tags = append([]string{}, state.Tags...)
	copyState.Prompts = append([]PromptItem{}, state.Prompts...)
	return copyState
}

func (a *App) savePrompt(payload json.RawMessage) (PromptItem, error) {
	var body struct {
		Prompt PromptItem `json:"prompt"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return PromptItem{}, errors.New("无法读取提示词数据")
	}
	p := body.Prompt
	p.Title = strings.TrimSpace(p.Title)
	if p.Title == "" {
		p.Title = "未命名提示词"
	}
	p.Category = strings.TrimSpace(p.Category)
	p.Tags = strings.TrimSpace(p.Tags)
	p.UpdatedAt = time.Now()

	a.mu.Lock()
	defer a.mu.Unlock()

	if p.Category == "" {
		p.Category = a.defaultCategoryLocked()
	}
	idx := a.promptIndexLocked(p.ID)
	if idx >= 0 {
		existing := a.state.Prompts[idx]
		p.CreatedAt = existing.CreatedAt
		p.UseCount = existing.UseCount
		p.LastUsedAt = existing.LastUsedAt
		p.SortOrder = existing.SortOrder
		a.state.Prompts[idx] = p
	} else {
		if strings.TrimSpace(p.ID) == "" {
			p.ID = newID()
		}
		p.CreatedAt = time.Now()
		p.SortOrder = a.nextSortOrderLocked()
		a.state.Prompts = append(a.state.Prompts, p)
	}
	a.ensureCategoryLocked(p.Category)
	a.ensureTagsLocked(p.Tags)
	return p, a.store.Save(&a.state)
}

func (a *App) deletePrompt(payload json.RawMessage) error {
	id := readString(payload, "id")
	if id == "" {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for i := range a.state.Prompts {
		if strings.EqualFold(a.state.Prompts[i].ID, id) {
			a.state.Prompts = append(a.state.Prompts[:i], a.state.Prompts[i+1:]...)
			break
		}
	}
	a.reindexPromptsLocked()
	return a.store.Save(&a.state)
}

func (a *App) duplicatePrompt(payload json.RawMessage) (string, error) {
	id := readString(payload, "id")
	a.mu.Lock()
	defer a.mu.Unlock()
	idx := a.promptIndexLocked(id)
	if idx < 0 {
		return "", nil
	}
	source := a.state.Prompts[idx]
	for i := range a.state.Prompts {
		if a.state.Prompts[i].SortOrder > source.SortOrder {
			a.state.Prompts[i].SortOrder++
		}
	}
	now := time.Now()
	dup := PromptItem{ID: newID(), Title: source.Title + " - 副本", Category: source.Category, Tags: source.Tags, Content: source.Content, Notes: source.Notes, SortOrder: source.SortOrder + 1, CreatedAt: now, UpdatedAt: now}
	a.state.Prompts = append(a.state.Prompts, dup)
	a.reindexPromptsLocked()
	return dup.ID, a.store.Save(&a.state)
}

func (a *App) toggleFavorite(payload json.RawMessage) error {
	id := readString(payload, "id")
	a.mu.Lock()
	defer a.mu.Unlock()
	idx := a.promptIndexLocked(id)
	if idx >= 0 {
		a.state.Prompts[idx].IsFavorite = !a.state.Prompts[idx].IsFavorite
		a.state.Prompts[idx].UpdatedAt = time.Now()
	}
	return a.store.Save(&a.state)
}

func formatClipboardText(title, content string) string {
	title = strings.TrimSpace(title)
	if title == "" {
		return content
	}
	if content == "" {
		return title
	}
	return title + "\r\n\r\n" + content
}

func (a *App) copyPrompt(payload json.RawMessage) error {
	id := readString(payload, "id")
	a.mu.Lock()
	idx := a.promptIndexLocked(id)
	if idx < 0 {
		a.mu.Unlock()
		return nil
	}
	clipboardText := formatClipboardText(a.state.Prompts[idx].Title, a.state.Prompts[idx].Content)
	now := time.Now()
	a.state.Prompts[idx].UseCount++
	a.state.Prompts[idx].LastUsedAt = &now
	a.state.Prompts[idx].UpdatedAt = now
	err := a.store.Save(&a.state)
	a.mu.Unlock()
	if err != nil {
		return err
	}
	if a.ctx != nil {
		return wailsrt.ClipboardSetText(a.ctx, clipboardText)
	}
	return nil
}

func (a *App) createPrompt() (string, error) {
	a.mu.Lock()
	defer a.mu.Unlock()
	now := time.Now()
	category := a.defaultCategoryLocked()
	p := PromptItem{ID: newID(), Title: "新的提示词", Category: category, SortOrder: a.nextSortOrderLocked(), CreatedAt: now, UpdatedAt: now}
	a.state.Prompts = append(a.state.Prompts, p)
	return p.ID, a.store.Save(&a.state)
}

func (a *App) reorderPrompts(payload json.RawMessage) error {
	ids := readStringList(payload, "ids")
	if len(ids) == 0 {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	byID := make(map[string]PromptItem, len(a.state.Prompts))
	for _, p := range a.state.Prompts {
		byID[strings.ToLower(p.ID)] = p
	}
	ordered := make([]PromptItem, 0, len(a.state.Prompts))
	for _, id := range ids {
		key := strings.ToLower(id)
		if p, ok := byID[key]; ok {
			ordered = append(ordered, p)
			delete(byID, key)
		}
	}
	left := make([]PromptItem, 0, len(byID))
	for _, p := range byID {
		left = append(left, p)
	}
	sort.SliceStable(left, func(i, j int) bool { return left[i].SortOrder < left[j].SortOrder })
	ordered = append(ordered, left...)
	for i := range ordered {
		ordered[i].SortOrder = i
	}
	a.state.Prompts = ordered
	return a.store.Save(&a.state)
}

func (a *App) addCategory(payload json.RawMessage) error {
	name := strings.TrimSpace(readString(payload, "name"))
	color := strings.TrimSpace(readString(payload, "color"))
	if name == "" {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if containsFold(a.state.Categories, name) {
		return fmt.Errorf("分类“%s”已存在", name)
	}
	a.ensureCategoryLocked(name)
	if isCategoryPaletteColor(color) {
		setCategoryColor(&a.state, name, color)
	}
	return a.store.Save(&a.state)
}

func (a *App) deleteCategory(payload json.RawMessage) error {
	name := strings.TrimSpace(readString(payload, "name"))
	if name == "" {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()

	remaining := make([]string, 0, len(a.state.Categories))
	for _, c := range a.state.Categories {
		if !strings.EqualFold(c, name) {
			remaining = append(remaining, c)
		}
	}
	a.state.Categories = remaining
	deleteCategoryColor(&a.state, name)

	fallback := ""
	if !strings.EqualFold(name, CommonCategory) && containsFold(a.state.Categories, CommonCategory) {
		fallback = findFold(a.state.Categories, CommonCategory)
	} else if len(a.state.Categories) > 0 {
		fallback = a.state.Categories[0]
	}

	hasAffected := false
	for i := range a.state.Prompts {
		if strings.EqualFold(a.state.Prompts[i].Category, name) {
			hasAffected = true
			break
		}
	}
	if hasAffected && fallback == "" {
		fallback = FallbackCategory
		a.ensureCategoryLocked(fallback)
	}
	for i := range a.state.Prompts {
		if strings.EqualFold(a.state.Prompts[i].Category, name) {
			a.state.Prompts[i].Category = fallback
			a.state.Prompts[i].UpdatedAt = time.Now()
		}
	}
	return a.store.Save(&a.state)
}

func (a *App) renameCategory(payload json.RawMessage) error {
	oldName := strings.TrimSpace(readString(payload, "oldName"))
	newName := strings.TrimSpace(readString(payload, "newName"))
	if oldName == "" || newName == "" || strings.EqualFold(oldName, newName) {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, c := range a.state.Categories {
		if !strings.EqualFold(c, oldName) && strings.EqualFold(c, newName) {
			return fmt.Errorf("分类“%s”已存在", newName)
		}
	}
	found := false
	for i := range a.state.Categories {
		if strings.EqualFold(a.state.Categories[i], oldName) {
			a.state.Categories[i] = newName
			found = true
			break
		}
	}
	if !found {
		return nil
	}
	renameCategoryColor(&a.state, oldName, newName)
	for i := range a.state.Prompts {
		if strings.EqualFold(a.state.Prompts[i].Category, oldName) {
			a.state.Prompts[i].Category = newName
			a.state.Prompts[i].UpdatedAt = time.Now()
		}
	}
	return a.store.Save(&a.state)
}

func (a *App) reorderCategories(payload json.RawMessage) error {
	items := readStringList(payload, "items")
	a.mu.Lock()
	defer a.mu.Unlock()
	a.state.Categories = mergeOrder(items, a.state.Categories)
	return a.store.Save(&a.state)
}

func (a *App) setCategoryColor(payload json.RawMessage) error {
	name := strings.TrimSpace(readString(payload, "name"))
	color := strings.TrimSpace(readString(payload, "color"))
	if name == "" {
		return nil
	}
	if !isCategoryPaletteColor(color) {
		return errors.New("请选择 PromptNest 提供的 20 个分类颜色之一")
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if !containsFold(a.state.Categories, name) {
		return fmt.Errorf("分类“%s”不存在", name)
	}
	setCategoryColor(&a.state, findFold(a.state.Categories, name), color)
	return a.store.Save(&a.state)
}

func (a *App) addTag(payload json.RawMessage) error {
	name := strings.TrimSpace(readString(payload, "name"))
	if name == "" {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if !containsFold(a.state.Tags, name) {
		a.state.Tags = append(a.state.Tags, name)
	}
	return a.store.Save(&a.state)
}

func (a *App) deleteTag(payload json.RawMessage) error {
	name := strings.TrimSpace(readString(payload, "name"))
	if name == "" {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	filtered := a.state.Tags[:0]
	for _, t := range a.state.Tags {
		if !strings.EqualFold(t, name) {
			filtered = append(filtered, t)
		}
	}
	a.state.Tags = filtered
	for i := range a.state.Prompts {
		parts := splitTags(a.state.Prompts[i].Tags)
		kept := parts[:0]
		changed := false
		for _, t := range parts {
			if strings.EqualFold(t, name) {
				changed = true
				continue
			}
			kept = append(kept, t)
		}
		if changed {
			a.state.Prompts[i].Tags = strings.Join(kept, ", ")
			a.state.Prompts[i].UpdatedAt = time.Now()
		}
	}
	return a.store.Save(&a.state)
}

func (a *App) renameTag(payload json.RawMessage) error {
	oldName := strings.TrimSpace(readString(payload, "oldName"))
	newName := strings.TrimSpace(readString(payload, "newName"))
	if oldName == "" || newName == "" || strings.EqualFold(oldName, newName) {
		return nil
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	for _, t := range a.state.Tags {
		if !strings.EqualFold(t, oldName) && strings.EqualFold(t, newName) {
			return fmt.Errorf("标签“%s”已存在", newName)
		}
	}
	for i := range a.state.Tags {
		if strings.EqualFold(a.state.Tags[i], oldName) {
			a.state.Tags[i] = newName
			break
		}
	}
	for i := range a.state.Prompts {
		parts := splitTags(a.state.Prompts[i].Tags)
		changed := false
		for j := range parts {
			if strings.EqualFold(parts[j], oldName) {
				parts[j] = newName
				changed = true
			}
		}
		if changed {
			a.state.Prompts[i].Tags = strings.Join(distinctPreserveOrder(parts), ", ")
			a.state.Prompts[i].UpdatedAt = time.Now()
		}
	}
	return a.store.Save(&a.state)
}

func (a *App) reorderTags(payload json.RawMessage) error {
	items := readStringList(payload, "items")
	a.mu.Lock()
	defer a.mu.Unlock()
	a.state.Tags = mergeOrder(items, a.state.Tags)
	return a.store.Save(&a.state)
}

func (a *App) setTheme(payload json.RawMessage) error {
	var body struct {
		IsDark bool `json:"isDark"`
	}
	if err := json.Unmarshal(payload, &body); err != nil {
		return err
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	a.state.IsDarkTheme = body.IsDark
	return a.store.Save(&a.state)
}

func (a *App) exportLibrary() ([]BridgeEnvelope, error) {
	if a.ctx == nil {
		return nil, errors.New("窗口尚未初始化")
	}
	path, err := wailsrt.SaveFileDialog(a.ctx, wailsrt.SaveDialogOptions{
		Title:           "导出提示词库",
		DefaultFilename: "PromptNest-" + time.Now().Format("20060102") + ".json",
		Filters:         []wailsrt.FileFilter{{DisplayName: "PromptNest 提示词库 (*.json)", Pattern: "*.json"}},
	})
	if err != nil || path == "" {
		return nil, err
	}
	a.mu.Lock()
	state := cloneState(a.state)
	a.mu.Unlock()
	if err := a.store.Export(state, path); err != nil {
		return nil, err
	}
	return a.toast("提示词库已导出", "success"), nil
}

func (a *App) importLibrary() ([]BridgeEnvelope, error) {
	if a.ctx == nil {
		return nil, errors.New("窗口尚未初始化")
	}
	path, err := wailsrt.OpenFileDialog(a.ctx, wailsrt.OpenDialogOptions{
		Title:   "导入提示词库",
		Filters: []wailsrt.FileFilter{{DisplayName: "PromptNest 提示词库 (*.json)", Pattern: "*.json"}, {DisplayName: "JSON 文件 (*.json)", Pattern: "*.json"}},
	})
	if err != nil || path == "" {
		return nil, err
	}
	imported, err := a.store.Import(path)
	if err != nil {
		return nil, err
	}
	a.mu.Lock()
	pending := cloneState(imported)
	a.pendingImport = &pending
	a.mu.Unlock()
	return []BridgeEnvelope{{Type: "confirmImport", Payload: map[string]any{"count": len(imported.Prompts)}}}, nil
}

func (a *App) confirmImportLibrary() ([]BridgeEnvelope, error) {
	a.mu.Lock()
	if a.pendingImport == nil {
		a.mu.Unlock()
		return nil, errors.New("没有待确认的导入数据")
	}
	imported := cloneState(*a.pendingImport)
	a.pendingImport = nil
	a.state = imported
	err := a.store.Save(&a.state)
	a.mu.Unlock()
	if err != nil {
		return nil, err
	}
	return []BridgeEnvelope{a.stateMessage("imported", "")}, nil
}

func (a *App) cancelImportLibrary() {
	a.mu.Lock()
	a.pendingImport = nil
	a.mu.Unlock()
}

func (a *App) chooseDataDirectory() ([]BridgeEnvelope, error) {
	if a.ctx == nil {
		return nil, errors.New("窗口尚未初始化")
	}
	path, err := wailsrt.OpenDirectoryDialog(a.ctx, wailsrt.OpenDialogOptions{
		Title: "选择 PromptNest 数据目录",
	})
	if err != nil || strings.TrimSpace(path) == "" {
		return nil, err
	}
	return a.switchDataDirectory(path)
}

func (a *App) resetDataDirectory() ([]BridgeEnvelope, error) {
	return a.switchDataDirectory(a.store.DefaultDataDirectory)
}

func (a *App) switchDataDirectory(target string) ([]BridgeEnvelope, error) {
	target = strings.TrimSpace(target)
	if target == "" {
		return nil, errors.New("数据目录不能为空")
	}
	if samePath(target, a.store.DataDirectory) {
		return a.toast("已经在使用这个数据目录", ""), nil
	}

	if a.store.DataFileExistsInDirectory(target) {
		a.mu.Lock()
		a.pendingDataDirectory = target
		a.mu.Unlock()
		return []BridgeEnvelope{{Type: "confirmDataDirectory", Payload: map[string]any{"path": target}}}, nil
	}

	return a.applyDataDirectorySwitch(target, false)
}

func (a *App) confirmDataDirectorySwitch() ([]BridgeEnvelope, error) {
	a.mu.Lock()
	target := a.pendingDataDirectory
	a.pendingDataDirectory = ""
	a.mu.Unlock()
	if strings.TrimSpace(target) == "" {
		return nil, errors.New("没有待确认的数据目录")
	}
	return a.applyDataDirectorySwitch(target, true)
}

func (a *App) cancelDataDirectorySwitch() {
	a.mu.Lock()
	a.pendingDataDirectory = ""
	a.mu.Unlock()
}

func (a *App) applyDataDirectorySwitch(target string, loadExisting bool) ([]BridgeEnvelope, error) {
	if loadExisting {
		loaded, err := a.store.LoadFromDirectory(target)
		if err != nil {
			return nil, err
		}
		a.mu.Lock()
		if err := a.store.SetDataDirectory(target); err != nil {
			a.mu.Unlock()
			return nil, err
		}
		a.state = loaded
		a.mu.Unlock()
	} else {
		a.mu.Lock()
		current := cloneState(a.state)
		if err := a.store.SaveToDirectory(current, target); err != nil {
			a.mu.Unlock()
			return nil, fmt.Errorf("无法把当前提示词库写入新目录: %w", err)
		}
		if err := a.store.SetDataDirectory(target); err != nil {
			a.mu.Unlock()
			return nil, err
		}
		a.mu.Unlock()
	}

	return []BridgeEnvelope{
		a.stateMessage("dataDirectoryChanged", ""),
		{Type: "toast", Payload: map[string]any{"text": "数据目录已切换", "kind": "success"}},
	}, nil
}

func (a *App) openDataFolder() error {
	return openPath(a.store.DataDirectory)
}

func openPath(path string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer.exe", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	return cmd.Start()
}

func (a *App) defaultCategoryLocked() string {
	if c := findFold(a.state.Categories, CommonCategory); c != "" {
		return c
	}
	if len(a.state.Categories) > 0 {
		return a.state.Categories[0]
	}
	a.state.Categories = append(a.state.Categories, FallbackCategory)
	return FallbackCategory
}

func (a *App) ensureCategoryLocked(name string) {
	name = strings.TrimSpace(name)
	if name == "" {
		return
	}
	if !containsFold(a.state.Categories, name) {
		a.state.Categories = append(a.state.Categories, name)
	}
	canonicalName := findFold(a.state.Categories, name)
	if canonicalName == "" {
		canonicalName = name
	}
	ensureCategoryColor(&a.state, canonicalName)
}

func (a *App) ensureTagsLocked(tags string) {
	for _, tag := range splitTags(tags) {
		if !containsFold(a.state.Tags, tag) {
			a.state.Tags = append(a.state.Tags, tag)
		}
	}
}

func (a *App) promptIndexLocked(id string) int {
	for i := range a.state.Prompts {
		if strings.EqualFold(a.state.Prompts[i].ID, id) {
			return i
		}
	}
	return -1
}

func (a *App) nextSortOrderLocked() int {
	max := -1
	for _, p := range a.state.Prompts {
		if p.SortOrder > max {
			max = p.SortOrder
		}
	}
	return max + 1
}

func (a *App) reindexPromptsLocked() {
	sort.SliceStable(a.state.Prompts, func(i, j int) bool {
		if a.state.Prompts[i].SortOrder == a.state.Prompts[j].SortOrder {
			return a.state.Prompts[i].CreatedAt.Before(a.state.Prompts[j].CreatedAt)
		}
		return a.state.Prompts[i].SortOrder < a.state.Prompts[j].SortOrder
	})
	for i := range a.state.Prompts {
		a.state.Prompts[i].SortOrder = i
	}
}

func readString(payload json.RawMessage, key string) string {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(payload, &obj); err != nil {
		return ""
	}
	var value string
	_ = json.Unmarshal(obj[key], &value)
	return value
}

func readStringList(payload json.RawMessage, key string) []string {
	var obj map[string]json.RawMessage
	if err := json.Unmarshal(payload, &obj); err != nil {
		return nil
	}
	var values []string
	_ = json.Unmarshal(obj[key], &values)
	return distinctPreserveOrder(values)
}

func containsFold(values []string, target string) bool { return findFold(values, target) != "" }

func findFold(values []string, target string) string {
	for _, v := range values {
		if strings.EqualFold(v, target) {
			return v
		}
	}
	return ""
}

func mergeOrder(preferred, existing []string) []string {
	return distinctPreserveOrder(append(append([]string{}, preferred...), existing...))
}
