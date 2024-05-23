package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"os/user"
	"path/filepath"
	"network-study-with-go/ch07/creds/auth"
)

func init() {
	flag.Usage = func() {
		_, _ = fmt.Fprintf(
			flag.CommandLine.Output(),
            "Usage:\n\t%s <group names>\n",
            filepath.Base(os.Args[0]),
		)
		flag.PrintDefaults()
	}
}

func parseGroupNames(args []string) map[string]struct{} {
    groups := make(map[string]struct{})

    for _, arg := range args {
        grp, err := user.LookupGroup(arg)
        if err != nil {
            log.Println(err)
            continue
        }

        groups[grp.Gid] = struct{}{}
    }

    return groups
}

func main() {
    flag.Parse()

    groups := parseGroupNames(flag.Args())
    socket := filepath.Join(os.TempDir(), "creds.sock")
    addr, err := net.ResolveUnixAddr("unix", socket)
    if err != nil {
        log.Fatal(err)
    }

    s, err := net.ListenUnix("unix", addr)
    if err != nil {
        log.Fatal(err)
    }

    // ListenUnix를 이용했음에도, Ctrl+C (Interrupt Signal)을 받으면
    // 즉시 종료되어 Socket 파일을 제거하지 못하게 된다.
    // 따라서 go routine을 통해 interrupt signal을 받을 시
    // close를 통해 socket 파일을 제거하도록 한다.
    c := make(chan os.Signal, 1)
    signal.Notify(c, os.Interrupt)
    go func() {
        <-c
        _ = s.Close()
    }()

    fmt.Printf("Listening on %s ...\n", socket)
    for {
        conn, err := s.AcceptUnix()
        if err != nil {
            break
        }

        if auth.Allowed(conn, groups) {
            _, err = conn.Write([]byte("Welcome\n"))
            if err == nil {
                continue
            }
        } else {
            _, err = conn.Write([]byte("Access denied\n"))
        }

        if err != nil {
            log.Println(err)
        }
        conn.Close()
    }
}
