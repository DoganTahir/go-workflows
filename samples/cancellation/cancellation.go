package main

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/cschleiden/go-workflows/backend"
	"github.com/cschleiden/go-workflows/backend/mysql"
	"github.com/cschleiden/go-workflows/client"
	"github.com/cschleiden/go-workflows/samples"
	"github.com/cschleiden/go-workflows/worker"
	"github.com/cschleiden/go-workflows/workflow"
	"github.com/google/uuid"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	//b := sqlite.NewInMemoryBackend()
	//b := sqlite.NewSqliteBackend("cancellation.sqlite")
	b := mysql.NewMysqlBackend("localhost", 3306, "root", "SqlPassw0rd", "cancellation")

	// Run worker
	go RunWorker(ctx, b)

	// Start workflow via client
	c := client.New(b)

	startWorkflow(ctx, c)

	c2 := make(chan os.Signal, 1)
	signal.Notify(c2, os.Interrupt)
	<-c2

	cancel()
}

func startWorkflow(ctx context.Context, c client.Client) {
	wf, err := c.CreateWorkflowInstance(ctx, client.WorkflowInstanceOptions{
		InstanceID: uuid.NewString(),
	}, Workflow1, "Hello world")
	if err != nil {
		panic("could not start workflow")
	}

	log.Println("Started workflow", wf.GetInstanceID())

	time.Sleep(2 * time.Second)

	if err := c.CancelWorkflowInstance(ctx, wf); err != nil {
		panic("could not cancel workflow")
	}

	log.Println("Canceled workflow", wf.GetInstanceID())
}

func RunWorker(ctx context.Context, mb backend.Backend) {
	w := worker.New(mb, nil)

	w.RegisterWorkflow(Workflow1)
	w.RegisterWorkflow(Workflow2)
	w.RegisterActivity(ActivityCancel)
	w.RegisterActivity(ActivitySkip)
	w.RegisterActivity(ActivitySuccess)
	w.RegisterActivity(ActivityCleanup)

	if err := w.Start(ctx); err != nil {
		panic("could not start worker")
	}
}

func Workflow1(ctx workflow.Context, msg string) (string, error) {
	samples.Trace(ctx, "Entering Workflow1", msg)
	defer samples.Trace(ctx, "Leaving Workflow1")

	defer func() {
		if errors.Is(ctx.Err(), workflow.Canceled) {
			samples.Trace(ctx, "Workflow1 was canceled")

			samples.Trace(ctx, "Do cleanup")
			ctx := workflow.NewDisconnectedContext(ctx)
			if _, err := workflow.ExecuteActivity[any](ctx, workflow.DefaultActivityOptions, ActivityCleanup).Get(ctx); err != nil {
				panic("could not execute cleanup activity")
			}
			samples.Trace(ctx, "Done with cleanup")
		}
	}()

	samples.Trace(ctx, "schedule ActivitySuccess")
	if r0, err := workflow.ExecuteActivity[int](ctx, workflow.DefaultActivityOptions, ActivitySuccess, 1, 2).Get(ctx); err != nil {
		samples.Trace(ctx, "error getting activity success result", err)
	} else {
		samples.Trace(ctx, "ActivitySuccess result:", r0)
	}

	samples.Trace(ctx, "schedule ActivityCancel")
	if rw, err := workflow.CreateSubWorkflowInstance[string](ctx, workflow.SubWorkflowOptions{
		InstanceID: "sub-workflow",
	}, Workflow2, "hello sub").Get(ctx); err != nil {
		samples.Trace(ctx, "error getting workflow2 result", err)
	} else {
		samples.Trace(ctx, "Workflow2 result:", rw)
	}

	samples.Trace(ctx, "schedule ActivitySkip")
	if r2, err := workflow.ExecuteActivity[int](ctx, workflow.DefaultActivityOptions, ActivitySkip, 1, 2).Get(ctx); err != nil {
		samples.Trace(ctx, "error getting activity skip result", err)
	} else {
		samples.Trace(ctx, "ActivitySkip result:", r2)
	}

	samples.Trace(ctx, "Workflow finished")
	return "result", nil
}

func Workflow2(ctx workflow.Context, msg string) (ret string, err error) {
	samples.Trace(ctx, "Entering Workflow2", msg)
	defer samples.Trace(ctx, "Leaving Workflow2")

	defer func() {
		if errors.Is(ctx.Err(), workflow.Canceled) {
			samples.Trace(ctx, "Workflow2 was canceled")

			samples.Trace(ctx, "Do cleanup")
			ctx := workflow.NewDisconnectedContext(ctx)
			if _, err := workflow.ExecuteActivity[any](ctx, workflow.DefaultActivityOptions, ActivityCleanup).Get(ctx); err != nil {
				panic("could not execute cleanup activity")
			}
			samples.Trace(ctx, "Done with cleanup")

			ret = "cleanup result"
		}
	}()

	samples.Trace(ctx, "schedule ActivityCancel")
	if r1, err := workflow.ExecuteActivity[int](ctx, workflow.DefaultActivityOptions, ActivityCancel, 1, 2).Get(ctx); err != nil {
		samples.Trace(ctx, "error getting activity cancel result", err)
	} else {
		samples.Trace(ctx, "ActivityCancel result:", r1)
	}

	return "some result", nil
}

func ActivitySuccess(ctx context.Context, a, b int) (int, error) {
	log.Println("Entering ActivitySuccess")
	defer log.Println("Leaving ActivitySuccess")

	return a + b, nil
}

func ActivityCancel(ctx context.Context, a, b int) (int, error) {
	log.Println("Entering ActivityCancel")
	defer log.Println("Leaving ActivityCancel")

	time.Sleep(10 * time.Second)

	return a + b, nil
}

func ActivitySkip(ctx context.Context, a, b int) (int, error) {
	log.Println("Entering ActivitySkip")
	defer log.Println("Leaving ActivitySkip")

	return a + b, nil
}

func ActivityCleanup(ctx context.Context) error {
	log.Println("Entering ActivityCleanup")
	defer log.Println("Leaving ActivityCleanup")

	log.Println("Do some cleanup")

	return nil
}
