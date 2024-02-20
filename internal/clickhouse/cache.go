package clickhouse

import "sync"

type Cache struct {
	sync.RWMutex

	pipelines          map[int64]float64
	jobs               map[int64]struct{}
	sections           map[int64]struct{}
	bridges            map[int64]struct{}
	testReports        map[string]struct{}
	testSuites         map[string]struct{}
	testCases          map[string]struct{}
	logEmbeddedMetrics map[int64]struct{}
	traceSpans         map[string]struct{}
}

func NewCache() *Cache {
	return &Cache{
		pipelines:          make(map[int64]float64),
		jobs:               make(map[int64]struct{}),
		sections:           make(map[int64]struct{}),
		bridges:            make(map[int64]struct{}),
		testReports:        make(map[string]struct{}),
		testSuites:         make(map[string]struct{}),
		testCases:          make(map[string]struct{}),
		logEmbeddedMetrics: make(map[int64]struct{}),
		traceSpans:         make(map[string]struct{}),
	}
}

func (c *Cache) UpdatePipelines(data map[int64]float64) map[int64]bool {
	c.Lock()
	defer c.Unlock()
	updated := make(map[int64]bool, len(data))
	for k, v := range data {
		timestamp, ok := c.pipelines[k]
		if !ok || timestamp < v {
			c.pipelines[k] = v
			updated[k] = true
		} else {
			updated[k] = false
		}
	}
	return updated
}

func (c *Cache) UpdateJobs(data []int64) []bool {
	c.Lock()
	defer c.Unlock()
	updated := make([]bool, len(data))
	for i, id := range data {
		_, ok := c.jobs[id]
		if !ok {
			c.jobs[id] = struct{}{}
			updated[i] = true
		}
	}
	return updated
}

func (c *Cache) UpdateSections(data []int64) []bool {
	c.Lock()
	defer c.Unlock()
	updated := make([]bool, len(data))
	for i, id := range data {
		_, ok := c.sections[id]
		if !ok {
			c.sections[id] = struct{}{}
			updated[i] = true
		}
	}
	return updated
}

func (c *Cache) UpdateBridges(data []int64) []bool {
	c.Lock()
	defer c.Unlock()
	updated := make([]bool, len(data))
	for i, id := range data {
		_, ok := c.bridges[id]
		if !ok {
			c.bridges[id] = struct{}{}
			updated[i] = true
		}
	}
	return updated
}

func (c *Cache) UpdateTestReports(data []string) []bool {
	c.Lock()
	defer c.Unlock()
	updated := make([]bool, len(data))
	for i, id := range data {
		_, ok := c.testReports[id]
		if !ok {
			c.testReports[id] = struct{}{}
			updated[i] = true
		}
	}
	return updated
}

func (c *Cache) UpdateTestSuites(data []string) []bool {
	c.Lock()
	defer c.Unlock()
	updated := make([]bool, len(data))
	for i, id := range data {
		_, ok := c.testSuites[id]
		if !ok {
			c.testSuites[id] = struct{}{}
			updated[i] = true
		}
	}
	return updated
}

func (c *Cache) UpdateTestCases(data []string) []bool {
	c.Lock()
	defer c.Unlock()
	updated := make([]bool, len(data))
	for i, id := range data {
		_, ok := c.testCases[id]
		if !ok {
			c.testCases[id] = struct{}{}
			updated[i] = true
		}
	}
	return updated
}

func (c *Cache) UpdateLogEmbeddedMetrics(data []int64) []bool {
	c.Lock()
	defer c.Unlock()

	newJobIDs := make(map[int64]struct{})

	updated := make([]bool, len(data))
	for i, jobID := range data {
		_, ok := c.logEmbeddedMetrics[jobID]
		if !ok {
			newJobIDs[jobID] = struct{}{}
			updated[i] = true
		}
	}

	for jobID := range newJobIDs {
		c.logEmbeddedMetrics[jobID] = struct{}{}
	}

	return updated
}

func (c *Cache) UpdateTraceSpans(data []string) []bool {
	c.Lock()
	defer c.Unlock()
	updated := make([]bool, len(data))
	for i, id := range data {
		_, ok := c.traceSpans[id]
		if !ok {
			c.traceSpans[id] = struct{}{}
			updated[i] = true
		}
	}
	return updated
}
