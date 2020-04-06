package voyager

import (
	"bufio"
	"context"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"github.com/golang/protobuf/proto"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"math/rand"
	"net"
	"net/http"
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

func NewTCPProbe() func(conn net.Conn) error {
	npi := networkProbeImpl{}
	return func(conn net.Conn) error {
		var buf []byte
		n, err := bufio.NewReader(conn).Read(buf)
		if err != nil {
			return err
		}
		log.Println(fmt.Sprintf("Read n: %d bytes", n, ))
		var req ProbeRequest
		err = proto.Unmarshal(buf, &req)
		if err != nil {
			return err
		}
		response, err := npi.Probe(context.Background(), &req)
		if err != nil {
			return err
		}
		bytes, err := proto.Marshal(response)
		if err != nil {
			return err
		}
		nn, err := bufio.NewWriter(conn).Write(bytes)
		if err != nil {
			return err
		}
		log.Println(fmt.Sprintf("Write %d bytes", nn))
		return nil
	}

}
func NewHttpProbe() http.Handler {
	return &networkProbeImpl{}
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

func (n *networkProbeImpl) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	pr := &ProbeRequest{}
	e := jsonpb.Unmarshal(req.Body, pr)
	if e != nil {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte(e.Error()))
	}
	response, e := n.Probe(req.Context(), pr)
	if e != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(e.Error()))
	}
	m := &jsonpb.Marshaler{}
	_ = m.Marshal(w, response)
}
