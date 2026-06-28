package main

import (
	"testing"

	"github.com/chabad360/go-osc/osc"
)

func TestProcPosUpdatesImmediatelyAfterPositionMovesBackward(t *testing.T) {
	originalPrev := posPrev
	originalDirection := directionForward
	originalLength := clipLength
	originalInvert := clipInvert
	originalTimeLeft := timeLeft
	t.Cleanup(func() {
		posPrev = originalPrev
		directionForward = originalDirection
		clipLength = originalLength
		clipInvert = originalInvert
		timeLeft = originalTimeLeft
	})

	directionForward = true
	clipInvert = false
	clipLength = 60
	posPrev = 0.8
	procPos(osc.NewMessage("/composition/selectedclip/transport/position", float32(0.1)))

	if posPrev != 0.1 {
		t.Fatalf("expected position 0.1, got %v", posPrev)
	}
	if timeLeft != "-00:00:54.000" {
		t.Fatalf("expected immediate countdown update, got %q", timeLeft)
	}
}

func TestMonitoredClipPathUsesPlayingClipForLayer(t *testing.T) {
	originalPath := clipPath
	originalActive := activeClipPath
	t.Cleanup(func() {
		clipPath = originalPath
		activeClipPath = originalActive
	})

	clipPath = "/composition/layers/3"
	activeClipPath = ""
	if got := monitoredClipPath(); got != "/composition/layers/3/clips/playing" {
		t.Fatalf("unexpected playing clip path: %q", got)
	}

	activeClipPath = "/composition/layers/3/clips/7"
	if got := monitoredClipPath(); got != activeClipPath {
		t.Fatalf("expected discovered clip path %q, got %q", activeClipPath, got)
	}
}
