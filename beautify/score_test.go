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

func TestScore(t *testing.T) {
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

	score := beautify.Score(controlPuzzle)
	log.Printf("score: %f", score)

	higherPuzzle := puzzlegen.Puzzle{
		Position: "6k1/3b3r/1p1p4/p1n2p2/1PPNpP1q/P3Q1p1/1R1RB1P1/5K2 b - - 0 1",
		Solution: []string{"h4f4", "e2f3", "f4e3", "f3h5", "h7h5", "d4f3", "h5h1", "f3g1", "h1g1"},
		MateIn:   5,
	}
	score = beautify.Score(higherPuzzle)
	log.Printf("score: %f", score)

	higherPuzzle = puzzlegen.Puzzle{
		Position: "2q1nk1r/4Rp2/1ppp1P2/6Pp/3p1B2/3P3P/PPP1Q3/6K1 w - - 0 1",
		Solution: []string{"e7e8", "c8e8", "f4d6", "e8e7", "e2e7", "f8g8", "e7e8", "g8h7", "e8f7"},
		MateIn:   5,
	}
	score = beautify.Score(higherPuzzle)
	log.Printf("score: %f", score)

	higherPuzzle = puzzlegen.Puzzle{
		Position: "8/8/8/8/8/2B1p3/K7/2N4k w - - 0 1",
		Solution: []string{},
		MateIn:   0,
		CP:       370,
	}
	score = beautify.Score(higherPuzzle)
	log.Printf("score: %f", score)
}
