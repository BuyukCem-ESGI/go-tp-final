package main

import (
	"errors"
	"fmt"
	"net/http"
)

type Grid struct {
	height, width int
	grid          []byte
}

func (g Grid) String() string {
	return string(g.grid)
}

func NewGrid(x, y int) Grid {
	wth := 2*x + 2 // +2 colonne de droite + '\n'
	hgt := 2*y + 1 // +1 ligne en dessous

	g := make([]byte, wth*hgt)

	for i := 0; i < hgt; i += 2 {
		row0 := i * wth
		row1 := (i + 1) * wth
		for j := 0; j < wth-2; j += 2 {
			g[row0+j], g[row0+j+1] = '+', '-'
			if row1+j+1 <= wth*hgt {
				g[row1+j], g[row1+j+1] = '|', ' '
			}
		}
		g[row0+wth-2], g[row0+wth-1] = '+', '\n'
		if row1+wth < wth*hgt {
			g[row1+wth-2], g[row1+wth-1] = '|', '\n'
		}
	}

	return Grid{
		height: y,
		width:  x,
		grid:   g,
	}
}

func (g Grid) cellAt(x, y int) (int, error) {
	woff := g.width*2 + 2
	foff := (y*2+1)*woff + x*2 + 1

	if foff > len(g.grid) {
		return 0, errors.New("Erreur: out of range")
	}

	return (y*2+1)*woff + x*2 + 1, nil
}

func (g Grid) set(c byte, x, y int) error {
	idx, err := g.cellAt(x, y)
	if err != nil {
		return err
	}
	g.grid[idx] = c
	return nil
}

func (g Grid) get(x, y int) string {
	idx, err := g.cellAt(x, y)
	if err != nil {
		return "error"
	}
	return string(g.grid[idx])
}

func (g Grid) draw(w http.ResponseWriter) {
	fmt.Fprint(w, "\033[H\033[2J")
	fmt.Fprint(w, "\x0c", g, "\n") 
}

func (g Grid) reset() {
	for i := 0; i < g.height; i++ {
		for j := 0; j < g.width; j++ {
			g.set(' ', i, j)
		}
	}
}