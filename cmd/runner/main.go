package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	containerd "github.com/containerd/containerd/v2/client"
	"github.com/containerd/containerd/v2/pkg/namespaces"

	"functiond/pkg"
	"functiond/pkg/runner"
)

func main() {
	log.SetOutput(os.Stdout)
	os.MkdirAll("/etc/functiond", 0777)
	cleanup()
	ctx := namespaces.WithNamespace(context.Background(), "example")

	wsm := pkg.NewManager()
	if err := wsm.Register(ctx, runner.BuildOptions(
		runner.WithWorkerSetName("node-server"),
		runner.WithFile("./lambda.zip"),
		runner.WithDownscaleTimeout(5*time.Second),
	)); err != nil {
		log.Fatal(err)
	}

	defer wsm.Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/{id}/execute", func(writer http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()
		if r.Method != http.MethodPost {
			http.Error(writer, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		//val := map[string]any{}
		//if err := json.NewDecoder(r.Body).Decode(&val); err != nil {
		//	log.Print(err)
		//	http.Error(writer, "Failed to parse json", http.StatusBadRequest)
		//	return
		//}
		//b, _ := json.Marshal(val)

		name := r.PathValue("id")
		w := wsm.RetrieveWorker(name)
		if w == nil {
			http.NotFound(writer, r)
			return
		}
		log.Printf("%+v", w)
		res, err := w.Execute(r.Context(), make([]byte, 1, 2))
		if err != nil {
			log.Print(err)
			http.Error(writer, "Failed to execute", http.StatusBadRequest)
			return
		}

		execResult := <-res
		if execResult.Err != nil {
			http.Error(writer, execResult.Err.Error(), http.StatusBadRequest)
			return
		}

		writer.WriteHeader(200)
		writer.Write(execResult.Result)
	})

	http.ListenAndServe(":8080", mux)
}

//	time.Sleep(5 * time.Second)
//	ws.Execute(context.Background(), []byte(fmt.Sprintf("Hello from the other side x %d time", 11)))
//	time.Sleep(20 * time.Second)
//}
//
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
// 	}10-functiond.conflist
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

	fmt.Println("Cleanup completed successfully")
}

func prepareCNI(ctx context.Context) {
	
}
