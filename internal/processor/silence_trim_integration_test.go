package processor

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/linuxmatters/jivetalking/internal/audio"
)

// TestSilenceTrimIntegration generates a synthetic audio file containing a long
// silence region and verifies that enabling SilenceTrim truncates that region
// to the configured MaxDuration.
func TestSilenceTrimIntegration(t *testing.T) {
	// Create a short test file: total 8s, silence from 2s..6s (4s long)
	inputPath := generateTestAudio(t, TestAudioOptions{
		DurationSecs: 8.0,
		SampleRate:   44100,
		ToneFreq:     440.0,
		ToneLevel:    -18.0,
		NoiseLevel:   -60.0, // small noise floor inside silence
		SilenceGap: struct {
			Start    float64
			Duration float64
		}{Start: 2.0, Duration: 4.0},
	})

	// Enable silence trimming to 2s
	cfg := DefaultFilterConfig()
	cfg.SilenceTrim.Enabled = true
	cfg.SilenceTrim.MaxDuration = 2 * time.Second
	cfg.Analysis.SilenceScanDuration = 0

	measurements, err := AnalyzeAudio(inputPath, cfg, nil)
	if err != nil {
		t.Fatalf("AnalyzeAudio failed: %v", err)
	}
	t.Logf("detected silence regions: %+v", measurements.SilenceRegions)
	if len(measurements.SilenceRegions) == 0 {
		t.Fatal("expected the synthetic gap to be detected")
	}
	region := measurements.SilenceRegions[0]
	if region.Start < 1500*time.Millisecond || region.Start > 2500*time.Millisecond {
		t.Fatalf("detected silence starts outside the synthetic gap: %+v", measurements.SilenceRegions)
	}
	if region.End < 5500*time.Millisecond || region.End > 6500*time.Millisecond {
		t.Fatalf("detected silence consumes trailing audio: %+v", measurements.SilenceRegions)
	}

	// Process the audio
	result, err := ProcessAudio(inputPath, cfg, true, nil)
	if err != nil {
		t.Fatalf("ProcessAudio failed: %v", err)
	}
	// Keep the output file for inspection. Print paths + durations as evidence.
	fmt.Printf("input: %s\n", inputPath)
	fmt.Printf("output: %s\n", result.OutputPath)

	// Open output and verify duration ~= 6s (8 - (4-2))
	_, meta, err := audio.OpenAudioFile(result.OutputPath)
	if err != nil {
		t.Fatalf("failed to open output file: %v", err)
	}

	want := 6.0
	fmt.Printf("input_duration: %.3f\n", 8.0)
	fmt.Printf("output_duration: %.3f\n", meta.Duration)

	if meta.Duration < want-0.15 || meta.Duration > want+0.15 {
		t.Fatalf("unexpected output duration: got %.3f s, want ~%.3f s", meta.Duration, want)
	}

	// Write a small marker file to make it easy for external scripts to find
	// the generated artifacts when running the demo.
	marker := fmt.Sprintf("%s\n%s\n", inputPath, result.OutputPath)
	// t.TempDir() creates a unique, isolated temp directory for the test duration
	// t.TempDir() creates a unique, isolated temp directory for the test duration
	tmpFile := filepath.Join(t.TempDir(), "jivetalking_silence_trim_demo_paths.txt")
	err = os.WriteFile(tmpFile, []byte(marker), 0600)
	if err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	if err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
}
