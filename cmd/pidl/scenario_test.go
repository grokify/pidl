package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// scenarioFixture is a small protocol with two main scenarios and a
// fragment, used to exercise --scenario handling.
const scenarioFixture = `{
  "protocol": {"id": "demo", "name": "Demo"},
  "entities": [
    {"id": "person", "name": "Person", "type": "user"},
    {"id": "agent", "name": "Agent", "type": "agent"},
    {"id": "resource", "name": "Resource", "type": "resource_server"},
    {"id": "ps", "name": "Person Server", "type": "authorization_server"}
  ],
  "scenarios": [
    {"id": "agent-identity", "name": "Agent identity", "flows": [
      {"from": "agent", "to": "resource", "action": "signed_request"},
      {"from": "resource", "to": "agent", "action": "response", "mode": "response"}
    ]},
    {"id": "consent", "name": "Consent", "fragment": true, "flows": [
      {"from": "ps", "to": "person", "action": "consent_prompt"},
      {"from": "person", "to": "ps", "action": "approve", "mode": "response"}
    ]},
    {"id": "three-party", "name": "PS authorization", "flows": [
      {"from": "agent", "to": "ps", "action": "auth_token_request"},
      {"include": "consent"},
      {"from": "ps", "to": "agent", "action": "auth_token", "mode": "response"}
    ]}
  ]
}`

// writeScenarioFixture writes scenarioFixture to dir and returns its path.
func writeScenarioFixture(t *testing.T, dir string) string {
	t.Helper()
	path := filepath.Join(dir, "demo_scenarios.json")
	if err := os.WriteFile(path, []byte(scenarioFixture), 0600); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
	return path
}

func TestRunScenarios(t *testing.T) {
	dir := t.TempDir()
	demo := writeScenarioFixture(t, dir)
	legacy := writeExample(t, dir, "oauth2_pkce")

	cases := []runCase{
		{name: "generate requires scenario", args: []string{"generate", "-f", "mermaid", demo}, wantCode: 1,
			wantStderr: []string{"demo has scenarios; choose one with --scenario: agent-identity, consent, three-party"}},
		{name: "generate unknown scenario", args: []string{"generate", "-f", "mermaid", "--scenario", "nope", demo}, wantCode: 1,
			wantStderr: []string{`unknown scenario "nope"`}},
		{name: "generate one scenario", args: []string{"generate", "-f", "mermaid", "--scenario", "agent-identity", demo}, wantCode: 0,
			wantStdout: []string{"participant agent", "participant resource", "signed_request"}},
		{name: "generate expands fragments for formats without include boxes", args: []string{"generate", "-f", "dot", "--scenario", "three-party", demo}, wantCode: 0,
			wantStdout: []string{"consent_prompt", "auth_token"}},
		{name: "generate draws includes as boxes", args: []string{"generate", "-f", "mermaid", "--scenario", "three-party", demo}, wantCode: 0,
			wantStdout: []string{"note over person,ps: ref: Consent", "auth_token"}},
		{name: "generate expand flag", args: []string{"generate", "-f", "plantuml", "--scenario", "three-party", "--expand", demo}, wantCode: 0,
			wantStdout: []string{"consent_prompt", "approve", "auth_token"}},
		{name: "generate legacy file rejects scenario", args: []string{"generate", "-f", "mermaid", "--scenario", "x", legacy}, wantCode: 1,
			wantStderr: []string{"has no scenarios"}},
		{name: "generate legacy file unchanged", args: []string{"generate", "-f", "mermaid", legacy}, wantCode: 0,
			wantStdout: []string{"sequenceDiagram"}},
		{name: "simulate requires scenario", args: []string{"simulate", demo}, wantCode: 1,
			wantStderr: []string{"choose one with --scenario"}},
		{name: "simulate one scenario expands fragments", args: []string{"simulate", "-v", "--scenario", "three-party", demo}, wantCode: 0,
			wantStdout: []string{"consent_prompt", "auth_token"}},
		{name: "debug requires scenario", args: []string{"debug", demo}, wantCode: 1,
			wantStderr: []string{"choose one with --scenario"}},
		{name: "debug one scenario", args: []string{"debug", "--scenario", "three-party", demo}, stdin: "list\nquit\n", wantCode: 0,
			wantStdout: []string{"Flows: 4", "consent_prompt"}},
		{name: "scenario matrix covers the whole file", args: []string{"generate", "-f", "scenario-matrix", demo}, wantCode: 0,
			wantStdout: []string{"Roles by Scenario", "1. Agent identity", "2. PS authorization", "3. Consent (fragment)"}},
		{name: "scenario matrix rejects --scenario", args: []string{"generate", "-f", "scenario-matrix", "--scenario", "three-party", demo}, wantCode: 1,
			wantStderr: []string{"covers every scenario"}},
		{name: "token table covers the whole file", args: []string{"generate", "-f", "token-table", demo}, wantCode: 0,
			wantStdout: []string{"Demo - Tokens", "*No tokens defined.*"}},
		{name: "outline covers the whole file", args: []string{"generate", "-f", "outline", demo}, wantCode: 0,
			wantStdout: []string{"## Roles", "### 2. PS authorization", "2. See *Consent* (scenario 3)", "### 3. Consent (fragment)"}},
		{name: "outline json", args: []string{"generate", "-f", "outline-json", demo}, wantCode: 0,
			wantStdout: []string{`"include_position": 3`}},
		{name: "walkthrough json covers the whole file", args: []string{"generate", "-f", "walkthrough-json", demo}, wantCode: 0,
			wantStdout: []string{`"fragments": [`, `"Consent"`, `"location": "scenarios[1].flows[0] (consent)"`}},
		{name: "scenarios table", args: []string{"scenarios", demo}, wantCode: 0,
			wantStdout: []string{"Scenarios in Demo", "1   agent-identity", "2   three-party", "3   consent", "fragment", "person, ps"}},
		{name: "scenarios json", args: []string{"scenarios", "-f", "json", demo}, wantCode: 0,
			wantStdout: []string{`"position": 2`, `"includes": [`, `"fragment": true`}},
		{name: "scenarios legacy file", args: []string{"scenarios", legacy}, wantCode: 0,
			wantStdout: []string{"has no scenarios; it has a single flow list"}},
		{name: "scenarios legacy json", args: []string{"scenarios", "-f", "json", legacy}, wantCode: 0,
			wantStdout: []string{"[]"}},
		{name: "scenarios no file", args: []string{"scenarios"}, wantCode: 1,
			wantStdout: []string{"Usage: pidl scenarios"}},
		{name: "analyze whole file", args: []string{"analyze", "-q", demo}, wantCode: 0,
			wantStdout: []string{"Score:"}},
		{name: "analyze one scenario", args: []string{"analyze", "-f", "json", "--scenario", "agent-identity", demo}, wantCode: 0,
			wantStdout: []string{`"summary"`}},
		{name: "analyze unknown scenario", args: []string{"analyze", "--scenario", "nope", demo}, wantCode: 1,
			wantStderr: []string{`unknown scenario "nope"`}},
	}

	origStdin := stdin
	t.Cleanup(func() { stdin = origStdin })
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			stdin = strings.NewReader(tc.stdin)
			var out, errOut bytes.Buffer
			code := run(tc.args, &out, &errOut)
			if code != tc.wantCode {
				t.Fatalf("run(%q) = %d, want %d\nstdout:\n%s\nstderr:\n%s", tc.args, code, tc.wantCode, out.String(), errOut.String())
			}
			for _, want := range tc.wantStdout {
				if !strings.Contains(out.String(), want) {
					t.Errorf("stdout missing %q\nstdout:\n%s", want, out.String())
				}
			}
			for _, want := range tc.wantStderr {
				if !strings.Contains(errOut.String(), want) {
					t.Errorf("stderr missing %q\nstderr:\n%s", want, errOut.String())
				}
			}
		})
	}
}

