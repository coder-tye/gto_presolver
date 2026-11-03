package solver

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"poker-bot/internal/models"
)

// FlopStrategy contains converged GTO strategies for a single flop board
type FlopStrategy struct {
	Board       string                        `json:"board"`
	Texture     string                        `json:"texture"`
	Category    string                        `json:"category"`
	Description string                        `json:"description"`
	OOPStrategy map[string]map[string]float64 `json:"oop_strategy"` // combo or class -> action -> freq
	IPStrategy  map[string]map[string]float64 `json:"ip_strategy"`  // combo or class -> action -> freq
}

// GTOFlopTable holds the complete precomputed database of canonical flop solutions
type GTOFlopTable struct {
	mu          sync.RWMutex
	Version     string                   `json:"version"`
	GeneratedAt string                   `json:"generated_at"`
	Iterations  int                      `json:"iterations"`
	Flops       map[string]*FlopStrategy `json:"flops"` // board string -> strategy
}

func NewGTOFlopTable(version string, iterations int) *GTOFlopTable {
	if version == "" {
		version = "gto-v1-subsets25"
	}
	return &GTOFlopTable{
		Version:     version,
		GeneratedAt: time.Now().UTC().Format(time.RFC3339),
		Iterations:  iterations,
		Flops:       make(map[string]*FlopStrategy),
	}
}

// AddStrategy registers a solved flop strategy into the table
func (t *GTOFlopTable) AddStrategy(s *FlopStrategy) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.Flops[s.Board] = s
}

// SaveToFile exports the GTO table to a JSON file
func (t *GTOFlopTable) SaveToFile(path string) error {
	t.mu.RLock()
	defer t.mu.RUnlock()

	data, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal gto table: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// LoadFromFile loads a pre-solved GTO table from disk
func LoadGTOFlopTable(path string) (*GTOFlopTable, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read gto file: %w", err)
	}

	var table GTOFlopTable
	if err := json.Unmarshal(data, &table); err != nil {
		return nil, fmt.Errorf("unmarshal gto table: %w", err)
	}
	return &table, nil
}

func (t *GTOFlopTable) LookupIP(boardStr string, isIP bool, comboOrClass string) (map[string]float64, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	cleanBoard := strings.TrimSpace(boardStr)
	flop, exists := t.Flops[cleanBoard]
	if !exists {
		return nil, false
	}

	var strategyMap map[string]map[string]float64
	if isIP {
		strategyMap = flop.IPStrategy
	} else {
		strategyMap = flop.OOPStrategy
	}

	if strategyMap == nil {
		return nil, false
	}

	dist, hit := strategyMap[comboOrClass]
	if !hit && len(comboOrClass) == 4 {
		// Try reversed card order e.g. "KdAc" -> "AcKd"
		rev := comboOrClass[2:4] + comboOrClass[0:2]
		dist, hit = strategyMap[rev]
	}
	return dist, hit
}

// Lookup queries the GTO action frequencies for a given board, position, and combo/hand class
func (t *GTOFlopTable) Lookup(boardStr string, position string, comboOrClass string) (map[string]float64, bool) {
	isIP := strings.EqualFold(position, "IP") || (!strings.EqualFold(position, "OOP") && (strings.EqualFold(position, "BTN") || strings.EqualFold(position, "CO")))
	return t.LookupIP(boardStr, isIP, comboOrClass)
}

var (
	globalGTOTable     *GTOFlopTable
	globalGTOTableOnce sync.Once
)

// GetGlobalGTOTable returns the globally loaded GTO flop table singleton
func GetGlobalGTOTable() *GTOFlopTable {
	globalGTOTableOnce.Do(func() {
		paths := []string{
			os.Getenv("GTO_TABLE_PATH"),
			"gto_flop_table.json",
			"data/gto_flop_table.json",
			"../gto_flop_table.json",
			"../data/gto_flop_table.json",
			"../../gto_flop_table.json",
			"../../data/gto_flop_table.json",
		}
		for _, p := range paths {
			if p == "" {
				continue
			}
			if tbl, err := LoadGTOFlopTable(p); err == nil && len(tbl.Flops) > 0 {
				globalGTOTable = tbl
				break
			}
		}
	})
	return globalGTOTable
}

