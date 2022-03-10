package workflow

import (
	a "github.com/cschleiden/go-workflows/internal/args"
	"github.com/cschleiden/go-workflows/internal/command"
	"github.com/cschleiden/go-workflows/internal/converter"
	"github.com/cschleiden/go-workflows/internal/fn"
	"github.com/cschleiden/go-workflows/internal/sync"
	"github.com/cschleiden/go-workflows/internal/workflowstate"
	"github.com/pkg/errors"
)

type SubWorkflowOptions struct {
	InstanceID string

	RetryOptions RetryOptions
}

var DefaultSubWorkflowOptions = SubWorkflowOptions{
	RetryOptions: DefaultRetryOptions,
}

func CreateSubWorkflowInstance(ctx sync.Context, options SubWorkflowOptions, workflow Workflow, args ...interface{}) sync.Future {
	return WithRetries(ctx, options.RetryOptions, func(ctx sync.Context) sync.Future {
		return createSubWorkflowInstance(ctx, options, workflow, args...)
	})
}

func createSubWorkflowInstance(ctx sync.Context, options SubWorkflowOptions, workflow Workflow, args ...interface{}) sync.Future {
	f := sync.NewFuture()

	inputs, err := a.ArgsToInputs(converter.DefaultConverter, args...)
	if err != nil {
		f.Set(nil, errors.Wrap(err, "failed to convert workflow input"))
		return f
	}

	wfState := workflowstate.WorkflowState(ctx)

	scheduleEventID := wfState.GetNextScheduleEventID()

	name := fn.Name(workflow)
	cmd := command.NewScheduleSubWorkflowCommand(scheduleEventID, options.InstanceID, name, inputs)
	wfState.AddCommand(&cmd)
	wfState.TrackFuture(scheduleEventID, f)

	// Handle cancellation
	if d := ctx.Done(); d != nil {
		if c, ok := d.(sync.ChannelInternal); ok {
			c.ReceiveNonBlocking(ctx, func(_ interface{}) {
				// Workflow has been canceled, check if the sub-workflow has already been scheduled
				if cmd.State == command.CommandState_Committed {
					// Command has already been committed, that means the sub-workflow has already been scheduled. Wait
					// until it is done.
					return
				}

				wfState.RemoveFuture(scheduleEventID)
				wfState.RemoveCommand(cmd)
				f.Set(nil, sync.Canceled)
			})
		}
	}

	return f
}