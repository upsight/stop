package stop

import (
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
	assert.Equal(t, 0, ms.stopCalls)
	assert.Equal(t, 0, ms.waitForStoppedCalls)
	close(sg.stop)
	sg.wg.Wait()
	assert.Equal(t, 1, ms.stopCalls)
	assert.Equal(t, 1, ms.waitForStoppedCalls)
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
	assert.False(t, sg.IsStopping())
	sg.isStopping = true
	assert.True(t, sg.IsStopping())
}

func TestGroupStop(t *testing.T) {
	sg := NewGroup()
	assert.False(t, sg.isStopping)
	sg.Stop()
	assert.True(t, sg.isStopping)
	select {
	case <-sg.stop:
	case <-time.After(1 * time.Second):
		assert.Fail(t, "Stop() did not close the stop channel")
	}

	// Test that you can call Stop more than once. Without the isStopping guard,
	// closing the channel twice would cause a panic.
	sg.Stop()
}

func TestGroupStopChannel(t *testing.T) {
	sg := NewGroup()
	assert.Exactly(t, sg.stop, sg.StopChannel())
}

func TestGroupStopOnSignal(t *testing.T) {
	sg := NewGroup()
	sg.StopOnSignal(syscall.SIGWINCH)
	syscall.Kill(syscall.Getpid(), syscall.SIGWINCH)
	select {
	case <-sg.stop:
	case <-time.After(1 * time.Second):
		assert.Fail(t, "Signal did not close stop channel")
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
		assert.Fail(t, "sg.Wait() didn't complete")
	}
}
