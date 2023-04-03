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
			pool, err := stockpool.NewStockPool(STOCKFISHPATH, 1, threads, 10)
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
	rootCmd.Flags().IntVarP(&threads, "threads", "t", 0, "The threads parameter")

	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("Error -- %s", err)
	}
}

type Puzzle struct {
	Position string   `json:"position"`
	Solution []string `json:"solution"`
}

type Puzzles struct {
	Puzzles []Puzzle `json:"puzzles"`
}

func write(fen string, sol *chess.Game) {
	f, err := ioutil.ReadFile("puzzles.json")
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

	puzzle := Puzzle{
		Position: fen,
		Solution: []string{},
	}

	solution := []string{}
	for _, m := range sol.Moves() {
		solution = append(solution, m.String())
	}
	puzzle.Solution = solution
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
