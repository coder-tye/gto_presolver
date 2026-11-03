package solver

import (
	"os"
	"path/filepath"
	"testing"

	"poker-bot/internal/evaluator"
	"poker-bot/internal/models"
)

func TestCanonicalFlopSubsets(t *testing.T) {
	if len(Default25FlopSubsets) != 25 {
		t.Fatalf("expected 25 canonical flop subsets, got %d", len(Default25FlopSubsets))
	}

	for _, flop := range Default25FlopSubsets {
		if len(flop.Cards) != 3 {
			t.Errorf("expected 3 cards for flop %s, got %d", flop.ID, len(flop.Cards))
		}
		if flop.BoardString() == "" {
			t.Errorf("empty board string for flop %s", flop.ID)
		}
	}
}

func TestBuildFlopTree(t *testing.T) {
	root := BuildFlopTree(5.5, 97.5, 100)
	if root == nil {
		t.Fatal("expected root node to be non-nil")
	}
	if root.Type != NodeTypeAction || root.Player != 0 {
		t.Errorf("root should be action node for OOP (player 0)")
	}
	if len(root.Actions) != 4 {
		t.Errorf("expected 4 root actions (check, bet_33, bet_67, allin), got %d", len(root.Actions))
	}
	if len(root.Children) != 4 {
		t.Errorf("expected 4 root children, got %d", len(root.Children))
	}
}

func TestSolverEngine_SolveFlop(t *testing.T) {
	eval := evaluator.NewChehsunliuEvaluator()
	engine := NewSolverEngine(eval)

	testFlop := Default25FlopSubsets[0] // As, Kd, 7c
	iterations := 10                    // Fast iteration test

	strat := engine.SolveFlop(testFlop, iterations, 5.5, 97.5)
	if strat == nil {
		t.Fatal("expected non-nil FlopStrategy")
	}
	if strat.Board != testFlop.BoardString() {
		t.Errorf("expected board %s, got %s", testFlop.BoardString(), strat.Board)
	}

	// Verify OOP Strategy was populated
	if len(strat.OOPStrategy) == 0 {
		t.Errorf("expected OOPStrategy to contain entries")
	}

	// Verify IP Strategy was populated
	if len(strat.IPStrategy) == 0 {
		t.Errorf("expected IPStrategy to contain entries")
	}

	// AA should have high checking or betting frequencies summing to 1.0
	if dist, ok := strat.OOPStrategy["AA"]; ok {
		sum := 0.0
		for _, f := range dist {
			sum += f
		}
		if sum < 0.95 || sum > 1.05 {
			t.Errorf("action frequencies for AA should sum to ~1.0, got %f", sum)
		}
	}
}

func TestGTOFlopTable_SaveLoadLookup(t *testing.T) {
	tmpDir := t.TempDir()
	outPath := filepath.Join(tmpDir, "test_gto_table.json")

	table := NewGTOFlopTable("test-v1", 20)
	strat := &FlopStrategy{
		Board:       "AsKd7c",
		Texture:     "rainbow",
		Category:    "A-high",
		Description: "A-K-7 rainbow test",
		OOPStrategy: map[string]map[string]float64{
			"AA": {
				"check":  0.30,
				"bet_33": 0.70,
			},
			"AsAh": {
				"check":  0.30,
				"bet_33": 0.70,
			},
		},
		IPStrategy: map[string]map[string]float64{
			"KK": {
				"check":  0.20,
				"bet_33": 0.80,
			},
		},
	}
	table.AddStrategy(strat)

	// Save to disk
	if err := table.SaveToFile(outPath); err != nil {
		t.Fatalf("SaveToFile failed: %v", err)
	}

	// Load back
	loaded, err := LoadGTOFlopTable(outPath)
	if err != nil {
		t.Fatalf("LoadGTOFlopTable failed: %v", err)
	}

	// Lookup exact match
	dist, ok := loaded.Lookup("AsKd7c", "OOP", "AA")
	if !ok {
		t.Fatalf("expected AA to be found in OOP lookup")
	}
	if dist["bet_33"] != 0.70 {
		t.Errorf("expected bet_33 freq 0.70, got %f", dist["bet_33"])
	}

	// Lookup closest board
	bCards, _ := models.ParseCards([]string{"Ah", "Ks", "8d"})
	closest, found := loaded.FindClosestBoard(bCards)
	if !found || closest != "AsKd7c" {
		t.Errorf("expected closest board AsKd7c, got %s found=%v", closest, found)
	}

	// Clean up verification
	_ = os.Remove(outPath)
}

