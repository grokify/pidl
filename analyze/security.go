// Package analyze provides security analysis for PIDL protocols.
package analyze

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/grokify/pidl"
)

// RiskSeverity represents the severity level of a security risk.
type RiskSeverity string

const (
	// SeverityCritical represents critical security risks.
	SeverityCritical RiskSeverity = "critical"
	// SeverityHigh represents high security risks.
	SeverityHigh RiskSeverity = "high"
	// SeverityMedium represents medium security risks.
	SeverityMedium RiskSeverity = "medium"
	// SeverityLow represents low security risks.
	SeverityLow RiskSeverity = "low"
	// SeverityInfo represents informational findings.
	SeverityInfo RiskSeverity = "info"
)

// severityWeight returns the numeric weight of a severity level.
func (s RiskSeverity) Weight() int {
	switch s {
	case SeverityCritical:
		return 5
	case SeverityHigh:
		return 4
	case SeverityMedium:
		return 3
	case SeverityLow:
		return 2
	case SeverityInfo:
		return 1
	default:
		return 0
	}
}

// RiskCategory represents the category of a security risk.
type RiskCategory string

const (
	CategoryTrustBoundary   RiskCategory = "trust_boundary"
	CategoryAuthentication  RiskCategory = "authentication"
	CategoryDataProtection  RiskCategory = "data_protection"
	CategoryTokenSecurity   RiskCategory = "token_security"
	CategoryCommunication   RiskCategory = "communication"
	CategoryConfiguration   RiskCategory = "configuration"
	CategoryProcessSecurity RiskCategory = "process_security"
)

// SecurityRisk represents a security risk identified in the protocol.
type SecurityRisk struct {
	// ID is a unique identifier for this risk type.
	ID string `json:"id"`

	// Severity indicates how serious the risk is.
	Severity RiskSeverity `json:"severity"`

	// Category classifies the type of risk.
	Category RiskCategory `json:"category"`

	// Title is a brief description of the risk.
	Title string `json:"title"`

	// Description provides detailed information about the risk.
	Description string `json:"description"`

	// Location identifies where in the protocol the risk was found.
	Location string `json:"location"`

	// Remediation provides guidance on how to address the risk.
	Remediation string `json:"remediation"`
}

// SecuritySummary provides aggregate statistics for the analysis.
type SecuritySummary struct {
	// TotalRisks is the total number of risks found.
	TotalRisks int `json:"total_risks"`

	// BySeverity breaks down risks by severity level.
	BySeverity map[RiskSeverity]int `json:"by_severity"`

	// ByCategory breaks down risks by category.
	ByCategory map[RiskCategory]int `json:"by_category"`

	// Score is a numeric security score (0-100, higher is better).
	Score int `json:"score"`
}

// SecurityAnalysis contains the complete analysis result.
type SecurityAnalysis struct {
	// ProtocolID identifies the analyzed protocol.
	ProtocolID string `json:"protocol_id"`

	// ProtocolName is the human-readable name.
	ProtocolName string `json:"protocol_name"`

	// AnalyzedAt is when the analysis was performed.
	AnalyzedAt time.Time `json:"analyzed_at"`

	// Risks contains all identified security risks.
	Risks []SecurityRisk `json:"risks"`

	// Summary provides aggregate statistics.
	Summary SecuritySummary `json:"summary"`
}

// AnalysisOptions controls analysis behavior.
type AnalysisOptions struct {
	// MinSeverity filters risks below this severity level.
	MinSeverity RiskSeverity

	// Categories filters to only these categories (empty = all).
	Categories []RiskCategory

	// EnabledRules specifies which rules to run (empty = all).
	EnabledRules []string

	// DisabledRules specifies which rules to skip.
	DisabledRules []string
}

// DefaultAnalysisOptions returns default analysis options.
func DefaultAnalysisOptions() AnalysisOptions {
	return AnalysisOptions{
		MinSeverity: SeverityInfo,
	}
}

// Analyze performs security analysis on a protocol.
func Analyze(p *pidl.Protocol, opts AnalysisOptions) *SecurityAnalysis {
	analysis := &SecurityAnalysis{
		ProtocolID:   p.ProtocolMeta.ID,
		ProtocolName: p.ProtocolMeta.Name,
		AnalyzedAt:   time.Now(),
		Risks:        make([]SecurityRisk, 0),
		Summary: SecuritySummary{
			BySeverity: make(map[RiskSeverity]int),
			ByCategory: make(map[RiskCategory]int),
		},
	}

	// Run all rules
	rules := DefaultRules()
	for _, rule := range rules {
		// Skip disabled rules
		if isRuleDisabled(rule.ID, opts.EnabledRules, opts.DisabledRules) {
			continue
		}

		// Run rule and collect risks
		risks := rule.Check(p)
		for _, risk := range risks {
			// Filter by severity
			if risk.Severity.Weight() < opts.MinSeverity.Weight() {
				continue
			}

			// Filter by category
			if len(opts.Categories) > 0 && !containsCategory(opts.Categories, risk.Category) {
				continue
			}

			analysis.Risks = append(analysis.Risks, risk)
		}
	}

	// Calculate summary
	analysis.calculateSummary()

	return analysis
}

