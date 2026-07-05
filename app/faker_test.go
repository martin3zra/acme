package app

import (
	"hash/fnv"
	"testing"

	"github.com/martin3zra/faker"
)

// newFaker returns a deterministic fake-data generator seeded from the test
// name. Each test gets its own reproducible stream: a failure replays with the
// same values, while different tests don't all draw identical data. Sequential
// calls within a test advance the PRNG, so repeated draws (names, companies)
// differ naturally.
//
// Note: fields the database requires to be unique (e.g. customer email) are not
// left to the generator — the builders suffix them with uniq() so a repeated
// draw can never trip a UNIQUE constraint. faker supplies realism where it is
// free; uniq() guarantees uniqueness where the schema demands it.
func newFaker(t *testing.T) *faker.Generator {
	t.Helper()
	h := fnv.New64a()
	_, _ = h.Write([]byte(t.Name()))
	return faker.New(faker.WithSeed(int64(h.Sum64())))
}

// fakeGen is the slice of the faker generator the builders depend on. Keeping it
// an interface lets the builder files stay decoupled from the concrete type;
// *faker.Generator satisfies it.
type fakeGen interface {
	Name() string
	Company() string
	Phone() string
	Email() string
	Price() float64
	Sentence() string
	IntRange(min, max int) int
}
