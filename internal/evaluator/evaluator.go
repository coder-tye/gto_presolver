package evaluator

import (
	"poker-bot/internal/models"
)

// HandRankCategory represents Texas Hold'em hand tiers
type HandRankCategory int

const (
	HighCard HandRankCategory = iota
	OnePair
	TwoPair
	ThreeOfAKind
	Straight
	Flush
	FullHouse
	FourOfAKind
	StraightFlush
	RoyalFlush
)

// EvaluatedScore represents normalized hand ranking (higher is better)
type EvaluatedScore uint32

// HandEvaluator provides hand assessment interface
// Implementations can be chehsunliu/poker, custom lookup table, or 7-card evaluator
type HandEvaluator interface {
	// Evaluate7 evaluates 2 hole cards + 5 board cards, returning a comparative score
	Evaluate7(hole []models.Card, board []models.Card) (EvaluatedScore, error)
	// Category returns the high-level category of the evaluated hand
	Category(score EvaluatedScore) HandRankCategory
}

// EvaluatorProvider allows pluggable switching of hand evaluator implementations
type EvaluatorProvider interface {
	GetEvaluator() HandEvaluator
}
