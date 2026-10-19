package svg

import (
	"reflect"
	"testing"

	"github.com/grokify/pidl"
)

// TestResolveBoundariesDeterministic guards against map iteration order
// leaking into boundary assignment and ordering.
func TestResolveBoundariesDeterministic(t *testing.T) {
	p := &pidl.Protocol{
		Entities: []pidl.Entity{
			{ID: "browser"}, {ID: "gateway"}, {ID: "api"}, {ID: "db"}, {ID: "idp"}, {ID: "audit"},
		},
		Metadata: &pidl.ProtocolMetadata{
			Networks: map[string]*pidl.NetworkConfig{
				"dmz":      {Entities: []string{"gateway", "api"}},
				"internal": {Entities: []string{"api", "db"}}, // api is listed twice
				"partner":  {Entities: []string{"idp"}},
				"logging":  {Entities: []string{"audit"}},
			},
			NetworkLayout: &pidl.NetworkLayoutConfig{Order: []string{"internal"}},
		},
	}
	overrides := map[string][]string{"edge": {"browser"}, "zone": {"browser"}}

	first := ResolveBoundaries(p, overrides)
	for i := 0; i < 50; i++ {
		if got := ResolveBoundaries(p, overrides); !reflect.DeepEqual(got, first) {
			t.Fatalf("run %d differs:\n got %+v\nwant %+v", i, got, first)
		}
	}

	// Later networks in sorted ID order win a shared entity: "internal" > "dmz".
	byEntity := map[string]string{}
	var order []string
	for _, b := range first {
		order = append(order, b.ID)
		for _, e := range b.Entities {
			byEntity[e] = b.ID
		}
	}
	if byEntity["api"] != "internal" || byEntity["browser"] != "zone" {
		t.Errorf("assignments = %v; want api in internal, browser in zone", byEntity)
	}
	if order[0] != "internal" {
		t.Errorf("order = %v; want the explicit order first", order)
	}
	if got := p.Metadata.NetworkLayout.Order; !reflect.DeepEqual(got, []string{"internal"}) {
		t.Errorf("NetworkLayout.Order was modified: %v", got)
	}
}
