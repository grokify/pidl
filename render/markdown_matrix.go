package render

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/grokify/pidl"
)

// MarkdownMatrixRenderer generates a Markdown table showing protocol roles by entity.
type MarkdownMatrixRenderer struct {
	// Title includes a title header
	Title bool
	// ShowDescription includes entity descriptions
	ShowDescription bool
}

// NewMarkdownMatrix creates a new MarkdownMatrixRenderer with default settings.
func NewMarkdownMatrix() *MarkdownMatrixRenderer {
	return &MarkdownMatrixRenderer{
		Title:           true,
		ShowDescription: false,
	}
}

// Format returns the output format identifier.
func (r *MarkdownMatrixRenderer) Format() Format {
	return FormatMarkdownMatrix
}

// Render writes the Markdown matrix to the given writer.
func (r *MarkdownMatrixRenderer) Render(w io.Writer, p *pidl.Protocol) error {
	s, err := r.RenderString(p)
	if err != nil {
		return err
	}
	_, err = w.Write([]byte(s))
	return err
}

// RenderString returns the Markdown matrix as a string.
func (r *MarkdownMatrixRenderer) RenderString(p *pidl.Protocol) (string, error) {
	var sb strings.Builder

	// Title
	if r.Title {
		sb.WriteString(fmt.Sprintf("# %s - Protocol Role Matrix\n\n", p.ProtocolMeta.Name))
	}

	// Collect all protocols used
	protocols := p.AllProtocols()
	if len(protocols) == 0 {
		sb.WriteString("*No protocol roles defined.*\n")
		return sb.String(), nil
	}
	sort.Strings(protocols)

	// Build header row
	sb.WriteString("| Entity |")
	for _, proto := range protocols {
		sb.WriteString(fmt.Sprintf(" %s |", proto))
	}
	sb.WriteString("\n")

	// Build separator row
	sb.WriteString("|--------|")
	for range protocols {
		sb.WriteString("--------|")
	}
	sb.WriteString("\n")

	// Build entity rows
	for _, e := range p.Entities {
		if len(e.ProtocolRoles) == 0 {
			continue
		}

		// Build a map of protocol -> roles for this entity
		roleMap := make(map[string][]string)
		for _, pr := range e.ProtocolRoles {
			role := pr.Role
			if pr.Variant != "" {
				role = fmt.Sprintf("%s (%s)", pr.Role, pr.Variant)
			}
			roleMap[pr.Protocol] = append(roleMap[pr.Protocol], role)
		}

		sb.WriteString(fmt.Sprintf("| **%s** |", e.Name))
		for _, proto := range protocols {
			roles := roleMap[proto]
			if len(roles) > 0 {
				sb.WriteString(fmt.Sprintf(" %s |", strings.Join(roles, ", ")))
			} else {
				sb.WriteString(" - |")
			}
		}
		sb.WriteString("\n")
	}

	// Add legend section
	sb.WriteString("\n## Protocol Legend\n\n")
	sb.WriteString("| Protocol | Description |\n")
	sb.WriteString("|----------|-------------|\n")

	protocolDescriptions := map[string]string{
		"oauth":   "OAuth 2.0 Authorization Framework",
		"scim":    "System for Cross-domain Identity Management",
		"spiffe":  "Secure Production Identity Framework for Everyone",
		"aauth":   "Agent Authorization Protocol",
		"idjag":   "Identity Assertion Authorization Grant",
		"authzen": "OpenID AuthZEN Authorization API",
		"mcp":     "Model Context Protocol",
		"a2a":     "Agent-to-Agent Protocol",
	}

	for _, proto := range protocols {
		desc := protocolDescriptions[proto]
		if desc == "" {
			desc = proto
		}
		sb.WriteString(fmt.Sprintf("| `%s` | %s |\n", proto, desc))
	}

	// Add component summary if present
	if len(p.Metadata.Components) > 0 {
		sb.WriteString("\n## Deployment Components\n\n")
		sb.WriteString("| Component | Type | Entities | Implements |\n")
		sb.WriteString("|-----------|------|----------|------------|\n")

		for _, c := range p.Metadata.Components {
			entities := strings.Join(c.Entities, ", ")
			var implements []string
			for _, impl := range c.Implements {
				implements = append(implements, fmt.Sprintf("%s:%s", impl.Protocol, impl.Role))
			}
			implStr := strings.Join(implements, ", ")
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", c.Name, c.Type, entities, implStr))
		}
	}

	// Add trust relationships summary if present
	if len(p.Metadata.TrustRelations) > 0 {
		sb.WriteString("\n## Trust Relationships\n\n")
		sb.WriteString("| From | Relationship | To | Credentials |\n")
		sb.WriteString("|------|--------------|----|--------------|\n")

		for _, t := range p.Metadata.TrustRelations {
			creds := strings.Join(t.Credentials, ", ")
			if creds == "" {
				creds = "-"
			}
			mutual := ""
			if t.Mutual {
				mutual = " ↔"
			}
			sb.WriteString(fmt.Sprintf("| %s | %s%s | %s | %s |\n", t.From, t.Type, mutual, t.To, creds))
		}
	}

	return sb.String(), nil
}
