package job

import (
	"context"
	"fmt"
	"time"

	"github.com/go-spring/spring-core/gs"
)

func init() {
	gs.Object(&Job{}).AsJob()
}

type Job struct {
}

func (x *Job) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("job exit")
			return nil
		case <-time.After(time.Second * 5):
			fmt.Println("job sleep end")
		}
	}
}
