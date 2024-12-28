package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/leaderelection"
	"k8s.io/client-go/tools/leaderelection/resourcelock"
)

func main() {
	podName := os.Getenv("POD_NAME")
	if podName == "" {
		log.Fatal("POD_NAME environment variable not set")
	}

	config, err := rest.InClusterConfig()
	if err != nil {
		log.Fatal(err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.Background()
	runLeaderElection(ctx, clientset, podName)
}

func updateLeaderFile(leader string) {
	content := fmt.Sprintf("%s: Leader is %s\n", time.Now().Format(time.RFC3339), leader)
	if err := os.WriteFile("/data/winner.txt", []byte(content), 0644); err != nil {
		log.Printf("Error writing to file: %v", err)
	}
}

func runLeaderElection(ctx context.Context, clientset *kubernetes.Clientset, podName string) {
	lock := &resourcelock.LeaseLock{
		LeaseMeta: metav1.ObjectMeta{
			Name:      "leader-election-lock",
			Namespace: "default",
		},
		Client: clientset.CoordinationV1(),
		LockConfig: resourcelock.ResourceLockConfig{
			Identity: podName,
		},
	}

	leaderelection.RunOrDie(ctx, leaderelection.LeaderElectionConfig{
		Lock:            lock,
		ReleaseOnCancel: true,
		LeaseDuration:   15 * time.Second,
		RenewDeadline:   10 * time.Second,
		RetryPeriod:     2 * time.Second,
		Callbacks: leaderelection.LeaderCallbacks{
			OnStartedLeading: func(ctx context.Context) {
				log.Printf("Pod %s became leader", podName)
				runAsLeader(ctx, podName)
			},
			OnStoppedLeading: func() {
				log.Printf("Pod %s stopped leading", podName)
			},
			OnNewLeader: func(identity string) {
				log.Printf("New leader is %s", identity)
				updateLeaderFile(identity)
			},
		},
	})
}

func runAsLeader(ctx context.Context, podName string) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			updateLeaderFile(podName)
		}
	}
}
