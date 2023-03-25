package puzzlegen

import (
	"github.com/garlicgarrison/go-chess"
)

type MoveNode struct {
	move *chess.Move
	responses []*MoveNode
}

type MoveTree struct {
	start *chess.Position
	head []*MoveNode
}