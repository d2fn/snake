package main

import (
	"strings"
)

type Grid struct {
	width, height int
	grid          [][]string
}

func NewGrid(width, height int) Grid {

	grid := make([][]string, height)
	for i := range grid {
		grid[i] = make([]string, width)
	}

	for x := range width {
		for y := range height {
			grid[y][x] = " "
		}
	}

	return Grid{width, height, grid}
}

func (g *Grid) Set(s string, x, y int) {
	g.grid[y][x] = s
}

func (g *Grid) PlaceText(s string, x, y int) {
	r := []rune(s)
	for xx, c := range r {
		g.Set(string(c), x + xx, y)
	}
}

func (g *Grid) View() string {
	lines := make([]string, len(g.grid))
	for i, row := range g.grid {
		lines[i] = strings.Join(row, "")
	}
	return strings.Join(lines, "\n")
}
