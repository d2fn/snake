package main

import (
	"fmt"
	lg "github.com/charmbracelet/lipgloss"
)

var tailStyle = lg.NewStyle().Foreground(lg.Color("180"))
var headStyle = lg.NewStyle().Foreground(lg.Color("204"))

type Snake struct {
	head      *element
	direction Dir
	maxLen    int
	frame     int
	length    int
}

type Dir int

const (
	Stopped Dir = iota
	Left
	Down
	Up
	Right
)

var directionVecs = map[Dir]vec {
	Stopped: {  0,  0},
	Left:    { -1,  0},
	Down:    {  0,  1},
	Up:      {  0, -1},
	Right:   {  1,  0},
}

var directionChars = map[Dir]string {
	Stopped: "\u25CB",
	Left: "\u25C0",
	Down: "\u25BC",
	Up: "\u25B2",
	Right: "\u25BA",
}

func (snake *Snake) Velocity() vec {
	return directionVecs[snake.direction]
}

func (snake *Snake) HeadString() string {
	return directionChars[snake.direction]
}

func (snake *Snake) IsMoving() bool {
	return snake.direction != Stopped
}

func (snake *Snake) Update() {
	if snake.head != nil && snake.IsMoving() {
		next := snake.head
		vel := snake.Velocity()
		newHead := &element{
			p: vec{
				x: next.p.x + vel.x,
				y: next.p.y + vel.y,
			}}
		newHead.next = next
		snake.head = newHead
		snake.frame++
		if snake.frame%10 == 0 {
			snake.maxLen++
		}
		snake.Trim(snake.maxLen)
	}
}

func (snake *Snake) CheckForCollisions() bool {
	var collisions = make(map[vec]int)
	hits := 0
	nodes := 0
	for n := snake.head; n != nil; n = n.next {
		nodes++
		collisions[n.p]++
		if collisions[n.p] > 1 {
			hits++
		}
	}

	if debug {
		if hits > 0 {
			fmt.Printf("hits = %d, nodes = %d\n", hits, nodes)
			fmt.Printf("%q\n", collisions)
		}
	}
	return hits > 0
}

func (snake *Snake) Up() {
	snake.direction = Up
}

func (snake *Snake) Down() {
	snake.direction = Down
}

func (snake *Snake) Left() {
	snake.direction = Left
}
func (snake *Snake) Right() {
	snake.direction = Right
}

func (snake *Snake) Trim(maxLength int) {
	var i = 0
	for n := snake.head; n != nil; n = n.next {
		i++
		snake.length = i
		if i >= maxLength {
			n.next = nil
			return
		}
	}
}

func (snake *Snake) View(g *Grid) {

	if snake.head != nil {
		for node := snake.head; node != nil; node = node.next {
			p := node.p
			g.Set(tailStyle.Render("\u25E6"), p.x, p.y)
		}
		snakeHead := snake.head.p
		g.Set(headStyle.Render(snake.HeadString()), snakeHead.x, snakeHead.y)
	}
}
