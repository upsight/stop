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

func TestNewStopGroup(t *testing.T) {
	sg := NewStopGroup()
	if assert.NotNil(t, sg) {
		assert.False(t, sg.isStopping)
		assert.NotNil(t, sg.stop)
		assert.NotNil(t, sg.wg)
	}
}

func TestStopGroupAdd(t *testing.T) {
	sg := NewStopGroup()
	ms := &MockStopper{}
	sg.Add(ms)
	assert.Equal(t, 0, ms.stopCalls)
	assert.Equal(t, 0, ms.waitForStoppedCalls)
	close(sg.stop)
	sg.wg.Wait()
	assert.Equal(t, 1, ms.stopCalls)
	assert.Equal(t, 1, ms.waitForStoppedCalls)
}

func TestStopGroupIsStopping(t *testing.T) {
	sg := NewStopGroup()
	assert.False(t, sg.IsStopping())
	sg.isStopping = true
	assert.True(t, sg.IsStopping())
}

func TestStopGroupStop(t *testing.T) {
	sg := NewStopGroup()
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

func TestStopGroupStopChannel(t *testing.T) {
	sg := NewStopGroup()
	assert.Exactly(t, sg.stop, sg.StopChannel())
}

func TestStopGroupStopOnSignal(t *testing.T) {
	sg := NewStopGroup()
	sg.StopOnSignal(syscall.SIGWINCH)
	syscall.Kill(syscall.Getpid(), syscall.SIGWINCH)
	select {
	case <-sg.stop:
	case <-time.After(1 * time.Second):
		assert.Fail(t, "Signal did not close stop channel")
	}
}

func TestStopGroupWait(t *testing.T) {
	ch := make(chan bool)
	sg := NewStopGroup()
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
