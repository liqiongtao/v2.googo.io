package goolog

import "testing"

func TestLogger_WithTag(t *testing.T) {

}

type A struct {
	Tag string
}

type B struct {
	Data map[string]any
}

func NewB() *B {
	return &B{
		Data: make(map[string]any),
	}
}

func (b *B) WithFeild(key string, value any) *B {
	data := map[string]any{}
	for k, v := range b.Data {
		data[k] = v
	}
	data[key] = value
	return &B{Data: data}
}
