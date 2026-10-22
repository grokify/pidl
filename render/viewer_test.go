package render_test

import (
	"encoding/json"
	"regexp"
	"strings"
	"testing"

	"github.com/grokify/pidl"
	"github.com/grokify/pidl/examples"
	"github.com/grokify/pidl/render"
)

var dataElement = regexp.MustCompile(`<script type="application/json" id="pidl-walkthrough">(.*?)</script>`)

func TestViewerEmbedsWalkthrough(t *testing.T) {
	p, err := examples.GetProtocol("oauth2_api_access")
	if err != nil {
		t.Fatalf("GetProtocol: %v", err)
	}
	page, err := render.RenderString(render.FormatViewer, p)
	if err != nil {
		t.Fatalf("RenderString: %v", err)
	}
	if !strings.Contains(page, "<title>OAuth 2.0 API Access walkthrough</title>") {
		t.Error("title not set from the protocol name")
	}
	m := dataElement.FindStringSubmatch(page)
	if m == nil {
		t.Fatal("walkthrough data element not found")
	}
	var w pidl.Walkthrough
	if err := json.Unmarshal([]byte(m[1]), &w); err != nil {
		t.Fatalf("embedded data is not walkthrough JSON: %v", err)
	}
	if w.Name != "OAuth 2.0 API Access" || len(w.Scenarios) != 7 {
		t.Errorf("embedded walkthrough = %s with %d scenarios", w.Name, len(w.Scenarios))
	}
	if render.FormatViewer.FileExtension() != ".html" || !render.IsDocumentFormat(render.FormatViewer) {
		t.Error("viewer should be a whole-document .html format")
	}
}

func TestViewerEscapesScriptBreakouts(t *testing.T) {
	p := &pidl.Protocol{
		ProtocolMeta: pidl.ProtocolMeta{ID: "x", Name: `Evil </script><script>alert(1)</script> & co`},
		Entities:     []pidl.Entity{{ID: "a", Name: "A", Type: pidl.EntityTypeClient}, {ID: "b", Name: "B", Type: pidl.EntityTypeServer}},
		Flows:        []pidl.Flow{{From: "a", To: "b", Action: "x", Label: "</script><b>"}},
	}
	page, err := render.RenderString(render.FormatViewer, p)
	if err != nil {
		t.Fatalf("RenderString: %v", err)
	}
	m := dataElement.FindStringSubmatch(page)
	if m == nil || strings.Contains(m[1], "</script") || strings.Contains(m[1], "<b>") {
		t.Fatal("protocol text escaped the data element")
	}
	if strings.Contains(page, "<title>Evil </script>") {
		t.Fatal("title not HTML-escaped")
	}
}
