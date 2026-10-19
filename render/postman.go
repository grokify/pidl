package render

import (
	"bytes"
	"crypto/sha1" // #nosec G505 -- name-based UUIDs (RFC 9562 version 5) use SHA-1 by definition
	"encoding/json"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/grokify/pidl"
)

// PostmanSchema is the Postman Collection format the renderer writes.
const PostmanSchema = "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"

// Signer profiles name how a signing helper receives a request.
const (
	// SignerNone sends requests directly, with their example headers.
	SignerNone = ""
	// SignerAAuthSandbox sends requests through the AAuth sandbox's
	// signing helper: the target URL follows the helper's address, and
	// X-Sandbox-Present and X-Sandbox-Token select the presented token.
	SignerAAuthSandbox = "aauth-sandbox"
)

// PostmanOptions configure the collection.
type PostmanOptions struct {
	// Client is the role whose requests become Postman requests. Empty
	// selects the first entity of type agent or client.
	Client string
	// Signer is a signer profile; SignerNone sends requests directly.
	Signer string
	// Mock saves example responses as Postman examples, adds the
	// x-mock-response-name header, and checks only status codes.
	Mock bool
	// PollLimit is how many times a deferred request is polled before the
	// run fails. Zero means 30.
	PollLimit int
}

// PostmanRenderer writes a Postman collection (format v2.1) with a folder
// per scenario and a request per step the client role sends, chaining
// values from responses into later requests.
type PostmanRenderer struct {
	Options PostmanOptions
}

// NewPostman creates a PostmanRenderer.
func NewPostman(opts PostmanOptions) *PostmanRenderer {
	return &PostmanRenderer{Options: opts}
}

// Format returns the output format identifier.
func (r *PostmanRenderer) Format() Format { return FormatPostman }

// Render writes the collection to w.
func (r *PostmanRenderer) Render(w io.Writer, p *pidl.Protocol) error {
	s, err := r.RenderString(p)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, s)
	return err
}

