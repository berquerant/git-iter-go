package runner_test

import (
	"errors"
	"testing"

	"github.com/berquerant/git-iter-go/internal/runner"
	"github.com/stretchr/testify/assert"
)

func TestTaskError(t *testing.T) {
	t.Parallel()

	baseErr := errors.New("exit status 2")
	tests := []struct {
		name          string
		taskErr       *runner.TaskError
		wantContains  []string
		wantUnwrapErr error
	}{
		{
			name: "full task error with exit code and base error",
			taskErr: &runner.TaskError{
				Dir:      "/repos/myorg/repo",
				Command:  []string{"git", "pull", "--rebase"},
				ExitCode: 2,
				Err:      baseErr,
			},
			wantContains: []string{
				"/repos/myorg/repo",
				"git pull --rebase",
				"exit code 2",
				"exit status 2",
			},
			wantUnwrapErr: baseErr,
		},
		{
			name: "task error without base error",
			taskErr: &runner.TaskError{
				Dir:      "/repos/myorg/repo2",
				Command:  []string{"make"},
				ExitCode: 1,
			},
			wantContains: []string{
				"/repos/myorg/repo2",
				"make",
				"exit code 1",
			},
			wantUnwrapErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			errMsg := tt.taskErr.Error()
			for _, exp := range tt.wantContains {
				assert.Contains(t, errMsg, exp)
			}
			if tt.wantUnwrapErr != nil {
				assert.True(t, errors.Is(tt.taskErr, tt.wantUnwrapErr))
			} else {
				assert.Nil(t, tt.taskErr.Unwrap())
			}
		})
	}
}
