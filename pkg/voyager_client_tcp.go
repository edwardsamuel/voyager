package voyager

import (
	"bufio"
	"context"
	"fmt"
	"github.com/golang/protobuf/proto"
	"google.golang.org/grpc"
	"log"
	"net"
	"time"
)

type vtc struct {
	address string
}

func NewTCPNetworkProbeClient(address string) NetworkProbeClient {
	return &vtc{address: address}
}

func (v *vtc) Probe(ctx context.Context, in *ProbeRequest, opts ...grpc.CallOption) (*ProbeResponse, error) {
	conn, e := net.Dial("tcp", v.address)
	if e != nil {
		return nil, e
	}
	reqBytes, e := proto.Marshal(in)
	if e != nil {
		return nil, e
	}
	nn, e := bufio.NewWriter(conn).Write(reqBytes)
	if e != nil {
		return nil, e
	}
	log.Println(fmt.Sprintf("Wrote %d bytes", nn))
	var resp []byte
	for {
		n, e := bufio.NewReader(conn).Read(resp)
		if e != nil {
			return nil, e
		}
		log.Println(fmt.Sprintf("Read %d bytes", n))
		if n > 0 {
			break
		} else {
			time.Sleep(time.Millisecond * 500)
		}
	}
	var pr ProbeResponse
	e = proto.Unmarshal(resp, &pr)
	if e != nil {
		return nil, e
	}
	return &pr, nil
}
