package pidl

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// WireBody returns the bytes an example body stands for on the wire, which
// is what a Content-Digest covers. A JSON string is sent as its contents,
// for media types that are not JSON; any other value is sent as compact
// JSON with its keys in written order.
func WireBody(body json.RawMessage) ([]byte, error) {
	trimmed := bytes.TrimSpace(body)
	if len(trimmed) == 0 {
		return nil, nil
	}
	if trimmed[0] == '"' {
		var s string
		if err := json.Unmarshal(trimmed, &s); err != nil {
			return nil, fmt.Errorf("decoding string body: %w", err)
		}
		return []byte(s), nil
	}
	var buf bytes.Buffer
	if err := json.Compact(&buf, trimmed); err != nil {
		return nil, fmt.Errorf("compacting JSON body: %w", err)
	}
	return buf.Bytes(), nil
}

// signatureHeaders are replayed, not created, by a generated command.
var signatureHeaders = map[string]bool{"signature": true, "signature-input": true}

// Curl returns a curl command equivalent to the request: method, URL, each
// header in order, and the wire body. curl cannot create HTTP message
// signatures, so when the request is signed a leading comment says the
// signature headers are replayed from the example.
func (r *ExampleRequest) Curl() (string, error) {
	if r == nil {
		return "", nil
	}
	body, err := WireBody(r.Body)
	if err != nil {
		return "", err
	}

	var sb strings.Builder
	for _, h := range r.Headers {
		if signatureHeaders[strings.ToLower(h.Name)] {
			sb.WriteString("# The signature headers are replayed from the example; curl cannot create HTTP message signatures.\n")
			break
		}
	}
	sb.WriteString("curl")
	if r.Method != "" && (r.Method != "GET" || body != nil) {
		sb.WriteString(" -X " + r.Method)
	}
	sb.WriteString(" " + shellQuote(r.URL))
	for _, h := range r.Headers {
		sb.WriteString(" \\\n  -H " + shellQuote(h.Name+": "+h.Value))
	}
	if body != nil {
		sb.WriteString(" \\\n  --data-raw " + shellQuote(string(body)))
	}
	return sb.String(), nil
}

// shellQuote quotes s for a POSIX shell.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
