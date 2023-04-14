package beautify

import (
	"io/ioutil"
	"log"
	"testing"

	"github.com/garlicgarrison/chess-puzzle-gen/puzzlegen"
	"github.com/garlicgarrison/chess-puzzle-gen/stockpool"
	"github.com/garlicgarrison/go-chess"
	"gopkg.in/yaml.v2"
)

func TestAnneal(t *testing.T) {
	// initialize stockfish pool
	pool, err := stockpool.NewStockPool("stockfish", 1, 8, 10)
	if err != nil {
		panic(err)
	}

	// get puzzle config
	yamlConfig, err := ioutil.ReadFile("../config/pieces.yaml")
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
			Depth:   10,
			MultiPV: 2,
		},
		config,
	}, pool, func(s string, i int, g *chess.Game) {}, 10)

	beautify := NewAnnealer(AnnealConfig{
		InitTemp:        500,
		FinalTemp:       5,
		Alpha:           100,
		Beta:            0,
		Method:          LINEAR,
		Iterations:      5,
		AcceptableScore: 10,
	}, gen)

	controlPuzzle := puzzlegen.Puzzle{
		Position: "R3nN2/8/Pk5P/3b4/7P/6r1/2pnN2p/2K5 b - - 0 1",
		Solution: []string{"h2h1q", "e2g1", "h1g1", "c1c2", "d5b3", "c2d2"},
		MateIn:   4,
	}
	puzzle := beautify.Anneal(&controlPuzzle)
	log.Printf("puzzle fen: %s", puzzle.Position)
}