func TestLoadedGTOProductionFile(t *testing.T) {
	prodPath := "../../gto_flop_table.json"
	if _, err := os.Stat(prodPath); os.IsNotExist(err) {
		t.Skip("gto_flop_table.json not found in root, skipping production file validation")
	}

	table, err := LoadGTOFlopTable(prodPath)
	if err != nil {
		t.Fatalf("failed to load production GTO table: %v", err)
	}

	if table.Iterations != 2000 {
		t.Errorf("expected 2000 iterations, got %d", table.Iterations)
	}

	if len(table.Flops) != 25 {
		t.Fatalf("expected 25 boards, got %d", len(table.Flops))
	}

	t.Logf("Successfully verified GTO table: %d boards, %d iterations, generated at %s",
		len(table.Flops), table.Iterations, table.GeneratedAt)

	// Test lookups for canonical boards
	testBoards := []string{"AsKd7c", "KsQdJc", "JhTs9c", "2s2d7c"}
	for _, b := range testBoards {
		strat, exists := table.Flops[b]
		if !exists {
			t.Errorf("expected board %s to be present in table", b)
			continue
		}
		if len(strat.OOPStrategy) == 0 || len(strat.IPStrategy) == 0 {
			t.Errorf("expected non-empty strategy for board %s", b)
		}
	}
}

func TestGameTree_NoRaiseFacingAllIn(t *testing.T) {
	root := BuildFlopTree(5.5, 97.5, 10)

	var checkTree func(node *GameNode, lastAction string)
	checkTree = func(node *GameNode, lastAction string) {
		if node == nil {
			return
		}
		if node.Type == NodeTypeAction {
			if lastAction == "allin" {
				for _, a := range node.Actions {
					if a == "raise" {
						t.Errorf("node %d facing allin has illegal action 'raise'", node.ID)
					}
					if a == "allin" {
						t.Errorf("node %d facing allin has duplicate action 'allin'", node.ID)
					}
				}
			}
			for i, child := range node.Children {
				act := node.Actions[i]
				checkTree(child, act)
			}
		}
	}

	checkTree(root, "")
}

func TestCategorizeBoardCategory_PairedAndTrips(t *testing.T) {
	pairedCards, _ := models.ParseCards([]string{"2s", "2d", "7c"})
	if cat := categorizeBoardCategory(pairedCards); cat != "Paired" {
		t.Errorf("expected 2s2d7c category 'Paired', got %s", cat)
	}

	tripsCards, _ := models.ParseCards([]string{"3s", "3d", "3c"})
	if cat := categorizeBoardCategory(tripsCards); cat != "Trips" {
		t.Errorf("expected 3s3d3c category 'Trips', got %s", cat)
	}

	aHighCards, _ := models.ParseCards([]string{"Ah", "Kd", "7s"})
	if cat := categorizeBoardCategory(aHighCards); cat != "A-high" {
		t.Errorf("expected AhKd7s category 'A-high', got %s", cat)
	}

	bwayCards, _ := models.ParseCards([]string{"Kh", "Qd", "Jc"})
	if cat := categorizeBoardCategory(bwayCards); cat != "Broadway" {
		t.Errorf("expected KhQdJc category 'Broadway', got %s", cat)
	}

	medCards, _ := models.ParseCards([]string{"Ts", "9d", "4c"})
	if cat := categorizeBoardCategory(medCards); cat != "Medium" {
		t.Errorf("expected Ts9d4c category 'Medium', got %s", cat)
	}

	lowCards, _ := models.ParseCards([]string{"7s", "6s", "5c"})
	if cat := categorizeBoardCategory(lowCards); cat != "Low" {
		t.Errorf("expected 7s6s5c category 'Low', got %s", cat)
	}
}

