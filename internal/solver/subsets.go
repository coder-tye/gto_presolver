package solver

import (
	"strings"

	"poker-bot/internal/models"
)

// CanonicalFlop represents a representative flop texture for GTO pre-solving
type CanonicalFlop struct {
	ID          string        `json:"id"`
	Cards       []models.Card `json:"cards"`
	Texture     string        `json:"texture"`     // rainbow, two_tone, monotone, paired, trips
	Category    string        `json:"category"`    // A-high, Broadway, Medium, Low
	Description string        `json:"description"`
}

func (f *CanonicalFlop) BoardString() string {
	var sb strings.Builder
	for _, c := range f.Cards {
		sb.WriteString(c.String())
	}
	return sb.String()
}

// Default25FlopSubsets contains 25 canonical boards representing the 1,755 flop isomorphism classes
var Default25FlopSubsets = []CanonicalFlop{
	// 1. A-high boards
	{
		ID:          "flop_01_A_dry_rainbow",
		Cards:       parseCardsHelper("As", "Kd", "7c"),
		Texture:     "rainbow",
		Category:    "A-high",
		Description: "A-K-7 rainbow (dry broadway)",
	},
	{
		ID:          "flop_02_A_two_tone",
		Cards:       parseCardsHelper("Ah", "Qh", "4c"),
		Texture:     "two_tone",
		Category:    "A-high",
		Description: "A-Q-4 two-tone (semi-wet)",
	},
	{
		ID:          "flop_03_A_monotone",
		Cards:       parseCardsHelper("As", "7s", "2s"),
		Texture:     "monotone",
		Category:    "A-high",
		Description: "A-7-2 monotone (flush board)",
	},
	{
		ID:          "flop_04_A_connected_wet",
		Cards:       parseCardsHelper("Ad", "Jc", "Tc"),
		Texture:     "two_tone",
		Category:    "A-high",
		Description: "A-J-T connected wet",
	},
	{
		ID:          "flop_05_A_dry_disconnected",
		Cards:       parseCardsHelper("As", "9d", "4c"),
		Texture:     "rainbow",
		Category:    "A-high",
		Description: "A-9-4 dry disconnected",
	},

	// 2. K-high & Q-high boards
	{
		ID:          "flop_06_broadway_rainbow",
		Cards:       parseCardsHelper("Ks", "Qd", "Jc"),
		Texture:     "rainbow",
		Category:    "Broadway",
		Description: "K-Q-J broadway rainbow",
	},
	{
		ID:          "flop_07_K_dry_rainbow",
		Cards:       parseCardsHelper("Kh", "8c", "2d"),
		Texture:     "rainbow",
		Category:    "Broadway",
		Description: "K-8-2 dry rainbow",
	},
	{
		ID:          "flop_08_K_two_tone",
		Cards:       parseCardsHelper("Ks", "Th", "4s"),
		Texture:     "two_tone",
		Category:    "Broadway",
		Description: "K-T-4 two-tone",
	},
	{
		ID:          "flop_09_K_monotone",
		Cards:       parseCardsHelper("Kh", "9h", "3h"),
		Texture:     "monotone",
		Category:    "Broadway",
		Description: "K-9-3 monotone",
	},
	{
		ID:          "flop_10_Q_connected",
		Cards:       parseCardsHelper("Qs", "Jd", "7c"),
		Texture:     "rainbow",
		Category:    "Broadway",
		Description: "Q-J-7 rainbow",
	},
	{
		ID:          "flop_11_Q_disconnected",
		Cards:       parseCardsHelper("Qh", "9s", "2d"),
		Texture:     "rainbow",
		Category:    "Broadway",
		Description: "Q-9-2 dry",
	},

	// 3. Medium & Heavy Connected boards
	{
		ID:          "flop_12_heavy_wet_straight",
		Cards:       parseCardsHelper("Jh", "Ts", "9c"),
		Texture:     "rainbow",
		Category:    "Medium",
		Description: "J-T-9 heavy straight board",
	},
	{
		ID:          "flop_13_J_dry",
		Cards:       parseCardsHelper("Js", "7d", "2c"),
		Texture:     "rainbow",
		Category:    "Medium",
		Description: "J-7-2 dry disconnected",
	},
	{
		ID:          "flop_14_T_two_tone_connected",
		Cards:       parseCardsHelper("Th", "8h", "6c"),
		Texture:     "two_tone",
		Category:    "Medium",
		Description: "T-8-6 gapped two-tone",
	},
	{
		ID:          "flop_15_T_connected",
		Cards:       parseCardsHelper("Ts", "9d", "4c"),
		Texture:     "rainbow",
		Category:    "Medium",
		Description: "T-9-4 connected rainbow",
	},
	{
		ID:          "flop_16_9_8_7_rainbow",
		Cards:       parseCardsHelper("9s", "8d", "7c"),
		Texture:     "rainbow",
		Category:    "Medium",
		Description: "9-8-7 coordinated rainbow",
	},
	{
		ID:          "flop_17_8_monotone_straight",
		Cards:       parseCardsHelper("8h", "7h", "6h"),
		Texture:     "monotone",
		Category:    "Medium",
		Description: "8-7-6 monotone straight",
	},
	{
		ID:          "flop_18_8_low_dry",
		Cards:       parseCardsHelper("8s", "4d", "2c"),
		Texture:     "rainbow",
		Category:    "Medium",
		Description: "8-4-2 dry low",
	},

	// 4. Low boards
	{
		ID:          "flop_19_7_6_5_rainbow",
		Cards:       parseCardsHelper("7s", "6s", "5c"),
		Texture:     "two_tone",
		Category:    "Low",
		Description: "7-6-5 connected two-tone",
	},
	{
		ID:          "flop_20_6_5_4_monotone",
		Cards:       parseCardsHelper("6h", "5h", "4h"),
		Texture:     "monotone",
		Category:    "Low",
		Description: "6-5-4 monotone straight",
	},

	// 5. Paired & Trips boards
	{
		ID:          "flop_21_paired_AA_K",
		Cards:       parseCardsHelper("As", "Ad", "Kc"),
		Texture:     "paired",
		Category:    "Paired",
		Description: "A-A-K paired high",
	},
	{
		ID:          "flop_22_paired_KK_7",
		Cards:       parseCardsHelper("Ks", "Kd", "7c"),
		Texture:     "paired",
		Category:    "Paired",
		Description: "K-K-7 paired broadway",
	},
	{
		ID:          "flop_23_paired_88_3",
		Cards:       parseCardsHelper("8s", "8d", "3c"),
		Texture:     "paired",
		Category:    "Paired",
		Description: "8-8-3 paired medium",
	},
	{
		ID:          "flop_24_paired_22_7",
		Cards:       parseCardsHelper("2s", "2d", "7c"),
		Texture:     "paired",
		Category:    "Paired",
		Description: "2-2-7 paired low",
	},
	{
		ID:          "flop_25_trips_333",
		Cards:       parseCardsHelper("3s", "3d", "3c"),
		Texture:     "trips",
		Category:    "Trips",
		Description: "3-3-3 trips board",
	},
}

