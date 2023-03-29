package puzzlegen

import (
	"errors"
	"fmt"
	"math/rand"
	"strings"
	"unicode"
)

var (
	ErrInvalidPuzzleConfig = errors.New("pieces must add up to 30")
)

var pieceToBit = map[rune]uint8{
	'P': 1,
	'N': 2,
	'B': 3,
	'R': 4,
	'Q': 5,
	'K': 6,
	'p': 9,
	'n': 10,
	'b': 11,
	'r': 12,
	'q': 13,
	'k': 14,
}

var bitToPiece = map[uint8]rune{
	1:  'P',
	2:  'N',
	3:  'B',
	4:  'R',
	5:  'Q',
	6:  'K',
	9:  'p',
	10: 'n',
	11: 'b',
	12: 'r',
	13: 'q',
	14: 'k',
}

var castleSquares = map[rune][][]int{
	'K': {
		{7, 5},
		{7, 6},
	},
	'Q': {
		{7, 1},
		{7, 2},
		{7, 3},
	},
	'k': {
		{0, 5},
		{0, 6},
	},
	'q': {
		{0, 1},
		{0, 2},
		{0, 3},
	},
}

// these are simply arbitrary
// maybe we can configure them later
type PuzzleConfig struct {
	WhiteQ uint8 `yaml:"white_q"`
	WhiteR uint8 `yaml:"white_r"`
	WhiteB uint8 `yaml:"white_b"`
	WhiteN uint8 `yaml:"white_n"`
	WhiteP uint8 `yaml:"white_p"`
	BlackQ uint8 `yaml:"black_q"`
	BlackR uint8 `yaml:"black_r"`
	BlackB uint8 `yaml:"black_b"`
	BlackN uint8 `yaml:"black_n"`
	BlackP uint8 `yaml:"black_p"`
}

func validatePuzzleCfg(cfg PuzzleConfig) bool {
	return cfg.WhiteQ+
		cfg.WhiteR+
		cfg.WhiteB+
		cfg.WhiteN+
		cfg.WhiteP+
		cfg.BlackQ+
		cfg.BlackR+
		cfg.BlackB+
		cfg.BlackN+
		cfg.BlackP <= 30
}

/*
	Generates a random valid FEN position from scratch
	NOTE: kings are not in check/checkmate
*/
func GenerateRandomFEN(cfg PuzzleConfig) (string, error) {
	ok := validatePuzzleCfg(cfg)
	if !ok {
		return "", ErrInvalidPuzzleConfig
	}

	board := [8][8]uint8{}

	pieceMap := map[rune]uint8{
		'Q': cfg.WhiteQ,
		'R': cfg.WhiteR,
		'B': cfg.WhiteB,
		'N': cfg.WhiteN,
		'P': cfg.WhiteP,
		'q': cfg.BlackQ,
		'r': cfg.BlackR,
		'b': cfg.BlackB,
		'n': cfg.BlackN,
		'p': cfg.BlackP,
	}

	whiteAttacks := make(map[string]bool)
	blackAttacks := make(map[string]bool)

	for piece, num := range pieceMap {
		for i := uint8(0); i < num; i++ {
			for {
				var pRow, pCol int
				switch piece {
				case 'P', 'p':
					pRow, pCol = rand.Intn(6)+1, rand.Intn(8)
				default:
					pRow, pCol = rand.Intn(8), rand.Intn(8)
				}

				if board[pRow][pCol] == 0 {
					board[pRow][pCol] = pieceToBit[piece]
					break
				}
			}
		}
	}

	for i, row := range board {
		for j, val := range row {
			if val == 0 {
				continue
			}

			piece := bitToPiece[val]
			attacks := attacks(piece, board, i, j)
			for _, a := range attacks {
				if unicode.IsUpper(piece) {
					whiteAttacks[a] = true
				} else {
					blackAttacks[a] = true
				}
			}
		}
	}

	// Add white king
	for {
		pRow, pCol := rand.Intn(8), rand.Intn(8)

		if board[pRow][pCol] == 0 && !blackAttacks[squareToString(pRow, pCol)] {
			board[pRow][pCol] = pieceToBit['K']
			attacks := kingAttacks(board, pRow, pCol)
			for _, a := range attacks {
				whiteAttacks[a] = true
			}
			break
		}
	}

	// Add black king
	for {
		pRow, pCol := rand.Intn(8), rand.Intn(8)

		if board[pRow][pCol] == 0 && !whiteAttacks[squareToString(pRow, pCol)] {
			board[pRow][pCol] = pieceToBit['k']
			break
		}
	}

	var sb strings.Builder
	for i, row := range board {
		empty := 0
		for _, val := range row {
			if val == 0 {
				empty++
				continue
			}

			if empty != 0 {
				sb.WriteString(fmt.Sprintf("%d", empty))
			}

			sb.WriteRune(bitToPiece[val])
			empty = 0
		}

		if empty != 0 {
			sb.WriteString(fmt.Sprintf("%d", empty))
		}

		if i != 7 {
			sb.WriteRune('/')
		}
	}

	// Randomly choose side -- 0 for black 1 for white
	player := 0
	if rand.Intn(2) == 0 {
		sb.WriteString(" b")
	} else {
		player = 1
		sb.WriteString(" w")
	}

	// Check castling rights
	sb.WriteRune(' ')
	if board[7][4] != 'K' && board[0][4] != 'k' {
		sb.WriteRune('-')
		goto EnPassant
	}

	for e, squares := range castleSquares {
		attackFound := false
		for _, s := range squares {
			var check map[string]bool
			if unicode.IsUpper(e) {
				check = blackAttacks
			} else {
				check = whiteAttacks
			}

			if check[squareToString(s[0], s[1])] {
				attackFound = true
				break
			}
		}

		if !attackFound {
			sb.WriteRune(e)
		}
	}

EnPassant:
	sb.WriteRune(' ')
	eSquare := rand.Intn(8)
	if player == 0 && board[3][eSquare] == pieceToBit['p'] && board[2][eSquare]+board[1][eSquare] == 0 {
		sb.WriteRune(rune(eSquare + 97))
		sb.WriteString(fmt.Sprintf("%d", 6))
	} else if player == 1 && board[4][eSquare] == pieceToBit['P'] && board[5][eSquare]+board[6][eSquare] == 0 {
		sb.WriteRune(rune(eSquare + 97))
		sb.WriteString(fmt.Sprintf("%d", 3))
	} else {
		sb.WriteRune('-')
	}

	sb.WriteString(" 0 ")
	sb.WriteRune('1')

	return sb.String(), nil
}

