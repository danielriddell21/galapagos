package neat

type innovations struct {
	nextInno int
	nextNode int
	cache    map[[2]int]int
}

func newInnovations(firstNodeID int) *innovations {
	return &innovations{nextNode: firstNodeID, cache: map[[2]int]int{}}
}

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

func (in *innovations) newNode() int {
	id := in.nextNode
	in.nextNode++
	return id
}

func (in *innovations) reset() { in.cache = map[[2]int]int{} }
