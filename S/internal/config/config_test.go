// Package config test cho defaults, Load với env override, parseSimple.
package config

import (
	"os"
	"testing"
)

func TestDefaults(t *testing.T) {
	cfg := defaults()
	if cfg.MaxCodeDump != 4096 {
		t.Errorf("MaxCodeDump = %d, want 4096", cfg.MaxCodeDump)
	}
	if cfg.MaxOutput != 10240 {
		t.Errorf("MaxOutput = %d, want 10240", cfg.MaxOutput)
	}
	if !cfg.EnableGitGuard {
		t.Error("EnableGitGuard phải true mặc định")
	}
	if len(cfg.SecretPatterns) == 0 {
		t.Error("SecretPatterns phải có ít nhất 1")
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	os.Setenv("WW_MAX_OUTPUT", "5555")
	defer os.Unsetenv("WW_MAX_OUTPUT")

	cfg := Load()
	if cfg.MaxOutput != 5555 {
		t.Errorf("MaxOutput = %d, want 5555", cfg.MaxOutput)
	}
}

func TestLoad_DisableGuard(t *testing.T) {
	os.Setenv("WW_NO_GITGUARD", "1")
	defer os.Unsetenv("WW_NO_GITGUARD")

	cfg := Load()
	if cfg.EnableGitGuard {
		t.Error("EnableGitGuard phải false khi WW_NO_GITGUARD=1")
	}
}

func TestParseStringArray(t *testing.T) {
	got := parseStringArray(`["a", "b", "c"]`)
	if len(got) != 3 || got[0] != "a" || got[2] != "c" {
		t.Errorf("parse = %v, want [a b c]", got)
	}
}

func TestParseSimple(t *testing.T) {
	cfg := defaults()
	data := "max_output = 777\nmax_code_dump = 888\nenable_git_guard = false\n"
	parseSimple(data, cfg)
	if cfg.MaxOutput != 777 || cfg.MaxCodeDump != 888 {
		t.Errorf("parse = %d/%d, want 777/888", cfg.MaxOutput, cfg.MaxCodeDump)
	}
	if cfg.EnableGitGuard {
		t.Error("EnableGitGuard phải false")
	}
}
