package api

import (
	"context"
	"io"
	"log"
	"sync"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kuberoute/orchestrator/internal/clusters"
	"github.com/kuberoute/orchestrator/internal/decisions"
	"github.com/kuberoute/orchestrator/internal/optimize"
	"github.com/kuberoute/orchestrator/internal/snapshot"
	pb "github.com/kuberoute/orchestrator/proto"
)

type AgentServer struct {
	pb.UnimplementedAgentServiceServer
	clusterService  *clusters.Service
	decisionService *decisions.Service
	db              *pgxpool.Pool
	optimizer       *optimize.Client
	mu              sync.Mutex
}

func NewAgentServer(clusterService *clusters.Service, decisionService *decisions.Service, db *pgxpool.Pool, optimizer *optimize.Client) *AgentServer {
	return &AgentServer{
		clusterService:  clusterService,
		decisionService: decisionService,
		db:              db,
		optimizer:       optimizer,
	}
}

func (s *AgentServer) StreamMetrics(stream pb.AgentService_StreamMetricsServer) error {
	ctx := stream.Context()
	for {
		batch, err := stream.Recv()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		clusterID, err := uuid.Parse(batch.ClusterId)
		if err != nil {
			continue
		}

		_ = s.clusterService.Heartbeat(ctx, clusterID, 0, "connected")
		snap := snapshot.FromMetrics(clusterID, batch.Metrics)
		if s.db != nil {
			if err := snapshot.Persist(ctx, s.db, snap); err != nil {
				log.Printf("persist snapshot: %v", err)
			}
		}
		go s.maybeOptimize(clusterID, snap)

		ready, _, err := s.decisionService.List(ctx, clusterID, "ready", 1, 50)
		if err != nil {
			return err
		}
		var pbDecisions []*pb.RoutingDecision
		for _, d := range ready {
			params := d.Params
			if params == nil {
				params = map[string]string{}
			}
			if d.WorkloadKind != "" {
				params["kind"] = d.WorkloadKind
			}
			if d.WorkloadNamespace != "" {
				params["namespace"] = d.WorkloadNamespace
			}
			pbDecisions = append(pbDecisions, &pb.RoutingDecision{
				DecisionId:        d.ID.String(),
				DecisionType:      d.DecisionType,
				WorkloadNamespace: d.WorkloadNamespace,
				WorkloadName:      d.WorkloadName,
				TargetNode:        d.TargetNodeName,
				Confidence:        d.ConfidenceScore,
				Reason:            d.Reason,
				Params:            params,
			})
		}
		if err := stream.Send(&pb.DecisionBatch{Decisions: pbDecisions}); err != nil {
			return err
		}
	}
}

func (s *AgentServer) maybeOptimize(clusterID uuid.UUID, snap snapshot.ClusterSnapshot) {
	if s.optimizer == nil || len(snap.Nodes) == 0 || len(snap.Workloads) == 0 {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	ctx := context.Background()
	out, err := s.optimizer.Optimize(ctx, snap)
	if err != nil {
		log.Printf("optimize: %v", err)
		return
	}
	for _, rec := range out {
		ns := rec.Params["namespace"]
		name := rec.Params["workload"]
		target := rec.Params["target_node"]
		if name == "" || target == "" {
			continue
		}
		exists, err := s.decisionService.HasPendingPlacement(ctx, clusterID, ns, name, target)
		if err != nil || exists {
			continue
		}
		d := &decisions.Decision{
			ClusterID:            clusterID,
			DecisionType:         rec.DecisionType,
			Reason:               rec.Reason,
			ConfidenceScore:      rec.Confidence,
			PredictedCostSavings: rec.PredictedCostSavings,
			Params:               rec.Params,
			WorkloadName:         name,
			WorkloadNamespace:    ns,
			WorkloadKind:         rec.Params["kind"],
			TargetNodeName:       target,
			TaskCost:             1,
			SolverTrace:          rec.SolverTrace,
			ProposedPatch:        rec.ProposedPatch,
		}
		if s.db != nil {
			d.TargetNodeID = snapshot.NodeIDByName(ctx, s.db, clusterID, target)
			d.WorkloadID = snapshot.WorkloadIDByName(ctx, s.db, clusterID, ns, name)
		}
		if err := s.decisionService.Create(ctx, d); err != nil {
			log.Printf("persist decision: %v", err)
		}
	}
}

func (s *AgentServer) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	clusterID, err := uuid.Parse(req.ClusterId)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid cluster id")
	}
	if err := s.clusterService.Heartbeat(ctx, clusterID, int(req.NodeCount), req.Status); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	return &pb.HeartbeatResponse{Accepted: true, ConfigVersion: "v1.0.0"}, nil
}

func (s *AgentServer) GetConfig(ctx context.Context, req *pb.ConfigRequest) (*pb.ConfigResponse, error) {
	return &pb.ConfigResponse{
		ConfigVersion: "v1.0.0",
		ConfigData:    []byte(`{"metrics_interval":"15s","buffer_size":1000}`),
	}, nil
}

func (s *AgentServer) ReportEvent(ctx context.Context, req *pb.EventRequest) (*pb.EventResponse, error) {
	log.Printf("Event from %s: [%s] %s", req.ClusterId, req.EventType, req.Message)
	if req.EventType == "decision_executed" {
		if id, err := uuid.Parse(req.Metadata["decision_id"]); err == nil {
			_ = s.decisionService.MarkExecuted(ctx, id)
		}
	}
	return &pb.EventResponse{Received: true}, nil
}
