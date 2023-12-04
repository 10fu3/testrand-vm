package reader

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

type SExpression interface {
	TypeId() string
	SExpressionTypeId() SExpressionType
	String() string
	IsList() bool
	Equals(sexp SExpression) bool
}

type Symbol interface {
	SExpression
	GetValue() string
}

type symbol struct {
	name string
}

func (s *symbol) SExpressionTypeId() SExpressionType {
	return SExpressionTypeSymbol
}

func (s *symbol) TypeId() string {
	return "symbol"
}

func (s *symbol) IsList() bool {
	return false
}

func (s *symbol) String() string {
	return s.name
}

func (s *symbol) Equals(sexp SExpression) bool {
	if sexp.TypeId() != "symbol" {
		return false
	}
	return s.name == (sexp).(Symbol).GetValue()
}

func (s *symbol) GetValue() string {
	return s.name
}

var internedSymbols = make(map[string]Symbol)
var internedSymbolRWLock = sync.RWMutex{}

func NewSymbol(sym string) Symbol {
	internedSymbolRWLock.RLock()
	if interned, ok := internedSymbols[sym]; ok {
		internedSymbolRWLock.RUnlock()
		return interned
	}
	internedSymbolRWLock.RUnlock()
	internedSymbolRWLock.Lock()
	defer internedSymbolRWLock.Unlock()
	newSymbol := &symbol{name: sym}
	internedSymbols[sym] = newSymbol
	return newSymbol
}

type _int int64

func (i _int) GetValue() int64 {
	return int64(i)
}

func (i _int) String() string {
	return strconv.FormatInt(int64(i), 10)
}

func (i _int) SExpressionTypeId() SExpressionType {
	return SExpressionTypeNumber
}

func (i _int) TypeId() string {
	return "number"
}

func (i _int) IsList() bool {
	return false
}

func (i _int) Equals(sexp SExpression) bool {
	if "number" != sexp.TypeId() {
		return false
	}
	return i.GetValue() == sexp.(Number).GetValue()
}

type Number interface {
	GetValue() int64
	String() string
	SExpression
}

func NewInt(val int64) Number {
	return _int(val)
}

type Bool interface {
	GetValue() bool
	String() string
	SExpression
}

type _bool bool

func (b _bool) Equals(sexp SExpression) bool {
	if b.SExpressionTypeId() != sexp.SExpressionTypeId() {
		return false
	}

	return b == sexp.(Bool)
}

func (b _bool) GetValue() bool {
	return bool(b)
}

func (b _bool) String() string {
	if b {
		return "#t"
	}
	return "#f"
}

func (b _bool) TypeId() string {
	return "bool"
}

func (b _bool) SExpressionTypeId() SExpressionType {
	return SExpressionTypeBool
}

func (b _bool) IsList() bool {
	return false
}

func NewBool(b bool) Bool {
	return _bool(b)
}

type Str interface {
	GetValue() string
	String() string
	SExpression
}

type _string string

func NewString(s string) Str {
	return _string(s)
}

func (s _string) Equals(sexp SExpression) bool {
	if "string" != sexp.TypeId() {
		return false
	}
	return s == sexp.(Str)
}

func (s _string) GetValue() string {
	return string(s)
}

func (s _string) String() string {
	return fmt.Sprintf("\"%s\"", string(s))
}

func (s _string) TypeId() string {
	return "string"
}

func (s _string) SExpressionTypeId() SExpressionType {
	return SExpressionTypeString
}

