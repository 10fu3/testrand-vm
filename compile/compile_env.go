package compile

import (
	"sync"
	"sync/atomic"
	"testrand-vm/infra"
)

type CompilerEnvironment struct {
	SharedEnvId         string
	Instr               []Instr
	RemoteJointVariable *infra.RemoteJointVariable
}

type RuntimeEnv struct {
	Frame     *sync.Map
	Parent    *RuntimeEnv
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
	symbolCount      atomic.Uint64
	symbolMap        *sync.Map
	reverseSymbolMap *sync.Map
}

var symbolTable = &SymbolTable{
	symbolCount:      atomic.Uint64{},
	symbolMap:        &sync.Map{},
	reverseSymbolMap: &sync.Map{},
}

func (s *SymbolTable) GetSymbolCount() uint64 {
	return s.symbolCount.Add(1)
}

func (s *SymbolTable) GetSymbol(symbol string) uint64 {
	if symbolId, ok := s.symbolMap.Load(symbol); ok {
		return symbolId.(uint64)
	}
	incrementCount := s.symbolCount.Add(1)

	s.symbolMap.Store(symbol, incrementCount)
	s.reverseSymbolMap.Store(incrementCount, symbol)

	return incrementCount
}

func (s *SymbolTable) GetSymbolById(symbolId uint64) string {
	if symbol, ok := s.reverseSymbolMap.Load(symbolId); ok {
		return symbol.(string)
	}
	return ""
}

func NewCompileEnvironment(sharedEndId string, remoteJointVariable *infra.RemoteJointVariable) *CompilerEnvironment {
	env := &CompilerEnvironment{
		SharedEnvId:         sharedEndId,
		Instr:               []Instr{},
		RemoteJointVariable: remoteJointVariable,
	}
	return env
}

func NewCompileEnvironmentBySharedEnvId(sharedEnvId string, remoteJointVariable *infra.RemoteJointVariable) *CompilerEnvironment {
	env := &CompilerEnvironment{
		Instr:               []Instr{},
		SharedEnvId:         sharedEnvId,
		RemoteJointVariable: remoteJointVariable,
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
