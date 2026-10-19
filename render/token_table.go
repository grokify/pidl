package render

import (
	"fmt"
	"io"
	"strings"

	"github.com/grokify/pidl"
)

// TokenTableRenderer generates a Markdown table of the protocol's token
// definitions, in the order they are declared: issuer, audience, binding,
// who checks each token and which claims they rely on, and the scenarios
// that issue and check it. A second table lists claim descriptions.
type TokenTableRenderer struct {
	// Title adds a heading with the protocol name.
	Title bool
}

// NewTokenTable creates a TokenTableRenderer.
func NewTokenTable() *TokenTableRenderer {
	return &TokenTableRenderer{Title: true}
}

// Format returns the output format identifier.
func (r *TokenTableRenderer) Format() Format {
	return FormatTokenTable
}

// Render writes the table to w.
func (r *TokenTableRenderer) Render(w io.Writer, p *pidl.Protocol) error {
	s, err := r.RenderString(p)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, s)
	return err
}

// RenderString returns the token table as Markdown.
func (r *TokenTableRenderer) RenderString(p *pidl.Protocol) (string, error) {
	var sb strings.Builder
	if r.Title {
		fmt.Fprintf(&sb, "# %s - Tokens\n\n", p.ProtocolMeta.Name)
	}
	if p.Metadata == nil || len(p.Metadata.Tokens) == 0 {
		sb.WriteString("*No tokens defined.*\n")
		return sb.String(), nil
	}

	withScenarios := p.HasScenarios()
	sb.WriteString("| Token | Type | Issuer | Audience | Binding | Verified by | Claims checked |")
	if withScenarios {
		sb.WriteString(" Issued in | Verified in |")
	}
	sb.WriteString("\n|-------|------|--------|----------|---------|-------------|----------------|")
	if withScenarios {
		sb.WriteString("-----------|-------------|")
	}
	sb.WriteString("\n")

	hasClaimDetail := false
	for _, t := range p.Metadata.Tokens {
		name := t.Name
		if name == "" {
			name = t.ID
		}
		var claims []string
		for _, c := range t.Claims {
			claim := "`" + c.Name + "`"
			if len(c.CheckedBy) > 0 {
				claim += " (" + entityNames(p, c.CheckedBy) + ")"
			}
			claims = append(claims, claim)
			hasClaimDetail = hasClaimDetail || c.Description != ""
		}
		fmt.Fprintf(&sb, "| **%s** | %s | %s | %s | %s | %s | %s |",
			escapeCell(name), cellOrDash(t.Type), cellOrDash(entityName(p, t.Issuer)), cellOrDash(entityName(p, t.Audience)),
			cellOrDash(t.Binding), cellOrDash(entityNames(p, t.VerifiedBy)), cellOrDash(strings.Join(claims, "; ")))
		if withScenarios {
			issued, verified := p.TokenUsage(t.ID)
			fmt.Fprintf(&sb, " %s | %s |", cellOrDash(scenarioNames(p, issued)), cellOrDash(scenarioNames(p, verified)))
		}
		sb.WriteString("\n")
	}

	if hasClaimDetail {
		sb.WriteString("\n## Claims\n\n| Token | Claim | Checked by | Description |\n|-------|-------|------------|-------------|\n")
		for _, t := range p.Metadata.Tokens {
			name := t.Name
			if name == "" {
				name = t.ID
			}
			for _, c := range t.Claims {
				fmt.Fprintf(&sb, "| %s | `%s` | %s | %s |\n", escapeCell(name), c.Name,
					cellOrDash(entityNames(p, c.CheckedBy)), cellOrDash(escapeCell(c.Description)))
			}
		}
	}
	return sb.String(), nil
}

// entityName returns the display name for entity id, or id when unknown.
func entityName(p *pidl.Protocol, id string) string {
	if id == "" {
		return ""
	}
	if e := p.EntityByID(id); e != nil && e.Name != "" {
		return escapeCell(e.Name)
	}
	return escapeCell(id)
}

// entityNames returns display names for ids, in entities order.
func entityNames(p *pidl.Protocol, ids []string) string {
	ordered := p.OrderEntities(ids)
	names := make([]string, len(ordered))
	for i, id := range ordered {
		names[i] = entityName(p, id)
	}
	return strings.Join(names, ", ")
}

// scenarioNames returns display names for scenario ids, keeping their order.
func scenarioNames(p *pidl.Protocol, ids []string) string {
	names := make([]string, len(ids))
	for i, id := range ids {
		names[i] = id
		if sc := p.ScenarioByID(id); sc != nil && sc.Name != "" {
			names[i] = escapeCell(sc.Name)
		}
	}
	return strings.Join(names, ", ")
}

// cellOrDash returns s, or an em dash for an empty table cell.
func cellOrDash(s string) string {
	if s == "" {
		return "—"
	}
	return s
}
