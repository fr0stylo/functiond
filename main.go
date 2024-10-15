package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/namespaces"

	"functiond/pkg/runner"
)

func main() {
	os.MkdirAll("/etc/functiond", 0777)
	cleanup()
	ws, err := runner.NewWorkerSet(runner.WithWorkerSetName("node-server"), runner.WithFile("./lambda.zip"))
	if err != nil {
		log.Fatal(err)
	}

	for i := 0; i < 1; i++ {
		t := time.Now()
		if err := ws.Start(); err != nil {
			log.Fatal(err)
		}
		log.Printf("Start time: %s", time.Since(t))
	}

	for i := 0; i < 100; i++ {
		ws.Execute(context.Background(), []byte(fmt.Sprintf("Hello from the other side x %d time", i)))
	}
	defer ws.Shutdown()

	time.Sleep(5 * time.Second)
	ws.Execute(context.Background(), []byte(fmt.Sprintf("Hello from the other side x %d time", 11)))
	time.Sleep(20 * time.Second)
}

// func CreateSubscriber(ctx context.Context, stream jetstream.Stream, name string) jetstream.ConsumeContext {
// 	// retrieve consumer handle from a stream
// 	cons, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
// 		Durable:       "foo",
// 		FilterSubject: "foo",
// 		DeliverPolicy: jetstream.DeliverAllPolicy,
// 	})
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	// consume messages from the consumer in callback
// 	cc, err := cons.Consume(func(msg jetstream.Msg) {
// 		fmt.Printf("[%s]Received jetstream message: %s\n", name, string(msg.Data()))
// 		msg.Ack()
// 	})
// 	if err != nil {
// 		log.Fatal(err)
// 	}
// 	return cc
// }

func cleanup() {
	client, _ := containerd.New("/run/containerd/containerd.sock")
	defer client.Close()
	ctx := namespaces.WithNamespace(context.Background(), "example")

	cl, _ := client.Containers(ctx)
	for _, container := range cl {
		if task, err := container.Task(ctx, nil); err == nil {
			if _, err := task.Delete(ctx, containerd.WithProcessKill); err != nil {
				fmt.Printf("Failed to delete task for container %s: %v\n", container.ID(), err)
			} else {
				fmt.Printf("Deleted task for container: %s\n", container.ID())
			}
		}

		if err := container.Delete(ctx, containerd.WithSnapshotCleanup); err != nil {
			fmt.Printf("Failed to delete container %s: %v\n", container.ID(), err)
		} else {
			fmt.Printf("Deleted container: %s\n", container.ID())
		}
	}

	//snapshotter := client.SnapshotService("overlayfs")

	//// List and remove all snapshots
	//if err := snapshotter.Walk(ctx, func(ctx context.Context, info snapshots.Info) error {
	//	if err := snapshotter.Remove(ctx, info.Name); err != nil {
	//		fmt.Printf("Failed to remove snapshot %s: %v\n", info.Name, err)
	//	} else {
	//		fmt.Printf("Removed snapshot: %s\n", info.Name)
	//	}
	//	return nil
	//}); err != nil {
	//	log.Fatalf("Failed to walk snapshots: %v", err)
	//}

	fmt.Println("Cleanup completed successfully")
}

