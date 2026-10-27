package pidl

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime"
	"net/url"
	"os"
	"strings"
	"time"
	"unicode/utf8"
)

// ObservationFormat names the observation format and its major version.
const ObservationFormat = "pidl-observation/v1"

// DefaultBodyLimit is the size above which NewObservedBody cuts a body.
const DefaultBodyLimit = 64 << 10

// Observation is a captured run of a real system: the HTTP exchanges it
// made, in order, whatever tool captured them. Generation matches its
// exchanges to a document's steps, using a Binding for the deployment, and
// shows them beside the steps' examples.
type Observation struct {
	// Format names the format and its major version: pidl-observation/v1.
	Format string `json:"format"`

	// Source says which tool produced the observation and what it covers.
	Source ObservationSource `json:"source"`

	// Exchanges are the requests and responses, in the order they happened.
	Exchanges []Exchange `json:"exchanges"`
}

// RequestFidelity says whether an observation's requests are the messages
// servers received or what a client sent before a proxy changed them.
type RequestFidelity string

const (
	// RequestFidelityWire marks requests as the messages on the wire, as a
	// server's trace or a proxy in front of the servers records them.
	RequestFidelityWire RequestFidelity = "wire"
	// RequestFidelityClient marks requests as what a client sent, before a
	// proxy such as a signing helper changed them.
	RequestFidelityClient RequestFidelity = "client"
)

// ObservationSource describes where an observation came from.
type ObservationSource struct {
	// Tool names the tool that captured the exchanges, such as
	// "aauth-sandbox", "har" or "newman".
	Tool string `json:"tool"`

	// RequestFidelity is wire or client; empty when unknown.
	RequestFidelity RequestFidelity `json:"requestFidelity,omitempty"`

	// From and To bound the time the observation covers.
	From *time.Time `json:"from,omitempty"`
	To   *time.Time `json:"to,omitempty"`
}

// Exchange is one HTTP request and its response.
type Exchange struct {
	// Time is when the request was made. A tool that has no clock keeps the
	// order and leaves it unset.
	Time *time.Time `json:"time,omitempty"`

	// From and To are the capturing tool's labels for the sender and the
	// receiver, kept as written. A binding maps them to participants.
	From string `json:"from,omitempty"`
	To   string `json:"to,omitempty"`

	// RequestID is the value of a request ID header, when the capture has
	// one. It joins client events and other sources to the exchange.
	RequestID string `json:"requestId,omitempty"`

	Request  ObservedRequest  `json:"request"`
	Response ObservedResponse `json:"response"`

	// BodyTruncated reports that a body was cut at the size limit.
	BodyTruncated bool `json:"bodyTruncated,omitempty"`

	// Label is a name a tool gave the exchange, such as a Newman request's
	// name. It never changes a match.
	Label string `json:"label,omitempty"`

	// Annotations are evidence a tool produced about the exchange, such as
	// assertions or a client's own events. They never change a match.
	Annotations []ExchangeAnnotation `json:"annotations,omitempty"`
}

// ObservedRequest is a captured HTTP request.
type ObservedRequest struct {
	Method  string          `json:"method"`
	URL     string          `json:"url"`
	Headers ObservedHeaders `json:"headers,omitempty"`
	// Body is a JSON value for a JSON body, a JSON string for other text,
	// or {"base64": "…", "mediaType": "…"} for binary content.
	Body json.RawMessage `json:"body,omitempty"`
}

// ObservedResponse is a captured HTTP response.
type ObservedResponse struct {
	Status  int             `json:"status"`
	Headers ObservedHeaders `json:"headers,omitempty"`
	// Body is encoded as ObservedRequest.Body is.
	Body json.RawMessage `json:"body,omitempty"`
}

// ObservedHeaders are header fields as name and value pairs, in their
// original order and case.
type ObservedHeaders [][2]string

// Get returns the first value of the named header, comparing names without
// regard to case.
func (h ObservedHeaders) Get(name string) (string, bool) {
	for _, kv := range h {
		if strings.EqualFold(kv[0], name) {
			return kv[1], true
		}
	}
	return "", false
}

// ExchangeAnnotation is a piece of evidence about an exchange.
type ExchangeAnnotation struct {
	// Kind classifies the annotation, such as "assertion" or "client-event".
	Kind string `json:"kind"`

	// Text is the annotation as a person reads it.
	Text string `json:"text"`

	// OK is the result of an assertion; unset for other kinds.
	OK *bool `json:"ok,omitempty"`

	// Time is when the annotated event happened, when it differs from the
	// exchange's time.
	Time *time.Time `json:"time,omitempty"`
}

// BinaryBody is the encoding of a body that is not text.
type BinaryBody struct {
	Base64    string `json:"base64"`
	MediaType string `json:"mediaType,omitempty"`
}

