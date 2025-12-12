package routines

import (
	"runtime"

	"github.com/jaysongiroux/mdserve/internal/logger"
)

const (
	minWorkersLimit = 4
	maxWorkersLimit = 50
)

// calculateMaxWorkers determines the optimal number of concurrent workers
func CalculateMaxWorkers(assetCount int) int {
	// Get number of CPU cores
	numCPU := runtime.NumCPU()

	// Base calculation: 2x CPU cores for I/O-bound tasks
	maxWorkers := min(
		max(
			numCPU*2,
			minWorkersLimit,
		),
		maxWorkersLimit,
	)

	// If we have fewer assets than workers, reduce workers
	if assetCount < maxWorkers {
		return assetCount
	}

	logger.Debug("Using %d concurrent workers", maxWorkers)
	return maxWorkers
}
