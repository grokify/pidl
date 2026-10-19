package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/grokify/pidl/examples"
)

// writeExample writes the named built-in example to dir and returns its path.
func writeExample(t *testing.T, dir, name string) string {
	t.Helper()
	data, err := examples.GetJSON(name)
	if err != nil {
		t.Fatalf("examples.GetJSON(%q): %v", name, err)
	}
	path := filepath.Join(dir, filepath.Base(name)+".json")
	if err := os.WriteFile(path, data, 0600); err != nil {
		t.Fatalf("writing %s: %v", path, err)
	}
	return path
}

// runCase is one table entry for TestRun.
type runCase struct {
	name       string
	args       []string
	stdin      string
	wantCode   int
	wantStdout []string // substrings that must appear on stdout
	wantStderr []string // substrings that must appear on stderr
	after      func(t *testing.T)
}

func TestRun(t *testing.T) {
	dir := t.TempDir()
	authCode := writeExample(t, dir, "oauth2_authorization_code")
	pkce := writeExample(t, dir, "oauth2_pkce")
	withStates := writeExample(t, dir, "oauth2_with_states")
	composed := writeExample(t, dir, "oauth_mcp_composed")
	oaiaf := writeExample(t, dir, "oaiaf_architecture")
	missing := filepath.Join(dir, "does-not-exist.json")
	genOut := filepath.Join(dir, "gen-out.mmd")
	resolveOut := filepath.Join(dir, "resolved.json")
	diffOut := filepath.Join(dir, "diff.md")
	analyzeOut := filepath.Join(dir, "analysis.json")
	traceOut := filepath.Join(dir, "trace.svg")
	initNew := filepath.Join(dir, "my-new-protocol.json")
	initFrom := filepath.Join(dir, "from-example.json")

	fileExists := func(path string) func(t *testing.T) {
		return func(t *testing.T) {
			t.Helper()
			if _, err := os.Stat(path); err != nil {
				t.Fatalf("expected %s to exist: %v", path, err)
			}
		}
	}
	validates := func(path string) func(t *testing.T) {
		return func(t *testing.T) {
			t.Helper()
			var out, errOut bytes.Buffer
			if code := run([]string{"validate", path}, &out, &errOut); code != 0 {
				t.Fatalf("validate %s: exit %d, stderr=%q", path, code, errOut.String())
			}
		}
	}

	cases := []runCase{
		// Dispatch
		{name: "no args", args: nil, wantCode: 1, wantStderr: []string{"Usage:"}},
		{name: "version", args: []string{"version"}, wantCode: 0, wantStdout: []string{"pidl version 0.7.0"}},
		{name: "version short flag", args: []string{"-v"}, wantCode: 0, wantStdout: []string{"pidl version 0.7.0"}},
		{name: "help", args: []string{"help"}, wantCode: 0, wantStdout: []string{"Usage:", "pidl <command>"}},
		{name: "unknown command", args: []string{"bogus"}, wantCode: 1, wantStderr: []string{"Unknown command: bogus", "Usage:"}},

		// validate
		{name: "validate ok", args: []string{"validate", authCode}, wantCode: 0, wantStdout: []string{"valid (OAuth 2.0 Authorization Code Flow)"}},
		{name: "validate quiet", args: []string{"validate", "-q", authCode}, wantCode: 0},
		{name: "validate missing file", args: []string{"validate", missing}, wantCode: 1, wantStderr: []string{"parse error"}},
		{name: "validate no file", args: []string{"validate"}, wantCode: 1, wantStdout: []string{"Usage: pidl validate"}},
		{name: "validate help flag", args: []string{"validate", "-h"}, wantCode: 0, wantStdout: []string{"Usage: pidl validate"}},

		// generate
		{name: "generate mermaid", args: []string{"generate", "-f", "mermaid", authCode}, wantCode: 0, wantStdout: []string{"sequenceDiagram"}},
		{name: "generate d2", args: []string{"generate", "-f", "d2", authCode}, wantCode: 0, wantStdout: []string{"sequence_diagram"}},
		{name: "generate plantuml default", args: []string{"gen", authCode}, wantCode: 0, wantStdout: []string{"@startuml"}},
		{name: "generate mermaid-state", args: []string{"generate", "-f", "mermaid-state", "--entity=client", withStates}, wantCode: 0, wantStdout: []string{"stateDiagram"}},
		{name: "generate svg", args: []string{"generate", "-f", "svg", "--template", "minimal", "--theme", "dark", authCode}, wantCode: 0, wantStdout: []string{"<svg"}},
		{name: "generate svg-animated", args: []string{"generate", "-f", "svg-animated", authCode}, wantCode: 0, wantStdout: []string{"<svg"}},
		{name: "generate svg-network with boundary", args: []string{"generate", "-f", "svg-network", "--boundary=dmz:client", authCode}, wantCode: 0, wantStdout: []string{"<svg"}},
		{name: "generate svg-network bad boundary", args: []string{"generate", "-f", "svg-network", "--boundary=nocolon", authCode}, wantCode: 1, wantStderr: []string{"Error parsing boundary flag"}},
		{name: "generate infographic", args: []string{"generate", "-f", "infographic", "--ig-size=datasheet-tile", "--ig-theme=tech", "--ig-direction=vertical", "--ig-no-animate", "--ig-no-shapes", authCode}, wantCode: 0, wantStdout: []string{"<svg"}},
		{name: "generate by example name", args: []string{"generate", "-f", "mermaid", "oauth2_pkce"}, wantCode: 0, wantStdout: []string{"sequenceDiagram"}},
		{name: "generate composed resolves imports", args: []string{"generate", "-f", "mermaid", composed}, wantCode: 1, wantStderr: []string{"Error resolving"}},
		{name: "generate to file", args: []string{"generate", "-f", "mermaid", "-o", genOut, authCode}, wantCode: 0, wantStdout: []string{"Wrote " + genOut}, after: fileExists(genOut)},
		{name: "generate bogus format", args: []string{"generate", "-f", "bogus", authCode}, wantCode: 1, wantStderr: []string{"Error:"}},
		{name: "generate missing file", args: []string{"generate", missing}, wantCode: 1, wantStderr: []string{"Error parsing"}},
		{name: "generate bad template", args: []string{"generate", "-f", "svg", "--template", "nope", authCode}, wantCode: 1, wantStderr: []string{"Error loading template"}},
		{name: "generate no file", args: []string{"generate"}, wantCode: 1, wantStdout: []string{"Usage: pidl generate"}},
		{name: "generate bad flag", args: []string{"generate", "-nosuchflag", "x"}, wantCode: 2, wantStderr: []string{"flag provided but not defined"}},

		// examples
		{name: "examples list", args: []string{"examples"}, wantCode: 0, wantStdout: []string{"Available examples:", "oauth2_authorization_code", "oauth2_pkce"}},
		{name: "examples show", args: []string{"examples", "oauth2_authorization_code"}, wantCode: 0, wantStdout: []string{"Example: oauth2_authorization_code", "Protocol: OAuth 2.0 Authorization Code Flow", "Entities:"}},
		{name: "examples json", args: []string{"examples", "-json", "oauth2_authorization_code"}, wantCode: 0, wantStdout: []string{`"$schema"`}},
		{name: "examples unknown", args: []string{"examples", "nope"}, wantCode: 1, wantStderr: []string{"Error:"}},
		{name: "examples unknown json", args: []string{"examples", "-json", "nope"}, wantCode: 1, wantStderr: []string{"Error:"}},

		// init
		{name: "init minimal", args: []string{"init", initNew}, wantCode: 0, wantStdout: []string{"Created " + initNew, "Protocol: My New Protocol"}, after: validates(initNew)},
		{name: "init from example", args: []string{"init", "-from", "oauth2_authorization_code", "-name", "Copied", initFrom}, wantCode: 0, wantStdout: []string{"Protocol: Copied"}, after: validates(initFrom)},
		{name: "init existing file", args: []string{"init", authCode}, wantCode: 1, wantStderr: []string{"file already exists"}},
		{name: "init unknown example", args: []string{"init", "-from", "nope", filepath.Join(dir, "x.json")}, wantCode: 1, wantStderr: []string{"Error:"}},
		{name: "init no file", args: []string{"init"}, wantCode: 1, wantStdout: []string{"Usage: pidl init"}},

		// roles / components / trust
		{name: "roles table", args: []string{"roles", oaiaf}, wantCode: 0, wantStdout: []string{"Protocol Roles in", "ENTITY ID", "oauth"}},
		{name: "roles table filtered", args: []string{"roles", "-protocol=scim", oaiaf}, wantCode: 0, wantStdout: []string{"scim"}},
		{name: "roles json filtered", args: []string{"roles", "-f", "json", "-protocol=oauth", oaiaf}, wantCode: 0, wantStdout: []string{`"entity_id"`, `"protocol": "oauth"`}},
		{name: "roles none", args: []string{"roles", authCode}, wantCode: 0, wantStdout: []string{"Protocol Roles in"}},
		{name: "roles missing", args: []string{"roles", missing}, wantCode: 1, wantStderr: []string{"Error:"}},
		{name: "roles no file", args: []string{"roles"}, wantCode: 1, wantStdout: []string{"Usage: pidl roles"}},
		{name: "components table", args: []string{"components", oaiaf}, wantCode: 0, wantStdout: []string{"Deployment Components in", "Identity Provider", "Total:"}},
		{name: "components json filtered", args: []string{"components", "-f", "json", "-type=idp", oaiaf}, wantCode: 0, wantStdout: []string{`"idp_component"`}},
		{name: "components none", args: []string{"components", authCode}, wantCode: 0, wantStdout: []string{"No components defined."}},
		{name: "components missing", args: []string{"components", missing}, wantCode: 1, wantStderr: []string{"Error:"}},
		{name: "components no file", args: []string{"components"}, wantCode: 1, wantStdout: []string{"Usage: pidl components"}},
		{name: "trust table", args: []string{"trust", oaiaf}, wantCode: 0, wantStdout: []string{"Trust Relationships in", "Total:"}},
		{name: "trust json filtered", args: []string{"trust", "-f", "json", "-type=issues", oaiaf}, wantCode: 0, wantStdout: []string{`"issues"`}},
		{name: "trust none", args: []string{"trust", authCode}, wantCode: 0, wantStdout: []string{"No trust relationships defined."}},
		{name: "trust missing", args: []string{"trust", missing}, wantCode: 1, wantStderr: []string{"Error:"}},
		{name: "trust no file", args: []string{"trust"}, wantCode: 1, wantStdout: []string{"Usage: pidl trust"}},

		// resolve
		{name: "resolve stdout", args: []string{"resolve", authCode}, wantCode: 0, wantStdout: []string{`"protocol": {`, `"id": "oauth2-authorization-code"`}, wantStderr: []string{"has no imports or extends to resolve"}},
		{name: "resolve to file no validate", args: []string{"resolve", "--validate=false", "-o", resolveOut, authCode}, wantCode: 0, wantStdout: []string{"Wrote resolved protocol to"}, after: fileExists(resolveOut)},
		{name: "resolve missing", args: []string{"resolve", missing}, wantCode: 1, wantStderr: []string{"Error parsing"}},
		{name: "resolve no file", args: []string{"resolve"}, wantCode: 1, wantStdout: []string{"Usage: pidl resolve"}},

		// analyze
		{name: "analyze text", args: []string{"analyze", authCode}, wantCode: 0, wantStdout: []string{"Security Analysis"}},
		{name: "analyze quiet", args: []string{"analyze", "-q", authCode}, wantCode: 0, wantStdout: []string{"Score:"}},
		{name: "analyze json to file", args: []string{"analyze", "-f", "json", "-o", analyzeOut, authCode}, wantCode: 0, wantStdout: []string{"Wrote analysis to"}, after: fileExists(analyzeOut)},
		{name: "analyze markdown category", args: []string{"analyze", "-f", "markdown", "--category", "authentication", "--min-severity", "low", authCode}, wantCode: 0, wantStdout: []string{"#"}},
		{name: "analyze fail-on", args: []string{"analyze", "-q", "--fail-on", "info", authCode}, wantCode: 1, wantStdout: []string{"Risks:"}},
		{name: "analyze bad fail-on", args: []string{"analyze", "-q", "--fail-on", "bogus", authCode}, wantCode: 1, wantStderr: []string{"Invalid fail-on severity"}},
		{name: "analyze bad severity", args: []string{"analyze", "--min-severity", "bogus", authCode}, wantCode: 1, wantStderr: []string{"Invalid severity"}},
		{name: "analyze bad category", args: []string{"analyze", "--category", "bogus", authCode}, wantCode: 1, wantStderr: []string{"Invalid category"}},
		{name: "analyze missing", args: []string{"analyze", missing}, wantCode: 1, wantStderr: []string{"Error:"}},
		{name: "analyze no file", args: []string{"analyze"}, wantCode: 1, wantStdout: []string{"Usage: pidl analyze"}},

		// simulate
		{name: "simulate text", args: []string{"simulate", withStates}, wantCode: 0, wantStdout: []string{"Step"}},
		{name: "simulate alias run json", args: []string{"run", "-json", withStates}, wantCode: 0, wantStdout: []string{"{"}},
		{name: "simulate verbose", args: []string{"sim", "-v", withStates}, wantCode: 0, wantStdout: []string{"✓ Step 1:", "Status: completed", "Final states:"}},
		{name: "simulate steps verbose", args: []string{"simulate", "-steps=2", "-v", withStates}, wantCode: 0, wantStdout: []string{"Steps executed: 2", "Status: partial"}},
		{name: "simulate steps", args: []string{"simulate", "-steps=2", "--show-timings", withStates}, wantCode: 0, wantStdout: []string{"Step"}},
		{name: "simulate mermaid", args: []string{"simulate", "--trace-format=mermaid", withStates}, wantCode: 0, wantStdout: []string{"sequenceDiagram"}},
		{name: "simulate svg to file", args: []string{"simulate", "--trace-format=svg", "--trace-output", traceOut, withStates}, wantCode: 0, wantStdout: []string{"Wrote trace to"}, after: fileExists(traceOut)},
		{name: "simulate missing", args: []string{"simulate", missing}, wantCode: 1, wantStderr: []string{"Error:"}},
		{name: "simulate no file", args: []string{"simulate"}, wantCode: 1, wantStdout: []string{"Usage: pidl simulate"}},

		// diff
		{name: "diff text", args: []string{"diff", authCode, pkce}, wantCode: 0, wantStdout: []string{"PKCE"}},
		{name: "diff quiet", args: []string{"diff", "-q", authCode, pkce}, wantCode: 0, wantStdout: []string{"Changes:"}},
		{name: "diff quiet identical", args: []string{"diff", "-q", authCode, authCode}, wantCode: 0, wantStdout: []string{"No differences found."}},
		{name: "diff json", args: []string{"diff", "-f", "json", "--ignore-metadata", "--ignore-descriptions", authCode, pkce}, wantCode: 0, wantStdout: []string{"{"}},
		{name: "diff markdown to file", args: []string{"diff", "-f", "markdown", "-o", diffOut, authCode, pkce}, wantCode: 0, wantStdout: []string{"Wrote diff to"}, after: fileExists(diffOut)},
		{name: "diff missing base", args: []string{"diff", missing, pkce}, wantCode: 1, wantStderr: []string{"Error loading base file"}},
		{name: "diff missing new", args: []string{"diff", authCode, missing}, wantCode: 1, wantStderr: []string{"Error loading new file"}},
		{name: "diff one arg", args: []string{"diff", authCode}, wantCode: 1, wantStdout: []string{"Usage: pidl diff"}},

		// debug (REPL reads from the package-level stdin)
		{
			name:     "debug scripted session",
			args:     []string{"debug", withStates},
			stdin:    "help\nlist\nstep\ninspect\ninspect entity client\ninspect flow 1\ninspect bogus\nbreak 3\nbp\ncontinue\ndelete 3\nbp\nset client idle\ntrace\nreset\nbogus\nquit\n",
			wantCode: 0,
			wantStdout: []string{
				"PIDL Debugger:", "Commands:", "Step 1:", "Entity:", "Flow 1:",
				"Usage: inspect", "Breakpoint set at flow 3", "Breakpoint hit at flow 3",
				"Breakpoint removed at flow 3", "No breakpoints set.", "Set client state to idle",
				"Steps executed:", "Execution reset to beginning.", "Unknown command: bogus", "Goodbye.",
			},
		},
		{
			name:       "debug error paths",
			args:       []string{"debug", withStates},
			stdin:      "break\nbreak x\ndelete\ndelete x\ninspect flow x\ninspect flow 999\ninspect entity nope\nset\nset nope x\n\n   \nq\n",
			wantCode:   0,
			wantStdout: []string{"Usage: break", "Invalid flow index: x", "Usage: delete", "Usage: set", "Error:", "Goodbye."},
		},
		{name: "debug runs to completion on continue", args: []string{"debug", withStates}, stdin: "c\n(pidl:done)\nexit\n", wantCode: 0, wantStdout: []string{"Execution complete.", "(pidl:done) "}},
		{name: "debug eof without newline", args: []string{"debug", withStates}, stdin: "step", wantCode: 0, wantStdout: []string{"Step 1:"}},
		{name: "debug empty stdin", args: []string{"debug", withStates}, stdin: "", wantCode: 0, wantStdout: []string{"PIDL Debugger:"}},
		{name: "debug missing", args: []string{"debug", missing}, wantCode: 1, wantStderr: []string{"Error:"}},
		{name: "debug no file", args: []string{"debug"}, wantCode: 1, wantStdout: []string{"Usage: pidl debug"}},
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
			if tc.after != nil {
				tc.after(t)
			}
		})
	}
}

