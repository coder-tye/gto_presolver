package solver

import (
	"math"
	"math/rand"

	"poker-bot/internal/evaluator"
	"poker-bot/internal/models"
)

// SolverEngine encapsulates Linear CFR+ execution for flop subgames
type SolverEngine struct {
	evaluator  evaluator.HandEvaluator
	SimRunouts int // 0 = default (128 for MC), 990 = exact full enumeration
}

func NewSolverEngine(eval evaluator.HandEvaluator) *SolverEngine {
	if eval == nil {
		eval = evaluator.NewChehsunliuEvaluator()
	}
	return &SolverEngine{
		evaluator:  eval,
		SimRunouts: 128, // Default accurate Monte Carlo sample
	}
}

// FlopContext holds the precalculated matchup matrix and indices for a specific board
type FlopContext struct {
	Flop        CanonicalFlop
	OOPCombos   []int       // combo indices in OOP range unblocked by board
	IPCombos    []int       // combo indices in IP range unblocked by board
	Matchup     [][]float64 // [oopIdx][ipIdx] -> equity of OOP vs IP (0..1), -1 if blocked
	OOPIndexMap map[int]int // comboIdx -> local OOP index
	IPIndexMap  map[int]int // comboIdx -> local IP index
}

// BuildFlopContext creates matchup equity matrix between OOP and IP ranges
func (e *SolverEngine) BuildFlopContext(flop CanonicalFlop) *FlopContext {
	board := flop.Cards
	boardCardsMap := make(map[int]bool)
	for _, c := range board {
		boardCardsMap[models.CardIndex(c)] = true
	}

	isBlockedByBoard := func(comboIdx int) bool {
		c1, c2 := models.ComboCards(comboIdx)
		return boardCardsMap[models.CardIndex(c1)] || boardCardsMap[models.CardIndex(c2)]
	}

	rawOOP := GetStandardOOPRange()
	rawIP := GetStandardIPRange()

	var oopCombos []int
	oopMap := make(map[int]int)
	for _, cIdx := range rawOOP {
		if !isBlockedByBoard(cIdx) {
			oopMap[cIdx] = len(oopCombos)
			oopCombos = append(oopCombos, cIdx)
		}
	}

	var ipCombos []int
	ipMap := make(map[int]int)
	for _, cIdx := range rawIP {
		if !isBlockedByBoard(cIdx) {
			ipMap[cIdx] = len(ipCombos)
			ipCombos = append(ipCombos, cIdx)
		}
	}

	// Precompute Matchup Matrix: OOP win probability against IP
	matchup := make([][]float64, len(oopCombos))
	for i := range matchup {
		matchup[i] = make([]float64, len(ipCombos))
	}

	// Deck remaining for runout evaluations
	deadMap := make(map[int]bool)
	for _, c := range board {
		deadMap[models.CardIndex(c)] = true
	}

	deck := make([]models.Card, 0, 49)
	for idx := 0; idx < 52; idx++ {
		if !deadMap[idx] {
			deck = append(deck, models.CardFromIndex(idx))
		}
	}

	simRunouts := e.SimRunouts
	if simRunouts <= 0 {
		simRunouts = 128
	}

	for i, oopIdx := range oopCombos {
		cO1, cO2 := models.ComboCards(oopIdx)
		o1Idx, o2Idx := models.CardIndex(cO1), models.CardIndex(cO2)

		for j, ipIdx := range ipCombos {
			cI1, cI2 := models.ComboCards(ipIdx)
			i1Idx, i2Idx := models.CardIndex(cI1), models.CardIndex(cI2)

			// Check card collision between players
			if o1Idx == i1Idx || o1Idx == i2Idx || o2Idx == i1Idx || o2Idx == i2Idx {
				matchup[i][j] = -1.0 // blocked
				continue
			}

			// Available runout deck without hole cards (45 cards)
			avail := make([]models.Card, 0, len(deck)-4)
			for _, c := range deck {
				ci := models.CardIndex(c)
				if ci != o1Idx && ci != o2Idx && ci != i1Idx && ci != i2Idx {
					avail = append(avail, c)
				}
			}

			var winCount float64
			validSims := 0

			if simRunouts >= 990 {
				// Exact full enumeration of all C(45, 2) = 990 Turn+River runouts
				for r1 := 0; r1 < len(avail); r1++ {
					for r2 := r1 + 1; r2 < len(avail); r2++ {
						fullBoard := []models.Card{board[0], board[1], board[2], avail[r1], avail[r2]}
						sOOP, errO := e.evaluator.Evaluate7([]models.Card{cO1, cO2}, fullBoard)
						sIP, errI := e.evaluator.Evaluate7([]models.Card{cI1, cI2}, fullBoard)
						if errO != nil || errI != nil {
							continue
						}
						if sOOP > sIP {
							winCount += 1.0
						} else if sOOP == sIP {
							winCount += 0.5
						}
						validSims++
					}
				}
			} else {
				// High-precision Monte Carlo with pair-seeded deterministic PRNG
				pairSeed := int64(oopIdx*10007 + ipIdx*31 + 42)
				rng := rand.New(rand.NewSource(pairSeed))
				availLen := len(avail)

				for s := 0; s < simRunouts; s++ {
					r1 := rng.Intn(availLen)
					r2 := rng.Intn(availLen - 1)
					if r2 >= r1 {
						r2++
					}
					fullBoard := []models.Card{board[0], board[1], board[2], avail[r1], avail[r2]}
					sOOP, errO := e.evaluator.Evaluate7([]models.Card{cO1, cO2}, fullBoard)
					sIP, errI := e.evaluator.Evaluate7([]models.Card{cI1, cI2}, fullBoard)
					if errO != nil || errI != nil {
						continue
					}
					if sOOP > sIP {
						winCount += 1.0
					} else if sOOP == sIP {
						winCount += 0.5
					}
					validSims++
				}
			}

			if validSims > 0 {
				matchup[i][j] = winCount / float64(validSims)
			} else {
				matchup[i][j] = 0.5
			}
		}
	}

	return &FlopContext{
		Flop:        flop,
		OOPCombos:   oopCombos,
		IPCombos:    ipCombos,
		Matchup:     matchup,
		OOPIndexMap: oopMap,
		IPIndexMap:  ipMap,
	}
}

