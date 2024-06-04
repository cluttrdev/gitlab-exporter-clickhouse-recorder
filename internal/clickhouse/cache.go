package clickhouse

import (
	"sync"
)

type Cache struct {
	sync.RWMutex

	pipelines     map[int64]float64
	jobs          map[int64]struct{}
	sections      map[int64]struct{}
	bridges       map[int64]struct{}
	testReports   map[string]struct{}
	testSuites    map[string]struct{}
	testCases     map[string]struct{}
	mergeRequests map[int64]float64
	metrics       map[int64]struct{}
	projects      map[int64]float64
	traceSpans    map[string]struct{}
}

func NewCache() *Cache {
	return &Cache{
		pipelines:     make(map[int64]float64),
		jobs:          make(map[int64]struct{}),
		sections:      make(map[int64]struct{}),
		bridges:       make(map[int64]struct{}),
		testReports:   make(map[string]struct{}),
		testSuites:    make(map[string]struct{}),
		testCases:     make(map[string]struct{}),
		mergeRequests: make(map[int64]float64),
		metrics:       make(map[int64]struct{}),
		projects:      make(map[int64]float64),
		traceSpans:    make(map[string]struct{}),
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

// UpdateSections updates the cache used to prevent inserting duplicate sections.
// For each key in the given map we check whether it is already cached.
// If it is, the correspopnding map value is set to `false`, else it will be
// added to the cache and the correponding map value is set to `true`.
// In order to not require holding each individual section ID in memory, we
// use the section's job ID as a cache key.
func (c *Cache) UpdateSections(keys map[int64]bool) {
	c.Lock()
	defer c.Unlock()

	newJobIDs := make(map[int64]struct{})

	for jobID := range keys {
		_, ok := c.sections[jobID]
		if !ok {
			newJobIDs[jobID] = struct{}{}
		}

		keys[jobID] = !ok
	}

	for jobID := range newJobIDs {
		c.sections[jobID] = struct{}{}
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

// UpdateTestCases updates the cache used to prevent inserting duplicate test cases.
// For each key in the given map we check whether it is already cached.
// If it is, the correspopnding map value is set to `false`, else it will be
// added to the cache and the correponding map value is set to `true`.
// In order to not require holding each individual test case ID in memory, we
// use the test case's test suite ID as a cache key.
func (c *Cache) UpdateTestCases(keys map[string]bool) {
	c.Lock()
	defer c.Unlock()

	newTestSuiteIDs := make(map[string]struct{})

	for suiteID := range keys {
		_, ok := c.testCases[suiteID]
		if !ok {
			newTestSuiteIDs[suiteID] = struct{}{}
		}

		keys[suiteID] = !ok
	}

	for suiteID := range newTestSuiteIDs {
		c.testCases[suiteID] = struct{}{}
	}
}

func (c *Cache) UpdateMergeRequests(data map[int64]float64, updated map[int64]bool) {
	c.Lock()
	defer c.Unlock()
	for k, v := range data {
		timestamp, ok := c.mergeRequests[k]
		if !ok || timestamp < v {
			c.mergeRequests[k] = v
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

// UpdateMetrics updates the cache used to prevent inserting duplicate metrics.
// For each key in the given map we check whether it is already cached.
// If it is, the correspopnding map value is set to `false`, else it will be
// added to the cache and the correponding map value is set to `true`.
// In order to not require holding each individual metric ID in memory, we
// use the metric's job ID as a cache key.
func (c *Cache) UpdateMetrics(keys map[int64]bool) {
	c.Lock()
	defer c.Unlock()

	newJobIDs := make(map[int64]struct{})

	for jobID := range keys {
		_, ok := c.metrics[jobID]
		if !ok {
			newJobIDs[jobID] = struct{}{}
		}

		keys[jobID] = !ok
	}

	for jobID := range newJobIDs {
		c.metrics[jobID] = struct{}{}
	}
}

func (c *Cache) UpdateProjects(data map[int64]float64, updated map[int64]bool) {
	c.Lock()
	defer c.Unlock()
	for k, v := range data {
		timestamp, ok := c.projects[k]
		if !ok || timestamp < v {
			c.projects[k] = v
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
