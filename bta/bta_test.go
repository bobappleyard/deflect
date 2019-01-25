package bta

import (
	"testing"
	"testing/quick"
)

type variable struct {
	Name string
}

type point struct {
	Def, Use           *variable
	Prev, Next         []*point
	LoopDeps, DataDeps []*point
}

func (p *point) link(q *point) {
	p.Next = append(p.Next, q)
	q.Prev = append(q.Prev, p)
}

func (p *point) graph() []*point {
	seen := map[*point]bool{}
	pp := transitiveClosure(seen, p, func(p *point) []*point {
		return p.Next
	})
	if !seen[p] {
		pp = append(pp, p)
	}
	return pp
}

func (p *point) successors() []*point {
	return transitiveClosure(map[*point]bool{}, p, func(p *point) []*point {
		return p.Next
	})
}

func (p *point) precursors() []*point {
	return transitiveClosure(map[*point]bool{}, p, func(p *point) []*point {
		return p.Prev
	})
}

func transitiveClosure(seen map[*point]bool, p *point, f func(*point) []*point) (res []*point) {
	var step func(p *point)
	step = func(p *point) {
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

type division map[variable]bool

func buildDependencyGraph(p *point) {
	for _, p := range p.graph() {
		if p.Use == nil {
			continue
		}
		prev := p.precursors()
		for _, q := range prev {
			if p.Use != q.Def {
				continue
			}
			if p != q {
				p.DataDeps = append(p.DataDeps, q)
			}
			if !reachable(p, q) {
				continue
			}
			// The point under consideration can reach a precursor; this implies
			// that we are in a loop. Search for branch points that could
			// terminate this loop.
			for _, r := range prev {
				if len(r.Next) <= 1 {
					continue
				}
				if reachable(p, r) {
					// We have found a branch point that could break the loop.
					p.LoopDeps = append(p.LoopDeps, r)
				}
			}
			// No such branch points were found, so the loop is infinite. By adding
			// a dependency on a point that is otherwise unreachable, we force our
			// value to be unknown.
			if len(p.LoopDeps) == 0 {
				p.LoopDeps = append(p.LoopDeps, &point{Use: p.Use})
			}
		}
	}
}

// whether it is possible to go from p to q
func reachable(p, q *point) bool {
	for _, p := range p.successors() {
		if p == q {
			return true
		}
	}
	return false
}

func pointwiseDivision(p *point, d division) map[*point]division {
	res := map[*point]division{p: d}
	nodes := p.graph()
	for _, p := range nodes {
		dd := division{}
		for k, v := range d {
			dd[k] = v
		}
		res[p] = dd
	}
	changed := true
	update := func(p *point, v variable, x bool) {
		y := res[p][v]
		x = x && y
		if y != x {
			changed = true
		}
		res[p][v] = x
	}
	for changed {
		changed = false
		for _, p := range nodes {
			for _, q := range p.DataDeps {
				for k, v := range res[q] {
					update(p, k, v)
				}
			}
			if p.Def != nil && p.Use != nil {
				update(p, *p.Def, res[p][*p.Use])
				for _, q := range p.LoopDeps {
					update(p, *p.Def, res[q][*q.Use])
				}
			}
		}
	}
	return res
}

func TestUse(t *testing.T) {
	x, y := variable{"x"}, variable{"y"}

	block := &point{Def: &x, Use: &y}
	buildDependencyGraph(block)

	type state struct{ X, Y bool }
	err := quick.CheckEqual(func(s state) bool {
		d := pointwiseDivision(block, division{x: s.X, y: s.Y})
		return d[block][x]
	}, func(s state) bool {
		return s.X && s.Y
	}, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestIndirectUse(t *testing.T) {
	x, y, z := variable{"x"}, variable{"y"}, variable{"z"}

	pre := &point{Def: &y, Use: &z}
	block := &point{Def: &x, Use: &y}
	pre.link(block)
	buildDependencyGraph(pre)

	type state struct{ X, Z bool }
	err := quick.CheckEqual(func(s state) bool {
		d := pointwiseDivision(pre, division{x: s.X, y: true, z: s.Z})
		return d[block][x]
	}, func(s state) bool {
		return s.X && s.Z
	}, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestInfiniteLoop(t *testing.T) {
	x := variable{"x"}

	block := &point{Def: &x, Use: &x}
	block.link(block)
	buildDependencyGraph(block)

	type state struct{ X bool }
	err := quick.CheckEqual(func(s state) bool {
		d := pointwiseDivision(block, division{x: s.X})
		return d[block][x]
	}, func(s state) bool {
		return false
	}, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestBasicLoop(t *testing.T) {
	x, y := variable{"x"}, variable{"y"}

	loop := &point{Use: &y}
	block := &point{Def: &x, Use: &x}
	ret := &point{}
	loop.link(block)
	loop.link(ret)
	block.link(loop)
	buildDependencyGraph(loop)

	type state struct{ X, Y bool }
	err := quick.CheckEqual(func(s state) bool {
		d := pointwiseDivision(loop, division{x: s.X, y: s.Y})
		return d[block][x]
	}, func(s state) bool {
		return s.X && s.Y
	}, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestIndirectLoop(t *testing.T) {
	x, y, z := variable{"x"}, variable{"y"}, variable{"z"}

	pre := &point{Def: &y, Use: &z}
	loop := &point{Use: &y}
	block := &point{Def: &x, Use: &x}
	ret := &point{}
	pre.link(loop)
	loop.link(block)
	loop.link(ret)
	block.link(loop)
	buildDependencyGraph(pre)

	type state struct{ X, Z bool }
	err := quick.CheckEqual(func(s state) bool {
		d := pointwiseDivision(pre, division{x: s.X, y: true, z: s.Z})
		return d[block][x]
	}, func(s state) bool {
		return s.X && s.Z
	}, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestLoopWithBreak(t *testing.T) {
	x, y, z := variable{"x"}, variable{"y"}, variable{"z"}

	loop := &point{Use: &y}
	block := &point{Def: &x, Use: &x}
	bk := &point{Use: &z}
	ret := &point{}
	loop.link(block)
	loop.link(ret)
	block.link(bk)
	bk.link(ret)
	bk.link(loop)
	buildDependencyGraph(loop)

	type state struct{ X, Y, Z bool }
	err := quick.CheckEqual(func(s state) bool {
		d := pointwiseDivision(loop, division{x: s.X, y: s.Y, z: s.Z})
		return d[block][x]
	}, func(s state) bool {
		return s.X && s.Y && s.Z
	}, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestLoopWithoutSelfDep(t *testing.T) {
	x, y, z := variable{"x"}, variable{"y"}, variable{"z"}

	loop := &point{Use: &y}
	block := &point{Def: &x, Use: &z}
	ret := &point{}
	loop.link(block)
	loop.link(ret)
	block.link(loop)
	buildDependencyGraph(loop)

	type state struct{ X, Y, Z bool }
	err := quick.CheckEqual(func(s state) bool {
		d := pointwiseDivision(loop, division{x: s.X, y: s.Y, z: s.Z})
		return d[block][x]
	}, func(s state) bool {
		return s.X && s.Z
	}, nil)
	if err != nil {
		t.Error(err)
	}
}
