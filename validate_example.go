package pidl

import (
	"encoding/json"
	"fmt"
)

// validateExamples checks every flow example in the document.
func (p *Protocol) validateExamples() ValidationErrors {
	tokenIDs := make(map[string]bool)
	if p.Metadata != nil {
		for _, t := range p.Metadata.Tokens {
			tokenIDs[t.ID] = true
		}
	}
	var errs ValidationErrors
	for _, list := range p.flowLists() {
		for i, f := range list.flows {
			if f.Example != nil {
				errs = append(errs, validateExample(fmt.Sprintf("%s[%d].example", list.field, i), f.Example, tokenIDs)...)
			}
		}
	}
	return errs
}

func validateExample(field string, ex *FlowExample, tokenIDs map[string]bool) ValidationErrors {
	var errs ValidationErrors
	add := func(f, msg string) { errs = append(errs, ValidationError{Field: field + f, Message: msg}) }

	if ex.Source == "" {
		add(".source", "required")
	}
	switch {
	case ex.Kind == "":
		add(".kind", "required")
	case !IsValidExampleKind(ex.Kind):
		add(".kind", fmt.Sprintf("invalid example kind %q (want draft or derived)", ex.Kind))
	}
	if ex.Request == nil && ex.Response == nil && len(ex.Tokens) == 0 {
		add("", "needs a request, a response or tokens")
	}
	if r := ex.Request; r != nil {
		if r.Method == "" {
			add(".request.method", "required")
		}
		if r.URL == "" {
			add(".request.url", "required")
		}
		errs = append(errs, validateHeaders(field+".request.headers", r.Headers)...)
		errs = append(errs, validateRawJSON(field+".request.body", r.Body)...)
	}
	if r := ex.Response; r != nil {
		if r.Status < 100 || r.Status > 599 {
			add(".response.status", fmt.Sprintf("invalid HTTP status %d", r.Status))
		}
		errs = append(errs, validateHeaders(field+".response.headers", r.Headers)...)
		errs = append(errs, validateRawJSON(field+".response.body", r.Body)...)
	}
	for i, t := range ex.Tokens {
		tf := fmt.Sprintf("%s.tokens[%d]", field, i)
		switch {
		case t.Token == "":
			errs = append(errs, ValidationError{Field: tf + ".token", Message: "required"})
		case !tokenIDs[t.Token]:
			errs = append(errs, ValidationError{Field: tf + ".token", Message: fmt.Sprintf("unknown token definition %q", t.Token)})
		}
		errs = append(errs, validateRawJSON(tf+".header", t.Header)...)
		errs = append(errs, validateRawJSON(tf+".payload", t.Payload)...)
	}
	return errs
}

func validateHeaders(field string, headers []ExampleHeader) ValidationErrors {
	var errs ValidationErrors
	for i, h := range headers {
		if h.Name == "" {
			errs = append(errs, ValidationError{Field: fmt.Sprintf("%s[%d].name", field, i), Message: "required"})
		}
	}
	return errs
}

// validateRawJSON rejects an empty-but-present or malformed JSON value. The
// decoder already rejects malformed JSON when parsing a file; this guards
// protocols built in Go.
func validateRawJSON(field string, raw json.RawMessage) ValidationErrors {
	if raw == nil || json.Valid(raw) {
		return nil
	}
	return ValidationErrors{{Field: field, Message: "must be valid JSON"}}
}
