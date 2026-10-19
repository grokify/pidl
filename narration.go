package pidl

import (
	"regexp"
	"sort"
)

// DefaultLanguage is the source language of narration when the protocol
// does not set one.
const DefaultLanguage = "en-US"

// Representation names for narration overrides. Sample languages, such as
// "go" or "python", are representations too.
const (
	RepresentationHTTP = "http"
	RepresentationCurl = "curl"
)

// languageTagPattern accepts BCP 47 tags such as "en-US", "fr" and
// "zh-Hans". It checks the shape, not the registry.
var languageTagPattern = regexp.MustCompile(`^[A-Za-z]{2,3}(-[A-Za-z0-9]{2,8})*$`)

// Narration is text written to be spoken, keyed by BCP 47 language tag.
// It is always an object, so a document can carry several languages and
// the schema needs no union type.
type Narration map[string]string

// Text returns the narration in lang, falling back to source, the
// document's source language. ok is false when neither is present.
func (n Narration) Text(lang, source string) (text string, ok bool) {
	if t, found := n[lang]; found && t != "" {
		return t, true
	}
	if t, found := n[source]; found && t != "" {
		return t, true
	}
	return "", false
}

// Languages returns the narration's languages, sorted.
func (n Narration) Languages() []string {
	langs := make([]string, 0, len(n))
	for l := range n {
		langs = append(langs, l)
	}
	sort.Strings(langs)
	return langs
}

// Briefing introduces a scenario before its steps, for example on slides
// that open a video.
type Briefing struct {
	// Why explains, as prose, why the scenario exists and what problem it
	// solves.
	Why string `json:"why,omitempty"`

	// Points are the key points, in display order.
	Points []string `json:"points,omitempty"`

	// Narration is spoken over the briefing.
	Narration Narration `json:"narration,omitempty"`
}

// RepresentationNarration overrides a step's narration when the reader
// views the step in one representation, such as cURL or Go code.
type RepresentationNarration struct {
	// Narration replaces the step's narration for this representation.
	Narration Narration `json:"narration,omitempty"`
}

// SourceLanguage returns the protocol's narration source language.
func (p *Protocol) SourceLanguage() string {
	if p.ProtocolMeta.Language != "" {
		return p.ProtocolMeta.Language
	}
	return DefaultLanguage
}

// StepNarration returns a flow's narration in lang for representation rep,
// falling back from the representation's narration to the flow's own, and
// from lang to the source language. ok is false when there is none; tools
// then use the flow's description.
func (p *Protocol) StepNarration(f Flow, rep, lang string) (text string, ok bool) {
	source := p.SourceLanguage()
	if r, found := f.Representations[rep]; found {
		if t, ok := r.Narration.Text(lang, source); ok {
			return t, true
		}
	}
	return f.Narration.Text(lang, source)
}
