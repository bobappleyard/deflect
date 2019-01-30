package bta

import (
	"go/types"
	"testing"
	"testing/quick"
)

type testPoint struct {
	def, use types.Object
	next     []Point
}

func (p *testPoint) Defs() types.Object          { return p.def }
func (p *testPoint) Uses() []types.Object        { return []types.Object{p.use} }
func (p *testPoint) Next() []Point               { return p.next }
func (p *testPoint) CouldBeTrue(d Division) bool { return d[p.use] }
func (p *testPoint) link(q Point)                { p.next = append(p.next, q) }

func variable(name string) types.Object {
	return types.NewVar(0, nil, name, &types.Basic{})
}

func TestUse(t *testing.T) {
	x, y := variable("x"), variable("y")

	block := &testPoint{def: x, use: y}
	g := NewGraph(block)

	type state struct{ X, Y bool }
	err := quick.CheckEqual(func(s state) bool {
		d := g.Division(Division{x: s.X, y: s.Y})
		return d[block][x]
	}, func(s state) bool {
		return s.X && s.Y
	}, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestIndirectUse(t *testing.T) {
	x, y, z := variable("x"), variable("y"), variable("z")

	pre := &testPoint{def: y, use: z}
	block := &testPoint{def: x, use: y}
	pre.link(block)
	g := NewGraph(pre)

	type state struct{ X, Z bool }
	err := quick.CheckEqual(func(s state) bool {
		d := g.Division(Division{x: s.X, y: true, z: s.Z})
		return d[block][x]
	}, func(s state) bool {
		return s.X && s.Z
	}, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestInfiniteLoop(t *testing.T) {
	x := variable("x")

	block := &testPoint{def: x, use: x}
	block.link(block)
	g := NewGraph(block)

	type state struct{ X bool }
	err := quick.CheckEqual(func(s state) bool {
		d := g.Division(Division{x: s.X})
		return d[block][x]
	}, func(s state) bool {
		return false
	}, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestBasicLoop(t *testing.T) {
	x, y := variable("x"), variable("y")

	loop := &testPoint{use: y}
	block := &testPoint{def: x, use: x}
	ret := &testPoint{}
	loop.link(block)
	loop.link(ret)
	block.link(loop)
	g := NewGraph(loop)

	type state struct{ X, Y bool }
	err := quick.CheckEqual(func(s state) bool {
		d := g.Division(Division{x: s.X, y: s.Y})
		return d[block][x]
	}, func(s state) bool {
		return s.X && s.Y
	}, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestIndirectLoop(t *testing.T) {
	x, y, z := variable("x"), variable("y"), variable("z")

	pre := &testPoint{def: y, use: z}
	loop := &testPoint{use: y}
	block := &testPoint{def: x, use: x}
	ret := &testPoint{}
	pre.link(loop)
	loop.link(block)
	loop.link(ret)
	block.link(loop)
	g := NewGraph(pre)

	type state struct{ X, Z bool }
	err := quick.CheckEqual(func(s state) bool {
		d := g.Division(Division{x: s.X, y: true, z: s.Z})
		return d[block][x]
	}, func(s state) bool {
		return s.X && s.Z
	}, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestLoopWithBreak(t *testing.T) {
	x, y, z := variable("x"), variable("y"), variable("z")

	loop := &testPoint{use: y}
	block := &testPoint{def: x, use: x}
	bk := &testPoint{use: z}
	ret := &testPoint{}
	loop.link(block)
	loop.link(ret)
	block.link(bk)
	bk.link(ret)
	bk.link(loop)
	g := NewGraph(loop)

	type state struct{ X, Y, Z bool }
	err := quick.CheckEqual(func(s state) bool {
		d := g.Division(Division{x: s.X, y: s.Y, z: s.Z})
		return d[block][x]
	}, func(s state) bool {
		return s.X && s.Y && s.Z
	}, nil)
	if err != nil {
		t.Error(err)
	}
}

func TestLoopWithoutSelfDep(t *testing.T) {
	x, y, z := variable("x"), variable("y"), variable("z")

	loop := &testPoint{use: y}
	block := &testPoint{def: x, use: z}
	ret := &testPoint{}
	loop.link(block)
	loop.link(ret)
	block.link(loop)
	g := NewGraph(loop)

	type state struct{ X, Y, Z bool }
	err := quick.CheckEqual(func(s state) bool {
		d := g.Division(Division{x: s.X, y: s.Y, z: s.Z})
		return d[block][x]
	}, func(s state) bool {
		return s.X && s.Z
	}, nil)
	if err != nil {
		t.Error(err)
	}
}
