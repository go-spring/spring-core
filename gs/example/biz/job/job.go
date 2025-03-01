package job

import (
	"fmt"
	"time"

	"github.com/go-spring/spring-core/gs"
)

func init() {
	gs.Object(&Job{}).AsJob()
}

type Job struct {
}

func (x *Job) Run() {
	for {
		if gs.Exiting() {
			fmt.Println("job exit 3")
			return
		}
		time.Sleep(time.Second * 5)
		fmt.Println("job sleep end")
	}
}
