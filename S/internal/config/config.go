// Package config đọc cấu hình ś từ ~/ten-reepo/S/config.toml (nếu có)
// và override bằng biến môi trường. Mục đích: không hard-code các ngưỡng
// (max output, max code dump) và danh sách secret pattern. Nếu không có
// file config, dùng default an toàn.
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Config chứa toàn bộ cấu hình runtime của ś.
type Config struct {
	// MaxCodeDump: nếu script ngắn hơn hoặc bằng ngưỡng này thì dump
	// nguyên vào report; dài hơn chỉ ghi path + size.
	MaxCodeDump int `toml:"max_code_dump"`
	// MaxOutput: cắt stdout/stderr ở ngưỡng này (bytes).
	MaxOutput int `toml:"max_output"`
	// SecretPatterns: danh sách substring cần chặn trước khi git sync.
	SecretPatterns []string `toml:"secret_patterns"`
	// EnableGitGuard: bật/tắt scan secret trước khi push.
	EnableGitGuard bool `toml:"enable_git_guard"`
}

// defaults trả về config mặc định an toàn.
func defaults() *Config {
	return &Config{
		MaxCodeDump: 4096,
		MaxOutput:   10240,
		EnableGitGuard: true,
		SecretPatterns: []string{
			"sk-", "sk_live_", "sk_test_",
			"ghp_", "gho_", "ghu_", "ghs_", "ghr_",
			"AKIA", "ASIA",
			"-----BEGIN PRIVATE KEY-----",
			"-----BEGIN RSA PRIVATE KEY-----",
			"xoxb-", "xoxp-",
			"AIza",
			"ya29.",
		},
	}
}

// Home trả về HOME hoặc fallback Termux.
func Home() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return "/data/data/com.termux/files/home"
}

// Load đọc config từ ~/ten-reepo/S/config.toml nếu tồn tại, override
// bằng env, trả về Config. Không lỗi nếu file thiếu hoặc parse lỗi —
// luôn trả default + override env.
func Load() *Config {
	cfg := defaults()
	path := filepath.Join(Home(), "ten-reepo", "S", "config.toml")
	if data, err := os.ReadFile(path); err == nil {
		parseSimple(string(data), cfg)
	}
	// Env override: WW_MAX_OUTPUT, WW_MAX_CODE_DUMP, WW_NO_GITGUARD.
	if v := os.Getenv("WW_MAX_OUTPUT"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.MaxOutput)
	}
	if v := os.Getenv("WW_MAX_CODE_DUMP"); v != "" {
		fmt.Sscanf(v, "%d", &cfg.MaxCodeDump)
	}
	if v := os.Getenv("WW_NO_GITGUARD"); v == "1" {
		cfg.EnableGitGuard = false
	}
	return cfg
}

// parseSimple parse TOML tối giản (key = value, array of string) đủ
// dùng cho config ś. Không kéo dependency TOML để giữ binary nhỏ.
func parseSimple(data string, cfg *Config) {
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		switch key {
		case "max_output":
			fmt.Sscanf(val, "%d", &cfg.MaxOutput)
		case "max_code_dump":
			fmt.Sscanf(val, "%d", &cfg.MaxCodeDump)
		case "enable_git_guard":
			cfg.EnableGitGuard = val == "true"
		case "secret_patterns":
			cfg.SecretPatterns = parseStringArray(val)
		}
	}
}

// parseStringArray parse `["a", "b"]` thành []string.
func parseStringArray(s string) []string {
	s = strings.Trim(s, "[]")
	var out []string
	for _, p := range strings.Split(s, ",") {
		p = strings.TrimSpace(p)
		p = strings.Trim(p, `"'`)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
