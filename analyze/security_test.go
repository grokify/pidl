package analyze

import (
	"strings"
	"testing"

	"github.com/grokify/pidl"
)

func createTestProtocolWithSecurity() *pidl.Protocol {
	return &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "test-protocol",
			Name: "Test Protocol",
		},
		Entities: []pidl.Entity{
			{
				ID:         "client",
				Name:       "Client",
				Type:       pidl.EntityTypeClient,
				TrustLevel: pidl.TrustLevelUntrusted,
			},
			{
				ID:         "server",
				Name:       "Server",
				Type:       pidl.EntityTypeServer,
				TrustLevel: pidl.TrustLevelTrusted,
			},
		},
		Flows: []pidl.Flow{
			{
				From:   "client",
				To:     "server",
				Action: "request",
				Mode:   pidl.FlowModeRequest,
				Security: &pidl.FlowSecurity{
					Requires: []pidl.SecurityRequirement{pidl.SecurityRequirementToken},
				},
			},
			{
				From:   "server",
				To:     "client",
				Action: "response",
				Mode:   pidl.FlowModeResponse,
			},
		},
	}
}

func createProtocolWithVulnerabilities() *pidl.Protocol {
	return &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "vulnerable-protocol",
			Name: "Vulnerable Protocol",
		},
		Entities: []pidl.Entity{
			{
				ID:         "client",
				Name:       "Client",
				Type:       pidl.EntityTypeClient,
				TrustLevel: pidl.TrustLevelUntrusted,
			},
			{
				ID:         "server",
				Name:       "Server",
				Type:       pidl.EntityTypeServer,
				TrustLevel: pidl.TrustLevelTrusted,
			},
		},
		Flows: []pidl.Flow{
			// Untrusted to trusted without security
			{
				From:   "client",
				To:     "server",
				Action: "insecure_request",
				Mode:   pidl.FlowModeRequest,
			},
		},
	}
}

func createProtocolWithConfidentialData() *pidl.Protocol {
	return &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "confidential-protocol",
			Name: "Confidential Protocol",
		},
		Entities: []pidl.Entity{
			{ID: "client", Name: "Client", Type: pidl.EntityTypeClient},
			{ID: "server", Name: "Server", Type: pidl.EntityTypeServer},
		},
		Flows: []pidl.Flow{
			{
				From:   "client",
				To:     "server",
				Action: "send_secret",
				Security: &pidl.FlowSecurity{
					Confidential: true,
					// Missing encryption requirement
				},
			},
		},
	}
}

func createProtocolWithTokens() *pidl.Protocol {
	return &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{
			ID:   "token-protocol",
			Name: "Token Protocol",
		},
		Entities: []pidl.Entity{
			{ID: "client", Name: "Client", Type: pidl.EntityTypeClient},
			{ID: "server", Name: "Server", Type: pidl.EntityTypeServer},
		},
		Flows: []pidl.Flow{
			{From: "client", To: "server", Action: "request"},
		},
		Metadata: &pidl.ProtocolMetadata{
			Tokens: []pidl.TokenDefinition{
				{ID: "access_token", Type: "jwt", Binding: "bearer"},
				{ID: "bound_token", Type: "jwt", Binding: "mtls", Audience: "server"},
			},
		},
	}
}

func TestAnalyze_NoRisks(t *testing.T) {
	p := createTestProtocolWithSecurity()
	opts := DefaultAnalysisOptions()

	analysis := Analyze(p, opts)

	// Should have minimal risks since we have proper security
	if analysis.Summary.TotalRisks > 5 {
		t.Errorf("expected few risks for secure protocol, got %d", analysis.Summary.TotalRisks)
	}
}

func TestAnalyze_TrustBoundaryViolation(t *testing.T) {
	p := createProtocolWithVulnerabilities()
	opts := DefaultAnalysisOptions()

	analysis := Analyze(p, opts)

	if !analysis.HasRisks() {
		t.Error("expected risks for vulnerable protocol")
	}

	// Should find trust boundary violation
	found := false
	for _, risk := range analysis.Risks {
		if risk.ID == "SEC001" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected SEC001 trust boundary violation")
	}
}

