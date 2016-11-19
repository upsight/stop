package stop

import (
	"os"
	"os/signal"
	"sync"
)

// StopGroup behaves like a WaitGroup, but it also coordinates shutting down
// the things attached to it when the StopGroup is stopped.
type StopGroup struct {
	isStopping bool
	mutex      sync.RWMutex
	stop       chan bool
	wg         *sync.WaitGroup
}

// NewStopGroup returns a StopGroup object.
func NewStopGroup() *StopGroup {
	return &StopGroup{
		stop: make(chan bool),
		wg:   &sync.WaitGroup{},
	}
}

// Add adds a Stopper to the stop group. The stop group will call Stop on the
// stopper when the group is stopped. The group's Wait method will block until
// WaitForStopped returns for all attached stoppers.
func (s *StopGroup) Add(stopper Stopper) {
	s.wg.Add(1)
	go func() {
		<-s.stop
		stopper.Stop()
		stopper.WaitForStopped()
		s.wg.Done()
	}()
}

// IsStopping returns true if Stop has been called.
func (s *StopGroup) IsStopping() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.isStopping
}

// Stop notifies the Stopped channel that attached stoppers should stop. If
// already stopped, this is a no-op.
func (s *StopGroup) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if !s.isStopping {
		close(s.stop)
		s.isStopping = true
	}
}

// StopChannel returns a channel that will be closed when Stop is called.
func (s *StopGroup) StopChannel() chan bool {
	return s.stop
}

// StopOnSignal will call stop the group when the given os signals are
// received. If no signals are passed, it will trigger for any signal.
func (s *StopGroup) StopOnSignal(sigs ...os.Signal) {
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
func (s *StopGroup) Wait() {
	s.wg.Wait()
}
