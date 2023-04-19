package puzzlegen

import (
	"log"
	"testing"
)

func TestPosition(t *testing.T) {
	cfg := PuzzleConfig{
		WhiteQ: 1,
		WhiteR: 1,
		WhiteB: 1,
		WhiteN: 1,
		WhiteP: 1,
		BlackQ: 1,
		BlackR: 1,
		BlackB: 1,
		BlackN: 1,
		BlackP: 1,
	}

	fen, err := GenerateRandomFEN(cfg)
	if err != nil {
		t.Fatalf("err -- %s", err)
	}
	log.Printf("fen generated -- %s", fen)

	for i := 0; i < 10; i++ {
		fen, err = MutateFEN(fen, 5)
		if err != nil {
			t.Fatalf("err -- %s", err)
		}
		log.Printf("fen mutated -- %s", fen)
	}
}
