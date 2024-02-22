package clickhouse

import (
	"sync"
)

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

func (c *Cache) UpdatePipelines(data map[int64]float64, updated map[int64]bool) {
	c.Lock()
	defer c.Unlock()
	for k, v := range data {
		timestamp, ok := c.pipelines[k]
		if !ok || timestamp < v {
			c.pipelines[k] = v
			if updated != nil {
				updated[k] = true
			}
		} else {
			if updated != nil {
				updated[k] = false
			}
		}
	}
}

func (c *Cache) UpdateJobs(data []int64, updated []bool) {
	c.Lock()
	defer c.Unlock()
	for i, id := range data {
		_, ok := c.jobs[id]
		if !ok {
			c.jobs[id] = struct{}{}
		}

		if i < len(updated) {
			updated[i] = !ok
		}
	}
}

func (c *Cache) UpdateSections(data []int64, updated []bool) {
	c.Lock()
	defer c.Unlock()
	for i, id := range data {
		_, ok := c.sections[id]
		if !ok {
			c.sections[id] = struct{}{}
		}

		if i < len(updated) {
			updated[i] = !ok
		}
	}
}

func (c *Cache) UpdateBridges(data []int64, updated []bool) {
	c.Lock()
	defer c.Unlock()
	for i, id := range data {
		_, ok := c.bridges[id]
		if !ok {
			c.bridges[id] = struct{}{}
		}

		if i < len(updated) {
			updated[i] = !ok
		}
	}
}

func (c *Cache) UpdateTestReports(data []string, updated []bool) {
	c.Lock()
	defer c.Unlock()
	for i, id := range data {
		_, ok := c.testReports[id]
		if !ok {
			c.testReports[id] = struct{}{}
		}

		if i < len(updated) {
			updated[i] = !ok
		}
	}
}

func (c *Cache) UpdateTestSuites(data []string, updated []bool) {
	c.Lock()
	defer c.Unlock()
	for i, id := range data {
		_, ok := c.testSuites[id]
		if !ok {
			c.testSuites[id] = struct{}{}
		}

		if i < len(updated) {
			updated[i] = !ok
		}
	}
}

func (c *Cache) UpdateTestCases(data []string, updated []bool) {
	c.Lock()
	defer c.Unlock()
	for i, id := range data {
		_, ok := c.testCases[id]
		if !ok {
			c.testCases[id] = struct{}{}
		}

		if i < len(updated) {
			updated[i] = !ok
		}
	}
}

func (c *Cache) UpdateLogEmbeddedMetrics(data []int64, updated []bool) {
	c.Lock()
	defer c.Unlock()

	newJobIDs := make(map[int64]struct{}, len(data))

	for i, jobID := range data {
		_, ok := c.logEmbeddedMetrics[jobID]
		if !ok {
			newJobIDs[jobID] = struct{}{}
		}

		if i < len(updated) {
			updated[i] = !ok
		}
	}

	for jobID := range newJobIDs {
		c.logEmbeddedMetrics[jobID] = struct{}{}
	}
}

func (c *Cache) UpdateTraceSpans(data []string, updated []bool) {
	c.Lock()
	defer c.Unlock()
	for i, id := range data {
		_, ok := c.traceSpans[id]
		if !ok {
			c.traceSpans[id] = struct{}{}
		}

		if i < len(updated) {
			updated[i] = !ok
		}
	}
}
