package render

//go:generate sh -c "cd ../viewer && npm ci && npm run build:embed"

import (
	_ "embed"
	"encoding/json"
	"errors"
	"html"
	"io"
	"strings"

	"github.com/grokify/pidl"
)

// viewerHTML is the single-file walkthrough viewer built from viewer/.
// Regenerate it with `go generate ./render`.
//
//go:embed viewer/viewer.html
var viewerHTML string

const (
	viewerDataElement = `<script type="application/json" id="pidl-walkthrough"></script>`
	viewerTitle       = `<title>PIDL Walkthrough</title>`
)

// ViewerRenderer writes a self-contained HTML walkthrough viewer with the
// protocol's walkthrough embedded, so the file opens in a browser with no
// server.
type ViewerRenderer struct {
	// Walkthrough produces the embedded data, including analyzer findings.
	Walkthrough *WalkthroughJSONRenderer
}

// NewViewer creates a ViewerRenderer with the default walkthrough options.
func NewViewer() *ViewerRenderer {
	return &ViewerRenderer{Walkthrough: NewWalkthroughJSON()}
}

// Format returns the output format identifier.
func (r *ViewerRenderer) Format() Format {
	return FormatViewer
}

// Render writes the viewer HTML to w.
func (r *ViewerRenderer) Render(w io.Writer, p *pidl.Protocol) error {
	s, err := r.RenderString(p)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, s)
	return err
}

// RenderString returns the viewer HTML with the walkthrough embedded.
func (r *ViewerRenderer) RenderString(p *pidl.Protocol) (string, error) {
	if !strings.Contains(viewerHTML, viewerDataElement) {
		return "", errors.New("embedded viewer has no data element; regenerate it with go generate ./render")
	}
	// json.Marshal escapes <, > and & as <, > and &, so the
	// data cannot close the script element it sits in.
	w, err := r.Walkthrough.Build(p)
	if err != nil {
		return "", err
	}
	data, err := json.Marshal(w)
	if err != nil {
		return "", err
	}
	page := strings.Replace(viewerHTML, viewerDataElement,
		`<script type="application/json" id="pidl-walkthrough">`+string(data)+`</script>`, 1)
	title := "<title>" + html.EscapeString(p.ProtocolMeta.Name) + " walkthrough</title>"
	return strings.Replace(page, viewerTitle, title, 1), nil
}
