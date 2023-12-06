package compile

import (
	"encoding/binary"
	"errors"
	"sync/atomic"
)

type CompilerEnvironment struct {
	instr []Instr
}

type SymbolTable struct {
	symbolCount      uint64
	symbolMap        map[string]uint64
	reverseSymbolMap map[uint64]string
}

var symbolTable = &SymbolTable{
	symbolCount:      0,
	symbolMap:        map[string]uint64{},
	reverseSymbolMap: map[uint64]string{},
}

func (s *SymbolTable) GetSymbolCount() uint64 {
	return s.symbolCount
}

func (s *SymbolTable) GetSymbol(symbol string) uint64 {
	if symbolId, ok := s.symbolMap[symbol]; ok {
		return symbolId
	}
	incrementCount := atomic.AddUint64(&s.symbolCount, 1)
	s.symbolMap[symbol] = incrementCount
	s.reverseSymbolMap[incrementCount] = symbol
	return incrementCount
}

func (s *SymbolTable) GetSymbolById(symbolId uint64) string {
	if val, ok := s.reverseSymbolMap[symbolId]; ok {
		return val
	}
	return ""
}

var _compileEnvIndex uint64 = 0

var _compileEnvRelation = map[uint64]*struct {
	hasParent             bool
	parent                uint64
	table                 map[uint64]uint64
	tableSymbolToSymbolId map[uint64]uint64
	tableSymbolCount      uint64
}{
	0: {
		hasParent:             false,
		parent:                0,
		table:                 map[uint64]uint64{},
		tableSymbolToSymbolId: map[uint64]uint64{},
		tableSymbolCount:      0,
	},
}

func NewCompileEnvironment() *CompilerEnvironment {
	env := &CompilerEnvironment{
		instr: []Instr{},
	}
	return env
}

func (c *CompilerEnvironment) SetInstr(instr []Instr) {
	c.instr = instr
}

func (c *CompilerEnvironment) Compile(sexp SExpression) error {
	stack, _, err := GenerateOpCode(c, sexp, 0, 0)
	if err != nil {
		return err
	}
	c.instr = stack
	return nil
}

func (c *CompilerEnvironment) Serialize() []byte {

	body := Serialize(c.instr)

	totalSymbolSize := uint64(8)

	for i := uint64(0); i < symbolTable.symbolCount+1; i++ {
		totalSymbolSize += 8 + uint64(len(symbolTable.reverseSymbolMap[i]))
	}

	b := make([]byte, totalSymbolSize+uint64(len(body)))
	binary.LittleEndian.PutUint64(b, symbolTable.symbolCount)
	byteIndex := uint64(8)

	for i := uint64(0); i < symbolTable.symbolCount+1; i++ {
		//symbolBody := make([]byte, 8+uint64(len(c.reverseSymbolMap[i])))
		symbolBody := b[byteIndex:]
		binary.LittleEndian.PutUint64(symbolBody, uint64(len(symbolTable.reverseSymbolMap[i])))
		byteIndex += 8
		copy(b[8:], symbolTable.reverseSymbolMap[i])
		byteIndex += uint64(len(symbolTable.reverseSymbolMap[i]))
	}

	copy(b[byteIndex:], body)

	return b
}

func DeserializeCompileEnvironment(b []byte) *CompilerEnvironment {
	c := &CompilerEnvironment{}
	symbolTable.symbolCount = binary.LittleEndian.Uint64(b)
	byteIndex := uint64(8)

	for i := uint64(0); i < symbolTable.symbolCount+1; i++ {
		symbolLen := binary.LittleEndian.Uint64(b[byteIndex:])
		byteIndex += 8
		symbolTable.reverseSymbolMap[i] = string(b[byteIndex : byteIndex+symbolLen])
		byteIndex += symbolLen
	}

	c.instr = DeserializeInstructions(b[byteIndex:])
	return c
}

func (c *CompilerEnvironment) GetInstr() []Instr {
	return c.instr
}

func (c *CompilerEnvironment) GetCompilerSymbol(symbol string) uint64 {
	return symbolTable.GetSymbol(symbol)
}

func (c *CompilerEnvironment) GetCompilerSymbolString(symbol uint64) string {
	return symbolTable.GetSymbolById(symbol)
}

func (c *CompilerEnvironment) GetNewEnvironmentIndex() uint64 {
	return atomic.AddUint64(&_compileEnvIndex, 1)
}

func (c *CompilerEnvironment) FindSymbolIndexInEnvironment(env uint64, symbol uint64) (uint64, uint64, error) {

	currentEnvId := env

	for {
		if _, ok := _compileEnvRelation[currentEnvId]; !ok {
			return 0, 0, errors.New("env not found")
		}
		reR := _compileEnvRelation[currentEnvId]
		if index, ok := reR.tableSymbolToSymbolId[symbol]; ok {
			return currentEnvId, index, nil
		}
		if reR.hasParent == false {
			return 0, 0, errors.New("symbol not found")
		}
		if currentEnvId == 0 {
			return 0, 0, errors.New("symbol not found")
		}
		currentEnvId = _compileEnvRelation[currentEnvId].parent
	}
}

func (c *CompilerEnvironment) FindSymbolInEnvironment(env uint64, symbolIndex uint64) (uint64, error) {

	currentEnvId := env

	for {
		if _, ok := _compileEnvRelation[currentEnvId]; !ok {
			return 0, errors.New("env not found")
		}
		reR := _compileEnvRelation[currentEnvId]
		if symId, ok := reR.table[symbolIndex]; ok {
			return symId, nil
		}
		if reR.hasParent == false {
			return 0, errors.New("symbol not found")
		}
		if currentEnvId == 0 {
			return 0, errors.New("symbol not found")
		}
		currentEnvId = _compileEnvRelation[currentEnvId].parent
	}
}

func (c *CompilerEnvironment) AddEnvironmentToEnvironment(parentEnv uint64, env uint64) {
	if _, ok := _compileEnvRelation[env]; !ok {
		_compileEnvRelation[env] = &struct {
			hasParent             bool
			parent                uint64
			table                 map[uint64]uint64
			tableSymbolToSymbolId map[uint64]uint64
			tableSymbolCount      uint64
		}{parent: parentEnv, table: map[uint64]uint64{}, hasParent: true, tableSymbolCount: 0, tableSymbolToSymbolId: map[uint64]uint64{}}
	}
}

func (c *CompilerEnvironment) AddSymbolToEnvironment(env uint64, symbol uint64) uint64 {
	index := _compileEnvRelation[env].tableSymbolCount
	_compileEnvRelation[env].table[index] = symbol
	_compileEnvRelation[env].tableSymbolToSymbolId[symbol] = index
	_compileEnvRelation[env].tableSymbolCount = index + 1
	return index
}
