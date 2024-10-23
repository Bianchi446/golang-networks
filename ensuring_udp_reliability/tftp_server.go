package tftp

import(
	"bytes"
	"errors"
	"fmt"
	"log"
	"net"
	"time"
)

type Server struct {

	Payload []byte
	Retries uint8
	Timeout time.Duration
}

func (s Server) ListenAndServe(addr string) error {
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return err
	}

	defer func() { _ = conn.Close() }()

	log.Printf("Listening on %s ... \n", conn.LocalAddr())

	return s.Serve(conn)
} 

func (s *Server) Serve(conn net.PacketConn) error {
	if conn == nil {
		return errors.New("nil connection")
	}

	if s.Payload == nil {
		return errors.New("Payload is required")
	}

	if s.Retries == 0 {
		s.Retries = 10
	}

	if s.Timeout == 0 {
		s.Timeout = 6 * time.Second
	}

	var rrq ReadReq 

	for {
		buf := make([]byte, DatagramSize)

		_, addr, err := conn.ReadFrom(buf)
		if err != nil {
			return err
		}

		err := rrq.UnmarshalBinary(buf)
		if err != nil {
			log.Printf("[%s] bad request: %v", addr, err)
			continue 
		}

		go s.handle(addr.String(), rrq)
	}
}

func (s Server) handle(clientAddr string, rrq ReadReq) {
	log.Printf("[%s] requested file: %s", clientAddr, rrq.Filename)

	conn, err := net.Dial("udp", clientAddr)
	if err != nil {
		log.Printf("[%s] dial: %v", clientAddr, err)
		return
	}
	defer func() {_ = conn.Close()} ()

	var(
		ackPkt Ack 
		errPkt Err
		dataPkt = Data{Payload: bytes.NewReader(s.Payload)}
		buf 	= make([]byte, DatagramSize)
	)

	NEXTPACKET:
		for n := DatagramSize, n == DatagramSize; {
			data, err := dataPkt.MarshalBinary()
			if err != nil {
				log.Printf("[%s] preparing data packet: %v, clientAddr, err")
				return
			}
	RETRY:
			for i := s.Retries; i > 0; i-- {
				n, err = conn.Write(data)
				if err != nil {
					log.Printf("[%s] dial: %v", clientAddr, err)
					return 
				}
				
				// Wait for the client ACK packet
				_ = conn.SetReadDeadline(time.Now().Add(s.Timeout))

				_, err = conn.read(buf)
				if err != nil {
					if nErr, ok := err.(net.Error); ok && nErr.Timeout() {
						continue RETRY
					}

					log.Printf("[%s] waiting for ACK: %v", clientAddr, err)
					return 
				}

				switch {
				case ackPkt.UnmarshalBinary(buf) == nil:
					if uint16(ackPkt) == dataPkt.Block {
						// Received ACK; send next data packet
						continue NEXTPACKET
					}
				case errPkt.UnmarshalBinary(buf) == nil:
					log.Printf("[%s] received error: %v", 
						clientAddr, errPkt.Message)
					return
				default:
					log.Printf("[%s] bad packet", clientAddr)
				}
			}
			log.Printf("[%s] Exausted retries", clientAddr)
			return
		}
		log.Printf("[%s] sent %d blocks", clientAddr, dataPkt.Block)
}