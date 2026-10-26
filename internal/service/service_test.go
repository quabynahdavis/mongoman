package service

import (
	"strings"
	"testing"
)

func TestStatusString(t *testing.T) {
	tests := []struct {
		s        Status
		expected string
	}{
		{StatusUnknown, "❓ Unknown"},
		{StatusRunning, "✅ Running"},
		{StatusStopped, "❌ Stopped"},
		{StatusEnabled, "🔁 Enabled"},
		{StatusDisabled, "⛔ Disabled"},
		{Status(99), "❓ Unknown"}, // out-of-range falls back to Unknown
	}
	for _, tt := range tests {
		if got := tt.s.String(); got != tt.expected {
			t.Errorf("Status(%d).String() = %q, want %q", tt.s, got, tt.expected)
		}
	}
}

func TestErrNotSupported(t *testing.T) {
	e := &errNotSupported{feature: "systemd/launchd"}
	if e.Error() == "" {
		t.Fatal("expected non-empty error message")
	}
	if !strings.Contains(e.Error(), "systemd/launchd") {
		t.Errorf("error should mention feature: %s", e.Error())
	}
	if !strings.Contains(strings.ToLower(e.Error()), "not supported") {
		t.Errorf("error should indicate unsupported: %s", e.Error())
	}
}

func TestNewManager(t *testing.T) {
	m := NewManager()
	if m == nil {
		t.Fatal("NewManager() returned nil")
	}
	// The stub platform's IsEnabled must report false for unknown instances
	// without invoking any external commands.
	if m.IsEnabled("definitely-not-a-real-instance-12345") {
		t.Fatal("stub IsEnabled should be false for unknown instances")
	}
}