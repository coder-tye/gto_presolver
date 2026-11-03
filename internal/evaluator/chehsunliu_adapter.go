package evaluator

import (
	"fmt"

	"github.com/chehsunliu/poker"
	"poker-bot/internal/models"
)

// ChehsunliuEvaluator wraps github.com/chehsunliu/poker
// In chehsunliu/poker, Evaluate returns rank between 1 (Royal Flush) and 7462 (worst 7-high).
// LOWER score means STRONGER hand in Cactus Kev's representation!
// To satisfy EvaluatedScore where HIGHER is STRONGER, we invert the score: Score = 8000 - rank
type ChehsunliuEvaluator struct{}

func NewChehsunliuEvaluator() *ChehsunliuEvaluator {
	return &ChehsunliuEvaluator{}
}

// Evaluate7 evaluates 2 hole cards + 5 board cards
func (e *ChehsunliuEvaluator) Evaluate7(hole []models.Card, board []models.Card) (EvaluatedScore, error) {
	if len(hole) != 2 {
		return 0, fmt.Errorf("hole cards must be 2, got %d", len(hole))
	}
	if len(board) != 5 {
		return 0, fmt.Errorf("board cards must be 5, got %d", len(board))
	}

	pokerCards := make([]poker.Card, 7)
	pokerCards[0] = poker.NewCard(hole[0].String())
	pokerCards[1] = poker.NewCard(hole[1].String())
	for i := 0; i < 5; i++ {
		pokerCards[2+i] = poker.NewCard(board[i].String())
	}

	rawRank := poker.Evaluate(pokerCards)
	if rawRank <= 0 {
		return 0, fmt.Errorf("invalid evaluation result: %d", rawRank)
	}

	// 7462 is the lowest possible hand (7-5-4-3-2 unsuited)
	// We map 1 (Royal Flush) -> 7462, and 7462 (7-high) -> 1
	normalized := EvaluatedScore(7500 - rawRank)
	return normalized, nil
}

// Category converts the evaluated score back to HandRankCategory
func (e *ChehsunliuEvaluator) Category(score EvaluatedScore) HandRankCategory {
	rawRank := int32(7500 - uint32(score))
	rankClass := poker.RankClass(rawRank)
	// rankClass from chehsunliu/poker:
	// 1: Straight Flush / Royal Flush
	// 2: Four of a Kind
	// 3: Full House
	// 4: Flush
	// 5: Straight
	// 6: Three of a Kind
	// 7: Two Pair
	// 8: One Pair
	// 9: High Card
	switch rankClass {
	case 1:
		if rawRank == 1 {
			return RoyalFlush
		}
		return StraightFlush
	case 2:
		return FourOfAKind
	case 3:
		return FullHouse
	case 4:
		return Flush
	case 5:
		return Straight
	case 6:
		return ThreeOfAKind
	case 7:
		return TwoPair
	case 8:
		return OnePair
	default:
		return HighCard
	}
}

// DefaultEvaluatorProvider implements EvaluatorProvider
type DefaultEvaluatorProvider struct {
	evaluator HandEvaluator
}

func NewDefaultEvaluatorProvider() *DefaultEvaluatorProvider {
	return &DefaultEvaluatorProvider{
		evaluator: NewChehsunliuEvaluator(),
	}
}

func (p *DefaultEvaluatorProvider) GetEvaluator() HandEvaluator {
	return p.evaluator
}
