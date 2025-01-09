package gs_core

import (
	"bytes"
	"container/list"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"

	"github.com/go-spring/spring-core/conf"
	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/gs/syslog"
	"github.com/go-spring/spring-core/util"
)

var (
	GsContextType = reflect.TypeOf((*gs.Context)(nil)).Elem()
)

type lazyField struct {
	v    reflect.Value
	path string
	tag  string
}

// destroyer 保存具有销毁函数的 bean 以及销毁函数的调用顺序。
type destroyer struct {
	current *gs.BeanDefinition
	earlier []*gs.BeanDefinition
}

func (d *destroyer) foundEarlier(b *gs.BeanDefinition) bool {
	for _, c := range d.earlier {
		if c == b {
			return true
		}
	}
	return false
}

// after 添加一个需要在该 bean 的销毁函数执行之前调用销毁函数的 bean 。
func (d *destroyer) after(b *gs.BeanDefinition) {
	if d.foundEarlier(b) {
		return
	}
	d.earlier = append(d.earlier, b)
}

// getBeforeDestroyers 获取排在 i 前面的 destroyer，用于 sort.Triple 排序。
func getBeforeDestroyers(destroyers *list.List, i interface{}) *list.List {
	d := i.(*destroyer)
	result := list.New()
	for e := destroyers.Front(); e != nil; e = e.Next() {
		c := e.Value.(*destroyer)
		if d.foundEarlier(c.current) {
			result.PushBack(c)
		}
	}
	return result
}

// wiringStack 记录 bean 的注入路径。
type wiringStack struct {
	destroyers   *list.List
	destroyerMap map[string]*destroyer
	beans        []*gs.BeanDefinition
	lazyFields   []lazyField
}

func newWiringStack() *wiringStack {
	return &wiringStack{
		destroyers:   list.New(),
		destroyerMap: make(map[string]*destroyer),
	}
}

// pushBack 添加一个即将注入的 bean 。
func (s *wiringStack) pushBack(b *gs.BeanDefinition) {
	syslog.Debug("push %s %s", b, gs.GetStatusString(b.GetStatus()))
	s.beans = append(s.beans, b)
}

// popBack 删除一个已经注入的 bean 。
func (s *wiringStack) popBack() {
	n := len(s.beans)
	b := s.beans[n-1]
	s.beans = s.beans[:n-1]
	syslog.Debug("pop %s %s", b, gs.GetStatusString(b.GetStatus()))
}

// path 返回 bean 的注入路径。
func (s *wiringStack) path() (path string) {
	for _, b := range s.beans {
		path += fmt.Sprintf("=> %s ↩\n", b)
	}
	return path[:len(path)-1]
}

// saveDestroyer 记录具有销毁函数的 bean ，因为可能有多个依赖，因此需要排重处理。
func (s *wiringStack) saveDestroyer(b *gs.BeanDefinition) *destroyer {
	d, ok := s.destroyerMap[b.ID()]
	if !ok {
		d = &destroyer{current: b}
		s.destroyerMap[b.ID()] = d
	}
	return d
}

// sortDestroyers 对具有销毁函数的 bean 按照销毁函数的依赖顺序进行排序。
func (s *wiringStack) sortDestroyers() []func() {

	destroy := func(v reflect.Value, fn interface{}) func() {
		return func() {
			if fn == nil {
				v.Interface().(gs.BeanDestroy).OnDestroy()
			} else {
				fnValue := reflect.ValueOf(fn)
				out := fnValue.Call([]reflect.Value{v})
				if len(out) > 0 && !out[0].IsNil() {
					syslog.Error(out[0].Interface().(error).Error())
				}
			}
		}
	}

	destroyers := list.New()
	for _, d := range s.destroyerMap {
		destroyers.PushBack(d)
	}
	destroyers = TripleSort(destroyers, getBeforeDestroyers)

	var ret []func()
	for e := destroyers.Front(); e != nil; e = e.Next() {
		d := e.Value.(*destroyer).current
		ret = append(ret, destroy(d.Value(), d.GetDestroy()))
	}
	return ret
}

// wireTag 注入语法的 tag 分解式，字符串形式的完整格式为 TypeName:BeanName? 。
// 注入语法的字符串表示形式分为三个部分，TypeName 是原始类型的全限定名，BeanName
// 是 bean 注册时设置的名称，? 表示注入结果允许为空。
type wireTag struct {
	typeName string
	beanName string
	nullable bool
}

