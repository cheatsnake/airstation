package mix

import (
	"strings"
	"testing"
)

func TestBuildFilterIncludesFadesDuckingAndDelay(t *testing.T) {
	filter, err := BuildFilter(Options{
		BaseDuration:  180,
		VoiceDuration: 20,
		StartOffset:   5,
		FadeIn:        2,
		FadeOut:       3,
		DuckDB:        -12,
	})
	if err != nil {
		t.Fatalf("BuildFilter returned an error: %v", err)
	}

	for _, want := range []string{
		"afade=t=in:st=0:d=2.000",
		"afade=t=out:st=17.000:d=3.000",
		"adelay=delays=5000:all=1",
		"apad=whole_dur=180.000",
		"volume=volume='1-(1-0.2512)*min(max(0,min(1,(t-5.000)/2.000)),max(0,min(1,(25.000-t)/3.000)))':eval=frame",
		"amix=inputs=2:duration=first",
		"alimiter",
	} {
		if !strings.Contains(filter, want) {
			t.Errorf("filter is missing %q:\n%s", want, filter)
		}
	}
	if strings.Contains(filter, "sidechaincompress") {
		t.Errorf("filter should duck with a deterministic envelope, not a level-triggered compressor:\n%s", filter)
	}
}

func TestBuildFilterWithoutDucking(t *testing.T) {
	filter, err := BuildFilter(Options{
		BaseDuration:  120,
		VoiceDuration: 10,
		StartOffset:   1,
		FadeIn:        1,
		FadeOut:       1,
		DuckDB:        0,
	})
	if err != nil {
		t.Fatalf("BuildFilter returned an error: %v", err)
	}
	if strings.Contains(filter, "volume=") {
		t.Errorf("music should not be ducked when DuckDB is zero:\n%s", filter)
	}
	if !strings.Contains(filter, "amix=inputs=2") {
		t.Errorf("filter should still mix the voice:\n%s", filter)
	}
}

func TestBuildFilterRejectsVoiceLongerThanMusic(t *testing.T) {
	_, err := BuildFilter(Options{
		BaseDuration:  10,
		VoiceDuration: 20,
	})
	if err == nil {
		t.Fatal("expected an error when the voice clip is longer than the music bed")
	}
}

func TestBuildFilterClampsStartOffset(t *testing.T) {
	filter, err := BuildFilter(Options{
		BaseDuration:  30,
		VoiceDuration: 10,
		StartOffset:   100,
		FadeIn:        0,
		FadeOut:       0,
		DuckDB:        -18,
	})
	if err != nil {
		t.Fatalf("BuildFilter returned an error: %v", err)
	}
	if !strings.Contains(filter, "adelay=delays=20000:all=1") {
		t.Errorf("start offset should clamp to base-voice duration:\n%s", filter)
	}
	if !strings.Contains(filter, "if(lt(t,20.000),0,1)") {
		t.Errorf("zero fade-in should still start the duck at the clamped offset:\n%s", filter)
	}
}

func TestBuildFilterCapsOverlappingFades(t *testing.T) {
	filter, err := BuildFilter(Options{
		BaseDuration:  60,
		VoiceDuration: 4,
		FadeIn:        5,
		FadeOut:       5,
	})
	if err != nil {
		t.Fatalf("BuildFilter returned an error: %v", err)
	}
	if !strings.Contains(filter, "afade=t=in:st=0:d=2.000") {
		t.Errorf("fade-in should be capped at half the voice duration:\n%s", filter)
	}
	if !strings.Contains(filter, "afade=t=out:st=2.000:d=2.000") {
		t.Errorf("fade-out should be capped at half the voice duration:\n%s", filter)
	}
}