func squareToString(row, col int) string {
	return fmt.Sprintf("%d:%d", row, col)
}

func attacks(piece rune, board [8][8]uint8, row, col int) []string {
	switch piece {
	case 'K', 'k':
		return kingAttacks(board, row, col)
	case 'p':
		return pawnAttacks(false, board, row, col)
	case 'P':
		return pawnAttacks(true, board, row, col)
	case 'N', 'n':
		return knightAttacks(board, row, col)
	case 'B', 'b':
		return bishopAttacks(board, row, col)
	case 'R', 'r':
		return rookAttacks(board, row, col)
	case 'Q', 'q':
		return queenAttacks(board, row, col)
	default:
		return nil
	}
}

func kingAttacks(board [8][8]uint8, row, col int) []string {
	attacks := make([]string, 0)
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if dx == 0 && dy == 0 {
				continue
			}

			x, y := row+dx, col+dy
			if x < 0 || x > 7 || y < 0 || y > 7 {
				continue
			}

			attacks = append(attacks, squareToString(x, y))
		}
	}

	return attacks
}

func pawnAttacks(white bool, board [8][8]uint8, row, col int) []string {
	attacks := make([]string, 0)
	if white {
		attacks = append(
			attacks,
			squareToString(row-1, col+1),
			squareToString(row-1, col-1),
		)
	} else {
		attacks = append(
			attacks,
			squareToString(row+1, col+1),
			squareToString(row+1, col-1),
		)
	}

	return attacks
}

var knightMoves = [][]int{{-2, -1}, {-1, -2}, {1, -2}, {2, -1}, {2, 1}, {1, 2}, {-1, 2}, {-2, 1}}

func knightAttacks(board [8][8]uint8, row, col int) []string {
	attacks := make([]string, 0)
	for _, move := range knightMoves {
		newRow := row + move[0]
		newCol := col + move[1]

		// Check if the new square is within the board bounds
		if newRow >= 0 && newRow < 8 && newCol >= 0 && newCol < 8 {
			attacks = append(attacks, squareToString(newRow, newCol))
		}
	}

	return attacks
}

func bishopAttacks(board [8][8]uint8, row, col int) []string {
	attacks := make([]string, 0)
	for i, j := row-1, col+1; i >= 0 && j < 8; i, j = i-1, j+1 {
		attacks = append(attacks, squareToString(i, j))

		if board[i][j] != 0 {
			break
		}
	}

	for i, j := row+1, col+1; i < 8 && j < 8; i, j = i+1, j+1 {
		attacks = append(attacks, squareToString(i, j))

		if board[i][j] != 0 {
			break
		}
	}

	for i, j := row+1, col-1; i < 8 && j >= 0; i, j = i+1, j-1 {
		attacks = append(attacks, squareToString(i, j))

		if board[i][j] != 0 {
			break
		}
	}

	for i, j := row-1, col-1; i >= 0 && j >= 0; i, j = i-1, j-1 {
		attacks = append(attacks, squareToString(i, j))

		if board[i][j] != 0 {
			break
		}
	}

	return attacks
}

func rookAttacks(board [8][8]uint8, row, col int) []string {
	attacks := make([]string, 0)
	for r := row - 1; r >= 0; r-- {
		attacks = append(attacks, squareToString(r, col))

		if board[r][col] != 0 {
			break
		}
	}

	for r := row + 1; r < 8; r++ {
		attacks = append(attacks, squareToString(r, col))

		if board[r][col] != 0 {
			break
		}
	}

	for c := col + 1; c < 8; c++ {
		attacks = append(attacks, squareToString(row, c))

		if board[row][c] != 0 {
			break
		}
	}

	for c := col - 1; c >= 0; c-- {
		attacks = append(attacks, squareToString(row, c))

		if board[row][c] != 0 {
			break
		}
	}

	return attacks
}

func queenAttacks(board [8][8]uint8, row, col int) []string {
	attacks := make([]string, 0)
	attacks = append(attacks, bishopAttacks(board, row, col)...)
	attacks = append(attacks, rookAttacks(board, row, col)...)
	return attacks
}
