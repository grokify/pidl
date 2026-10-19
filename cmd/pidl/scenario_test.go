package main

import (
	"bytes"
	"os"
	"path/filepath"
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
