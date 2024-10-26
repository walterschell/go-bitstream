package smallcurve
import (
	"testing"
)

func TestBitStream(t *testing.T) {
	stream := BitStream{}
	if stream.Size() != 0 {
		t.Errorf("Expected 0, got %d", stream.Size())
	}
	stream.AppendBits(8, []byte{0xa5})
	if stream.Size() != 8 {
		t.Errorf("Expected 8, got %d", stream.Size())
	}
	if stream.String() != "10100101" {
		t.Errorf("Expected 10100101, got %s", stream.String())
	}
	stream.AppendUint(3, 4)
	if stream.Size() != 12 {
		t.Errorf("Expected 12, got %d", stream.Size())
	}
	if stream.String() != "101001010011" {
		t.Errorf("Expected 101001010011, got %s", stream.String())
	}

}

func TestBitStreamAppendBit(t *testing.T) {
	stream := BitStream{}
	if stream.Size() != 0 {
		t.Errorf("Expected 0, got %d", stream.Size())
	}
	stream.AppendBit(1)
	if stream.Size() != 1 {
		t.Errorf("Expected size 1, got %d", stream.Size())
	}
	if stream.String() != "1" {
		t.Errorf("Expected 1, got %s", stream.String())
	}
	stream.AppendBit(0)
	if stream.Size() != 2 {
		t.Errorf("Expected size 2, got %d", stream.Size())
	}
	if stream.String() != "10" {
		t.Errorf("Expected 10, got %s", stream.String())
	}
	stream2 := BitStream{}
	stream2.AppendUint(255, 8)
	if stream2.Size() != 8 {
		t.Errorf("Expected 8, got %d", stream2.Size())
	}
	if stream2.String() != "11111111" {
		t.Errorf("Expected 11111111, got %s", stream2.String())
	}
	stream.AppendBitstream(&stream2)
	if stream.Size() != 10 {
		t.Errorf("Expected 10, got %d", stream.Size())
	}
	if stream.String() != "1011111111" {
		t.Errorf("Expected 1011111111, got %s", stream.String())
	}
}