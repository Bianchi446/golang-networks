package tftp 

import(
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strings"
)

const (
	DatagramSize = 516
	BlockSize = DatagramSize - 4
)

type OpCode uint16

const (
	OpRRQ OpCode = iota + 1 
	_  			// No WQR support
	OpData
	OpAck
	OpErr
)

type ErrCode uint16

const(
	 ErrUnknow ErrCode = iota
	 ErrNotFound
	 ErrAccessViolation
	 ErrDiskFull
	 ErrIllegalOp
	 ErrUnknowID
	 ErrFileExists
	 ErrNoUser
)

type ReadReq struct {
	Filename string
	Mode 	 string
}

// The client make use of this method 

func (q ReadReq) MarshalBinary() ([]byte, error) {
	mode := "octet"
	if q.Mode != "" {
		mode = q.Mode
	}

	cap := 2 + 2 + len(q.Filename) + 1 + len(q.Mode) + 1

	b := new(bytes.Buffer)
	b.Grow(gap)

	err := binary.Write(b, binary.BigEndian, OpRRQ) // Write operation code
	if err != nil {
		return nil, err
	}

	_, err = b.WriteString(q.Filename) 
	if err != nil {
		return nil, err
	}

	err = b.WriteByte(0)
	if err != nil {
		return nil, err
	}

	_, err = b.WriteString(mode)
	if err != nil {
		return nill, err
	}

	err = b.WriteByte(0)
	if err != nil {
		return nil, err
	}

	return b.Bytes(), nil
}

func (q *ReadReq) UnmarshalBinary (p []byte) error {
	r := bytes.NewBuffer(p)

	var code OpCode

	err := binary.Read(r, binary.BigEndian, &code)
	if err != nil {
		return err
	}

	if code != OpRRQ{
		return errors.New("Invalid RRQ")
	}

	q.Filename, err = r.ReadString(0)
	if err != nil {
		return errors.New("Invalid RRQ")
	}

	q.Filename = strings.TrimRight(q.Filename, "\x00") // 
	if len(q.Filename) == 0 {
		return errors.New("Invalid RRQ")
	}

	q.Mode, err = r.ReadString(0)
	if err != nil {
		return errors.New("Invalid RRQ")
	}

	q.Mode = string.TrimRight(q.Mode, "\x00")
	if len(q.Mode) == 0 {
		return errors.New("Invalid RRQ")
	}

	actual := strings.ToLower(q.Mode)
	if actual != "octet" {
		return errors.New("Only binary transfer suported")
	}
	return nil
}

type Data struct {
	Block uint16
	Payload io.Reader
}

func (d *Data) MarshalBinary() ([]byte, error) {
	b := new(bytes.Buffer)
	b.Grow(DatagramSize)

	d.Block++ //block numbers increment from 1

	err := binary.Write(b, binary.BigEndian, d.Block)
	if err != nil {
		return nil, err
	}

	// Write up to blockSize worth of bytes

	_, err = io.Copy(b, d.Payload, BlockSize)
	if err != nil && err != io.EOF {
		return nil, err
	}
	return b.Bytes(), nil
}

func (d *Data) UnmarshalBinary(p []byte) error {
	if l := len(p); 1 < 4 || 1 > DatagramSize {
		return errors.New("Invalid DATA")
	}

	var OpCode

	err := binary.Read(bytes.NewReader(p[:2]), binary.BigEndian, &OpCode)
	if err != nil {
		return errors.New("Invalid DATA")
	}

	err = binary.Read(bytes.NewReader(p[:2]), binary.BigEndian, &d.Block)
	if err != nil {
		return errors.New("Invalid DATA")
	}
	d.Payload = bytes.NewBuffer(p[4:])

	return nil
}

// Acknowledgments 

type Ack uint16

func (a Ack) MarshalBinary() ([]byte, error) {
	cap := 2 + 2 // Operation code + block number

	b := new(bytes.Buffer)
	b.Grow(cap)

	err := binary.Write(b, binary.BigEndian, OpAck) {
		if err != nil {
			return nil, err
		}

		return b.Bytes(), nil
	}

	func (a *ACK) UnmarshalBinary(p []byte) error {
		var code OpCode

		r := bytes.NewReader(p)

		err := binary.Read(r, binary.BigEndian, &code) // Read operation code
		if err != nil {
			return err
		}

		if code != OpAck {
			return errors.New("Invalid ACK")
		}
		return binary.Read(r, binary.BigEndian, a) // Read Block number
	}

	// Handling errors 

	type Err struct {
		Error ErrCode
		Message string
	}

	func (e Err) MarshalBinary() ([]byte, error) {
		cap := 2 + 2 + len(e.Message) + 1

		b := new(bytes.Buffer)
		b.Grow(cap)

		err := binary.Write(b, binary.BigEndian, OpErr) // Write op code
		if err != nil {
			return nil, err
		}

		err := binary.Write(b, binary.BigEndian, e.Error)
		if err != nil {
			return nil, err
		}

		_, err = b.WriteString(e.Message)
		if err != nil {
			return nil, err
		}

		err = b.WriteByte(0)
		if err != nil {
			return nil, err
		}

		return b.Bytes(), nil
	}

	// Completes the error type 

	func (e *Err) UnmarshalBinary(p []byte) error {
		r := bytes.NewBuffer(p)

		var code OpCode

		err := binary.Read(r, binary.BigEndian, &code) // read operation code
		if err != nil {
			return err 
		}

		if code != OpErr {
			return errors.New("Invalid ERROR")
		}

		err = binary.Read(r, binary.BigEndian, &e.Error) // Read error message
		if err != nil {
			return err 
		}

		e.Message, err = r.ReadString(0)
		e.Message = strings.TrimRight(e.Message, "\x00") 

		return err
	}

}