// RenderString returns the collection as indented JSON.
func (r *PostmanRenderer) RenderString(p *pidl.Protocol) (string, error) {
	c, err := r.Collection(p)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(c); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// Postman Collection v2.1 types, limited to what the renderer writes.

// PostmanCollection is a collection.
type PostmanCollection struct {
	Info     PostmanInfo       `json:"info"`
	Item     []PostmanItem     `json:"item"`
	Event    []PostmanEvent    `json:"event,omitempty"`
	Variable []PostmanVariable `json:"variable,omitempty"`
}

// PostmanInfo describes a collection.
type PostmanInfo struct {
	PostmanID   string `json:"_postman_id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Schema      string `json:"schema"`
}

// PostmanItem is a folder (Item set) or a request (Request set).
type PostmanItem struct {
	ID          string            `json:"id,omitempty"`
	Name        string            `json:"name"`
	Description string            `json:"description,omitempty"`
	Item        []PostmanItem     `json:"item,omitempty"`
	Request     *PostmanRequest   `json:"request,omitempty"`
	Response    []PostmanResponse `json:"response,omitempty"`
	Event       []PostmanEvent    `json:"event,omitempty"`
}

// PostmanRequest is a request.
type PostmanRequest struct {
	Method string          `json:"method"`
	Header []PostmanHeader `json:"header"`
	Body   *PostmanBody    `json:"body,omitempty"`
	URL    string          `json:"url"`
}

// PostmanHeader is one header.
type PostmanHeader struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// PostmanBody is a raw request body.
type PostmanBody struct {
	Mode    string             `json:"mode"`
	Raw     string             `json:"raw"`
	Options *PostmanRawOptions `json:"options,omitempty"`
}

// PostmanRawOptions names a raw body's language.
type PostmanRawOptions struct {
	Raw struct {
		Language string `json:"language"`
	} `json:"raw"`
}

// PostmanResponse is a saved example response.
type PostmanResponse struct {
	Name            string          `json:"name"`
	OriginalRequest *PostmanRequest `json:"originalRequest,omitempty"`
	Status          string          `json:"status"`
	Code            int             `json:"code"`
	Header          []PostmanHeader `json:"header"`
	Body            string          `json:"body"`
	PreviewLanguage string          `json:"_postman_previewlanguage,omitempty"`
}

// PostmanEvent is a script run before a request or after its response.
type PostmanEvent struct {
	Listen string        `json:"listen"`
	Script PostmanScript `json:"script"`
}

// PostmanScript is JavaScript, one line per element.
type PostmanScript struct {
	Type string   `json:"type"`
	Exec []string `json:"exec"`
}

// PostmanVariable is a collection variable.
type PostmanVariable struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// capture is a value a response delivers that later requests reuse.
type capture struct {
	name    string
	value   string
	header  string // header name, or "" for a body member
	pointer string // JSON pointer into the body
}

// postmanStep is a client request with its paired response.
type postmanStep struct {
	step     pidl.WalkthroughStep
	response *pidl.WalkthroughStep
	// poll is the index of the request that polls this one's Location,
	// or -1; final is the response that eventually answers it.
	poll  int
	final *pidl.WalkthroughStep
}

// Collection builds the collection.
func (r *PostmanRenderer) Collection(p *pidl.Protocol) (*PostmanCollection, error) {
	w := p.Walkthrough(nil)
	client := r.Options.Client
	if client == "" {
		client = defaultClient(p)
	}
	if client == "" {
		return nil, fmt.Errorf("no client role: set one, or add an entity of type agent or client")
	}

	roles := roleOrigins(w)
	c := &PostmanCollection{
		Info: PostmanInfo{
			PostmanID:   nameUUID("pidl:" + w.ID),
			Name:        w.Name,
			Description: w.Description,
			Schema:      PostmanSchema,
		},
		Item: []PostmanItem{},
	}

	byGroup := map[string][]PostmanItem{}
	var ungrouped []PostmanItem
	for _, sc := range w.Scenarios {
		if sc.Fragment {
			continue
		}
		folder, err := r.scenarioFolder(w, sc, client, roles)
		if err != nil {
			return nil, fmt.Errorf("scenario %s: %w", sc.ID, err)
		}
		if sc.Group != "" {
			byGroup[sc.Group] = append(byGroup[sc.Group], folder)
		} else {
			ungrouped = append(ungrouped, folder)
		}
	}
	for _, g := range w.Groups {
		if items := byGroup[g.ID]; len(items) > 0 {
			c.Item = append(c.Item, PostmanItem{Name: g.Name, Description: g.Description, Item: items})
		}
	}
	c.Item = append(c.Item, ungrouped...)

	vars := make([]string, 0, len(roles))
	for role := range roles {
		vars = append(vars, role)
	}
	sort.Strings(vars)
	for _, role := range vars {
		c.Variable = append(c.Variable, PostmanVariable{Key: role + "_url", Value: roles[role]})
	}
	if r.Options.Signer != SignerNone {
		c.Variable = append(c.Variable, PostmanVariable{Key: "signer_url", Value: "http://127.0.0.1:8099"})
	}
	return c, nil
}

// defaultClient returns the first entity of type agent or client.
func defaultClient(p *pidl.Protocol) string {
	for _, e := range p.Entities {
		if e.Type == pidl.EntityTypeAgent || e.Type == pidl.EntityTypeClient {
			return e.ID
		}
	}
	return ""
}

// roleOrigins maps each role to the origin its example requests are sent
// to, from the first example request each role receives.
func roleOrigins(w pidl.Walkthrough) map[string]string {
	out := map[string]string{}
	for _, sc := range w.Scenarios {
		for _, st := range sc.Steps {
			if st.Example == nil || st.Example.Request == nil {
				continue
			}
			if _, ok := out[st.To]; ok {
				continue
			}
			if u, err := url.Parse(st.Example.Request.URL); err == nil && u.Host != "" {
				out[st.To] = u.Scheme + "://" + u.Host
			}
		}
	}
	return out
}

// pairResponse returns the index of the response to request i: the first
// later response from the request's receiver to its sender, before another
// request between the same two roles. It returns -1 when there is none.
func pairResponse(steps []pidl.WalkthroughStep, i int) int {
	req := steps[i]
	for j := i + 1; j < len(steps); j++ {
		s := steps[j]
		if !s.Response && s.From == req.From && s.To == req.To {
			return -1
		}
		if s.Response && s.From == req.To && s.To == req.From {
			return j
		}
	}
	return -1
}

func (r *PostmanRenderer) scenarioFolder(w pidl.Walkthrough, sc pidl.WalkthroughScenario, client string, roles map[string]string) (PostmanItem, error) {
	var steps []postmanStep
	var others []string
	for i, st := range sc.Steps {
		if st.Response {
			continue
		}
		if st.From != client || st.Example == nil || st.Example.Request == nil {
			others = append(others, fmt.Sprintf("%d. %s → %s: %s", st.Number, roleName(w, st.From), roleName(w, st.To), st.Label))
			continue
		}
		ps := postmanStep{step: st, poll: -1}
		if j := pairResponse(sc.Steps, i); j >= 0 && sc.Steps[j].Example != nil && sc.Steps[j].Example.Response != nil {
			ps.response = &sc.Steps[j]
		}
		steps = append(steps, ps)
	}

	// Link deferred requests to the request that polls their Location.
	for i := range steps {
		loc := headerValue(steps[i].response, "Location")
		if loc == "" || steps[i].response.Example.Response.Status != 202 {
			continue
		}
		for k := i + 1; k < len(steps); k++ {
			if steps[k].step.Example.Request.URL == loc {
				steps[i].poll = k
				steps[i].final = steps[k].response
				break
			}
		}
	}

	caps := collectCaptures(steps)
	items := make([]PostmanItem, 0, len(steps))
	for i := range steps {
		item, err := r.requestItem(sc, steps, i, caps, roles)
		if err != nil {
			return PostmanItem{}, err
		}
		items = append(items, item)
	}

	desc := sc.Description
	if text, ok := sc.Narration.Text(w.Language, w.Language); ok {
		desc = strings.TrimSpace(desc + "\n\n" + text)
	}
	if len(others) > 0 {
		desc += "\n\nSteps between other roles, not sent by this collection:\n\n- " + strings.Join(others, "\n- ")
	}
	return PostmanItem{
		Name:        sc.Name,
		Description: strings.TrimSpace(desc),
		Item:        items,
		Event:       []PostmanEvent{folderOverrides(sc, roles)},
	}, nil
}

// folderOverrides lets an environment point a role at another server for
// one scenario: an environment variable "<scenario>.<role>_url" overrides
// "<role>_url" for the requests in that scenario's folder.
func folderOverrides(sc pidl.WalkthroughScenario, roles map[string]string) PostmanEvent {
	names := make([]string, 0, len(roles))
	for role := range roles {
		names = append(names, role)
	}
	sort.Strings(names)
	lines := []string{"// An environment can point a role at another server for this scenario."}
	for _, role := range names {
		key := sc.ID + "." + role + "_url"
		lines = append(lines, fmt.Sprintf("if (pm.environment.has(%q)) { pm.variables.set(%q, pm.environment.get(%q)); }", key, role+"_url", key))
	}
	return PostmanEvent{Listen: "prerequest", Script: PostmanScript{Type: "text/javascript", Exec: lines}}
}

func roleName(w pidl.Walkthrough, id string) string {
	for _, r := range w.Roles {
		if r.ID == id {
			return r.Name
		}
	}
	return id
}

func headerValue(st *pidl.WalkthroughStep, name string) string {
	if st == nil || st.Example == nil || st.Example.Response == nil {
		return ""
	}
	for _, h := range st.Example.Response.Headers {
		if strings.EqualFold(h.Name, name) {
			return h.Value
		}
	}
	return ""
}

// collectCaptures finds values in responses that later requests reuse:
// string body members and Location headers. Names come from the member or
// header name, made unique within the scenario.
func collectCaptures(steps []postmanStep) map[int][]capture {
	out := map[int][]capture{}
	used := map[string]bool{}
	taken := map[string]string{} // name -> value
	unique := func(name, value string) string {
		base, n := name, 2
		for {
			if v, ok := taken[name]; !ok || v == value {
				taken[name] = value
				return name
			}
			name = fmt.Sprintf("%s_%d", base, n)
			n++
		}
	}
	sources := func(st *pidl.WalkthroughStep) []capture {
		if st == nil {
			return nil
		}
		var cs []capture
		for _, h := range st.Example.Response.Headers {
			if strings.EqualFold(h.Name, "Location") || (len(h.Value) >= 16 && !volatileHeader(h.Name)) {
				cs = append(cs, capture{name: strings.ToLower(h.Name), value: h.Value, header: h.Name})
			}
		}
		var body any
		if b := st.Example.Response.Body; len(b) > 0 && json.Unmarshal(b, &body) == nil {
			walkStrings(body, "", func(ptr, s string) {
				if len(s) >= 16 {
					cs = append(cs, capture{name: lastSegment(ptr), value: s, pointer: ptr})
				}
			})
		}
		return cs
	}
	for i, ps := range steps {
		for _, src := range append(sources(ps.response), sources(ps.final)...) {
			if usedLater(steps, i, src.value) || isPresented(steps, i, src.name) {
				key := src.name + "\x00" + src.value
				if used[key] {
					continue
				}
				used[key] = true
				src.name = unique(sanitizeVar(src.name), src.value)
				out[i] = append(out[i], src)
			}
		}
	}
	// A deferred request's final values are captured by its poll too, since
	// the final response arrives there when the request is deferred.
	for i, ps := range steps {
		if ps.poll < 0 {
			continue
		}
		for _, c := range out[i] {
			if c.source(ps.final) {
				out[ps.poll] = append(out[ps.poll], c)
			}
		}
	}
	return out
}

// replaceOrigin replaces origin where it appears as a whole origin: not
// followed by more host characters, so "https://a.example" does not match
// inside "https://a.example.org".
func replaceOrigin(s, origin, repl string) string {
	if origin == "" {
		return s
	}
	var b strings.Builder
	for {
		i := strings.Index(s, origin)
		if i < 0 {
			b.WriteString(s)
			return b.String()
		}
		end := i + len(origin)
		if end < len(s) && (isHostChar(s[end]) || s[end] == ':') {
			b.WriteString(s[:end])
			s = s[end:]
			continue
		}
		b.WriteString(s[:i])
		b.WriteString(repl)
		s = s[end:]
	}
}

func isHostChar(c byte) bool {
	return c == '.' || c == '-' || (c >= '0' && c <= '9') || (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// volatileHeader reports headers whose values describe the response
// itself rather than carry protocol values.
func volatileHeader(name string) bool {
	switch strings.ToLower(name) {
	case "date", "content-type", "content-length", "cache-control", "retry-after", "etag", "last-modified", "vary", "aauth-requirement", "www-authenticate":
		return true
	}
	return isSignatureHeader(name)
}

// usedLater reports whether a request after step i carries value in its
// URL, a header or a body string.
func usedLater(steps []postmanStep, i int, value string) bool {
	for _, ps := range steps[i+1:] {
		req := ps.step.Example.Request
		if strings.Contains(req.URL, value) {
			return true
		}
		for _, h := range req.Headers {
			if strings.Contains(h.Value, value) {
				return true
			}
		}
		if bytes.Contains(req.Body, []byte(value)) {
			return true
		}
	}
	return false
}

// isPresented reports whether a later request presents a token named name
// in its Signature-Key header.
func isPresented(steps []postmanStep, i int, name string) bool {
	for _, ps := range steps[i+1:] {
		if presentedToken(ps.step) == name {
			return true
		}
	}
	return false
}

// presentedToken returns the token definition ID of the token a request
// presents in Signature-Key, from the example's token labels.
func presentedToken(st pidl.WalkthroughStep) string {
	for _, t := range st.Example.Tokens {
		if strings.HasPrefix(t.Label, "Signature-Key") {
			return t.Token
		}
	}
	return ""
}

// signatureScheme returns the Signature-Key scheme of a request, such as
// "jwt", "hwk" or "jkt-jwt".
var schemePattern = regexp.MustCompile(`^\s*[A-Za-z0-9_-]+=([A-Za-z0-9_-]+)`)

func signatureScheme(req *pidl.ExampleRequest) string {
	for _, h := range req.Headers {
		if strings.EqualFold(h.Name, "Signature-Key") {
			if m := schemePattern.FindStringSubmatch(h.Value); m != nil {
				return m[1]
			}
		}
	}
	return ""
}

// tokenType returns the typ of a decoded example token.
func tokenType(st pidl.WalkthroughStep, id string) string {
	for _, t := range st.Example.Tokens {
		if t.Token != id {
			continue
		}
		var h struct {
			Typ string `json:"typ"`
		}
		if json.Unmarshal(t.Header, &h) == nil {
			return h.Typ
		}
	}
	return ""
}

// sandboxPresent maps a request's Signature-Key to the AAuth sandbox
// helper's X-Sandbox-Present value.
func sandboxPresent(st pidl.WalkthroughStep) string {
	switch scheme := signatureScheme(st.Example.Request); scheme {
	case "hwk", "jkt-jwt":
		return scheme
	case "":
		return "none"
	}
	switch tokenType(st, presentedToken(st)) {
	case "aa-person+jwt":
		return "person"
	case "aa-auth+jwt":
		return "auth"
	default:
		return "agent"
	}
}

func (r *PostmanRenderer) requestItem(sc pidl.WalkthroughScenario, steps []postmanStep, i int, caps map[int][]capture, roles map[string]string) (PostmanItem, error) {
	ps := steps[i]
	st := ps.step
	ex := st.Example.Request

	// Values captured before this step, by value and by name.
	var earlier []capture
	for k := 0; k < i; k++ {
		earlier = append(earlier, caps[k]...)
	}
	sub := func(s string) string {
		for _, c := range earlier {
			s = strings.ReplaceAll(s, c.value, "{{"+c.name+"}}")
		}
		return s
	}
	// subRoles also replaces role origins that appear in values, such as a
	// resource named in a body, so they follow the environment.
	origins := make([]string, 0, len(roles))
	for role := range roles {
		origins = append(origins, role)
	}
	sort.Slice(origins, func(a, b int) bool { return len(roles[origins[a]]) > len(roles[origins[b]]) })
	subRoles := func(s string) string {
		s = sub(s)
		for _, role := range origins {
			s = replaceOrigin(s, roles[role], "{{"+role+"_url}}")
		}
		return s
	}

	target := sub(ex.URL)
	if !strings.HasPrefix(target, "{{") {
		if u, err := url.Parse(ex.URL); err == nil && u.Host != "" {
			target = "{{" + st.To + "_url}}" + strings.TrimPrefix(ex.URL, u.Scheme+"://"+u.Host)
		}
	}

	req := &PostmanRequest{Method: ex.Method, Header: []PostmanHeader{}, URL: target}
	signed := r.Options.Signer == SignerAAuthSandbox
	for _, h := range ex.Headers {
		if signed && isSignatureHeader(h.Name) {
			continue
		}
		req.Header = append(req.Header, PostmanHeader{Key: h.Name, Value: subRoles(h.Value)})
	}
	if signed {
		req.URL = "{{signer_url}}/" + target
		present := sandboxPresent(st)
		req.Header = append(req.Header, PostmanHeader{Key: "X-Sandbox-Present", Value: present})
		if tok := presentedToken(st); tok != "" && present != "jkt-jwt" && present != "hwk" && hasCapture(earlier, tok) {
			req.Header = append(req.Header, PostmanHeader{Key: "X-Sandbox-Token", Value: "{{" + tok + "}}"})
		}
	}
	if r.Options.Mock && ps.response != nil {
		req.Header = append(req.Header, PostmanHeader{Key: "x-mock-response-name", Value: exampleName(st)})
	}
	if len(ex.Body) > 0 {
		wire, err := pidl.WireBody(ex.Body)
		if err != nil {
			return PostmanItem{}, fmt.Errorf("step %d body: %w", st.Number, err)
		}
		raw := subRoles(string(wire))
		if pretty, err := indentJSON(raw); err == nil {
			raw = pretty
		}
		body := &PostmanBody{Mode: "raw", Raw: raw}
		if json.Valid([]byte(raw)) || strings.HasPrefix(strings.TrimSpace(raw), "{") {
			body.Options = &PostmanRawOptions{}
			body.Options.Raw.Language = "json"
		}
		req.Body = body
	}

	item := PostmanItem{
		ID:          sc.ID + "/" + stepKey(st),
		Name:        fmt.Sprintf("%d. %s", st.Number, st.Label),
		Description: stepDescription(st),
		Request:     req,
	}
	if r.Options.Mock && ps.response != nil {
		item.Response = []PostmanResponse{savedResponse(st, *ps.response, req)}
	}
	if script := r.testScript(sc, steps, i, caps[i]); len(script) > 0 {
		item.Event = []PostmanEvent{{Listen: "test", Script: PostmanScript{Type: "text/javascript", Exec: script}}}
	}
	return item, nil
}

func hasCapture(cs []capture, name string) bool {
	for _, c := range cs {
		if c.name == name {
			return true
		}
	}
	return false
}

func isSignatureHeader(name string) bool {
	switch strings.ToLower(name) {
	case "signature", "signature-input", "signature-key", "content-digest":
		return true
	}
	return false
}

func stepKey(st pidl.WalkthroughStep) string {
	if st.ID != "" {
		return st.ID
	}
	return fmt.Sprint(st.Number)
}

func exampleName(st pidl.WalkthroughStep) string {
	return fmt.Sprintf("%d. %s", st.Number, st.Label)
}

func stepDescription(st pidl.WalkthroughStep) string {
	parts := []string{}
	if st.Description != "" {
		parts = append(parts, st.Description)
	}
	if st.Example.Note != "" {
		parts = append(parts, st.Example.Note)
	}
	if st.Example.Source != "" {
		parts = append(parts, "Source: "+st.Example.Source)
	}
	return strings.Join(parts, "\n\n")
}

func savedResponse(req, res pidl.WalkthroughStep, original *PostmanRequest) PostmanResponse {
	r := res.Example.Response
	out := PostmanResponse{
		Name:            exampleName(req),
		OriginalRequest: original,
		Status:          statusText(r.Status),
		Code:            r.Status,
		Header:          []PostmanHeader{},
	}
	for _, h := range r.Headers {
		out.Header = append(out.Header, PostmanHeader{Key: h.Name, Value: h.Value})
	}
	if len(r.Body) > 0 {
		if pretty, err := indentJSON(string(r.Body)); err == nil {
			out.Body = pretty
			out.PreviewLanguage = "json"
		} else {
			out.Body = string(r.Body)
		}
	}
	return out
}

// testScript checks the response and stores captured values. For a
// deferred request it accepts either the 202 (and polls) or the final
// response directly (and skips the poll); a poll repeats on 202.
func (r *PostmanRenderer) testScript(sc pidl.WalkthroughScenario, steps []postmanStep, i int, caps []capture) []string {
	ps := steps[i]
	var lines []string
	limit := r.Options.PollLimit
	if limit <= 0 {
		limit = 30
	}

	polledBy := -1
	for k := 0; k < i; k++ {
		if steps[k].poll == i {
			polledBy = k
		}
	}

	switch {
	case polledBy >= 0 && ps.response != nil:
		// A poll: repeat while pending.
		counter := "poll_" + sanitizeVar(sc.ID+"_"+stepKey(ps.step))
		lines = append(lines,
			"if (pm.response.code === 202) {",
			fmt.Sprintf("  const n = Number(pm.collectionVariables.get(%q) || 0) + 1;", counter),
			fmt.Sprintf("  pm.collectionVariables.set(%q, n);", counter),
			fmt.Sprintf("  pm.test('still pending (attempt ' + n + ')', () => pm.expect(n).to.be.below(%d));", limit),
			fmt.Sprintf("  if (n < %d) { setTimeout(() => {}, 1000); %s }", limit, nextRequest(sc.ID+"/"+stepKey(ps.step))),
			"} else {",
			fmt.Sprintf("  pm.collectionVariables.unset(%q);", counter),
		)
		for _, l := range r.checks(*ps.response) {
			lines = append(lines, "  "+l)
		}
		for _, c := range caps {
			lines = append(lines, "  "+c.script())
		}
		return append(lines, "}")
	case ps.poll >= 0 && ps.final != nil:
		// A deferred request: 202 leads to the poll; the final status skips it.
		lines = append(lines, fmt.Sprintf("if (pm.response.code === %d) {", ps.final.Example.Response.Status))
		for _, l := range r.checks(*ps.final) {
			lines = append(lines, "  "+l)
		}
		for _, c := range caps {
			if c.source(ps.final) {
				lines = append(lines, "  "+c.script())
			}
		}
		next := ""
		if ps.poll+1 < len(steps) {
			next = sc.ID + "/" + stepKey(steps[ps.poll+1].step)
		}
		if next != "" {
			lines = append(lines, "  "+nextRequest(next))
		} else {
			lines = append(lines, "  "+nextRequest(""))
		}
		lines = append(lines, "} else {")
		for _, l := range r.checks(*ps.response) {
			lines = append(lines, "  "+l)
		}
		for _, c := range caps {
			if c.source(ps.response) {
				lines = append(lines, "  "+c.script())
			}
		}
		return append(lines, "}")
	case ps.response != nil:
		lines = append(lines, r.checks(*ps.response)...)
	}
	for _, c := range caps {
		if ps.final == nil || c.source(ps.response) {
			lines = append(lines, c.script())
		}
	}
	return lines
}

// nextRequest returns script that runs the request with the given ID next,
// or ends the run for "". It uses pm.execution where the runtime has it and
// postman.setNextRequest otherwise.
func nextRequest(id string) string {
	arg := "null"
	if id != "" {
		q, err := json.Marshal(id)
		if err != nil {
			panic(err) // marshaling a string cannot fail
		}
		arg = string(q)
	}
	return fmt.Sprintf("if (pm.execution && pm.execution.setNextRequest) { pm.execution.setNextRequest(%s); } else { postman.setNextRequest(%s); }", arg, arg)
}

// source reports whether c is captured from st's response.
func (c capture) source(st *pidl.WalkthroughStep) bool {
	if st == nil || st.Example == nil || st.Example.Response == nil {
		return false
	}
	if c.header != "" {
		return headerValue(st, c.header) == c.value
	}
	return bytes.Contains(st.Example.Response.Body, []byte(c.value))
}

// script stores the captured value in a collection variable.
func (c capture) script() string {
	if c.header != "" {
		return fmt.Sprintf("pm.collectionVariables.set(%q, pm.response.headers.get(%q));", c.name, c.header)
	}
	return fmt.Sprintf("pm.collectionVariables.set(%q, pm.response.json()%s);", c.name, pointerExpr(c.pointer))
}

// checks asserts the status, the example's header names and body members,
// and the claim names of tokens the body carries. Mock collections check
// only the status.
func (r *PostmanRenderer) checks(res pidl.WalkthroughStep) []string {
	ex := res.Example.Response
	lines := []string{fmt.Sprintf("pm.test('status %d', () => pm.response.to.have.status(%d));", ex.Status, ex.Status)}
	if r.Options.Mock {
		return lines
	}
	for _, h := range ex.Headers {
		if isSignatureHeader(h.Name) || strings.EqualFold(h.Name, "Date") {
			continue
		}
		lines = append(lines, fmt.Sprintf("pm.test('header %s', () => pm.response.to.have.header(%q));", h.Name, h.Name))
	}
	var body map[string]any
	if len(ex.Body) > 0 && json.Unmarshal(ex.Body, &body) == nil {
		keys := make([]string, 0, len(body))
		for k := range body {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		lines = append(lines, "const pidlBody = pm.response.json();")
		for _, k := range keys {
			lines = append(lines, fmt.Sprintf("pm.test('body has %s', () => pm.expect(pidlBody).to.have.property(%q));", k, k))
		}
		for _, t := range res.Example.Tokens {
			if _, ok := body[t.Token]; !ok {
				continue
			}
			var claims map[string]any
			if json.Unmarshal(t.Payload, &claims) != nil {
				continue
			}
			names := make([]string, 0, len(claims))
			for n := range claims {
				names = append(names, n)
			}
			sort.Strings(names)
			quoted, err := json.Marshal(names)
			if err != nil {
				continue
			}
			lines = append(lines,
				fmt.Sprintf("pm.test('%s claims', () => {", t.Token),
				fmt.Sprintf("  const payload = JSON.parse(Buffer.from(String(pidlBody[%q]).split('.')[1].replace(/-/g, '+').replace(/_/g, '/'), 'base64').toString('utf8'));", t.Token),
				fmt.Sprintf("  pm.expect(payload).to.include.keys(%s);", quoted),
				"});",
			)
		}
	}
	return lines
}

// walkStrings calls fn for every string leaf with its JSON pointer.
func walkStrings(v any, ptr string, fn func(ptr, s string)) {
	switch t := v.(type) {
	case string:
		fn(ptr, t)
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			walkStrings(t[k], ptr+"/"+strings.NewReplacer("~", "~0", "/", "~1").Replace(k), fn)
		}
	case []any:
		for i, e := range t {
			walkStrings(e, fmt.Sprintf("%s/%d", ptr, i), fn)
		}
	}
}

// pointerExpr turns a JSON pointer into a JavaScript property path, such as
// "/a/0/b" into ["a"][0]["b"].
func pointerExpr(ptr string) string {
	var sb strings.Builder
	for _, seg := range strings.Split(strings.TrimPrefix(ptr, "/"), "/") {
		seg = strings.NewReplacer("~1", "/", "~0", "~").Replace(seg)
		if n, err := strconv.Atoi(seg); err == nil && n >= 0 {
			fmt.Fprintf(&sb, "[%d]", n)
			continue
		}
		q, err := json.Marshal(seg)
		if err != nil {
			panic(err) // marshaling a string cannot fail
		}
		fmt.Fprintf(&sb, "[%s]", q)
	}
	return sb.String()
}

func lastSegment(ptr string) string {
	if i := strings.LastIndex(ptr, "/"); i >= 0 {
		return ptr[i+1:]
	}
	return ptr
}

var nonVar = regexp.MustCompile(`[^A-Za-z0-9_]+`)

func sanitizeVar(s string) string {
	s = nonVar.ReplaceAllString(s, "_")
	if s == "" || (s[0] >= '0' && s[0] <= '9') {
		s = "v_" + s
	}
	return s
}

func indentJSON(s string) (string, error) {
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(s), "", "  "); err != nil {
		return "", err
	}
	return buf.String(), nil
}

// nameUUID returns a deterministic RFC 9562 version 5 UUID for name, so a
// regenerated collection keeps its identity.
func nameUUID(name string) string {
	h := sha1.Sum([]byte(name)) // #nosec G401 -- identifier, not security
	b := h[:16]
	b[6] = (b[6] & 0x0f) | 0x50
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

func statusText(code int) string {
	names := map[int]string{200: "OK", 201: "Created", 202: "Accepted", 204: "No Content", 400: "Bad Request", 401: "Unauthorized", 402: "Payment Required", 403: "Forbidden", 404: "Not Found"}
	if n, ok := names[code]; ok {
		return n
	}
	return fmt.Sprint(code)
}
