package neat

// innovations assigns globally unique, reproducible innovation numbers to new
// connections and ids to new nodes. Within a generation, identical structural
// mutations (the same from→to pair) share an innovation number, as in the
// original NEAT; the per-generation cache is reset by reset.
type innovations struct {
	nextInno int
	nextNode int
	cache    map[[2]int]int
}

// newInnovations starts numbering after the given node id (so initial nodes are
// not reallocated).
func newInnovations(firstNodeID int) *innovations {
	return &innovations{nextNode: firstNodeID, cache: map[[2]int]int{}}
}

// conn returns the innovation number for a from→to connection, reusing the
// number already assigned to that pair this generation.
func (in *innovations) conn(from, to int) int {
	key := [2]int{from, to}
	if n, ok := in.cache[key]; ok {
		return n
	}
	n := in.nextInno
	in.nextInno++
	in.cache[key] = n
	return n
}

// newNode allocates a fresh hidden node id.
func (in *innovations) newNode() int {
	id := in.nextNode
	in.nextNode++
	return id
}

// reset clears the per-generation connection cache while keeping the global
// counters monotonic.
func (in *innovations) reset() { in.cache = map[[2]int]int{} }
