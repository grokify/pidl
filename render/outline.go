package render

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/grokify/pidl"
)

// OutlineRenderer writes a protocol's Outline as Markdown: roles, then
// scenarios with numbered steps, then tokens, all in document order. It is
// a starting point for guides and slide decks.
type OutlineRenderer struct{}

// NewOutline creates an OutlineRenderer.
func NewOutline() *OutlineRenderer {
	return &OutlineRenderer{}
}

// Format returns the output format identifier.
func (r *OutlineRenderer) Format() Format {
	return FormatOutline
}

// Render writes the outline to w.
func (r *OutlineRenderer) Render(w io.Writer, p *pidl.Protocol) error {
	s, err := r.RenderString(p)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, s)
	return err
}

// RenderString returns the outline as Markdown.
func (r *OutlineRenderer) RenderString(p *pidl.Protocol) (string, error) {
	o := p.Outline()
	name := func(id string) string {
		if e := p.EntityByID(id); e != nil && e.Name != "" {
			return e.Name
		}
		return id
	}

	var sb strings.Builder
	fmt.Fprintf(&sb, "# %s\n\n", o.Name)
	if o.Description != "" {
		fmt.Fprintf(&sb, "%s\n\n", o.Description)
	}

	sb.WriteString("## Roles\n\n")
	for _, role := range o.Roles {
		fmt.Fprintf(&sb, "%d. **%s**", role.Position, role.Name)
		if role.Description != "" {
			fmt.Fprintf(&sb, ": %s", role.Description)
		}
		sb.WriteString("\n")
	}

	sb.WriteString("\n## Scenarios\n")
	for _, sc := range o.Scenarios {
		title := sc.Name
		if sc.Fragment {
			title += " (fragment)"
		}
		fmt.Fprintf(&sb, "\n### %d. %s\n\n", sc.Position, title)
		if sc.Description != "" {
			fmt.Fprintf(&sb, "%s\n\n", sc.Description)
		}
		if g := groupName(o.Groups, sc.Group); g != "" {
			fmt.Fprintf(&sb, "Group: %s\n\n", g)
		}
		if sc.VariantOf != "" {
			fmt.Fprintf(&sb, "Variant of %s: %s\n\n", scenarioName(o.Scenarios, sc.VariantOf), sc.VariantLabel)
		}
		names := make([]string, len(sc.Participants))
		for i, id := range sc.Participants {
			names[i] = name(id)
		}
		fmt.Fprintf(&sb, "Participants: %s\n\n", strings.Join(names, ", "))
		for _, st := range sc.Steps {
			if st.Include != "" {
				ref := ""
				if st.IncludePosition > 0 {
					ref = fmt.Sprintf(" (scenario %d)", st.IncludePosition)
				}
				fmt.Fprintf(&sb, "%d. See *%s*%s\n", st.Number, st.Label, ref)
				continue
			}
			arrow := "→"
			if st.Response {
				arrow = "⇢"
			}
			fmt.Fprintf(&sb, "%d. %s %s %s: %s\n", st.Number, name(st.From), arrow, name(st.To), st.Label)
		}
	}

	if len(o.Tokens) > 0 {
		sb.WriteString("\n## Tokens\n\n")
		for _, t := range o.Tokens {
			fmt.Fprintf(&sb, "- **%s**", t.Name)
			var parts []string
			if t.Issuer != "" {
				parts = append(parts, "issued by "+name(t.Issuer))
			}
			if t.Audience != "" {
				parts = append(parts, "for "+name(t.Audience))
			}
			if len(t.VerifiedBy) > 0 {
				checkers := make([]string, len(t.VerifiedBy))
				for i, id := range t.VerifiedBy {
					checkers[i] = name(id)
				}
				parts = append(parts, "checked by "+strings.Join(checkers, ", "))
			}
			if len(parts) > 0 {
				fmt.Fprintf(&sb, ": %s", strings.Join(parts, "; "))
			}
			sb.WriteString("\n")
		}
	}
	return sb.String(), nil
}

// OutlineJSONRenderer writes a protocol's Outline as indented JSON, for
// deck and guide tooling.
type OutlineJSONRenderer struct{}

// NewOutlineJSON creates an OutlineJSONRenderer.
func NewOutlineJSON() *OutlineJSONRenderer {
	return &OutlineJSONRenderer{}
}

// Format returns the output format identifier.
func (r *OutlineJSONRenderer) Format() Format {
	return FormatOutlineJSON
}

// Render writes the outline JSON to w.
func (r *OutlineJSONRenderer) Render(w io.Writer, p *pidl.Protocol) error {
	s, err := r.RenderString(p)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, s)
	return err
}

// RenderString returns the outline as indented JSON.
func (r *OutlineJSONRenderer) RenderString(p *pidl.Protocol) (string, error) {
	data, err := json.MarshalIndent(p.Outline(), "", "  ")
	if err != nil {
		return "", err
	}
	return string(data) + "\n", nil
}

// groupName returns the display name of group id, or "" when id is empty
// or unknown.
func groupName(groups []pidl.Group, id string) string {
	for _, g := range groups {
		if g.ID == id {
			return g.Name
		}
	}
	return ""
}

// scenarioName returns the name of scenario id, or id when it is unknown.
func scenarioName(scenarios []pidl.OutlineScenario, id string) string {
	for _, sc := range scenarios {
		if sc.ID == id {
			return sc.Name
		}
	}
	return id
}