// calculateSummary computes aggregate statistics.
func (a *SecurityAnalysis) calculateSummary() {
	a.Summary.TotalRisks = len(a.Risks)
	a.Summary.BySeverity = make(map[RiskSeverity]int)
	a.Summary.ByCategory = make(map[RiskCategory]int)

	totalWeight := 0
	for _, risk := range a.Risks {
		a.Summary.BySeverity[risk.Severity]++
		a.Summary.ByCategory[risk.Category]++
		totalWeight += risk.Severity.Weight()
	}

	// Calculate score (0-100, higher is better)
	// Base score of 100, deduct based on risk severity
	if a.Summary.TotalRisks == 0 {
		a.Summary.Score = 100
	} else {
		// Deduct 20 points per critical, 10 per high, 5 per medium, 2 per low, 1 per info
		deduction := a.Summary.BySeverity[SeverityCritical]*20 +
			a.Summary.BySeverity[SeverityHigh]*10 +
			a.Summary.BySeverity[SeverityMedium]*5 +
			a.Summary.BySeverity[SeverityLow]*2 +
			a.Summary.BySeverity[SeverityInfo]*1

		a.Summary.Score = 100 - deduction
		if a.Summary.Score < 0 {
			a.Summary.Score = 0
		}
	}
}

// HasRisks returns true if any risks were found.
func (a *SecurityAnalysis) HasRisks() bool {
	return len(a.Risks) > 0
}

// HasRisksAtOrAbove returns true if any risks at or above the given severity exist.
func (a *SecurityAnalysis) HasRisksAtOrAbove(severity RiskSeverity) bool {
	for _, risk := range a.Risks {
		if risk.Severity.Weight() >= severity.Weight() {
			return true
		}
	}
	return false
}

// RisksByCategory returns risks filtered by category.
func (a *SecurityAnalysis) RisksByCategory(category RiskCategory) []SecurityRisk {
	var risks []SecurityRisk
	for _, risk := range a.Risks {
		if risk.Category == category {
			risks = append(risks, risk)
		}
	}
	return risks
}

// RisksBySeverity returns risks filtered by severity.
func (a *SecurityAnalysis) RisksBySeverity(severity RiskSeverity) []SecurityRisk {
	var risks []SecurityRisk
	for _, risk := range a.Risks {
		if risk.Severity == severity {
			risks = append(risks, risk)
		}
	}
	return risks
}

// String returns a human-readable text representation.
func (a *SecurityAnalysis) String() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("Security Analysis: %s\n", a.ProtocolName))
	sb.WriteString(strings.Repeat("=", 60) + "\n\n")

	if !a.HasRisks() {
		sb.WriteString("No security risks identified.\n")
		sb.WriteString(fmt.Sprintf("Security Score: %d/100\n", a.Summary.Score))
		return sb.String()
	}

	// Group by severity
	severities := []RiskSeverity{SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow, SeverityInfo}

	for _, severity := range severities {
		risks := a.RisksBySeverity(severity)
		if len(risks) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("%s Risks (%d)\n", strings.ToUpper(string(severity)), len(risks)))
		sb.WriteString(strings.Repeat("-", 40) + "\n")

		for _, risk := range risks {
			sb.WriteString(fmt.Sprintf("  [%s] %s\n", risk.ID, risk.Title))
			sb.WriteString(fmt.Sprintf("    Location: %s\n", risk.Location))
			sb.WriteString(fmt.Sprintf("    %s\n", risk.Description))
			if risk.Remediation != "" {
				sb.WriteString(fmt.Sprintf("    Fix: %s\n", risk.Remediation))
			}
			sb.WriteString("\n")
		}
	}

	sb.WriteString(strings.Repeat("-", 60) + "\n")
	sb.WriteString("Summary:\n")
	sb.WriteString(fmt.Sprintf("  Total Risks: %d\n", a.Summary.TotalRisks))
	sb.WriteString(fmt.Sprintf("  Critical: %d, High: %d, Medium: %d, Low: %d, Info: %d\n",
		a.Summary.BySeverity[SeverityCritical],
		a.Summary.BySeverity[SeverityHigh],
		a.Summary.BySeverity[SeverityMedium],
		a.Summary.BySeverity[SeverityLow],
		a.Summary.BySeverity[SeverityInfo]))
	sb.WriteString(fmt.Sprintf("  Security Score: %d/100\n", a.Summary.Score))

	return sb.String()
}

