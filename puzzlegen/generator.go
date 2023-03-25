package puzzlegen

type Generator[T any] interface {
	Start()
	Close()

	Add(T) error
}