package main

import (
	"context"
	"crypto/tls"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/kuberoute/agent/internal/collector"
	"github.com/kuberoute/agent/internal/executor"
	pb "github.com/kuberoute/agent/proto"
)

func main() {
	clusterID := os.Getenv("KUBEROUTE_CLUSTER_ID")
	if clusterID == "" {
		log.Fatal("KUBEROUTE_CLUSTER_ID required")
	}

	endpoint := os.Getenv("KUBEROUTE_API_ENDPOINT")
	if endpoint == "" {
		endpoint = "localhost:50051"
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	kubeClient, err := newKubeClient()
	if err != nil {
		log.Printf("kubernetes client unavailable (%v); scrape/apply will no-op until kubeconfig is present", err)
	}

	opts, err := dialOpts()
	if err != nil {
		log.Fatal(err)
	}

	conn, err := grpc.Dial(endpoint, opts...)
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := pb.NewAgentServiceClient(conn)
	metricsCollector := collector.New(15*time.Second, kubeClient)
	exec := executor.New(kubeClient)

	stream, err := client.StreamMetrics(ctx)
	if err != nil {
		log.Fatalf("failed to start stream: %v", err)
	}

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_, err := client.Heartbeat(ctx, &pb.HeartbeatRequest{
					ClusterId:    clusterID,
					AgentVersion: "v1.0.0",
					NodeCount:    metricsCollector.NodeCount(),
					Status:       "connected",
				})
				if err != nil {
					log.Printf("heartbeat failed: %v", err)
				}
			}
		}
	}()

	go func() {
		for batch := range metricsCollector.Collect(ctx, clusterID) {
			if err := stream.Send(batch); err != nil {
				log.Printf("send error: %v", err)
				continue
			}
			decisions, err := stream.Recv()
			if err != nil {
				continue
			}
			for _, d := range decisions.Decisions {
				if err := exec.Apply(ctx, d); err != nil {
					log.Printf("failed to apply decision %s: %v", d.DecisionId, err)
					continue
				}
				_, _ = client.ReportEvent(ctx, &pb.EventRequest{
					ClusterId: clusterID,
					EventType: "decision_executed",
					Message:   "applied " + d.DecisionType,
					Metadata:  map[string]string{"decision_id": d.DecisionId},
					Timestamp: time.Now().Unix(),
				})
			}
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("agent shutting down")
}

func dialOpts() ([]grpc.DialOption, error) {
	if os.Getenv("KUBEROUTE_INSECURE") == "true" {
		return []grpc.DialOption{grpc.WithTransportCredentials(insecure.NewCredentials())}, nil
	}
	tlsCfg := &tls.Config{}
	if os.Getenv("KUBEROUTE_TLS_SKIP_VERIFY") == "true" {
		log.Println("warning: KUBEROUTE_TLS_SKIP_VERIFY=true; certificate verification is disabled")
		tlsCfg.InsecureSkipVerify = true
	}
	return []grpc.DialOption{grpc.WithTransportCredentials(credentials.NewTLS(tlsCfg))}, nil
}

func newKubeClient() (kubernetes.Interface, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		kubeconfig := os.Getenv("KUBECONFIG")
		if kubeconfig == "" {
			kubeconfig = clientcmd.RecommendedHomeFile
		}
		cfg, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
		if err != nil {
			return nil, err
		}
	}
	return kubernetes.NewForConfig(cfg)
}