func TestParseBoundaryFlag(t *testing.T) {
	cases := []struct {
		in       string
		wantID   string
		wantIDs  []string
		wantErr  bool
		scenario string
	}{
		{in: "dmz:auth,gateway", wantID: "dmz", wantIDs: []string{"auth", "gateway"}, scenario: "two entities"},
		{in: " dmz : a , ,b ", wantID: "dmz", wantIDs: []string{"a", "b"}, scenario: "whitespace and empty entries trimmed"},
		{in: "nocolon", wantErr: true, scenario: "missing colon"},
		{in: ":a", wantErr: true, scenario: "empty boundary id"},
		{in: "dmz:", wantErr: true, scenario: "empty entity list"},
		{in: "dmz: , ", wantErr: true, scenario: "only separators"},
	}
	for _, tc := range cases {
		t.Run(tc.scenario, func(t *testing.T) {
			id, ids, err := parseBoundaryFlag(tc.in)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("parseBoundaryFlag(%q) expected error, got id=%q ids=%v", tc.in, id, ids)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseBoundaryFlag(%q): %v", tc.in, err)
			}
			if id != tc.wantID || strings.Join(ids, ",") != strings.Join(tc.wantIDs, ",") {
				t.Errorf("parseBoundaryFlag(%q) = %q, %v; want %q, %v", tc.in, id, ids, tc.wantID, tc.wantIDs)
			}
		})
	}
}

func TestTruncate(t *testing.T) {
	cases := []struct {
		in, want string
		max      int
	}{
		{in: "short", max: 10, want: "short"},
		{in: "exactly10!", max: 10, want: "exactly10!"},
		{in: "this is too long", max: 10, want: "this is..."},
		{in: "abcdef", max: 3, want: "abc"},
	}
	for _, tc := range cases {
		if got := truncate(tc.in, tc.max); got != tc.want {
			t.Errorf("truncate(%q, %d) = %q, want %q", tc.in, tc.max, got, tc.want)
		}
	}
}

func TestFlagTypes(t *testing.T) {
	var b boundaryFlags
	var c categoryFlags
	for _, v := range []string{"x", "y"} {
		if err := b.Set(v); err != nil {
			t.Fatal(err)
		}
		if err := c.Set(v); err != nil {
			t.Fatal(err)
		}
	}
	if b.String() != "x, y" || c.String() != "x, y" {
		t.Errorf("String() = %q / %q, want %q", b.String(), c.String(), "x, y")
	}
}