func TestAnalyze_MissingEncryption(t *testing.T) {
	p := createProtocolWithConfidentialData()
	opts := DefaultAnalysisOptions()

	analysis := Analyze(p, opts)

	if !analysis.HasRisks() {
		t.Error("expected risks for protocol with missing encryption")
	}

	// Should find missing encryption
	found := false
	for _, risk := range analysis.Risks {
		if risk.ID == "SEC002" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected SEC002 missing encryption risk")
	}
}

func TestAnalyze_TokenWithoutBinding(t *testing.T) {
	p := createProtocolWithTokens()
	opts := DefaultAnalysisOptions()

	analysis := Analyze(p, opts)

	// Should find unbound bearer token
	found := false
	for _, risk := range analysis.Risks {
		if risk.ID == "SEC003" {
			found = true
			break
		}
	}

	if !found {
		t.Error("expected SEC003 unbound bearer token risk")
	}
}

func TestAnalyze_FilterBySeverity(t *testing.T) {
	p := createProtocolWithVulnerabilities()
	opts := AnalysisOptions{
		MinSeverity: SeverityHigh,
	}

	analysis := Analyze(p, opts)

	// Should only have high severity or above
	for _, risk := range analysis.Risks {
		if risk.Severity.Weight() < SeverityHigh.Weight() {
			t.Errorf("found risk below high severity: %s (%s)", risk.ID, risk.Severity)
		}
	}
}

func TestAnalyze_FilterByCategory(t *testing.T) {
	p := createProtocolWithVulnerabilities()
	opts := AnalysisOptions{
		Categories: []RiskCategory{CategoryTrustBoundary},
	}

	analysis := Analyze(p, opts)

	// Should only have trust boundary risks
	for _, risk := range analysis.Risks {
		if risk.Category != CategoryTrustBoundary {
			t.Errorf("found risk outside trust boundary category: %s (%s)", risk.ID, risk.Category)
		}
	}
}

func TestSecurityAnalysis_HasRisksAtOrAbove(t *testing.T) {
	p := createProtocolWithVulnerabilities()
	opts := DefaultAnalysisOptions()

	analysis := Analyze(p, opts)

	if !analysis.HasRisksAtOrAbove(SeverityMedium) {
		t.Error("expected risks at medium or above")
	}

	// Create empty analysis
	emptyAnalysis := &SecurityAnalysis{Risks: []SecurityRisk{}}
	if emptyAnalysis.HasRisksAtOrAbove(SeverityInfo) {
		t.Error("expected no risks in empty analysis")
	}
}

func TestSecurityAnalysis_RisksByCategory(t *testing.T) {
	p := createProtocolWithVulnerabilities()
	opts := DefaultAnalysisOptions()

	analysis := Analyze(p, opts)

	trustRisks := analysis.RisksByCategory(CategoryTrustBoundary)
	if len(trustRisks) == 0 {
		t.Error("expected trust boundary risks")
	}

	for _, risk := range trustRisks {
		if risk.Category != CategoryTrustBoundary {
			t.Errorf("wrong category: expected %s, got %s", CategoryTrustBoundary, risk.Category)
		}
	}
}

func TestSecurityAnalysis_RisksBySeverity(t *testing.T) {
	p := createProtocolWithVulnerabilities()
	opts := DefaultAnalysisOptions()

	analysis := Analyze(p, opts)

	highRisks := analysis.RisksBySeverity(SeverityHigh)
	for _, risk := range highRisks {
		if risk.Severity != SeverityHigh {
			t.Errorf("wrong severity: expected %s, got %s", SeverityHigh, risk.Severity)
		}
	}
}

func TestSecurityAnalysis_String(t *testing.T) {
	p := createProtocolWithVulnerabilities()
	opts := DefaultAnalysisOptions()

	analysis := Analyze(p, opts)
	output := analysis.String()

	if output == "" {
		t.Error("expected non-empty string output")
	}

	if !strings.Contains(output, "Security Analysis") {
		t.Error("expected header in output")
	}

	if !strings.Contains(output, "Security Score") {
		t.Error("expected security score in output")
	}
}

