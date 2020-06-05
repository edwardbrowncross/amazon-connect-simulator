package simulator

import (
	"fmt"
	"sync"

	"github.com/edwardbrowncross/amazon-connect-simulator/event"
)

// TestCoverageReporter tracks which branches of a call flow have been explored during testing.
type TestCoverageReporter struct {
	s    *Simulator
	seen map[string]bool
	m    sync.Mutex
}

// NewTestCoverageReporter creates a TestCoverageReporter.
// TestCoverageReporter tracks which branches of a call flow have been explored during testing.
// The given simulator is used to determine the routes that exist.
// Use TestCoverageReporter.Track to track all calls under test.
// Use GetCoverage to get a coverage value between 0 and 1 after all calls are complete.
func NewTestCoverageReporter(s *Simulator) TestCoverageReporter {
	return TestCoverageReporter{
		s:    s,
		seen: map[string]bool{},
		m:    sync.Mutex{},
	}
}

// Track monitors an ongoing call.
// Any route covered by the call after it is passed to Track is considered tested.
func (cr *TestCoverageReporter) Track(call *Call) {
	evts := make(chan event.Event)
	call.Subscribe(evts)
	go func() {
		for {
			evt, ok := <-evts
			if !ok {
				return
			}
			if evt.Type() != event.BranchType {
				continue
			}
			cr.add(evt.(event.BranchEvent))
		}
	}()
}

// GetCoverage returns a value between 0 and 1, where 0 indicates no test coverage, and 1 represents 100% coverage.
// Coverage is measured as the percentage of connections between blocks that calls passed through during testing.
func (cr *TestCoverageReporter) GetCoverage() float64 {
	cr.m.Lock()
	defer cr.m.Unlock()
	return float64(len(cr.seen)) / float64(cr.numBranches())
}

func (cr *TestCoverageReporter) add(evt event.BranchEvent) {
	key := fmt.Sprintf("%s->%s", evt.From, evt.To)
	cr.m.Lock()
	defer cr.m.Unlock()
	cr.seen[key] = true
}

func (cr *TestCoverageReporter) numBranches() int {
	var n int
	for _, m := range cr.s.modules {
		n += len(m.Branches)
	}
	return n
}
