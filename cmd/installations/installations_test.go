package installations

import (
	"testing"

	"github.com/giantswarm/backstage-catalog-importer/pkg/input/installations"
	bscatalog "github.com/giantswarm/backstage-catalog-importer/pkg/output/bscatalog/v1alpha1"
)

func TestToCustomerGroupEntities(t *testing.T) {
	ins := []*installations.Installation{
		{Codename: "gazelle", Customer: "giantswarm"},
		{Codename: "golem", Customer: "giantswarm"},
		{Codename: "alpha", Customer: "adidas"},
		{Codename: "nameless", Customer: ""},
	}

	entities := toCustomerGroupEntities(ins)

	if len(entities) != 2 {
		t.Fatalf("expected 2 customer groups, got %d", len(entities))
	}

	wantNames := []string{"adidas", "giantswarm"}
	for i, e := range entities {
		if e.Kind != bscatalog.EntityKindGroup {
			t.Errorf("entity %d: expected kind %q, got %q", i, bscatalog.EntityKindGroup, e.Kind)
		}
		if e.Metadata.Name != wantNames[i] {
			t.Errorf("entity %d: expected name %q, got %q", i, wantNames[i], e.Metadata.Name)
		}
		spec, ok := e.Spec.(bscatalog.GroupSpec)
		if !ok {
			t.Fatalf("entity %d: unexpected spec type %T", i, e.Spec)
		}
		if spec.Type != "customer" {
			t.Errorf("entity %d: expected type customer, got %q", i, spec.Type)
		}
	}
}
