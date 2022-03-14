package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"

	"github.com/cschleiden/go-workflows/backend"
	"github.com/cschleiden/go-workflows/backend/sqlite"
	"github.com/cschleiden/go-workflows/client"
	"github.com/cschleiden/go-workflows/worker"
	"github.com/cschleiden/go-workflows/workflow"
	"github.com/google/uuid"
	errs "github.com/pkg/errors"
)

func main() {
	ctx := context.Background()

	b := sqlite.NewInMemoryBackend()

	// Run worker
	go RunWorker(ctx, b)

	// Start workflow via client
	c := client.New(b)

	startWorkflow(ctx, c)
	// startWorkflow(ctx, c)

	c2 := make(chan os.Signal, 1)
	signal.Notify(c2, os.Interrupt)
	<-c2
}

func startWorkflow(ctx context.Context, c client.Client) {
	wf, err := c.CreateWorkflowInstance(ctx, client.WorkflowInstanceOptions{
		InstanceID: uuid.NewString(),
	}, Workflow1, "Hello world"+uuid.NewString())
	if err != nil {
		panic("could not start workflow")
	}

	log.Println("Started workflow", wf.GetInstanceID())
}

func RunWorker(ctx context.Context, mb backend.Backend) {
	w := worker.New(mb, nil)

	w.RegisterWorkflow(Workflow1)

	w.RegisterActivity(Activity1)

	if err := w.Start(ctx); err != nil {
		panic("could not start worker")
	}
}

func Workflow1(ctx workflow.Context, msg string) error {
	log.Println("Entering Workflow1")
	log.Println("\tWorkflow instance input:", msg)
	log.Println("\tIsReplaying:", workflow.Replaying(ctx))
	defer func() { log.Println("Leaving Workflow1") }()

	a1 := workflow.ExecuteActivity[int](ctx, workflow.DefaultActivityOptions, Activity1, 35, 12)

	r1, err := a1.Get(ctx)
	if err != nil {
		log.Println("Error from Activity 1", err)
		return errs.Wrap(err, "error getting result from activity 1")
	}
	log.Println("R1 result:", r1)
	log.Println("\tIsReplaying:", workflow.Replaying(ctx))

	log.Println("Completing workflow 1")
	return nil
}

func Activity1(ctx context.Context, a, b int) (int, error) {
	log.Println("Entering Activity1")
	defer func() { log.Println("Leaving Activity1") }()

	return 0, errors.New("some activity error")
}
