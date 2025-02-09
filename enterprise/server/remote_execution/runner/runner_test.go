package runner_test

import (
	"context"
	"io/ioutil"
	"os"
	"testing"

	"github.com/buildbuddy-io/buildbuddy/enterprise/server/remote_execution/platform"
	"github.com/buildbuddy-io/buildbuddy/enterprise/server/remote_execution/runner"
	"github.com/buildbuddy-io/buildbuddy/enterprise/server/remote_execution/workspace"
	"github.com/buildbuddy-io/buildbuddy/enterprise/server/tasksize"
	"github.com/buildbuddy-io/buildbuddy/server/config"
	"github.com/buildbuddy-io/buildbuddy/server/testutil/testauth"
	"github.com/buildbuddy-io/buildbuddy/server/testutil/testenv"
	"github.com/buildbuddy-io/buildbuddy/server/util/status"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	repb "github.com/buildbuddy-io/buildbuddy/proto/remote_execution"
)

const (
	unlimited = -1

	sysMemoryBytes = tasksize.DefaultMemEstimate * 10
	sysMilliCPU    = tasksize.DefaultCPUEstimate * 10
)

var (
	defaultCfg = &config.RunnerPoolConfig{}

	noLimitsCfg = &config.RunnerPoolConfig{
		MaxRunnerCount:            unlimited,
		MaxRunnerDiskSizeBytes:    unlimited,
		MaxRunnerMemoryUsageBytes: unlimited,
	}
)

func newTask() *repb.ExecutionTask {
	return &repb.ExecutionTask{
		Command: &repb.Command{
			Platform: &repb.Platform{
				Properties: []*repb.Platform_Property{
					{Name: platform.RecycleRunnerPropertyName, Value: "true"},
				},
			},
		},
	}
}

func newWorkspace(t *testing.T, env *testenv.TestEnv) *workspace.Workspace {
	tmpDir, err := ioutil.TempDir("/tmp", "buildbuddy_test_runner_workspace_*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(tmpDir); err != nil {
			t.Fatal(err)
		}
	})
	ws, err := workspace.New(env, tmpDir, &workspace.Opts{})
	if err != nil {
		t.Fatal(err)
	}
	return ws
}

func newUUID(t *testing.T) string {
	id, err := uuid.NewRandom()
	if err != nil {
		t.Fatal(err)
	}
	return id.String()
}

func newTestEnv(t *testing.T) *testenv.TestEnv {
	env := testenv.GetTestEnv(t)
	env.SetAuthenticator(testauth.NewTestAuthenticator(testauth.TestUsers(
		"US1", "GR1",
		"US2", "GR2",
	)))
	return env
}

func withAuthenticatedUser(t *testing.T, ctx context.Context, userID string) context.Context {
	jwt, err := testauth.TestJWTForUserID(userID)
	require.NoError(t, err)
	return context.WithValue(ctx, "x-buildbuddy-jwt", jwt)
}

func runMinimalCommand(t *testing.T, r *runner.CommandRunner) {
	res := r.Run(context.Background(), &repb.Command{Arguments: []string{"pwd"}})
	if res.Error != nil {
		t.Fatal(res.Error)
	}
}

func runShellCommand(t *testing.T, r *runner.CommandRunner, command string) {
	res := r.Run(context.Background(), &repb.Command{Arguments: []string{"sh", "-c", command}})
	if res.Error != nil {
		t.Fatal(res.Error)
	}
}

func newRunnerPool(t *testing.T, env *testenv.TestEnv, cfg *config.RunnerPoolConfig) *runner.Pool {
	env.GetConfigurator().GetExecutorConfig().RunnerPool = *cfg
	p, err := runner.NewPool(env)
	require.NoError(t, err)
	require.NotNil(t, p)
	return p
}

func TestRunnerPool_CanAddAndGetBackSameRunner(t *testing.T) {
	env := newTestEnv(t)
	pool := newRunnerPool(t, env, noLimitsCfg)
	ctx := withAuthenticatedUser(t, context.Background(), "US1")

	r, err := pool.Get(ctx, newTask())

	require.NoError(t, err)

	runMinimalCommand(t, r)

	err = pool.Add(context.Background(), r)

	require.NoError(t, err)
	assert.Equal(t, 1, pool.Size())

	taken, err := pool.Get(ctx, newTask())

	require.NoError(t, err)
	assert.Same(t, r, taken)
	assert.Equal(t, 0, pool.Size())
}

func TestRunnerPool_CannotTakeRunnerFromOtherGroup(t *testing.T) {
	env := newTestEnv(t)
	pool := newRunnerPool(t, env, noLimitsCfg)
	ctxUser1 := withAuthenticatedUser(t, context.Background(), "US1")
	ctxUser2 := withAuthenticatedUser(t, context.Background(), "US2")

	r, err := pool.Get(ctxUser1, newTask())

	require.NoError(t, err)

	runMinimalCommand(t, r)

	err = pool.Add(context.Background(), r)

	require.NoError(t, err)

	taken, err := pool.Get(ctxUser2, newTask())

	require.NoError(t, err)
	assert.NotSame(t, r, taken)
	assert.Equal(t, 1, pool.Size())
}

