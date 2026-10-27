package sandbox

import (
	"bytes"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/grokify/pidl"
)

var update = flag.Bool("update", false, "rewrite golden files")

// petstore.trace.json is a real sandbox trace of the Petstore run behind
// the sandbox: a health check, browsing pets, and ordering one with the
// person's approval. Tokens and signatures are truncated with an ellipsis.
const fixture = "testdata/petstore.trace.json"

func TestReadFileGolden(t *testing.T) {
	o, err := ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	got, err := json.MarshalIndent(o, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	got = append(got, '\n')
	golden := filepath.Join("testdata", "petstore.observation.json")
	if *update {
		if err := os.WriteFile(golden, got, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(golden)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Errorf("observation differs from %s; run go test -update to rewrite it after checking the change", golden)
	}
}

func TestReadFileStructure(t *testing.T) {
	o, err := ReadFile(fixture)
	if err != nil {
		t.Fatal(err)
	}
	if errs := o.Validate(); len(errs) > 0 {
		t.Fatalf("observation does not validate: %v", errs)
	}
	if o.Source.Tool != Tool || o.Source.RequestFidelity != pidl.RequestFidelityWire {
		t.Errorf("source = %+v", o.Source)
	}
	if len(o.Exchanges) != 19 {
		t.Fatalf("exchanges = %d, want 19", len(o.Exchanges))
	}
	first, last := o.Exchanges[0].Time, o.Exchanges[len(o.Exchanges)-1].Time
	if !o.Source.From.Equal(*first) || !o.Source.To.Equal(*last) {
		t.Errorf("source range = %v to %v, want %v to %v", o.Source.From, o.Source.To, first, last)
	}
	labels := map[string]bool{}
	for _, e := range o.Exchanges {
		labels[e.From] = true
		labels[e.To] = true
	}
	var got []string
	for l := range labels {
		got = append(got, l)
	}
	sort.Strings(got)
	if want := "agent ap client petstore.aauth.test ps"; strings.Join(got, " ") != want {
		t.Errorf("labels = %v, want %s: the sandbox's own names, kept as written", got, want)
	}
	order := o.Exchanges[6]
	if order.Request.Method != "POST" || !strings.HasSuffix(order.Request.URL, "/v2/store/order") || order.Response.Status != 401 {
		t.Errorf("exchange 7 = %s %s %d, want the first order's challenge", order.Request.Method, order.Request.URL, order.Response.Status)
	}
	if v, ok := order.Response.Headers.Get("AAuth-Requirement"); !ok || v != "requirement=person-token" {
		t.Errorf("challenge header = %q, %v", v, ok)
	}
	if order.Request.Body == nil || o.Exchanges[2].Request.Body != nil {
		t.Error("a request body was dropped, or a null body kept")
	}
}

func TestReadErrors(t *testing.T) {
	if _, err := Read(strings.NewReader(`{"not": "a list"}`)); err == nil {
		t.Error("an object was read as a trace")
	}
	if _, err := ReadFile(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Error("a missing file was read")
	}
	o, err := Read(strings.NewReader(`[]`))
	if err != nil {
		t.Fatal(err)
	}
	if len(o.Exchanges) != 0 || o.Source.From != nil {
		t.Errorf("empty trace = %+v", o)
	}
}

func TestReadKeepsOrderWithoutTimes(t *testing.T) {
	o, err := Read(strings.NewReader(`[
	  {"from": "agent", "to": "ap", "request": {"method": "POST", "url": "https://ap.example.test/a"}, "response": {"status": 200}},
	  {"from": "agent", "to": "ps", "request": {"method": "GET", "url": "https://ps.example.test/b"}, "response": {"status": 204}}
	]`))
	if err != nil {
		t.Fatal(err)
	}
	if o.Exchanges[0].Time != nil || o.Exchanges[1].To != "ps" || o.Source.From != nil {
		t.Errorf("exchanges without times: %+v", o.Exchanges)
	}
}
