package sim

import (
	"math/rand/v2"
	"testing"

	"github.com/danielriddell21/galapagos/internal/core"
)

type statusEnv struct {
	*fakeSingleEnv
	status string
}

func (e *statusEnv) Status() string { return e.status }

func TestBodyStatus(t *testing.T) {
	// A body that exposes Status reports it through the matching index.
	env := AsMulti(func() core.Environment {
		return &statusEnv{fakeSingleEnv: &fakeSingleEnv{life: 10}, status: "pipes 3"}
	}).(*multiAdapter)
	env.ResetAll(2, rand.New(rand.NewPCG(1, 2)))

	if s, ok := env.BodyStatus(1); !ok || s != "pipes 3" {
		t.Fatalf("BodyStatus(1) = %q, %v; want \"pipes 3\", true", s, ok)
	}

	// Out-of-range indices report no status.
	if _, ok := env.BodyStatus(-1); ok {
		t.Fatal("BodyStatus(-1) should report false")
	}
	if _, ok := env.BodyStatus(2); ok {
		t.Fatal("BodyStatus(2) out of range should report false")
	}

	// A body whose environment has no Status method reports false.
	plain := AsMulti(func() core.Environment { return &fakeSingleEnv{life: 10} }).(*multiAdapter)
	plain.ResetAll(1, rand.New(rand.NewPCG(1, 2)))
	if _, ok := plain.BodyStatus(0); ok {
		t.Fatal("a body without Status should report false")
	}
}
