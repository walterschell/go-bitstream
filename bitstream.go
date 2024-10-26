/*
Package bitstream provides a toy library for bit-level marshaling and
unmarshaling.

This package is designed to handle streams of bits, which are not necessarily
byte-sized or byte-aligned. It includes various utilities for appending and
extracting bits, integers, and big integers to and from a bitstream. The package
also supports base32 encoding and decoding.

The BitStream type represents a stream of bits with methods to manipulate and
query the bitstream.

Key Features:
- Append and extract individual bits, bytes, and integers.
- Support for appending and extracting big integers.
- Base32 encoding and decoding.
- Utility functions for bit manipulation.

Note: This package is not optimized for performance and is intended for
educational or experimental purposes.

Example usage:

	package main

	import (
	    "fmt"
	    "github.com/walterschell/go-bitstream"
	)

	func main() {
	    stream := bitstream.BitStream{}
	    stream.AppendUint(255, 8)
	    fmt.Println(stream.String()) // Output: 11111111
	}
*/
package bitstream

import (
	"encoding/binary"
	"fmt"
	"math/big"
)

// Represents a stream of bits (not nesseciarily byte size or byte aligned)
// Not in any way shape or form performant
type BitStream struct {
	nextBitIndex uint
	backingStore []byte
}

// Returns the number of bits inserted into the stream so far
func (s *BitStream) Size() uint {
	return s.nextBitIndex
}

// Returns the minimum number of bytes needed to represent bitcount bits
func bytesNeededForBits(bitcount uint) uint {
	result := bitcount / 8
	if bitcount%8 != 0 {
		result += 1
	}
	return result
}

// Returns the most signifigant n bits from b right shifted
func topNBits(b byte, n uint) byte {
	return (b >> (8 - n))
}

// Returns the least signifigant
func bottomNBits(b byte, n uint) byte {
	return b & ((1 << n) - 1)
}

