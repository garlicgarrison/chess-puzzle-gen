package puzzlegen

type Puzzle struct {
	Position string   `json:"position"`
	Solution []string `json:"solution"`
	MateIn   int      `json:"mate_in"`
	CP       int      `json:"cp"`
}

type Puzzles struct {
	Puzzles []Puzzle `json:"puzzles"`
}
