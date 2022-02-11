package main

import (
	"context"
	"log"
	"time"

	"github.com/cschleiden/go-dt/pkg/workflow"
	"github.com/cschleiden/go-dt/samples"
)

type Inputs struct {
	Msg   string
	Times int
}

func Workflow1(ctx workflow.Context, msg string, times int, inputs Inputs) error {
	samples.Trace(ctx, "Entering Workflow1", msg, times, inputs)

	defer samples.Trace(ctx, "Leaving Workflow1")

	var r1 int
	err := workflow.ExecuteActivity(ctx, workflow.DefaultActivityOptions, Activity1, 35, 12, nil, "test").Get(ctx, &r1)
	if err != nil {
		panic("error getting activity 1 result")
	}
	samples.Trace(ctx, "R1 result:", r1)

	var r2 int
	err = workflow.ExecuteActivity(ctx, workflow.DefaultActivityOptions, Activity2).Get(ctx, &r2)
	if err != nil {
		panic("error getting activity 1 result")
	}
	samples.Trace(ctx, "R2 result:", r2)

	return nil
}

func Activity1(ctx context.Context, a, b int, x, y *string) (int, error) {
	log.Println("Entering Activity1")
	defer log.Println("Leaving Activity1")

	log.Println(x, *y)

	time.Sleep(5 * time.Second)

	return a + b, nil
}

func Activity2(ctx context.Context) (int, error) {
	log.Println("Entering Activity2")
	defer log.Println("Leaving Activity2")

	time.Sleep(1 * time.Second)

	return 12, nil
}