func parseWireTag(str string) (tag wireTag) {

	if str == "" {
		return
	}

	if n := len(str) - 1; str[n] == '?' {
		tag.nullable = true
		str = str[:n]
	}

	i := strings.Index(str, ":")
	if i < 0 {
		tag.beanName = str
		return
	}

	tag.typeName = str[:i]
	tag.beanName = str[i+1:]
	return
}

func (tag wireTag) String() string {
	b := bytes.NewBuffer(nil)
	if tag.typeName != "" {
		b.WriteString(tag.typeName)
		b.WriteString(":")
	}
	b.WriteString(tag.beanName)
	if tag.nullable {
		b.WriteString("?")
	}
	return b.String()
}

func toWireString(tags []wireTag) string {
	var buf bytes.Buffer
	for i, tag := range tags {
		buf.WriteString(tag.String())
		if i < len(tags)-1 {
			buf.WriteByte(',')
		}
	}
	return buf.String()
}

// resolveTag tag 预处理，可能通过属性值进行指定。
func (c *Container) resolveTag(tag string) (string, error) {
	if strings.HasPrefix(tag, "${") {
		s, err := c.p.Data().Resolve(tag)
		if err != nil {
			return "", err
		}
		return s, nil
	}
	return tag, nil
}

func (c *Container) toWireTag(selector gs.BeanSelector) (wireTag, error) {
	switch s := selector.(type) {
	case string:
		s, err := c.resolveTag(s)
		if err != nil {
			return wireTag{}, err
		}
		return parseWireTag(s), nil
	case gs.BeanDefinition:
		return parseWireTag(s.ID()), nil
	case *gs.BeanDefinition:
		return parseWireTag(s.ID()), nil
	default:
		return parseWireTag(util.TypeName(s) + ":"), nil
	}
}

func (c *Container) autowire(v reflect.Value, tags []wireTag, nullable bool, stack *wiringStack) error {
	switch v.Kind() {
	case reflect.Map, reflect.Slice, reflect.Array:
		return c.collectBeans(v, tags, nullable, stack)
	default:
		var tag wireTag
		if len(tags) > 0 {
			tag = tags[0]
		} else if nullable {
			tag.nullable = true
		}
		return c.getBean(v, tag, stack)
	}
}

type byBeanName []SimpleBean

func (b byBeanName) Len() int           { return len(b) }
func (b byBeanName) Less(i, j int) bool { return b[i].GetName() < b[j].GetName() }
func (b byBeanName) Swap(i, j int)      { b[i], b[j] = b[j], b[i] }

// filterBean 返回 tag 对应的 bean 在数组中的索引，找不到返回 -1。
func filterBean(beans []SimpleBean, tag wireTag, t reflect.Type) (int, error) {

	var found []int
	for i, b := range beans {
		if b.Match(tag.typeName, tag.beanName) {
			found = append(found, i)
		}
	}

	if len(found) > 1 {
		msg := fmt.Sprintf("found %d beans, bean:%q type:%q [", len(found), tag, t)
		for _, i := range found {
			msg += "( " + beans[i].String() + " ), "
		}
		msg = msg[:len(msg)-2] + "]"
		return -1, errors.New(msg)
	}

	if len(found) > 0 {
		i := found[0]
		return i, nil
	}

	if tag.nullable {
		return -1, nil
	}

	return -1, fmt.Errorf("can't find bean, bean:%q type:%q", tag, t)
}

