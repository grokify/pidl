package pidl

import (
	"strings"
	"testing"
	"testing/fstest"
)

func TestResolveFromFS(t *testing.T) {
	const base = `{"protocol":{"id":"base","name":"Base"},
		"entities":[{"id":"a","name":"A","type":"client"},{"id":"b","name":"B","type":"server"}],
		"flows":[{"from":"a","to":"b","action":"request","label":"Hello"}]}`
	fsys := fstest.MapFS{
		"lib/base.json":   {Data: []byte(base)},
		"derived.json":    {Data: []byte(`{"protocol":{"id":"d","name":"D"},"extends":{"path":"lib/base.json"}}`)},
		"escape.json":     {Data: []byte(`{"protocol":{"id":"e","name":"E"},"extends":{"path":"../outside.json"}}`)},
		"lib/loop-a.json": {Data: []byte(`{"protocol":{"id":"la","name":"LA"},"extends":{"path":"loop-b.json"}}`)},
		"lib/loop-b.json": {Data: []byte(`{"protocol":{"id":"lb","name":"LB"},"extends":{"path":"loop-a.json"}}`)},
		"imports.json":    {Data: []byte(`{"protocol":{"id":"i","name":"I"},"imports":[{"path":"lib/base.json","entities":["a"]}],"entities":[{"id":"c","name":"C","type":"server"}],"flows":[{"from":"a","to":"c","action":"request","label":"Hi"}]}`)},
	}
	resolve := func(name string) (*Protocol, error) {
		data, err := fsys.ReadFile(name)
		if err != nil {
			t.Fatal(err)
		}
		p, err := Parse(data)
		if err != nil {
			t.Fatal(err)
		}
		opts := DefaultResolveOptions()
		opts.FS = fsys
		return p.Resolve(opts)
	}

	p, err := resolve("derived.json")
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Entities) != 2 || len(p.Flows) != 1 || p.ProtocolMeta.ID != "d" {
		t.Fatalf("derived = %+v", p)
	}
	if p, err = resolve("imports.json"); err != nil || len(p.Entities) != 2 {
		t.Fatalf("imports = %v, %v", p, err)
	}
	if _, err := resolve("escape.json"); err == nil || !strings.Contains(err.Error(), "leaves the file system") {
		t.Errorf("escaping path: %v", err)
	}
	data, _ := fsys.ReadFile("lib/loop-a.json")
	loop, _ := Parse(data)
	opts := DefaultResolveOptions()
	opts.FS, opts.BasePath = fsys, "lib"
	if _, err := loop.Resolve(opts); err == nil || !strings.Contains(err.Error(), "circular extends") {
		t.Errorf("cycle: %v", err)
	}
}
