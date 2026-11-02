package ui

import (
	"strings"
	"testing"
)

func TestBuildTimeLabelUsesInjectedValue(t *testing.T) {
	original := BuildTime
	defer func() { BuildTime = original }()

	BuildTime = "2026-11-02T17:20"
	if got := buildTimeLabel(); got != "2026-11-02 17:20" {
		t.Errorf("buildTimeLabel() = %q", got)
	}
}

func TestBuildTimeLabelFallsBackToExecutable(t *testing.T) {
	original := BuildTime
	defer func() { BuildTime = original }()

	BuildTime = ""
	if got := buildTimeLabel(); got == "" || got == "unknown" {
		t.Errorf("buildTimeLabel() = %q", got)
	}
}

func TestVersionShownOnAboutOnly(t *testing.T) {
	original := BuildTime
	defer func() { BuildTime = original }()
	BuildTime = "2026-11-02T17:20"

	about := RenderAbout()
	for _, want := range []string{"V" + Version, "2026-11-02 17:20"} {
		if !strings.Contains(about, want) {
			t.Errorf("about page missing %q:\n%s", want, about)
		}
	}

	if welcome := RenderWelcome(&WelcomeAnimationState{}, 0); strings.Contains(welcome, "V"+Version) {
		t.Errorf("welcome page should not show the version:\n%s", welcome)
	}
}
