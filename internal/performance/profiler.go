package performance

import (
	"fmt"
	"log/slog"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"time"
)

// Profiler provides profiling capabilities
type Profiler struct {
	logger     *slog.Logger
	cpuProfile *os.File
	memProfile *os.File
	gorProfile *os.File
	traceFile  *os.File
	isRunning  bool
	startTime  time.Time
	profileDir string
}

// ProfileConfig holds profiling configuration
type ProfileConfig struct {
	OutputDir       string
	EnableCPU       bool
	EnableMemory    bool
	EnableGoroutine bool
	EnableTrace     bool
	SampleRate      time.Duration
}

// NewProfiler creates a new profiler
func NewProfiler(config ProfileConfig, logger *slog.Logger) *Profiler {
	if config.OutputDir == "" {
		config.OutputDir = "./profiles"
	}
	if config.SampleRate == 0 {
		config.SampleRate = 100 * time.Millisecond
	}

	// Create output directory
	os.MkdirAll(config.OutputDir, 0755)

	return &Profiler{
		logger:     logger,
		profileDir: config.OutputDir,
		startTime:  time.Now(),
	}
}

// StartCPUProfile starts CPU profiling
func (p *Profiler) StartCPUProfile() error {
	if p.cpuProfile != nil {
		return fmt.Errorf("CPU profile already running")
	}

	filename := fmt.Sprintf("%s/cpu-%d.prof", p.profileDir, p.startTime.Unix())
	f, err := os.Create(filename)
	if err != nil {
		p.logger.Error("failed to create CPU profile file", slog.String("error", err.Error()))
		return err
	}

	if err := pprof.StartCPUProfile(f); err != nil {
		f.Close()
		p.logger.Error("failed to start CPU profile", slog.String("error", err.Error()))
		return err
	}

	p.cpuProfile = f
	p.logger.Info("CPU profiling started", slog.String("file", filename))
	return nil
}

// StopCPUProfile stops CPU profiling
func (p *Profiler) StopCPUProfile() error {
	if p.cpuProfile == nil {
		return fmt.Errorf("CPU profile not running")
	}

	pprof.StopCPUProfile()
	p.cpuProfile.Close()
	p.cpuProfile = nil

	p.logger.Info("CPU profiling stopped")
	return nil
}

// CaptureMemoryProfile captures memory profile
func (p *Profiler) CaptureMemoryProfile() error {
	runtime.GC()

	filename := fmt.Sprintf("%s/mem-%d.prof", p.profileDir, p.startTime.Unix())
	f, err := os.Create(filename)
	if err != nil {
		p.logger.Error("failed to create memory profile file", slog.String("error", err.Error()))
		return err
	}
	defer f.Close()

	if err := pprof.WriteHeapProfile(f); err != nil {
		p.logger.Error("failed to write memory profile", slog.String("error", err.Error()))
		return err
	}

	p.logger.Info("memory profile captured", slog.String("file", filename))
	return nil
}

// CaptureGoroutineProfile captures goroutine profile
func (p *Profiler) CaptureGoroutineProfile() error {
	filename := fmt.Sprintf("%s/goroutine-%d.prof", p.profileDir, p.startTime.Unix())
	f, err := os.Create(filename)
	if err != nil {
		p.logger.Error("failed to create goroutine profile file", slog.String("error", err.Error()))
		return err
	}
	defer f.Close()

	if err := pprof.Lookup("goroutine").WriteTo(f, 0); err != nil {
		p.logger.Error("failed to write goroutine profile", slog.String("error", err.Error()))
		return err
	}

	p.logger.Info("goroutine profile captured", slog.String("file", filename))
	return nil
}

// StartTracing starts execution tracing
func (p *Profiler) StartTracing() error {
	if p.traceFile != nil {
		return fmt.Errorf("tracing already running")
	}

	filename := fmt.Sprintf("%s/trace-%d.out", p.profileDir, p.startTime.Unix())
	f, err := os.Create(filename)
	if err != nil {
		p.logger.Error("failed to create trace file", slog.String("error", err.Error()))
		return err
	}

	if err := trace.Start(f); err != nil {
		f.Close()
		p.logger.Error("failed to start tracing", slog.String("error", err.Error()))
		return err
	}

	p.traceFile = f
	p.logger.Info("execution tracing started", slog.String("file", filename))
	return nil
}

// StopTracing stops execution tracing
func (p *Profiler) StopTracing() error {
	if p.traceFile == nil {
		return fmt.Errorf("tracing not running")
	}

	trace.Stop()
	p.traceFile.Close()
	p.traceFile = nil

	p.logger.Info("execution tracing stopped")
	return nil
}

