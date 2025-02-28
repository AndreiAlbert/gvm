package common

import (
	"fmt"
	"log"
	"math"
	"strings"
)

type ValueKind byte

type StructField struct {
	Name       string
	Type       ValueKind
	Offset     uint
	ArrayType  *ValueKind
	StructType string
}

type StructType struct {
	Name    string
	Fields  []StructField
	Size    uint
	Methods map[string]uint
}

const (
	ValueInt32 ValueKind = iota
	ValueFloat32
	ValuePtr
	ValueString
	ValueArray
	ValueVoid
	ValueStruct
	ValueByte
)

type Value struct {
	Kind ValueKind
	Raw  uint32 // 4 bytes
	Ptr  uintptr
}

func (sf StructField) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%s: ", sf.Name))

	if sf.ArrayType != nil {
		sb.WriteString(fmt.Sprintf("array<%v>", *sf.ArrayType))
	} else if sf.StructType != "" {
		sb.WriteString(sf.StructType)
	} else {
		sb.WriteString(fmt.Sprintf("%v", sf.Type))
	}

	sb.WriteString(fmt.Sprintf(" (offset: %d)", sf.Offset))
	return sb.String()
}

func (st StructType) String() string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("struct %s {\n", st.Name))

	// Fields
	for _, field := range st.Fields {
		sb.WriteString(fmt.Sprintf("  %s\n", field.String()))
	}

	// Size
	sb.WriteString(fmt.Sprintf("  size: %d bytes\n", st.Size))

	// Methods
	if len(st.Methods) > 0 {
		sb.WriteString("  methods: {\n")
		for name, addr := range st.Methods {
			sb.WriteString(fmt.Sprintf("    %s: @%d\n", name, addr))
		}
		sb.WriteString("  }\n")
	}

	sb.WriteString("}")
	return sb.String()
}

func (v ValueKind) String() string {
	return [...]string{"int32", "float32", "ptr", "string", "array", "void", "struct"}[v]
}

func (v Value) AsInt32() int32 {
	if v.Kind != ValueInt32 {
		log.Fatalf("Value is not int32, its %v\n", v.Kind)
	}
	return int32(v.Raw)
}

func (v Value) AsFloat32() float32 {
	if v.Kind != ValueFloat32 {
		log.Fatalf("Value is not float32, its %v\n", v.Kind)
	}
	return math.Float32frombits(v.Raw)
}

func (v Value) AsPtr() uintptr {
	if v.Kind != ValuePtr {
		log.Fatalf("Value is not ptr, its %v\n", v.Kind)
	}
	return v.Ptr
}

func (v Value) AsByte() byte {
	if v.Kind != ValueByte {
		log.Fatalf("Value is not byte, its %v\n", v.Kind)
	}
	return byte(v.Raw & 0xFF)
}

func ByteValue(val byte) Value {
	return Value{
		Kind: ValueByte,
		Raw:  uint32(val),
	}
}

func PtrValue(ptr uintptr) Value {
	return Value{
		Kind: ValuePtr,
		Ptr:  ptr,
	}
}

func Int32Value(val int32) Value {
	return Value{
		Kind: ValueInt32,
		Raw:  uint32(val),
	}
}

func Float32Value(val float32) Value {
	return Value{
		Kind: ValueFloat32,
		Raw:  math.Float32bits(val),
	}
}

func (v Value) String() string {
	switch v.Kind {
	case ValueInt32:
		return fmt.Sprintf("%d", v.AsInt32())
	case ValueFloat32:
		return fmt.Sprintf("%f", v.AsFloat32())
	case ValuePtr, ValueString, ValueStruct:
		return fmt.Sprintf("%d", v.AsPtr())
	case ValueByte:
		return fmt.Sprintf("%d", v.AsByte())
	default:
		return fmt.Sprintf("<unknown ValueKind %d: raw=0x%08X>", v.Kind, v.Raw)
	}
}

func Equals(v1, v2 Value) bool {
	if v1.Kind != v2.Kind {
		log.Fatal("type mismatch for comparison")
	}
	switch v1.Kind {
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
	if v1.Kind != v2.Kind {
		log.Fatal("Type mismatch for comaprison")
	}
	switch v1.Kind {
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
	if v1.Kind != v2.Kind {
		log.Fatal("Type mismatch for comaprison")
	}
	switch v1.Kind {
	case ValueFloat32:
		return v1.AsFloat32() < v2.AsFloat32()
	case ValueInt32:
		return v1.AsInt32() < v2.AsInt32()
	default:
		log.Fatal("Unsupported types")
		return false
	}
}
