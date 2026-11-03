package solver

import "math"

type NodeType int

const (
	NodeTypeAction NodeType = iota
	NodeTypeTerminalFold
	NodeTypeTerminalShowdown
)

// GameNode represents a node in the Flop subgame game tree
type GameNode struct {
	ID          int
	Type        NodeType
	Player      int      // 0 = OOP, 1 = IP
	Actions     []string // e.g. ["check", "bet_33", "bet_67", "allin"] or ["fold", "call", "raise", "allin"]
	Children    []*GameNode
	Pot         float64
	StartingPot float64
	OOPInvest   float64
	IPInvest    float64

	// FoldWinner: if NodeTypeTerminalFold, who won? (0 = OOP, 1 = IP)
	FoldWinner int

	// CFR regret and strategy tables
	// RegretSum: [comboIdx][actionIdx]
	RegretSum [][]float64
	// StrategySum: [comboIdx][actionIdx]
	StrategySum [][]float64
}

// BettingState tracks exact chip investments and remaining stacks along a betting line
type BettingState struct {
	StartingPot   float64
	OOPInvest     float64
	IPInvest      float64
	OOPStackLeft  float64
	IPStackLeft   float64
	CurrentBet    float64 // Highest single-round bet level faced
	LastRaiseSize float64 // Incremental size of last raise for min-raise
	ActingPlayer  int     // 0 = OOP, 1 = IP
	NumBets       int     // 0 = unopened, 1 = bet, 2 = raise
}

