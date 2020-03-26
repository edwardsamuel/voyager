package voyager

import (
	"context"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"math/rand"
	"os"
	"time"
)

var (
	nodeName string
	podName  string
)

func init() {
	if nn := os.Getenv("NODE_NAME"); nn != "" {
		nodeName = nn
	} else if hn, er := os.Hostname(); er == nil {
		nodeName = hn
	} else {
		nodeName = "unknown"
	}
	if pn := os.Getenv("POD_NAME"); pn != "" {
		podName = pn
	} else {
		podName = "unknown"
	}
}

type networkProbeImpl struct {
}

func NewGrpcProbe() NetworkProbeServer {
	return &networkProbeImpl{}
}

func generateRandomData(size int64) []byte {
	b := make([]byte, size)
	rand.Read(b)
	return b

}
func (n *networkProbeImpl) Probe(ctx context.Context, req *ProbeRequest) (*ProbeResponse, error) {
	st := time.Now()
	time.Sleep(time.Millisecond * 300)
	resp := &ProbeResponse{
		NodeName: nodeName,
		PodName:  podName,
	}
	pst, er := ptypes.TimestampProto(st)
	if er == nil {
		resp.ServerTime = pst
	} else {
		log.Println(er)
	}
	if req.ShouldKillServer {
		log.Println("Killing myself")
		panic("Self destruct")
	}
	if req.ResponseSize > 0 {
		resp.Meta = &Meta{Data: generateRandomData(req.ResponseSize)}
	}
	if req.ResponseCode > 0 {
		return resp, status.Errorf(codes.Code(req.ResponseCode), "Asked to return error")
	}
	return resp, nil
}
