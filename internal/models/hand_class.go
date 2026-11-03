package models

import (
	"strings"
)

// RankOrder maps a Rank to its numeric value 2..14
var RankOrder = map[Rank]int{
	Rank2: 2,
	Rank3: 3,
	Rank4: 4,
	Rank5: 5,
	Rank6: 6,
	Rank7: 7,
	Rank8: 8,
	Rank9: 9,
	RankT: 10,
	RankJ: 11,
	RankQ: 12,
	RankK: 13,
	RankA: 14,
}

// RankChars contains the character representations of ranks in ascending order
var RankChars = []byte{'2', '3', '4', '5', '6', '7', '8', '9', 'T', 'J', 'Q', 'K', 'A'}

var suitOrder = map[Suit]int{
	SuitSpades:   0,
	SuitHearts:   1,
	SuitDiamonds: 2,
	SuitClubs:    3,
}

var suitChars = []Suit{SuitSpades, SuitHearts, SuitDiamonds, SuitClubs}

const (
	TotalCombos      = 1326
	TotalHandClasses = 169
)

var (
	cardPairToCombo [52][52]int
	comboCards      [TotalCombos][2]Card
	comboToClass    [TotalCombos]string
	classToCombos   map[string][]int
)

func init() {
	classToCombos = make(map[string][]int, TotalHandClasses)
	idx := 0
	for a := 0; a < 52; a++ {
		c1 := CardFromIndex(a)
		for b := a + 1; b < 52; b++ {
			c2 := CardFromIndex(b)
			cardPairToCombo[a][b] = idx
			cardPairToCombo[b][a] = idx
			comboCards[idx] = [2]Card{c1, c2}
			cls := CanonicalHandClass(c1, c2)
			comboToClass[idx] = cls
			classToCombos[cls] = append(classToCombos[cls], idx)
			idx++
		}
	}
}

// CardIndex returns 0..51 index for a Card (rank * 4 + suit)
func CardIndex(c Card) int {
	rVal := RankOrder[c.Rank] - 2 // 0..12
	sVal := suitOrder[c.Suit]    // 0..3
	if rVal < 0 || rVal > 12 || sVal < 0 || sVal > 3 {
		return 0
	}
	return rVal*4 + sVal
}

// CardFromIndex returns the Card for index 0..51
func CardFromIndex(idx int) Card {
	if idx < 0 || idx >= 52 {
		return Card{Rank: Rank2, Suit: SuitSpades}
	}
	rIdx := idx / 4
	sIdx := idx % 4
	return Card{
		Rank: Rank(RankChars[rIdx]),
		Suit: suitChars[sIdx],
	}
}

// ComboIndex returns the 0..1325 unique combo index for any two distinct cards
func ComboIndex(c1, c2 Card) int {
	a := CardIndex(c1)
	b := CardIndex(c2)
	if a == b {
		return -1
	}
	return cardPairToCombo[a][b]
}

// ComboCards returns the two cards corresponding to a combo index 0..1325
func ComboCards(comboIdx int) (Card, Card) {
	if comboIdx < 0 || comboIdx >= TotalCombos {
		return Card{}, Card{}
	}
	return comboCards[comboIdx][0], comboCards[comboIdx][1]
}

// ComboClass returns the canonical hand class (e.g. "AA", "AKs", "AKo") for a combo index
func ComboClass(comboIdx int) string {
	if comboIdx < 0 || comboIdx >= TotalCombos {
		return ""
	}
	return comboToClass[comboIdx]
}

// ClassCombos returns the combo indices belonging to a hand class
func ClassCombos(handClass string) []int {
	return classToCombos[handClass]
}

// CanonicalHandClass converts two cards into one of the 169 canonical poker hand classes
// e.g. "AA", "AKs", "AKo", "72o", "87s"
func CanonicalHandClass(c1, c2 Card) string {
	r1Val := RankOrder[c1.Rank]
	r2Val := RankOrder[c2.Rank]

	if r1Val == r2Val {
		return string([]byte{byte(c1.Rank), byte(c2.Rank)})
	}

	high := c1
	low := c2
	if r2Val > r1Val {
		high = c2
		low = c1
	}

	if high.Suit == low.Suit {
		return string([]byte{byte(high.Rank), byte(low.Rank), 's'})
	}
	return string([]byte{byte(high.Rank), byte(low.Rank), 'o'})
}

// ExpandHandRange parses standard poker range shorthand expressions into a set of 169 classes
// Supports: "99+", "22+", "ATs+", "KTs+", "87s+", "ATo+", "KTo+", and individual classes "AKs, A5s"
func ExpandHandRange(expr string) map[string]bool {
	result := make(map[string]bool)
	tokens := strings.Split(expr, ",")

	for _, rawToken := range tokens {
		token := strings.TrimSpace(rawToken)
		if token == "" {
			continue
		}

		if strings.HasSuffix(token, "+") {
			base := strings.TrimSuffix(token, "+")
			// Pair plus, e.g., "99+", "TT+"
			if len(base) == 2 && base[0] == base[1] {
				startRank := Rank(base[0])
				startIdx := RankOrder[startRank] - 2 // rank2 is index 0
				for i := startIdx; i < len(RankChars); i++ {
					r := RankChars[i]
					result[string([]byte{r, r})] = true
				}
				continue
			}

			// Suited or Offsuit plus, e.g., "ATs+", "87s+", "ATo+"
			if len(base) == 3 {
				rHigh := Rank(base[0])
				rLow := Rank(base[1])
				suffix := base[2] // 's' or 'o'

				highIdx := RankOrder[rHigh] - 2
				lowIdx := RankOrder[rLow] - 2

				// Suited connector plus e.g. "54s+", "87s+"
				if highIdx == lowIdx+1 && suffix == 's' {
					for i := lowIdx; i < len(RankChars)-1; i++ {
						rL := RankChars[i]
						rH := RankChars[i+1]
						result[string([]byte{rH, rL, 's'})] = true
					}
					continue
				}

				// Broadway plus e.g. "ATs+" -> ATs, AJs, AQs, AKs
				for i := lowIdx; i < highIdx; i++ {
					rL := RankChars[i]
					result[string([]byte{byte(rHigh), rL, suffix})] = true
				}
				continue
			}
		}

		// Exact hand class, e.g., "AKs", "AA", "72o"
		result[token] = true
	}

	return result
}
