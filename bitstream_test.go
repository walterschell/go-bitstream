package bitstream

import (
	"math/big"
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

func TestAppendByOne(t *testing.T) {
	stream := BitStream{}
	expected := ""
	for i := 0; i < 1000; i++ {
		stream.AppendBit(1)
		expected += "1"
		if stream.String() != expected {
			t.Logf("Backing store: %v\n", stream.debugDump())
			t.Fatalf("Expected %s, got %s", expected, stream.String())
		}
	}
}

func TestAppendByTwo(t *testing.T) {
	stream := BitStream{}
	expected := ""
	for i := 0; i < 1000; i++ {
		stream.AppendUint(3, 2)
		expected += "11"
		if stream.String() != expected {
			t.Fatalf("Expected %s, got %s", expected, stream.String())
		}
	}
}

func TestAppendByThree(t *testing.T) {
	stream := BitStream{}
	expected := ""
	for i := 0; i < 1000; i++ {
		stream.AppendUint(7, 3)
		expected += "111"
		if stream.String() != expected {
			t.Logf("Backing store: %v\n", stream.debugDump())

			t.Fatalf("Expected %s, got %s", expected, stream.String())
		}
	}
}

func TestAppendByFour(t *testing.T) {
	stream := BitStream{}
	expected := ""
	for i := 0; i < 1000; i++ {
		stream.AppendUint(15, 4)
		expected += "1111"
		if stream.String() != expected {
			t.Fatalf("Expected %s, got %s", expected, stream.String())
		}
	}
}

func TestAppendByFive(t *testing.T) {
	stream := BitStream{}
	expected := ""
	for i := 0; i < 1000; i++ {
		stream.AppendUint(31, 5)
		expected += "11111"
		if stream.String() != expected {
			t.Fatalf("Expected %s, got %s", expected, stream.String())
		}
	}
}

func TestAppendBySix(t *testing.T) {
	stream := BitStream{}
	expected := ""
	for i := 0; i < 1000; i++ {
		stream.AppendUint(63, 6)
		expected += "111111"
		if stream.String() != expected {
			t.Fatalf("Expected %s, got %s", expected, stream.String())
		}
	}
}
func TestAppendBySeven(t *testing.T) {
	stream := BitStream{}
	expected := ""
	for i := 0; i < 1000; i++ {
		stream.AppendUint(127, 7)
		expected += "1111111"
		if stream.String() != expected {
			t.Logf("Backing store: %v\n", stream.debugDump())
			t.Fatalf("Expected %s, got %s", expected, stream.String())
		}
	}
}

func toBinaryString(n uint64, size int) string {
	result := make([]byte, (size))
	for i := 0; i < size; i++ {
		if n&(1<<uint(size-i-1)) != 0 {
			result[i] = '1'
		} else {
			result[i] = '0'
		}
	}
	return string(result)

}
func TestAppendB32Patterns(t *testing.T) {
	for n := uint64(1); n < 32; n++ {

		stream := BitStream{}
		expected := ""
		for i := 0; i < 1000; i++ {
			stream.AppendUint(n, 5)
			expected += toBinaryString(n, 5)
			if stream.String() != expected {
				t.Fatalf("Expected %s, got %s", expected, stream.String())
			}
		}
	}
}

func TestAppendBigInt(t *testing.T) {
	value := big.NewInt(7)

	for i := 3; i < 20; i++ {
		stream := BitStream{}
		stream.AppendBigInt(value, uint(i))

		expected := ""
		for j := 3; j < i; j++ {
			expected += "0"
		}
		expected += "111"

		if stream.String() != expected {
			t.Fatalf("(i=%d) Expected %s, got %s", i, expected, stream.String())
		}
		extracted := stream.BigIntAt(0, uint(i))
		if extracted.Cmp(value) != 0 {
			t.Fatalf("(i=%d) Expected 7, got %d", i, extracted)
		}

	}

}

func TestExpandingBackingStore(t *testing.T) {
	message := make([]byte, 16)
	stream := BitStream{}
	expectedSize := uint(0)
	for i := 0; i < 256; i++ {
		stream.AppendBits(128, message)
		expectedSize += 128
		if stream.Size() != expectedSize {
			t.Fatalf("Expected %d, got %d", expectedSize, stream.Size())
		}
	}
}
