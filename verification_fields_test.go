package pidl

import (
	"strings"
	"testing"
)

func verificationProtocol() *Protocol {
	p := scenarioProtocol()
	p.Metadata = &ProtocolMetadata{Tokens: []TokenDefinition{
		{
			ID: "resource_token", Issuer: "resource", Audience: "ps",
			VerifiedBy: []string{"ps"},
			Claims: []TokenClaim{
				{Name: "aud", CheckedBy: []string{"ps"}},
				{Name: "agent_jkt", CheckedBy: []string{"ps"}, Description: "Must match the key that signed the request"},
			},
		},
		{ID: "auth_token", Issuer: "ps", Audience: "resource", VerifiedBy: []string{"resource"}},
	}}
	tp := &p.Scenarios[2]
	tp.Flows[1].Security = &FlowSecurity{Verifies: []string{"resource_token"}, SignedBy: "agent"}
	tp.Flows[3].Security = &FlowSecurity{Issues: []string{"auth_token"}}
	return p
}

func TestValidateVerificationFields(t *testing.T) {
	if errs := verificationProtocol().Validate(); len(errs) > 0 {
		t.Fatalf("Validate() = %v, want no errors", errs)
	}

	tests := []struct {
		name      string
		mutate    func(p *Protocol)
		wantField string
		wantMsg   string
	}{
		{"unknown verifier", func(p *Protocol) { p.Metadata.Tokens[0].VerifiedBy = []string{"as"} },
			"metadata.tokens[0].verified_by[0]", `unknown entity "as"`},
		{"duplicate verifier", func(p *Protocol) { p.Metadata.Tokens[0].VerifiedBy = []string{"ps", "ps"} },
			"metadata.tokens[0].verified_by[1]", "duplicate entity"},
		{"claim without name", func(p *Protocol) { p.Metadata.Tokens[0].Claims[0].Name = "" },
			"metadata.tokens[0].claims[0].name", "required"},
		{"duplicate claim", func(p *Protocol) { p.Metadata.Tokens[0].Claims[1].Name = "aud" },
			"metadata.tokens[0].claims[1].name", `duplicate claim "aud"`},
		{"unknown claim checker", func(p *Protocol) { p.Metadata.Tokens[0].Claims[0].CheckedBy = []string{"nobody"} },
			"metadata.tokens[0].claims[0].checked_by[0]", `unknown entity "nobody"`},
		{"unknown signer", func(p *Protocol) { p.Scenarios[2].Flows[1].Security.SignedBy = "nobody" },
			"scenarios[2].flows[1].security.signed_by", `unknown entity "nobody"`},
		{"unknown issued token", func(p *Protocol) { p.Scenarios[2].Flows[3].Security.Issues = []string{"nope"} },
			"scenarios[2].flows[3].security.issues[0]", `unknown token definition "nope"`},
		{"unknown verified token", func(p *Protocol) { p.Scenarios[2].Flows[1].Security.Verifies = []string{"nope"} },
			"scenarios[2].flows[1].security.verifies[0]", `unknown token definition "nope"`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := verificationProtocol()
			tt.mutate(p)
			for _, e := range p.Validate() {
				if e.Field == tt.wantField && strings.Contains(e.Message, tt.wantMsg) {
					return
				}
			}
			t.Fatalf("Validate() = %v, want %s: ...%s...", p.Validate(), tt.wantField, tt.wantMsg)
		})
	}
}

func TestVerificationFieldsRoundTrip(t *testing.T) {
	p := verificationProtocol()
	data, err := p.ToJSON()
	if err != nil {
		t.Fatalf("ToJSON: %v", err)
	}
	for _, want := range []string{`"verified_by"`, `"claims"`, `"checked_by"`, `"signed_by": "agent"`, `"issues"`, `"verifies"`} {
		if !strings.Contains(string(data), want) {
			t.Errorf("JSON missing %s", want)
		}
	}
	back, err := Parse(data)
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if errs := back.Validate(); len(errs) > 0 {
		t.Fatalf("round-tripped protocol does not validate: %v", errs)
	}
}
