package stop

import (
	"sync"
)

// Stopper is an interface for objects that can be used with StopGroup. See
// ChannelStopper for a simple implementation using channels.
//
// IsStopping should return true if Stop has been called.
//
// IsStopped should return a true if Stopped has been called.
//
// Stop should be a non-blocking notification to stop any work in progress.
// It should be safe to call this multiple times.
//
// Stopped should be a non-blocking notification that work has been stopped.
// It should be safe to call this multiple times.
//
// WaitForStopped should block until the work has been stopped.
type Stopper interface {
	IsStopping() bool
	IsStopped() bool
	Stop()
	StopChannel() chan bool
	Stopped()
	StoppedChannel() chan bool
	WaitForStopped()
}

// ChannelStopper is an implementation of Stopper using channels.
type ChannelStopper struct {
	stop       chan bool
	stopped    chan bool
	mutex      sync.RWMutex
	isStopped  bool
	isStopping bool
}

// NewChannelStopper returns a new ChannelStopper.
func NewChannelStopper() *ChannelStopper {
	return &ChannelStopper{
		stop:    make(chan bool),
		stopped: make(chan bool),
	}
}

// IsStopped returns true if Stopped has been called.
func (s *ChannelStopper) IsStopped() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.isStopped
}

// IsStopping returns true if Stop has been called.
func (s *ChannelStopper) IsStopping() bool {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.isStopping
}

// Stop notifies that work should stop by closing the stop channel.
func (s *ChannelStopper) Stop() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if !s.isStopping {
		close(s.stop)
		s.isStopping = true
	}
}

// StopChannel returns a channel that will be closed when Stop is called.
func (s *ChannelStopper) StopChannel() chan bool {
	return s.stop
}

// Stopped notifies that work has stopped by closing the stopped channel.
func (s *ChannelStopper) Stopped() {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if !s.isStopped {
		close(s.stopped)
		s.isStopped = true
	}
}

// StoppedChannel returns a channel that will be closed when Stopped is called.
func (s *ChannelStopper) StoppedChannel() chan bool {
	return s.stopped
}

// WaitForStopped blocks until Stopped is called.
func (s *ChannelStopper) WaitForStopped() {
	<-s.stopped
}
