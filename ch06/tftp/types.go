package tftp

import (
	"bytes"
	"encoding/binary"
	"errors"
	"io"
	"strings"
)

const (
    // 파편화를 피하기 위해 데이트그램 크기 작게 제한
    DatagramSize = 516 // 최대 지원하는 데이터그램 크기
    BlockSize = DatagramSize - 4 // DatagramSize - 4바이트 헤더
)

type OpCode uint16
const (
    OpRRQ OpCode = iota + 1 // Read Request
    _ // WRQ 미지원 (read only로 생성할 것이므로)
    OpData
    OpAck
    OpErr
)

type ErrCode uint16
const (
    ErrUnknown ErrCode = iota
    ErrNotFound
    ErrAccessViolation
    ErrDiskFull
    ErrIllegalOp
    ErrUnknownId
    ErrFileExists
    ErrNoUser
)


type ReadReq struct {
    Filename string
    Mode string
}

// Implement encoding.BinaryMarshaler
// 서버에서 사용되지 않지만 클라이언트가 이 method 사용
func (q ReadReq) MarshalBinary() ([]byte, error) {
    mode := "octet"
    if q.Mode != "" {
        mode = q.Mode
    }

    // OP 코드 + 파일명 + null + 모드 정보 + null
    cap := 2 + 2 + len(q.Filename) + 1 + len(q.Mode) + 1

    b := new(bytes.Buffer)
    b.Grow(cap)

    // Write OpCode
    err := binary.Write(b, binary.BigEndian, OpRRQ)
    if err != nil {
        return nil, err
    }

    // Write filename
    _, err = b.WriteString(q.Filename)
    if err != nil {
        return nil, err
    }

    // Write null byte
    err = b.WriteByte(0)
    if err != nil {
        return nil, err
    }

    // Write mode
    _, err = b.WriteString(mode)
    if err != nil {
        return nil, err
    }

    // Write null byte
    err = b.WriteByte(0)
    if err != nil {
        return nil, err
    }

    return b.Bytes(), nil
}

// Implement encoding.BinaryUnmarshaler
func (q *ReadReq) UnmarshalBinary(p []byte) error {
    r := bytes.NewBuffer(p)

    var code OpCode

    err := binary.Read(r, binary.BigEndian, &code)
    if err != nil {
        return err
    }
    if code != OpRRQ {
        return errors.New("invalid RRQ")
    }

    // 파일명 읽기
    q.Filename, err = r.ReadString(0)
    if err != nil {
        return errors.New("invalid RRQ")
    }

    q.Filename = strings.TrimRight(q.Filename, "\x00") // 0바이트 제거
    if len(q.Filename) == 0 {
        return errors.New("invalid RRQ")
    }

    q.Mode, err = r.ReadString(0) // 모든 정보 읽기
    if err != nil {
        return errors.New("invalid RRQ")
    }

    q.Mode = strings.TrimRight(q.Mode, "\x00") // 0바이트 제거
    if len(q.Mode) == 0 {
        return errors.New("invalid RRQ")
    }

    // 예제에서는 octet모드만 사용
    actual := strings.ToLower(q.Mode)
    if actual != "octet" {
        return errors.New("only binary transfers supported")
    }

    return nil
}


type Data struct {
    Block uint16 // Overflow가 발생할 수 있음에 유의하라
    Payload io.Reader // io.Reader를 사용함으로써 페이로드를 어느 소스로부터든 얻어올 수 있도록 구현
}

// Implement encoding.BinaryMarshaler
func (d *Data) MarshalBinary() ([]byte, error) {
    b := new(bytes.Buffer)
    b.Grow(DatagramSize)

    // Increase block number by 1
    d.Block++ 

    err := binary.Write(b, binary.BigEndian, OpData) // OP 코드 쓰기
    if err != nil {
        return nil, err
    }

    err = binary.Write(b, binary.BigEndian, d.Block) // Block 번호 쓰기
    if err != nil {
        return nil, err
    }

    // BlockSize 크기만큼 쓰기
    _, err = io.CopyN(b, d.Payload, BlockSize)
    if err != nil  && err != io.EOF {
        return nil, err
    }

    return b.Bytes(), nil
}

// Implement encoding.BinaryUnmarshaler
func (d *Data) UnmarshalBinary(p []byte) error {
    if l := len(p); l < 4 || l > DatagramSize { // 패킷 사이즈 체크
        return errors.New("invalid DATA")
    }

    // OpCode 확인
    var opcode OpCode
    err := binary.Read(bytes.NewReader(p[:2]), binary.BigEndian, &opcode)
    if err != nil || opcode != OpData {
        return errors.New("invalid Data")
    }

    // Block Number 확인
    err = binary.Read(bytes.NewReader(p[2:4]), binary.BigEndian, &d.Block)
    if err != nil {
        return errors.New("invalid Data")
    }

    // 남은 bytes 새로운 버퍼에 넣고 Payload 필드에 할당
    d.Payload = bytes.NewBuffer(p[4:])

    return nil
}


type Ack uint16

// Implement encoding.BinaryMarshaler
func (a Ack) MarshalBinary() ([]byte, error) {
    cap := 2 + 2 // OP 코드 + 블록 번호

    b := new(bytes.Buffer)
    b.Grow(cap)

    // Write OP 코드
    err := binary.Write(b, binary.BigEndian, OpAck)
    if err != nil {
        return nil, err
    }

    // Write Block Number
    err = binary.Write(b, binary.BigEndian, a)
    if err != nil {
        return nil, err
    }

    return b.Bytes(), nil
}

// Implement encoding.BinaryUnmarshaler
func (a *Ack) UnmarshalBinary(p []byte) error {
    var code OpCode

    r := bytes.NewReader(p)

    // Check OpCode
    err := binary.Read(r, binary.BigEndian, &code)
    if err != nil {
        return err
    }
    if code != OpAck  {
        return errors.New("invalid ACK")
    }

    // Read Block Number
    return binary.Read(r, binary.BigEndian, a)
}


type Err struct {
    Error ErrCode
    Message string
}

// Implement encoding.BinaryMarshaler
func (e Err) MarshalBinary() ([]byte, error) {
    // OP 코드 + Error 코드 + 메세지 + null 바이트
    cap := 2 + 2 + len(e.Message) + 1

    b := new(bytes.Buffer)
    b.Grow(cap)

    // Write OpCode
    err := binary.Write(b, binary.BigEndian, OpErr)
    if err != nil {
        return nil, err
    }

    // Write Error Code
    err = binary.Write(b, binary.BigEndian, e.Error)
    if err != nil {
        return nil, err
    }

    // Write Message
    _, err = b.WriteString(e.Message)
    if err != nil {
        return nil, err
    }

    // null 바이트 쓰기
    err = b.WriteByte(0)
    if err != nil {
        return nil, err
    }

    return b.Bytes(), nil
}

// Implement encoding.BinaryUnmarshaler
func (e *Err) UnmarshalBinary(p []byte) error {
    r := bytes.NewBuffer(p)

    // Check OpCOde
    var code OpCode
    err := binary.Read(r, binary.BigEndian, &code)
    if err != nil {
        return err
    }
    if code != OpErr {
        return errors.New("invalid ERROR")
    }

    // Read Error Code
    err = binary.Read(r, binary.BigEndian, e.Error)
    if err != nil {
        return err
    }

    // Read Message
    e.Message, err = r.ReadString(0)
    if err != nil {
        return err
    }
    e.Message = strings.TrimRight(e.Message, "\x00")

    return nil
}
