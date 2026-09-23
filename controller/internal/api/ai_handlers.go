package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/managerhub/managerhub/controller/internal/auth"
)

// AIConfig stores the OpenRouter API key and default model.
type AIConfig struct {
	APIKey      string  `json:"api_key"`
	Model       string  `json:"model"`
	BaseURL     string  `json:"base_url"`
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
}

type aiChatReq struct {
	Message   string `json:"message"`
	Model     string `json:"model,omitempty"`
	System    string `json:"system,omitempty"`
	MaxTokens int    `json:"max_tokens,omitempty"`
}

type aiChatResp struct {
	Content string `json:"content"`
	Model   string `json:"model"`
}

func (s *Server) loadAIConfig(ctx context.Context) AIConfig {
	var raw string
	err := s.Store.Pool.QueryRow(ctx,
		`SELECT value FROM settings WHERE key='ai_config'`).Scan(&raw)
	if err != nil {
		return AIConfig{Model: "openai/gpt-4o-mini", BaseURL: "https://openrouter.ai/api/v1", MaxTokens: 2048, Temperature: 0.7}
	}
	var cfg AIConfig
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return AIConfig{Model: "openai/gpt-4o-mini", BaseURL: "https://openrouter.ai/api/v1", MaxTokens: 2048, Temperature: 0.7}
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://openrouter.ai/api/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "openai/gpt-4o-mini"
	}
	return cfg
}

func (s *Server) handleAIChat(w http.ResponseWriter, r *http.Request) {
	var req aiChatReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	cfg := s.loadAIConfig(r.Context())
	if cfg.APIKey == "" {
		writeErr(w, http.StatusBadRequest, "OpenRouter API key not configured — go to Settings")
		return
	}
	model := req.Model
	if model == "" {
		model = cfg.Model
	}
	maxTok := req.MaxTokens
	if maxTok <= 0 {
		maxTok = cfg.MaxTokens
	}
	if maxTok <= 0 {
		maxTok = 2048
	}
	sys := req.System
	if sys == "" {
		sys = "You are ManagerHub AI assistant for infrastructure management. Be concise and precise."
	}

	payload := map[string]any{
		"model": model,
		"messages": []map[string]string{
			{"role": "system", "content": sys},
			{"role": "user", "content": req.Message},
		},
		"max_tokens":  maxTok,
		"temperature": cfg.Temperature,
	}
	body, _ := json.Marshal(payload)

	httpReq, _ := http.NewRequestWithContext(r.Context(), "POST", cfg.BaseURL+"/chat/completions", bytes.NewReader(body))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+cfg.APIKey)
	httpReq.Header.Set("HTTP-Referer", s.Cfg.PublicURL)
	httpReq.Header.Set("X-Title", "ManagerHub")

	client := &http.Client{Timeout: 120 * time.Second}
	resp, err := client.Do(httpReq)
	if err != nil {
		writeErr(w, http.StatusBadGateway, "OpenRouter error: "+err.Error())
		return
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if resp.StatusCode != 200 {
		writeErr(w, resp.StatusCode, "OpenRouter: "+string(raw))
		return
	}

	var orResp struct {
		Model   string `json:"model"`
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(raw, &orResp); err != nil {
		writeErr(w, http.StatusBadGateway, "bad OpenRouter response")
		return
	}
	content := ""
	if len(orResp.Choices) > 0 {
		content = orResp.Choices[0].Message.Content
	}
	writeJSON(w, http.StatusOK, aiChatResp{Content: content, Model: orResp.Model})
}

func (s *Server) handleAIConfig(w http.ResponseWriter, r *http.Request) {
	cfg := s.loadAIConfig(r.Context())
	masked := cfg
	if len(masked.APIKey) > 8 {
		masked.APIKey = masked.APIKey[:4] + "..." + masked.APIKey[len(masked.APIKey)-4:]
	} else if masked.APIKey != "" {
		masked.APIKey = "****"
	}
	writeJSON(w, http.StatusOK, masked)
}

func (s *Server) handleAISetConfig(w http.ResponseWriter, r *http.Request) {
	var cfg AIConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		writeErr(w, http.StatusBadRequest, "invalid body")
		return
	}
	// If a masked key is sent back, preserve the real one.
	if strings.Contains(cfg.APIKey, "...") || cfg.APIKey == "****" {
		cfg.APIKey = s.loadAIConfig(r.Context()).APIKey
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://openrouter.ai/api/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "openai/gpt-4o-mini"
	}
	b, _ := json.Marshal(cfg)
	_, err := s.Store.Pool.Exec(r.Context(),
		`INSERT INTO settings (key, value) VALUES ('ai_config', $1) ON CONFLICT (key) DO UPDATE SET value=EXCLUDED.value`,
		string(b))
	if err != nil {
		writeErr(w, http.StatusInternalServerError, "save failed")
		return
	}
	c, _ := auth.ClaimsFrom(r.Context())
	_ = s.Store.AppendAudit(r.Context(), c.Subject, "ai.config_update", "", "")
	writeJSON(w, http.StatusOK, map[string]string{"status": "saved"})
}

func (s *Server) handleAIModels(w http.ResponseWriter, _ *http.Request) {
	models := []string{
		"openai/gpt-4o-mini",
		"openai/gpt-4o",
		"anthropic/claude-3.5-sonnet",
		"anthropic/claude-3-haiku",
		"google/gemini-2.0-flash-exp:free",
		"google/gemini-pro-1.5",
		"meta-llama/llama-3.1-8b-instruct",
		"deepseek/deepseek-chat",
		"qwen/qwen-2.5-72b-instruct",
	}
	writeJSON(w, http.StatusOK, models)
}
