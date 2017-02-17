package stop

import (
	"os"
	"os/signal"
	"sync"
)

// Group behaves like a WaitGroup, but it also coordinates shutting down
// the things attached to it when the Group is stopped.
type Group struct {
	isStopping bool
	mutex      sync.RWMutex
	stop       chan bool
	wg         *sync.WaitGroup
}

// NewGroup returns a Group object.
func NewGroup() *Group {
	return &Group{
		stop: make(chan bool),
		wg:   &sync.WaitGroup{},
	}
}

// Add adds a Stopper to the stop group. The stop group will call Stop on the
// stopper when the group is stopped. The group's Wait method will block until
// WaitForStopped returns for all attached stoppers.
func (s *Group) Add(stopper Stopper) {
	s.wg.Add(1)
	go func() {
		select {
		case <-stopper.StoppedChannel():
		case <-s.stop:
			stopper.Stop()
			stopper.WaitForStopped()
		}
		s.wg.Done()
	}()
}

// IsStopping returns true if Stop has been called.
func (s *Group) IsStopping() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.isStopping
}

// Stop notifies the Stopped channel that attached stoppers should stop. If
// already stopped, this is a no-op.
func (s *Group) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if !s.isStopping {
		close(s.stop)
		s.isStopping = true
	}
}

// StopChannel returns a channel that will be closed when Stop is called.
func (s *Group) StopChannel() chan bool {
	return s.stop
}

// StopOnSignal will call stop the group when the given os signals are
// received. If no signals are passed, it will trigger for any signal.
func (s *Group) StopOnSignal(sigs ...os.Signal) {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, sigs...)
	go func() {
		select {
		case <-s.stop:
		case <-ch:
			s.Stop()
		}
		signal.Stop(ch)
		close(ch)
	}()
}

// Wait blocks until everything attached to the group has stopped.
func (s *Group) Wait() {
	s.wg.Wait()
}
