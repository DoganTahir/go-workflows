package main

import (
	"testing"

	"github.com/cschleiden/go-workflows/internal/tester"
	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func Test_Workflow(t *testing.T) {
	tester := tester.NewWorkflowTester(Workflow1)

	tester.OnActivity(Activity1, mock.Anything, 35, 12).Return(47, nil)

	tester.Execute("Hello world" + uuid.NewString())

	require.True(t, tester.WorkflowFinished())

	var wr string
	var werr string
	tester.WorkflowResult(&wr, &werr)
	require.Equal(t, "result", wr)
	require.Empty(t, werr)
	tester.AssertExpectations(t)
}