func (c *Container) collectBeans(v reflect.Value, tags []wireTag, nullable bool, stack *wiringStack) error {

	t := v.Type()
	if t.Kind() != reflect.Slice && t.Kind() != reflect.Map {
		return fmt.Errorf("should be slice or map in collection mode")
	}

	et := t.Elem()
	if !util.IsBeanReceiver(et) {
		return fmt.Errorf("%s is not valid receiver type", t.String())
	}

	var beans []SimpleBean
	beans = c.beansByType[et]

	{
		var arr []SimpleBean
		for _, b := range beans {
			if b.GetStatus() == gs.Deleted {
				continue
			}
			arr = append(arr, b)
		}
		beans = arr
	}

	if len(tags) > 0 {

		var (
			anyBeans  []SimpleBean
			afterAny  []SimpleBean
			beforeAny []SimpleBean
		)

		foundAny := false
		for _, item := range tags {

			// 是否遇到了"无序"标记
			if item.beanName == "*" {
				if foundAny {
					return fmt.Errorf("more than one * in collection %q", tags)
				}
				foundAny = true
				continue
			}

			index, err := filterBean(beans, item, et)
			if err != nil {
				return err
			}
			if index < 0 {
				continue
			}

			if foundAny {
				afterAny = append(afterAny, beans[index])
			} else {
				beforeAny = append(beforeAny, beans[index])
			}

			tmpBeans := append([]SimpleBean{}, beans[:index]...)
			beans = append(tmpBeans, beans[index+1:]...)
		}

		if foundAny {
			anyBeans = append(anyBeans, beans...)
		}

		n := len(beforeAny) + len(anyBeans) + len(afterAny)
		arr := make([]SimpleBean, 0, n)
		arr = append(arr, beforeAny...)
		arr = append(arr, anyBeans...)
		arr = append(arr, afterAny...)
		beans = arr
	}

	if len(beans) == 0 && !nullable {
		if len(tags) == 0 {
			return fmt.Errorf("no beans collected for %q", toWireString(tags))
		}
		for _, tag := range tags {
			if !tag.nullable {
				return fmt.Errorf("no beans collected for %q", toWireString(tags))
			}
		}
		return nil
	}

	for _, b := range beans {
		switch c.state {
		case Refreshing:
			if err := c.wireBeanInRefreshing(b.(*gs.BeanDefinition), stack); err != nil {
				return err
			}
		case Refreshed:
			if err := c.wireBeanAfterRefreshed(b.(*gs.BeanRuntimeMeta), stack); err != nil {
				return err
			}
		default:
			return fmt.Errorf("state is error for wiring")
		}
	}

	var ret reflect.Value
	switch t.Kind() {
	case reflect.Slice:
		sort.Sort(byBeanName(beans))
		ret = reflect.MakeSlice(t, 0, 0)
		for _, b := range beans {
			ret = reflect.Append(ret, b.Value())
		}
	case reflect.Map:
		ret = reflect.MakeMap(t)
		for _, b := range beans {
			ret.SetMapIndex(reflect.ValueOf(b.GetName()), b.Value())
		}
	default:
	}
	v.Set(ret)
	return nil
}

// getBean 获取 tag 对应的 bean 然后赋值给 v，因此 v 应该是一个未初始化的值。
func (c *Container) getBean(v reflect.Value, tag wireTag, stack *wiringStack) error {

	if !v.IsValid() {
		return fmt.Errorf("receiver must be ref type, bean:%q", tag)
	}

	t := v.Type()
	if !util.IsBeanReceiver(t) {
		return fmt.Errorf("%s is not valid receiver type", t.String())
	}

	var foundBeans []SimpleBean
	for _, b := range c.beansByType[t] {
		if b.GetStatus() == gs.Deleted {
			continue
		}
		if !b.Match(tag.typeName, tag.beanName) {
			continue
		}
		foundBeans = append(foundBeans, b)
	}

	// 指定 bean 名称时通过名称获取，防止未通过 Export 方法导出接口。
	if t.Kind() == reflect.Interface && tag.beanName != "" {
		for _, b := range c.beansByName[tag.beanName] {
			if b.GetStatus() == gs.Deleted {
				continue
			}
			if !b.Type().AssignableTo(t) {
				continue
			}
			if !b.Match(tag.typeName, tag.beanName) {
				continue
			}

			found := false // 对结果排重
			for _, r := range foundBeans {
				if r == b {
					found = true
					break
				}
			}
			if !found {
				foundBeans = append(foundBeans, b)
				syslog.Warn("you should call Export() on %s", b)
			}
		}
	}

	if len(foundBeans) == 0 {
		if tag.nullable {
			return nil
		}
		return fmt.Errorf("can't find bean, bean:%q type:%q", tag, t)
	}

	// 优先使用设置成主版本的 bean
	var primaryBeans []SimpleBean

	for _, b := range foundBeans {
		if b.IsPrimary() {
			primaryBeans = append(primaryBeans, b)
		}
	}

	if len(primaryBeans) > 1 {
		msg := fmt.Sprintf("found %d primary beans, bean:%q type:%q [", len(primaryBeans), tag, t)
		for _, b := range primaryBeans {
			msg += "( " + b.String() + " ), "
		}
		msg = msg[:len(msg)-2] + "]"
		return errors.New(msg)
	}

	if len(primaryBeans) == 0 && len(foundBeans) > 1 {
		msg := fmt.Sprintf("found %d beans, bean:%q type:%q [", len(foundBeans), tag, t)
		for _, b := range foundBeans {
			msg += "( " + b.String() + " ), "
		}
		msg = msg[:len(msg)-2] + "]"
		return errors.New(msg)
	}

	var result SimpleBean
	if len(primaryBeans) == 1 {
		result = primaryBeans[0]
	} else {
		result = foundBeans[0]
	}

	// 确保找到的 bean 已经完成依赖注入。
	switch c.state {
	case Refreshing:
		if err := c.wireBeanInRefreshing(result.(*gs.BeanDefinition), stack); err != nil {
			return err
		}
	case Refreshed:
		if err := c.wireBeanAfterRefreshed(result, stack); err != nil {
			return err
		}
	default:
		return fmt.Errorf("state is error for wiring")
	}

	v.Set(result.Value())
	return nil
}

