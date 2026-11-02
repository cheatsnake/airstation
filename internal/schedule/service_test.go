package schedule

import (
	"testing"
	"time"
)

func TestValidateTriggers(t *testing.T) {
	tests := []struct {
		name    string
		everyN  int
		atMin   int
		wantErr bool
	}{
		{name: "every n tracks", everyN: 3, atMin: -1},
		{name: "top of hour", everyN: 0, atMin: 0},
		{name: "both", everyN: 5, atMin: 30},
		{name: "nothing", everyN: 0, atMin: -1, wantErr: true},
		{name: "negative every n", everyN: -1, atMin: -1, wantErr: true},
		{name: "minute too high", everyN: 0, atMin: 60, wantErr: true},
		{name: "minute too low", everyN: 0, atMin: -2, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateTriggers(tt.everyN, tt.atMin)
			if (err != nil) != tt.wantErr {
				t.Fatalf("validateTriggers(%d, %d) error = %v, wantErr %v", tt.everyN, tt.atMin, err, tt.wantErr)
			}
		})
	}
}

func TestClockDue(t *testing.T) {
	now := time.Date(2026, 11, 1, 15, 30, 0, 0, time.Local)
	lastTopHour := hourStart(now)

	if !clockDue(&Rule{AtMinute: 30, Enabled: true}, now, 0) {
		t.Fatal("expected rule to be due at minute 30")
	}
	if clockDue(&Rule{AtMinute: 30, Enabled: true}, now, lastTopHour) {
		t.Fatal("rule should not fire twice in the same hour")
	}
	if clockDue(&Rule{AtMinute: 31, Enabled: true}, now, 0) {
		t.Fatal("rule should not fire at the wrong minute")
	}
	if clockDue(&Rule{AtMinute: 30, Enabled: false}, now, 0) {
		t.Fatal("disabled rule should not fire")
	}
}

func TestRotationDue(t *testing.T) {
	if !rotationDue(&Rule{EveryNTracks: 3, Enabled: true}, 3) {
		t.Fatal("expected rule to be due after 3 tracks")
	}
	if rotationDue(&Rule{EveryNTracks: 3, Enabled: true}, 2) {
		t.Fatal("rule should not fire before 3 tracks")
	}
	if rotationDue(&Rule{EveryNTracks: 0, Enabled: true}, 100) {
		t.Fatal("rule with every_n_tracks=0 should be disabled")
	}
}
