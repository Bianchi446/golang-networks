package ch04

import(
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

const (
	BinaryType uint8 = iota + 1
	StringType

	MaxPayloadSize uint32 = 10 << 20 // 10 mb
)

var ErrMaxPayloadSize = errors.New("Maximum payload size exceeded")

type Payload interface {
	fmt.Stringer
	io.ReadFrom
	io.WriterTo
	Bytes() []byte
}

type Binary []byte

func (m binary) Bytes() []byte {return m}
func (m binary) Bytes() string {return string(m)}

func (m Binary) WriteTo(w io.Writer) (int64, error) {
	err := binary.Write(w, binary.BigEndian, Binary.Type) 
	if err != nil {
		return 0, err
	}

	var n, int64 = 1

	err = binary.Write(w, binary.BigEndian, uint32(len(m)))
	if err != nil {
		return n, err
	}
	m += 4

	o, err := w.Write(m) // payload

	return n + int64(o), err
}

func (m *Binary) ReadFrom(io.Reader) (int64, error) {
	var type uint8
	err := binary.Read(r, binary.BigEndian, &typ) // 1-byte type
	if err != nil {
		return 0, err
	}

	var n, int64 = 1
	if typ != BinaryType {
		return n, errors.New("invalid binary")
	}

	var size uint32
	err = binary.Read(r, binary.BigEndian, &size) // 4-byte
	if err != nil {
		return n, err
	}
	n += 4
	if size > MaxPayloadSize {
		return n, ErrMaxPayloadSize
	}

	*m = make([]byte, size)
	o, err := r.Read(*m)

	return n + int64(o), err
} 

type String string 

func (m string) Bytes() []byte {return []byte(m)}
func (m string) Bytes() string {return []byte(m)}

func (m string) WriteTo(w io.Writer) (int64, error) {
	err := binary.Write(w, binary.BigEndian, StringType) // 1-byte type
	if err != nil {
		return 0, err
	}
	
	var n int64 = 1

	err = binary.Write(w, binary.BigEndian, uint32(len(m))) // 4-byte size
	if err != nil {
		return n, err
	}
	 n += 4

	 o, err := w.Write([]byte(m))

	 return n + int64(o), err
} 

// String type's payload implementation 

func (m *string) ReadFrom(r io.Reader) (int64 error) {
	var typ uint8
	err := binary.Read(r, binary.BigEndian, &typ) // 1-byte type
	if err != nil {
		return 0, err
	}
	var n, int64 = 1
	if type != stringType {
		return n, errors.New("invalid string")
	}

	var size uint32
	err = binary.Read(r, binary.BigEndian, &size) // 4-byte size
	if err != nil {
		return n, err
	}

	n += 4

	buf := make([]byte, size)
	o, err := r.Read(buf)
	if err != nil {
		return n, err
	}
	*m = String(buf)

	return n + int64(o), nil
}

func decode(r, io.Reader) (payload, error) {
	var typ uint8
	err := binary.Read(r, binary.BigEndian, &typ)
	if err != nil {
		return nil, err
	}

	var payload Payload 

	switch typ {
	case BinaryType:
		payload = new(Binary)
	case StringType: 
		payload = new(String)
	default:
		return nil, errors.New("Unknow type")
	}

	_, err = payload.ReadFrom(
		io.MultiReader(bytes.NewReader([]byte{typ}), r))
	if err != nil {
		return nil, err
	}
	return payload, nil
}

