package main

import (
	"encoding/json"
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
	var threads int

	rootCmd := &cobra.Command{
		Use:   "puzzlegen",
		Short: "Generate beautiful puzzles",
		Long:  "Generate beautiful puzzles",
		Run: func(cmd *cobra.Command, args []string) {
			// initialize stockfish pool
			pool, err := stockpool.NewStockPool(CRYSTALPATH, 10, threads, 10)
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
			gen := puzzlegen.NewMatePuzzleGenerator(puzzlegen.Cfg{
				puzzlegen.AnalysisConfig{
					Depth:   depth,
					MultiPV: multipv,
				},
				config,
			}, pool, write, 10)
			gen.Start()

			defer gen.Close()

			// closing operations
			sigChan := make(chan os.Signal, 1)
			signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

			<-sigChan
			log.Printf("exit")
			os.Exit(0)
		},
	}

	rootCmd.Flags().IntVarP(&depth, "depth", "d", 0, "The depth parameter")
	rootCmd.Flags().IntVarP(&multipv, "multipv", "m", 0, "The multipv parameter")
	rootCmd.Flags().IntVarP(&multipv, "threads", "t", 2, "The threads parameter")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error -- %s", err)
	}
}

type Puzzle struct {
	Position  string     `json:"position"`
	Solutions [][]string `json:"solutions"`
}

type Puzzles struct {
	Puzzles []Puzzle `json:"puzzles"`
}

func write(fen string, sols []*chess.Game) {
	f, err := ioutil.ReadFile("puzzles.json")
	if err != nil {
		return
	}

	p := &Puzzles{}
	err = json.Unmarshal(f, p)
	if err != nil {
		return
	}

	puzzle := Puzzle{
		Position:  fen,
		Solutions: [][]string{},
	}
	for _, s := range sols {
		solution := []string{}
		for _, m := range s.Moves() {
			solution = append(solution, m.String())
		}

		puzzle.Solutions = append(puzzle.Solutions, solution)
	}
	p.Puzzles = append(p.Puzzles, puzzle)

	b, err := json.Marshal(p)
	if err != nil {
		return
	}

	err = ioutil.WriteFile("puzzles.json", b, 0777)
	if err != nil {
		return
	}
}
