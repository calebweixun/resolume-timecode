package main

import (
	"testing"

	"github.com/chabad360/go-osc/osc"
)

func TestProcPosRecoversAfterPositionMovesBackward(t *testing.T) {
	originalPrev := posPrev
	originalDirection := directionForward
	t.Cleanup(func() {
		posPrev = originalPrev
		directionForward = originalDirection
	})

	directionForward = true
	posPrev = 0.8
	procPos(osc.NewMessage("/composition/selectedclip/transport/position", float32(0.1)))

	if posPrev != 0.1 {
		t.Fatalf("expected backward jump to establish baseline 0.1, got %v", posPrev)
	}
}
