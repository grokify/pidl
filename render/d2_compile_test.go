package render

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/grokify/pidl"
	"github.com/grokify/pidl/examples"
)

// keyedMessage matches the `seqN:` / `noteN:` prefixes that made D2 read a
// message as a labelled shape.
var keyedMessage = regexp.MustCompile(`(?m)^\s*(seq|note)\d+:`)

// awkwardProtocol exercises every D2 output path that once produced text D2
// could not compile: conditions, annotations, alternatives, responses, notes,
// and labels holding characters that are significant in unquoted D2 strings.
func awkwardProtocol() *pidl.Protocol {
	return &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:      "awkward",
			Name:    "Awkward: labels; with [brackets] and {braces} #1",
			Version: "1.0",
		},
		Entities: []pidl.Entity{
			{ID: "agent", Name: "Agent (local@domain)", Type: pidl.EntityTypeAgent},
			{ID: "ps", Name: "Person Server: chosen by the person", Type: pidl.EntityTypeAuthorizationServer},
			{ID: "res", Name: "Resource \"quoted\"", Type: pidl.EntityTypeResourceServer},
		},
		Phases: []pidl.Phase{
			{ID: "p1", Name: "Phase one: tokens [a, b]"},
			{ID: "p2", Name: "Phase two", Parent: "p1"},
		},
		Flows: []pidl.Flow{
			{
				From: "agent", To: "ps", Action: "req", Phase: "p1",
				Label: "POST /token with scope=read; grant_type=code",
				Mode:  pidl.FlowModeRequest,
				Note:  "A note: with a colon",
				Annotations: []pidl.Annotation{
					{Type: pidl.AnnotationTypeSecurity, Text: "Bearer: possession is authority [RFC 6750]"},
				},
			},
			{
				From: "ps", To: "agent", Action: "resp", Phase: "p1",
				Label: "200 OK {token}",
				Mode:  pidl.FlowModeResponse,
			},
			{
				From: "agent", To: "res", Action: "call", Phase: "p2",
				Label:     "GET /.well-known/aauth-resource.json",
				Mode:      pidl.FlowModeRequest,
				Condition: "keys_not_cached",
				Alternatives: []pidl.Alternative{
					{
						Condition: "user_denies: [deny]",
						Flows: []pidl.Flow{
							{From: "res", To: "agent", Action: "deny", Label: "403 Forbidden", Mode: pidl.FlowModeResponse},
						},
					},
				},
			},
			{
				From: "res", To: "res", Action: "verify", Phase: "p2",
				Label: "Verify signature",
				Mode:  pidl.FlowModeRequest,
			},
			{
				From: "res", To: "agent", Action: "evt", Phase: "p2",
				Label: "event: key rotated",
				Mode:  pidl.FlowModeEvent,
			},
		},
	}
}

func TestD2SequenceEmitsEdgesNotShapes(t *testing.T) {
	s, err := NewD2().RenderString(awkwardProtocol())
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}
	if m := keyedMessage.FindString(s); m != "" {
		t.Errorf("D2 sequence output must not key messages or notes (%q):\n%s", strings.TrimSpace(m), s)
	}
	if strings.Contains(s, "<-") {
		t.Errorf("D2 sequence output must not reverse arrows:\n%s", s)
	}
	for _, want := range []string{
		`agent -> ps: "POST /token with scope=read; grant_type=code"`,
		`ps -> agent: "200 OK {token}"`,
		"style.stroke-dash: 3",
		`ps."⚠️ SECURITY: Bearer: possession is authority [RFC 6750]"`,
		`agent -> res: "[keys_not_cached] GET /.well-known/aauth-resource.json"`,
		`res: "Resource \"quoted\""`,
	} {
		if !strings.Contains(s, want) {
			t.Errorf("D2 sequence output missing %q:\n%s", want, s)
		}
	}
}

// TestD2OutputCompiles renders every built-in example and the awkward protocol in
// each D2 style and asks the d2 CLI to compile the result. It is skipped when d2
// is not installed, so CI without d2 still passes the string-level tests above.
func TestD2OutputCompiles(t *testing.T) {
	d2Bin, err := exec.LookPath("d2")
	if err != nil {
		t.Skip("d2 not installed; skipping compile check")
	}

	protocols := map[string]*pidl.Protocol{"awkward": awkwardProtocol()}
	for _, name := range examples.List() {
		ex, err := examples.Get(name)
		if err != nil {
			t.Fatalf("examples.Get(%q) error = %v", name, err)
		}
		p, err := ex.Protocol()
		if err != nil {
			t.Fatalf("example %q: %v", name, err)
		}
		protocols[name] = p
	}

	renderers := map[string]func() *D2Renderer{
		"sequence": NewD2,
		"flow":     NewD2Flow,
		"arch":     NewD2Arch,
	}

	dir := t.TempDir()
	for pname, p := range protocols {
		for rname, newR := range renderers {
			t.Run(pname+"/"+rname, func(t *testing.T) {
				src, err := newR().RenderString(p)
				if err != nil {
					t.Fatalf("RenderString() error = %v", err)
				}
				base := strings.ReplaceAll(pname, "/", "_") + "-" + rname
				in := filepath.Join(dir, base+".d2")
				out := filepath.Join(dir, base+".svg")
				if err := os.WriteFile(in, []byte(src), 0o600); err != nil {
					t.Fatal(err)
				}
				ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
				defer cancel()
				cmd := exec.CommandContext(ctx, d2Bin, in, out) //nolint:gosec // test-only; paths are from t.TempDir()
				if output, err := cmd.CombinedOutput(); err != nil {
					t.Fatalf("d2 failed to compile %s output: %v\n%s\n--- source ---\n%s", rname, err, output, src)
				}
			})
		}
	}
}
