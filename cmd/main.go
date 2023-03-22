package main

import (
	"time"

	generator "github.com/garlicgarrison/chess-puzzle-gen/puzzlegen"
	"github.com/garlicgarrison/chess-puzzle-gen/stockpool"
	chess "github.com/garlicgarrison/go-chess"
	"github.com/garlicgarrison/go-chess/uci"
)

const (
	STOCKFISHPATH = "stockfish"
	CRYSTALPATH = "./stockfish/crystal"
)

func main() {
	eng, err := uci.New("stockfish")
	if err != nil {
		panic(err)
	}

	defer eng.Close()

	pool, err := stockpool.NewStockPool(STOCKFISHPATH, 10, 10)
	if err != nil {
		panic(err)
	}
	gen := generator.NewMatePuzzleGenerator(pool, 1000, 20)
	gen.Start()
	
	defer gen.Close()

	if err := eng.Run(uci.CmdUCI, uci.CmdIsReady, uci.CmdUCINewGame); err != nil {
		panic(err)
	}

	game := chess.NewGame()
	for game.Outcome() == chess.NoOutcome {
		cmdPos := uci.CmdPosition{Position: game.Position()}
		cmdGo := uci.CmdGo{MoveTime: time.Second / 1000}
		if err := eng.Run(cmdPos, cmdGo); err != nil {
			panic(err)
		}
		move := eng.SearchResults().BestMove
		if err := game.Move(move); err != nil {
			panic(err)
		}

		gen.Add(game.Position())
	}

	time.Sleep(time.Minute * 3)
}