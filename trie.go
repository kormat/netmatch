package netmatch

import (
	"errors"
)

type Data interface{}

var singleBitMask [8]byte

func init() {
	for i := 0; i < 8; i++ {
		singleBitMask[i] = 0x1 << uint(7-i)
	}
}

// ErrInvalidLength is returned for invalid lengths.
var ErrInvalidLength = errors.New("invalid length")

// ErrNotContained is returned for deletes if the given prefix is not contained.
var ErrNotContained = errors.New("not contained")

// ErrAlreadyPresent is returned when adding duplicate prefix.
var ErrAlreadyPresent = errors.New("already present")

type node struct {
	children [2]*node
	data     Data
}

// Trie implements a prefix tree to match CIDR-like IP subnets.
type Trie struct {
	root *node
}

// New returns a new Trie.
func New() *Trie {
	return &Trie{
		root: &node{},
	}
}

// Add adds the given network to the Trie.
// Note that a valid IPv6 prefix and appropriate length are expected.
func (t *Trie) Add(prefix [16]byte, length int, data Data) error {
	if length >= 127 {
		return ErrInvalidLength
	}
	current := t.root
	next := t.root
	for i := 0; i < length; i++ {
		maskPosition := i % 8
		currentByte := i / 8

		left := (prefix[currentByte] & singleBitMask[maskPosition]) == 0
		child := 0
		if !left {
			child = 1
		}

		next = current.children[child]
		if next == nil {
			next = &node{}
			current.children[child] = next
		}

		current = next
	}

	if next.data != nil {
		return ErrNotContained
	}
	next.data = data
	return nil
}

// Match matches the given IP address (in IPv6 format) against the Trie.
func (t *Trie) Match(addr [16]byte) (Data, error) {
	current := t.root
	next := t.root
	for i := 0; i < 127; i++ {
		maskPosition := i % 8
		currentByte := i / 8

		left := (addr[currentByte] & singleBitMask[maskPosition]) == 0

		if left {
			next = current.children[0]
		} else {
			next = current.children[1]
		}

		if next == nil {
			return current.data, nil
		}

		current = next
	}

	return nil, nil
}

// Remove removes the given network from the Trie.
// It expects the same parameters used to Add the prefix earlier.
func (t *Trie) Remove(prefix [16]byte, length int) error {
	if length >= 127 {
		return ErrInvalidLength
	}

	return t.delRecur(prefix, length, 0, t.root)
}

func (t *Trie) delRecur(prefix [16]byte, length, pos int, current *node) error {
	var next *node

	maskPosition := pos % 8
	currentByte := pos / 8

	left := (prefix[currentByte] & singleBitMask[maskPosition]) == 0
	child := 0
	if !left {
		child = 1
	}
	next = current.children[child]

	if next == nil {
		return ErrNotContained
	}

	if pos == length-1 {
		//break
		if next.data == nil {
			return ErrNotContained
		}
		next.data = nil
		if next.children[0] == nil && next.children[1] == nil {
			current.children[child] = nil
		}
	} else {
		//go deeper
		err := t.delRecur(prefix, length, pos+1, next)
		if err != nil {
			return err
		}
		if next.children[0] == nil && next.children[1] == nil {
			current.children[child] = nil
		}
	}
	return nil
}
