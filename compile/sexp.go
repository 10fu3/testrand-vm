package compile

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type SExpression interface {
	TypeId() string
	SExpressionTypeId() SExpressionType
	String(compEnv *CompilerEnvironment) string
	IsList() bool
	Equals(sexp SExpression) bool
}

type Symbol uint64

func (s Symbol) SExpressionTypeId() SExpressionType {
	return SExpressionTypeSymbol
}

func (s Symbol) TypeId() string {
	return "symbol"
}

func (s Symbol) IsList() bool {
	return false
}

func (s Symbol) String(compEnv *CompilerEnvironment) string {
	return compEnv.GetCompilerSymbolString(uint64(s))
}

func (s Symbol) Equals(sexp SExpression) bool {
	if sexp.SExpressionTypeId() != SExpressionTypeSymbol {
		return false
	}
	return s == (sexp).(Symbol)
}

func (s Symbol) GetSymbolIndex() uint64 {
	return uint64(s)
}

func NewSymbol(symNum uint64) Symbol {
	return Symbol(symNum)
}

func (i Number) GetValue() int64 {
	return int64(i)
}

func (i Number) String(compEnv *CompilerEnvironment) string {
	return strconv.FormatInt(int64(i), 10)
}

func (i Number) SExpressionTypeId() SExpressionType {
	return SExpressionTypeNumber
}

func (i Number) TypeId() string {
	return "number"
}

func (i Number) IsList() bool {
	return false
}

func (i Number) Equals(sexp SExpression) bool {
	if sexp.SExpressionTypeId() != SExpressionTypeNumber {
		return false
	}
	return i == sexp.(Number)
}

type Number int64

type Bool bool

func (b Bool) Equals(sexp SExpression) bool {
	if b.SExpressionTypeId() != sexp.SExpressionTypeId() {
		return false
	}

	return b == sexp.(Bool)
}

func (b Bool) GetValue() bool {
	return bool(b)
}

func (b Bool) String(compEnv *CompilerEnvironment) string {
	if b {
		return "#t"
	}
	return "#f"
}

func (b Bool) TypeId() string {
	return "bool"
}

func (b Bool) SExpressionTypeId() SExpressionType {
	return SExpressionTypeBool
}

func (b Bool) IsList() bool {
	return false
}

func NewBool(b bool) Bool {
	return Bool(b)
}

type Str uint64

func NewString(s uint64) Str {
	return Str(s)
}

func (s Str) Equals(sexp SExpression) bool {
	if sexp.SExpressionTypeId() != SExpressionTypeString {
		return false
	}
	return s == sexp.(Str)
}

func (s Str) GetValue(compEnv *CompilerEnvironment) string {
	return compEnv.GetCompilerSymbolString(uint64(s))
}

func (s Str) GetSymbolIndex() uint64 {
	return uint64(s)
}

func (s Str) String(compEnv *CompilerEnvironment) string {
	return fmt.Sprintf("\"%s\"", compEnv.GetCompilerSymbolString(uint64(s)))
}

func (s Str) TypeId() string {
	return "string"
}

func (s Str) SExpressionTypeId() SExpressionType {
	return SExpressionTypeString
}

func (s Str) IsList() bool {
	return false
}

type Nil struct{}

func (n Nil) Equals(sexp SExpression) bool {
	return "nil" == sexp.TypeId()
}

func (n Nil) TypeId() string {
	return "nil"
}

func (n Nil) SExpressionTypeId() SExpressionType {
	return SExpressionTypeNil
}

func (n Nil) String(compEnv *CompilerEnvironment) string {
	return "#nil"
}

func (n Nil) IsList() bool {
	return false
}

var _nilInstance = struct{}{}

func NewNil() Nil {
	return _nilInstance
}

type ConsCell struct {
	Car     SExpression
	Cdr     SExpression
	compEnv *CompilerEnvironment
}

func (cell ConsCell) Equals(sexp SExpression) bool {
	if SExpressionTypeConsCell != sexp.SExpressionTypeId() {
		return false
	}
	c := sexp.(ConsCell)
	if !cell.Car.Equals(c.GetCar()) {
		return false
	}
	return !cell.Cdr.Equals(c.GetCdr())
}

func NewConsCell(car SExpression, cdr SExpression) ConsCell {
	return ConsCell{
		Car: car,
		Cdr: cdr,
	}
}

