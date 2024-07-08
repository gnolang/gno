package consensus

/*

Something like this could be useful, but currently our state machine doesn't
emit all the round step events expected, say to break on PrecommitWait (no
EventNewRoundStep is emitted). So far it's been fine to just subscribe/ensure
as the tests already do, so this abstraction maybe isn't useful.

type debugger struct {
	rsChan       <-chan events.Event  // event.NewRoundStep's
	newBreakChan chan cstypes.HRS     // new breaks
	didBreakChan chan chan<- struct{} // close chan struct{} to resume
	breakHRS     cstypes.HRS
}

func makeDebugger(cs *ConsensusState) *debugger {
	dbg := &debugger{
		rsChan:       subscribe(cs.evsw, cstypes.EventNewRoundStep{}),
		newBreakChan: make(chan cstypes.HRS, 0),
		didBreakChan: make(chan chan<- struct{}, 0),
		breakHRS:     cstypes.HRS{},
	}
	go dbg.listenRoutine()
	return dbg
}

func (dbg *debugger) listenRoutine() {
	for {
		select {
		case event, ok := <-dbg.rsChan:
			if !ok {
				return // done
			}
			if dbg.breakHRS.IsHRSZero() {
				continue
			}
			newStepEvent := event.(cstypes.EventNewRoundStep)
			if dbg.breakHRS.Compare(newStepEvent.HRS) <= 0 {
				resumeChan := make(chan struct{}, 0)
				dbg.didBreakChan <- resumeChan // block once
				// OnBreak()
				<-resumeChan                 // block twice
				dbg.breakHRS = cstypes.HRS{} // reset
			} else {
				// TODO use some log
				fmt.Printf("[INFO] ignoring event that comes before breakpoint %v\n", event)
				continue
			}
		case breakHRS := <-dbg.newBreakChan:
			if dbg.breakHRS.IsHRSZero() {
				dbg.breakHRS = breakHRS
			} else {
				panic("debugger breakpoint already set")
			}
		}
	}
}

func (dbg *debugger) SetBreak(hrs cstypes.HRS) {
	dbg.newBreakChan <- hrs
}

func (dbg *debugger) OnBreak(cb func()) {
	resumeChan := <-dbg.didBreakChan // block until break
	cb()                             // run callback body
	close(resumeChan)                // resume
}

*/
