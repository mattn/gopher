package main

import (
	gopherlib "github.com/mattn/gopher"
)

type (
	mode  int
	scene int
)

const (
	Walking mode = iota
	Jumping
	HighJumping
	Waiting
	Stopping
	SL
)

type Gopher struct {
	msg    gopherlib.Msg      // current message
	task   chan gopherlib.Msg // message queue
	x, y   int                // location of gopher
	w, h   int                // size of gopher
	dx, dy int                // amount of movement
	wait   int                // counter for delay
	mode   mode               // animation mode
	scene  scene              // index of scenes
}

func (g *Gopher) NextScene() scene {
	g.scene += 1
	if g.scene == 4 {
		g.scene = 0
	}
	return g.scene
}

func (g *Gopher) Scene() scene {
	return g.scene
}

func (g *Gopher) X() int {
	return g.x
}

func (g *Gopher) Y() int {
	return g.y
}

func (g *Gopher) W() int {
	return g.w
}

func (g *Gopher) H() int {
	return g.h
}

func (g *Gopher) Dx() int {
	return g.dx
}

func (g *Gopher) Dy() int {
	return g.dy
}

func (g *Gopher) Motion() {
	switch g.mode {
	case HighJumping:
		g.x += g.dx / 2
		g.y += g.dy
		g.dy++
		if g.dy > 20 {
			g.SetMode(Walking)
		}
	case Jumping:
		g.x += g.dx / 2
		g.y += g.dy
		g.dy++
		if g.dy > 10 {
			g.SetMode(Walking)
		}
	default:
		g.x += g.dx
		g.y += g.dy
	}
}

func (g *Gopher) Move(dx, dy int) {
	g.x = dx
	g.y = dy
}

func (g *Gopher) MoveTo(x, y int) {
	g.x = x
	g.y = y
}

func (g *Gopher) Take() (string, bool) {
	select {
	case g.msg = <-g.task:
		return g.msg.Method, true
	default:
	}
	return "", false
}

func (g *Gopher) Mode() mode {
	return g.mode
}

func (g *Gopher) ClearMsg() {
	for len(g.task) > 0 {
		<-g.task
	}
}

func (g *Gopher) PushMsg(m gopherlib.Msg) {
	g.task <- m
}

func (g *Gopher) SetMode(m mode) {
	g.mode = m
	switch m {
	case Walking:
		g.dy = 0
	case Stopping:
		g.wait = 10
	case Jumping:
		g.dy = -10
	case HighJumping:
		g.scene = 0
		g.dy = -20
	case Waiting:
		g.scene = 0
		g.wait = 100
	}
}

func (g *Gopher) Turn() {
	g.dx = -g.dx // turn over
}

func (g *Gopher) WakeUp() {
	g.wait = 0
}

func (g *Gopher) Time() int {
	return g.wait
}

func (g *Gopher) Idle() int {
	g.wait--
	return g.wait
}

func (g *Gopher) CurrentSceneInfo() sceneInfo {
	s := g.scene
	if g.mode == Waiting {
		s = 4
	}
	dir := 0
	if g.dx < 0 {
		dir = 1
	}
	return scenes[dir][s]
}

func (g *Gopher) SetContent(c string) {
	g.msg.Content = c
}

func (g *Gopher) Content() string {
	return g.msg.Content
}

func (g *Gopher) Link() string {
	return g.msg.Link
}

func (g *Gopher) Busy() bool {
	return len(gopher.task) == cap(gopher.task)
}
