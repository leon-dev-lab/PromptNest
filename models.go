package main

import (
	"encoding/json"
	"time"
)

type AppState struct {
	IsDarkTheme    bool              `json:"isDarkTheme"`
	Categories     []string          `json:"categories"`
	CategoryColors map[string]string `json:"categoryColors,omitempty"`
	Tags           []string          `json:"tags"`
	Prompts        []PromptItem      `json:"prompts"`
}

type PromptItem struct {
	ID         string     `json:"id"`
	Title      string     `json:"title"`
	Category   string     `json:"category"`
	Tags       string     `json:"tags"`
	Content    string     `json:"content"`
	Notes      string     `json:"notes"`
	IsFavorite bool       `json:"isFavorite"`
	SortOrder  int        `json:"sortOrder"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
	LastUsedAt *time.Time `json:"lastUsedAt,omitempty"`
	UseCount   int        `json:"useCount"`
}

type BridgeMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
}

type BridgeEnvelope struct {
	Type    string `json:"type"`
	Payload any    `json:"payload"`
}
