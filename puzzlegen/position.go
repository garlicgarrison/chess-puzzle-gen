package puzzlegen

import (
	"errors"
	"fmt"
	"math"
	"math/rand"
	"strconv"
	"strings"
	"unicode"
)

const NonKingPieces = "PNBRQpnbrq"
const NoiseSTD = 0.5

var (
	ErrInvalidPuzzleConfig = errors.New("pieces must add up to 30")
	ErrInvalidFEN          = errors.New("invalid fen")
)

var PieceToBit = map[rune]int8{
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

var BitToPiece = map[int8]rune{
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

var StartPieces = map[rune]int{
	'P': 8,
	'N': 2,
	'B': 2,
	'R': 2,
	'Q': 1,
	'K': 1,
	'p': 8,
	'n': 2,
	'b': 2,
	'r': 2,
	'q': 1,
	'k': 1,
}

var castleSquares = map[rune][][]int8{
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
	WhiteQ int8 `yaml:"white_q"`
	WhiteR int8 `yaml:"white_r"`
	WhiteB int8 `yaml:"white_b"`
	WhiteN int8 `yaml:"white_n"`
	WhiteP int8 `yaml:"white_p"`
	BlackQ int8 `yaml:"black_q"`
	BlackR int8 `yaml:"black_r"`
	BlackB int8 `yaml:"black_b"`
	BlackN int8 `yaml:"black_n"`
	BlackP int8 `yaml:"black_p"`
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

	board := [8][8]int8{}
	pieceMap := map[rune]int8{
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

	whiteAttacks := make(map[int8]bool)
	blackAttacks := make(map[int8]bool)

	for piece, num := range pieceMap {
		for i := int8(0); i < num; i++ {
			for {
				var pRow, pCol int8
				switch piece {
				case 'P', 'p':
					pRow, pCol = int8(rand.Intn(6)+1), int8(rand.Intn(8))
				default:
					pRow, pCol = int8(rand.Intn(8)), int8(rand.Intn(8))
				}

				if board[pRow][pCol] == 0 {
					board[pRow][pCol] = PieceToBit[piece]
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

			piece := BitToPiece[val]
			attacks := attacks(piece, board, int8(i), int8(j))
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
		pRow, pCol := int8(rand.Intn(8)), int8(rand.Intn(8))

		if board[pRow][pCol] == 0 && !blackAttacks[squareHash(pRow, pCol)] {
			board[pRow][pCol] = PieceToBit['K']
			attacks := kingAttacks(board, pRow, pCol)
			for _, a := range attacks {
				whiteAttacks[a] = true
			}
			break
		}
	}

	// Add black king
	for {
		pRow, pCol := int8(rand.Intn(8)), int8(rand.Intn(8))

		if board[pRow][pCol] == 0 && !whiteAttacks[squareHash(pRow, pCol)] {
			board[pRow][pCol] = PieceToBit['k']
			break
		}
	}

	// Randomly choose side -- 0 for black 1 for white
	var sb strings.Builder
	player := int8(rand.Intn(2))
	writeFEN(&sb, player, board, blackAttacks, whiteAttacks)

	return sb.String(), nil
}

/*
	We can mutate a string by first adding the non-king pieces first, and then adding/removing
	one of those pieces, and then adding the kings with enpassant/casting rights
	We have to assume the fen is a valid position to start with

	asymptote is the number of pieces we want to converge to
*/
func MutateFEN(fen string, asymptote int) (string, error) {
	if asymptote > 30 {
		return "", ErrInvalidFEN
	}

	board := [8][8]int8{}
	fen = strings.TrimSpace(fen)
	parts := strings.Split(fen, " ")

	startMap := make(map[rune]int)
	for key, value := range StartPieces {
		startMap[key] = value
	}

	var blackK int8
	var whiteK int8

	rankStrs := strings.Split(parts[0], "/")
	totalPieces := 0

	for rank, row := range rankStrs {
		var file int8
		for _, p := range row {
			if p == 'k' {
				blackK = int8(rank)*8 + file
				continue
			}
			if p == 'K' {
				whiteK = int8(rank)*8 + file
				continue
			}

			pieceBit, ok := PieceToBit[p]
			if !ok {
				skip, err := strconv.ParseInt(string(p), 10, 8)
				if err != nil {
					return "", ErrInvalidFEN
				}

				file += int8(skip)
				continue
			}

			board[rank][file] = pieceBit
			startMap[p]--
			file++
			totalPieces++
		}
	}

	noise := rand.NormFloat64()*NoiseSTD + 1
	toAdd := int(math.Floor((float64(asymptote) - float64(totalPieces)) * noise))
	if toAdd < -1*totalPieces {
		toAdd = 1
	} else if toAdd+totalPieces > 30 {
		toAdd = 30 - totalPieces
	}

	pieceOperations := int(math.Abs(float64(toAdd)))
	randRow := rand.Intn(8)
	randCol := rand.Intn(8)
	if toAdd == 0 {
		for {
			randomPiece := rune(NonKingPieces[rand.Intn(8)])
			pieceToRemove := board[randRow][randCol]
			if (randomPiece == 'P' || randomPiece == 'p') && (randRow == 7 || randRow == 0) ||
				board[randRow][randCol] == 0 ||
				PieceToBit[randomPiece] == pieceToRemove {
				randRow = rand.Intn(8)
				randCol = rand.Intn(8)
				continue
			}
			startMap[BitToPiece[pieceToRemove]]++

			for {
				randRow = rand.Intn(8)
				randCol = rand.Intn(8)
				startMap[randomPiece]--
				if board[randRow][randCol] == 0 &&
					!((randomPiece == 'P' || randomPiece == 'p') && (randRow == 7 || randRow == 0)) {
					board[randRow][randCol] = PieceToBit[randomPiece]
					break
				}
			}

			break
		}
	}

	for i := 0; i < pieceOperations; i++ {
		if toAdd > 0 {
			for {
				randomPiece := rune(NonKingPieces[rand.Intn(8)])
				if (randomPiece == 'P' || randomPiece == 'p') && (randRow == 7 || randRow == 0) ||
					startMap[randomPiece] == 0 ||
					board[randRow][randCol] != 0 {
					randRow = rand.Intn(8)
					randCol = rand.Intn(8)
					continue
				}

				board[randRow][randCol] = PieceToBit[randomPiece]
				startMap[randomPiece]--
				break
			}
		} else if toAdd < 0 {
			for {
				if board[randRow][randCol] != 0 {
					pieceToRemove := board[randRow][randCol]
					startMap[BitToPiece[pieceToRemove]]++
					board[randRow][randCol] = 0
					break
				}

				randRow = rand.Intn(8)
				randCol = rand.Intn(8)
			}
		}
	}

	// Add attacks
	whiteAttacks := make(map[int8]bool)
	blackAttacks := make(map[int8]bool)
	for i, row := range board {
		for j, val := range row {
			if val == 0 {
				continue
			}

			piece := BitToPiece[val]
			attacks := attacks(piece, board, int8(i), int8(j))
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
	placementAttempted := false
	for {
		var pRow, pCol int8
		if !placementAttempted {
			pRow, pCol = whiteK/8, whiteK%8
			placementAttempted = true
		} else {
			pRow, pCol = int8(rand.Intn(8)), int8(rand.Intn(8))
		}

		if board[pRow][pCol] == 0 && !blackAttacks[squareHash(pRow, pCol)] {
			board[pRow][pCol] = PieceToBit['K']
			attacks := kingAttacks(board, pRow, pCol)
			for _, a := range attacks {
				whiteAttacks[a] = true
			}
			break
		}
	}

	// Add black king
	placementAttempted = false
	for {
		var pRow, pCol int8
		if !placementAttempted {
			pRow, pCol = blackK/8, blackK%8
			placementAttempted = true
		} else {
			pRow, pCol = int8(rand.Intn(8)), int8(rand.Intn(8))
		}

		if board[pRow][pCol] == 0 && !whiteAttacks[squareHash(pRow, pCol)] {
			board[pRow][pCol] = PieceToBit['k']
			break
		}
	}

	// Pieces of FEN
	var sb strings.Builder
	var player int8
	if parts[1] == "w" {
		player = 1
	}
	writeFEN(&sb, player, board, blackAttacks, whiteAttacks)

	return sb.String(), nil
}

func squareHash(row, col int8) int8 {
	return row*8 + col
}

func attacks(piece rune, board [8][8]int8, row, col int8) []int8 {
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

func kingAttacks(board [8][8]int8, row, col int8) []int8 {
	attacks := make([]int8, 0)
	for dx := -1; dx <= 1; dx++ {
		for dy := -1; dy <= 1; dy++ {
			if dx == 0 && dy == 0 {
				continue
			}

			x, y := row+int8(dx), col+int8(dy)
			if x < 0 || x > 7 || y < 0 || y > 7 {
				continue
			}

			attacks = append(attacks, squareHash(x, y))
		}
	}

	return attacks
}

func pawnAttacks(white bool, board [8][8]int8, row, col int8) []int8 {
	attacks := make([]int8, 0)
	if white {
		attacks = append(
			attacks,
			squareHash(row-1, col+1),
			squareHash(row-1, col-1),
		)
	} else {
		attacks = append(
			attacks,
			squareHash(row+1, col+1),
			squareHash(row+1, col-1),
		)
	}

	return attacks
}

var knightMoves = [][]int8{{-2, -1}, {-1, -2}, {1, -2}, {2, -1}, {2, 1}, {1, 2}, {-1, 2}, {-2, 1}}

func knightAttacks(board [8][8]int8, row, col int8) []int8 {
	attacks := make([]int8, 0)
	for _, move := range knightMoves {
		newRow := int8(int8(row) + move[0])
		newCol := int8(int8(col) + move[1])

		// Check if the new square is within the board bounds
		if newRow >= 0 && newRow < 8 && newCol >= 0 && newCol < 8 {
			attacks = append(attacks, squareHash(newRow, newCol))
		}
	}

	return attacks
}

func bishopAttacks(board [8][8]int8, row, col int8) []int8 {
	attacks := make([]int8, 0)
	for i, j := row-1, col+1; i >= 0 && j < 8; i, j = i-1, j+1 {
		attacks = append(attacks, squareHash(i, j))

		if board[i][j] != 0 {
			break
		}
	}

	for i, j := row+1, col+1; i < 8 && j < 8; i, j = i+1, j+1 {
		attacks = append(attacks, squareHash(i, j))

		if board[i][j] != 0 {
			break
		}
	}

	for i, j := row+1, col-1; i < 8 && j >= 0; i, j = i+1, j-1 {
		attacks = append(attacks, squareHash(i, j))

		if board[i][j] != 0 {
			break
		}
	}

	for i, j := row-1, col-1; i >= 0 && j >= 0; i, j = i-1, j-1 {
		attacks = append(attacks, squareHash(i, j))

		if board[i][j] != 0 {
			break
		}
	}

	return attacks
}

func rookAttacks(board [8][8]int8, row, col int8) []int8 {
	attacks := make([]int8, 0)
	for r := row - 1; r >= 0; r-- {
		attacks = append(attacks, squareHash(r, col))

		if board[r][col] != 0 {
			break
		}
	}

	for r := row + 1; r < 8; r++ {
		attacks = append(attacks, squareHash(r, col))

		if board[r][col] != 0 {
			break
		}
	}

	for c := col + 1; c < 8; c++ {
		attacks = append(attacks, squareHash(row, c))

		if board[row][c] != 0 {
			break
		}
	}

	for c := col - 1; c >= 0; c-- {
		attacks = append(attacks, squareHash(row, c))

		if board[row][c] != 0 {
			break
		}
	}

	return attacks
}

func queenAttacks(board [8][8]int8, row, col int8) []int8 {
	attacks := make([]int8, 0)
	attacks = append(attacks, bishopAttacks(board, row, col)...)
	attacks = append(attacks, rookAttacks(board, row, col)...)
	return attacks
}

func writeFEN(sb *strings.Builder, player int8, board [8][8]int8, blackAttacks, whiteAttacks map[int8]bool) {
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
			sb.WriteRune(BitToPiece[val])
			empty = 0
		}

		if empty != 0 {
			sb.WriteString(fmt.Sprintf("%d", empty))
		}
		if i != 7 {
			sb.WriteRune('/')
		}
	}

	if player == 0 {
		sb.WriteString(" b")
	} else {
		sb.WriteString(" w")
	}

	sb.WriteRune(' ')
	if board[7][4] != 'K' && board[0][4] != 'k' {
		sb.WriteRune('-')
		goto EnPassant
	}

	for e, squares := range castleSquares {
		attackFound := false
		for _, s := range squares {
			var check map[int8]bool
			if unicode.IsUpper(e) {
				check = blackAttacks
			} else {
				check = whiteAttacks
			}

			if check[squareHash(s[0], s[1])] {
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
	if player == 0 && board[3][eSquare] == PieceToBit['p'] &&
		board[2][eSquare]+board[1][eSquare] == 0 {
		sb.WriteRune(rune(eSquare + 97))
		sb.WriteString(fmt.Sprintf("%d", 6))
	} else if player == 1 &&
		board[4][eSquare] == PieceToBit['P'] &&
		board[5][eSquare]+board[6][eSquare] == 0 {
		sb.WriteRune(rune(eSquare + 97))
		sb.WriteString(fmt.Sprintf("%d", 3))
	} else {
		sb.WriteRune('-')
	}

	sb.WriteString(" 0 ")
	sb.WriteRune('1')
}
