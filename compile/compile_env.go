package compile

import (
	"encoding/binary"
	"sync/atomic"
)

type CompilerEnvironment struct {
	symbolCount      uint64
	symbolMap        map[string]uint64
	reverseSymbolMap map[uint64]string
	instr            []Instr
}

func NewCompileEnvironment() *CompilerEnvironment {
	env := &CompilerEnvironment{
		symbolCount:      0,
		symbolMap:        map[string]uint64{},
		reverseSymbolMap: map[uint64]string{},
		instr:            []Instr{},
	}
	return env
}

func (c *CompilerEnvironment) SetInstr(instr []Instr) {
	c.instr = instr
}

func (c *CompilerEnvironment) Compile(sexp SExpression) error {
	stack, _, err := GenerateOpCode(c, sexp, 0)
	if err != nil {
		return err
	}
	c.instr = stack
	return nil
}

func (c *CompilerEnvironment) Serialize() []byte {

	body := Serialize(c.instr)

	totalSymbolSize := uint64(8)

	for i := uint64(0); i < c.symbolCount+1; i++ {
		totalSymbolSize += 8 + uint64(len(c.reverseSymbolMap[i]))
	}

	b := make([]byte, totalSymbolSize+uint64(len(body)))
	binary.LittleEndian.PutUint64(b, c.symbolCount)
	byteIndex := uint64(8)

	for i := uint64(0); i < c.symbolCount+1; i++ {
		//symbolBody := make([]byte, 8+uint64(len(c.reverseSymbolMap[i])))
		symbolBody := b[byteIndex:]
		binary.LittleEndian.PutUint64(symbolBody, uint64(len(c.reverseSymbolMap[i])))
		byteIndex += 8
		copy(b[8:], c.reverseSymbolMap[i])
		byteIndex += uint64(len(c.reverseSymbolMap[i]))
	}

	copy(b[byteIndex:], body)

	return b
}

func DeserializeCompileEnvironment(b []byte) *CompilerEnvironment {
	c := &CompilerEnvironment{}
	c.symbolCount = binary.LittleEndian.Uint64(b)
	byteIndex := uint64(8)

	for i := uint64(0); i < c.symbolCount+1; i++ {
		symbolLen := binary.LittleEndian.Uint64(b[byteIndex:])
		byteIndex += 8
		c.reverseSymbolMap[i] = string(b[byteIndex : byteIndex+symbolLen])
		byteIndex += symbolLen
	}

	c.instr = DeserializeInstructions(b[byteIndex:])
	return c
}

func (c *CompilerEnvironment) GetInstr() []Instr {
	return c.instr
}

func (c *CompilerEnvironment) GetCompilerSymbol(symbol string) uint64 {
	val, ok := c.symbolMap[symbol]
	if !ok {
		v := atomic.AddUint64(&c.symbolCount, 1)
		c.reverseSymbolMap[v] = symbol
		c.symbolMap[symbol] = v
		return v
	}
	return val
}

func (c *CompilerEnvironment) GetCompilerSymbolString(symbol uint64) string {
	return c.reverseSymbolMap[symbol]
}
