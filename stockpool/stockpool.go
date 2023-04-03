package stockpool

import (
	"errors"
	"strconv"
	"time"

	"github.com/garlicgarrison/go-chess/uci"
	guuid "github.com/google/uuid"
)

var (
	ErrPathNotFound       = errors.New("path not found")
	ErrWrongStockInstance = errors.New("wrong instance released")
)

// TODO: add threads
type StockInstance struct {
	id     guuid.UUID
	Engine *uci.Engine
}

type StockPool struct {
	idSet   map[guuid.UUID]bool
	pool    chan *StockInstance
	threads int
	timeout int
}

func NewStockPool(path string, limit, threads, timeout int) (*StockPool, error) {
	idSet := make(map[guuid.UUID]bool)
	ch := make(chan *StockInstance, limit)

	for i := 0; i < limit; i++ {
		eng, err := uci.New(path)
		if err != nil {
			return nil, ErrPathNotFound
		}

		eng.Run(uci.CmdSetOption{
			Name:  "Threads",
			Value: strconv.FormatInt(int64(threads), 10),
		})

		eng.Run(uci.CmdSetOption{
			Name:  "Hash",
			Value: "2048",
		})

		id := guuid.New()
		idSet[id] = true
		ch <- &StockInstance{
			id:     id,
			Engine: eng,
		}
	}

	return &StockPool{
		idSet:   idSet,
		pool:    ch,
		threads: threads,
		timeout: timeout,
	}, nil
}

func (sp *StockPool) Acquire() *StockInstance {
	for {
		select {
		case instance := <-sp.pool:
			return instance
		default:
			time.Sleep(time.Duration(sp.timeout) * time.Millisecond)
		}
	}
}

func (sp *StockPool) Release(si *StockInstance) error {
	_, ok := sp.idSet[si.id]
	if !ok {
		return ErrWrongStockInstance
	}

	sp.pool <- si
	return nil
}