func parseCardsHelper(strs ...string) []models.Card {
	cards, _ := models.ParseCards(strs)
	return cards
}

// GetStandardIPRange returns standard BTN single-raised-pot opening range (~45% of combos)
func GetStandardIPRange() []int {
	classes := []string{
		// Pocket pairs: all
		"AA", "KK", "QQ", "JJ", "TT", "99", "88", "77", "66", "55", "44", "33", "22",
		// Suited Ax: all
		"AKs", "AQs", "AJs", "ATs", "A9s", "A8s", "A7s", "A6s", "A5s", "A4s", "A3s", "A2s",
		// Offsuit Ax: A7o+
		"AKo", "AQo", "AJo", "ATo", "A9o", "A8o", "A7o",
		// Suited Kx: K2s+
		"KQs", "KJs", "KTs", "K9s", "K8s", "K7s", "K6s", "K5s", "K4s", "K3s", "K2s",
		// Offsuit Kx: K9o+
		"KQo", "KJo", "KTo", "K9o",
		// Suited Qx: Q4s+
		"QJs", "QTs", "Q9s", "Q8s", "Q7s", "Q6s", "Q5s", "Q4s",
		// Offsuit Qx: Q9o+
		"QJo", "QTo", "Q9o",
		// Suited Jx: J6s+
		"JTs", "J9s", "J8s", "J7s", "J6s",
		// Offsuit Jx: J9o+
		"JTo", "J9o",
		// Suited connectors & one-gappers
		"T9s", "T8s", "T7s", "T6s",
		"98s", "97s", "96s",
		"87s", "86s",
		"76s", "75s",
		"65s", "64s",
		"54s", "53s",
		"43s",
	}

	return collectCombos(classes)
}

// GetStandardOOPRange returns standard BB defense range vs BTN 2.5bb open (~55% of combos)
func GetStandardOOPRange() []int {
	classes := []string{
		// Defensive pairs (TT-22, top pairs usually 3bet)
		"TT", "99", "88", "77", "66", "55", "44", "33", "22",
		// Suited Ax
		"A9s", "A8s", "A7s", "A6s", "A5s", "A4s", "A3s", "A2s",
		// Offsuit Ax: A2o - ATo
		"ATo", "A9o", "A8o", "A7o", "A6o", "A5o", "A4o", "A3o", "A2o",
		// Suited Kx: K2s - KTs
		"KTs", "K9s", "K8s", "K7s", "K6s", "K5s", "K4s", "K3s", "K2s",
		// Offsuit Kx: K6o - KJo
		"KJo", "KTo", "K9o", "K8o", "K7o", "K6o",
		// Suited Qx: Q2s - QTs
		"QTs", "Q9s", "Q8s", "Q7s", "Q6s", "Q5s", "Q4s", "Q3s", "Q2s",
		// Offsuit Qx: Q7o - QTo
		"QTo", "Q9o", "Q8o", "Q7o",
		// Suited Jx: J2s - JTs
		"JTs", "J9s", "J8s", "J7s", "J6s", "J5s", "J4s", "J3s", "J2s",
		// Offsuit Jx: J7o - JTo
		"JTo", "J9o", "J8o", "J7o",
		// Connectors & gappers
		"T9s", "T8s", "T7s", "T6s", "T5s",
		"T9o", "T8o", "T7o",
		"98s", "97s", "96s", "95s",
		"98o", "97o",
		"87s", "86s", "85s", "84s",
		"87o", "86o",
		"76s", "75s", "74s",
		"76o", "75o",
		"65s", "64s", "63s",
		"65o", "64o",
		"54s", "53s", "52s",
		"43s", "32s",
	}

	return collectCombos(classes)
}

func collectCombos(classes []string) []int {
	seen := make(map[int]bool)
	var result []int
	for _, cls := range classes {
		combos := models.ClassCombos(cls)
		for _, cIdx := range combos {
			if !seen[cIdx] {
				seen[cIdx] = true
				result = append(result, cIdx)
			}
		}
	}
	return result
}

