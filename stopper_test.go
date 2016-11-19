package stop

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewChannelStopper(t *testing.T) {
	s := NewChannelStopper()
	if assert.NotNil(t, s) {
		assert.NotNil(t, s.stop)
		assert.NotNil(t, s.stopped)
		assert.False(t, s.isStopped)
		assert.False(t, s.isStopping)
	}
}

func TestChannelStopperIsStopped(t *testing.T) {
	s := NewChannelStopper()
	assert.False(t, s.IsStopped())
	s.isStopped = true
	assert.True(t, s.IsStopped())
}

func TestChannelStopperIsStopping(t *testing.T) {
	s := NewChannelStopper()
	assert.False(t, s.IsStopping())
	s.isStopping = true
	assert.True(t, s.IsStopping())
}

func TestChannelStopperStop(t *testing.T) {
	s := NewChannelStopper()
	assert.False(t, s.isStopping)
	s.Stop()
	assert.True(t, s.isStopping)
	select {
	case <-s.stop:
	case <-time.After(1 * time.Second):
		assert.Fail(t, "Stop() did not close stop channel")
	}

}

func TestChannelStopperStopChannel(t *testing.T) {
	s := NewChannelStopper()
	assert.Exactly(t, s.stop, s.StopChannel())
}

func TestChannelStopperStopped(t *testing.T) {
	s := NewChannelStopper()
	assert.False(t, s.isStopped)
	s.Stopped()
	assert.True(t, s.isStopped)
	select {
	case <-s.stopped:
	case <-time.After(1 * time.Second):
		assert.Fail(t, "Stopped() did not close stopped channel")
	}
}

func TestChannelStopperStoppedChannel(t *testing.T) {
	s := NewChannelStopper()
	assert.Exactly(t, s.stopped, s.StoppedChannel())
}

func TestChannelStopperWaitForStopped(t *testing.T) {
	s := NewChannelStopper()
	ch := make(chan bool)
	go func() {
		s.WaitForStopped()
		close(ch)
	}()
	s.Stopped()
	select {
	case <-ch:
	case <-time.After(1 * time.Second):
		assert.Fail(t, "WaitForStopped() didn't return")
	}
}
