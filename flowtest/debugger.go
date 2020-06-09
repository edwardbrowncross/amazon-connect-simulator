package flowtest

import (
	"sync"

	simulator "github.com/edwardbrowncross/amazon-connect-simulator"
	"github.com/edwardbrowncross/amazon-connect-simulator/event"
	"github.com/edwardbrowncross/amazon-connect-simulator/flow"
)

type breakpoint struct {
	before bool
	after  bool
}

// Debugger is used to control the flow of a running call.
// It allows you to pause, step and resume the flow, and add and remove breakpoints.
type Debugger struct {
	bp      map[flow.ModuleID]breakpoint
	mutex   sync.RWMutex
	pause   chan<- interface{}
	resume  chan<- interface{}
	step    chan<- interface{}
	detatch chan<- interface{}
	pos     <-chan flow.ModuleID
}

// NewDebugger creates a debugger for controlling the flow of the given call.
func NewDebugger(call *simulator.Call) *Debugger {
	lookahead := make(chan event.Event)
	flowBlock := make(chan event.Event)
	call.Subscribe(lookahead)
	call.Subscribe(flowBlock)
	resume := make(chan interface{})
	step := make(chan interface{})
	pause := make(chan interface{})
	detatch := make(chan interface{})
	pos := make(chan flow.ModuleID)

	d := Debugger{
		bp:      map[flow.ModuleID]breakpoint{},
		step:    step,
		resume:  resume,
		pause:   pause,
		detatch: detatch,
		pos:     pos,
	}

	go func() {
		var paused bool
		for {
			// See what the next event emitted is.
			evt, ok := <-lookahead
			if !ok {
				return
			}
			if evt.Type() != event.ModuleType {
				<-flowBlock
				continue
			}
			// Event is for entering a block (the only time we ever pause).
			// Should we pause at caller request?
			select {
			case <-detatch:
				return
			case <-pause:
				paused = true
			default:
			}
			// If not, should we pause because of a breakpoint?
			id := evt.(event.ModuleEvent).ID
			var deferPause bool
			if !paused {
				d.mutex.RLock()
				if bp, found := d.bp[id]; found {
					paused = true
					deferPause = bp.after
				}
				d.mutex.RUnlock()
			}
			// Have we paused?
			// (either for above reasons or because we were paused from last loop).
			if paused && !deferPause {
			block:
				for {
					select {
					// Send the current position if anyone wants it.
					case pos <- id:
					// Wait for resume instruction.
					case <-resume:
						paused = false
						break block
					case <-step:
						break block
					case <-detatch:
						return
					case <-pause:
					}
				}
			}
			// Unblock the call flow.
			<-flowBlock
		}
	}()

	return &d
}

// SetBreakpoint adds a new breakpoint, to pause the flow as the call enters the given block.
func (d *Debugger) SetBreakpoint(id flow.ModuleID) {
	d.mutex.Lock()
	if bp, ok := d.bp[id]; ok {
		bp.before = true
		d.bp[id] = bp
	} else {
		d.bp[id] = breakpoint{before: true}
	}
	d.mutex.Unlock()
}

// SetBreakpointAfter adds a new breakpoint, to pause the flow as the call leaves the given block.
func (d *Debugger) SetBreakpointAfter(id flow.ModuleID) {
	d.mutex.Lock()
	if bp, ok := d.bp[id]; ok {
		bp.after = true
		d.bp[id] = bp
	} else {
		d.bp[id] = breakpoint{after: true}
	}
	d.mutex.Unlock()
}

// RemoveBreakpoint removes all breakpoints attached to the given block (entry and exist both).
func (d *Debugger) RemoveBreakpoint(id flow.ModuleID) {
	d.mutex.Lock()
	delete(d.bp, id)
	d.mutex.Unlock()
}

// Step moves a paused call on one block. If the flow is not paused, this call blocks until it is paused.
func (d *Debugger) Step() {
	d.step <- true
}

// Resume unpauses a paused flow. If the flow is not paused, this call blocks until it is paused.
func (d *Debugger) Resume() {
	d.resume <- true
}

// Pause pauses a flow when it next enters a new block.
func (d *Debugger) Pause() {
	d.pause <- true
}

// Paused returns true if the flow is currently paused.
func (d *Debugger) Paused() bool {
	_, paused := d.Position()
	return paused
}

// Position returns the current position (id of the current block) of the flow.
// It can only give a valid answer if the flow is paused.
// If the flow is no paused, ok will be false and id is unreliable.
func (d *Debugger) Position() (id flow.ModuleID, ok bool) {
	select {
	case id := <-d.pos:
		return id, true
	default:
		return flow.ModuleID(""), false
	}
}

// Wait waits until the flow is paused.
func (d *Debugger) Wait() {
	<-d.pos
}
