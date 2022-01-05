package dot

import "reflect"

//颜色预留，可能用作后续区分业务模块
const (
	GroupBgColorRed    = GroupBgColor("red")
	GroupBgColorGreen  = GroupBgColor("green")
	GroupBgColorBlue   = GroupBgColor("blue")
	GroupBgColorPink   = GroupBgColor("pink")
	GroupBgColorYellow = GroupBgColor("yellow")
	GroupBgColorBlack  = GroupBgColor("black")
	GroupBgColorPurple = GroupBgColor("purple")
	GroupBgColorOrange = GroupBgColor("orange")
	GroupBgColorCoral  = GroupBgColor("coral")
	GroupBgColorBeige  = GroupBgColor("beige")
	GroupBgColorBrown  = GroupBgColor("brown")
	GroupBgColorWhite  = GroupBgColor("white")
)

//使用 NodeShapeEgg (egg) 表示 interface
//使用 NodeShapeRecord (record) 表示 struct
//使用 NodeShapeCircle (circle) 表示 collection : map, array, slice
//使用 NodeShapePlaintext (plaintext) 表示基本类型
//其余使用默认值
//其余形状保留
const (
	//NodeShapeMrecord 圆角矩形
	NodeShapeMrecord = Shape("Mrecord")
	//NodeShapeRecord 矩形
	NodeShapeRecord = Shape("record")
	//NodeShapeCircle 圆形
	NodeShapeCircle = Shape("circle")
	//NodeShapeEgg 蛋形
	NodeShapeEgg = Shape("egg")
	//NodeShapePlaintext 无形状边框，纯文本
	NodeShapePlaintext = Shape("plaintext")
	//NodeShapeEllipse 椭圆, 默认椭圆
	NodeShapeEllipse = Shape("Ellipse")
)

var typeShapeMapping = map[reflect.Kind]Shape{
	reflect.Array: NodeShapeCircle,
	reflect.Map: NodeShapeCircle,
	reflect.Slice: NodeShapeCircle,

	reflect.Struct: NodeShapeRecord,

	reflect.Bool: NodeShapePlaintext,
	reflect.Int: NodeShapePlaintext,
	reflect.Int8: NodeShapePlaintext,
	reflect.Int16: NodeShapePlaintext,
	reflect.Int32: NodeShapePlaintext,
	reflect.Int64: NodeShapePlaintext,
	reflect.Uint: NodeShapePlaintext,
	reflect.Uint8: NodeShapePlaintext,
	reflect.Uint16: NodeShapePlaintext,
	reflect.Uint32: NodeShapePlaintext,
	reflect.Uint64: NodeShapePlaintext,
	reflect.Uintptr: NodeShapePlaintext,
	reflect.Float32: NodeShapePlaintext,
	reflect.Float64: NodeShapePlaintext,
	reflect.Complex64: NodeShapePlaintext,
	reflect.Complex128: NodeShapePlaintext,

	reflect.Interface: NodeShapeEgg,
}
