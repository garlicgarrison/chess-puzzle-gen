package puzzlegen

// import (
// 	"io/ioutil"
// 	"log"
// 	"testing"

// 	"github.com/garlicgarrison/chess-puzzle-gen/stockpool"
// 	chess "github.com/garlicgarrison/go-chess"
// 	"gopkg.in/yaml.v2"
// )

// func write(fen string, sols *chess.Game) {

// }

// func TestMateSolutions(t *testing.T) {
// 	pool, err := stockpool.NewStockPool("stockfish", 10, 8, 10)
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

// 	f, err := chess.FEN("1k6/3b3p/3r1Pp1/2N5/4np1p/8/N7/1nR3K1 w - - 0 1")
// 	if err != nil {
// 		log.Printf("error -- %s", err)
// 	}

// 	game := chess.NewGame(f)
// 	solutions := gen.MateSolutions(game.Position())

// 	log.Printf("solutions -- %d", len(solutions.Moves()))
// }
