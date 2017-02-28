package stop

import (
	"testing"
	"time"
)

func TestNewChannelStopper(t *testing.T) {
	s := NewChannelStopper()
	if s == nil {
		t.Fatal("s is nil")
	}
	if s.stop == nil {
		t.Errorf("s.stop = nil")
	}
	if s.stopped == nil {
		t.Errorf("s.stopped = nil")
	}
	if s.isStopped {
		t.Error("s.isStopped = true, expected false")
	}
	if s.isStopping {
		t.Error("s.isStopping = true, expected false")
	}
}

func TestChannelStopperIsStopped(t *testing.T) {
	s := NewChannelStopper()
	if s.IsStopped() {
		t.Error("s.IsStopped() = true, expected false")
	}

	s.isStopped = true
	if !s.IsStopped() {
		t.Error("s.IsStopped() = false, expected true")
	}
}

func TestChannelStopperIsStopping(t *testing.T) {
	s := NewChannelStopper()
	if s.IsStopping() {
		t.Error("s.IsStopping() = true, expected false")
	}

	s.isStopping = true
	if !s.IsStopping() {
		t.Error("s.IsStopping() = false, expected true")
	}
}

func TestChannelStopperStop(t *testing.T) {
	s := NewChannelStopper()
	if s.isStopping {
		t.Error("s.isStopping = true, expected false")
	}

	s.Stop()
	if !s.isStopping {
		t.Error("s.isStopping = false, expected true")
	}
	select {
	case <-s.stop:
	case <-time.After(1 * time.Second):
		t.Error("Stop() did not close stop channel")
	}

}

func TestChannelStopperStopChannel(t *testing.T) {
	s := NewChannelStopper()
	if s.StopChannel() != s.stop {
		t.Errorf("s.StopChannel() = %#v, expected %#v", s.StopChannel(), s.stop)
	}
}

func TestChannelStopperStopped(t *testing.T) {
	s := NewChannelStopper()
	if s.isStopped {
		t.Errorf("s.isStopped is true, expected false")
	}

	s.Stopped()
	if !s.isStopped {
		t.Errorf("s.isStopped is false, expected true")
	}

	select {
	case <-s.stopped:
	case <-time.After(1 * time.Second):
		t.Error("Stopped() did not close stopped channel")
	}
}

func TestChannelStopperStoppedChannel(t *testing.T) {
	s := NewChannelStopper()
	if s.StoppedChannel() != s.stopped {
		t.Errorf("s.StoppedChannel() = %#v, expected %#v", s.StoppedChannel(), s.stopped)
	}
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
		t.Error("WaitForStopped() didn't return")
	}
}
