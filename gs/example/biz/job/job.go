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

func (x *Job) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("job exit")
			return
		case <-time.After(time.Second * 5):
			fmt.Println("job sleep end")
		}
	}
}
