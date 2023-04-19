package beautify

import (
	"log"
	"math"
	"math/rand"

	"github.com/garlicgarrison/chess-puzzle-gen/puzzlegen"
	"github.com/garlicgarrison/go-chess"
)

func energy(deltaE float64, temperature float64) float64 {
	return math.Pow(math.E, -1*deltaE/temperature)
}

type Method string

const (
	GEOMETRIC Method = "geometric"
	LINEAR    Method = "linear"
	SLOW      Method = "slow"
)

type AnnealConfig struct {
	InitTemp        float64
	FinalTemp       float64
	Alpha           float64
	Beta            float64
	Method          Method
	Iterations      int
	AcceptableScore float64

	NumPieces int
}

type Annealer struct {
	cfg AnnealConfig
	g   puzzlegen.Generator[*chess.Position]
}

func NewAnnealer(cfg AnnealConfig, g puzzlegen.Generator[*chess.Position]) *Annealer {
	return &Annealer{
		cfg: cfg,
		g:   g,
	}
}

func (a *Annealer) Anneal(p *puzzlegen.Puzzle) *puzzlegen.Puzzle {
	temperature := a.cfg.InitTemp
	currentScore := a.Score(*p)
	nextScore := 0.0
	for temperature >= a.cfg.FinalTemp {
		for i := 0; i < a.cfg.Iterations; i++ {
			nextFEN, err := puzzlegen.MutateFEN(p.Position, a.cfg.NumPieces)
			if err != nil {
				return nil
			}
			log.Printf("mutated fen: %s", nextFEN)

			f, err := chess.FEN(nextFEN)
			if err != nil {
				return nil
			}

			var puzzle *puzzlegen.Puzzle
			game := chess.NewGame(f)
			sol, res := a.g.Create(game.Position())
			solution := []string{}
			if sol != nil {
				for _, m := range sol.Moves() {
					solution = append(solution, m.String())
				}
			}

			if res == nil {
				puzzle = &puzzlegen.Puzzle{
					Position: nextFEN,
					Solution: solution,
					MateIn:   0,
					CP:       0,
				}
			} else {
				puzzle = &puzzlegen.Puzzle{
					Position: nextFEN,
					Solution: solution,
					MateIn:   res.Info.Score.Mate,
					CP:       res.Info.Score.CP,
				}
			}

			nextScore = a.Score(*puzzle)
			log.Printf("nextScore: %f", nextScore)
			log.Printf("currentScore: %f", currentScore)
			if nextScore > currentScore {
				p = puzzle
				currentScore = nextScore
			} else {
				energy := energy(currentScore-nextScore, temperature)
				log.Printf("energy: %f", energy)
				if energy > rand.Float64() {
					p = puzzle
					currentScore = nextScore
				}
			}
		}

		switch a.cfg.Method {
		case SLOW:
			temperature = temperature / (1 + a.cfg.Beta*temperature)
		case LINEAR:
			temperature -= a.cfg.Alpha
		case GEOMETRIC:
			temperature *= a.cfg.Alpha
		}
	}

	return p
}
