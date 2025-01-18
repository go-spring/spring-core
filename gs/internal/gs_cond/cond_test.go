package gs_cond

import (
	"fmt"
	"testing"

	"github.com/go-spring/spring-core/gs/internal/gs"
	"github.com/go-spring/spring-core/util/assert"
)

func TestCondition(t *testing.T) {
	var c gs.Condition

	c = &onProperty{name: "a", havingValue: "x", matchIfMissing: true}
	assert.Equal(t, fmt.Sprint(c), `OnProperty(name=a, havingValue=x, matchIfMissing=true)`)

	c = &onMissingProperty{name: "a"}
	assert.Equal(t, fmt.Sprint(c), `OnMissingProperty(name=a)`)

	c = &onBean{selector: "a"}
	assert.Equal(t, fmt.Sprint(c), `OnBean(selector=a)`)

	c = &onBean{selector: new(error)}
	assert.Equal(t, fmt.Sprint(c), `OnBean(selector=error:)`)

	c = &onMissingBean{selector: "a"}
	assert.Equal(t, fmt.Sprint(c), `OnMissingBean(selector=a)`)

	c = &onMissingBean{selector: new(error)}
	assert.Equal(t, fmt.Sprint(c), `OnMissingBean(selector=error:)`)

	c = &onSingleBean{selector: "a"}
	assert.Equal(t, fmt.Sprint(c), `OnSingleBean(selector=a)`)

	c = &onSingleBean{selector: new(error)}
	assert.Equal(t, fmt.Sprint(c), `OnSingleBean(selector=error:)`)

	c = &onExpression{expression: "a"}
	assert.Equal(t, fmt.Sprint(c), `OnExpression(expression=a)`)

	c = &onFunc{fn: func(ctx gs.CondContext) (bool, error) { return false, nil }}
	assert.Equal(t, fmt.Sprint(c), `OnFunc(fn=gs_cond.TestCondition.func1)`)

	c = Not(&onBean{selector: "a"})
	assert.Equal(t, fmt.Sprint(c), `Not(OnBean(selector=a))`)

	c = Or(&onBean{selector: "a"})
	assert.Equal(t, fmt.Sprint(c), `OnBean(selector=a)`)

	c = Or(&onBean{selector: "a"}, &onBean{selector: "b"})
	assert.Equal(t, fmt.Sprint(c), `Or(OnBean(selector=a), OnBean(selector=b))`)

	c = And(&onBean{selector: "a"})
	assert.Equal(t, fmt.Sprint(c), `OnBean(selector=a)`)

	c = And(&onBean{selector: "a"}, &onBean{selector: "b"})
	assert.Equal(t, fmt.Sprint(c), `And(OnBean(selector=a), OnBean(selector=b))`)

	c = None(&onBean{selector: "a"})
	assert.Equal(t, fmt.Sprint(c), `Not(OnBean(selector=a))`)

	c = None(&onBean{selector: "a"}, &onBean{selector: "b"})
	assert.Equal(t, fmt.Sprint(c), `None(OnBean(selector=a), OnBean(selector=b))`)

	c = And(
		&onBean{selector: "a"},
		Or(
			&onBean{selector: "b"},
			Not(&onBean{selector: "c"}),
		),
	)
	assert.Equal(t, fmt.Sprint(c), `And(OnBean(selector=a), Or(OnBean(selector=b), Not(OnBean(selector=c))))`)

	c = OnBean("a").And().OnProperty("a")
	assert.Equal(t, fmt.Sprint(c), `Conditional(cond=OnBean(selector=a), op=and, next=(cond=OnProperty(name=a, havingValue=, matchIfMissing=false)))`)
}
