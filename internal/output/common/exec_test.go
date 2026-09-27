package common_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/berquerant/git-iter-go/internal/output/common"
	"github.com/berquerant/git-iter-go/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExecuteTasksCore(t *testing.T) {
	t.Parallel()

	t.Run("basic execution", func(t *testing.T) {
		t.Parallel()
		tasks := []common.Task{
			{
				RepoAbsPath: "/repos/a",
				ReposRoot:   "/repos",
				Dir:         "/repos/a",
				Command:     []string{"echo", "hello"},
				Runner:      &testutil.FakeRunner{Output: "hello\n"},
			},
		}

		var collected []common.Result
		err := common.ExecuteTasksCore(context.Background(), tasks, 1, 0, false, func(r common.Result) error {
			collected = append(collected, r)
			return nil
		})

		require.NoError(t, err)
		require.Len(t, collected, 1)
		assert.Equal(t, "/repos/a", collected[0].RepoAbsPath)
		assert.Equal(t, "a", collected[0].RepoRelPath)
		assert.Equal(t, 0, collected[0].ExitCode)
		assert.Equal(t, "hello\n", collected[0].Stdout)
	})

	t.Run("fail fast aborts remaining", func(t *testing.T) {
		t.Parallel()
		tasks := []common.Task{
			{
				RepoAbsPath: "/repos/a",
				Dir:         "/repos/a",
				Command:     []string{"cmd"},
				Runner:      &testutil.FakeRunner{Err: errors.New("fail")},
			},
			{
				RepoAbsPath: "/repos/b",
				Dir:         "/repos/b",
				Command:     []string{"cmd"},
				Runner:      &testutil.FakeRunner{Err: errors.New("fail")},
			},
		}

		err := common.ExecuteTasksCore(context.Background(), tasks, 1, 0, true, func(_ common.Result) error {
			return nil
		})
		require.Error(t, err)
	})

	t.Run("callback error is captured", func(t *testing.T) {
		t.Parallel()
		tasks := []common.Task{
			{
				RepoAbsPath: "/repos/a",
				Dir:         "/repos/a",
				Command:     []string{"cmd"},
				Runner:      &testutil.FakeRunner{},
			},
		}

		callbackErr := errors.New("write error")
		err := common.ExecuteTasksCore(context.Background(), tasks, 1, 0, true, func(_ common.Result) error {
			return callbackErr
		})
		require.ErrorIs(t, err, callbackErr)
	})

	t.Run("timeout applies to task", func(t *testing.T) {
		t.Parallel()
		tasks := []common.Task{
			{
				RepoAbsPath: "/repos/a",
				Dir:         "/repos/a",
				Command:     []string{"cmd"},
				Runner:      &testutil.FakeRunner{},
			},
		}

		err := common.ExecuteTasksCore(context.Background(), tasks, 1, time.Second, false, func(_ common.Result) error {
			return nil
		})
		require.NoError(t, err)
	})
}
