package compile

import (
	"github.com/google/uuid"
	"sync/atomic"
)

type CompilerEnvironment struct {
	SharedEnvId     string
	Instr           []Instr
	CompileEnvIndex uint64
	CompileEnvLock  uint32
	GlobalEnv       []RuntimeEnv
}

type RuntimeEnv struct {
	SelfIndex uint64
	Frame     map[uint64]SExpression
	Parent    uint64
	HasParent bool
}

func (e RuntimeEnv) TypeId() string {
	return "environment"
}

func (e RuntimeEnv) SExpressionTypeId() SExpressionType {
	return SExpressionTypeEnvironment
}

func (e RuntimeEnv) String(env *CompilerEnvironment) string {
	return "environment"
}

func (e RuntimeEnv) IsList() bool {
	return false
}

func (e RuntimeEnv) Equals(sexp SExpression) bool {
	panic("implement me")
}

type SymbolTable struct {
	symbolCount      uint64
	symbolMap        map[string]uint64
	reverseSymbolMap map[uint64]string
}

var symbolWriteMutex = uint32(0)
var symbolReadMutex = int64(0)
var symbolTable = &SymbolTable{
	symbolCount:      0,
	symbolMap:        map[string]uint64{},
	reverseSymbolMap: map[uint64]string{},
}

func (s *SymbolTable) GetSymbolCount() uint64 {
	atomic.AddInt64(&symbolReadMutex, 1)
	for atomic.CompareAndSwapUint32(&symbolWriteMutex, 0, 0) == false {
	}
	defer atomic.AddInt64(&symbolReadMutex, -1)
	return s.symbolCount
}

func (s *SymbolTable) GetSymbol(symbol string) uint64 {
	atomic.AddInt64(&symbolReadMutex, 1)
	for atomic.CompareAndSwapUint32(&symbolWriteMutex, 0, 0) == false {
	}
	if symbolId, ok := s.symbolMap[symbol]; ok {
		return symbolId
	}
	incrementCount := atomic.AddUint64(&s.symbolCount, 1)
	atomic.AddInt64(&symbolReadMutex, -1)
	for atomic.LoadInt64(&symbolReadMutex) == 0 && atomic.CompareAndSwapUint32(&symbolWriteMutex, 0, 1) == false {
	}
	s.symbolMap[symbol] = incrementCount
	s.reverseSymbolMap[incrementCount] = symbol
	atomic.StoreUint32(&symbolWriteMutex, 0)
	return incrementCount
}

func (s *SymbolTable) GetSymbolById(symbolId uint64) string {
	atomic.AddInt64(&symbolReadMutex, 1)
	for atomic.CompareAndSwapUint32(&symbolWriteMutex, 0, 0) == false {
	}
	if val, ok := s.reverseSymbolMap[symbolId]; ok {
		return val
	}
	atomic.AddInt64(&symbolReadMutex, -1)
	return ""
}

func NewCompileEnvironment() *CompilerEnvironment {
	env := &CompilerEnvironment{
		SharedEnvId:     uuid.NewString(),
		Instr:           []Instr{},
		CompileEnvIndex: 0,
		CompileEnvLock:  0,
		GlobalEnv: []RuntimeEnv{
			{
				SelfIndex: 0,
				Frame:     map[uint64]SExpression{},
			},
		},
	}
	return env
}

func NewCompileEnvironmentBySharedEnvId(sharedEnvId string) *CompilerEnvironment {
	env := &CompilerEnvironment{
		Instr:           []Instr{},
		CompileEnvIndex: 0,
		CompileEnvLock:  0,
		GlobalEnv: []RuntimeEnv{
			{
				SelfIndex: 0,
				Frame:     map[uint64]SExpression{},
			},
		},
		SharedEnvId: sharedEnvId,
	}
	return env
}

func (c *CompilerEnvironment) SetInstr(instr []Instr) {
	c.Instr = instr
}

func (c *CompilerEnvironment) Compile(sexp SExpression) error {
	stack, _, err := GenerateOpCode(c, sexp, 0)
	if err != nil {
		return err
	}
	c.Instr = stack
	return nil
}

func (c *CompilerEnvironment) GetInstr() []Instr {
	return c.Instr
}

func (c *CompilerEnvironment) GetCompilerSymbol(symbol string) uint64 {
	return symbolTable.GetSymbol(symbol)
}

