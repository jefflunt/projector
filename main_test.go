package main

import (
	"projector/config"
	"testing"
)

func TestPrepareProjectList(t *testing.T) {
	cfg := &config.Config{
		Projects: map[string]config.ProjectDetails{
			"b": {Starred: false, Show: true},
			"a": {Starred: true, Show: true},
			"c": {Starred: false, Show: false},
		},
	}

	m := &model{config: cfg}
	m.prepareProjectList()

	if len(m.projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(m.projects))
	}

	// Starred should come first
	if m.projects[0].name != "a" {
		t.Errorf("expected 'a' to be first (starred), got %s", m.projects[0].name)
	}

	// Then unstarred alphabetically
	if m.projects[1].name != "b" {
		t.Errorf("expected 'b' to be second, got %s", m.projects[1].name)
	}
}
