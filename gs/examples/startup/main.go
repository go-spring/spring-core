package main

import (
	"context"
	"fmt"
	"time"

	"github.com/go-spring/spring-core/gs"
	"github.com/go-spring/spring-core/util/sysconf"
	"github.com/go-spring/spring-core/util/syslog"
)

func init() {
	gs.Object(&Service{})     // Register the Service object.
	gs.Object(&Job{}).AsJob() // Register the Job object and mark it as a scheduled job.
}

type Service struct{}

// Echo prints a formatted log message using the syslog package.
func (s *Service) Echo(format string, a ...any) {
	syslog.Infof(fmt.Sprintf(format, a...))
}

// Job struct represents a scheduled task that depends on the Service.
type Job struct {
	Service *Service `autowire:""`                // Automatically inject the Service dependency.
	AppName string   `value:"${spring.app.name}"` // Read the application name from the configuration.
}

// Run method is executed when the Job is triggered.
func (j *Job) Run(ctx context.Context) error {
	time.Sleep(time.Second * 2)
	j.Service.Echo("app '%s' will exit", j.AppName)
	gs.ShutDown() // Shut down the application.
	return nil
}

func main() {
	// Set the application name in the configuration.
	_ = sysconf.Set("spring.app.name", "test")

	// Start the Go-Spring application. If it fails, log the error.
	if err := gs.Run(); err != nil {
		syslog.Errorf("app run failed: %s", err.Error())
	}
}
