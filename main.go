package main

import (
	"fmt"
	"os"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
)

var debug = false

var maxSnakeLength = 500

type Drawable interface {
	Update(m *model)
	View(g *Grid)
	AccumulatePositions(dst map[vec]int)
}

type Wall struct {
	p vec
}

func (w *Wall) Update(m *model) {

}

func (w Wall) View(g *Grid) {
	var wallStyle = lg.NewStyle().Foreground(lg.Color("87")).Background(lg.Color("67"))
	g.Set(wallStyle.Render(" "), w.p.x, w.p.y)
}

func (w Wall) AccumulatePositions(dst map[vec]int) {
	dst[w.p]++
}

func InitWall(x, y int) Drawable {
	return &Wall{vec{x: x, y: y}}
}

type ScoreBanner struct {
	text string
}

func (sb *ScoreBanner) Update(m *model) {
	sb.text = fmt.Sprintf("  Score: %10d  ", m.score)
}

func (sb ScoreBanner) View(g *Grid) {
	g.PlaceText(sb.text, 10, 0)
}

func (sb *ScoreBanner) AccumulatePositions(dst map[vec]int) {
	x0 := 10
	for x := range utf8.RuneCountInString(sb.text) {
		dst[vec{x0+x,0}]++
	}
}



type model struct {
	player    Snake
	drawables []Drawable
	scoreDrawable Drawable
	score     int
	width     int
	height    int
	quitting  bool
}

type element struct {
	p    vec
	next *element
}

type vec struct {
	x, y int
}

type TickMsg struct {
	Time time.Time
	tag  int
	ID   int
}

func initialModel() model {
	return model{
		drawables: make([]Drawable, 0),
		scoreDrawable: &ScoreBanner {},
	}
}

func (m *model) SpawnSnakeAt(p vec) {
	m.player = Snake { direction: m.player.direction }
	m.player.head = &element{p, nil}
	m.player.maxLen = maxSnakeLength
	m.score = 0
}

func (m model) Tick() tea.Cmd {
	return tea.Tick(
		time.Second/15,
		func(t time.Time) tea.Msg { return TickMsg{
				Time: time.Now(),
				ID:   1,
				tag:  1,
			}
		})
}

func (m model) Init() tea.Cmd {
	return m.Tick()
}

func (m model) CenterPoint() vec {
	return vec{m.width / 2, m.height / 2}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "q", "esc", "ctrl+c":
			m.quitting = true
			return m, tea.Quit
		case tea.KeyUp.String():
			m.player.Up()
			return m, nil
		case tea.KeyDown.String():
			m.player.Down()
			return m, nil
		case tea.KeyLeft.String():
			m.player.Left()
			return m, nil
		case tea.KeyRight.String():
			m.player.Right()
			return m, nil
		}
		return m, nil
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.SpawnSnakeAt(m.CenterPoint())
		m.InitWalls()
		return m, tea.EnterAltScreen
	case TickMsg:
		m.Advance()
		return m, m.Tick()
	}

	return m, nil
}

func (m *model) InitWalls() {
	for x := range m.width {
		m.drawables = append(m.drawables, InitWall(x, 0))
	}

	for y := 1; y < m.height; y++ {
		m.drawables = append(m.drawables, InitWall(m.width-1, y))
	}

	for x := m.width - 2; x >= 0; x-- {
		m.drawables = append(m.drawables, InitWall(x, m.height-1))
	}

	for y := m.height - 1; y >= 1; y-- {
		m.drawables = append(m.drawables, InitWall(0, y))
	}
}

func (m *model) Advance() {

	m.score += m.player.length
	m.scoreDrawable.Update(m)

	s := &m.player
	s.Update()
	
	if s.CheckForCollisions() {
		m.SpawnSnakeAt(s.head.p)
	}

	obstacles := make(map[vec]int)
	for _, d := range m.drawables {
		d.AccumulatePositions(obstacles)
	}

	for n := s.head; n != nil; n = n.next {
		if obstacles[n.p] > 0 {
			m.SpawnSnakeAt(m.CenterPoint())
		}
	}
}

func (m model) View() string {

	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	grid := NewGrid(m.width, m.height)
	for _, d := range m.drawables {
		d.View(&grid)
	}
	m.player.View(&grid)
	m.scoreDrawable.View(&grid)
	return grid.View()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
