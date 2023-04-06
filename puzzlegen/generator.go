package puzzlegen

import chess "github.com/garlicgarrison/go-chess"

type Generator[T any] interface {
	Start()
	Close()

	Create(*chess.Position) (*chess.Game, int)
}