// BuildFlopTree constructs the Flop subgame betting tree using the BettingState state machine
func BuildFlopTree(startingPot, effectiveStack float64, numCombos int) *GameNode {
	nodeCount := 0

	makeNode := func(nodeType NodeType, player int, actions []string, startingPot, pot, oopInv, ipInv float64) *GameNode {
		nodeCount++
		node := &GameNode{
			ID:          nodeCount,
			Type:        nodeType,
			Player:      player,
			Actions:     actions,
			Pot:         pot,
			StartingPot: startingPot,
			OOPInvest:   oopInv,
			IPInvest:    ipInv,
		}
		if nodeType == NodeTypeAction && len(actions) > 0 {
			node.RegretSum = make([][]float64, numCombos)
			node.StrategySum = make([][]float64, numCombos)
			for i := 0; i < numCombos; i++ {
				node.RegretSum[i] = make([]float64, len(actions))
				node.StrategySum[i] = make([]float64, len(actions))
			}
		}
		return node
	}

	makeTerminalFold := func(winner int, startingPot, pot, oopInv, ipInv float64) *GameNode {
		n := makeNode(NodeTypeTerminalFold, -1, nil, startingPot, pot, oopInv, ipInv)
		n.FoldWinner = winner
		return n
	}

	makeTerminalShowdown := func(startingPot, pot, oopInv, ipInv float64) *GameNode {
		return makeNode(NodeTypeTerminalShowdown, -1, nil, startingPot, pot, oopInv, ipInv)
	}

	var buildNode func(state BettingState) *GameNode
	buildNode = func(state BettingState) *GameNode {
		currentPot := state.StartingPot + state.OOPInvest + state.IPInvest
		actor := state.ActingPlayer
		defender := actor
		bettor := 1 - actor

		var defenderInvest, bettorInvest, defenderStack, bettorStack float64
		if defender == 0 {
			defenderInvest = state.OOPInvest
			bettorInvest = state.IPInvest
			defenderStack = state.OOPStackLeft
			bettorStack = state.IPStackLeft
		} else {
			defenderInvest = state.IPInvest
			bettorInvest = state.OOPInvest
			defenderStack = state.IPStackLeft
			bettorStack = state.OOPStackLeft
		}

		// 1. CurrentBet == 0: Unopened pot (Check or Bet options)
		if state.CurrentBet == 0 {
			var actions []string
			// Actor can check
			actions = append(actions, "check")

			// Sized bets: only if stack allows
			b33 := math.Round(state.StartingPot*0.33*10) / 10.0
			if b33 <= 0 {
				b33 = 1.0
			}
			if b33 < defenderStack {
				actions = append(actions, "bet_33")
			}
			b67 := math.Round(state.StartingPot*0.67*10) / 10.0
			if b67 <= 0 {
				b67 = 2.0
			}
			if b67 < defenderStack && b67 > b33 {
				actions = append(actions, "bet_67")
			}
			if defenderStack > 0 {
				actions = append(actions, "allin")
			}

			node := makeNode(NodeTypeAction, actor, actions, state.StartingPot, currentPot, state.OOPInvest, state.IPInvest)

			for _, act := range actions {
				switch act {
				case "check":
					if actor == 0 {
						// OOP checks -> IP's turn
						nextState := state
						nextState.ActingPlayer = 1
						node.Children = append(node.Children, buildNode(nextState))
					} else {
						// IP checks behind -> Showdown!
						node.Children = append(node.Children, makeTerminalShowdown(state.StartingPot, currentPot, state.OOPInvest, state.IPInvest))
					}

				case "bet_33":
					nextState := state
					if actor == 0 {
						nextState.OOPInvest += b33
						nextState.OOPStackLeft -= b33
					} else {
						nextState.IPInvest += b33
						nextState.IPStackLeft -= b33
					}
					nextState.CurrentBet = b33
					nextState.LastRaiseSize = b33
					nextState.ActingPlayer = bettor
					nextState.NumBets = 1
					node.Children = append(node.Children, buildNode(nextState))

				case "bet_67":
					nextState := state
					if actor == 0 {
						nextState.OOPInvest += b67
						nextState.OOPStackLeft -= b67
					} else {
						nextState.IPInvest += b67
						nextState.IPStackLeft -= b67
					}
					nextState.CurrentBet = b67
					nextState.LastRaiseSize = b67
					nextState.ActingPlayer = bettor
					nextState.NumBets = 1
					node.Children = append(node.Children, buildNode(nextState))

				case "allin":
					nextState := state
					shoveSize := defenderStack
					if actor == 0 {
						nextState.OOPInvest += shoveSize
						nextState.OOPStackLeft = 0
					} else {
						nextState.IPInvest += shoveSize
						nextState.IPStackLeft = 0
					}
					nextState.CurrentBet = shoveSize
					nextState.LastRaiseSize = shoveSize
					nextState.ActingPlayer = bettor
					nextState.NumBets = 1
					node.Children = append(node.Children, buildNode(nextState))
				}
			}
			return node
		}

		// 2. CurrentBet > 0: Facing a bet or raise
		toCall := bettorInvest - defenderInvest
		if toCall > defenderStack {
			toCall = defenderStack
		}

		// Determine legal actions
		var actions []string
		actions = append(actions, "fold", "call")

		canRaise := false
		var raiseSize float64

		// Opponent must have chips left and defender must have more than toCall to raise
		if bettorStack > 0 && defenderStack > toCall && state.NumBets < 2 {
			// Sized raise: 2.5x of previous bet or min-raise
			minRaise := toCall + math.Max(state.LastRaiseSize, toCall)
			targetRaise := toCall + math.Max(state.LastRaiseSize*1.5, toCall*1.5)
			if targetRaise < minRaise {
				targetRaise = minRaise
			}
			if targetRaise < defenderStack && (targetRaise-toCall) < bettorStack {
				canRaise = true
				raiseSize = targetRaise
			}
		}

		if canRaise {
			actions = append(actions, "raise")
		}

		// All-in is available if defender has more than toCall and bettor has chips left
		if defenderStack > toCall && bettorStack > 0 {
			actions = append(actions, "allin")
		}

		node := makeNode(NodeTypeAction, actor, actions, state.StartingPot, currentPot, state.OOPInvest, state.IPInvest)

		for _, act := range actions {
			switch act {
			case "fold":
				// Bettor wins the current pot; loser's loss is strictly what was invested
				node.Children = append(node.Children, makeTerminalFold(bettor, state.StartingPot, currentPot, state.OOPInvest, state.IPInvest))

			case "call":
				// Defender calls toCall; pot is matched, round ends in Showdown
				callOOPInv := state.OOPInvest
				callIPInv := state.IPInvest
				if defender == 0 {
					callOOPInv += toCall
				} else {
					callIPInv += toCall
				}
				finalPot := state.StartingPot + callOOPInv + callIPInv
				node.Children = append(node.Children, makeTerminalShowdown(state.StartingPot, finalPot, callOOPInv, callIPInv))

			case "raise":
				nextState := state
				incrRaise := raiseSize - toCall
				if defender == 0 {
					nextState.OOPInvest += raiseSize
					nextState.OOPStackLeft -= raiseSize
				} else {
					nextState.IPInvest += raiseSize
					nextState.IPStackLeft -= raiseSize
				}
				nextState.CurrentBet = defenderInvest + raiseSize
				nextState.LastRaiseSize = incrRaise
				nextState.ActingPlayer = bettor
				nextState.NumBets = state.NumBets + 1
				node.Children = append(node.Children, buildNode(nextState))

			case "allin":
				nextState := state
				shoveAdd := defenderStack
				incrRaise := shoveAdd - toCall
				if defender == 0 {
					nextState.OOPInvest += shoveAdd
					nextState.OOPStackLeft = 0
				} else {
					nextState.IPInvest += shoveAdd
					nextState.IPStackLeft = 0
				}
				nextState.CurrentBet = defenderInvest + shoveAdd
				nextState.LastRaiseSize = incrRaise
				nextState.ActingPlayer = bettor
				nextState.NumBets = state.NumBets + 1
				node.Children = append(node.Children, buildNode(nextState))
			}
		}

		return node
	}

	initialState := BettingState{
		StartingPot:   startingPot,
		OOPInvest:     0,
		IPInvest:      0,
		OOPStackLeft:  effectiveStack,
		IPStackLeft:   effectiveStack,
		CurrentBet:    0,
		LastRaiseSize: 0,
		ActingPlayer:  0,
		NumBets:       0,
	}

	return buildNode(initialState)
}
