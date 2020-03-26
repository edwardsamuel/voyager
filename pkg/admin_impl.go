package voyager

import (
	"context"
	"github.com/golang/protobuf/ptypes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"log"
	"math/rand"
	"strings"
	"sync"
	"time"
)

type adminImpl struct {
}

func NewAdminImpl() AdminServer {
	return &adminImpl{}
}

type worker struct {
	client NetworkProbeClient
	conn   *grpc.ClientConn
}

func (w *worker) Close() error {
	return w.conn.Close()
}

type unitResp struct {
	nodeName    string
	podName     string
	expectError bool
	isError     bool
}

func (w *worker) do(numReq int64, reqSize int64, respSize int64, errRate float32, panicRate float32) <-chan *unitResp {
	respChan := make(chan *unitResp, 10)
	go func() {
		defer close(respChan)
		ctx := context.Background()
		for i := 0; i < int(numReq); i++ {
			req := &ProbeRequest{
				ResponseSize:     respSize,
				Meta:             &Meta{Data: generateRandomData(reqSize)},
				ShouldKillServer: rand.Float32() < panicRate,
			}
			if rand.Float32() < errRate {
				req.ResponseCode = int32(codes.Internal)
			}
			r, er := w.client.Probe(ctx, req)
			ur := &unitResp{
				expectError: req.ShouldKillServer || req.ResponseCode > 0,
			}
			if er != nil {
				ur.isError = true
			}
			if r != nil {
				ur.podName = r.PodName
				ur.nodeName = r.NodeName
			}
			respChan <- ur
		}
	}()
	return respChan
}

type stat struct {
	totalRequest int64
	success      int64
	failure      int64
}

type stats struct {
	nodePodStats map[string]map[string]*stat
}

func (s *stats) add(ur *unitResp) {
	nn := ur.nodeName
	pn := ur.podName
	if nn == "" {
		nn = "empty"
	}
	if pn == "" {
		pn = "empty"
	}
	if s.nodePodStats == nil {
		s.nodePodStats = make(map[string]map[string]*stat)
	}
	if s.nodePodStats[nn] == nil {
		s.nodePodStats[nn] = make(map[string]*stat)
	}
	if s.nodePodStats[nn][pn] == nil {
		s.nodePodStats[nn][pn] = &stat{}
	}
	s.nodePodStats[nn][pn].totalRequest++
	if ur.isError && !ur.expectError {
		s.nodePodStats[nn][pn].failure++
	} else {
		s.nodePodStats[nn][pn].success++
	}

}
func (s *stats) response() *NetworkProbeStats {
	nps := new(NetworkProbeStats)
	for node, podStats := range s.nodePodStats {
		for pod, stats := range podStats {
			if nps.Stats == nil {
				nps.Stats = &NetworkProbeStats_Stats{}
			}
			nps.Stats.NumRequestsCompleted += stats.totalRequest
			nps.Stats.NumSuccess += stats.success
			nps.Stats.NumFailure += stats.failure
			nps.HostSplits = append(nps.HostSplits, &NetworkProbeStats_HostSplit{
				NodeName: node,
				PodName:  pod,
				Stats: &NetworkProbeStats_Stats{
					NumRequestsCompleted: stats.totalRequest,
					NumSuccess:           stats.success,
					NumFailure:           stats.failure,
					Latencies:            nil,
				},
			})
		}
	}
	nps.CreatedAt = ptypes.TimestampNow()
	return nps
}

func (a *adminImpl) StartProbe(req *InitiateNetworkProbe, stream Admin_StartProbeServer) error {
	w := &worker{}
	if conn, err := grpc.Dial(strings.Join(req.HostPorts, ","), grpc.WithInsecure()); err != nil {
		return err
	} else {
		w.conn = conn
		w.client = NewNetworkProbeClient(w.conn)
	}

	aggChan := make(chan *unitResp, 100)
	wg := sync.WaitGroup{}
	for i := 0; i < int(req.NumThreads); i++ {
		log.Println("Spawning worker", i)
		wg.Add(1)
		rC := w.do(req.NumberRequests, req.RequestSize, req.ResponseSize, req.ErrorRate, req.PanicRate)
		go func() {
			defer wg.Done()
			for r := range rC {
				aggChan <- r
			}
			log.Println("Completed worker")
		}()
	}
	go func() {
		wg.Wait()
		close(aggChan)
		log.Println("All workers returned")
	}()
	s := new(stats)
	t := time.NewTicker(time.Second * 1)
	defer t.Stop()
	for {
		select {
		case _ = <-t.C:
			if er := stream.Send(s.response()); er != nil {
				return er
			}
		case ur, ok := <-aggChan:
			if ok {
				s.add(ur)
			} else {
				log.Println("Agg chan closed")
				aggChan = nil
			}
		}
		if aggChan == nil {
			if er := stream.Send(s.response()); er != nil {
				return er
			}
			break
		}
	}
	return nil
}
