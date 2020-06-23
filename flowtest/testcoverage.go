package flowtest

import (
	"bytes"
	"fmt"
	"sync"

	simulator "github.com/edwardbrowncross/amazon-connect-simulator"
	"github.com/edwardbrowncross/amazon-connect-simulator/event"
	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

// CoverageReporter tracks which branches of a call flow have been explored during testing.
type CoverageReporter struct {
	s    *simulator.Simulator
	seen map[string]bool
	m    sync.Mutex
}

// NewCoverageReporter creates a CoverageReporter.
// CoverageReporter tracks which branches of a call flow have been explored during testing.
// The given simulator is used to determine the routes that exist.
// Use CoverageReporter.Track to track all calls under test.
// Use GetCoverage to get a coverage value between 0 and 1 after all calls are complete.
func NewCoverageReporter(s *simulator.Simulator) CoverageReporter {
	return CoverageReporter{
		s:    s,
		seen: map[string]bool{},
		m:    sync.Mutex{},
	}
}

// Track monitors an ongoing call.
// Any route covered by the call after it is passed to Track is considered tested.
func (cr *CoverageReporter) Track(call *simulator.Call) {
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

// Coverage returns a value between 0 and 1, where 0 indicates no test coverage, and 1 represents 100% coverage.
// Coverage is measured as the percentage of connections between blocks that calls passed through during testing.
func (cr *CoverageReporter) Coverage() float64 {
	cr.m.Lock()
	defer cr.m.Unlock()
	return float64(len(cr.seen)) / float64(cr.numBranches())
}

// CoverageReportFlow is one element of the return value of CoverageReport().
type CoverageReportFlow struct {
	Name    string
	Modules []CoverageReportModule
}

// CoverageReportModule is one element of CoverageReportFlow.
type CoverageReportModule struct {
	ID       flow.ModuleID
	Type     flow.ModuleType
	Branches []CoverageReportBranch
}

// CoverageReportBranch is one element of CoverageReportModule.
type CoverageReportBranch struct {
	Type    flow.ModuleBranchCondition
	Dest    flow.ModuleID
	Covered bool
}

// CoverageReport generates a structured list of all the branches and whether they have been covered.
func (cr *CoverageReporter) CoverageReport() []CoverageReportFlow {
	cr.m.Lock()
	flows := cr.s.Flows()
	cFlows := make([]CoverageReportFlow, len(flows))
	for i, f := range cr.s.Flows() {
		cFlows[i] = CoverageReportFlow{
			Name:    f.Metadata.Name,
			Modules: make([]CoverageReportModule, len(f.Modules)),
		}
		for j, m := range f.Modules {
			cFlows[i].Modules[j] = CoverageReportModule{
				ID:       m.ID,
				Type:     m.Type,
				Branches: make([]CoverageReportBranch, len(m.Branches)),
			}
			for k, b := range m.Branches {
				cFlows[i].Modules[j].Branches[k] = CoverageReportBranch{
					Type:    b.Condition,
					Dest:    b.Transition,
					Covered: cr.seen[formatKey(m.ID, b.Transition)],
				}
			}
		}
	}
	cr.m.Unlock()
	return cFlows
}

func (cr *CoverageReporter) add(evt event.BranchEvent) {
	key := formatKey(evt.From, evt.To)
	cr.m.Lock()
	defer cr.m.Unlock()
	cr.seen[key] = true
}

func (cr *CoverageReporter) numBranches() int {
	var n int
	for _, f := range cr.s.Flows() {
		for _, m := range f.Modules {
			n += len(m.Branches)
		}
	}
	return n
}

func formatKey(src flow.ModuleID, dst flow.ModuleID) string {
	return fmt.Sprintf("%s->%s", src, dst)
}

// FormatCoverageReport takes the output of CoverageReport() and formats it into a printable form.
// If colors is true, ansi color highlighting is added to indicate covered or missed branches.
func FormatCoverageReport(report []CoverageReportFlow, colors bool) string {
	buf := bytes.NewBufferString("")
	for _, f := range report {
		buf.WriteString(fmt.Sprintf("%s:\n", f.Name))
		for _, m := range f.Modules {
			buf.WriteString(fmt.Sprintf("\t%s (%s):\n", m.Type, m.ID))
			for _, b := range m.Branches {
				c := fmt.Sprintf("%v", b.Covered)
				c2 := ""
				if colors {
					if b.Covered {
						c = "\u001b[32m"
					} else {
						c = "\u001b[31m"
					}
					c2 = "\u001b[0m"
				}
				buf.WriteString(fmt.Sprintf("\t\t%s %s (%s)%s\n", c, b.Type, b.Dest, c2))
			}
		}
	}
	return buf.String()
}
