package pidl

import (
	"reflect"
	"strings"
	"testing"
)

// narratedProtocol returns a valid protocol with narration in two
// languages, a briefing, a cURL override and pronunciations.
func narratedProtocol() *Protocol {
	p := structuredProtocol()
	p.ProtocolMeta.Language = "en-US"
	p.Pronunciations = map[string]map[string]string{
		"agent_jkt": {"en-US": "agent J K T", "fr-FR": "agent J K T"},
	}
	p.Scenarios[0].Narration = Narration{"en-US": "The agent proves who it is.", "fr-FR": "L'agent prouve son identité."}
	p.Scenarios[0].Briefing = &Briefing{
		Why:       "Resources need to know which agent is calling.",
		Points:    []string{"Signed requests", "No tokens from a server"},
		Narration: Narration{"en-US": "Why agent identity matters."},
	}
	f := &p.Scenarios[0].Flows[0]
	f.Narration = Narration{"en-US": "The agent signs the request.", "fr-FR": "L'agent signe la requête."}
	f.Representations = map[string]RepresentationNarration{
		RepresentationCurl: {Narration: Narration{"en-US": "This curl command replays the signed request."}},
	}
	return p
}

func TestNarrationValid(t *testing.T) {
	if errs := narratedProtocol().Validate(); len(errs) > 0 {
		t.Fatalf("Validate() = %v, want no errors", errs)
	}
}

func TestNarrationText(t *testing.T) {
	n := Narration{"en-US": "hello", "fr-FR": "bonjour", "de-DE": ""}
	tests := []struct {
		lang, source, want string
		ok                 bool
	}{
		{"fr-FR", "en-US", "bonjour", true},
		{"zh-Hans", "en-US", "hello", true},
		{"de-DE", "en-US", "hello", true}, // empty text falls back
		{"zh-Hans", "ja-JP", "", false},
	}
	for _, tt := range tests {
		got, ok := n.Text(tt.lang, tt.source)
		if got != tt.want || ok != tt.ok {
			t.Errorf("Text(%q, %q) = %q, %v; want %q, %v", tt.lang, tt.source, got, ok, tt.want, tt.ok)
		}
	}
	if got := n.Languages(); !reflect.DeepEqual(got, []string{"de-DE", "en-US", "fr-FR"}) {
		t.Errorf("Languages() = %v", got)
	}
}

func TestStepNarration(t *testing.T) {
	p := narratedProtocol()
	f := p.Scenarios[0].Flows[0]
	tests := []struct {
		rep, lang, want string
	}{
		{RepresentationCurl, "en-US", "This curl command replays the signed request."},
		{RepresentationCurl, "fr-FR", "This curl command replays the signed request."}, // override falls back to its source language
		{RepresentationHTTP, "fr-FR", "L'agent signe la requête."},
		{"go", "zh-Hans", "The agent signs the request."},
	}
	for _, tt := range tests {
		if got, ok := p.StepNarration(f, tt.rep, tt.lang); !ok || got != tt.want {
			t.Errorf("StepNarration(%q, %q) = %q, %v; want %q", tt.rep, tt.lang, got, ok, tt.want)
		}
	}
	if _, ok := p.StepNarration(p.Scenarios[0].Flows[1], RepresentationHTTP, "en-US"); ok {
		t.Error("StepNarration on a flow without narration reported ok")
	}
	p.ProtocolMeta.Language = ""
	if got := p.SourceLanguage(); got != DefaultLanguage {
		t.Errorf("SourceLanguage() = %q, want %q", got, DefaultLanguage)
	}
}

