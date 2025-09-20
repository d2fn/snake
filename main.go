package main

import (
	"fmt"
	"os"
	"slices"
	"sort"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	lg "github.com/charmbracelet/lipgloss"
)

var debug = false

var maxSnakeLength = 500

type model struct {
	player        Snake

	// game board
	drawables     []Drawable
	gameBoard     *Window

	// score board
	scoreBoard    *Window
	scoreDrawable Drawable
	hiScores      []int
	score         int

	// screen size
	width         int
	height        int

	quitting      bool
}

type Window struct {
	ul, lr vec
}

func (w Window) Width() int {
	return w.lr.x - w.ul.x
}

func (w Window) Height() int {
	return w.lr.y - w.ul.y
}

func (w Window) ToScreen(p vec) vec {
	return vec{p.x + w.ul.x, p.y + w.ul.y}
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

type Drawable interface {
	Update(m *model)
	View(g *Grid, window *Window)
	AccumulatePositions(dst map[vec]int)
}

type Wall struct {
	p vec
}

func (w *Wall) Update(m *model) {

}

func (w Wall) View(g *Grid, window *Window) {
	var wallStyle = lg.NewStyle().Foreground(lg.Color("87")).Background(lg.Color("67"))
	p := window.ToScreen(w.p)
	g.Set(wallStyle.Render(" "), p.x, p.y)
}

func (w Wall) AccumulatePositions(dst map[vec]int) {
	dst[w.p]++
}

func InitWall(x, y int) Drawable {
	return &Wall{vec{x: x, y: y}}
}

type ScoreBanner struct {
	text []string
}

func (sb *ScoreBanner) Update(m *model) {

	lines := make([]string, 1)
	lines[0] = "HI SCORES"

	rankedScores := make([]int, 0)
	rankedScores = append(rankedScores, m.hiScores...)
	rankedScores = append(rankedScores, m.score)

	sort.Sort(sort.Reverse(sort.IntSlice(rankedScores)))

	if len(rankedScores) == 0 {
		line := fmt.Sprintf("   > %10d", m.score)
		lines = append(lines, line)
	} else {
		for i := range len(rankedScores) {
			rank := i+1
			score := rankedScores[i]
			var line string
			if score == m.score {
				line = fmt.Sprintf("   > %10d", rankedScores[i])
			} else {
				line = fmt.Sprintf("%3d: %10d", rank, rankedScores[i])
			}
			lines = append(lines, line)
		}
	}

	sb.text = lines
}


func (sb ScoreBanner) View(g *Grid, window *Window) {
	p := window.ToScreen(vec { 0, 0 })
	for i := range len(sb.text) {
		g.PlaceText(sb.text[i], p.x, p.y + i)
	}
}

func (sb *ScoreBanner) AccumulatePositions(dst map[vec]int) { }

func initialModel() model {
	return model{
		drawables:     make([]Drawable, 0),
		scoreDrawable: &ScoreBanner{},
		hiScores:      make([]int, 0),
	}
}

func (m *model) resetScore() {
	m.hiScores = append(m.hiScores, m.score)
	slices.Sort(m.hiScores)
	m.score = 0
}

func (m *model) SpawnSnakeAt(p vec) {
	m.player = Snake{direction: m.player.direction}
	m.player.head = &element{p, nil}
	m.player.maxLen = maxSnakeLength
}

func (m model) Tick() tea.Cmd {
	return tea.Tick(
		time.Second/15,
		func(t time.Time) tea.Msg {
			return TickMsg{
				Time: time.Now(),
				ID:   1,
				tag:  1,
			}
		})
}

func (m model) Init() tea.Cmd {
	return m.Tick()
}

func (w Window) CenterPoint() vec {
	return vec{ (w.lr.x - w.ul.x) / 2, (w.lr.y - w.ul.y) / 2 }
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
		m.scoreBoard = &Window{ vec { 0, 0 }, vec { 24, m.height-1} }
		m.gameBoard = &Window{ vec { 25, 0 }, vec { m.width-1, m.height-1 } }
		maxSnakeLength = 2 * m.gameBoard.Width()
		m.SpawnSnakeAt(m.gameBoard.CenterPoint())
		m.InitWalls()
		return m, tea.EnterAltScreen
	case TickMsg:
		m.Advance()
		return m, m.Tick()
	}

	return m, nil
}

func (m *model) InitWalls() {

	for x := range m.gameBoard.Width() {
		m.drawables = append(m.drawables, InitWall(x, 0))
		m.drawables = append(m.drawables, InitWall(x, m.gameBoard.Height()-1))
	}

	for y := range m.gameBoard.Height() {
		m.drawables = append(m.drawables, InitWall(0, y))
		m.drawables = append(m.drawables, InitWall(m.gameBoard.Width()-1, y))
	}
}
	
func (m *model) Advance() {

	m.score += m.player.length
	m.scoreDrawable.Update(m)

	s := &m.player
	s.Update()

	// check for snake collisions with itself
	if s.CheckForCollisions() {
		m.SpawnSnakeAt(s.head.p)
		m.resetScore()
	}

	obstacles := make(map[vec]int)
	for _, d := range m.drawables {
		d.AccumulatePositions(obstacles)
	}

	if obstacles[s.head.p] > 0 {
		m.SpawnSnakeAt(m.gameBoard.CenterPoint())
		m.player.direction = Stopped
		m.resetScore()
	}
}

func (m model) View() string {

	if m.width == 0 || m.height == 0 {
		return "Initializing..."
	}

	grid := NewGrid(m.width, m.height)
	for _, d := range m.drawables {
		d.View(&grid, m.gameBoard)
	}
	m.player.View(&grid, m.gameBoard)
	m.scoreDrawable.View(&grid, m.scoreBoard)
	return grid.View()
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
