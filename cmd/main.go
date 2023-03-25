package main

import (
	"io/ioutil"

	"github.com/garlicgarrison/chess-puzzle-gen/puzzlegen"
	"github.com/garlicgarrison/chess-puzzle-gen/stockpool"
	"github.com/garlicgarrison/go-chess"
	"github.com/go-yaml/yaml"
)

const (
	STOCKFISHPATH = "stockfish"
	CRYSTALPATH   = "./stockfish/crystal"

	CONFIGPATH = "./config/pieces.yaml"
)

func main() {
	// initialize stockfish pool
	pool, err := stockpool.NewStockPool(CRYSTALPATH, 10, 4, 10)
	if err != nil {
		panic(err)
	}

	// initilialize mate generator
	gen := puzzlegen.NewMatePuzzleGenerator(puzzlegen.AnalysisConfig{
		Depth:   50,
		MultiPV: 2,
	}, pool, 1000)
	gen.Start()

	defer gen.Close()

	// initialize feeder
	yamlConfig, err := ioutil.ReadFile("config.yaml")
	if err != nil {
		panic(err)
	}

	var config puzzlegen.PuzzleConfig
	err = yaml.Unmarshal(yamlConfig, &config)
	if err != nil {
		panic(err)
	}

	feeder := puzzlegen.NewPositionFeeder(func() *chess.Position {
		fen, err := puzzlegen.GenerateRandomFEN(config)
		if err != nil {
			return nil
		}

		gameF, err := chess.FEN(fen)
		if err != nil {
			return nil
		}

		game := chess.NewGame(gameF)
		return game.Position()
	}, gen)
	feeder.Start(500)

	defer feeder.Close()
}