// SolveFlop executes Linear CFR+ on the game tree for the given flop
func (e *SolverEngine) SolveFlop(flop CanonicalFlop, iterations int, startingPot, effectiveStack float64) *FlopStrategy {
	fCtx := e.BuildFlopContext(flop)
	numOOP := len(fCtx.OOPCombos)
	numIP := len(fCtx.IPCombos)
	maxCombos := numOOP
	if numIP > maxCombos {
		maxCombos = numIP
	}

	root := BuildFlopTree(startingPot, effectiveStack, maxCombos)

	// Reaching probability vectors
	p0Reach := make([]float64, numOOP)
	for i := range p0Reach {
		p0Reach[i] = 1.0 / float64(numOOP)
	}
	p1Reach := make([]float64, numIP)
	for j := range p1Reach {
		p1Reach[j] = 1.0 / float64(numIP)
	}

	// Linear CFR+ iterations: later iterations are weighted linearly by t
	for t := 1; t <= iterations; t++ {
		e.cfr(root, 0, p0Reach, p1Reach, fCtx, t)
		e.cfr(root, 1, p0Reach, p1Reach, fCtx, t)
	}

	// Extract and normalize average strategy into FlopStrategy
	return e.extractStrategy(flop, root, fCtx, iterations)
}

func (e *SolverEngine) cfr(node *GameNode, updatingPlayer int, p0Reach, p1Reach []float64, fCtx *FlopContext, iteration int) []float64 {
	numOOP := len(fCtx.OOPCombos)
	numIP := len(fCtx.IPCombos)

	// 1. Terminal Fold: Net payoff based strictly on actual chips invested
	if node.Type == NodeTypeTerminalFold {
		oopStartingContrib := node.StartingPot / 2.0
		ipStartingContrib := node.StartingPot / 2.0

		// Winner receives total pot; net payoff = Pot - WinnerTotalCost = LoserTotalCost
		// Loser loses what they contributed = - LoserTotalCost
		// Zero-sum is strictly preserved!
		oopTotalCost := oopStartingContrib + node.OOPInvest
		ipTotalCost := ipStartingContrib + node.IPInvest

		if updatingPlayer == 0 {
			res := make([]float64, numOOP)
			var payoffPerOpp float64
			if node.FoldWinner == 0 {
				payoffPerOpp = ipTotalCost // OOP won: wins IP's investment
			} else {
				payoffPerOpp = -oopTotalCost // OOP lost: loses OOP's investment
			}
			for i := 0; i < numOOP; i++ {
				oppWeight := 0.0
				for j := 0; j < numIP; j++ {
					if fCtx.Matchup[i][j] >= 0 {
						oppWeight += p1Reach[j]
					}
				}
				res[i] = payoffPerOpp * oppWeight
			}
			return res
		} else {
			res := make([]float64, numIP)
			var payoffPerOpp float64
			if node.FoldWinner == 1 {
				payoffPerOpp = oopTotalCost // IP won: wins OOP's investment
			} else {
				payoffPerOpp = -ipTotalCost // IP lost: loses IP's investment
			}
			for j := 0; j < numIP; j++ {
				oppWeight := 0.0
				for i := 0; i < numOOP; i++ {
					if fCtx.Matchup[i][j] >= 0 {
						oppWeight += p0Reach[i]
					}
				}
				res[j] = payoffPerOpp * oppWeight
			}
			return res
		}
	}

	// 2. Terminal Showdown: Net payoff based on actual chips invested and equity
	if node.Type == NodeTypeTerminalShowdown {
		oopTotalCost := node.StartingPot/2.0 + node.OOPInvest
		ipTotalCost := node.StartingPot/2.0 + node.IPInvest
		totalPot := node.Pot

		if updatingPlayer == 0 {
			res := make([]float64, numOOP)
			for i := 0; i < numOOP; i++ {
				ev := 0.0
				for j := 0; j < numIP; j++ {
					eq := fCtx.Matchup[i][j]
					if eq >= 0 {
						// Expected net return = eq * Pot - Cost
						netEV := eq*totalPot - oopTotalCost
						ev += p1Reach[j] * netEV
					}
				}
				res[i] = ev
			}
			return res
		} else {
			res := make([]float64, numIP)
			for j := 0; j < numIP; j++ {
				ev := 0.0
				for i := 0; i < numOOP; i++ {
					eq := fCtx.Matchup[i][j]
					if eq >= 0 {
						// IP win rate is (1.0 - eq)
						ipEq := 1.0 - eq
						netEV := ipEq*totalPot - ipTotalCost
						ev += p0Reach[i] * netEV
					}
				}
				res[j] = ev
			}
			return res
		}
	}

	// 3. Action Node
	curPlayer := node.Player
	numActions := len(node.Actions)
	combosCount := numOOP
	if curPlayer == 1 {
		combosCount = numIP
	}

	// Calculate current strategy using Regret Matching (CFR+)
	strategy := make([][]float64, combosCount)
	for c := 0; c < combosCount; c++ {
		strategy[c] = make([]float64, numActions)
		posSum := 0.0
		for a := 0; a < numActions; a++ {
			r := node.RegretSum[c][a]
			if r > 0 {
				posSum += r
			}
		}
		if posSum > 0 {
			for a := 0; a < numActions; a++ {
				r := node.RegretSum[c][a]
				if r > 0 {
					strategy[c][a] = r / posSum
				}
			}
		} else {
			// Uniform fallback
			prob := 1.0 / float64(numActions)
			for a := 0; a < numActions; a++ {
				strategy[c][a] = prob
			}
		}
	}

	// Compute action child utilities
	actionUtilities := make([][]float64, numActions)
	for a := 0; a < numActions; a++ {
		nextP0 := p0Reach
		nextP1 := p1Reach
		if curPlayer == 0 {
			nextP0 = make([]float64, numOOP)
			for c := 0; c < numOOP; c++ {
				nextP0[c] = p0Reach[c] * strategy[c][a]
			}
		} else {
			nextP1 = make([]float64, numIP)
			for c := 0; c < numIP; c++ {
				nextP1[c] = p1Reach[c] * strategy[c][a]
			}
		}
		actionUtilities[a] = e.cfr(node.Children[a], updatingPlayer, nextP0, nextP1, fCtx, iteration)
	}

	// Compute Node Expected Utility
	targetCombos := numOOP
	if updatingPlayer == 1 {
		targetCombos = numIP
	}
	nodeUtility := make([]float64, targetCombos)

	if curPlayer == updatingPlayer {
		// Player's own node: expectation weighted by player's current strategy
		for a := 0; a < numActions; a++ {
			for c := 0; c < targetCombos; c++ {
				nodeUtility[c] += strategy[c][a] * actionUtilities[a][c]
			}
		}
	} else {
		// Opponent node:
		// Opponent's strategy probabilities were ALREADY incorporated into the next reach probabilities!
		// The counterfactual value is simply the sum over all opponent action subtrees.
		for a := 0; a < numActions; a++ {
			for c := 0; c < targetCombos; c++ {
				nodeUtility[c] += actionUtilities[a][c]
			}
		}
	}

	// Update Regrets and StrategySum for updatingPlayer
	if curPlayer == updatingPlayer {
		weight := float64(iteration) // Linear CFR+ weight
		for c := 0; c < targetCombos; c++ {
			reach := p0Reach[c]
			if updatingPlayer == 1 {
				reach = p1Reach[c]
			}
			for a := 0; a < numActions; a++ {
				regret := actionUtilities[a][c] - nodeUtility[c]
				// CFR+: clamp negative cumulative regrets to 0
				node.RegretSum[c][a] = math.Max(0, node.RegretSum[c][a]+regret)
				// Linear CFR+: accumulate strategy weighted by t * reach
				node.StrategySum[c][a] += weight * reach * strategy[c][a]
			}
		}
	}

	return nodeUtility
}

