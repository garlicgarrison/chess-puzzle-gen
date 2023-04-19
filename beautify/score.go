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
	Sacrifice     = 10.0
	// Any underpromotion
	UnderPromotion = 15.0

	CP         = 2.0 / 100.0
	PieceDiff  = -1.5
	MateReward = 250.0
)

var PieceValueMap = map[chess.PieceType]int{
	chess.Pawn:   1,
	chess.Bishop: 3,
	chess.Knight: 3,
	chess.Rook:   5,
	chess.Queen:  9,
}

//TODO: the more pieces the opponent has compared to you, the higher the score should be
func (a *Annealer) Score(p puzzlegen.Puzzle) float64 {
	score := 0.0
	f, err := chess.FEN(p.Position)
	if err != nil {
		return -100
	}

	game := chess.NewGame(f)
	if game == nil {
		return -100
	}

	if p.MateIn > 0 {
		score += MateReward
	}

	// number of pieces
	// pieces := 0.0
	// for _, piece := range p.Position {
	// 	_, ok := puzzlegen.PieceToBit[piece]
	// 	if ok {
	// 		pieces++
	// 	}
	// }
	// score += (32.0 - pieces) * NumPieces

	// diff pieces on each side
	whitePieces := 0.0
	blackPieces := 0.0
	for _, p := range game.Position().Board().SquareMap() {
		if p.Type() == chess.NoPieceType {
			continue
		}

		if p.Color() == chess.White {
			whitePieces += float64(PieceValueMap[p.Type()])
			continue
		}
		blackPieces += float64(PieceValueMap[p.Type()])
	}
	if game.Position().Turn() == chess.White {
		score += (whitePieces - blackPieces) * PieceDiff
	} else {
		score += (blackPieces - whitePieces) * PieceDiff
	}

	if game.Outcome() != chess.NoOutcome {
		return score
	}

	// if there is no mate, calculate by the cp score
	if len(p.Solution) == 0 {
		return score + float64(p.CP)*CP
	}

	// mate moves diff
	diff := len(p.Solution)/2 - p.MateIn + 1
	if diff == 0 {
		score += MateMovesDiff
	} else {
		score += (float64(1) / float64(diff)) * MateMovesDiff
	}

	// see how much material is lost/gained
	underPromotions := 0
	totalLost := 0
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
