// Package gitguard test cho Scan, Blocked, Summary. Cover các pattern
// phổ biến (OpenAI, GitHub, AWS, private key) và trường hợp sạch.
package gitguard

import "testing"

var defaultPatterns = []string{
	"sk-", "ghp_", "AKIA", "-----BEGIN PRIVATE KEY-----",
}

func TestScan_Clean(t *testing.T) {
	content := "#!/bin/bash\necho hello\necho world\n"
	findings := Scan(content, defaultPatterns)
	if Blocked(findings) {
		t.Errorf("nội dung sạch nhưng bị block: %+v", findings)
	}
}

func TestScan_OpenAIKey(t *testing.T) {
	content := "#!/bin/bash\necho sk-abc123def456\n"
	findings := Scan(content, defaultPatterns)
	if !Blocked(findings) {
		t.Fatal("phải phát hiện sk-")
	}
	if findings[0].Line != 2 {
		t.Errorf("line = %d, want 2", findings[0].Line)
	}
}

func TestScan_GitHubToken(t *testing.T) {
	content := "ghp_xxxxxxxxxxxxxxxxxxxx"
	findings := Scan(content, defaultPatterns)
	if !Blocked(findings) {
		t.Fatal("phải phát hiện ghp_")
	}
}

func TestScan_AWSKey(t *testing.T) {
	content := "const key = \"AKIAIOSFODNN7EXAMPLE\";"
	findings := Scan(content, defaultPatterns)
	if !Blocked(findings) {
		t.Fatal("phải phát hiện AKIA")
	}
}

func TestScan_PrivateKey(t *testing.T) {
	content := "-----BEGIN PRIVATE KEY-----\nMIIEv...\n-----END PRIVATE KEY-----"
	findings := Scan(content, defaultPatterns)
	if !Blocked(findings) {
		t.Fatal("phải phát hiện private key")
	}
}

func TestScan_MultipleLines(t *testing.T) {
	content := "line1 sk-foo\nline2 clean\nline3 ghp_bar\n"
	findings := Scan(content, defaultPatterns)
	if len(findings) < 2 {
		t.Errorf("phải phát hiện ≥ 2, được %d", len(findings))
	}
}

func TestSummary_Empty(t *testing.T) {
	if s := Summary(nil); s != "" {
		t.Errorf("Summary(nil) = %q, want empty", s)
	}
}

func TestSummary_ContainsPatternAndLine(t *testing.T) {
	content := "x sk-abc"
	findings := Scan(content, defaultPatterns)
	s := Summary(findings)
	if s == "" {
		t.Fatal("Summary rỗng")
	}
}