func (e *SolverEngine) extractStrategy(flop CanonicalFlop, root *GameNode, fCtx *FlopContext, iterations int) *FlopStrategy {
	res := &FlopStrategy{
		Board:       flop.BoardString(),
		Texture:     flop.Texture,
		Category:    flop.Category,
		Description: flop.Description,
		OOPStrategy: make(map[string]map[string]float64),
		IPStrategy:  make(map[string]map[string]float64),
	}

	// 1. OOP Strategy at Root (Node 1)
	extractNodeStrategy(root, fCtx.OOPCombos, res.OOPStrategy)

	// 2. IP Strategy after OOP check (Node 2: root.Children[0])
	if len(root.Children) > 0 {
		ipCheckNode := root.Children[0]
		if ipCheckNode != nil && ipCheckNode.Type == NodeTypeAction {
			extractNodeStrategy(ipCheckNode, fCtx.IPCombos, res.IPStrategy)
		}
	}

	return res
}

// extractNodeStrategy extracts individual combo strategies AND rigorous weighted averages for 169 hand classes
func extractNodeStrategy(node *GameNode, combos []int, targetMap map[string]map[string]float64) {
	numActions := len(node.Actions)
	if numActions == 0 {
		return
	}

	// Map each hand class (e.g. "AKs", "AA", "72o") to its member local combo indices
	classCombos := make(map[string][]int)

	for localIdx, cIdx := range combos {
		c1, c2 := models.ComboCards(cIdx)
		comboName := c1.String() + c2.String()
		handClass := models.ComboClass(cIdx)
		classCombos[handClass] = append(classCombos[handClass], localIdx)

		dist := make(map[string]float64)
		sum := 0.0
		for a := 0; a < numActions; a++ {
			sum += node.StrategySum[localIdx][a]
		}
		for a, actionName := range node.Actions {
			if sum > 0 {
				dist[actionName] = math.Round((node.StrategySum[localIdx][a]/sum)*10000) / 10000.0
			} else {
				dist[actionName] = math.Round((1.0/float64(numActions))*10000) / 10000.0
			}
		}
		targetMap[comboName] = dist
	}

	// Rigorous 169 Hand Class Aggregation: sum strategy across all member combos
	for handClass, memberIndices := range classCombos {
		actionSums := make(map[string]float64)
		totalSum := 0.0
		for _, localIdx := range memberIndices {
			for a, actionName := range node.Actions {
				val := node.StrategySum[localIdx][a]
				actionSums[actionName] += val
				totalSum += val
			}
		}

		classDist := make(map[string]float64)
		for _, actionName := range node.Actions {
			if totalSum > 0 {
				classDist[actionName] = math.Round((actionSums[actionName]/totalSum)*10000) / 10000.0
			} else {
				classDist[actionName] = math.Round((1.0/float64(numActions))*10000) / 10000.0
			}
		}
		targetMap[handClass] = classDist
	}
}