// ToMarkdown returns a markdown representation.
func (a *SecurityAnalysis) ToMarkdown() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("# Security Analysis: %s\n\n", a.ProtocolName))
	sb.WriteString(fmt.Sprintf("**Analyzed:** %s\n\n", a.AnalyzedAt.Format(time.RFC3339)))

	if !a.HasRisks() {
		sb.WriteString("No security risks identified.\n\n")
		sb.WriteString(fmt.Sprintf("**Security Score:** %d/100\n", a.Summary.Score))
		return sb.String()
	}

	// Summary table
	sb.WriteString("## Summary\n\n")
	sb.WriteString("| Metric | Value |\n|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Total Risks | %d |\n", a.Summary.TotalRisks))
	sb.WriteString(fmt.Sprintf("| Critical | %d |\n", a.Summary.BySeverity[SeverityCritical]))
	sb.WriteString(fmt.Sprintf("| High | %d |\n", a.Summary.BySeverity[SeverityHigh]))
	sb.WriteString(fmt.Sprintf("| Medium | %d |\n", a.Summary.BySeverity[SeverityMedium]))
	sb.WriteString(fmt.Sprintf("| Low | %d |\n", a.Summary.BySeverity[SeverityLow]))
	sb.WriteString(fmt.Sprintf("| Info | %d |\n", a.Summary.BySeverity[SeverityInfo]))
	sb.WriteString(fmt.Sprintf("| Security Score | %d/100 |\n\n", a.Summary.Score))

	// Risks by severity
	severities := []RiskSeverity{SeverityCritical, SeverityHigh, SeverityMedium, SeverityLow, SeverityInfo}

	for _, severity := range severities {
		risks := a.RisksBySeverity(severity)
		if len(risks) == 0 {
			continue
		}

		sb.WriteString(fmt.Sprintf("## %s Risks\n\n", strings.Title(string(severity))))

		for _, risk := range risks {
			sb.WriteString(fmt.Sprintf("### [%s] %s\n\n", risk.ID, risk.Title))
			sb.WriteString(fmt.Sprintf("- **Category:** %s\n", risk.Category))
			sb.WriteString(fmt.Sprintf("- **Location:** `%s`\n", risk.Location))
			sb.WriteString(fmt.Sprintf("- **Description:** %s\n", risk.Description))
			if risk.Remediation != "" {
				sb.WriteString(fmt.Sprintf("- **Remediation:** %s\n", risk.Remediation))
			}
			sb.WriteString("\n")
		}
	}

	return sb.String()
}

// ToJSON returns the analysis as JSON bytes.
func (a *SecurityAnalysis) ToJSON() ([]byte, error) {
	return json.MarshalIndent(a, "", "  ")
}

// Helper functions

func isRuleDisabled(ruleID string, enabled, disabled []string) bool {
	// If enabled list is specified, rule must be in it
	if len(enabled) > 0 {
		found := false
		for _, id := range enabled {
			if id == ruleID {
				found = true
				break
			}
		}
		if !found {
			return true
		}
	}

	// Check if explicitly disabled
	for _, id := range disabled {
		if id == ruleID {
			return true
		}
	}

	return false
}

func containsCategory(categories []RiskCategory, category RiskCategory) bool {
	for _, c := range categories {
		if c == category {
			return true
		}
	}
	return false
}

// ParseSeverity parses a severity string.
func ParseSeverity(s string) (RiskSeverity, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "critical":
		return SeverityCritical, nil
	case "high":
		return SeverityHigh, nil
	case "medium":
		return SeverityMedium, nil
	case "low":
		return SeverityLow, nil
	case "info", "informational":
		return SeverityInfo, nil
	default:
		return "", fmt.Errorf("unknown severity: %s", s)
	}
}

// ParseCategory parses a category string.
func ParseCategory(s string) (RiskCategory, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "trust_boundary", "trust-boundary", "trustboundary":
		return CategoryTrustBoundary, nil
	case "authentication", "auth":
		return CategoryAuthentication, nil
	case "data_protection", "data-protection", "dataprotection":
		return CategoryDataProtection, nil
	case "token_security", "token-security", "tokensecurity", "token":
		return CategoryTokenSecurity, nil
	case "communication", "comm":
		return CategoryCommunication, nil
	case "configuration", "config":
		return CategoryConfiguration, nil
	case "process_security", "process-security", "processsecurity", "process":
		return CategoryProcessSecurity, nil
	default:
		return "", fmt.Errorf("unknown category: %s", s)
	}
}