// SetGlobalGTOTable allows injecting a GTO table (e.g. for testing)
func SetGlobalGTOTable(t *GTOFlopTable) {
	globalGTOTable = t
}

// FindClosestBoard finds the best matching canonical flop board from the pre-solved dataset
func (t *GTOFlopTable) FindClosestBoard(boardCards []models.Card) (string, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if len(boardCards) < 3 || len(t.Flops) == 0 {
		return "", false
	}

	// Exact match check
	var sb strings.Builder
	for i := 0; i < 3; i++ {
		sb.WriteString(boardCards[i].String())
	}
	exactStr := sb.String()
	if _, ok := t.Flops[exactStr]; ok {
		return exactStr, true
	}

	// Deterministic selection: collect and sort keys
	keys := make([]string, 0, len(t.Flops))
	for k := range t.Flops {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Feature match: board category and texture
	targetCategory := categorizeBoardCategory(boardCards[:3])
	targetTexture := categorizeTexture(boardCards[:3])
	targetMaxRank := getMaxRank(boardCards[:3])

	// First pass: match both category and texture, choosing closest high card
	var bestBoth string
	bestBothDiff := 999
	for _, bStr := range keys {
		s := t.Flops[bStr]
		if s.Category == targetCategory && s.Texture == targetTexture {
			diff := absDiff(getBoardMaxRank(bStr), targetMaxRank)
			if diff < bestBothDiff {
				bestBothDiff = diff
				bestBoth = bStr
			}
		}
	}
	if bestBoth != "" {
		return bestBoth, true
	}

	// Second pass: match category, choosing closest high card
	var bestCat string
	bestCatDiff := 999
	for _, bStr := range keys {
		s := t.Flops[bStr]
		if s.Category == targetCategory {
			diff := absDiff(getBoardMaxRank(bStr), targetMaxRank)
			if diff < bestCatDiff {
				bestCatDiff = diff
				bestCat = bStr
			}
		}
	}
	if bestCat != "" {
		return bestCat, true
	}

	// If no category matched with confidence, return false to trigger clean fallback to Monte Carlo EV engine
	return "", false
}

func absDiff(a, b int) int {
	if a > b {
		return a - b
	}
	return b - a
}

func getBoardMaxRank(bStr string) int {
	maxR := 0
	for i := 0; i+1 < len(bStr); i += 2 {
		r := models.RankOrder[models.Rank(bStr[i])]
		if r > maxR {
			maxR = r
		}
	}
	return maxR
}

func getMaxRank(cards []models.Card) int {
	maxRank := 0
	for _, c := range cards {
		r := models.RankOrder[c.Rank]
		if r > maxRank {
			maxRank = r
		}
	}
	return maxRank
}

func categorizeBoardCategory(cards []models.Card) string {
	if len(cards) < 3 {
		return "Low"
	}
	ranks := make(map[models.Rank]int)
	maxRank := 0
	for _, c := range cards[:3] {
		ranks[c.Rank]++
		r := models.RankOrder[c.Rank]
		if r > maxRank {
			maxRank = r
		}
	}
	if len(ranks) == 1 {
		return "Trips"
	}
	if len(ranks) == 2 {
		return "Paired"
	}
	switch {
	case maxRank == 14:
		return "A-high"
	case maxRank >= 11:
		return "Broadway"
	case maxRank >= 8:
		return "Medium"
	default:
		return "Low"
	}
}

func categorizeTexture(cards []models.Card) string {
	suits := make(map[models.Suit]int)
	ranks := make(map[models.Rank]int)
	for _, c := range cards {
		suits[c.Suit]++
		ranks[c.Rank]++
	}

	if len(ranks) == 1 {
		return "trips"
	}
	if len(ranks) == 2 {
		return "paired"
	}
	if len(suits) == 1 {
		return "monotone"
	}
	if len(suits) == 2 {
		return "two_tone"
	}
	return "rainbow"
}

