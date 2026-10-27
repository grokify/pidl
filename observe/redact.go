package observe

import (
	"encoding/json"
	"strings"

	"github.com/grokify/pidl"
)

// Masks put in place of redacted values.
const (
	MaskJWT   = "[redacted:jwt]"
	MaskValue = "[redacted]"
)

// masker replaces raw token, signature and credential values in observed
// messages. Decoded token headers and claims are kept: they are what a
// reader learns from. A nil masker leaves messages as they are.
type masker struct{}

func newMasker() *masker { return &masker{} }

// mask redacts a message in place. tokens are raw token values found by
// the token definitions, masked wherever they appear even when they are not
// JWTs.
func (m *masker) mask(msg *pidl.ObservedMessage, tokens []string) {
	if m == nil {
		return
	}
	tokens = sortedLongestFirst(tokens)
	// JWTs first, so they keep the JWT mask; then the other token values
	// the definitions found, such as opaque keys.
	text := func(s string) string {
		s = jwtPattern.ReplaceAllString(s, MaskJWT)
		for _, t := range tokens {
			if t != "" && !jwtPattern.MatchString(t) {
				s = strings.ReplaceAll(s, t, MaskValue)
			}
		}
		return s
	}
	headers := func(h []pidl.ExampleHeader) {
		for i := range h {
			h[i].Value = maskHeader(h[i].Name, text(h[i].Value))
		}
	}
	body := func(b json.RawMessage) json.RawMessage {
		if len(b) == 0 {
			return b
		}
		// Masks are plain text, so replacing inside the JSON text keeps it
		// valid and keeps the members' order.
		return json.RawMessage(text(string(b)))
	}
	if r := msg.Request; r != nil {
		cp := *r
		cp.URL = text(cp.URL)
		cp.Headers = append([]pidl.ExampleHeader(nil), cp.Headers...)
		headers(cp.Headers)
		cp.Body = body(cp.Body)
		msg.Request = &cp
	}
	if r := msg.Response; r != nil {
		cp := *r
		cp.Headers = append([]pidl.ExampleHeader(nil), cp.Headers...)
		headers(cp.Headers)
		cp.Body = body(cp.Body)
		msg.Response = &cp
	}
	msg.Label = text(msg.Label)
	if len(msg.Annotations) > 0 {
		anns := append([]pidl.ExchangeAnnotation(nil), msg.Annotations...)
		for i := range anns {
			anns[i].Text = text(anns[i].Text)
		}
		msg.Annotations = anns
	}
}

// maskHeader masks credentials by header: a signature's value, and the
// credentials after the scheme in Authorization or in a cookie.
func maskHeader(name, value string) string {
	switch strings.ToLower(name) {
	case "signature":
		label, _, found := strings.Cut(value, "=")
		if found {
			return label + "=:" + MaskValue + ":"
		}
		return MaskValue
	case "authorization", "proxy-authorization":
		if scheme, _, found := strings.Cut(value, " "); found {
			return scheme + " " + MaskValue
		}
		return MaskValue
	case "cookie", "set-cookie":
		return MaskValue
	}
	return value
}
