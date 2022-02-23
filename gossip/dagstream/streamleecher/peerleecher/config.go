package peerleecher

import (
	"time"

	"github.com/MugamboBC/mugambo-base/inter/dag"
)

type EpochDownloaderConfig struct {
	RecheckInterval        time.Duration
	DefaultChunkSize       dag.Metric
	ParallelChunksDownload int
}
