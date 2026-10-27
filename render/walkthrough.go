package render

import (
	"encoding/json"
	"io"

	"github.com/grokify/pidl"
	"github.com/grokify/pidl/analyze"
)

// WalkthroughJSONRenderer writes a protocol's walkthrough (pidl.Walkthrough)
// as JSON for step-by-step viewers, with analyzer findings attached to the
// steps they concern.
type WalkthroughJSONRenderer struct {
	// Analysis configures the analyzer run whose findings are attached.
	// Set Analyze to false to attach none.
	Analyze  bool
	Analysis analyze.AnalysisOptions

	// Transform, when set, changes the walkthrough after it is built and
	// before it is written, such as adding a run observed in a real system.
	Transform func(*pidl.Walkthrough) error
}

// NewWalkthroughJSON creates a WalkthroughJSONRenderer that attaches
// findings of every severity except SEC016 (flow provenance), which is
// meant for incident attack-flows and would flag every step of an ordinary
// protocol.
func NewWalkthroughJSON() *WalkthroughJSONRenderer {
	opts := analyze.DefaultAnalysisOptions()
	opts.DisabledRules = []string{"SEC016"}
	return &WalkthroughJSONRenderer{Analyze: true, Analysis: opts}
}

// Format returns the output format identifier.
func (r *WalkthroughJSONRenderer) Format() Format {
	return FormatWalkthroughJSON
}

// Render writes the walkthrough JSON to w.
func (r *WalkthroughJSONRenderer) Render(w io.Writer, p *pidl.Protocol) error {
	s, err := r.RenderString(p)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, s)
	return err
}

// RenderString returns the walkthrough as indented JSON.
func (r *WalkthroughJSONRenderer) RenderString(p *pidl.Protocol) (string, error) {
	w, err := r.Build(p)
	if err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(w, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data) + "\n", nil
}

// Build returns the walkthrough with findings attached and Transform
// applied.
func (r *WalkthroughJSONRenderer) Build(p *pidl.Protocol) (pidl.Walkthrough, error) {
	w := r.Walkthrough(p)
	if r.Transform != nil {
		if err := r.Transform(&w); err != nil {
			return pidl.Walkthrough{}, err
		}
	}
	return w, nil
}

// Walkthrough returns the walkthrough with findings attached.
func (r *WalkthroughJSONRenderer) Walkthrough(p *pidl.Protocol) pidl.Walkthrough {
	var findings []pidl.Finding
	if r.Analyze {
		findings = analyze.Analyze(p, r.Analysis).Findings()
	}
	return p.Walkthrough(findings)
}
