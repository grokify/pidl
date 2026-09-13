package pidl

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestFlowVerificationJSONRoundTrip(t *testing.T) {
	orig := Flow{
		From:   "attacker",
		To:     "target",
		Action: "exploit",
		Verification: &FlowVerification{
			Tier:     VerificationTierCorroborated,
			Source:   "METR",
			Citation: "https://example.com/report",
		},
	}

	data, err := json.Marshal(orig)
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}

	// Provenance is serialized under the "verification" key.
	if !strings.Contains(string(data), `"verification"`) {
		t.Errorf("marshaled flow missing verification key; got %s", data)
	}

	var got Flow
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}

	if got.Verification == nil {
		t.Fatal("round-tripped flow has nil Verification")
	}
	if got.Verification.Tier != VerificationTierCorroborated {
		t.Errorf("Tier = %q, want %q", got.Verification.Tier, VerificationTierCorroborated)
	}
	if got.Verification.Source != "METR" {
		t.Errorf("Source = %q, want %q", got.Verification.Source, "METR")
	}
	if got.Verification.Citation != "https://example.com/report" {
		t.Errorf("Citation = %q, want %q", got.Verification.Citation, "https://example.com/report")
	}
}

func TestFlowVerificationOmittedWhenNil(t *testing.T) {
	data, err := json.Marshal(Flow{From: "a", To: "b", Action: "x"})
	if err != nil {
		t.Fatalf("Marshal() error = %v", err)
	}
	if strings.Contains(string(data), "verification") {
		t.Errorf("verification should be omitted when nil; got %s", data)
	}
}

func TestFlowHasVerification(t *testing.T) {
	if (Flow{}).HasVerification() {
		t.Error("empty flow should not report verification")
	}
	f := Flow{Verification: &FlowVerification{Tier: VerificationTierReported}}
	if !f.HasVerification() {
		t.Error("flow with tier should report verification")
	}
	if (Flow{Verification: &FlowVerification{}}).HasVerification() {
		t.Error("flow with empty tier should not report verification")
	}
}
