package stop

import (
	"syscall"
	"testing"
	"time"
)

type MockStopper struct {
	stopCalls           int
	waitForStoppedCalls int
}

func (ms *MockStopper) IsStopping() bool {
	return false
}

func (ms *MockStopper) IsStopped() bool {
	return false
}

func (ms *MockStopper) Stop() {
	ms.stopCalls++
}

func (ms *MockStopper) StopChannel() chan bool {
	return nil
}

func (ms *MockStopper) Stopped() {}

func (ms *MockStopper) StoppedChannel() chan bool {
	return nil
}

func (ms *MockStopper) WaitForStopped() {
	ms.waitForStoppedCalls++
}

func TestNewGroup(t *testing.T) {
	sg := NewGroup()
	if sg == nil {
		t.Fatal("sg was unexpectedly nil")
	}
	if sg.isStopping {
		t.Errorf("expected sg.isStopping to be false, got true")
	}
	if sg.stop == nil {
		t.Errorf("sg.stop was unexpectedly nil")
	}
	if sg.wg == nil {
		t.Errorf("sg.wg was unexpectedly nil")
	}
}

func TestGroupAdd(t *testing.T) {
	sg := NewGroup()
	ms := &MockStopper{}
	sg.Add(ms)
	if ms.stopCalls != 0 {
		t.Errorf("ms.stopCalls = %d; expected 0", ms.stopCalls)
	}
	if ms.waitForStoppedCalls != 0 {
		t.Errorf("ms.waitForStoppedCalls = %d; expected 0", ms.waitForStoppedCalls)
	}
	close(sg.stop)
	sg.wg.Wait()
	if ms.stopCalls != 1 {
		t.Errorf("ms.stopCalls = %d; expected 1", ms.stopCalls)
	}
	if ms.waitForStoppedCalls != 1 {
		t.Errorf("ms.waitForStoppedCalls = %d; expected 1", ms.waitForStoppedCalls)
	}
}

func TestGroupAddEarlyStop(t *testing.T) {
	// This test is validating that calling stop on a stopper unblocks the
	// stop group cleanup goroutine for the stopper.
	sg := NewGroup()
	ms := NewChannelStopper()
	sg.Add(ms)
	ms.Stopped()
	timer := time.AfterFunc(100*time.Millisecond, func() {
		t.Error("Group Add() did not unblock after 100ms")
		close(sg.stop)
	})
	sg.wg.Wait()
	timer.Stop()
}

func TestGroupIsStopping(t *testing.T) {
	sg := NewGroup()
	if sg.IsStopping() {
		t.Error("sg.IsStopping() is true, expected false")
	}
	sg.isStopping = true
	if !sg.IsStopping() {
		t.Error("sg.IsStopping() is false, expected true")
	}
}

func TestGroupStop(t *testing.T) {
	sg := NewGroup()
	if sg.isStopping {
		t.Error("sg.isStopping is true, expected false")
	}
	sg.Stop()
	if !sg.isStopping {
		t.Error("sg.isStopping is false, expected true")
	}
	select {
	case <-sg.stop:
	case <-time.After(1 * time.Second):
		t.Error("Stop() did not close the stop channel")
	}

	// Test that you can call Stop more than once. Without the isStopping guard,
	// closing the channel twice would cause a panic.
	sg.Stop()
}

func TestGroupStopChannel(t *testing.T) {
	sg := NewGroup()
	if sg.stop != sg.StopChannel() {
		t.Errorf("sg.stop = %#v; expected %#v", sg.stop, sg.StopChannel())
	}
}

func TestGroupStopOnSignal(t *testing.T) {
	sg := NewGroup()
	sg.StopOnSignal(syscall.SIGWINCH)
	err := syscall.Kill(syscall.Getpid(), syscall.SIGWINCH)
	if err != nil {
		t.Error(err)
	}
	select {
	case <-sg.stop:
	case <-time.After(1 * time.Second):
		t.Error("Signal did not close stop channel")
	}
}

func TestGroupWait(t *testing.T) {
	ch := make(chan bool)
	sg := NewGroup()
	sg.wg.Add(1)
	go func() {
		sg.Wait()
		close(ch)
	}()
	sg.wg.Done()
	select {
	case <-ch:
	case <-time.After(1 * time.Second):
		t.Error("sg.Wait() didn't complete")
	}
}
