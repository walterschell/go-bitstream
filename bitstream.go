package smallcurve

import (
	"encoding/binary"
	"fmt"
)

// Represents a stream of bits (not nesseciarily byte size or byte aligned)
// Not in any way shape or form performant
type BitStream struct {
	nextBitIndex int
	backingStore []byte
}

// Returns the number of bits inserted into the stream so far
func (s *BitStream) Size() int {
	return s.nextBitIndex
}



// Returns the minimum number of bytes needed to represent bitcount bits
func bytesNeededForBits(bitcount int) int {
	result := bitcount / 8
	if bitcount % 8 != 0 {
		result += 1
	}
	return result
}

// Returns the most signifigant n bits from b right shifted 
func topNBits(b byte, n int) byte {
	return (b >> (8-n))
} 

// Returns the least signifigant
func bottomNBits(b byte, n int) byte {
	return b & ((1 << n) - 1)
}

// Appends bits to the stream. They must be MSB justified
func (s *BitStream) AppendBits(size int, bits []byte) {
	if len(bits) < bytesNeededForBits(size) {
		panic(fmt.Sprintf("Too few bytes (%d) for %d bits (need %d)", len(bits), size, bytesNeededForBits(size)))
	}
	neededSize := s.nextBitIndex + size

	// Ensure there is enough size
	if s.backingStore == nil || len(s.backingStore) <= neededSize {
		nextSize := 16
		if s.backingStore != nil {
			nextSize = len(s.backingStore) * 2
		}
		nextStore := make([]byte, nextSize)
		copy(nextStore[:], s.backingStore)
		s.backingStore = nextStore
	}
	numFullBytes := size / 8

	offset := s.nextBitIndex % 8
	bitsLeftInByte := 8 - offset

	for _, bt := range(bits[:numFullBytes]) {
		nextByteIndex := (s.nextBitIndex + 1) / 8
		top := topNBits(bt, bitsLeftInByte)
		bottom := bottomNBits(bt, offset)

		s.backingStore[nextByteIndex] |= top
		s.backingStore[nextByteIndex+1] |= bottom << bitsLeftInByte
		s.nextBitIndex += 8		
	}

	// Handle the last byte
	if size % 8 != 0 {
		trailingBits := size % 8
		topCount := min(trailingBits, bitsLeftInByte)
		bottomCount := trailingBits - topCount
		nextByteIndex := (s.nextBitIndex + 1) / 8
		top := topNBits(bits[numFullBytes], topCount)
		shiftAmount := 8 - (s.nextBitIndex % 8) - topCount
		if shiftAmount < 0 {
			panic(fmt.Sprintf("Shift amount is negative: %d", shiftAmount))
		}
		top = top << shiftAmount
		s.backingStore[nextByteIndex] |= top

		bottom := bottomNBits(bits[numFullBytes] << topCount, bottomCount)
		s.backingStore[nextByteIndex] |= top
		if bottomCount != 0 {
			s.backingStore[nextByteIndex+1] |= bottom << (8 - bottomCount)
		}
		s.nextBitIndex += trailingBits
	}
}

// Appends a single bit to the stream
// bit must be 0 or 1
func (s *BitStream) AppendBit(bit byte) {
	if bit != 0 && bit != 1 {
		panic(fmt.Sprintf("Bit must be 0 or 1, got %d", bit))
	}
	switch bit {
	case 0:
		s.AppendBits(1, []byte{0})
	case 1:
		s.AppendBits(1, []byte{128})
	}
}

// Appends a Uint64 to the stream
func (s *BitStream) AppendUint64(value uint64) {
	content := make([]byte, 8)
	binary.BigEndian.PutUint64(content, value)
	s.AppendBits(64, content)
}

// Appends a Uint to the stream
func (s *BitStream) AppendUint(value uint64, size int) {
	if size > 64 {
		panic("Size out of range")
	}
	mask := ((uint64)(1) << size) - 1
	if (value & mask) != value {
		panic(fmt.Sprintf("%d is larger than %d bits", value, size))
	}
	shifted := value << (64 - size)

	content := make([]byte, 8)
	binary.BigEndian.PutUint64(content, shifted)
	s.AppendBits(size, content)
}

// Returns the bit at a given index
func (s *BitStream) BitAt(index int) byte {
	if index >= s.nextBitIndex {
		panic(fmt.Sprintf("Index %d is out of bounds", index))
	}
	byteIndex := index / 8
	bitIndex := index % 8
	return (s.backingStore[byteIndex] >> (7 - bitIndex)) & 1
}

// Returns the MSB justified size bits at given index
func (s *BitStream) BitsAt(index int, size int) []byte {
	if index + size > s.nextBitIndex {
		panic(fmt.Sprintf("Index %d is out of bounds", index))
	}
	resultSize := bytesNeededForBits(size)
	result := make([]byte, resultSize)
	byteIndex := 0
	shift := 7
	for i := 0; i < size; i++ {
		if shift < 0 {
			byteIndex++
			shift = 7
		}
		result[byteIndex] |= s.BitAt(index + i) << shift
		shift--
	}
	return result	
}

// Interprets the (possibly unaligned) bits at a given index as a uint64
func (s *BitStream) Uint64At(index int) uint64 {
	data := s.BitsAt(index, 64)
	return binary.BigEndian.Uint64(data)
}

// Interprets the (possibly unaligned) size bits at a given index as a uint64)
func (s *BitStream) UintAt(index int, size int) uint64 {
	if size > 64 {
		panic("Size is too big")
	}
	uintBuffer := make([]byte, 8)
	data := s.BitsAt(index, size)
	copy(uintBuffer, data)
	shifted := binary.BigEndian.Uint64(uintBuffer) >> (64 - size)
	return shifted
}

// Returns a string representation of the bitstream
func (s *BitStream) String() string {
	result := ""
	for i := 0; i < s.nextBitIndex; i++ {
		result += fmt.Sprintf("%d", s.BitAt(i))
	}
	return result
}

// Appends another bitstream to this one
func (s *BitStream) AppendBitstream(other *BitStream) {
	s.AppendBits(other.Size(), other.backingStore)
}

// Returns a byte array representation of the bitstream.
// The size is the minimum number of bytes needed to represent the bits
func (s *BitStream) ToBytes() []byte {
	size := s.Size()
	result := make([]byte, bytesNeededForBits(size))
	copy(result, s.backingStore)
	return result
}
