/*
 * Copyright 2012-2024 the original author or authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *      https://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package gs_app

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"reflect"
	"strings"
	"syscall"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/internal/gs_conf"
	"github.com/go-spring/spring-core/gs/internal/gs_ctx"
)

const (
	Version = "go-spring@v1.1.3"
	Website = "https://go-spring.com/"
)

// SpringBannerVisible 是否显示 banner。
const SpringBannerVisible = "spring.banner.visible"

// AppRunner 命令行启动器接口
type AppRunner interface {
	Run(ctx gs.Context)
}

// AppEvent 应用运行过程中的事件
type AppEvent interface {
	OnAppStart(ctx gs.Context)     // 应用启动的事件
	OnAppStop(ctx context.Context) // 应用停止的事件
}

type tempApp struct {
	banner string
}

// App 应用
type App struct {
	*tempApp

	b *Boot
	c *gs_ctx.Container
	p *gs_conf.Configuration

	exitChan chan struct{}

	Events []AppEvent `autowire:"${application-event.collection:=*?}"`
}

// NewApp application 的构造函数
func NewApp() *App {
	return &App{
		c:        gs_ctx.New(),
		p:        gs_conf.NewConfiguration(),
		tempApp:  &tempApp{},
		exitChan: make(chan struct{}),
	}
}

// Banner 自定义 banner 字符串。
func (app *App) Banner(banner string) {
	app.banner = banner
}

func (app *App) Start() (gs.Context, error) {

	//showBanner, _ := strconv.ParseBool(e.p.Get(SpringBannerVisible))
	//if showBanner {
	//	app.printBanner(app.getBanner(e))
	//}

	if app.b != nil {
		err := app.b.run()
		if err != nil {
			return nil, err
		}
	}

	p, err := app.p.Refresh()
	if err != nil {
		return nil, err
	}

	app.Object(app)

	err = app.c.RefreshProperties(p)
	if err != nil {
		return nil, err
	}

	err = app.c.Refresh(false)
	if err != nil {
		return nil, err
	}

	//var runners []AppRunner
	//err = app.c.Get(&runners, "${command-line-runner.collection:=*?}")
	//if err != nil {
	//	return nil, err
	//}
	//
	//// 执行命令行启动器
	//for _, r := range runners {
	//	r.Run(app.c)
	//}

	// 通知应用启动事件
	for _, event := range app.Events {
		event.OnAppStart(app.c)
	}

	// 通知应用停止事件
	app.c.Go(func(ctx context.Context) {
		<-ctx.Done()
		for _, event := range app.Events {
			event.OnAppStop(context.Background())
		}
	})

	return app.c, nil
}

func (app *App) wait() {
	go func() {
		ch := make(chan os.Signal, 1)
		signal.Notify(ch, os.Interrupt, syscall.SIGTERM)
		sig := <-ch
		app.ShutDown(fmt.Sprintf("signal %v", sig))
	}()
	<-app.exitChan
}

func (app *App) Stop() {

	// if app.b != nil {
	// 	app.b.c.Close()
	// }

	app.c.Close()
}

func (app *App) Run() error {
	_, err := app.Start()
	if err != nil {
		return err
	}
	app.wait()
	app.Stop()
	return nil
}

const DefaultBanner = `
                                              (_)              
  __ _    ___             ___   _ __    _ __   _   _ __     __ _ 
 / _' |  / _ \   ______  / __| | '_ \  | '__| | | | '_ \   / _' |
| (_| | | (_) | |______| \__ \ | |_) | | |    | | | | | | | (_| |
 \__, |  \___/           |___/ | .__/  |_|    |_| |_| |_|  \__, |
  __/ |                        | |                          __/ |
 |___/                         |_|                         |___/ 
`

func (app *App) getBanner() string {
	if app.banner != "" {
		return app.banner
	}
	banner := DefaultBanner
	// for _, resource := range resources {
	// 	if b, _ := ioutil.ReadAll(resource); b != nil {
	// 		banner = string(b)
	// 	}
	// }
	return banner
}

// printBanner 打印 banner 到控制台
func (app *App) printBanner(banner string) {

	if banner[0] != '\n' {
		fmt.Println()
	}

	maxLength := 0
	for _, s := range strings.Split(banner, "\n") {
		fmt.Printf("\x1b[36m%s\x1b[0m\n", s) // CYAN
		if len(s) > maxLength {
			maxLength = len(s)
		}
	}

	if banner[len(banner)-1] != '\n' {
		fmt.Println()
	}

	var padding []byte
	if n := (maxLength - len(Version)) / 2; n > 0 {
		padding = make([]byte, n)
		for i := range padding {
			padding[i] = ' '
		}
	}
	fmt.Println(string(padding) + Version + "\n")
}

// ShutDown 关闭执行器
func (app *App) ShutDown(msg ...string) {
	select {
	case <-app.exitChan:
		// chan 已关闭，无需再次关闭。
	default:
		close(app.exitChan)
	}
}

// Boot 返回 *bootstrap 对象。
func (app *App) Boot() *Boot {
	if app.b == nil {
		app.b = newBoot()
	}
	return app.b
}

func (app *App) Group(fn gs.GroupFunc) {
	app.c.Group(fn)
}

// Accept 参考 Container.Accept 的解释。
func (app *App) Accept(b *gs.BeanDefinition) *gs.BeanDefinition {
	return app.c.Accept(b)
}

// Object 参考 Container.Object 的解释。
func (app *App) Object(i interface{}) *gs.BeanDefinition {
	return app.c.Accept(gs_ctx.NewBean(reflect.ValueOf(i)))
}

// Provide 参考 Container.Provide 的解释。
func (app *App) Provide(ctor interface{}, args ...gs.Arg) *gs.BeanDefinition {
	return app.c.Accept(gs_ctx.NewBean(ctor, args...))
}
