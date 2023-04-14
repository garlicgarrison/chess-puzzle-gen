package beautify

import (
	"log"
	"testing"

	"github.com/garlicgarrison/chess-puzzle-gen/puzzlegen"
)

func TestScore(t *testing.T) {
	controlPuzzle := puzzlegen.Puzzle{
		Position: "R3nN2/8/Pk5P/3b4/7P/6r1/2pnN2p/2K5 b - - 0 1",
		Solution: []string{"h2h1q", "e2g1", "h1g1", "c1c2", "d5b3", "c2d2"},
		MateIn:   4,
	}

	score := Score(controlPuzzle)
	log.Printf("score: %f", score)

	higherPuzzle := puzzlegen.Puzzle{
		Position: "6k1/3b3r/1p1p4/p1n2p2/1PPNpP1q/P3Q1p1/1R1RB1P1/5K2 b - - 0 1",
		Solution: []string{"h4f4", "e2f3", "f4e3", "f3h5", "h7h5", "d4f3", "h5h1", "f3g1", "h1g1"},
		MateIn:   5,
	}
	score = Score(higherPuzzle)
	log.Printf("score: %f", score)

	higherPuzzle = puzzlegen.Puzzle{
		Position: "2q1nk1r/4Rp2/1ppp1P2/6Pp/3p1B2/3P3P/PPP1Q3/6K1 w - - 0 1",
		Solution: []string{"e7e8", "c8e8", "f4d6", "e8e7", "e2e7", "f8g8", "e7e8", "g8h7", "e8f7"},
		MateIn:   5,
	}
	score = Score(higherPuzzle)
	log.Printf("score: %f", score)
}
