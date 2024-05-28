package auth

import (
	"log"
	"net"
	"os/user"
	"fmt"

	"golang.org/x/sys/unix"
)

// net.UnixConn: 유닉스 도메인 소켓 연결 객체
// file descriptor를 얻기 위해 필요
func Allowed(conn *net.UnixConn, groups map[string]struct{}) bool {
	if conn == nil || groups == nil || len(groups) == 0 {
		return false
	}

	// Copy the underlying file descriptor of net.Conn, and return it
	file, _ := conn.File()
	defer func() { file.Close() }()

	var (
		err   error
		ucred *unix.Ucred
	)

	// Ucred 구조
	// type Ucred struct {
	// 	Pid int32 // process ID
	// 	Uid uint32 // user ID
	// 	Gid uint32 // group ID
	// }

	for {
		ucred, err = unix.GetsockoptUcred(
			int(file.Fd()),   // file의 descriptor
			unix.SOL_SOCKET,  // 어느 프로토콜 계층에 속하였는지
			unix.SO_PEERCRED, // 옵션 값
		)
		if err == unix.EINTR {
			continue // syscall로 인해 중단됨 => 다시 시도
		} else if err != nil {
			log.Println(err)
			return false
		}

		break
	}

	// Get User Information from user ID
	u, err := user.LookupId(fmt.Sprint(ucred.Uid))
	if err != nil {
		log.Println(err)
		return false
	}

	// Get Group IDs from User Information
	gids, err := u.GroupIds()
	if err != nil {
		log.Println(err)
		return false
	}

	for _, gid := range gids {
		if _, found := groups[gid]; found {
			return true
		}
	}

	return false
}
