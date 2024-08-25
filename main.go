package main

import (
	"fmt"
	"image/color"
	"log"
	"math/rand"

	"github.com/hajimehoshi/ebiten/v2"
)

const (
	Columns = 640
	Rows    = 320
	MaxTPS  = 60
)
const (
	CellEmpty Cell = iota // 0
	CellAlive             // 1
	CellBorn
	CellDying
)

var (
	Colors = []color.Color{
		color.RGBA{0x34, 0x49, 0x5e, 0xff}, // empty #34495e
		color.RGBA{0x1e, 0x84, 0x49, 0xff}, // alive #1e8449
		// color.RGBA{0xab, 0xeb, 0xc6, 0xff},  // born  #abebc6
		color.RGBA{0x27, 0xae, 0x60, 0xff}, // born  #27ae60
		// color.RGBA{0xf5, 0xb7, 0xb1, 0xff},  // dying #f5b7b1
		color.RGBA{0x64, 0x1e, 0x16, 0xff}, // dying #641e16
	}
)

type Cell byte

type Row []Cell

type Board struct {
	Rows []Row
}

func NewBoard(nColumns, nRows int) *Board {
	rows := make([]Row, nRows)
	for i := range rows {
		r := make(Row, nColumns)
		rows[i] = r
		for j := 0; j < nColumns; j++ {
			r[j] = CellEmpty
		}
	}
	return &Board{
		Rows: rows,
	}
}

func (b *Board) RandomFill() {
	for _, row := range b.Rows {
		for j := range row {
			row[j] = Cell(rand.Intn(len(Colors)))
		}
	}
}

type Game struct {
	Board     *Board
	FeedPhase bool
	Paused    bool
	Speed     int
}

func NewGame(nColumns, nRows int) *Game {
	g := &Game{
		Board:     NewBoard(nColumns, nRows),
		FeedPhase: false,
		Paused:    false,
		Speed:     10,
	}
	g.Board.RandomFill()
	ebiten.SetTPS(g.Speed)
	return g
}

func (g *Game) Update() error {
	g.processInput()
	if g.Paused {
		return nil
	}
	fp := g.FeedPhase
	g.FeedPhase = !fp
	if fp {
		return g.feed()
	}
	return g.lifeOn()
}

func (g *Game) processInput() {
	switch {
	case ebiten.IsKeyPressed(ebiten.KeySpace):
		g.Paused = !g.Paused
	case ebiten.IsKeyPressed(ebiten.KeyLeft):
		g.Speed--
		if g.Speed < 1 {
			g.Speed = 1
		}
		ebiten.SetTPS(g.Speed)
	case ebiten.IsKeyPressed(ebiten.KeyRight):
		g.Speed++
		if g.Speed > MaxTPS {
			g.Speed = MaxTPS
		}
		ebiten.SetTPS(g.Speed)
	}
}

func (g *Game) feed() error {
	var countBoard []Row
	rows := g.Board.Rows
	tmp := make(Row, len(rows[0]))
	for i, row := range rows {
		i1 := i - 1
		i2 := i + 1
		if i1 < 0 {
			i1 = len(rows) - 1
		}
		if i2 >= len(rows) {
			i2 = 0
		}
		// Make sum of three adjacent rows into tmp.
		for j, cell := range row {
			count := cell + rows[i1][j] + rows[i2][j]
			tmp[j] = count
		}
		// Make sum of three adjacent columns into crow.
		crow := make(Row, len(row))
		for j, cnt := range tmp {
			j1 := j - 1
			j2 := j + 1
			if j1 < 0 {
				j1 = len(tmp) - 1
			}
			if j2 >= len(tmp) {
				j2 = 0
			}
			// We count all 9 cells, minus the central one.
			crow[j] = cnt + tmp[j1] + tmp[j2] - row[j]
		}
		countBoard = append(countBoard, crow)
	}
	// We make a final sweep to check the fate of the cells.
	for i, row := range rows {
		for j, cell := range row {
			nbrs := countBoard[i][j]
			switch cell {
			case CellEmpty:
				if nbrs == 3 {
					row[j] = CellBorn
				}
			case CellAlive:
				if nbrs < 2 || nbrs > 3 {
					row[j] = CellDying
				}
			default:
				panic(fmt.Sprintf("wrong cell value %d at feed phase", cell))
			}
		}
	}
	return nil
}

func (g *Game) lifeOn() error {
	empty := 0
	filled := 0
	for i, row := range g.Board.Rows {
		for j, cell := range row {
			switch cell {
			case CellEmpty, CellDying:
				row[j] = CellEmpty
				empty++
			case CellBorn, CellAlive:
				row[j] = CellAlive
				filled++
			default:
				panic(fmt.Sprintf("cell value %d is wrong at (%d,%d)", cell, j, i))
			}
		}
	}
	log.Printf("lifeOn: empty=%d filled=%d", empty, filled)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	for y, row := range g.Board.Rows {
		for x, cell := range row {
			clr := Colors[cell]
			screen.Set(x, y, clr)
		}
	}
}

func (g *Game) Layout(oW, oH int) (int, int) {
	return len(g.Board.Rows[0]), len(g.Board.Rows)
}

func main() {
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	ebiten.SetWindowSize(Columns, Rows)
	ebiten.SetWindowTitle("convay's life")
	ebiten.SetVsyncEnabled(true)
	log.Printf("about to start game")
	log.Printf("cell values: %d, %d, %d, %d", CellEmpty, CellBorn, CellAlive, CellDying)
	if err := ebiten.RunGame(NewGame(Columns, Rows)); err != nil {
		log.Fatal(err)
	}
}
