// Copyright (c) 2020-2021 Blockwatch Data Inc.
// Author: alex@blockwatch.cc

package tezos

import (
	"bytes"
	"errors"
	"fmt"
	"strings"

	"blockwatch.cc/tzgo/base58"
)

var (
	// ErrChecksumMismatch describes an error where decoding failed due
	// to a bad checksum.
	ErrChecksumMismatch = errors.New("checksum mismatch")

	// ErrUnknownAddressType describes an error where an address can not
	// decoded as a specific address type due to the string encoding
	// begining with an identifier byte unknown to any standard or
	// registered (via Register) network.
	ErrUnknownAddressType = errors.New("unknown address type")

	// InvalidAddress is an empty invalid address
	InvalidAddress = Address{Type: AddressTypeInvalid, Hash: nil}

	// ZeroAddress is a tz1 address with all bytes zero
	ZeroAddress = Address{Type: AddressTypeEd25519, Hash: make([]byte, HashTypePkhEd25519.Len())}
)

type AddressType byte

const (
	AddressTypeInvalid AddressType = iota
	AddressTypeEd25519
	AddressTypeSecp256k1
	AddressTypeP256
	AddressTypeContract
	AddressTypeBlinded
	AddressTypeBaker
)

func ParseAddressType(s string) AddressType {
	switch s {
	case "ed25519", ED25519_PUBLIC_KEY_HASH_PREFIX:
		return AddressTypeEd25519
	case "secp256k1", SECP256K1_PUBLIC_KEY_HASH_PREFIX:
		return AddressTypeSecp256k1
	case "p256", P256_PUBLIC_KEY_HASH_PREFIX:
		return AddressTypeP256
	case "contract", NOCURVE_PUBLIC_KEY_HASH_PREFIX:
		return AddressTypeContract
	case "blinded", BLINDED_PUBLIC_KEY_HASH_PREFIX:
		return AddressTypeBlinded
	case "baker", BAKER_PUBLIC_KEY_HASH_PREFIX:
		return AddressTypeBaker
	default:
		return AddressTypeInvalid
	}
}

func (t AddressType) IsValid() bool {
	return t != AddressTypeInvalid
}

func (t AddressType) String() string {
	switch t {
	case AddressTypeEd25519:
		return "ed25519"
	case AddressTypeSecp256k1:
		return "secp256k1"
	case AddressTypeP256:
		return "p256"
	case AddressTypeContract:
		return "contract"
	case AddressTypeBlinded:
		return "blinded"
	case AddressTypeBaker:
		return "baker"
	default:
		return "invalid"
	}
}

func (t AddressType) Prefix() string {
	switch t {
	case AddressTypeEd25519:
		return ED25519_PUBLIC_KEY_HASH_PREFIX
	case AddressTypeSecp256k1:
		return SECP256K1_PUBLIC_KEY_HASH_PREFIX
	case AddressTypeP256:
		return P256_PUBLIC_KEY_HASH_PREFIX
	case AddressTypeContract:
		return NOCURVE_PUBLIC_KEY_HASH_PREFIX
	case AddressTypeBlinded:
		return BLINDED_PUBLIC_KEY_HASH_PREFIX
	case AddressTypeBaker:
		return BAKER_PUBLIC_KEY_HASH_PREFIX
	default:
		return ""
	}
}

func (t AddressType) Tag() byte {
	switch t {
	case AddressTypeEd25519:
		return 0
	case AddressTypeSecp256k1:
		return 1
	case AddressTypeP256:
		return 2
	case AddressTypeBlinded:
		return 3
	default:
		return 255
	}
}

func ParseAddressTag(b byte) AddressType {
	switch b {
	case 0:
		return AddressTypeEd25519
	case 1:
		return AddressTypeSecp256k1
	case 2:
		return AddressTypeP256
	case 3:
		return AddressTypeBlinded
	default:
		return AddressTypeInvalid
	}
}

func (t *AddressType) UnmarshalText(data []byte) error {
	typ := ParseAddressType(string(data))
	if !typ.IsValid() {
		return ErrUnknownAddressType
	}
	*t = typ
	return nil
}

func (t AddressType) MarshalText() ([]byte, error) {
	return []byte(t.String()), nil
}

