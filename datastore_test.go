package main

import (
	"path/filepath"
	"testing"
)

func TestCustomDataDirectoryPersists(t *testing.T) {
	root := t.TempDir()
	defaultDir := filepath.Join(root, "default")
	customDir := filepath.Join(root, "custom")

	store := &DataStore{
		DefaultDataDirectory: defaultDir,
		ConfigDirectory:      defaultDir,
		ConfigFilePath:       filepath.Join(defaultDir, "settings.json"),
		DataDirectory:        defaultDir,
		DataFilePath:         filepath.Join(defaultDir, "prompts.json"),
	}

	state := createDefaultState()
	if err := store.Save(&state); err != nil {
		t.Fatalf("save default state: %v", err)
	}
	if err := store.SaveToDirectory(state, customDir); err != nil {
		t.Fatalf("seed custom directory: %v", err)
	}
	if err := store.SetDataDirectory(customDir); err != nil {
		t.Fatalf("set custom data directory: %v", err)
	}
	if !store.IsCustomDataDirectory() {
		t.Fatal("expected custom data directory")
	}

	reloaded := &DataStore{
		DefaultDataDirectory: defaultDir,
		ConfigDirectory:      defaultDir,
		ConfigFilePath:       filepath.Join(defaultDir, "settings.json"),
		DataDirectory:        defaultDir,
		DataFilePath:         filepath.Join(defaultDir, "prompts.json"),
	}
	reloaded.loadSettings()
	if !samePath(reloaded.DataDirectory, customDir) {
		t.Fatalf("custom directory was not restored: got %q want %q", reloaded.DataDirectory, customDir)
	}
	loaded := reloaded.Load()
	if len(loaded.Prompts) != len(state.Prompts) {
		t.Fatalf("loaded prompt count mismatch: got %d want %d", len(loaded.Prompts), len(state.Prompts))
	}

	if err := reloaded.ResetDataDirectory(); err != nil {
		t.Fatalf("reset data directory: %v", err)
	}
	if reloaded.IsCustomDataDirectory() {
		t.Fatal("expected default data directory after reset")
	}
}

func TestLegacyCategoriesReceiveDistinctPaletteColors(t *testing.T) {
	state := AppState{
		Categories: []string{"通用分类", "temu", "软件开发", "1"},
		Prompts: []PromptItem{
			{ID: "1", Title: "A", Category: "通用分类"},
			{ID: "2", Title: "B", Category: "temu"},
			{ID: "3", Title: "C", Category: "软件开发"},
			{ID: "4", Title: "D", Category: "1"},
		},
	}
	normalizeState(&state)
	if len(state.CategoryColors) != 4 {
		t.Fatalf("expected 4 category colors, got %d", len(state.CategoryColors))
	}
	seen := map[string]bool{}
	for _, category := range state.Categories {
		color := categoryColorFromMap(state.CategoryColors, category)
		if !isCategoryPaletteColor(color) {
			t.Fatalf("category %q got invalid palette color %q", category, color)
		}
		if seen[color] {
			t.Fatalf("expected unique initial colors, duplicate %q", color)
		}
		seen[color] = true
	}
}

func TestCategoryColorSurvivesRename(t *testing.T) {
	state := AppState{Categories: []string{"软件开发"}, CategoryColors: map[string]string{"软件开发": CategoryPalette[7]}}
	normalizeState(&state)
	renameCategoryColor(&state, "软件开发", "开发")
	if got := categoryColorFromMap(state.CategoryColors, "开发"); got != CategoryPalette[7] {
		t.Fatalf("renamed category color mismatch: got %q want %q", got, CategoryPalette[7])
	}
	if got := categoryColorFromMap(state.CategoryColors, "软件开发"); got != "" {
		t.Fatalf("old category color key still exists: %q", got)
	}
}
