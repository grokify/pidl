package observe

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"regexp"
	"sort"
	"strings"

	"github.com/grokify/pidl"
	"github.com/grokify/pidl/observe/match"
)

// jwtPattern finds a compact JWT: three base64url segments, the last
// possibly empty.
var jwtPattern = regexp.MustCompile(`eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+\.[A-Za-z0-9_-]*`)

// tokenFinder finds tokens in messages where the document's token
// definitions say they travel.
type tokenFinder struct {
	defs []tokenDef
}

type tokenDef struct {
	id     string
	places []pidl.TokenCarriage
}

func newTokenFinder(p *pidl.Protocol) *tokenFinder {
	f := &tokenFinder{}
	if p == nil || p.Metadata == nil {
		return f
	}
	for _, t := range p.Metadata.Tokens {
		if len(t.CarriedIn) > 0 {
			f.defs = append(f.defs, tokenDef{id: t.ID, places: t.CarriedIn})
		}
	}
	return f
}

// find returns the defined tokens a message carries, decoded but not
// verified, in the order of the token definitions.
func (f *tokenFinder) find(h pidl.ObservedHeaders, body json.RawMessage, part string) []pidl.ExampleToken {
	var out []pidl.ExampleToken
	for _, d := range f.defs {
		for _, c := range d.places {
			value, where := locate(c, h, body)
			if value == "" {
				continue
			}
			hdr, payload, ok := decodeJWT(value)
			if c.Typ != "" && (!ok || jwtTyp(hdr) != c.Typ) {
				continue
			}
			t := pidl.ExampleToken{Token: d.id, Label: where + " in the " + part + "; decoded, not verified"}
			if ok {
				t.Header, t.Payload = hdr, payload
			}
			out = append(out, t)
		}
	}
	return out
}

// values returns every raw token value an exchange carries, for masking.
func (f *tokenFinder) values(x *pidl.Exchange) []string {
	var out []string
	for _, d := range f.defs {
		for _, c := range d.places {
			for _, part := range []struct {
				h    pidl.ObservedHeaders
				body json.RawMessage
			}{{x.Request.Headers, x.Request.Body}, {x.Response.Headers, x.Response.Body}} {
				if v, _ := locate(c, part.h, part.body); v != "" {
					out = append(out, v)
				}
			}
		}
	}
	return out
}

// locate finds a token's value at one place, and names the place.
func locate(c pidl.TokenCarriage, h pidl.ObservedHeaders, body json.RawMessage) (string, string) {
	switch c.In {
	case "header":
		v, ok := h.Get(c.Name)
		if !ok {
			return "", ""
		}
		if c.Param != "" {
			p, ok := param(v, c.Param)
			if !ok {
				return "", ""
			}
			return p, c.Name + " header (" + c.Param + ")"
		}
		// A token after an authentication scheme, such as "Bearer <token>".
		if scheme, rest, found := strings.Cut(v, " "); found && scheme != "" && !strings.Contains(scheme, "=") {
			v = strings.TrimSpace(rest)
		}
		return v, c.Name + " header"
	case "body":
		var doc any
		dec := json.NewDecoder(bytes.NewReader(body))
		dec.UseNumber()
		if len(body) == 0 || dec.Decode(&doc) != nil {
			return "", ""
		}
		v, ok := match.Pointer(doc, c.Path)
		if s, isString := v.(string); ok && isString && s != "" {
			return s, c.Path + " in the body"
		}
	}
	return "", ""
}

// param returns a structured field parameter's value, as in
// `sig=jwt;jwt="eyJ…"` or `requirement=auth-token; resource-token="eyJ…"`.
func param(value, name string) (string, bool) {
	re := regexp.MustCompile(`(?:^|[;,\s])` + regexp.QuoteMeta(name) + `=(?:"((?:[^"\\]|\\.)*)"|([^;,\s]*))`)
	m := re.FindStringSubmatch(value)
	if m == nil {
		return "", false
	}
	if m[1] != "" {
		return m[1], true
	}
	return m[2], m[2] != ""
}

// decodeJWT splits a compact JWT and decodes its header and payload. It
// does not verify the signature.
func decodeJWT(token string) (json.RawMessage, json.RawMessage, bool) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, nil, false
	}
	hdr, err1 := base64.RawURLEncoding.DecodeString(parts[0])
	payload, err2 := base64.RawURLEncoding.DecodeString(parts[1])
	if err1 != nil || err2 != nil || !json.Valid(hdr) || !json.Valid(payload) {
		return nil, nil, false
	}
	return hdr, payload, true
}

// jwtTyp returns a decoded JWT header's typ.
func jwtTyp(hdr json.RawMessage) string {
	var h struct {
		Typ string `json:"typ"`
	}
	if json.Unmarshal(hdr, &h) != nil {
		return ""
	}
	return h.Typ
}

// sortedLongestFirst orders values so a longer one is replaced before a
// shorter one it contains.
func sortedLongestFirst(values []string) []string {
	out := append([]string(nil), values...)
	sort.SliceStable(out, func(i, j int) bool { return len(out[i]) > len(out[j]) })
	return out
}