//func JoinList(compEnv *CompilerEnvironment, left, right SExpression) (ConsCell, error) {
//
//	if !left.IsList() {
//		return nil, errors.New("left is not a list")
//	}
//
//	if !right.IsList() {
//		return nil, errors.New("right is not a list")
//	}
//
//	baseRoot := left.(*_cons_cell)
//	baseLook := baseRoot
//
//	copyRoot := &_cons_cell{
//		Car: NewNil(),
//		Cdr: NewNil(),
//	}
//	copyLook := copyRoot
//
//	for {
//		if IsEmptyList(baseLook.GetCdr()) {
//			copyLook.Car = baseLook.GetCar()
//			copyLook.Cdr = right
//			return copyRoot, nil
//		} else {
//			copyLook.Car = baseLook.GetCar()
//			copyLook.Cdr = NewConsCell(NewNil(), NewNil())
//			copyLook = copyLook.Cdr.(*_cons_cell)
//			baseLook = baseLook.GetCdr().(*_cons_cell)
//		}
//	}
//}

func (cell ConsCell) TypeId() string {
	return "cons_cell"
}

func (cell ConsCell) SExpressionTypeId() SExpressionType {
	return SExpressionTypeConsCell
}

func (cell ConsCell) String(compEnv *CompilerEnvironment) string {
	if SExpressionTypeSymbol == cell.Car.SExpressionTypeId() &&
		cell.compEnv.GetCompilerSymbol("quote") == ((cell.Car).(Symbol)).GetSymbolIndex() &&
		SExpressionTypeConsCell == cell.Cdr.SExpressionTypeId() &&
		SExpressionTypeNil == ((cell.Cdr).(ConsCell)).GetCdr().SExpressionTypeId() {
		return fmt.Sprintf("'%s", ((cell.Cdr).(ConsCell)).GetCar().String(compEnv))
	}
	var joinedString strings.Builder
	joinedString.WriteString("(")
	var lookCell = cell

	for {
		if SExpressionTypeNil != lookCell.GetCar().SExpressionTypeId() {
			joinedString.WriteString(lookCell.GetCar().String(compEnv))
			if SExpressionTypeConsCell == lookCell.GetCdr().SExpressionTypeId() {
				if SExpressionTypeNil != lookCell.GetCdr().(ConsCell).GetCar().SExpressionTypeId() &&
					SExpressionTypeNil != lookCell.GetCdr().(ConsCell).GetCdr().SExpressionTypeId() {
					joinedString.WriteString(" ")
				}
			}
		}

		if SExpressionTypeConsCell != lookCell.GetCdr().SExpressionTypeId() {
			if SExpressionTypeNil != lookCell.GetCdr().SExpressionTypeId() {
				joinedString.WriteString(" . " + lookCell.GetCdr().String(compEnv))
			}
			joinedString.WriteString(")")
			break
		}
		lookCell = (lookCell.GetCdr()).(ConsCell)
	}
	return joinedString.String()
}

//func ToArray(sexp SExpression) ([]SExpression, error) {
//	list := make([]SExpression, 0)
//	look := sexp
//
//	for !IsEmptyList(look) {
//		if "cons_cell" != look.TypeId() {
//			return nil, errors.New("need list")
//		}
//		if look.(ConsCell).GetCdr().TypeId() != "cons_cell" {
//			list = append(list, NewConsCell(look.(ConsCell).GetCar(), look.(ConsCell).GetCdr()))
//			return list, nil
//		}
//		list = append(list, look.(ConsCell).GetCar())
//		look = look.(ConsCell).GetCdr()
//	}
//	return list, nil
//}

func (cell ConsCell) IsList() bool {
	if SExpressionTypeConsCell != cell.Cdr.SExpressionTypeId() {
		if IsEmptyList(cell.Cdr) {
			return true
		}
		return cell.Cdr.IsList()
	}
	return false
}

func (cell ConsCell) GetCar() SExpression {
	return cell.Car
}

func (cell ConsCell) GetCdr() SExpression {
	return cell.Cdr
}

//func ToConsCell(list []SExpression) ConsCell {
//	var head = (NewConsCell(NewNil(), NewNil())).(*_cons_cell)
//	var look = head
//	var beforeLook *_cons_cell = nil
//
//	for _, sexp := range list {
//		look.Car = sexp
//		look.Cdr = NewConsCell(NewNil(), NewNil())
//		beforeLook = look
//		look = (look.Cdr).(*_cons_cell)
//	}
//	if beforeLook != nil {
//		beforeLook.Cdr = NewConsCell(NewNil(), NewNil())
//	}
//	return head
//}

func IsEmptyList(list SExpression) bool {
	if SExpressionTypeConsCell != list.SExpressionTypeId() {
		return false
	}
	tmp := (list).(ConsCell)

	return SExpressionTypeNil == tmp.GetCar().SExpressionTypeId() && SExpressionTypeNil == tmp.GetCdr().SExpressionTypeId()
}