func (s _string) IsList() bool {
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

func (n *_nil) String() string {
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
	Car SExpression
	Cdr SExpression
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

func JoinList(left, right SExpression) (ConsCell, error) {

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

func (cell *_cons_cell) String() string {
	if "symbol" == cell.Car.TypeId() && "quote" == ((cell.Car).(Symbol)).GetValue() && cell.Cdr.TypeId() == "cons_cell" && "nil" == ((cell.Cdr).(ConsCell)).GetCdr().TypeId() {
		return fmt.Sprintf("'%s", ((cell.Cdr).(ConsCell)).GetCar().String())
	}
	var joinedString strings.Builder
	joinedString.WriteString("(")
	var lookCell ConsCell = cell

	for {
		if lookCell.GetCar().TypeId() != "nil" {
			joinedString.WriteString(lookCell.GetCar().String())
			if lookCell.GetCdr().TypeId() == "cons_cell" {
				if lookCell.GetCdr().(ConsCell).GetCar().TypeId() != "nil" && lookCell.GetCdr().(ConsCell).GetCdr().TypeId() != "nil" {
					joinedString.WriteString(" ")
				}
			}
		}

		if lookCell.GetCdr().TypeId() != "cons_cell" {
			if lookCell.GetCdr().TypeId() != "nil" {
				joinedString.WriteString(" . " + lookCell.GetCdr().String())
			}
			joinedString.WriteString(")")
			break
		}
		lookCell = (lookCell.GetCdr()).(ConsCell)
	}
	return joinedString.String()
}

func ToArray(sexp SExpression) ([]SExpression, error) {
	list := make([]SExpression, 0)
	look := sexp

	for !IsEmptyList(look) {
		if "cons_cell" != look.TypeId() {
			return nil, errors.New("need list")
		}
		if look.(ConsCell).GetCdr().TypeId() != "cons_cell" {
			list = append(list, NewConsCell(look.(ConsCell).GetCar(), look.(ConsCell).GetCdr()))
			return list, nil
		}
		list = append(list, look.(ConsCell).GetCar())
		look = look.(ConsCell).GetCdr()
	}
	return list, nil
}

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

func ToConsCell(list []SExpression) ConsCell {
	var head = (NewConsCell(NewNil(), NewNil())).(*_cons_cell)
	var look = head
	var beforeLook *_cons_cell = nil

	for _, sexp := range list {
		look.Car = sexp
		look.Cdr = NewConsCell(NewNil(), NewNil())
		beforeLook = look
		look = (look.Cdr).(*_cons_cell)
	}
	if beforeLook != nil {
		beforeLook.Cdr = NewConsCell(NewNil(), NewNil())
	}
	return head
}

func IsEmptyList(list SExpression) bool {
	if "cons_cell" != list.TypeId() {
		return false
	}
	tmp := (list).(ConsCell)

	return "nil" == tmp.GetCar().TypeId() && "nil" == tmp.GetCdr().TypeId()
}

type NativeArray struct {
	elements []SExpression
}

func (a *NativeArray) TypeId() string {
	return "native_array"
}

func (a *NativeArray) SExpressionTypeId() SExpressionType {
	return SExpressionTypeNativeArray
}

func (a *NativeArray) String() string {
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

func NewNativeArray(elements []SExpression) *NativeArray {
	return &NativeArray{elements: elements}
}

type NativeHashMap struct {
	elements map[string]SExpression
}

func (h *NativeHashMap) TypeId() string {
	return "native_hashmap"
}

func (h *NativeHashMap) SExpressionTypeId() SExpressionType {
	return SExpressionTypeNativeHashmap
}

func (h *NativeHashMap) String() string {
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

func (h *NativeHashMap) Get(key string) (SExpression, bool) {
	if val, ok := h.elements[key]; ok {
		return val, true
	}
	return nil, false
}

func (h *NativeHashMap) Set(key string, value SExpression) {
	h.elements[key] = value
}

func (h *NativeHashMap) Length() int64 {
	return int64(len(h.elements))
}

func (h *NativeHashMap) Delete(key string) {
	delete(h.elements, key)
}

func (h *NativeHashMap) Keys() []SExpression {
	keys := make([]SExpression, len(h.elements))
	i := 0
	for k := range h.elements {
		keys[i] = NewString(k)
		i++
	}
	return keys
}

func NewNativeHashmap(elements map[string]SExpression) *NativeHashMap {
	return &NativeHashMap{elements: elements}
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
