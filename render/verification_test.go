package render

import (
	"strings"
	"testing"

	"github.com/grokify/pidl"
)

// verifiedProtocol returns a protocol containing a flow with verification provenance.
func verifiedProtocol() *pidl.Protocol {
	return &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:       "attack",
			Name:     "Attack Flow",
			Category: pidl.CategoryAgent,
		},
		Entities: []pidl.Entity{
			{ID: "attacker", Name: "Attacker", Type: pidl.EntityTypeAgent},
			{ID: "target", Name: "Target", Type: pidl.EntityTypeServer},
		},
		Flows: []pidl.Flow{
			{
				From:   "attacker",
				To:     "target",
				Action: "exploit",
				Label:  "Exploit",
				Mode:   pidl.FlowModeRequest,
				Verification: &pidl.FlowVerification{
					Tier:     pidl.VerificationTierCorroborated,
					Source:   "METR",
					Citation: "https://example.com/report",
				},
			},
		},
	}
}

func TestD2RendererVerificationBadge(t *testing.T) {
	p := verifiedProtocol()
	r := NewD2()

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	if !strings.Contains(s, "🔎 corroborated — METR") {
		t.Errorf("D2 output missing verification badge; got:\n%s", s)
	}
}

func TestD2RendererVerificationDisabled(t *testing.T) {
	p := verifiedProtocol()
	r := NewD2()
	r.ShowVerification = false

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	if strings.Contains(s, "🔎") {
		t.Errorf("D2 output should not contain verification badge when disabled; got:\n%s", s)
	}
}

func TestMermaidRendererVerificationBadge(t *testing.T) {
	p := verifiedProtocol()
	r := NewMermaid()

	s, err := r.RenderString(p)
	if err != nil {
		t.Fatalf("RenderString() error = %v", err)
	}

	if !strings.Contains(s, "🔎 corroborated — METR") {
		t.Errorf("Mermaid output missing verification badge; got:\n%s", s)
	}
}

func TestVerificationBadgeSourceOmitted(t *testing.T) {
	got := verificationBadge(&pidl.FlowVerification{Tier: pidl.VerificationTierReported})
	if got != "🔎 reported" {
		t.Errorf("verificationBadge() = %q, want %q", got, "🔎 reported")
	}
}
