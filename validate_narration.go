package pidl

import (
	"fmt"
	"sort"
)

// validateNarration checks the source language, pronunciations, and every
// narration object: language tags, non-empty text, and text in the source
// language, which is the fallback for every other language.
func (p *Protocol) validateNarration() ValidationErrors {
	var errs ValidationErrors
	if l := p.ProtocolMeta.Language; l != "" && !languageTagPattern.MatchString(l) {
		errs = append(errs, ValidationError{Field: "protocol.language", Message: fmt.Sprintf("%q is not a BCP 47 language tag", l)})
	}
	source := p.SourceLanguage()

	terms := make([]string, 0, len(p.Pronunciations))
	for term := range p.Pronunciations {
		terms = append(terms, term)
	}
	sort.Strings(terms)
	for _, term := range terms {
		field := fmt.Sprintf("pronunciations[%q]", term)
		if term == "" {
			errs = append(errs, ValidationError{Field: "pronunciations", Message: "term must not be empty"})
		}
		errs = append(errs, validateLanguageMap(field, p.Pronunciations[term])...)
	}

	for i, sc := range p.Scenarios {
		field := fmt.Sprintf("scenarios[%d]", i)
		errs = append(errs, validateNarrationText(field+".narration", sc.Narration, source)...)
		if b := sc.Briefing; b != nil {
			errs = append(errs, validateNarrationText(field+".briefing.narration", b.Narration, source)...)
			for j, pt := range b.Points {
				if pt == "" {
					errs = append(errs, ValidationError{Field: fmt.Sprintf("%s.briefing.points[%d]", field, j), Message: "must not be empty"})
				}
			}
		}
	}

	for _, list := range p.flowLists() {
		for j, f := range list.flows {
			field := fmt.Sprintf("%s[%d]", list.field, j)
			errs = append(errs, validateNarrationText(field+".narration", f.Narration, source)...)
			reps := make([]string, 0, len(f.Representations))
			for rep := range f.Representations {
				reps = append(reps, rep)
			}
			sort.Strings(reps)
			for _, rep := range reps {
				rf := fmt.Sprintf("%s.representations[%q]", field, rep)
				if !validRepresentation(rep) && !sampleLanguages(f)[rep] {
					errs = append(errs, ValidationError{Field: rf, Message: fmt.Sprintf("unknown representation %q; use %q, %q or a sample language on this step", rep, RepresentationHTTP, RepresentationCurl)})
				}
				errs = append(errs, validateNarrationText(rf+".narration", f.Representations[rep].Narration, source)...)
			}
		}
	}
	return errs
}

// validRepresentation reports whether rep is a representation every step
// with an example has. Sample languages on the step are representations
// too.
func validRepresentation(rep string) bool {
	return rep == RepresentationHTTP || rep == RepresentationCurl
}

// validateNarrationText checks a narration object. An empty one is valid.
func validateNarrationText(field string, n Narration, source string) ValidationErrors {
	if len(n) == 0 {
		return nil
	}
	errs := validateLanguageMap(field, n)
	if _, ok := n[source]; !ok {
		errs = append(errs, ValidationError{Field: field, Message: fmt.Sprintf("must include the source language %q", source)})
	}
	return errs
}

// validateLanguageMap checks that every key is a language tag and every
// value is non-empty, reporting keys in sorted order.
func validateLanguageMap(field string, m map[string]string) ValidationErrors {
	var errs ValidationErrors
	langs := make([]string, 0, len(m))
	for l := range m {
		langs = append(langs, l)
	}
	sort.Strings(langs)
	for _, l := range langs {
		lf := fmt.Sprintf("%s[%q]", field, l)
		if !languageTagPattern.MatchString(l) {
			errs = append(errs, ValidationError{Field: lf, Message: fmt.Sprintf("%q is not a BCP 47 language tag", l)})
		}
		if m[l] == "" {
			errs = append(errs, ValidationError{Field: lf, Message: "must not be empty"})
		}
	}
	return errs
}

// validateAllSamples checks scenario samples and example samples.
func (p *Protocol) validateAllSamples() ValidationErrors {
	entityIDs := make(map[string]bool, len(p.Entities))
	for _, e := range p.Entities {
		entityIDs[e.ID] = true
	}
	var errs ValidationErrors
	for i, sc := range p.Scenarios {
		errs = append(errs, validateSamples(fmt.Sprintf("scenarios[%d].samples", i), sc.Samples, entityIDs)...)
	}
	for _, list := range p.flowLists() {
		for j, f := range list.flows {
			if f.Example != nil {
				errs = append(errs, validateSamples(fmt.Sprintf("%s[%d].example.samples", list.field, j), f.Example.Samples, entityIDs)...)
			}
		}
	}
	return errs
}