// wireBean 对 bean 进行属性绑定和依赖注入，同时追踪其注入路径。如果 bean 有初始
// 化函数，则在注入完成之后执行其初始化函数。如果 bean 依赖了其他 bean，则首先尝试
// 实例化被依赖的 bean 然后对它们进行注入。
func (c *Container) wireBeanInRefreshing(b *gs.BeanDefinition, stack *wiringStack) error {

	if b.GetStatus() == gs.Deleted {
		return fmt.Errorf("bean:%q have been deleted", b.ID())
	}

	// 运行时 Get 或者 Wire 会出现下面这种情况。
	if c.state == Refreshed && b.GetStatus() == gs.Wired {
		return nil
	}

	haveDestroy := false

	defer func() {
		if haveDestroy {
			stack.destroyers.Remove(stack.destroyers.Back())
		}
	}()

	// 记录注入路径上的销毁函数及其执行的先后顺序。
	if _, ok := b.Interface().(gs.BeanDestroy); ok || b.GetDestroy() != nil {
		haveDestroy = true
		d := stack.saveDestroyer(b)
		if i := stack.destroyers.Back(); i != nil {
			d.after(i.Value.(*gs.BeanDefinition))
		}
		stack.destroyers.PushBack(b)
	}

	stack.pushBack(b)

	if b.GetStatus() == gs.Creating && b.Callable() != nil {
		prev := stack.beans[len(stack.beans)-2]
		if prev.GetStatus() == gs.Creating {
			return errors.New("found circle autowire")
		}
	}

	if b.GetStatus() >= gs.Creating {
		stack.popBack()
		return nil
	}

	b.SetStatus(gs.Creating)

	// 对当前 bean 的间接依赖项进行注入。
	for _, s := range b.GetDepends() {
		beans, err := c.Find(s)
		if err != nil {
			return err
		}
		for _, d := range beans {
			err = c.wireBeanInRefreshing(d, stack)
			if err != nil {
				return err
			}
		}
	}

	v, err := c.getBeanValue(b, stack)
	if err != nil {
		return err
	}

	b.SetStatus(gs.Created)

	t := v.Type()
	for _, typ := range b.GetExports() {
		if !t.Implements(typ) {
			return fmt.Errorf("%s doesn't implement interface %s", b, typ)
		}
	}

	err = c.wireBeanValue(v, t, stack)
	if err != nil {
		return err
	}

	// 如果 bean 有初始化函数，则执行其初始化函数。
	if b.GetInit() != nil {
		fnValue := reflect.ValueOf(b.GetInit())
		out := fnValue.Call([]reflect.Value{b.Value()})
		if len(out) > 0 && !out[0].IsNil() {
			return out[0].Interface().(error)
		}
	}

	// 如果 bean 实现了 BeanInit 接口，则执行其 OnInit 方法。
	if f, ok := b.Interface().(gs.BeanInit); ok {
		if err = f.OnInit(c); err != nil {
			return err
		}
	}

	// 如果 bean 实现了 dync.Refreshable 接口，则将 bean 添加到可刷新对象列表中。
	if b.IsRefreshEnable() {
		i := b.Interface().(gs.Refreshable)
		refreshParam := b.GetRefreshParam()
		watch := c.state == Refreshing
		if err = c.p.RefreshBean(i, refreshParam, watch); err != nil {
			return err
		}
	}

	b.SetStatus(gs.Wired)
	stack.popBack()
	return nil
}

func (c *Container) wireBeanAfterRefreshed(b SimpleBean, stack *wiringStack) error {

	v, err := c.getBeanValue(b, stack)
	if err != nil {
		return err
	}

	t := v.Type()
	err = c.wireBeanValue(v, t, stack)
	if err != nil {
		return err
	}

	// 如果 bean 实现了 BeanInit 接口，则执行其 OnInit 方法。
	if f, ok := b.Interface().(gs.BeanInit); ok {
		if err = f.OnInit(c); err != nil {
			return err
		}
	}

	return nil
}

type argContext struct {
	c     *Container
	stack *wiringStack
}

func (a *argContext) Matches(c gs.Condition) (bool, error) {
	return c.Matches(a.c)
}

func (a *argContext) Bind(v reflect.Value, tag string) error {
	return a.c.p.Data().Bind(v, conf.Tag(tag))
}

