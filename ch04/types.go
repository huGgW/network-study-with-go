package ch04

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"io"
)

// Constants

const (
    BinaryType uint8 = iota + 1
    StringType

    MaxPayloadSize uint32 = 10 << 20 // 10MB
)

var ErrMaxPayloadSize = errors.New("maximum payload size exceeded")


// Interface Definition for Payload types

type Payload interface {
    fmt.Stringer
    io.ReaderFrom
    io.WriterTo
    Bytes() []byte
}


// Binary TVL Payload Data Type

type Binary []byte

func (m Binary) Bytes() []byte { return m }

func (m Binary) String() string { return string(m) }

func (m Binary) WriteTo(w io.Writer) (int64, error) {
    var n int64 = 0 // 총 읽은 byte 수 트래킹
    err := binary.Write(w, binary.BigEndian, BinaryType) // 1byte를 사용해 타입 기록
    if err != nil {
        return n, err
    }
    n += 1


    err = binary.Write(w, binary.BigEndian, uint32(len(m))) // 4byte를 사용해 데이터 크기 기록
    if err != nil {
        return n, err

    }
    n += 4

    // 실제 데이터 write 후 반환
    o, err := w.Write(m)
    return n + int64(o), err // 실제 데이터 write 크기까지 트래킹
}

func (m *Binary) ReadFrom(r io.Reader) (int64, error) {
    var n int64 = 0 // 총 읽은 byte 수 트래킹

    // 초기 1byte로부터 type 읽기
    var typ uint8
    err := binary.Read(r, binary.BigEndian, &typ)
    if err != nil {
        return n, err
    }
    n += 1

    if typ != BinaryType {
        return n, errors.New("invalid Binary Type")
    }

    // 이후 4byte로부터 size 읽기
    var size uint32
    err = binary.Read(r, binary.BigEndian, &size)
    if err != nil {
        return n, err
    }
    n += 4

    // 최대 payload 사이즈를 정하여 악의적인 공격으로 인해 ram을 모두 소비하는 것을 방지
    if size > MaxPayloadSize { 
        return n, ErrMaxPayloadSize
    }

    // 읽은 size만큼 buffer를 생성 후 읽어서 반환.
    *m = make([]byte, size)
    o, err := r.Read(*m)
    return n + int64(o), err
}


// String TVL Payload Data Type

type String string

func (m String) Bytes() []byte { return []byte(m) }

func (m String) String() string { return string(m) }

func (m String) WriteTo(w io.Writer) (int64, error) {
    // 총 읽은 byte 수 트래킹
    var n int64 = 0

    // 초기 1byte에 type 기록
    err := binary.Write(w, binary.BigEndian, StringType)
    if err != nil {
        return n, err
    }
    n += 1

    // string type은 기본적으로 immutable한 []byte이다.
    // 따라서 len이나 index를 이용한 접근은 기본적으로 []byte와 동일하다 생각 가능
    // 그러나 range를 이용한 iteration의 경우 UTF-8 한 문자씩 읽어오는 []rune type과 같게 동작.

    // 이후 4byte에 문자열의 크기 기록
    err = binary.Write(w, binary.BigEndian, uint32(len(m)))
    if err != nil {
        return n, err
    }
    n += 4

    // 데이터를 writer에 기록 후 반환
    o, err := w.Write([]byte(m))
    return n + int64(o), err // write한 데이터의 크기까지 트래킹
}

func (m *String) ReadFrom(r io.Reader) (int64, error) {
    // 총 읽은 byte 수 트래킹
    var n int64 = 0

    // 초기 1byte로부터 타입 읽기
    var typ byte
    err := binary.Read(r, binary.BigEndian, &typ)
    if err != nil {
        return n, err
    }
    n += 1

    if typ != StringType {
        return n, errors.New("Invalid String")
    }

    // 이후 4byte로부터 문자열의 크기 읽기
    var size uint32
    err = binary.Read(r, binary.BigEndian, &size)
    if err != nil {
        return n, err
    }
    n += 4

    buf := make([]byte, size)
    o, err := r.Read(buf)
    if err != nil {
        return n, err
    }

    *m = String(buf) // buf를 String type으로 convert하여 m에 할당
    return n + int64(o), err
}


// reader에서 byte를 읽어 Binary, String 타입으로 디코딩
func decode(r io.Reader) (Payload, error) {
    var typ byte
    // type 추론을 위해 1byte를 읽음
    err := binary.Read(r, binary.BigEndian, &typ)
    if err != nil {
        return nil, err
    }

    // new: Allocate memory for given type with zero values, and return pointer
    // 읽은 type을 중심으로 적절한 type의 payload로 초기화
    var payload Payload
    switch uint8(typ) {
    case BinaryType:
        payload = new(Binary) 
    case StringType:
        payload = new(String)
    default:
        return nil, errors.New("unknown type")
    }

    _, err = payload.ReadFrom(
        // 이미 앞에서 1byte를 읽었기 때문에, 해당 부분을 앞에 reader로 제공하여
        // io.MultiReader를 통해 원래의 byte 전체를 읽는 것과 동일하도록 한다.
        io.MultiReader(bytes.NewReader([]byte{typ}), r),
    )
    if err != nil {
        return nil, err
    }

    return payload, nil
}
