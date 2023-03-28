package main

import (
	"io/ioutil"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/garlicgarrison/chess-puzzle-gen/puzzlegen"
	"github.com/garlicgarrison/chess-puzzle-gen/stockpool"
	"github.com/garlicgarrison/go-chess"
	"github.com/spf13/cobra"
	yaml "gopkg.in/yaml.v2"
)

const (
	STOCKFISHPATH = "stockfish"
	CRYSTALPATH   = "./stockfish/crystal"

	CONFIGPATH = "./config/pieces.yaml"
)

func main() {
	var depth int
	var multipv int

	rootCmd := &cobra.Command{
		Use:   "puzzlegen",
		Short: "Generate beautiful puzzles",
		Long:  "Generate beautiful puzzles",
		Run: func(cmd *cobra.Command, args []string) {
			// initialize stockfish pool
			pool, err := stockpool.NewStockPool(CRYSTALPATH, 10, 4, 10)
			if err != nil {
				panic(err)
			}

			// initilialize mate generator
			gen := puzzlegen.NewMatePuzzleGenerator(puzzlegen.AnalysisConfig{
				Depth:   depth,
				MultiPV: multipv,
			}, pool, 10)
			gen.Start()

			defer gen.Close()

			// initialize feeder
			yamlConfig, err := ioutil.ReadFile("config/pieces.yaml")
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
					log.Printf("FEN: %s", fen)
					log.Printf("Error -- %s", err)
					return nil
				}

				gameF, err := chess.FEN(fen)
				if err != nil {
					log.Printf("FEN: %s", fen)
					log.Printf("Error -- %s", err)
					return nil
				}

				game := chess.NewGame(gameF)
				return game.Position()
			}, gen)
			feeder.Start(500)

			defer feeder.Close()

			// closing operations
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			<-sigChan
			log.Printf("exit")
		},
	}

	rootCmd.Flags().IntVarP(&depth, "depth", "d", 0, "The depth parameter")
	rootCmd.Flags().IntVarP(&multipv, "multipv", "m", 0, "The multipv parameter")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error -- %s", err)
	}
}
