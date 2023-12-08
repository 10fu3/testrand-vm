package compile

import (
	"errors"
	"sync/atomic"
)

type CompilerEnvironment struct {
	instr               []Instr
	_compileEnvIndex    uint64
	_compileEnvLock     uint32
	_compileEnvRelation map[uint64]*struct {
		hasParent             bool
		parent                uint64
		table                 map[uint64]uint64
		tableSymbolToSymbolId map[uint64]uint64
		children              []uint64
		tableSymbolCount      uint64
	}
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

func NewCompileEnvironment() *CompilerEnvironment {
	env := &CompilerEnvironment{
		instr:            []Instr{},
		_compileEnvIndex: 0,
		_compileEnvLock:  0,
		_compileEnvRelation: map[uint64]*struct {
			hasParent             bool
			parent                uint64
			table                 map[uint64]uint64
			tableSymbolToSymbolId map[uint64]uint64
			children              []uint64
			tableSymbolCount      uint64
		}{
			0: {parent: 0, table: map[uint64]uint64{}, hasParent: false, tableSymbolCount: 0, tableSymbolToSymbolId: map[uint64]uint64{}, children: []uint64{}},
		},
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
	return atomic.AddUint64(&c._compileEnvIndex, 1)
}

func (c *CompilerEnvironment) FindSymbolIndexInEnvironment(env uint64, symbol uint64) (uint64, uint64, error) {

	currentEnvId := env

	for !atomic.CompareAndSwapUint32(&c._compileEnvLock, 0, 1) {
	}
	for {
		if _, ok := c._compileEnvRelation[currentEnvId]; !ok {
			atomic.StoreUint32(&c._compileEnvLock, 0)
			return 0, 0, errors.New("env not found")
		}
		reR := c._compileEnvRelation[currentEnvId]
		if index, ok := reR.tableSymbolToSymbolId[symbol]; ok {
			atomic.StoreUint32(&c._compileEnvLock, 0)
			return currentEnvId, index, nil
		}
		if reR.hasParent == false {
			return 0, 0, errors.New("symbol not found")
		}
		if currentEnvId == 0 {
			return 0, 0, errors.New("symbol not found")
		}
		parent := c._compileEnvRelation[currentEnvId].parent
		currentEnvId = parent
	}
}

func (c *CompilerEnvironment) FindSymbolInEnvironment(env uint64, symbolIndex uint64) (uint64, error) {

	currentEnvId := env

	for {
		for atomic.CompareAndSwapUint32(&c._compileEnvLock, 0, 1) == false {
		}
		if _, ok := c._compileEnvRelation[currentEnvId]; !ok {
			atomic.StoreUint32(&c._compileEnvLock, 0)
			return 0, errors.New("env not found")
		}
		reR := c._compileEnvRelation[currentEnvId]
		if symId, ok := reR.table[symbolIndex]; ok {
			atomic.StoreUint32(&c._compileEnvLock, 0)
			return symId, nil
		}
		if reR.hasParent == false {
			return 0, errors.New("symbol not found")
		}
		if currentEnvId == 0 {
			return 0, errors.New("symbol not found")
		}
		parent := c._compileEnvRelation[currentEnvId].parent
		atomic.StoreUint32(&c._compileEnvLock, 0)
		currentEnvId = parent
	}
}

func (c *CompilerEnvironment) AddEnvironmentToEnvironment(parentEnv uint64, env uint64) {
	for atomic.CompareAndSwapUint32(&c._compileEnvLock, 0, 1) == false {
	}
	if _, ok := c._compileEnvRelation[env]; !ok {
		c._compileEnvRelation[env] = &struct {
			hasParent             bool
			parent                uint64
			table                 map[uint64]uint64
			tableSymbolToSymbolId map[uint64]uint64
			children              []uint64
			tableSymbolCount      uint64
		}{parent: parentEnv, table: map[uint64]uint64{}, hasParent: true, tableSymbolCount: 0, tableSymbolToSymbolId: map[uint64]uint64{}, children: []uint64{}}
		atomic.StoreUint32(&c._compileEnvLock, 0)
		return
	}
	atomic.StoreUint32(&c._compileEnvLock, 0)
	panic("found env")
}

func (c *CompilerEnvironment) AddSymbolToEnvironment(env uint64, symbol uint64) uint64 {
	for atomic.CompareAndSwapUint32(&c._compileEnvLock, 0, 1) == false {
	}
	index := c._compileEnvRelation[env].tableSymbolCount
	c._compileEnvRelation[env].table[index] = symbol
	c._compileEnvRelation[env].tableSymbolToSymbolId[symbol] = index
	c._compileEnvRelation[env].tableSymbolCount = index + 1
	atomic.StoreUint32(&c._compileEnvLock, 0)
	return index
}
