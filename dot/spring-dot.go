package dot

import (
	"bytes"
	"fmt"
	SpringCore "github.com/go-spring/spring-core"
	"io"
	"math"
	"reflect"
)

//type declare here
type GroupBgColor string

type DiGraph struct {
	//name 图名, 生成图时使用此字段声明
	name string
	//bgColor 背景色，可使用 #rrggbb 形式
	bgColor GroupBgColor
	//subGraphs 子图集合，目前无用，计划用于模块划分
	subGraphs []DiGraph
	//nodes 此图包含的全部节点
	nodes []Node
	//edges 此节点的边集合
	edges []Edge
}

type Shape string

type Node struct {
	//name 节点名，声明时使用，取 bean的name
	name string
	//shape 节点形状, 现在用于区分各种数据类型, 参考 typeShapeMapping 声明
	shape Shape
}

type Edge struct {
	//fromName 从此节点出发
	fromName string
	//toName 节点指向 node
	toName string
}

//WithContext 根据 ctx 生成图形结构
func WithContext(ctx SpringCore.SpringContext) *DiGraph {
	beans := ctx.GetBeanDefinitions()
	nodes := make([]Node, len(beans))
	edges := make([]Edge, 0, len(beans))

	typeNameMap := make(map[reflect.Type]string)
	//构建节点并记录 type - name
	for i := range beans {
		beanType := reflect.TypeOf(beans[i].Bean())
		typeNameMap[beanType] = beans[i].Name()
		nodes[i] = Node{
			shape: getShape(beanType.Kind()),
			name:  beans[i].Name(),
		}
	}
	// 构建边
	for i:= range beans {
		beanType := reflect.TypeOf(beans[i].Bean())
		if beanType.Kind() == reflect.Struct ||
			beanType.Kind() == reflect.Ptr && beanType.Elem().Kind() == reflect.Struct {
			aType := beanType
			if aType.Kind() == reflect.Ptr {
				aType = aType.Elem()
			}
			for j := 0; j < aType.NumField(); j++ {
				val, ok := aType.Field(j).Tag.Lookup("autowire")
				if ok {
					if val != "" {
						edges = append(edges, Edge{
							fromName: beans[i].Name(),
							toName:   val,
						})
					} else {
						edges = append(edges, Edge{
							fromName: beans[i].Name(),
							toName:   typeNameMap[aType.Field(j).Type],
						})
					}
				}
			}
		}
	}
	return &DiGraph{
		name:    "springGraph",
		bgColor: GroupBgColorWhite,
		nodes:   nodes,
		edges:   edges,
	}
}

//util func here
func RGBToBgColor(red, green, blue int) GroupBgColor {
	return rgbToBgColor(
		limitColorValue(red),
		limitColorValue(green),
		limitColorValue(blue),
	)
}

func limitColorValue(val int) uint8 {
	if val < 0 {
		return 0
	}
	if val > math.MaxUint8 {
		return math.MaxUint8
	}
	return uint8(val)
}

func rgbToBgColor(red, green, blue uint8) GroupBgColor {
	return GroupBgColor(fmt.Sprintf("#%x%x%x", red, green, blue))
}

func getShape(kind reflect.Kind) Shape {
	if shape, have := typeShapeMapping[kind]; have {
		return shape
	}
	return NodeShapeEllipse
}

func (graph *DiGraph) ToDot(writer io.Writer) error {
	buffer := &bytes.Buffer{}
	buffer.Write([]byte(fmt.Sprintf("digraph %s {\n", graph.name)))
	buffer.Write([]byte(fmt.Sprintf("  graph [bgcolor = %s]\n", graph.bgColor)))
	for _, node := range graph.nodes {
		buffer.Write([]byte("  "))
		node.ToDot(buffer)
	}
	for _, edge := range graph.edges {
		buffer.Write([]byte("  "))
		edge.ToDot(buffer)
	}
	buffer.Write([]byte("}\n"))
	_, err := writer.Write(buffer.Bytes())
	return err
}

func (node *Node) ToDot(buffer *bytes.Buffer) {
	buffer.Write([]byte(fmt.Sprintf("\"%s\"[shape = %s]\n", node.name, node.shape)))
}

func (edge *Edge) ToDot(buffer *bytes.Buffer) {
	buffer.Write([]byte(fmt.Sprintf("\"%s\" -> \"%s\" \n", edge.fromName, edge.toName)))
}
