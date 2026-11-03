package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"poker-bot/internal/evaluator"
	"poker-bot/internal/solver"
)

func main() {
	iterations := flag.Int("iterations", 300, "Number of CFR+ iterations per canonical flop")
	workers := flag.Int("workers", runtime.NumCPU(), "Number of parallel worker goroutines")
	outFile := flag.String("out", "gto_flop_table.json", "Output JSON path for solved GTO tables")
	boardLimit := flag.Int("boards", 25, "Number of canonical flop boards to solve (1..25)")
	startingPot := flag.Float64("pot", 5.5, "Starting pot size (in BB, default 5.5 for SRP)")
	effStack := flag.Float64("stack", 97.5, "Effective stack size (in BB, default 97.5)")
	version := flag.String("version", "gto-v1-subsets25", "Version tag for the GTO database")
	runouts := flag.Int("runouts", 256, "Number of Turn+River runouts for equity (128..990, 990=exact)")
	flag.Parse()

	if *workers < 1 {
		*workers = 1
	}
	if *boardLimit > len(solver.Default25FlopSubsets) {
		*boardLimit = len(solver.Default25FlopSubsets)
	}

	selectedBoards := solver.Default25FlopSubsets[:*boardLimit]

	fmt.Println("==================================================================")
	fmt.Println("       iPoker GTO Flop Offline Solver (CFR+ Pre-solver)          ")
	fmt.Println("==================================================================")
	fmt.Printf(" [Config] Canonical Boards: %d\n", len(selectedBoards))
	fmt.Printf(" [Config] CFR Iterations:   %d per board\n", *iterations)
	fmt.Printf(" [Config] Parallel Workers: %d (CPU Cores: %d)\n", *workers, runtime.NumCPU())
	fmt.Printf(" [Config] Subgame Tree:     Pot=%.1f BB, Stack=%.1f BB\n", *startingPot, *effStack)
	fmt.Printf(" [Config] Runout Precision: %d runouts (990 = exact)\n", *runouts)
	fmt.Printf(" [Config] Output Target:    %s\n", *outFile)
	fmt.Println("------------------------------------------------------------------")

	table := solver.NewGTOFlopTable(*version, *iterations)
	eval := evaluator.NewChehsunliuEvaluator()
	engine := solver.NewSolverEngine(eval)
	engine.SimRunouts = *runouts

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	taskCh := make(chan solver.CanonicalFlop, len(selectedBoards))
	for _, b := range selectedBoards {
		taskCh <- b
	}
	close(taskCh)

	var completed int64
	totalBoards := int64(len(selectedBoards))
	startTime := time.Now()

	var wg sync.WaitGroup
	for w := 1; w <= *workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case <-ctx.Done():
					return
				case flop, ok := <-taskCh:
					if !ok {
						return
					}
					bStart := time.Now()
					log.Printf("[Worker %d] Starting solving board %s (%s, %s)...",
						workerID, flop.BoardString(), flop.Category, flop.Texture)

					strat := engine.SolveFlop(flop, *iterations, *startingPot, *effStack)
					table.AddStrategy(strat)

					done := atomic.AddInt64(&completed, 1)
					elapsed := time.Since(startTime)
					bElapsed := time.Since(bStart)
					avgPerBoard := elapsed / time.Duration(done)
					remaining := avgPerBoard * time.Duration(totalBoards-done)

					log.Printf("[Worker %d] ✓ Finished %s in %v | Progress: %d/%d (%.1f%%) | ETA: %v",
						workerID, flop.BoardString(), bElapsed.Round(time.Millisecond),
						done, totalBoards, float64(done)/float64(totalBoards)*100, remaining.Round(time.Second))
				}
			}
		}(w)
	}

	wg.Wait()

	totalElapsed := time.Since(startTime)
	fmt.Println("------------------------------------------------------------------")
	log.Printf("Solving completed for %d boards in %v. Saving output...", completed, totalElapsed.Round(time.Millisecond))

	if err := table.SaveToFile(*outFile); err != nil {
		log.Fatalf("Failed to save GTO table to %s: %v", *outFile, err)
	}

	fileInfo, err := os.Stat(*outFile)
	fileSizeKB := int64(0)
	if err == nil {
		fileSizeKB = fileInfo.Size() / 1024
	}

	fmt.Printf("\nSUCCESS: GTO Flop Database successfully saved to: %s (%d KB)\n", *outFile, fileSizeKB)
	fmt.Println("This file can be placed into poker_bot/data/ or server configs for Phase 5 O(1) in-memory lookup.")
	fmt.Println("==================================================================")
}

