package main

import (
	"path/filepath"
	"testing"
)

func TestRunOutputPathIncludesRunID(t *testing.T) {
	got := runOutputPath("data/raw", "tokopedia.json", 42)
	want := filepath.Join("data/raw", "run-42-tokopedia.json")
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestBuildJobsIncludesBothMarketplaceScrapers(t *testing.T) {
	jobs, err := buildJobs("tokopedia,shopee", "beras 5 kg", "surabaya", "", 20, 0.07)
	if err != nil {
		t.Fatalf("build jobs: %v", err)
	}
	if len(jobs) != 2 {
		t.Fatalf("expected two jobs, got %d", len(jobs))
	}
	for _, job := range jobs {
		if !job.paid || job.maxItems != 20 || job.maxChargeUSD != 0.07 {
			t.Fatalf("job does not preserve paid API limits: %#v", job)
		}
		if job.adapt == nil {
			t.Fatalf("marketplace job %s has no raw product adapter", job.name)
		}
	}
}
