package maze

import (
	"math/rand/v2"
	"testing"

	"github.com/danielriddell21/galapagos/internal/render"
)

func newRNG() *rand.Rand { return rand.New(rand.NewPCG(1, 2)) }

type act int

func (a act) Vector() []float64 { return []float64{float64(a)} }

func TestMazeIsFullyConnected(t *testing.T) {
	// A perfect maze reaches every cell from the start; flood fill must visit
	// all of them.
	e := New(DefaultConfig())
	e.Reset(newRNG())
	w, h := e.cfg.Width, e.cfg.Height
	seen := make([]bool, w*h)
	stack := []int{0}
	seen[0] = true
	count := 0
	for len(stack) > 0 {
		c := stack[len(stack)-1]
		stack = stack[:len(stack)-1]
		count++
		cx, cy := c%w, c/w
		for d := range 4 {
			if e.open[c][d] {
				nc := e.cell(cx+dx[d], cy+dy[d])
				if !seen[nc] {
					seen[nc] = true
					stack = append(stack, nc)
				}
			}
		}
	}
	if count != w*h {
		t.Fatalf("reached %d cells, want %d (maze not fully connected)", count, w*h)
	}
}

func TestStatusAndRenderReflectGoal(t *testing.T) {
	e := New(Config{Width: 6, Height: 6, MaxSteps: 10000})
	e.Reset(newRNG())

	// Before reaching the exit the agent is exploring; Render draws the goal and
	// agent in their base colors.
	if got := e.Status(); got != "exploring" {
		t.Fatalf("fresh maze status = %q, want %q", got, "exploring")
	}
	e.Render(render.NewNop())

	if !exploreToGoal(e) {
		t.Fatal("goal not reachable")
	}

	// At the exit the status flips and Render takes the lit-up (solved) path.
	if got := e.Status(); got != "solved" {
		t.Fatalf("status at goal = %q, want %q", got, "solved")
	}
	e.Render(render.NewNop())
}

func TestMazeReachableByGreedyDFS(t *testing.T) {
	// Walking the maze with a wall-following depth-first policy reaches the goal,
	// confirming the goal is solvable and Step honors walls.
	e := New(Config{Width: 6, Height: 6, MaxSteps: 10000})
	e.Reset(newRNG())
	if exploreToGoal(e) == false {
		t.Fatal("goal not reachable")
	}
	if !e.Solved() {
		t.Fatal("Solved() should report true at the goal")
	}
}

// exploreToGoal does an iterative DFS over cells, moving the agent one step at a
// time, and returns whether the goal was reached.
func exploreToGoal(e *Env) bool {
	w, h := e.cfg.Width, e.cfg.Height
	visited := make([]bool, w*h)
	var dfs func() bool
	dfs = func() bool {
		if e.Solved() {
			return true
		}
		visited[e.cell(e.x, e.y)] = true
		for d := range 4 {
			cx, cy := e.x, e.y
			if !e.open[e.cell(cx, cy)][d] {
				continue
			}
			if visited[e.cell(cx+dx[d], cy+dy[d])] {
				continue
			}
			e.Step(act(d))
			if dfs() {
				return true
			}
			e.Step(act((d + 2) % 4)) // backtrack
		}
		return false
	}
	return dfs()
}
