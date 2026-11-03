package models

import (
	"fmt"
	"strings"
)

// Rank represents a card rank (2..9, T, J, Q, K, A)
type Rank byte

const (
	Rank2 Rank = '2'
	Rank3 Rank = '3'
	Rank4 Rank = '4'
	Rank5 Rank = '5'
	Rank6 Rank = '6'
	Rank7 Rank = '7'
	Rank8 Rank = '8'
	Rank9 Rank = '9'
	RankT Rank = 'T'
	RankJ Rank = 'J'
	RankQ Rank = 'Q'
	RankK Rank = 'K'
	RankA Rank = 'A'
)

// Suit represents card suits (s, h, d, c)
type Suit byte

const (
	SuitSpades   Suit = 's'
	SuitHearts   Suit = 'h'
	SuitDiamonds Suit = 'd'
	SuitClubs    Suit = 'c'
)

// Card represents a standardized playing card
type Card struct {
	Rank Rank
	Suit Suit
}

// String returns the canonical notation e.g., "As", "Td", "7h"
func (c Card) String() string {
	return string([]byte{byte(c.Rank), byte(c.Suit)})
}

// NormalizedCardString normalizes any representation (e.g. "as" -> "As", "10s" -> "Ts")
func NormalizeCardString(raw string) (string, error) {
	c, err := ParseCard(raw)
	if err != nil {
		return "", err
	}
	return c.String(), nil
}

// ParseCard parses standard string notations such as "As", "as", "10s", "Ts"
func ParseCard(raw string) (Card, error) {
	s := strings.TrimSpace(raw)
	if len(s) < 2 {
		return Card{}, fmt.Errorf("invalid card string: %q", raw)
	}

	var rChar byte
	var sChar byte

	// Special handle "10s", "10h", etc.
	if strings.HasPrefix(s, "10") {
		if len(s) != 3 {
			return Card{}, fmt.Errorf("invalid card string length for 10: %q", raw)
		}
		rChar = 'T'
		sChar = s[2]
	} else {
		if len(s) != 2 {
			return Card{}, fmt.Errorf("invalid card string length: %q", raw)
		}
		rChar = s[0]
		sChar = s[1]
	}

	// Normalize Rank
	var rank Rank
	switch rChar {
	case '2', '3', '4', '5', '6', '7', '8', '9':
		rank = Rank(rChar)
	case 't', 'T':
		rank = RankT
	case 'j', 'J':
		rank = RankJ
	case 'q', 'Q':
		rank = RankQ
	case 'k', 'K':
		rank = RankK
	case 'a', 'A':
		rank = RankA
	default:
		return Card{}, fmt.Errorf("unknown rank in card %q", raw)
	}

	// Normalize Suit
	var suit Suit
	switch strings.ToLower(string(sChar)) {
	case "s":
		suit = SuitSpades
	case "h":
		suit = SuitHearts
	case "d":
		suit = SuitDiamonds
	case "c":
		suit = SuitClubs
	default:
		return Card{}, fmt.Errorf("unknown suit in card %q", raw)
	}

	return Card{Rank: rank, Suit: suit}, nil
}

// ParseCards parses a slice of card strings
func ParseCards(rawCards []string) ([]Card, error) {
	res := make([]Card, len(rawCards))
	seen := make(map[string]bool)
	for i, raw := range rawCards {
		c, err := ParseCard(raw)
		if err != nil {
			return nil, err
		}
		str := c.String()
		if seen[str] {
			return nil, fmt.Errorf("duplicate card detected: %s", str)
		}
		seen[str] = true
		res[i] = c
	}
	return res, nil
}