func (c *CompilerEnvironment) GetCompilerSymbolString(symbol uint64) string {
	return symbolTable.GetSymbolById(symbol)
}

func (c *CompilerEnvironment) GetNewEnvironmentIndex() uint64 {
	return atomic.AddUint64(&c.CompileEnvIndex, 1)
}

//func (c *CompilerEnvironment) FindSymbolIndexInEnvironment(env uint64, symbol uint64) (uint64, uint64, error) {
//
//	currentEnvId := env
//
//	for !atomic.CompareAndSwapUint32(&c.CompileEnvLock, 0, 1) {
//	}
//	for {
//		if _, ok := c.CompileEnvRelation[currentEnvId]; !ok {
//			atomic.StoreUint32(&c.CompileEnvLock, 0)
//			return 0, 0, errors.New("env not found")
//		}
//		reR := c.CompileEnvRelation[currentEnvId]
//		if index, ok := reR.tableSymbolToSymbolId[symbol]; ok {
//			atomic.StoreUint32(&c.CompileEnvLock, 0)
//			return currentEnvId, index, nil
//		}
//		if reR.hasParent == false {
//			atomic.StoreUint32(&c.CompileEnvLock, 0)
//			return 0, 0, errors.New("symbol not found")
//		}
//		if currentEnvId == 0 {
//			atomic.StoreUint32(&c.CompileEnvLock, 0)
//			return 0, 0, errors.New("symbol not found")
//		}
//		parent := c.CompileEnvRelation[currentEnvId].parent
//		currentEnvId = parent
//	}
//}
//
//func (c *CompilerEnvironment) FindSymbolInEnvironment(env uint64, symbolIndex uint64) (uint64, error) {
//
//	currentEnvId := env
//
//	for {
//		for atomic.CompareAndSwapUint32(&c.CompileEnvLock, 0, 1) == false {
//		}
//		if _, ok := c.CompileEnvRelation[currentEnvId]; !ok {
//			atomic.StoreUint32(&c.CompileEnvLock, 0)
//			return 0, errors.New("env not found")
//		}
//		reR := c.CompileEnvRelation[currentEnvId]
//		if symId, ok := reR.table[symbolIndex]; ok {
//			atomic.StoreUint32(&c.CompileEnvLock, 0)
//			return symId, nil
//		}
//		if reR.hasParent == false {
//			return 0, errors.New("symbol not found")
//		}
//		if currentEnvId == 0 {
//			return 0, errors.New("symbol not found")
//		}
//		parent := c.CompileEnvRelation[currentEnvId].parent
//		atomic.StoreUint32(&c.CompileEnvLock, 0)
//		currentEnvId = parent
//	}
//}
//
//func (c *CompilerEnvironment) AddEnvironmentToEnvironment(parentEnv uint64, env uint64) {
//	for atomic.CompareAndSwapUint32(&c.CompileEnvLock, 0, 1) == false {
//	}
//	if _, ok := c.CompileEnvRelation[env]; !ok {
//		c.CompileEnvRelation[env] = &struct {
//			hasParent             bool
//			parent                uint64
//			table                 map[uint64]uint64
//			tableSymbolToSymbolId map[uint64]uint64
//			children              []uint64
//			tableSymbolCount      uint64
//		}{parent: parentEnv, table: map[uint64]uint64{}, hasParent: true, tableSymbolCount: 0, tableSymbolToSymbolId: map[uint64]uint64{}, children: []uint64{}}
//		atomic.StoreUint32(&c.CompileEnvLock, 0)
//		return
//	}
//	atomic.StoreUint32(&c.CompileEnvLock, 0)
//	panic("found env")
//}
//
//func (c *CompilerEnvironment) AddSymbolToEnvironment(env uint64, symbol uint64) uint64 {
//	for atomic.CompareAndSwapUint32(&c.CompileEnvLock, 0, 1) == false {
//	}
//	index := c.CompileEnvRelation[env].tableSymbolCount
//	c.CompileEnvRelation[env].table[index] = symbol
//	c.CompileEnvRelation[env].tableSymbolToSymbolId[symbol] = index
//	c.CompileEnvRelation[env].tableSymbolCount = index + 1
//	atomic.StoreUint32(&c.CompileEnvLock, 0)
//	return index
//}
