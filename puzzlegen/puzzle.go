package puzzlegen

type Puzzle struct {
	Position string   `json:"position"`
	Solution []string `json:"solution"`
	MateIn   int      `json:"mate_in"`
}

type Puzzles struct {
	Puzzles []Puzzle `json:"puzzles"`
}
