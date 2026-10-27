// Package sandbox reads the message trace of the aauth-go sandbox as a PIDL
// observation.
//
// The sandbox records every request and response between the roles it
// serves, including those a client never sees, so its trace holds the
// messages on the wire. Print it with "aauth-sandbox trace", or fetch
// /sandbox/trace from any of its role hosts.
//
// The trace's labels are the sandbox's names, not a document's
// participants: its roles ("agent", "ap", "ps"), a resource it fronts by
// host name, and "client" for any caller that is not one of its roles. A
// binding maps them.
package sandbox

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/grokify/pidl"
)

// Tool names the sandbox as an observation's source.
const Tool = "aauth-sandbox"

// exchange and message are the fields of the sandbox's trace the adapter
// reads.
type exchange struct {
	Time     time.Time `json:"time"`
	From     string    `json:"from"`
	To       string    `json:"to"`
	Request  message   `json:"request"`
	Response message   `json:"response"`
}

type message struct {
	Method  string          `json:"method"`
	URL     string          `json:"url"`
	Status  int             `json:"status"`
	Headers [][2]string     `json:"headers"`
	Body    json.RawMessage `json:"body"`
}

// Read reads a sandbox trace, a JSON list of exchanges, as an observation
// of the messages on the wire. Exchanges keep their order and labels.
func Read(r io.Reader) (*pidl.Observation, error) {
	var trace []exchange
	if err := json.NewDecoder(r).Decode(&trace); err != nil {
		return nil, fmt.Errorf("reading sandbox trace: %w", err)
	}
	o := &pidl.Observation{
		Format:    pidl.ObservationFormat,
		Source:    pidl.ObservationSource{Tool: Tool, RequestFidelity: pidl.RequestFidelityWire},
		Exchanges: make([]pidl.Exchange, 0, len(trace)),
	}
	for _, e := range trace {
		x := pidl.Exchange{
			From:     e.From,
			To:       e.To,
			Request:  pidl.ObservedRequest{Method: e.Request.Method, URL: e.Request.URL, Headers: e.Request.Headers, Body: nonNull(e.Request.Body)},
			Response: pidl.ObservedResponse{Status: e.Response.Status, Headers: e.Response.Headers, Body: nonNull(e.Response.Body)},
		}
		if !e.Time.IsZero() {
			t := e.Time
			x.Time = &t
			if o.Source.From == nil || t.Before(*o.Source.From) {
				o.Source.From = &t
			}
			if o.Source.To == nil || t.After(*o.Source.To) {
				o.Source.To = &t
			}
		}
		o.Exchanges = append(o.Exchanges, x)
	}
	return o, nil
}

// ReadFile reads a sandbox trace file.
func ReadFile(path string) (*pidl.Observation, error) {
	f, err := os.Open(path) // #nosec G304 -- the path is chosen by the caller
	if err != nil {
		return nil, err
	}
	o, readErr := Read(f)
	if err := f.Close(); err != nil && readErr == nil {
		return nil, err
	}
	if readErr != nil {
		return nil, fmt.Errorf("%s: %w", path, readErr)
	}
	return o, nil
}

// nonNull drops a JSON null body, which the trace writes for a message
// without one.
func nonNull(b json.RawMessage) json.RawMessage {
	if string(b) == "null" {
		return nil
	}
	return b
}
