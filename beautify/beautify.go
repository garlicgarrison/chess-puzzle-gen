package beautify

import (
	"log"
	"math"
	"math/rand"

	"github.com/garlicgarrison/chess-puzzle-gen/puzzlegen"
	"github.com/garlicgarrison/go-chess"
)

func energy(deltaE float64, temperature float64) float64 {
	return math.Pow(math.E, deltaE/temperature)
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
	currentScore := Score(*p)
	nextScore := 0.0

	for temperature >= a.cfg.FinalTemp {
		for i := 0; i < a.cfg.Iterations; i++ {
			nextFEN, err := puzzlegen.MutateFEN(p.Position)
			if err != nil {
				return nil
			}
			log.Printf("mutated fen: %s", nextFEN)

			f, err := chess.FEN(nextFEN)
			if err != nil {
				return nil
			}

			game := chess.NewGame(f)
			sol, mateIn := a.g.Create(game.Position())
			solution := []string{}
			if sol != nil {
				for _, m := range sol.Moves() {
					solution = append(solution, m.String())
				}
			}

			puzzle := &puzzlegen.Puzzle{
				Position: nextFEN,
				Solution: solution,
				MateIn:   mateIn,
			}

			if len(solution) > 0 {
				nextScore = Score(*puzzle)
			}

			if nextScore > currentScore {
				p = puzzle
				currentScore = nextScore
			} else {
				energy := energy(currentScore-nextScore, temperature)
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
