package bta

import (
	"go/types"
	"log"
)

// Point represents a subject of control flow.
type Point interface {
	Next() []Point
	Defs() types.Object
	Uses() []types.Object
	CouldBeTrue(d Division) bool
}

// Division describes the known-ness of a set of variables.
type Division map[types.Object]bool

// Graph stores the control flow structure under examination.
type Graph struct {
	point              Point
	defs               types.Object
	uses               []types.Object
	prev, next         []*Graph
	loopDeps, dataDeps []*Graph
}

// NewGraph creates a new graph given a starting Point.
func NewGraph(p Point) *Graph {
	g := newGraph(map[Point]*Graph{}, p)
	calculateDependencies(g)
	return g
}

// Division calculates the pointwise division of the graph given an initial division.
func (p *Graph) Division(d Division) map[Point]Division {
	res := map[Point]Division{p.point: d}
	nodes := p.graph()
	for _, p := range nodes {
		dd := Division{}
		for k, v := range d {
			dd[k] = v
		}
		res[p.point] = dd
	}
	changed := true
	update := func(p *Graph, v types.Object, x bool) {
		y := res[p.point][v]
		x = x && y
		if y != x {
			changed = true
		}
		res[p.point][v] = x
	}
	for changed {
		changed = false
		for _, p := range nodes {
			for _, q := range p.dataDeps {
				for k, v := range res[q.point] {
					update(p, k, v)
				}
			}
			for _, d := range p.uses {
				update(p, p.defs, res[p.point][d])
			}
			for _, q := range p.loopDeps {
				log.Println(p, p.defs, q.point)
				update(p, p.defs, q.point.CouldBeTrue(res[q.point]))
			}
		}
	}
	return res
}

func newGraph(seen map[Point]*Graph, p Point) *Graph {
	g := seen[p]
	if g != nil {
		return g
	}
	g = &Graph{point: p}
	seen[p] = g
	if d := p.Defs(); p != nil {
		g.defs = d
	}
	if d := p.Uses(); p != nil {
		g.uses = d
	}
	for _, q := range p.Next() {
		g.link(newGraph(seen, q))
	}
	return g
}

func (p *Graph) link(q *Graph) {
	p.next = append(p.next, q)
	q.prev = append(q.prev, p)
}

func (p *Graph) graph() []*Graph {
	seen := map[*Graph]bool{}
	pp := transitiveClosure(seen, p, func(p *Graph) []*Graph {
		return p.next
	})
	if !seen[p] {
		pp = append(pp, p)
	}
	return pp
}

func (p *Graph) successors() []*Graph {
	return transitiveClosure(map[*Graph]bool{}, p, func(p *Graph) []*Graph {
		return p.next
	})
}

func (p *Graph) precursors() []*Graph {
	return transitiveClosure(map[*Graph]bool{}, p, func(p *Graph) []*Graph {
		return p.prev
	})
}

func transitiveClosure(seen map[*Graph]bool, p *Graph, f func(*Graph) []*Graph) (res []*Graph) {
	var step func(p *Graph)
	step = func(p *Graph) {
		if seen[p] {
			return
		}
		seen[p] = true
		res = append(res, p)
		for _, p := range f(p) {
			step(p)
		}
	}
	step(p)
	return res
}

func calculateDependencies(p *Graph) {
	for _, p := range p.graph() {
		if p.uses == nil {
			continue
		}
		for _, q := range p.precursors() {
			if !p.dependsUpon(q) {
				continue
			}
			if p != q {
				p.dataDeps = append(p.dataDeps, q)
			}
			calculateDependencyLoops(p, q)
		}
	}
}

func calculateDependencyLoops(p, q *Graph) {
	if !reachable(p, q) {
		return
	}
	for _, r := range p.precursors() {
		if len(r.next) <= 1 {
			continue
		}
		if reachable(p, r) {
			p.loopDeps = append(p.loopDeps, r)
		}
	}
	if len(p.loopDeps) == 0 {
		p.loopDeps = infiniteLoop
	}
}

func (p *Graph) dependsUpon(q *Graph) bool {
	for _, d := range p.uses {
		if d == q.defs {
			return true
		}
	}
	return false
}

// whether it is possible to go from p to q
func reachable(p, q *Graph) bool {
	for _, p := range p.successors() {
		if p == q {
			return true
		}
	}
	return false
}

type infinitePoint struct{}

func (infinitePoint) Defs() types.Object          { return nil }
func (infinitePoint) Uses() []types.Object        { return nil }
func (infinitePoint) Next() []Point               { return nil }
func (infinitePoint) CouldBeTrue(d Division) bool { return false }

var infiniteLoop = []*Graph{&Graph{point: infinitePoint{}}}
