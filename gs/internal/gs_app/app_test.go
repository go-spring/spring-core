package gs_app

import (
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/util/assert"
	"github.com/go-spring/spring-core/util/sysconf"
	"go.uber.org/mock/gomock"
)

func clean() {
	os.Args = nil
	os.Clearenv()
	sysconf.Clear()
}

func TestApp(t *testing.T) {

	t.Run("config refresh error", func(t *testing.T) {
		t.Cleanup(clean)
		_ = sysconf.Set("a", "123")
		_ = os.Setenv("GS_A_B", "456")
		app := NewApp()
		err := app.Run()
		assert.Error(t, err, "property 'a' is a value but 'a.b' wants other type")
	})

	t.Run("container refresh error", func(t *testing.T) {
		app := NewApp()
		app.C.Provide(func() (*http.Server, error) {
			return nil, errors.New("fail to create bean")
		})
		err := app.Run()
		assert.Error(t, err, "fail to create bean")
	})

	t.Run("runner return error", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		defer ctrl.Finish()
		r := gs.NewMockRunner(ctrl)
		r.EXPECT().Run().Return(errors.New("runner return error"))
		app := NewApp()
		app.C.Object(r).AsRunner()
		err := app.Run()
		assert.Error(t, err, "runner return error")
	})

}

//func TestApp_Start_RunnerError(t *testing.T) {
//	app := NewApp()
//	mockRunner := new(gs.MockRunner)
//	mockRunner.EXPECT().Run().Return(errors.New("runner error"))
//	app.Runners = []gs.Runner{mockRunner}
//
//	// 模拟配置加载成功
//	app.P = gs_conf.NewAppConfig() // 假设默认配置有效
//
//	err := app.Start()
//	assert.Error(t, err, "runner error")
//}
//
//func TestApp_JobsWithShutdown(t *testing.T) {
//	app := NewApp()
//	app.EnableJobs = true
//
//	mockJob := new(gs.MockJob)
//	mockJob.EXPECT().Run(gomock.Any()).Return(errors.New("job error"))
//	app.Jobs = []gs.Job{mockJob}
//
//	// 启动并快速触发关闭
//	go func() {
//		time.Sleep(100 * time.Millisecond)
//		app.ShutDown()
//	}()
//
//	err := app.Start()
//	assert.Nil(t, err)
//	assert.True(t, app.Exiting())
//}
//
//func TestApp_ServersPanicRecovery(t *testing.T) {
//	app := NewApp()
//	app.EnableServers = true
//
//	mockServer := new(gs.MockServer)
//	mockServer.EXPECT().ListenAndServe(gomock.Any()).Do(func() {
//		panic("server panic")
//	})
//	app.Servers = []gs.Server{mockServer}
//
//	app.Start()
//	assert.True(t, app.Exiting())
//}
//
//func TestApp_ShutDown(t *testing.T) {
//	app := NewApp()
//	assert.False(t, app.Exiting())
//
//	app.ShutDown()
//	assert.True(t, app.Exiting())
//
//	// 确保上下文被取消
//	select {
//	case <-app.ctx.Done():
//	default:
//		assert.Fail(t, "context should be canceled")
//	}
//}
//
//func TestApp_Stop(t *testing.T) {
//	app := NewApp()
//	mockServer := new(gs.MockServer)
//	mockServer.EXPECT().Shutdown(gomock.Any()).Return(nil)
//	app.Servers = []gs.Server{mockServer}
//
//	// 模拟等待组
//	app.wg.Add(1)
//	go func() {
//		defer app.wg.Done()
//		time.Sleep(50 * time.Millisecond)
//	}()
//
//	app.Stop()
//}
//
//func TestApp_RunWithSignal(t *testing.T) {
//	app := NewApp()
//
//	// 模拟接收SIGTERM
//	go func() {
//		time.Sleep(100 * time.Millisecond)
//		p, _ := os.FindProcess(os.Getpid())
//		_ = p.Signal(syscall.SIGTERM)
//	}()
//
//	err := app.Run()
//	assert.Nil(t, err)
//	assert.True(t, app.Exiting())
//}
//
//// 测试并发安全
//func TestApp_ConcurrentShutdown(t *testing.T) {
//	app := NewApp()
//	var wg sync.WaitGroup
//
//	for i := 0; i < 10; i++ {
//		wg.Add(1)
//		go func() {
//			defer wg.Done()
//			app.ShutDown()
//		}()
//	}
//
//	wg.Wait()
//	assert.True(t, app.Exiting())
//}
