package instr

import (
	"bytes"
	"testing"
)

func TestSerialize(t *testing.T) {
	instr := []Instr{
		CreatePushNumberInstr(0),
		CreateDefineInstr("a"),
		CreatePopInstr(),
		CreatePushSymbolInstr("a"),
		CreateLoadInstr(),
		CreatePushNumberInstr(10),
		CreateLessThanNumInstr(2),
		CreateJmpElseInstr(18),
		CreatePushSymbolInstr("a"),
		CreateLoadInstr(),
		CreatePushNumberInstr(1),
		CreatePlusNumInstr(2),
		CreateSetInstr("a"),
		CreatePopInstr(),
		CreatePushSymbolInstr("a"),
		CreateLoadInstr(),
		CreatePrintlnInstr(1),
		CreateJmpInstr(3),
		CreateEndCodeInstr(),
	}

	serialized := Serialize(instr)
	deserialized := Deserialize(serialized)

	for i := 0; i < len(instr); i++ {
		d := deserialized[i]
		if d.Type != instr[i].Type {
			t.Fatalf("type not match %d %d", d.Type, instr[i].Type)
		}
		if d.Length != instr[i].Length {
			t.Fatalf("length not match %d %d", d.Length, instr[i].Length)
		}
		if !bytes.Equal(d.Data, instr[i].Data) {
			t.Fatalf("data not match %v %v", d.Data, instr[i].Data)
		}
	}
}
