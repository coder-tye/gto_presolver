package models

import "fmt"

// ActionType defines the legal action enum
type ActionType string

const (
	ActionFold  ActionType = "fold"
	ActionCheck ActionType = "check"
	ActionCall  ActionType = "call"
	ActionBet   ActionType = "bet"
	ActionRaise ActionType = "raise"
	ActionAllIn ActionType = "all_in"
)

// Street represents the current stage of Texas Hold'em
type Street string

const (
	StreetPreflop Street = "preflop"
	StreetFlop    Street = "flop"
	StreetTurn    Street = "turn"
	StreetRiver   Street = "river"
)

// Position represents standard table positions
type Position string

const (
	PositionSB  Position = "SB"
	PositionBB  Position = "BB"
	PositionUTG Position = "UTG"
	PositionMP  Position = "MP"
	PositionCO  Position = "CO"
	PositionBTN Position = "BTN"
)

// IsValidAction checks if an action string is a known ActionType
func IsValidAction(action string) bool {
	switch ActionType(action) {
	case ActionFold, ActionCheck, ActionCall, ActionBet, ActionRaise, ActionAllIn:
		return true
	default:
		return false
	}
}

// ValidateActionCompatibility verifies if the chosen action is in legal actions
func ValidateActionCompatibility(chosen ActionType, legalActions []ActionType) error {
	for _, a := range legalActions {
		if a == chosen {
			return nil
		}
	}
	return fmt.Errorf("action %s is not among legal actions %v", chosen, legalActions)
}