type NativeArray struct {
	elements []SExpression
	compEnv  *CompilerEnvironment
}

func (a *NativeArray) TypeId() string {
	return "native_array"
}

func (a *NativeArray) SExpressionTypeId() SExpressionType {
	return SExpressionTypeNativeArray
}

func (a *NativeArray) String(compEnv *CompilerEnvironment) string {
	var joinedString strings.Builder
	joinedString.WriteString("[")
	for i, elm := range a.elements {
		if i != 0 {
			joinedString.WriteString(" ")
		}
		joinedString.WriteString(fmt.Sprintf("%v", elm))
	}
	joinedString.WriteString("]")
	return joinedString.String()
}

func (a *NativeArray) IsList() bool {
	return false
}

func (a *NativeArray) Equals(sexp SExpression) bool {
	if "native_array" != sexp.TypeId() {
		return false
	}
	return a == sexp.(*NativeArray)
}

func (a *NativeArray) Get(index int64) SExpression {
	return a.elements[index]
}

func (a *NativeArray) Length() int64 {
	return int64(len(a.elements))
}

func (a *NativeArray) Set(index int64, value SExpression) error {
	if index < 0 || index >= int64(len(a.elements)) {
		return errors.New("index out of bounds")
	}
	a.elements[index] = value
	return nil
}

func (a *NativeArray) Push(value SExpression) {
	a.elements = append(a.elements, value)
}

func NewNativeArray(compEnv *CompilerEnvironment, elements []SExpression) *NativeArray {
	return &NativeArray{elements: elements, compEnv: compEnv}
}

type NativeHashMap struct {
	elements map[uint64]SExpression
	compEnv  *CompilerEnvironment
}

func (h *NativeHashMap) TypeId() string {
	return "native_hashmap"
}

func (h *NativeHashMap) SExpressionTypeId() SExpressionType {
	return SExpressionTypeNativeHashmap
}

func (h *NativeHashMap) String(compEnv *CompilerEnvironment) string {
	var joinedString strings.Builder
	joinedString.WriteString("{\n")
	i := 0
	for k, elm := range h.elements {
		if i != 0 {
			joinedString.WriteString(" ")
		}
		joinedString.WriteString(fmt.Sprintf("%s: %v,\n", compEnv.GetCompilerSymbolString(k), elm.String(compEnv)))
		i++
	}
	joinedString.WriteString("}")
	return joinedString.String()
}

func (h *NativeHashMap) IsList() bool {
	return false
}

func (h *NativeHashMap) Equals(sexp SExpression) bool {
	if "native_hashmap" != sexp.TypeId() {
		return false
	}
	return h == sexp.(*NativeHashMap)
}

func (h *NativeHashMap) Get(key uint64) (SExpression, bool) {
	if val, ok := h.elements[key]; ok {
		return val, true
	}
	return nil, false
}

func (h *NativeHashMap) Set(key uint64, value SExpression) {
	h.elements[key] = value
}

func (h *NativeHashMap) Length() int64 {
	return int64(len(h.elements))
}

func (h *NativeHashMap) Delete(key uint64) {
	delete(h.elements, key)
}

//func (h *NativeHashMap) Keys() []SExpression {
//	keys := make([]SExpression, len(h.elements))
//	i := 0
//	for k := range h.elements {
//		keys[i] = NewString(h.compEnv, k)
//		i++
//	}
//	return keys
//}

func NewNativeHashmap(compEnv *CompilerEnvironment, elements map[uint64]SExpression) *NativeHashMap {
	return &NativeHashMap{elements: elements, compEnv: compEnv}
}

type NativeValue struct {
	Value interface{}
}

func (v NativeValue) TypeId() string {
	return "native_value"
}

func (v NativeValue) SExpressionTypeId() SExpressionType {
	return SExpressionTypeNativeValue
}

func (v NativeValue) String(compEnv *CompilerEnvironment) string {
	return fmt.Sprintf("%v", v.Value)
}

func (v NativeValue) IsList() bool {
	return false
}

func (v NativeValue) Equals(sexp SExpression) bool {
	tmp, ok := sexp.(NativeValue)
	if !ok {
		return false
	}
	return v.Value == tmp.Value
}

func NewNativeValue[T any](value T) NativeValue {
	return NativeValue{Value: value}
}

type SExpressionType int

const (
	SExpressionTypeSymbol SExpressionType = iota
	SExpressionTypeNumber
	SExpressionTypeBool
	SExpressionTypeString
	SExpressionTypeNil
	SExpressionTypeConsCell
	SExpressionTypeClosure
	SExpressionTypeNativeHashmap
	SExpressionTypeNativeArray
	SExpressionTypeEnvironment
	SExpressionTypeNativeValue
)
