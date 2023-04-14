package beautify

import (
	"github.com/garlicgarrison/chess-puzzle-gen/puzzlegen"
	"github.com/garlicgarrison/go-chess"
)

/*
	These are the weights to score the beauty
	NOTE: these weights could probably be trained by NNs
*/
const (
	// Moves to mate
	MateMovesDiff = 5.0
	Sacrifice     = 6.0
	// Any underpromotion
	UnderPromotion = 9.0
	NumPieces      = 0.6
)

var PieceValueMap = map[chess.PieceType]int{
	chess.Pawn:   1,
	chess.Bishop: 3,
	chess.Knight: 3,
	chess.Rook:   5,
	chess.Queen:  9,
}

func Score(p puzzlegen.Puzzle) float64 {
	score := 0.0
	f, err := chess.FEN(p.Position)
	if err != nil {
		return -1
	}

	// number of pieces
	pieces := 0.0
	for _, piece := range p.Position {
		_, ok := puzzlegen.PieceToBit[piece]
		if ok {
			pieces++
		}
	}
	score += (32.0 - pieces) * NumPieces

	diff := len(p.Solution)/2 - p.MateIn + 1
	if diff == 0 {
		score += MateMovesDiff
	} else {
		score += (float64(1) / float64(diff)) * MateMovesDiff
	}

	// see how much material is lost/gained
	underPromotions := 0
	totalLost := 0
	game := chess.NewGame(f)
	for i, m := range p.Solution {
		move, err := chess.UCINotation{}.Decode(nil, m)
		if err != nil {
			return 0
		}

		game.Move(move)
		if i%2 == 1 {
			continue
		}

		if move.Promo() != chess.Queen {
			underPromotions++
		}
		if move.Promo() != chess.NoPieceType {
			continue
		}

		gameClone := game.Clone()

		lost := materialLost(gameClone)
		if lost < 0.0 {
			continue
		}
		totalLost = totalLost + lost
	}
	score += float64(totalLost)*Sacrifice + float64(underPromotions)*UnderPromotion

	return score
}

// starts with the opponent's move
func materialLost(game *chess.Game) int {
	var maxMove *chess.Move
	maxMaterial := 0
	squareMap := game.Position().Board().SquareMap()
	validMoves := game.ValidMoves()
	for _, move := range validMoves {
		material := PieceValueMap[squareMap[move.S2()].Type()]
		if material > maxMaterial {
			maxMaterial = material
			maxMove = move
		}
	}

	if maxMove == nil || maxMaterial == 0 {
		return 0
	}
	game.Move(maxMove)

	return maxMaterial - materialLost(game)
}