//client, err := containerd.New("/run/containerd/containerd.sock")
//cni, err := gocni.New(
//	gocni.WithMinNetworkCount(2),
//	gocni.WithPluginConfDir("/etc/cni/net.d"),
//	gocni.WithPluginDir([]string{"/opt/cni/bin"}),
//	gocni.WithInterfacePrefix("eth"),
//	gocni.WithConfListFile("/etc/cni/net.d/10-functiond.conflist"))
//if err != nil {
//	log.Fatalf("failed to initialize CNI: %v", err)
//}
//
//ctx := namespaces.WithNamespace(context.Background(), "example")
//
//worker := runner.NewWorker(client, cni)
//if err := worker.Start(ctx); err != nil {
//	log.Fatal(err)
//}
//time.Sleep(10 * time.Second)
//if err := worker.Shutdown(ctx); err != nil {
//	log.Fatal(err)
//}
//if err := redisExample(); err != nil {
//	log.Fatal(err)
//}
//}
//
//func redisExample() error {
//	// create a new client connected to the default socket path for containerd
//	client, err := containerd.New("/run/containerd/containerd.sock")
//	if err != nil {
//		return err
//	}
//	defer client.Close()
//
//	// create a new context with an "example" namespace
//	ctx := namespaces.WithNamespace(context.Background(), "example")
//
//	// pull the redis image from DockerHub
//	image, err := client.Pull(ctx, "docker.io/library/redis:alpine", containerd.WithPullUnpack)
//	if err != nil {
//		return err
//	}
//
//	// create a container
//	container, err := client.NewContainer(
//		ctx,
//		"redis-server",
//		containerd.WithImage(image),
//		containerd.WithNewSnapshot("redis-server-snapshot", image),
//		containerd.WithNewSpec(oci.WithImageConfig(image)),
//	)
//	if err != nil {
//		return err
//	}
//	defer container.Delete(ctx, containerd.WithSnapshotCleanup)
//
//	// create a task from the container
//	task, err := container.NewTask(ctx, cio.NewCreator(cio.WithStdio))
//	if err != nil {
//		return err
//	}
//	defer task.Delete(ctx)
//
//	// make sure we wait before calling start
//	exitStatusC, err := task.Wait(ctx)
//	if err != nil {
//		return err
//	}
//
//	// call start on the task to execute the redis server
//	if err := task.Start(ctx); err != nil {
//		return err
//	}
//
//	pid := task.Pid()
//	fmt.Printf("Container started with PID %d\n", pid)
//
//	//cniPath := "/opt/cni/bin"
//	//cniConfigDir := "/etc/cni/net.d"
//	cni, err := gocni.New(
//		gocni.WithMinNetworkCount(2),
//		gocni.WithPluginConfDir("/etc/cni/net.d"),
//		gocni.WithPluginDir([]string{"/opt/cni/bin"}),
//		gocni.WithInterfacePrefix("eth"),
//		gocni.WithConfListFile("/etc/cni/net.d/10-functiond.conflist"))
//	if err != nil {
//		log.Fatalf("failed to initialize CNI: %v", err)
//	}
//
//	// Load the CNI configuration
//	if err := cni.Load(
//		gocni.WithLoNetwork,
//		gocni.WithDefaultConf); err != nil {
//		log.Fatalf("failed to load CNI configuration: %v", err)
//	}
//
//	// Get the container's network namespace path
//	netNS := fmt.Sprintf("/proc/%d/ns/net", pid)
//
//	labels := map[string]string{
//		"RUNNING":       "example",
//		"IgnoreUnknown": "1",
//	}
//	// Attach the container to the network
//	result, err := cni.Setup(ctx, "example", netNS, gocni.WithLabels(labels), gocni.WithCapabilityPortMap([]gocni.PortMapping{{
//		HostPort:      6379,
//		ContainerPort: 6379,
//		Protocol:      "tcp",
//	}}))
//	if err != nil {
//		log.Fatalf("failed to attach network to container: %v", err)
//	}
//	defaultIfName := "eth0"
//
//	log.Println(result)
//	IP := result.Interfaces[defaultIfName].IPConfigs[0].IP.String()
//	fmt.Printf("IP of the default interface %s:%s", defaultIfName, IP)
//	// Wait for a bit to allow the container to serve requests
//	fmt.Println("Container is running and attached to network. Press Ctrl+C to stop...")
//
//	// sleep for a lil bit to see the logs
//	time.Sleep(60 * time.Second)
//	// Tear down the network when the container is done
//	if err := cni.Remove(ctx, "example", netNS, gocni.WithLabels(labels), gocni.WithCapabilityPortMap([]gocni.PortMapping{{
//		HostPort:      6379,
//		ContainerPort: 6379,
//		Protocol:      "tcp",
//	}})); err != nil {
//		log.Fatalf("failed to remove network: %v", err)
//	}
//
//	// kill the process and get the exit status
//	if err := task.Kill(ctx, syscall.SIGTERM); err != nil {
//		return err
//	}
//
//	// wait for the process to fully exit and print out the exit status
//	status := <-exitStatusC
//	code, _, err := status.Result()
//	if err != nil {
//		return err
//	}
//	fmt.Printf("redis-server exited with status: %d\n", code)
//
//	return nil
//}
