package puzzlegen

import (
	"errors"
	"log"
	"sync"

	"github.com/garlicgarrison/chess-puzzle-gen/stockpool"
	chess "github.com/garlicgarrison/go-chess"
	"github.com/garlicgarrison/go-chess/uci"
)

/*
	When only given certain positions, it is difficult to find material advantage
	puzzles, but finding mate in N puzzles are quite simple
*/
var (
	ErrQueueEmpty = errors.New("queue empty")
   	ErrQueueFull  = errors.New("queue full")
)

type MatePuzzleGenerator struct {
	pool *stockpool.StockPool
	q chan *chess.Position
	cache map[string]bool
	depth int
	mutex sync.Mutex
	quit chan bool
}

func NewMatePuzzleGenerator(pool *stockpool.StockPool, queueLimit int, depth int) *MatePuzzleGenerator {
	return &MatePuzzleGenerator{
		pool: pool,
		q: make(chan *chess.Position, queueLimit),
		cache: make(map[string]bool),

		depth: depth,

		quit: make(chan bool),
		mutex: sync.Mutex{},
	}
}

func (g *MatePuzzleGenerator) Start() {
	go func() {
		for {
			log.Printf("Current Queue length: %d", len(g.q))
			
			quit := false
			select {
			case position := <- g.q:
				go func() {
					log.Printf("Analyzing Position: %s", position.String())
					searchRes := g.analyzePosition(position)
					log.Printf("Position %s analyzed", position.String())
					
					if searchRes == nil {
						return 
					}

					// log.Printf("Puzzle position: %s", position.String())
					// mateInN, solutions := g.analyzeResults(searchRes.MultiPV)
					// log.Printf("Puzzle solution (mate in %d): %s", mateInN, solutions)
				}()
			case <- g.quit:
				log.Printf("Goodbye :)")
				quit = true
			} 
	
			if quit {
				break
			}
		}
	}()

}

func (g *MatePuzzleGenerator) Add(position *chess.Position) error {
	if g.cache[position.String()] {
		return nil
	}
	
	select {
	case g.q <- position: 
		g.mutex.Lock()
		g.cache[position.String()] = true
		g.mutex.Unlock()

		return nil
	default: 
		return ErrQueueFull
	}
}

func (g *MatePuzzleGenerator) Close() {
	g.quit <- true
}

/* This takes the position and returns the search results of that position */
func (g *MatePuzzleGenerator) analyzePosition(position *chess.Position) *uci.SearchResults {
	if position == nil {
		return nil
	}

	cmdPos := uci.CmdPosition{Position: position}
	cmdGo := uci.CmdGo{Depth: g.depth}

	instance := g.pool.Acquire()

	defer g.pool.Release(instance)

	err := instance.Engine.Run(cmdPos, cmdGo)
	if err != nil {
		log.Printf("Error --  %s", err)
	}

	res := instance.Engine.SearchResults()
	return &res
}

/* 
	Returns a move node given a position and results 
	1. If it is the opponent's move, just return their best move
	2. If own move, we rank our moves and look deeper
*/
func (g *MatePuzzleGenerator) createMateTree(position *chess.Position) *MoveTree {
	// startPos, err := chess.FEN(position.String())
	// if err != nil {
	// 	return nil
	// }

	// startGame := chess.NewGame(startPos)
	// if startGame == nil {
	// 	return nil
	// }

	// getLines := func(pos *chess.Position) []*chess.Move {
	// 	search := g.analyzePosition(pos)
	// 	if search == nil {
	// 		return nil
	// 	}
	
	// 	pvs := search.MultiPV
	// 	possibleLines := make(map[string]bool)
	// 	moves := []*chess.Move{}
	// 	for _, info := range pvs {
	// 		if info.Score.Mate <= 0 || possibleLines[info.PV[0].String()] {
	// 			continue
	// 		}
			
	// 		moves = append(moves, info.PV[0])
	// 		possibleLines[info.PV[0].String()] = true
	// 	}

	// 	return moves
	// }

	// getBest := func(pos *chess.Position) *chess.Move {
	// 	search := g.analyzePosition(pos)
	// 	if search == nil {
	// 		return nil
	// 	}

	// 	return search.BestMove
	// }

	// traverse := func() {
		
	// }

	tree := &MoveTree{
		start: position,
		head: []*MoveNode{},
	}

	return tree
} 