func HasAddressPrefix(s string) bool {
	for _, prefix := range []string{
		ED25519_PUBLIC_KEY_HASH_PREFIX,
		SECP256K1_PUBLIC_KEY_HASH_PREFIX,
		P256_PUBLIC_KEY_HASH_PREFIX,
		NOCURVE_PUBLIC_KEY_HASH_PREFIX,
		BLINDED_PUBLIC_KEY_HASH_PREFIX,
		BAKER_PUBLIC_KEY_HASH_PREFIX,
	} {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}

func (t AddressType) HashType() HashType {
	switch t {
	case AddressTypeEd25519:
		return HashTypePkhEd25519
	case AddressTypeSecp256k1:
		return HashTypePkhSecp256k1
	case AddressTypeP256:
		return HashTypePkhP256
	case AddressTypeContract:
		return HashTypePkhNocurve
	case AddressTypeBlinded:
		return HashTypePkhBlinded
	case AddressTypeBaker:
		return HashTypePkhBaker
	default:
		return HashTypeInvalid
	}
}

func (t AddressType) KeyType() KeyType {
	switch t {
	case AddressTypeEd25519:
		return KeyTypeEd25519
	case AddressTypeSecp256k1:
		return KeyTypeSecp256k1
	case AddressTypeP256:
		return KeyTypeP256
	default:
		return KeyTypeInvalid
	}
}

type Address struct {
	Type AddressType
	Hash []byte
}

func NewAddress(typ AddressType, hash []byte) Address {
	a := Address{
		Type: typ,
		Hash: make([]byte, len(hash)),
	}
	copy(a.Hash, hash)
	return a
}

func (a Address) IsValid() bool {
	return a.Type != AddressTypeInvalid && len(a.Hash) == a.Type.HashType().Len()
}

func (a Address) Equal(b Address) bool {
	return a.Type == b.Type && bytes.Compare(a.Hash, b.Hash) == 0
}

func (a Address) Clone() Address {
	x := Address{
		Type: a.Type,
		Hash: make([]byte, len(a.Hash)),
	}
	copy(x.Hash, a.Hash)
	return x
}

// String returns the string encoding of the address.
func (a Address) String() string {
	s, _ := EncodeAddress(a.Type, a.Hash)
	return s
}

func (a Address) Short() string {
	s := a.String()
	if len(s) < 12 {
		return s
	}
	return s[:8] + "..." + s[32:]
}

func (a *Address) UnmarshalText(data []byte) error {
	astr := strings.Split(string(data), "%")[0]
	addr, err := ParseAddress(astr)
	if err != nil {
		return err
	}
	*a = addr
	return nil
}

func (a Address) MarshalText() ([]byte, error) {
	return []byte(a.String()), nil
}

// output the 21 (implicit) or 22 byte (contract) version
func (a Address) Bytes() []byte {
	if !a.Type.IsValid() {
		return nil
	}
	if a.Type == AddressTypeContract {
		buf := append([]byte{01}, a.Hash...)
		buf = append(buf, byte(0)) // padding
		return buf
	}
	return append([]byte{a.Type.Tag()}, a.Hash...)
}

// Tezos compatible binary encoding with padding for contracts and
// leading 0-byte for EOAs
func (a Address) Bytes22() []byte {
	if !a.Type.IsValid() {
		return nil
	}
	if a.Type == AddressTypeContract {
		buf := append([]byte{01}, a.Hash...)
		buf = append(buf, byte(0)) // padding
		return buf
	}
	return append([]byte{00, a.Type.Tag()}, a.Hash...)
}

// output the 22 byte version
func (a Address) MarshalBinary() ([]byte, error) {
	if !a.Type.IsValid() {
		return nil, ErrUnknownAddressType
	}
	if a.Type == AddressTypeContract {
		buf := append([]byte{01}, a.Hash...)
		buf = append(buf, byte(0)) // padding
		return buf, nil
	}
	return append([]byte{00, a.Type.Tag()}, a.Hash...), nil
}

// support both the 21 byte and 22 byte versions
// be resilient to longer byte strings with extra padding
func (a *Address) UnmarshalBinary(b []byte) error {
	switch true {
	case len(b) >= 22 && (b[0] == 0 || b[0] == 1):
		if b[0] == 0 {
			a.Type = ParseAddressTag(b[1])
			b = b[2:22]
		} else if b[0] == 1 {
			a.Type = AddressTypeContract
			b = b[1:21]
		} else {
			return fmt.Errorf("invalid binary address prefix %x", b[0])
		}
	case len(b) >= 21:
		a.Type = ParseAddressTag(b[0])
		b = b[1:21]
	default:
		return fmt.Errorf("invalid binary address length %d", len(b))
	}
	if !a.Type.IsValid() {
		return ErrUnknownAddressType
	}
	if cap(a.Hash) < 20 {
		a.Hash = make([]byte, 20)
	} else {
		a.Hash = a.Hash[:20]
	}
	copy(a.Hash, b)
	return nil
}

func IsAddressBytes(b []byte) bool {
	if len(b) < 21 {
		return false
	}
	switch true {
	case len(b) == 22 && (b[0] == 0 || b[0] == 1):
		return true
	case len(b) == 21:
		return ParseAddressTag(b[0]) != AddressTypeInvalid
	default:
		return false
	}
}

// ContractAddress returns the string encoding of the address when used
// as originated contract.
func (a Address) ContractAddress() string {
	s, _ := EncodeAddress(AddressTypeContract, a.Hash)
	return s
}

// BakerAddress returns the string encoding of the address when used
// as baker contract.
func (a Address) BakerAddress() string {
	s, _ := EncodeAddress(AddressTypeBaker, a.Hash)
	return s
}

func MustParseAddress(addr string) Address {
	a, err := ParseAddress(addr)
	if err != nil {
		panic(err)
	}
	return a
}

func ParseAddress(addr string) (Address, error) {
	if len(addr) == 0 {
		return InvalidAddress, nil
	}
	a := Address{}
	// check for blinded address first
	if strings.HasPrefix(addr, BLINDED_PUBLIC_KEY_HASH_PREFIX) {
		return DecodeBlindedAddress(addr)
	}
	decoded, version, err := base58.CheckDecode(addr, 3, nil)
	if err != nil {
		if err == base58.ErrChecksum {
			return a, ErrChecksumMismatch
		}
		return a, fmt.Errorf("decoded address is of unknown format: %w", err)
	}
	if len(decoded) != 20 {
		return a, errors.New("decoded address hash is of invalid length")
	}
	switch true {
	case bytes.Compare(version, BAKER_PUBLIC_KEY_HASH_ID) == 0:
		return Address{Type: AddressTypeBaker, Hash: decoded}, nil
	case bytes.Compare(version, ED25519_PUBLIC_KEY_HASH_ID) == 0:
		return Address{Type: AddressTypeEd25519, Hash: decoded}, nil
	case bytes.Compare(version, SECP256K1_PUBLIC_KEY_HASH_ID) == 0:
		return Address{Type: AddressTypeSecp256k1, Hash: decoded}, nil
	case bytes.Compare(version, P256_PUBLIC_KEY_HASH_ID) == 0:
		return Address{Type: AddressTypeP256, Hash: decoded}, nil
	case bytes.Compare(version, NOCURVE_PUBLIC_KEY_HASH_ID) == 0:
		return Address{Type: AddressTypeContract, Hash: decoded}, nil
	default:
		return a, fmt.Errorf("decoded address %s is of unknown type %x", addr, version)
	}
}

func EncodeAddress(typ AddressType, addrhash []byte) (string, error) {
	if len(addrhash) != 20 {
		return "", fmt.Errorf("invalid address hash")
	}
	switch typ {
	case AddressTypeBaker:
		return base58.CheckEncode(addrhash, BAKER_PUBLIC_KEY_HASH_ID), nil
	case AddressTypeEd25519:
		return base58.CheckEncode(addrhash, ED25519_PUBLIC_KEY_HASH_ID), nil
	case AddressTypeSecp256k1:
		return base58.CheckEncode(addrhash, SECP256K1_PUBLIC_KEY_HASH_ID), nil
	case AddressTypeP256:
		return base58.CheckEncode(addrhash, P256_PUBLIC_KEY_HASH_ID), nil
	case AddressTypeContract:
		return base58.CheckEncode(addrhash, NOCURVE_PUBLIC_KEY_HASH_ID), nil
	case AddressTypeBlinded:
		return base58.CheckEncode(addrhash, BLINDED_PUBLIC_KEY_HASH_ID), nil
	default:
		return "", fmt.Errorf("unknown address type %s for hash=%x\n", typ, addrhash)
	}
}
