package puzzlegen

import (
	"time"

	"github.com/garlicgarrison/go-chess"
)

type Feeder[T any] struct {
	f    func() T
	gen  Generator[T]
	quit chan bool
}

func NewPositionFeeder[T *chess.Position](f func() T, gen Generator[T]) *Feeder[T] {
	return &Feeder[T]{
		f:    f,
		gen:  gen,
		quit: make(chan bool),
	}
}

func NewGameFeeder[T *chess.Game](f func() T, gen Generator[T]) *Feeder[T] {
	return &Feeder[T]{
		f:    f,
		gen:  gen,
		quit: make(chan bool),
	}
}

func (f *Feeder[T]) Start(timeout int) {
	go func() {
		for {
			quit := false
			select {
			case <-f.quit:
				quit = true
			default:
				pos := f.f()

				err := f.gen.Add(pos)
				if err != nil {
					time.Sleep(time.Millisecond * time.Duration(timeout))
				}
			}

			if quit {
				break
			}
		}
	}()
}

func (f *Feeder[T]) Close() {
	f.quit <- true
}