func (a *argContext) Wire(v reflect.Value, tag string) error {
	return a.c.wireByTag(v, tag, a.stack)
}

// getBeanValue 获取 bean 的值，如果是构造函数 bean 则执行其构造函数然后返回执行结果。
func (c *Container) getBeanValue(b SimpleBean, stack *wiringStack) (reflect.Value, error) {

	if b.Callable() == nil {
		return b.Value(), nil
	}

	out, err := b.Callable().Call(&argContext{c: c, stack: stack})
	if err != nil {
		return reflect.Value{}, err /* fmt.Errorf("%s:%s return error: %v", b.getClass(), b.ID(), err) */
	}

	// 构造函数的返回值为值类型时 b.Type() 返回其指针类型。
	if val := out[0]; util.IsBeanType(val.Type()) {
		// 如果实现接口的是值类型，那么需要转换成指针类型然后再赋值给接口。
		if !val.IsNil() && val.Kind() == reflect.Interface && util.IsValueType(val.Elem().Type()) {
			v := reflect.New(val.Elem().Type())
			v.Elem().Set(val.Elem())
			b.Value().Set(v)
		} else {
			b.Value().Set(val)
		}
	} else {
		b.Value().Elem().Set(val)
	}

	if b.Value().IsNil() {
		return reflect.Value{}, fmt.Errorf("%s return nil", b.String()) // b.GetClass(), b.FileLine())
	}

	v := b.Value()
	// 结果以接口类型返回时需要将原始值取出来才能进行注入。
	if b.Type().Kind() == reflect.Interface {
		v = v.Elem()
	}
	return v, nil
}

// wireBeanValue 对 v 进行属性绑定和依赖注入，v 在传入时应该是一个已经初始化的值。
func (c *Container) wireBeanValue(v reflect.Value, t reflect.Type, stack *wiringStack) error {

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
		t = t.Elem()
	}

	// 如整数指针类型的 bean 是无需注入的。
	if v.Kind() != reflect.Struct {
		return nil
	}

	typeName := t.Name()
	if typeName == "" { // 简单类型没有名字
		typeName = t.String()
	}

	param := conf.BindParam{Path: typeName}
	return c.wireStruct(v, t, param, stack)
}

// wireStruct 对结构体进行依赖注入，需要注意的是这里不需要进行属性绑定。
func (c *Container) wireStruct(v reflect.Value, t reflect.Type, opt conf.BindParam, stack *wiringStack) error {

	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		fv := v.Field(i)

		if !fv.CanInterface() {
			fv = util.PatchValue(fv)
			if !fv.CanInterface() {
				continue
			}
		}

		fieldPath := opt.Path + "." + ft.Name

		// 支持 autowire 和 inject 两个标签。
		tag, ok := ft.Tag.Lookup("autowire")
		if !ok {
			tag, ok = ft.Tag.Lookup("inject")
		}
		if ok {
			if strings.HasSuffix(tag, ",lazy") {
				f := lazyField{v: fv, path: fieldPath, tag: tag}
				stack.lazyFields = append(stack.lazyFields, f)
			} else {
				if ft.Type == GsContextType {
					c.ContextAware = true
				}
				if err := c.wireByTag(fv, tag, stack); err != nil {
					return fmt.Errorf("%q wired error: %w", fieldPath, err)
				}
			}
			continue
		}

		subParam := conf.BindParam{
			Key:  opt.Key,
			Path: fieldPath,
		}

		if tag, ok = ft.Tag.Lookup("value"); ok {
			if err := subParam.BindTag(tag, ft.Tag); err != nil {
				return err
			}
			if ft.Anonymous {
				err := c.wireStruct(fv, ft.Type, subParam, stack)
				if err != nil {
					return err
				}
			} else {
				watch := c.state == Refreshing
				err := c.p.RefreshField(fv.Addr(), subParam, watch)
				if err != nil {
					return err
				}
			}
			continue
		}

		if ft.Anonymous && ft.Type.Kind() == reflect.Struct {
			if err := c.wireStruct(fv, ft.Type, subParam, stack); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Container) wireByTag(v reflect.Value, tag string, stack *wiringStack) error {

	tag, err := c.resolveTag(tag)
	if err != nil {
		return err
	}

	if tag == "" {
		return c.autowire(v, nil, false, stack)
	}

	var tags []wireTag
	if tag != "?" {
		for _, s := range strings.Split(tag, ",") {
			var g wireTag
			g, err = c.toWireTag(s)
			if err != nil {
				return err
			}
			tags = append(tags, g)
		}
	}
	return c.autowire(v, tags, tag == "?", stack)
}
