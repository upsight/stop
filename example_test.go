package stop_test

import (
	"fmt"
	"time"

	"github.com/upsight/stop"
)

func ExampleChannelStopper() {
	s := stop.NewChannelStopper()

	go func() {
		<-s.StopChannel()
		fmt.Println("Stopper is stopping")
		s.Stopped()
	}()

	s.Stop()
	s.WaitForStopped()
	fmt.Println("Stopper has stopped")

	// Output:
	// Stopper is stopping
	// Stopper has stopped
}

func ExampleStopGroup() {
	type Service struct {
		*stop.ChannelStopper
		t *time.Timer
	}

	runService := func(s *Service, sg *stop.StopGroup, n int) {
		defer s.Stopped()
		select {
		case <-s.StopChannel():
			fmt.Println("Service", n, "stopping")
		case <-s.t.C:
			fmt.Println("Service", n, "timer expired")
			sg.Stop()
		}
	}

	// Add two services to a stop group. The timer for the first
	// service will finish first. It will stop the group, which will
	// cause the stop for the other service to be called.

	sg := stop.NewStopGroup()

	s1 := &Service{
		ChannelStopper: stop.NewChannelStopper(),
		t:              time.NewTimer(1 * time.Second),
	}
	sg.Add(s1)
	go runService(s1, sg, 1)

	s2 := &Service{
		ChannelStopper: stop.NewChannelStopper(),
		t:              time.NewTimer(2 * time.Second),
	}
	sg.Add(s2)
	go runService(s2, sg, 2)

	// Wait for all the services to finish stopping.

	sg.Wait()
	fmt.Println("Services have stopped")

	// Output:
	// Service 1 timer expired
	// Service 2 stopping
	// Services have stopped
}