func TestRunnerPool_CannotTakeRunnerFromOtherInstanceName(t *testing.T) {
	env := newTestEnv(t)
	ctx := withAuthenticatedUser(t, context.Background(), "US1")
	pool := newRunnerPool(t, env, noLimitsCfg)
	task1 := newTask()
	task1.ExecuteRequest = &repb.ExecuteRequest{InstanceName: "instance/1"}
	task2 := newTask()
	task2.ExecuteRequest = &repb.ExecuteRequest{InstanceName: "instance/2"}

	r, err := pool.Get(ctx, task1)

	require.NoError(t, err)

	runMinimalCommand(t, r)

	err = pool.Add(context.Background(), r)

	require.NoError(t, err)
	assert.Equal(t, 1, pool.Size())

	taken, err := pool.Get(ctx, task2)

	assert.NotSame(t, r, taken)
	assert.Equal(t, 1, pool.Size())
}

func TestRunnerPool_CannotTakeRunnerFromOtherWorkflow(t *testing.T) {
	env := newTestEnv(t)
	ctx := withAuthenticatedUser(t, context.Background(), "US1")
	pool := newRunnerPool(t, env, noLimitsCfg)
	task1 := newTask()
	task1.Command.Platform.Properties = append(
		task1.Command.Platform.Properties,
		&repb.Platform_Property{Name: "workflow-id", Value: "WF1"},
	)
	task2 := newTask()
	task2.Command.Platform.Properties = append(
		task1.Command.Platform.Properties,
		&repb.Platform_Property{Name: "workflow-id", Value: "WF2"},
	)
	r, err := pool.Get(ctx, task1)

	require.NoError(t, err)

	runMinimalCommand(t, r)

	err = pool.Add(context.Background(), r)

	require.NoError(t, err)
	assert.Equal(t, 1, pool.Size())

	taken, err := pool.Get(ctx, task2)

	assert.NotSame(t, r, taken)
	assert.Equal(t, 1, pool.Size())
}

func TestRunnerPool_Shutdown_RemovesAllRunners(t *testing.T) {
	env := newTestEnv(t)
	pool := newRunnerPool(t, env, noLimitsCfg)
	ctx := withAuthenticatedUser(t, context.Background(), "US1")

	r, err := pool.Get(ctx, newTask())

	require.NoError(t, err)

	runMinimalCommand(t, r)

	err = pool.Add(context.Background(), r)

	require.NoError(t, err)
	assert.Equal(t, 1, pool.Size())

	err = pool.Shutdown(context.Background())

	require.NoError(t, err)
	assert.Equal(t, 0, pool.Size())
}

func TestRunnerPool_DefaultSystemBasedLimits_CanAddAtLeastOneRunner(t *testing.T) {
	env := newTestEnv(t)
	pool := newRunnerPool(t, env, defaultCfg)
	ctx := withAuthenticatedUser(t, context.Background(), "US1")

	r, err := pool.Get(ctx, newTask())

	require.NoError(t, err)

	runMinimalCommand(t, r)

	err = pool.Add(context.Background(), r)

	require.NoError(t, err)
	assert.Equal(t, 1, pool.Size())
}

func TestRunnerPool_ExceedMaxRunnerCount_OldestRunnerEvicted(t *testing.T) {
	env := newTestEnv(t)
	pool := newRunnerPool(t, env, &config.RunnerPoolConfig{
		MaxRunnerCount:            1,
		MaxRunnerDiskSizeBytes:    unlimited,
		MaxRunnerMemoryUsageBytes: unlimited,
	})
	ctx := withAuthenticatedUser(t, context.Background(), "US1")

	get := func() *runner.CommandRunner {
		r, err := pool.Get(ctx, newTask())
		require.NoError(t, err)
		runMinimalCommand(t, r)
		return r
	}

	r1 := get()
	r2 := get()

	err := pool.Add(ctx, r1)

	require.NoError(t, err)
	require.Equal(t, 1, pool.Size())

	err = pool.Add(ctx, r2)

	require.NoError(t, err)
	require.Equal(t, 1, pool.Size(), "should have evicted an older runner and added another one, leaving the size at 1")

	taken, err := pool.Get(ctx, newTask())

	require.NoError(t, err)
	assert.Same(t, r2, taken, "since runner limit is 1, last added runner should be the only runner in the pool")
	assert.Equal(t, 0, pool.Size())
}

func TestRunnerPool_DiskLimitExceeded_CannotAdd(t *testing.T) {
	env := newTestEnv(t)
	pool := newRunnerPool(t, env, &config.RunnerPoolConfig{
		MaxRunnerCount:            unlimited,
		MaxRunnerMemoryUsageBytes: unlimited,
		// At least one byte should be needed for the workspace root dir.
		MaxRunnerDiskSizeBytes: 1,
	})
	ctx := withAuthenticatedUser(t, context.Background(), "US1")

	r, err := pool.Get(ctx, newTask())

	runMinimalCommand(t, r)

	err = pool.Add(context.Background(), r)

	assert.True(t, status.IsResourceExhaustedError(err), "should exceed disk limit")
	assert.Equal(t, 0, pool.Size())
}

// TODO: Test mem limit. We currently don't compute mem usage for bare runners,
// so there's not a great way to test this yet.
