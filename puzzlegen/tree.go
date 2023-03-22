package puzzlegen

import (
	"github.com/garlicgarrison/go-chess"
)

type MoveNode struct {
	move string
	responses []*MoveNode
}

type MoveTree struct {
	start *chess.Position
	head []*MoveNode
}