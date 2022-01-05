package dot

import (
	"bytes"
	"fmt"
	SpringCore "github.com/go-spring/spring-core"
	"testing"
)

func TestDotBase(t *testing.T) {
	shapes := []Shape{NodeShapeEgg, NodeShapeEllipse, NodeShapeCircle, NodeShapeRecord, NodeShapeMrecord}
	nodes := make([]Node, 5)
	for i := 0; i < 5; i++ {
		nodes[i] = Node{
			name:  fmt.Sprintf("node%d", i),
			shape: shapes[i],
		}
	}
	edge0 := Edge{
		fromName:  nodes[0].name,
		toName:    nodes[1].name,
	}
	edge1 := Edge{
		fromName:  nodes[0].name,
		toName:    nodes[2].name,
	}
	edge2 := Edge{
		fromName:  nodes[1].name,
		toName:    nodes[3].name,
	}
	edge3 := Edge{
		fromName:  nodes[1].name,
		toName:    nodes[4].name,
	}

	root := DiGraph{
		name:    "test",
		bgColor: GroupBgColorWhite,
		nodes:   nodes,
		edges: []Edge{
			edge0,
			edge1,
			edge2,
			edge3,
		},
	}
	buffer := &bytes.Buffer{}
	err := root.ToDot(buffer)
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(buffer.String())
	}
}

func TestDotWithContext(t *testing.T) {
	applicationContext := SpringCore.DefaultApplicationContext()
	applicationContext.RegisterBean(&IamTestHandler{})
	applicationContext.RegisterBean(&IamTestSimpleController{})
	applicationContext.RegisterBean(&IamTestComplexController{})

	var validators = []IamTestValidator{
		&IamTest0Validator{},
		&IamTest1Validator{},
		&IamTest2Validator{},
		&IamTest3Validator{},
		&IamTest4Validator{},
		&IamTest5Validator{},
	}

	applicationContext.RegisterNameBeanFn("iamTestValidator", func() IamTestValidator {
		return validators[0]
	})

	applicationContext.RegisterBean(validators)

	var services = []IamTestService{
		&IamTestServiceRemixImpl{},
		&IamTestServiceDBImpl{},
		&IamTestServiceRedisImpl{},
	}
	applicationContext.RegisterNameBeanFn("iamTestService", func() IamTestService {
		return services[0]
	})

	applicationContext.RegisterBean(services)

	applicationContext.RegisterBean(&Redis{})
	applicationContext.RegisterBean(&Datasource{})
	applicationContext.AutoWireBeans()

	buffer := bytes.Buffer{}
	_ = WithContext(applicationContext).ToDot(&buffer)
	fmt.Println(buffer.String())
}


// 测试用结构体
type (
	IamTestHandler struct {
		IamTestSimpleController  *IamTestSimpleController  `autowire:""`
		IamTestComplexController *IamTestComplexController `autowire:""`
	}
	IamTestValidator interface {
		Validate() error
	}
	IamTestService interface {
		Test() bool
	}
)

type (
	IamTest0Validator struct {
	}
	IamTest1Validator struct {
	}
	IamTest2Validator struct {
	}
	IamTest3Validator struct {
	}
	IamTest4Validator struct {
	}
	IamTest5Validator struct {
	}
)

type (
	IamTestSimpleController struct {
		Validator IamTestValidator `autowire:"iamTestValidator"`
		Service   IamTestService   `autowire:"iamTestService"`
	}
	IamTestComplexController struct {
		Validators []IamTestValidator `autowire:""`
		Services   []IamTestService   `autowire:""`
	}
	IamTestServiceDBImpl struct {
		Datasource *Datasource `autowire:""`
	}
	Datasource struct {
		Host     string
		Port     uint16
		Protocol string
	}
	IamTestServiceRedisImpl struct {
		Redis *Redis `autowire:""`
	}
	Redis struct {
		Host string
		Port uint16
	}
	IamTestServiceRemixImpl struct {
		Datasource *Datasource `autowire:""`
		Redis      *Redis      `autowire:""`
	}
)


func (validator *IamTest0Validator) Validate() error {
	return nil
}

func (validator *IamTest1Validator) Validate() error {
	return nil
}

func (validator *IamTest2Validator) Validate() error {
	return nil
}

func (validator *IamTest3Validator) Validate() error {
	return nil
}

func (validator *IamTest4Validator) Validate() error {
	return nil
}

func (validator *IamTest5Validator) Validate() error {
	return nil
}

func (service *IamTestServiceDBImpl) Test() bool {
	return true
}

func (service *IamTestServiceRedisImpl) Test() bool {
	return true
}

func (service *IamTestServiceRemixImpl) Test() bool {
	return true
}