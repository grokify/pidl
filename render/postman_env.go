package render

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"

	"github.com/grokify/pidl"
)

// PostmanTarget says where a collection's roles run: a base URL per role,
// optional per-scenario overrides, and an optional signing helper.
type PostmanTarget struct {
	// Signer is the signing helper's base URL, such as
	// "http://127.0.0.1:8099".
	Signer string `json:"signer,omitempty"`
	// Roles maps a role ID to its base URL.
	Roles map[string]string `json:"roles"`
	// Scenarios maps a scenario ID to role base URLs that apply to that
	// scenario's folder only.
	Scenarios map[string]map[string]string `json:"scenarios,omitempty"`
	// Variables are extra environment variables, such as "mock_url".
	Variables map[string]string `json:"variables,omitempty"`
}

// ReadPostmanTargets reads a targets file: a JSON object mapping target
// names (such as "sandbox" or "mock") to targets.
func ReadPostmanTargets(path string) (map[string]PostmanTarget, error) {
	data, err := os.ReadFile(path) // #nosec G304 -- the path is chosen by the caller
	if err != nil {
		return nil, err
	}
	var targets map[string]PostmanTarget
	if err := json.Unmarshal(data, &targets); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return targets, nil
}

// PostmanEnvironment is a Postman environment file.
type PostmanEnvironment struct {
	ID     string                    `json:"id"`
	Name   string                    `json:"name"`
	Values []PostmanEnvironmentValue `json:"values"`
	Scope  string                    `json:"_postman_variable_scope"`
}

// PostmanEnvironmentValue is one environment variable.
type PostmanEnvironmentValue struct {
	Key     string `json:"key"`
	Value   string `json:"value"`
	Type    string `json:"type"`
	Enabled bool   `json:"enabled"`
}

// NewPostmanEnvironment builds the environment for one target, checking
// that every role and scenario it names exists in the protocol.
func NewPostmanEnvironment(p *pidl.Protocol, name string, t PostmanTarget) (*PostmanEnvironment, error) {
	roles := map[string]bool{}
	for _, e := range p.Entities {
		roles[e.ID] = true
	}
	env := &PostmanEnvironment{
		ID:    nameUUID("pidl:" + p.ProtocolMeta.ID + ":env:" + name),
		Name:  p.ProtocolMeta.Name + " (" + name + ")",
		Scope: "environment",
	}
	add := func(k, v string) {
		env.Values = append(env.Values, PostmanEnvironmentValue{Key: k, Value: v, Type: "default", Enabled: true})
	}
	for _, role := range sortedKeys(t.Roles) {
		if !roles[role] {
			return nil, fmt.Errorf("target %s: unknown role %q", name, role)
		}
		add(role+"_url", t.Roles[role])
	}
	for _, sc := range sortedKeys(t.Scenarios) {
		if p.ScenarioByID(sc) == nil {
			return nil, fmt.Errorf("target %s: unknown scenario %q", name, sc)
		}
		for _, role := range sortedKeys(t.Scenarios[sc]) {
			if !roles[role] {
				return nil, fmt.Errorf("target %s: scenario %s: unknown role %q", name, sc, role)
			}
			add(sc+"."+role+"_url", t.Scenarios[sc][role])
		}
	}
	if t.Signer != "" {
		add("signer_url", t.Signer)
	}
	for _, k := range sortedKeys(t.Variables) {
		add(k, t.Variables[k])
	}
	return env, nil
}

// JSON returns the environment as indented JSON.
func (e *PostmanEnvironment) JSON() (string, error) {
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(e); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func sortedKeys[V any](m map[string]V) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// PostmanEnvironmentRenderer writes the environment for one target. With
// no target, roles point at the origins of the protocol's example URLs.
type PostmanEnvironmentRenderer struct {
	Name   string
	Target *PostmanTarget
}

// Format returns the output format identifier.
func (r *PostmanEnvironmentRenderer) Format() Format { return FormatPostmanEnvironment }

// Render writes the environment to w.
func (r *PostmanEnvironmentRenderer) Render(w io.Writer, p *pidl.Protocol) error {
	s, err := r.RenderString(p)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, s)
	return err
}

// RenderString returns the environment as indented JSON.
func (r *PostmanEnvironmentRenderer) RenderString(p *pidl.Protocol) (string, error) {
	name, t := r.Name, r.Target
	if t == nil {
		name = "examples"
		t = &PostmanTarget{Roles: roleOrigins(p.Walkthrough(nil))}
	}
	env, err := NewPostmanEnvironment(p, name, *t)
	if err != nil {
		return "", err
	}
	return env.JSON()
}
