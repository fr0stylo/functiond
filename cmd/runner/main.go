package main

import (
	"context"
	"functiond/pkg/orchestration"
	"functiond/pkg/runtime/shell"
	"log/slog"
	"os"
	"sync"
	"time"
)

func main() {
	m := orchestration.NewManager()
	ws, err := orchestration.NewWorkerSet(
		context.Background(),
		orchestration.WithConcurency(50),
		orchestration.WithWorkerFactory(shell.NewFactory()),
	)

	if err != nil {
		slog.Error("Error while creating WS", "error", err)
		os.Exit(1)
	}

	m.Register(context.Background(), "test",
		ws,
	)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Go(func() {
			w := m.RetrieveWorker("test")
			if err != nil {
				slog.Error("Error while creating WS", "error", err)
				os.Exit(1)
			}
			r, err := w.Execute(context.Background(), []byte("sleep $(( RANDOM % 10 + 1 )) && date"))
			if err != nil {
				slog.Error("Error while calling execute", "error", err)
				os.Exit(1)
			}
			<-r
		})
	}

	wg.Wait()

	time.Sleep(30 * time.Second)
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