// NewObservedBody encodes a captured body: as a JSON value when it is JSON,
// as a JSON string when it is other text, and as a BinaryBody otherwise. A
// body longer than limit bytes is cut to limit, in which case truncated is
// true and a JSON body is kept as text. A limit of zero or less means
// DefaultBodyLimit. An empty body gives a nil result.
func NewObservedBody(data []byte, mediaType string, limit int) (body json.RawMessage, truncated bool, err error) {
	if len(data) == 0 {
		return nil, false, nil
	}
	if limit <= 0 {
		limit = DefaultBodyLimit
	}
	if len(data) > limit {
		data, truncated = data[:limit], true
	}
	if !truncated && isJSONMediaType(mediaType) && json.Valid(data) {
		var buf bytes.Buffer
		if err := json.Compact(&buf, data); err != nil {
			return nil, false, err
		}
		return buf.Bytes(), false, nil
	}
	if isTextMediaType(mediaType) || (mediaType == "" && utf8.Valid(data)) {
		text := string(data)
		if truncated {
			text = strings.ToValidUTF8(text, "")
		}
		b, err := json.Marshal(text)
		return b, truncated, err
	}
	b, err := json.Marshal(BinaryBody{Base64: base64.StdEncoding.EncodeToString(data), MediaType: mediaType})
	return b, truncated, err
}

func isJSONMediaType(mediaType string) bool {
	mt, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		return false
	}
	return mt == "application/json" || strings.HasSuffix(mt, "+json")
}

func isTextMediaType(mediaType string) bool {
	mt, _, err := mime.ParseMediaType(mediaType)
	if err != nil {
		return false
	}
	return strings.HasPrefix(mt, "text/") || isJSONMediaType(mediaType) ||
		mt == "application/x-www-form-urlencoded" || strings.HasSuffix(mt, "+xml") || mt == "application/xml"
}

// IsObservation reports whether data is an observation, by its format.
func IsObservation(data []byte) bool {
	var head struct {
		Format string `json:"format"`
	}
	return json.Unmarshal(data, &head) == nil && strings.HasPrefix(head.Format, "pidl-observation/")
}

// ParseObservation parses an observation. It refuses another major version
// of the format.
func ParseObservation(data []byte) (*Observation, error) {
	var o Observation
	if err := json.Unmarshal(data, &o); err != nil {
		return nil, err
	}
	if o.Format != ObservationFormat {
		return nil, fmt.Errorf("observation format %q is not %s", o.Format, ObservationFormat)
	}
	return &o, nil
}

// ReadObservation reads an observation file.
func ReadObservation(path string) (*Observation, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- the path is chosen by the caller
	if err != nil {
		return nil, err
	}
	o, err := ParseObservation(data)
	if err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return o, nil
}

// WriteFile writes the observation as indented JSON.
func (o *Observation) WriteFile(path string) error {
	data, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o600)
}

// Validate checks the observation's structure: a format, a source tool,
// and for each exchange a method, an absolute URL and a response status.
func (o *Observation) Validate() ValidationErrors {
	var errs ValidationErrors
	if o.Format != ObservationFormat {
		errs = append(errs, ValidationError{Field: "format", Message: fmt.Sprintf("must be %s", ObservationFormat)})
	}
	if o.Source.Tool == "" {
		errs = append(errs, ValidationError{Field: "source.tool", Message: "required"})
	}
	switch o.Source.RequestFidelity {
	case "", RequestFidelityWire, RequestFidelityClient:
	default:
		errs = append(errs, ValidationError{Field: "source.requestFidelity", Message: fmt.Sprintf("must be %s or %s", RequestFidelityWire, RequestFidelityClient)})
	}
	if o.Source.From != nil && o.Source.To != nil && o.Source.To.Before(*o.Source.From) {
		errs = append(errs, ValidationError{Field: "source.to", Message: "is before source.from"})
	}
	for i, e := range o.Exchanges {
		field := fmt.Sprintf("exchanges[%d]", i)
		if e.Request.Method == "" {
			errs = append(errs, ValidationError{Field: field + ".request.method", Message: "required"})
		}
		if u, err := url.Parse(e.Request.URL); err != nil || u.Scheme == "" || u.Host == "" {
			errs = append(errs, ValidationError{Field: field + ".request.url", Message: "must be an absolute URL"})
		}
		if e.Response.Status < 100 || e.Response.Status > 599 {
			errs = append(errs, ValidationError{Field: field + ".response.status", Message: "must be an HTTP status from 100 to 599"})
		}
		errs = append(errs, validateObservedHeaders(field+".request.headers", e.Request.Headers)...)
		errs = append(errs, validateObservedHeaders(field+".response.headers", e.Response.Headers)...)
		for _, b := range []struct {
			field string
			body  json.RawMessage
		}{{field + ".request.body", e.Request.Body}, {field + ".response.body", e.Response.Body}} {
			if len(b.body) > 0 && !json.Valid(b.body) {
				errs = append(errs, ValidationError{Field: b.field, Message: "must be JSON"})
			}
		}
		for j, a := range e.Annotations {
			if a.Kind == "" {
				errs = append(errs, ValidationError{Field: fmt.Sprintf("%s.annotations[%d].kind", field, j), Message: "required"})
			}
		}
	}
	return errs
}

func validateObservedHeaders(field string, h ObservedHeaders) ValidationErrors {
	var errs ValidationErrors
	for i, kv := range h {
		if kv[0] == "" {
			errs = append(errs, ValidationError{Field: fmt.Sprintf("%s[%d]", field, i), Message: "header name required"})
		}
	}
	return errs
}
