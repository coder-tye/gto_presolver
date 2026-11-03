package evaluator

import (
	"testing"

	"poker-bot/internal/models"
)

func TestChehsunliuEvaluator(t *testing.T) {
	eval := NewChehsunliuEvaluator()

	// Royal Flush: As Ks with Qs Js Ts
	royalHole, _ := models.ParseCards([]string{"As", "Ks"})
	royalBoard, _ := models.ParseCards([]string{"Qs", "Js", "Ts", "2d", "3c"})
	royalScore, err := eval.Evaluate7(royalHole, royalBoard)
	if err != nil {
		t.Fatalf("failed to eval royal flush: %v", err)
	}
	if eval.Category(royalScore) != RoyalFlush {
		t.Errorf("expected RoyalFlush category, got %v", eval.Category(royalScore))
	}

	// High Card: 7s 5c with 4d 3h 2s
	worstHole, _ := models.ParseCards([]string{"7s", "5c"})
	worstBoard, _ := models.ParseCards([]string{"4d", "3h", "2s", "8h", "Td"})
	worstScore, err := eval.Evaluate7(worstHole, worstBoard)
	if err != nil {
		t.Fatalf("failed to eval high card: %v", err)
	}

	// Royal Flush score must be strictly greater than High Card
	if royalScore <= worstScore {
		t.Errorf("royalScore %d should be greater than worstScore %d", royalScore, worstScore)
	}

	// Full House vs Flush
	// Full house: As Ah with Ad Ks Kd
	fhHole, _ := models.ParseCards([]string{"As", "Ah"})
	fhBoard, _ := models.ParseCards([]string{"Ad", "Ks", "Kd", "2c", "3c"})
	fhScore, _ := eval.Evaluate7(fhHole, fhBoard)

	// Flush: 2h 4h with 6h 8h Kh
	flushHole, _ := models.ParseCards([]string{"2h", "4h"})
	flushBoard, _ := models.ParseCards([]string{"6h", "8h", "Kh", "As", "Qd"})
	flushScore, _ := eval.Evaluate7(flushHole, flushBoard)

	if fhScore <= flushScore {
		t.Errorf("Full House (%d) must beat Flush (%d)", fhScore, flushScore)
	}
	if eval.Category(fhScore) != FullHouse {
		t.Errorf("expected FullHouse category, got %v", eval.Category(fhScore))
	}
	if eval.Category(flushScore) != Flush {
		t.Errorf("expected Flush category, got %v", eval.Category(flushScore))
	}
}

func TestEvaluatorOrdinalConsistency(t *testing.T) {
	eval := NewChehsunliuEvaluator()

	// 1. Straight Flush (Steel Wheel: 5h-4h-3h-2h-Ah)
	sfHole, _ := models.ParseCards([]string{"5h", "4h"})
	sfBoard, _ := models.ParseCards([]string{"3h", "2h", "Ah", "Kd", "Qc"})
	sfScore, _ := eval.Evaluate7(sfHole, sfBoard)

	// 2. Four of a Kind (9s 9h 9d 9c Kd)
	quadHole, _ := models.ParseCards([]string{"9s", "9h"})
	quadBoard, _ := models.ParseCards([]string{"9d", "9c", "Kd", "2s", "3c"})
	quadScore, _ := eval.Evaluate7(quadHole, quadBoard)

	// 3. Full House (Jh Jd Js 8c 8d)
	fhHole, _ := models.ParseCards([]string{"Jh", "Jd"})
	fhBoard, _ := models.ParseCards([]string{"Js", "8c", "8d", "2c", "3c"})
	fhScore, _ := eval.Evaluate7(fhHole, fhBoard)

	// 4. Flush (Kh Jh 8h 4h 2h)
	flHole, _ := models.ParseCards([]string{"Kh", "Jh"})
	flBoard, _ := models.ParseCards([]string{"8h", "4h", "2h", "As", "Qd"})
	flScore, _ := eval.Evaluate7(flHole, flBoard)

	// 5. Straight (Ts 9h 8d 7c 6s)
	strHole, _ := models.ParseCards([]string{"Ts", "9h"})
	strBoard, _ := models.ParseCards([]string{"8d", "7c", "6s", "2h", "3d"})
	strScore, _ := eval.Evaluate7(strHole, strBoard)

	// 6. Wheel Straight (Ah 2s 3d 4c 5s)
	wheelHole, _ := models.ParseCards([]string{"Ah", "2s"})
	wheelBoard, _ := models.ParseCards([]string{"3d", "4c", "5s", "8h", "Kd"})
	wheelScore, _ := eval.Evaluate7(wheelHole, wheelBoard)

	// 7. Three of a Kind (7s 7h 7d Ac Kd)
	tripsHole, _ := models.ParseCards([]string{"7s", "7h"})
	tripsBoard, _ := models.ParseCards([]string{"7d", "Ac", "Kd", "2c", "3c"})
	tripsScore, _ := eval.Evaluate7(tripsHole, tripsBoard)

	// 8. Two Pair (As Ah Ks Kh Qd)
	tp1Hole, _ := models.ParseCards([]string{"As", "Ah"})
	tp1Board, _ := models.ParseCards([]string{"Ks", "Kh", "Qd", "2c", "3c"})
	tp1Score, _ := eval.Evaluate7(tp1Hole, tp1Board)

	// 9. Two Pair lower kicker (As Ah Ks Kh Jd)
	tp2Hole, _ := models.ParseCards([]string{"As", "Ah"})
	tp2Board, _ := models.ParseCards([]string{"Ks", "Kh", "Jd", "2c", "3c"})
	tp2Score, _ := eval.Evaluate7(tp2Hole, tp2Board)

	// 10. One Pair (As Ah Kd Qc Jd)
	pairHole, _ := models.ParseCards([]string{"As", "Ah"})
	pairBoard, _ := models.ParseCards([]string{"Kd", "Qc", "Jd", "2c", "3c"})
	pairScore, _ := eval.Evaluate7(pairHole, pairBoard)

	// 11. High Card (As Kd Qc Jd 9s)
	hcHole, _ := models.ParseCards([]string{"As", "Kd"})
	hcBoard, _ := models.ParseCards([]string{"Qc", "Jd", "9s", "2c", "3c"})
	hcScore, _ := eval.Evaluate7(hcHole, hcBoard)

	// Assert strict descending order (Hierarchy Check)
	if !(sfScore > quadScore &&
		quadScore > fhScore &&
		fhScore > flScore &&
		flScore > strScore &&
		strScore > wheelScore &&
		wheelScore > tripsScore &&
		tripsScore > tp1Score &&
		tp1Score > tp2Score && // Kicker difference check: Queen kicker > Jack kicker
		tp2Score > pairScore &&
		pairScore > hcScore) {
		t.Fatalf("hierarchy check failed: scores not strictly descending")
	}

	// Board play tie check
	boardPlayHole1, _ := models.ParseCards([]string{"2s", "3s"})
	boardPlayHole2, _ := models.ParseCards([]string{"2c", "3c"})
	broadwayBoard, _ := models.ParseCards([]string{"As", "Ks", "Qs", "Js", "Ts"}) // Royal flush on board
	score1, _ := eval.Evaluate7(boardPlayHole1, broadwayBoard)
	score2, _ := eval.Evaluate7(boardPlayHole2, broadwayBoard)
	if score1 != score2 {
		t.Errorf("board play tie check failed: expected %d == %d", score1, score2)
	}
}