// MemoryStats holds memory statistics
type MemoryStats struct {
	Alloc        uint64
	TotalAlloc   uint64
	Sys          uint64
	NumGC        uint32
	Goroutines   int
	HeapAlloc    uint64
	HeapSys      uint64
	HeapIdle     uint64
	HeapInuse    uint64
	HeapObjects  uint64
	StackInuse   uint64
	MSpanInuse   uint64
	MCacheInuse  uint64
	PauseTotalNs uint64
	LastPauseNs  uint64
	Timestamp    time.Time
}

// GetMemoryStats captures current memory statistics
func (p *Profiler) GetMemoryStats() *MemoryStats {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &MemoryStats{
		Alloc:        m.Alloc,
		TotalAlloc:   m.TotalAlloc,
		Sys:          m.Sys,
		NumGC:        m.NumGC,
		Goroutines:   runtime.NumGoroutine(),
		HeapAlloc:    m.HeapAlloc,
		HeapSys:      m.HeapSys,
		HeapIdle:     m.HeapIdle,
		HeapInuse:    m.HeapInuse,
		HeapObjects:  m.HeapObjects,
		StackInuse:   m.StackInuse,
		MSpanInuse:   m.MSpanInuse,
		MCacheInuse:  m.MCacheInuse,
		PauseTotalNs: m.PauseNs[(m.NumGC+255)%256],
		LastPauseNs:  m.PauseNs[m.NumGC%256],
		Timestamp:    time.Now(),
	}
}

// MonitorMemory continuously monitors memory usage
type MemoryMonitor struct {
	profiler    *Profiler
	interval    time.Duration
	stopChan    chan struct{}
	stoppedChan chan struct{}
	stats       []*MemoryStats
}

// NewMemoryMonitor creates a new memory monitor
func NewMemoryMonitor(profiler *Profiler, interval time.Duration) *MemoryMonitor {
	if interval == 0 {
		interval = 1 * time.Second
	}
	return &MemoryMonitor{
		profiler:    profiler,
		interval:    interval,
		stopChan:    make(chan struct{}),
		stoppedChan: make(chan struct{}),
		stats:       make([]*MemoryStats, 0),
	}
}

// Start starts memory monitoring
func (mm *MemoryMonitor) Start() {
	go func() {
		defer close(mm.stoppedChan)
		ticker := time.NewTicker(mm.interval)
		defer ticker.Stop()

		for {
			select {
			case <-mm.stopChan:
				return
			case <-ticker.C:
				stats := mm.profiler.GetMemoryStats()
				mm.stats = append(mm.stats, stats)

				mm.profiler.logger.Debug("memory stats",
					slog.Uint64("alloc", stats.Alloc),
					slog.Uint64("heap_alloc", stats.HeapAlloc),
					slog.Int("goroutines", stats.Goroutines),
					slog.Uint64("gc_count", uint64(stats.NumGC)),
				)
			}
		}
	}()
}

// Stop stops memory monitoring
func (mm *MemoryMonitor) Stop() {
	close(mm.stopChan)
	<-mm.stoppedChan
}

// GetStats returns collected memory statistics
func (mm *MemoryMonitor) GetStats() []*MemoryStats {
	return mm.stats
}

// AnalyzeMemoryTrend analyzes memory usage trend
func (mm *MemoryMonitor) AnalyzeMemoryTrend() map[string]interface{} {
	if len(mm.stats) < 2 {
		return nil
	}

	first := mm.stats[0]
	last := mm.stats[len(mm.stats)-1]

	allocGrowth := float64(last.Alloc-first.Alloc) / float64(first.Alloc) * 100
	gcCount := last.NumGC - first.NumGC

	return map[string]interface{}{
		"alloc_growth_percent": allocGrowth,
		"gc_count":             gcCount,
		"heap_objects_delta":   int64(last.HeapObjects) - int64(first.HeapObjects),
		"goroutines_delta":     last.Goroutines - first.Goroutines,
		"duration":             last.Timestamp.Sub(first.Timestamp),
	}
}

// PrintMemoryStats prints memory statistics
func (stats *MemoryStats) Print() {
	fmt.Printf("\nMemory Statistics:\n")
	fmt.Printf("  Alloc:         %.2f MB\n", float64(stats.Alloc)/1024/1024)
	fmt.Printf("  TotalAlloc:    %.2f MB\n", float64(stats.TotalAlloc)/1024/1024)
	fmt.Printf("  Sys:           %.2f MB\n", float64(stats.Sys)/1024/1024)
	fmt.Printf("  HeapAlloc:     %.2f MB\n", float64(stats.HeapAlloc)/1024/1024)
	fmt.Printf("  HeapObjects:   %d\n", stats.HeapObjects)
	fmt.Printf("  Goroutines:    %d\n", stats.Goroutines)
	fmt.Printf("  NumGC:         %d\n", stats.NumGC)
	fmt.Printf("  LastPauseNs:   %d ns\n", stats.LastPauseNs)
}
