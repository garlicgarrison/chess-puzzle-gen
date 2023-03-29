package puzzlegen

import (
	"errors"
	"log"
	"sort"
	"strconv"

	"github.com/garlicgarrison/chess-puzzle-gen/stockpool"
	chess "github.com/garlicgarrison/go-chess"
	"github.com/garlicgarrison/go-chess/uci"
)

const (
	MAX_PROCESSES = 10
	TIMEOUT       = 1000
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

type Cfg struct {
	AnalysisConfig
	PuzzleConfig
}

type MatePuzzleGenerator struct {
	cfg   Cfg
	pool  *stockpool.StockPool
	write func(string, []*chess.Game)
	q     chan *chess.Position
	quit  chan bool
}

func NewMatePuzzleGenerator(cfg Cfg, pool *stockpool.StockPool, write func(string, []*chess.Game), queueLimit int) Generator[*chess.Position] {
	return &MatePuzzleGenerator{
		cfg:   cfg,
		pool:  pool,
		write: write,
		q:     make(chan *chess.Position, queueLimit),
		quit:  make(chan bool),
	}
}

func (g *MatePuzzleGenerator) Start() {
	go func() {
		for {
			fen, err := GenerateRandomFEN(g.cfg.PuzzleConfig)
			if err != nil {
				log.Printf("error -- %s", err)
			}

			f, err := chess.FEN(fen)
			if err != nil {
				log.Printf("error -- %s", err)
			}
			game := chess.NewGame(f)
			log.Printf("new position -- %s", fen)

			solutions := g.create(game.Position())
			if solutions != nil {
				g.write(fen, solutions)
			}
		}
	}()
}

func (g *MatePuzzleGenerator) Close() {
	g.quit <- true
}

// TODO: make into a tree
func (g *MatePuzzleGenerator) create(position *chess.Position) []*chess.Game {
	solutions := g.mateSolutions(position)
	if len(solutions) == 0 {
		return nil
	}

	return solutions
}

/*
	This takes the position and returns the search results of that position
*/
func (g *MatePuzzleGenerator) analyzePosition(position *chess.Position, depth int, multiPV int) *uci.SearchResults {
	if position == nil {
		return nil
	}

	cmdPos := uci.CmdPosition{Position: position}
	cmdGo := uci.CmdGo{Depth: g.cfg.Depth}

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
	2. We only need to check the moves after the mate solution is found
*/
func (g *MatePuzzleGenerator) mateSolutions(position *chess.Position) []*chess.Game {
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
		queueLen := len(queue)

		for i := 0; i < queueLen; i++ {
			currGame := queue[0]
			queue = queue[1:]

			mateMoves := g.mateMoves(currGame.Position())
			if len(mateMoves) == 0 {
				return solutions
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
					continue
				}

				move := g.getBestMove(newGame.Position())
				newGame.Move(move)
				queue = append(queue, newGame)
			}
		}
	}

	return solutions
}

/*
	NOTE: if all the scores are the same, there could possibly be other lines
	that have the same exact score, therefore, the result is incomplete and invalid

	This returns the moves with the shortest mating moves
*/
func (g *MatePuzzleGenerator) mateMoves(position *chess.Position) []*chess.Move {
	search := g.analyzePosition(position, g.cfg.Depth, g.cfg.MultiPV)

	if search == nil {
		return nil
	}

	pvs := search.MultiPV
	sort.Slice(pvs, func(i, j int) bool {
		return pvs[i].Score.Mate < pvs[j].Score.Mate
	})

	possibleLines := make(map[string]bool)
	moves := []*chess.Move{}
	prev := -1
	allSame := true
	for _, info := range pvs {
		if info.Score.Mate <= 0 {
			continue
		}

		if prev != -1 && info.Score.Mate != prev {
			allSame = false
			break
		}

		if !possibleLines[info.PV[0].String()] {
			prev = info.Score.Mate
			possibleLines[info.PV[0].String()] = true
			moves = append(moves, info.PV[0])
		}
	}

	if allSame {
		return nil
	}

	return moves
}

func (g *MatePuzzleGenerator) getBestMove(position *chess.Position) *chess.Move {
	search := g.analyzePosition(position, g.cfg.Depth, 1)
	if search == nil {
		return nil
	}

	return search.BestMove
}