func TestSecurityAnalysis_ToMarkdown(t *testing.T) {
	p := createProtocolWithVulnerabilities()
	opts := DefaultAnalysisOptions()

	analysis := Analyze(p, opts)
	output := analysis.ToMarkdown()

	if output == "" {
		t.Error("expected non-empty markdown output")
	}

	if !strings.Contains(output, "# Security Analysis") {
		t.Error("expected markdown header in output")
	}
}

func TestSecurityAnalysis_ToJSON(t *testing.T) {
	p := createProtocolWithVulnerabilities()
	opts := DefaultAnalysisOptions()

	analysis := Analyze(p, opts)
	jsonBytes, err := analysis.ToJSON()

	if err != nil {
		t.Fatalf("ToJSON failed: %v", err)
	}

	if len(jsonBytes) == 0 {
		t.Error("expected non-empty JSON output")
	}

	if !strings.Contains(string(jsonBytes), "protocol_id") {
		t.Error("expected protocol_id in JSON")
	}
}

func TestSecurityAnalysis_Score(t *testing.T) {
	// Secure protocol should have high score
	secureP := createTestProtocolWithSecurity()
	secureAnalysis := Analyze(secureP, DefaultAnalysisOptions())

	// Vulnerable protocol should have lower score
	vulnP := createProtocolWithVulnerabilities()
	vulnAnalysis := Analyze(vulnP, DefaultAnalysisOptions())

	if vulnAnalysis.Summary.Score >= secureAnalysis.Summary.Score {
		t.Error("vulnerable protocol should have lower score than secure protocol")
	}
}

func TestParseSeverity(t *testing.T) {
	tests := []struct {
		input    string
		expected RiskSeverity
		wantErr  bool
	}{
		{"critical", SeverityCritical, false},
		{"CRITICAL", SeverityCritical, false},
		{"high", SeverityHigh, false},
		{"medium", SeverityMedium, false},
		{"low", SeverityLow, false},
		{"info", SeverityInfo, false},
		{"informational", SeverityInfo, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseSeverity(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSeverity(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.expected {
				t.Errorf("ParseSeverity(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestParseCategory(t *testing.T) {
	tests := []struct {
		input    string
		expected RiskCategory
		wantErr  bool
	}{
		{"trust_boundary", CategoryTrustBoundary, false},
		{"authentication", CategoryAuthentication, false},
		{"auth", CategoryAuthentication, false},
		{"data_protection", CategoryDataProtection, false},
		{"token_security", CategoryTokenSecurity, false},
		{"token", CategoryTokenSecurity, false},
		{"communication", CategoryCommunication, false},
		{"configuration", CategoryConfiguration, false},
		{"invalid", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := ParseCategory(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCategory(%q) error = %v, wantErr %v", tt.input, err, tt.wantErr)
			}
			if got != tt.expected {
				t.Errorf("ParseCategory(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestRiskSeverity_Weight(t *testing.T) {
	if SeverityCritical.Weight() <= SeverityHigh.Weight() {
		t.Error("critical should have higher weight than high")
	}
	if SeverityHigh.Weight() <= SeverityMedium.Weight() {
		t.Error("high should have higher weight than medium")
	}
	if SeverityMedium.Weight() <= SeverityLow.Weight() {
		t.Error("medium should have higher weight than low")
	}
	if SeverityLow.Weight() <= SeverityInfo.Weight() {
		t.Error("low should have higher weight than info")
	}
}

func TestDefaultRules(t *testing.T) {
	rules := DefaultRules()

	if len(rules) == 0 {
		t.Error("expected default rules")
	}

	// Check rules have required fields
	for _, rule := range rules {
		if rule.ID == "" {
			t.Error("rule missing ID")
		}
		if rule.Name == "" {
			t.Error("rule missing name")
		}
		if rule.Check == nil {
			t.Error("rule missing Check function")
		}
	}
}

func TestDefaultAnalysisOptions(t *testing.T) {
	opts := DefaultAnalysisOptions()

	if opts.MinSeverity != SeverityInfo {
		t.Errorf("expected MinSeverity to be Info, got %s", opts.MinSeverity)
	}
}
