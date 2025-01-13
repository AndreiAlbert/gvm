package vm

import (
	"fmt"
	"log"
	"math"
)

type ValueKind byte

const (
	ValueInt32 ValueKind = iota
	ValueFloat32
)

type Value struct {
	kind ValueKind
	raw  uint32 // 4 bytes
}

func (v Value) AsInt32() int32 {
	if v.kind != ValueInt32 {
		log.Fatalf("Value is not int32, its %v\n", v.kind)
	}
	return int32(v.raw)
}

func (v Value) AsFloat32() float32 {
	if v.kind != ValueFloat32 {
		log.Fatalf("Value is not float32, its %v\n", v.kind)
	}
	return math.Float32frombits(v.raw)
}

func Int32Value(val int32) Value {
	return Value{
		kind: ValueInt32,
		raw:  uint32(val),
	}
}

func Float32Value(val float32) Value {
	return Value{
		kind: ValueFloat32,
		raw:  math.Float32bits(val),
	}
}

func (v Value) String() string {
	switch v.kind {
	case ValueInt32:
		return fmt.Sprintf("%d", v.AsInt32())
	case ValueFloat32:
		return fmt.Sprintf("%f", v.AsFloat32())
	default:
		return fmt.Sprintf("<unknown ValueKind %d: raw=0x%08X>", v.kind, v.raw)
	}
}

func Equals(v1, v2 Value) bool {
	if v1.kind != v2.kind {
		log.Fatal("type mismatch for comparison")
	}
	switch v1.kind {
	case ValueFloat32:
		return v1.AsFloat32() == v2.AsFloat32()
	case ValueInt32:
		return v1.AsInt32() == v2.AsInt32()
	default:
		log.Fatal("Unuspported types")
		return false
	}
}

func (v1 Value) LesserOrEqual(v2 Value) bool {
	if v1.kind != v2.kind {
		log.Fatal("Type mismatch for comaprison")
	}
	switch v1.kind {
	case ValueFloat32:
		return v1.AsFloat32() <= v2.AsFloat32()
	case ValueInt32:
		return v1.AsInt32() <= v2.AsInt32()
	default:
		log.Fatal("Unsupported types")
		return false
	}
}

func (v1 Value) Lesser(v2 Value) bool {
	if v1.kind != v2.kind {
		log.Fatal("Type mismatch for comaprison")
	}
	switch v1.kind {
	case ValueFloat32:
		return v1.AsFloat32() < v2.AsFloat32()
	case ValueInt32:
		return v1.AsInt32() < v2.AsInt32()
	default:
		log.Fatal("Unsupported types")
		return false
	}
}