func TestRunAllScenarios(t *testing.T) {
	dir := t.TempDir()
	demo := writeScenarioFixture(t, dir)
	legacy := writeExample(t, dir, "oauth2_pkce")

	listDir := func(t *testing.T, d string) []string {
		t.Helper()
		entries, err := os.ReadDir(d)
		if err != nil {
			t.Fatalf("ReadDir(%s): %v", d, err)
		}
		var names []string
		for _, e := range entries {
			names = append(names, e.Name())
		}
		return names
	}

	tests := []struct {
		name      string
		args      []string
		outDir    string
		wantCode  int
		wantFiles []string
		wantErr   string
	}{
		{
			name:      "main scenarios in document order",
			args:      []string{"generate", "-f", "mermaid", "--all-scenarios"},
			outDir:    "main",
			wantFiles: []string{"01-agent-identity.mmd", "02-three-party.mmd"},
		},
		{
			name:      "fragments after main scenarios",
			args:      []string{"generate", "-f", "d2", "--all-scenarios", "--include-fragments"},
			outDir:    "frags",
			wantFiles: []string{"01-agent-identity.d2", "02-three-party.d2", "03-consent.d2"},
		},
		{
			name:     "needs an output directory",
			args:     []string{"generate", "-f", "mermaid", "--all-scenarios"},
			wantCode: 1,
			wantErr:  "needs an output directory",
		},
		{
			name:     "cannot combine with --scenario",
			args:     []string{"generate", "-f", "mermaid", "--all-scenarios", "--scenario", "three-party"},
			outDir:   "combined",
			wantCode: 1,
			wantErr:  "cannot be combined",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			args := append([]string{}, tt.args...)
			out := ""
			if tt.outDir != "" {
				out = filepath.Join(dir, tt.outDir)
				args = append(args, "-o", out)
			}
			args = append(args, demo)
			var stdout, stderr bytes.Buffer
			code := run(args, &stdout, &stderr)
			if code != tt.wantCode {
				t.Fatalf("run(%q) = %d, want %d\nstderr: %s", args, code, tt.wantCode, stderr.String())
			}
			if tt.wantErr != "" && !strings.Contains(stderr.String(), tt.wantErr) {
				t.Fatalf("stderr = %q, want %q", stderr.String(), tt.wantErr)
			}
			if tt.wantFiles != nil {
				if got := listDir(t, out); !reflect.DeepEqual(got, tt.wantFiles) {
					t.Fatalf("files = %v, want %v", got, tt.wantFiles)
				}
			}
		})
	}

	// The collapsed include box appears in the scenario that includes the fragment.
	data, err := os.ReadFile(filepath.Join(dir, "main", "02-three-party.mmd"))
	if err != nil {
		t.Fatalf("reading three-party output: %v", err)
	}
	if !strings.Contains(string(data), "ref: Consent") {
		t.Fatalf("three-party output missing include box:\n%s", data)
	}

	// A legacy file has no scenarios to render.
	var stdout, stderr bytes.Buffer
	if code := run([]string{"generate", "--all-scenarios", "-o", filepath.Join(dir, "legacy"), legacy}, &stdout, &stderr); code != 1 ||
		!strings.Contains(stderr.String(), "has no scenarios") {
		t.Fatalf("legacy --all-scenarios: code %d, stderr %q", code, stderr.String())
	}
}
