package ch03

import (
    "net"
    "testing"
)

func TestListener(t *testing.T) {
    // net.Listen
    // Bind to given network and address
    // if ip / port is not given, all unicast and broadcast addr / random port
    // IP를 지정하는 경우, 해당 NIC에 binding 된 IP로만 요청을 받겠다!!
    listener, err := net.Listen("tcp", "127.0.0.1:0")
    if err != nil {
        t.Fatal(err)
    }
    // close the connection before exit the funciton
    defer func() { listener.Close() }()

    t.Logf("bound to %q", listener.Addr())
}
