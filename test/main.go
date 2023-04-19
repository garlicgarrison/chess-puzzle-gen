package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"time"

	"github.com/garlicgarrison/chess-puzzle-gen/beautify"
	"github.com/garlicgarrison/chess-puzzle-gen/puzzlegen"
	"github.com/garlicgarrison/chess-puzzle-gen/stockpool"
	"github.com/garlicgarrison/go-chess"
	"gopkg.in/yaml.v2"
)

func main() {
	// initialize stockfish pool
	pool, err := stockpool.NewStockPool("stockfish", 1, 8, 10)
	if err != nil {
		panic(err)
	}

	// get puzzle config
	yamlConfig, err := ioutil.ReadFile("config/pieces.yaml")
	if err != nil {
		panic(err)
	}

	var config puzzlegen.PuzzleConfig
	err = yaml.Unmarshal(yamlConfig, &config)
	if err != nil {
		panic(err)
	}

	// initilialize mate generator
	gen := puzzlegen.NewMatePuzzleGenerator(&puzzlegen.Cfg{
		puzzlegen.AnalysisConfig{
			Depth:   14,
			MultiPV: 2,
		},
		config,
	}, pool, func(s string, i int, g *chess.Game) {}, 10)

	beautify := beautify.NewAnnealer(beautify.AnnealConfig{
		InitTemp:        200,
		FinalTemp:       0.5,
		Alpha:           1,
		Beta:            0.02,
		Method:          beautify.LINEAR,
		Iterations:      1000,
		AcceptableScore: 10,

		NumPieces: 5,
	}, gen)

	controlPuzzle := puzzlegen.Puzzle{
		Position: "R3nN2/8/Pk5P/3b4/7P/6r1/2pnN2p/2K5 b - - 0 1",
		Solution: []string{"h2h1q", "e2g1", "h1g1", "c1c2", "d5b3", "c2d2"},
		MateIn:   4,
	}

	now := time.Now()
	puzzle := beautify.Anneal(&controlPuzzle)
	if puzzle != nil {
		write(*puzzle)
	}

	log.Printf("puzzle fen: %s", puzzle.Position)
	log.Printf("time: %d", time.Since(now))
}

func write(puzzle puzzlegen.Puzzle) {
	f, err := ioutil.ReadFile("puzzles.json")
	if err != nil {
		log.Printf("read error -- %s", err)
		return
	}

	p := &puzzlegen.Puzzles{}
	err = json.Unmarshal(f, p)
	if err != nil {
		log.Printf("unmarshal error -- %s", err)
		return
	}

	p.Puzzles = append(p.Puzzles, puzzle)

	b, err := json.Marshal(p)
	if err != nil {
		log.Printf("marshal error -- %s", err)
		return
	}

	err = ioutil.WriteFile("puzzles.json", b, 0777)
	if err != nil {
		log.Printf("write error -- %s", err)
		return
	}
}
