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
	cfg   *Cfg
	pool  *stockpool.StockPool
	write func(string, int, *chess.Game)
	q     chan *chess.Position
	quit  chan bool
}

func NewMatePuzzleGenerator(cfg *Cfg, pool *stockpool.StockPool, write func(string, int, *chess.Game), queueLimit int) Generator[*chess.Position] {
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

			solution, res := g.Create(game.Position())
			if solution != nil {
				g.write(fen, res.Info.Score.Mate, solution)
			}
		}
	}()
}

func (g *MatePuzzleGenerator) Close() {
	g.quit <- true
}

func (g *MatePuzzleGenerator) Create(position *chess.Position) (*chess.Game, *uci.SearchResults) {
	return g.mateSolutions(position)
}

/*
	This takes the position and returns the search results of that position
*/
func (g *MatePuzzleGenerator) Analyze(position *chess.Position, depth int, multiPV int) *uci.SearchResults {
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
		log.Printf("error -- %s -- position: %s", err, position.String())
		panic(err)
	}

	res := instance.Engine.SearchResults()
	return &res
}

/*
	Returns a move tree given a position and results
	1. If it is the opponent's move, just return their best move
	2. We only need to check the moves after the mate solution is found

	NOTE: decrease the depth every iteration by 1
*/
func (g *MatePuzzleGenerator) mateSolutions(position *chess.Position) (*chess.Game, *uci.SearchResults) {
	startPos, err := chess.FEN(position.String())
	if err != nil {
		return nil, nil
	}

	game := chess.NewGame(startPos)
	if game == nil {
		return nil, nil
	}

	if game.Outcome() != chess.NoOutcome {
		return nil, nil
	}

	var searchResults *uci.SearchResults
	for {
		res := g.Analyze(game.Position(), g.cfg.Depth, g.cfg.MultiPV)
		mateMove := g.mateMove(res)
		if mateMove == nil {
			if searchResults != nil {
				return game, searchResults
			} else {
				return nil, res
			}
		}

		if searchResults == nil {
			searchResults = res
		}

		game.Move(mateMove)
		if game.Outcome() == chess.NoOutcome {
			res = g.Analyze(game.Position(), g.cfg.Depth, g.cfg.MultiPV)
			bestReply := g.bestMove(res)
			game.Move(bestReply)
			continue
		}

		return game, searchResults
	}
}

/*
	NOTE: if all the scores are the same, there could possibly be other lines
	that have the same exact score, therefore, the result is incomplete and invalid

	This returns the moves with the shortest mating moves, and mate in N
*/
func (g *MatePuzzleGenerator) mateMove(search *uci.SearchResults) *chess.Move {
	pvs := search.MultiPV
	sort.Slice(pvs, func(i, j int) bool {
		return pvs[i].Score.Mate < pvs[j].Score.Mate
	})

	var solution *chess.Move
	minMate := -1
	for _, info := range pvs {
		if info.Score.Mate <= 0 {
			continue
		}

		if solution == nil {
			solution = info.PV[0]
			minMate = info.Score.Mate
			continue
		}

		if minMate == info.Score.Mate {
			return nil
		}

		return solution
	}

	return solution
}

func (g *MatePuzzleGenerator) bestMove(search *uci.SearchResults) *chess.Move {
	return search.BestMove
}