func TestValidateNarrationErrors(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(p *Protocol)
		wantField string
		wantMsg   string
	}{
		{"bad source language", func(p *Protocol) { p.ProtocolMeta.Language = "english" }, "protocol.language", "not a BCP 47"},
		{"bad narration tag", func(p *Protocol) { p.Scenarios[0].Narration["fr_FR"] = "x" }, `scenarios[0].narration["fr_FR"]`, "not a BCP 47"},
		{"empty narration text", func(p *Protocol) { p.Scenarios[0].Narration["fr-FR"] = "" }, `scenarios[0].narration["fr-FR"]`, "must not be empty"},
		{"missing source language", func(p *Protocol) { delete(p.Scenarios[0].Flows[0].Narration, "en-US") }, "scenarios[0].flows[0].narration", `source language "en-US"`},
		{"briefing narration without source", func(p *Protocol) { p.Scenarios[0].Briefing.Narration = Narration{"fr-FR": "x"} }, "scenarios[0].briefing.narration", "source language"},
		{"empty briefing point", func(p *Protocol) { p.Scenarios[0].Briefing.Points[1] = "" }, "scenarios[0].briefing.points[1]", "must not be empty"},
		{"unknown representation", func(p *Protocol) {
			p.Scenarios[0].Flows[0].Representations["ruby"] = RepresentationNarration{Narration: Narration{"en-US": "x"}}
		}, `scenarios[0].flows[0].representations["ruby"]`, "unknown representation"},
		{"bad pronunciation tag", func(p *Protocol) { p.Pronunciations["agent_jkt"]["e"] = "x" }, `pronunciations["agent_jkt"]["e"]`, "not a BCP 47"},
		{"empty pronunciation", func(p *Protocol) { p.Pronunciations["agent_jkt"]["fr-FR"] = "" }, `pronunciations["agent_jkt"]["fr-FR"]`, "must not be empty"},
		{"narration on include step", func(p *Protocol) { p.Scenarios[2].Flows[2].Narration = Narration{"en-US": "x"} }, "scenarios[2].flows[2]", "include step cannot set narration"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := narratedProtocol()
			tt.mutate(p)
			errs := p.Validate()
			for _, e := range errs {
				if e.Field == tt.wantField && strings.Contains(e.Message, tt.wantMsg) {
					return
				}
			}
			t.Fatalf("Validate() = %v, want %s: %s", errs, tt.wantField, tt.wantMsg)
		})
	}
}

func TestWalkthroughNarration(t *testing.T) {
	w := narratedProtocol().Walkthrough(nil)
	if w.Language != "en-US" || w.Pronunciations["agent_jkt"]["fr-FR"] != "agent J K T" {
		t.Fatalf("language %q, pronunciations %v", w.Language, w.Pronunciations)
	}
	for _, sc := range w.Scenarios {
		if sc.ID != "agent-identity" {
			continue
		}
		if sc.Narration["fr-FR"] == "" || sc.Briefing == nil || len(sc.Briefing.Points) != 2 {
			t.Fatalf("scenario narration %v, briefing %+v", sc.Narration, sc.Briefing)
		}
		st := sc.Steps[0]
		if st.Narration["en-US"] == "" || st.Representations[RepresentationCurl].Narration["en-US"] == "" {
			t.Fatalf("step narration %v, representations %v", st.Narration, st.Representations)
		}
		return
	}
	t.Fatal("agent-identity not in walkthrough")
}

func TestResolvePronunciations(t *testing.T) {
	dir := t.TempDir()
	base := narratedProtocol()
	base.ProtocolMeta.Language = "fr-FR"
	base.Pronunciations = map[string]map[string]string{
		"agent_jkt": {"en-US": "base form", "fr-FR": "forme de base"},
		"AAuth":     {"en-US": "ay auth"},
	}
	// Base narration must include its own source language, fr-FR.
	for i := range base.Scenarios {
		base.Scenarios[i].Narration = nil
		base.Scenarios[i].Briefing = nil
		for j := range base.Scenarios[i].Flows {
			base.Scenarios[i].Flows[j].Narration = nil
			base.Scenarios[i].Flows[j].Representations = nil
		}
	}
	writeProtocol(t, dir, "base.json", base)
	derived := &Protocol{
		ProtocolMeta:   ProtocolMeta{ID: "derived", Name: "Derived"},
		Extends:        &ProtocolExtends{Path: "base.json"},
		Pronunciations: map[string]map[string]string{"agent_jkt": {"en-US": "agent J K T"}},
	}
	got := resolveIn(t, dir, derived)
	want := map[string]map[string]string{
		"agent_jkt": {"en-US": "agent J K T", "fr-FR": "forme de base"},
		"AAuth":     {"en-US": "ay auth"},
	}
	if !reflect.DeepEqual(got.Pronunciations, want) {
		t.Fatalf("pronunciations = %v, want %v", got.Pronunciations, want)
	}
	if got.ProtocolMeta.Language != "fr-FR" {
		t.Errorf("language = %q, want the base's fr-FR", got.ProtocolMeta.Language)
	}
	if base.Pronunciations["agent_jkt"]["en-US"] != "base form" {
		t.Error("resolve mutated the base pronunciations")
	}
}
