package flightrecorderreceiver

import (
	"unsafe"

	"github.com/zeebo/xxh3"
)

// fn is a helper struct for the functions table.
type fn struct {
	nameIdx       int32
	systemNameIdx int32
	fileNameIdx   int32
	startLine     int64
}

// kvu is a helper struct for the KeyValueAndUnit attributes table.
type kvu struct {
	keyIdx  int32
	unitIdx int32

	// TODO: convert value to something like AnyValue
	// https://github.com/open-telemetry/opentelemetry-proto/blob/d6dc40fe54b8d441fd7a920af29d96c3ba6ed36a/opentelemetry/proto/profiles/v1development/profiles.proto#L495
	value string
}

// stackinfo is a helper struct for the stacks table.
type stackinfo struct {
	locationIndices []int32
	stackTableIdx   int32
}

type lookupTable struct {
	// mappings - currently not used
	// location - TODO
	functions map[fn]int32
	// links - currently not used
	strings    map[string]int32
	attributes map[kvu]int32
	stacks     map[uint64]stackinfo
}

func createLookupTable() lookupTable {
	// By convention, all tables in the dictionary hold
	// an empty element as their first element.
	functions := make(map[fn]int32)
	functions[fn{}] = 0

	strings := make(map[string]int32)
	strings[""] = 0

	attributes := make(map[kvu]int32)
	attributes[kvu{}] = 0

	stacks := make(map[uint64]stackinfo)
	stacks[0] = stackinfo{stackTableIdx: 0}

	return lookupTable{
		functions:  functions,
		strings:    strings,
		attributes: attributes,
		stacks:     stacks,
	}
}

// AddString returns an index to the given string in the lookup table.
func (lt *lookupTable) AddString(s string) int32 {
	if idx, exists := lt.strings[s]; exists {
		return idx
	}
	idx := int32(len(lt.strings))
	lt.strings[s] = idx
	return idx
}

// AddKeyValueUnit returns an index to the given attribute in the lookup table.
func (lt *lookupTable) AddKeyValueUnit(k, v, u string) int32 {
	attr := kvu{
		keyIdx:  lt.AddString(k),
		value:   v,
		unitIdx: lt.AddString(u),
	}
	if idx, exists := lt.attributes[attr]; exists {
		return idx
	}

	idx := int32(len(lt.attributes))
	lt.attributes[attr] = idx
	return idx
}

// AddFunction returns an index to the given function in the lookup table.
func (lt *lookupTable) AddFunction(name, systemName, fileName string, startLine int64) int32 {
	attr := fn{
		nameIdx:       lt.AddString(name),
		systemNameIdx: lt.AddString(systemName),
		fileNameIdx:   lt.AddString(fileName),
		startLine:     startLine,
	}
	if idx, exists := lt.functions[attr]; exists {
		return idx
	}

	idx := int32(len(lt.functions))
	lt.functions[attr] = idx
	return idx
}

func hashLocations(locs []int32) uint64 {
	// Reinterpret the []int32 as a []byte slice.
	// Since an int32 is 4 bytes, the new length is len(locs) * 4.
	// This does NOT allocate; it only creates a new slice header on the stack.
	b := unsafe.Slice((*byte)(unsafe.Pointer(&locs[0])), len(locs)*4)

	return xxh3.Hash(b)
}

func (lt *lookupTable) AddStack(locs []int32) int32 {
	if len(locs) == 0 {
		return 0
	}

	h := hashLocations(locs)
	if s, exists := lt.stacks[h]; exists {
		return s.stackTableIdx
	}
	idx := int32(len(lt.stacks))
	lt.stacks[h] = stackinfo{
		locationIndices: locs,
		stackTableIdx:   idx,
	}
	return idx
}
