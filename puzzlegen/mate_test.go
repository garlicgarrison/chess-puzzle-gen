package puzzlegen

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"testing"

	chess "github.com/garlicgarrison/go-chess"
)

type PuzzleJSON struct {
	Position  *string    `json:"position"`
	Solutions [][]string `json:"solutions"`
}

type Puzzles struct {
	Puzzles []*PuzzleJSON `json:"puzzles"`
}

func write(fen string, sols []*chess.Game) {
	f, err := ioutil.ReadFile("../puzzles.json")
	if err != nil {
		log.Printf("read error -- %s", err)
		return
	}

	p := &Puzzles{}
	err = json.Unmarshal(f, p)
	if err != nil {
		log.Printf("unmarshal error -- %s", err)
		return
	}

	puzzle := PuzzleJSON{
		Position:  &fen,
		Solutions: [][]string{},
	}
	for _, s := range sols {
		solution := []string{}
		for _, m := range s.Moves() {
			solution = append(solution, m.String())
		}

		puzzle.Solutions = append(puzzle.Solutions, solution)
	}
	p.Puzzles = append(p.Puzzles, &puzzle)

	b, err := json.Marshal(p)
	if err != nil {
		log.Printf("marshal error -- %s", err)
		return
	}

	err = ioutil.WriteFile("../puzzles.json", b, 0777)
	if err != nil {
		log.Printf("write error -- %s", err)
		return
	}
}

// func TestMateSolutions(t *testing.T) {
// 	pool, err := stockpool.NewStockPool("../stockfish/crystal", 10, 8, 10)
// 	if err != nil {
// 		panic(err)
// 	}

// 	yamlConfig, err := ioutil.ReadFile("../config/pieces.yaml")
// 	if err != nil {
// 		panic(err)
// 	}

// 	var config PuzzleConfig
// 	err = yaml.Unmarshal(yamlConfig, &config)
// 	if err != nil {
// 		panic(err)
// 	}

// 	gen := NewMatePuzzleGenerator(Cfg{
// 		AnalysisConfig{
// 			Depth:   35,
// 			MultiPV: 3,
// 		},
// 		config,
// 	}, pool, write, 10)

// 	f, err := chess.FEN("8/K1Rp2p1/p1r1p1P1/1Qn1n3/8/k7/7b/8 b - - 0 1")
// 	if err != nil {
// 		log.Printf("error -- %s", err)
// 	}

// 	game := chess.NewGame(f)
// 	solutions := gen.MateSolutions(game.Position())
// 	assert.Equal(t, len(solutions), 2, "This puzzle has two solutions")

// 	write("8/K1Rp2p1/p1r1p1P1/1Qn1n3/8/k7/7b/8 b - - 0 1", solutions)
// }

func TestWrite(t *testing.T) {
	f, err := chess.FEN("8/K1Rp2p1/p1r1p1P1/1Qn1n3/8/k7/7b/8 b - - 0 1")
	if err != nil {
		log.Printf("error -- %s", err)
	}

	game := chess.NewGame(f)
	game.Move(game.ValidMoves()[0])

	write("8/K1Rp2p1/p1r1p1P1/1Qn1n3/8/k7/7b/8 b - - 0 1", []*chess.Game{game})
}
