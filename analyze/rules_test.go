package analyze

import (
	"testing"

	"github.com/grokify/pidl"
)

func TestVerificationTierWeight(t *testing.T) {
	tests := []struct {
		name string
		tier pidl.VerificationTier
		want int
	}{
		{name: "reproduced", tier: pidl.VerificationTierReproduced, want: 4},
		{name: "partially reproduced", tier: pidl.VerificationTierPartiallyReproduced, want: 3},
		{name: "corroborated", tier: pidl.VerificationTierCorroborated, want: 2},
		{name: "reported", tier: pidl.VerificationTierReported, want: 1},
		{name: "empty", tier: "", want: 0},
		{name: "unknown", tier: pidl.VerificationTier("inferred"), want: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := verificationTierWeight(tt.tier); got != tt.want {
				t.Errorf("verificationTierWeight(%q) = %d, want %d", tt.tier, got, tt.want)
			}
		})
	}
}

func TestVerificationTierWeight_Ordering(t *testing.T) {
	ordered := []pidl.VerificationTier{
		pidl.VerificationTierReported,
		pidl.VerificationTierCorroborated,
		pidl.VerificationTierPartiallyReproduced,
		pidl.VerificationTierReproduced,
	}
	for i := 1; i < len(ordered); i++ {
		lo, hi := ordered[i-1], ordered[i]
		if verificationTierWeight(lo) >= verificationTierWeight(hi) {
			t.Errorf("weight(%q) should be < weight(%q)", lo, hi)
		}
	}
}

// agentProtocol builds an agent-category protocol with one flow carrying the
// given verification (nil for none).
func agentProtocol(category pidl.Category, verification *pidl.FlowVerification) *pidl.Protocol {
	return &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{ID: "agent", Name: "Agent", Category: category},
		Entities: []pidl.Entity{
			{ID: "attacker", Name: "Attacker", Type: pidl.EntityTypeClient},
			{ID: "target", Name: "Target", Type: pidl.EntityTypeServer},
		},
		Flows: []pidl.Flow{
			{From: "attacker", To: "target", Action: "exploit", Verification: verification},
		},
	}
}

func TestRuleFlowWithoutVerification(t *testing.T) {
	tests := []struct {
		name         string
		protocol     *pidl.Protocol
		wantCount    int
		wantSeverity RiskSeverity
	}{
		{
			name:     "non-agent category is skipped",
			protocol: agentProtocol(pidl.CategoryAuth, nil),
		},
		{
			name:         "no verification is medium",
			protocol:     agentProtocol(pidl.CategoryAgent, nil),
			wantCount:    1,
			wantSeverity: SeverityMedium,
		},
		{
			name:         "empty tier counts as no verification",
			protocol:     agentProtocol(pidl.CategoryAgent, &pidl.FlowVerification{Source: "blog"}),
			wantCount:    1,
			wantSeverity: SeverityMedium,
		},
		{
			name:         "reported tier is low",
			protocol:     agentProtocol(pidl.CategoryAgent, &pidl.FlowVerification{Tier: pidl.VerificationTierReported}),
			wantCount:    1,
			wantSeverity: SeverityLow,
		},
		{
			name:         "unknown tier is low",
			protocol:     agentProtocol(pidl.CategoryAgent, &pidl.FlowVerification{Tier: "inferred"}),
			wantCount:    1,
			wantSeverity: SeverityLow,
		},
		{
			name:     "corroborated tier passes",
			protocol: agentProtocol(pidl.CategoryAgent, &pidl.FlowVerification{Tier: pidl.VerificationTierCorroborated}),
		},
		{
			name:     "reproduced tier passes",
			protocol: agentProtocol(pidl.CategoryAgent, &pidl.FlowVerification{Tier: pidl.VerificationTierReproduced}),
		},
	}

	rule := ruleFlowWithoutVerification()
	if rule.ID != "SEC016" {
		t.Fatalf("rule ID = %q, want SEC016", rule.ID)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			risks := rule.Check(tt.protocol)
			if len(risks) != tt.wantCount {
				t.Fatalf("got %d risks, want %d: %+v", len(risks), tt.wantCount, risks)
			}
			for _, r := range risks {
				if r.ID != "SEC016" {
					t.Errorf("risk ID = %q, want SEC016", r.ID)
				}
				if r.Severity != tt.wantSeverity {
					t.Errorf("severity = %q, want %q", r.Severity, tt.wantSeverity)
				}
				if r.Category != CategoryProcessSecurity {
					t.Errorf("category = %q, want %q", r.Category, CategoryProcessSecurity)
				}
				if r.Location != "flows[0]" {
					t.Errorf("location = %q, want flows[0]", r.Location)
				}
			}
		})
	}
}

func TestIsRuleDisabled(t *testing.T) {
	tests := []struct {
		name     string
		ruleID   string
		enabled  []string
		disabled []string
		want     bool
	}{
		{name: "no lists", ruleID: "SEC001", want: false},
		{name: "in enabled list", ruleID: "SEC001", enabled: []string{"SEC001", "SEC002"}, want: false},
		{name: "not in enabled list", ruleID: "SEC003", enabled: []string{"SEC001", "SEC002"}, want: true},
		{name: "in disabled list", ruleID: "SEC001", disabled: []string{"SEC001"}, want: true},
		{name: "not in disabled list", ruleID: "SEC002", disabled: []string{"SEC001"}, want: false},
		{name: "enabled but also disabled", ruleID: "SEC001", enabled: []string{"SEC001"}, disabled: []string{"SEC001"}, want: true},
		{name: "empty enabled slice means all", ruleID: "SEC001", enabled: []string{}, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isRuleDisabled(tt.ruleID, tt.enabled, tt.disabled); got != tt.want {
				t.Errorf("isRuleDisabled(%q, %v, %v) = %v, want %v", tt.ruleID, tt.enabled, tt.disabled, got, tt.want)
			}
		})
	}
}