// Appends bits to the stream. They must be MSB justified
func (s *BitStream) AppendBits(size uint, bits []byte) {
	if (uint)(len(bits)) < bytesNeededForBits(size) {
		panic(fmt.Sprintf("Too few bytes (%d) for %d bits (need %d)", len(bits), size, bytesNeededForBits(size)))
	}
	neededSize := bytesNeededForBits(s.nextBitIndex + size)

	// Ensure there is enough size
	if s.backingStore == nil || (uint)(len(s.backingStore)) <= neededSize {
		nextSize := uint(16)
		if s.backingStore != nil {
			nextSize = uint(len(s.backingStore))
		}
		for nextSize <= neededSize {
			nextSize *= 2
		}
		nextStore := make([]byte, nextSize)
		copy(nextStore[:], s.backingStore)
		s.backingStore = nextStore
	}
	numFullBytes := size / 8

	offset := s.nextBitIndex % 8 // Offset in the current byte
	bitsLeftInByte := 8 - offset // Number of bits left in the current byte

	for _, bt := range bits[:numFullBytes] {
		nextByteIndex := s.nextBitIndex / 8
		top := topNBits(bt, bitsLeftInByte)
		bottom := bottomNBits(bt, offset)

		s.backingStore[nextByteIndex] |= top
		s.backingStore[nextByteIndex+1] |= bottom << bitsLeftInByte
		s.nextBitIndex += 8
	}

	// Handle the last byte
	if size%8 != 0 {
		trailingBits := size % 8
		topCount := min(trailingBits, bitsLeftInByte)
		bottomCount := trailingBits - topCount
		nextByteIndex := (s.nextBitIndex) / 8
		top := topNBits(bits[numFullBytes], topCount)
		shiftAmount := 8 - (s.nextBitIndex % 8) - topCount
		shiftedTop := top << shiftAmount
		s.backingStore[nextByteIndex] |= shiftedTop
		s.nextBitIndex += topCount
		nextByteIndex = (s.nextBitIndex) / 8

		if 0 < bottomCount {
			bottom := bottomNBits(bits[numFullBytes]>>(8-byte(topCount)-byte(bottomCount)), bottomCount)
			shiftedBottom := bottom << (8 - bottomCount)
			s.backingStore[nextByteIndex] |= shiftedBottom
			s.nextBitIndex += bottomCount
		}
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
func (s *BitStream) AppendUint(value uint64, size uint) {
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
func (s *BitStream) BitAt(index uint) byte {
	if index >= s.nextBitIndex {
		panic(fmt.Sprintf("Index %d is out of bounds", index))
	}
	byteIndex := index / 8
	bitIndex := index % 8
	return (s.backingStore[byteIndex] >> (7 - bitIndex)) & 1
}

// Returns the MSB justified size bits at given index
func (s *BitStream) BitsAt(index uint, size uint) []byte {
	if index+size > s.nextBitIndex {
		panic(fmt.Sprintf("Index %d is out of bounds", index))
	}
	resultSize := bytesNeededForBits(size)
	result := make([]byte, resultSize)
	byteIndex := 0
	shift := 7
	for i := uint(0); i < size; i++ {
		if shift < 0 {
			byteIndex++
			shift = 7
		}
		result[byteIndex] |= s.BitAt(index+i) << shift
		shift--
	}
	return result
}

// Interprets the (possibly unaligned) bits at a given index as a uint64
func (s *BitStream) Uint64At(index uint) uint64 {
	data := s.BitsAt(index, 64)
	return binary.BigEndian.Uint64(data)
}

// Interprets the (possibly unaligned) size bits at a given index as a uint64)
func (s *BitStream) UintAt(index uint, size uint) uint64 {
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
	for i := uint(0); i < s.nextBitIndex; i++ {
		result += fmt.Sprintf("%d", s.BitAt(i))
	}
	return result
}

func (s *BitStream) debugDump() string {
	result := ""
	for i := uint(0); i < bytesNeededForBits(s.nextBitIndex); i++ { // Loop through bytes
		result += fmt.Sprintf("%08b ", s.backingStore[i])
	}

	return result
}

// Appends another bitstream to this one
func (s *BitStream) AppendBitstream(other *BitStream) {
	s.AppendBits(other.Size(), other.backingStore)
}

// Returns a new BitStream that is a copy of s with other appended to it
func (s *BitStream) Concat(other *BitStream) *BitStream {
	result := &BitStream{}
	result.backingStore = make([]byte, bytesNeededForBits(s.Size()+other.Size()))
	copy(result.backingStore, s.backingStore)
	result.nextBitIndex = s.nextBitIndex
	result.AppendBitstream(other)
	return result
}

// Returns a byte array representation of the bitstream.
// The size is the minimum number of bytes needed to represent the bits
func (s *BitStream) ToBytes() []byte {
	size := s.Size()
	result := make([]byte, bytesNeededForBits(size))
	copy(result, s.backingStore)
	return result
}

const base32Alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"

var base32Reversed map[byte]byte

func ensureBase32Reversed() {
	if base32Reversed == nil {
		base32Reversed = make(map[byte]byte)
		for i := 0; i < len(base32Alphabet); i++ {
			base32Reversed[base32Alphabet[i]] = byte(i)
		}
	}
}

// Marshals the bitstream as a base32 string. The bitstream length must be a multiple of 5
func (s *BitStream) MarshalBase32() string {
	if s.Size()%5 != 0 {
		panic(fmt.Sprintf("Bitstream size must be a multiple of 5 for base32 encoding, got %d", s.Size()))
	}
	result := ""
	for i := uint(0); i < s.nextBitIndex; i += 5 {
		value := s.UintAt(i, 5)
		result += string(base32Alphabet[uint(value)])
	}
	return result
}

// Unmarshals a base32 string.
func (s *BitStream) UnmarshalBase32(encoded string) error {
	s.backingStore = nil
	s.nextBitIndex = 0
	ensureBase32Reversed() // Ensure the base32 reverse map is initialized
	for i := 0; i < len(encoded); i++ {
		value, ok := base32Reversed[encoded[i]]
		if !ok {
			return fmt.Errorf("invalid base32 character: %c", encoded[i])
		}
		s.AppendUint(uint64(value), 5)
	}
	return nil
}

// Returns a new BitStream from the given slice of s
func (s *BitStream) BitstreamAt(index uint, size uint) *BitStream {
	result := &BitStream{}
	result.AppendBits(size, s.BitsAt(index, size))
	return result
}

// Appends a BigInt into the bitstream. nbits must set explicitly to allow unambigous unmarshalling
func (s *BitStream) AppendBigInt(value *big.Int, nbits uint) {
	if value.BitLen() > int(nbits) {
		panic(fmt.Sprintf("Value is too big for %d bits", nbits))
	}
	bytes := make([]byte, (nbits+7)/8)
	value.FillBytes(bytes)

	offset := nbits % 8

	if offset == 0 {
		s.AppendBits(nbits, bytes)
		return
	}
	firstByte := uint64(bytes[0])
	s.AppendUint(firstByte, offset)

	remainingSize := nbits - offset
	remainingBytes := bytes[1:]
	s.AppendBits(remainingSize, remainingBytes)
}

// Extracts a BigInt from a BitStream
func (s *BitStream) BigIntAt(index uint, nbits uint) *big.Int {
	offset := nbits % 8
	var resultBytes []byte
	if offset == 0 {
		resultBytes = s.BitsAt(index, nbits)
	} else {
		firstByte := s.UintAt(index, offset)
		remainingBytes := s.BitsAt(index+offset, nbits-offset)
		resultBytes = append([]byte{byte(firstByte)}, remainingBytes...)
	}

	result := new(big.Int)
	result.SetBytes(resultBytes)
	return result
}
