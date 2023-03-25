package puzzlegen

import (
	"errors"
	"log"
	"strconv"
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

/*
	NOTE: if multiPV is 2, there is only 1 unique solution
*/
type AnalysisConfig struct {
	Depth   int
	MultiPV int
}

type MatePuzzleGenerator struct {
	analysisCfg AnalysisConfig
	pool        *stockpool.StockPool
	q           chan *chess.Position
	cache       map[string]bool
	mutex       sync.Mutex
	quit        chan bool
}

func NewMatePuzzleGenerator(analysisConfig AnalysisConfig, pool *stockpool.StockPool, queueLimit int) Generator[*chess.Position] {
	return &MatePuzzleGenerator{
		analysisCfg: analysisConfig,
		pool:        pool,
		q:           make(chan *chess.Position, queueLimit),
		cache:       make(map[string]bool),
		quit:        make(chan bool),
		mutex:       sync.Mutex{},
	}
}

func (g *MatePuzzleGenerator) Start() {
	go func() {
		for {
			quit := false
			select {
			case position := <-g.q:
				go func() {
					solutions := g.getMateSolutions(position)
					log.Printf("%d solutions found...", len(solutions))
					if len(solutions) > 0 {
						log.Printf("Puzzle position: %s", position.String())
						log.Printf("Puzzle solution: %v", solutions[0].Moves())
						for _, m := range solutions[0].Moves() {
							log.Printf("Puzzle solution: %v", m.String())
						}
					}
				}()
			case <-g.quit:
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
func (g *MatePuzzleGenerator) analyzePosition(position *chess.Position, depth int, multiPV int) *uci.SearchResults {
	if position == nil {
		return nil
	}

	cmdPos := uci.CmdPosition{Position: position}
	cmdGo := uci.CmdGo{Depth: g.analysisCfg.Depth}

	instance := g.pool.Acquire()
	instance.Engine.Run(uci.CmdSetOption{
		Name:  "MultiPV",
		Value: strconv.Itoa(multiPV),
	})

	defer g.pool.Release(instance)

	err := instance.Engine.Run(cmdPos, cmdGo)
	if err != nil {
		log.Printf("Error --  %s", err)
	}

	res := instance.Engine.SearchResults()
	return &res
}

/*
	Returns a move tree given a position and results
	1. If it is the opponent's move, just return their best move
	2. If own move, we rank our moves and look deeper
*/
func (g *MatePuzzleGenerator) getMateSolutions(position *chess.Position) []*chess.Game {
	startPos, err := chess.FEN(position.String())
	if err != nil {
		return nil
	}

	startGame := chess.NewGame(startPos)
	if startGame == nil {
		return nil
	}

	queue := []*chess.Game{startGame}
	solutions := []*chess.Game{}
	for len(queue) > 0 {
		solutionFound := false
		queueLen := len(queue)

		for i := 0; i < queueLen; i++ {
			currGame := queue[0]
			queue = queue[1:]

			mateMoves := g.getMateMoves(currGame.Position())
			if len(mateMoves) == 0 {
				return nil
			}

			for _, move := range mateMoves {
				newGame := currGame.Clone()
				err := newGame.Move(move)
				if err != nil {
					continue
				}

				if (newGame.Outcome() == chess.BlackWon || newGame.Outcome() == chess.WhiteWon) &&
					newGame.Method() == chess.Checkmate {
					solutions = append(solutions, newGame)
					solutionFound = true
					continue
				}

				if !solutionFound {
					move := g.getBestMove(newGame.Position())
					newGame.Move(move)
					queue = append(queue, newGame)
				}
			}
		}

		if solutionFound {
			break
		}
	}

	return solutions
}

/*
	NOTE: if all the scores are the same, there could possibly be other lines
	that have the same exact score, therefore, the result is incomplete and invalid
*/
func (g *MatePuzzleGenerator) getMateMoves(position *chess.Position) []*chess.Move {
	search := g.analyzePosition(position, g.analysisCfg.Depth, g.analysisCfg.MultiPV)

	if search == nil {
		return nil
	}

	pvs := search.MultiPV
	possibleLines := make(map[string]bool)
	moves := []*chess.Move{}
	prev := -1
	allSame := true
	for _, info := range pvs {
		if info.Score.Mate <= 0 || possibleLines[info.PV[0].String()] {
			continue
		}

		if prev != 0 && info.Score.Mate != prev {
			allSame = false
		}

		moves = append(moves, info.PV[0])
		prev = info.Score.Mate
		possibleLines[info.PV[0].String()] = true
	}

	if allSame {
		return nil
	}

	return moves
}

func (g *MatePuzzleGenerator) getBestMove(position *chess.Position) *chess.Move {
	search := g.analyzePosition(position, g.analysisCfg.Depth, 1)
	if search == nil {
		return nil
	}

	return search.BestMove
}
