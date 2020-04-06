package voyager

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net/http"
)

type vhc struct {
	address string
	client  http.Client
}

func NewHttpNetworkProbeClient(address string) NetworkProbeClient {
	return &vhc{address: address, client: http.Client{}}
}

func (v *vhc) Probe(ctx context.Context, in *ProbeRequest, opts ...grpc.CallOption) (*ProbeResponse, error) {
	var buf bytes.Buffer
	m := &jsonpb.Marshaler{}
	er := m.Marshal(&buf, in)
	if er != nil {
		log.Println(er)
	}
	resp, err := v.client.Post(fmt.Sprintf("http://%s/voyager", v.address), "application/json", &buf)
	if err != nil {
		log.Println(err)
		return nil, err
	}
	var response ProbeResponse
	if resp.StatusCode != http.StatusOK {
		return nil, status.Error(codes.Internal, bufio.NewScanner(resp.Body).Text())
	}
	er = jsonpb.Unmarshal(resp.Body, &response)
	if er != nil {
		return nil, er
	}
	return &response, nil
}
