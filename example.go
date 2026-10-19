package pidl

import "encoding/json"

// ExampleKind says where a flow example comes from.
type ExampleKind string

const (
	// ExampleKindDraft is taken from the cited source, with values aligned
	// to the scenario so steps link up.
	ExampleKindDraft ExampleKind = "draft"
	// ExampleKindDerived is written from the cited source's normative text
	// where the source has no example.
	ExampleKindDerived ExampleKind = "derived"
)

// FlowExample is an illustrative message for one step: the HTTP request,
// the HTTP response and decoded tokens. Examples are for explanation; they
// are not message schemas.
type FlowExample struct {
	// Source cites where the example comes from, such as "draft-11 §7.2.1".
	Source string `json:"source"`

	// Kind is draft or derived.
	Kind ExampleKind `json:"kind"`

	// Request is the HTTP request, for steps that send one.
	Request *ExampleRequest `json:"request,omitempty"`

	// Response is the HTTP response, for steps that return one.
	Response *ExampleResponse `json:"response,omitempty"`

	// Tokens are decoded tokens the message carries, in display order.
	Tokens []ExampleToken `json:"tokens,omitempty"`

	// Note explains anything the reader should notice in this example.
	Note string `json:"note,omitempty"`
}

// ExampleRequest is an illustrative HTTP request.
type ExampleRequest struct {
	Method  string          `json:"method"`
	URL     string          `json:"url"`
	Headers []ExampleHeader `json:"headers,omitempty"`
	// Body is the request body as JSON, kept byte for byte so key order
	// survives a round trip.
	Body json.RawMessage `json:"body,omitempty"`
}

// ExampleResponse is an illustrative HTTP response.
type ExampleResponse struct {
	Status  int             `json:"status"`
	Headers []ExampleHeader `json:"headers,omitempty"`
	Body    json.RawMessage `json:"body,omitempty"`
}

// ExampleHeader is one HTTP header. Headers are a list, not a map, so they
// keep the order they are written in.
type ExampleHeader struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

// ExampleToken is a decoded token carried by a message.
type ExampleToken struct {
	// Token is the token definition ID.
	Token string `json:"token"`
	// Label says where the token travels, such as "Signature-Key header".
	Label string `json:"label,omitempty"`
	// Header and Payload are the decoded JWT header and claims.
	Header  json.RawMessage `json:"header,omitempty"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

// IsValidExampleKind reports whether k is a known example kind.
func IsValidExampleKind(k ExampleKind) bool {
	return k == ExampleKindDraft || k == ExampleKindDerived
}
