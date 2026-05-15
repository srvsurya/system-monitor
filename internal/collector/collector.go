package collector

import (
	"log"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"
)

type jobKind int

const (
	jobCPU jobKind = iota
	jobMemory
	jobDisk
	jobNetwork
)

const totalJobsPerTick = 4

type result struct {
	kind jobKind

	cpuUsage float64

	memoryUsed  int
	memoryTotal int

	diskUsed  int
	diskTotal int

	// network
	netUpload   float64
	netDownload float64
}

// metricRow is what we INSERT — one complete row per tick.
type metricRow struct {
	CPUUsage    float64 `db:"cpu_usage"`
	MemoryUsed  int     `db:"memory_used"`
	MemoryTotal int     `db:"memory_total"`
	DiskUsed    int     `db:"disk_used"`
	DiskTotal   int     `db:"disk_total"`
	NetUpload   float64 `db:"net_upload"`
	NetDownload float64 `db:"net_download"`
}

const (
	workerCount   = 5
	tickInterval  = 5 * time.Second
	jobBufferSize = 20
)

type Collector struct {
	db      *sqlx.DB
	jobs    chan jobKind
	results chan result
	done    chan struct{}

	lastNetBytes *net.IOCountersStat
	lastNetTime  time.Time
}

func New(db *sqlx.DB) *Collector {
	return &Collector{
		db:      db,
		jobs:    make(chan jobKind, jobBufferSize),
		results: make(chan result, jobBufferSize),
		done:    make(chan struct{}),
	}
}

func (c *Collector) Start() {
	if err := c.seedNetworkBaseline(); err != nil {
		log.Printf("[collector] warning: could not seed network baseline: %v", err)
	}

	for i := 0; i < workerCount; i++ {
		go c.worker()
	}

	go c.aggregator()
	go c.scheduler()

	log.Println("[collector] started — 5 workers, 5-second tick")
}

func (c *Collector) Stop() {
	close(c.done)
}

// scheduler - defines tick rate

func (c *Collector) scheduler() {
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.enqueueAll()
		case <-c.done:
			return
		}
	}
}

func (c *Collector) enqueueAll() { // -- manages queue
	jobs := []jobKind{jobCPU, jobMemory, jobDisk, jobNetwork}
	for _, j := range jobs {
		select {
		case c.jobs <- j:
		default:
			log.Printf("[collector] warning: job channel full, dropping job %d", j)
		}
	}
}

// worker goroutine, sends to results chan
func (c *Collector) worker() {
	for {
		select {
		case job := <-c.jobs:
			res, err := c.collect(job)
			if err != nil {
				log.Printf("[collector] error on job %d: %v", job, err)
				// Still send a zero result so the aggregator isn't left waiting
				// for a result that will never arrive.
				c.results <- result{kind: job}
				continue
			}
			c.results <- res
		case <-c.done:
			return
		}
	}
}

func (c *Collector) collect(j jobKind) (result, error) {
	switch j {
	case jobCPU:
		return c.collectCPU()
	case jobMemory:
		return c.collectMemory()
	case jobDisk:
		return c.collectDisk()
	case jobNetwork:
		return c.collectNetwork()
	}
	return result{}, nil
}

// aggregator waits for exactly totalJobsPerTick results, merges them into
// one metricRow, and inserts a single row into system_metrics.
func (c *Collector) aggregator() {
	for {
		row, ok := c.collectTick()
		if !ok {
			return // done channel was closed
		}
		if err := c.insertRow(row); err != nil {
			log.Printf("[collector] DB insert error: %v", err)
		}
	}
}

// collectTick blocks until all 4 results for one tick arrive, or until done.
func (c *Collector) collectTick() (metricRow, bool) {
	var row metricRow
	received := 0

	for received < totalJobsPerTick {
		select {
		case res := <-c.results:
			c.mergeResult(&row, res)
			received++
		case <-c.done:
			return metricRow{}, false
		}
	}

	return row, true
}

func (c *Collector) mergeResult(row *metricRow, res result) {
	switch res.kind {
	case jobCPU:
		row.CPUUsage = res.cpuUsage
	case jobMemory:
		row.MemoryUsed = res.memoryUsed
		row.MemoryTotal = res.memoryTotal
	case jobDisk:
		row.DiskUsed = res.diskUsed
		row.DiskTotal = res.diskTotal
	case jobNetwork:
		row.NetUpload = res.netUpload
		row.NetDownload = res.netDownload
	}
}

func (c *Collector) insertRow(row metricRow) error {
	_, err := c.db.Exec(`
		INSERT INTO system_metrics
			(cpu_usage, memory_used, memory_total, disk_used, disk_total, net_upload, net_download, timestamp)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, NOW())`,
		row.CPUUsage,
		row.MemoryUsed,
		row.MemoryTotal,
		row.DiskUsed,
		row.DiskTotal,
		row.NetUpload,
		row.NetDownload,
	)
	return err
}

func (c *Collector) collectCPU() (result, error) {
	percentages, err := cpu.Percent(0, false)
	if err != nil {
		return result{}, err
	}
	return result{kind: jobCPU, cpuUsage: percentages[0]}, nil
}

func (c *Collector) collectMemory() (result, error) {
	v, err := mem.VirtualMemory()
	if err != nil {
		return result{}, err
	}
	return result{
		kind:        jobMemory,
		memoryUsed:  int(v.Used / 1024 / 1024),
		memoryTotal: int(v.Total / 1024 / 1024),
	}, nil
}

func (c *Collector) collectDisk() (result, error) {
	usage, err := disk.Usage("/")
	if err != nil {
		return result{}, err
	}
	return result{
		kind:      jobDisk,
		diskUsed:  int(usage.Used / 1024 / 1024),
		diskTotal: int(usage.Total / 1024 / 1024),
	}, nil
}

func (c *Collector) collectNetwork() (result, error) {
	counters, err := net.IOCounters(false)
	if err != nil {
		return result{}, err
	}
	if len(counters) == 0 {
		return result{kind: jobNetwork}, nil
	}

	now := time.Now()
	current := counters[0]

	if c.lastNetBytes == nil {
		c.lastNetBytes = &current
		c.lastNetTime = now
		return result{kind: jobNetwork}, nil
	}

	elapsed := now.Sub(c.lastNetTime).Seconds()
	uploadKBs := float64(current.BytesSent-c.lastNetBytes.BytesSent) / elapsed / 1024
	downloadKBs := float64(current.BytesRecv-c.lastNetBytes.BytesRecv) / elapsed / 1024

	c.lastNetBytes = &current
	c.lastNetTime = now

	return result{
		kind:        jobNetwork,
		netUpload:   uploadKBs,
		netDownload: downloadKBs,
	}, nil
}

func (c *Collector) seedNetworkBaseline() error {
	counters, err := net.IOCounters(false)
	if err != nil {
		return err
	}
	if len(counters) == 0 {
		return nil
	}
	c.lastNetBytes = &counters[0]
	c.lastNetTime = time.Now()
	return nil
}
