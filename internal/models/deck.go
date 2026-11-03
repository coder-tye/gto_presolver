package models

// All52Cards returns all 52 standard cards in a single slice
func All52Cards() []Card {
	ranks := []Rank{Rank2, Rank3, Rank4, Rank5, Rank6, Rank7, Rank8, Rank9, RankT, RankJ, RankQ, RankK, RankA}
	suits := []Suit{SuitSpades, SuitHearts, SuitDiamonds, SuitClubs}
	deck := make([]Card, 0, 52)
	for _, r := range ranks {
		for _, s := range suits {
			deck = append(deck, Card{Rank: r, Suit: s})
		}
	}
	return deck
}

// AvailableDeck returns the remaining deck after excluding dead/known cards
func AvailableDeck(deadCards []Card) []Card {
	deadMap := make(map[string]bool, len(deadCards))
	for _, c := range deadCards {
		deadMap[c.String()] = true
	}

	all := All52Cards()
	available := make([]Card, 0, 52-len(deadCards))
	for _, c := range all {
		if !deadMap[c.String()] {
			available = append(available, c)
		}
	}
	return available
}
