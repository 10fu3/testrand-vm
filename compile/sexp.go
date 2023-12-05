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

func NewInt(val int64) Number {
	return Number(val)
}

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

type Nil interface {
	SExpression
}

type _nil struct {
}

func (n *_nil) Equals(sexp SExpression) bool {
	return "nil" == sexp.TypeId()
}

func (n *_nil) TypeId() string {
	return "nil"
}

func (n *_nil) SExpressionTypeId() SExpressionType {
	return SExpressionTypeNil
}

func (n *_nil) String(compEnv *CompilerEnvironment) string {
	return "#nil"
}

func (n *_nil) IsList() bool {
	return false
}

var _nilInstance = &_nil{}

func NewNil() Nil {
	return _nilInstance
}

type ConsCell interface {
	SExpression
	GetCar() SExpression
	GetCdr() SExpression
}

type _cons_cell struct {
	Car     SExpression
	Cdr     SExpression
	compEnv *CompilerEnvironment
}

func (cell *_cons_cell) Equals(sexp SExpression) bool {
	if "cons_cell" != sexp.TypeId() {
		return false
	}
	c := sexp.(ConsCell)
	if !cell.Car.Equals(c.GetCar()) {
		return false
	}
	return !cell.Cdr.Equals(c.GetCdr())
}

func NewConsCell(car SExpression, cdr SExpression) ConsCell {
	return &_cons_cell{
		Car: car,
		Cdr: cdr,
	}
}

func JoinList(compEnv *CompilerEnvironment, left, right SExpression) (ConsCell, error) {

	if !left.IsList() {
		return nil, errors.New("left is not a list")
	}

	if !right.IsList() {
		return nil, errors.New("right is not a list")
	}

	baseRoot := left.(*_cons_cell)
	baseLook := baseRoot

	copyRoot := &_cons_cell{
		Car: NewNil(),
		Cdr: NewNil(),
	}
	copyLook := copyRoot

	for {
		if IsEmptyList(baseLook.GetCdr()) {
			copyLook.Car = baseLook.GetCar()
			copyLook.Cdr = right
			return copyRoot, nil
		} else {
			copyLook.Car = baseLook.GetCar()
			copyLook.Cdr = NewConsCell(NewNil(), NewNil())
			copyLook = copyLook.Cdr.(*_cons_cell)
			baseLook = baseLook.GetCdr().(*_cons_cell)
		}
	}
}

func (cell *_cons_cell) TypeId() string {
	return "cons_cell"
}

func (cell *_cons_cell) SExpressionTypeId() SExpressionType {
	return SExpressionTypeConsCell
}

func (cell *_cons_cell) String(compEnv *CompilerEnvironment) string {
	if "symbol" == cell.Car.TypeId() && cell.compEnv.GetCompilerSymbol("quote") == ((cell.Car).(Symbol)).GetSymbolIndex() && cell.Cdr.TypeId() == "cons_cell" && "nil" == ((cell.Cdr).(ConsCell)).GetCdr().TypeId() {
		return fmt.Sprintf("'%s", ((cell.Cdr).(ConsCell)).GetCar().String(compEnv))
	}
	var joinedString strings.Builder
	joinedString.WriteString("(")
	var lookCell ConsCell = cell

	for {
		if lookCell.GetCar().TypeId() != "nil" {
			joinedString.WriteString(lookCell.GetCar().String(compEnv))
			if lookCell.GetCdr().TypeId() == "cons_cell" {
				if lookCell.GetCdr().(ConsCell).GetCar().TypeId() != "nil" && lookCell.GetCdr().(ConsCell).GetCdr().TypeId() != "nil" {
					joinedString.WriteString(" ")
				}
			}
		}

		if lookCell.GetCdr().TypeId() != "cons_cell" {
			if lookCell.GetCdr().TypeId() != "nil" {
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

func (cell *_cons_cell) IsList() bool {
	if "cons_cell" == cell.Cdr.TypeId() {
		if IsEmptyList(cell.Cdr) {
			return true
		}
		return cell.Cdr.IsList()
	}
	return false
}

func (cell *_cons_cell) GetCar() SExpression {
	return cell.Car
}

func (cell *_cons_cell) GetCdr() SExpression {
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
	if "cons_cell" != list.TypeId() {
		return false
	}
	tmp := (list).(ConsCell)

	return "nil" == tmp.GetCar().TypeId() && "nil" == tmp.GetCdr().TypeId()
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
		joinedString.WriteString(fmt.Sprintf("%s: %v\n", k, elm))
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
)
