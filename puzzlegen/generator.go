package puzzlegen

import (
	chess "github.com/garlicgarrison/go-chess"
	"github.com/garlicgarrison/go-chess/uci"
)

type Generator[T any] interface {
	Start()
	Close()

	Create(*chess.Position) (*chess.Game, *uci.SearchResults)
	Analyze(*chess.Position, int, int) *uci.SearchResults
}
