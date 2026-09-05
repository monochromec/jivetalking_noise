package processor

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDebugMP3OutputFormat(t *testing.T) {
	cfg := DefaultFilterConfig()
	cfg.OutputFormat = "mp3"
	cfg.Resample.Enabled = true
	cfg.Resample.SampleRate = 44100
	cfg.Resample.Format = "s16"
	cfg.Resample.FrameSize = 4096

	t.Logf("base.OutputFormat=%q", cfg.OutputFormat)
	eff := deriveEffectiveFilterConfig(cfg)
	t.Logf("eff.OutputFormat=%q", eff.OutputFormat)
	t.Logf("eff.RequiredFmt=%q", eff.requiredOutputSampleFmt())
	spec := eff.BuildFilterSpec()
	t.Logf("filter spec=%s", spec)
	t.Logf("contains s16p=%t", strings.Contains(spec, "sample_fmts=s16p"))
}

func TestProcessAudioMP3(t *testing.T) {
	inputPath := generateTestAudio(t, TestAudioOptions{
		DurationSecs: 1.0,
		SampleRate:   44100,
		ToneFreq:     440.0,
		ToneLevel:    -18.0,
		SilenceGap: struct {
			Start    float64
			Duration float64
		}{Start: 0.5, Duration: 0.2},
	})
	defer os.Remove(inputPath)

	config := DefaultFilterConfig()
	config.OutputFormat = "mp3"

	result, err := ProcessAudio(inputPath, config, false, nil)
	if err != nil {
		t.Fatalf("ProcessAudio failed: %v", err)
	}
	if result == nil || result.OutputPath == "" {
		t.Fatalf("ProcessAudio returned no result or output path")
	}
	defer os.Remove(result.OutputPath)

	if filepath.Ext(result.OutputPath) != ".mp3" {
		t.Fatalf("expected mp3 output file, got %q", result.OutputPath)
	}
}
