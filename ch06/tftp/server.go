package tftp

import (
	"bytes"
	"errors"
	"log"
	"net"
	"time"
)

type Server struct {
	Payload []byte        // 모든 읽기 요청에 반환될 페이로드
	Retries uint8         // 전송 실패 시 재시도 횟수
	Timeout time.Duration // 전송 승인을 기다릴 시간
}

func (s Server) ListenAndServe(addr string) error {
	conn, err := net.ListenPacket("udp", addr)
	if err != nil {
		return err
	}
	defer func() { conn.Close() }()

	log.Printf("Listening on %s...\n", conn.LocalAddr())

	return s.Serve(conn)
}

func (s *Server) Serve(conn net.PacketConn) error {
	// Check fields, and set default
	if conn == nil {
		return errors.New("nil connection")
	}

	if s.Payload == nil {
		return errors.New("payload is required")
	}

	if s.Retries == 0 {
		s.Retries = 10 // set default retries to 10
	}

	if s.Timeout == 0 {
		s.Timeout = time.Second * 6 // set default timeout to 6 seconds
	}

	var rrq ReadReq

	for {
		buf := make([]byte, DatagramSize)

		// Connection으로부터 데이터를 읽음
		_, addr, err := conn.ReadFrom(buf)
		if err != nil {
			return err
		}

		// Read request 객체를 통해 unmarshalling 시도
		// 본 서버는 다운로드만 지원하므로 wrq 체크 x
		err = rrq.UnmarshalBinary(buf)
		if err != nil {
			log.Printf("[%s] bad request: %v", addr, err)
			continue
		}

		go s.handle(addr.String(), rrq)
	}
}

func (s Server) handle(clientAddr string, rrq ReadReq) {
	log.Printf("[%s] requested file: %s", clientAddr, rrq.Filename)

	// net.Dial로 udp 연결을 맺어 별도 확인 없이 이 때 주어진 주소에 대해서만 통신되도록 함.
	conn, err := net.Dial("udp", clientAddr)
	if err != nil {
		log.Printf("[%s] dial: %v", clientAddr, err)
		return
	}
	defer func() { _ = conn.Close() }()

	var (
		ackPkt  Ack
		errPkt  Err
		dataPkt = Data{Payload: bytes.NewReader(s.Payload)} // 서버의 payload로 데이터 객체 생성
		buf     = make([]byte, DatagramSize)
	)

NEXTPACKET:
	for n := DatagramSize; n == DatagramSize; {
		// payload로부터 데이터 패킷을 얻어오기
		data, err := dataPkt.MarshalBinary()
		if err != nil {
			log.Printf("[%s] preparing data packet: %v", clientAddr, err)
			return
		}

	RETRY:
		for i := s.Retries; i > 0; i-- {
			// 데이터 패킷 전송
			n, err = conn.Write(data)
			if err != nil {
				log.Printf("[%s] write: %v", clientAddr, err)
				return
			}

			// Client의 ACK 패킷 대기 제한시간 적용
			conn.SetReadDeadline(time.Now().Add(s.Timeout))

			_, err = conn.Read(buf)
			if err != nil {
				// Timeout인 경우 재시도 횟수 내에서 패킷 재전송 시도
				if nErr, ok := err.(net.Error); ok && nErr.Timeout() {
					continue RETRY
				}

				log.Printf("[%s] waiting for ACK: %v", clientAddr, err)
				return
			}

			// read한 데이터가 어떤 패킷인지 switch, unmarshalbinary를 통해 처리
			switch {
			case ackPkt.UnmarshalBinary(buf) == nil:
				if uint16(ackPkt) == dataPkt.Block {
					// 데이터패킷의 block number 일치하면 다음 패킷 전송
					continue NEXTPACKET
				}
			case errPkt.UnmarshalBinary(buf) == nil:
				// 에러 패킷일 경우 데이터 전송 중단
				log.Printf("[%s] received error: %v", clientAddr, errPkt.Message)
				return
			default:
				log.Printf("[%s] bad packet", clientAddr)
			}
		}

		log.Printf("[%s] exhausted retries", clientAddr)
		return
	}

	log.Printf("[%s] sent %d blocks", clientAddr, dataPkt.Block)
}
