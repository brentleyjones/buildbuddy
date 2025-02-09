package runner

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/buildbuddy-io/buildbuddy/enterprise/server/auth"
	"github.com/buildbuddy-io/buildbuddy/enterprise/server/remote_execution/commandutil"
	"github.com/buildbuddy-io/buildbuddy/enterprise/server/remote_execution/container"
	"github.com/buildbuddy-io/buildbuddy/enterprise/server/remote_execution/containers/bare"
	"github.com/buildbuddy-io/buildbuddy/enterprise/server/remote_execution/containers/containerd"
	"github.com/buildbuddy-io/buildbuddy/enterprise/server/remote_execution/containers/docker"
	"github.com/buildbuddy-io/buildbuddy/enterprise/server/remote_execution/platform"
	"github.com/buildbuddy-io/buildbuddy/enterprise/server/remote_execution/workspace"
	"github.com/buildbuddy-io/buildbuddy/enterprise/server/tasksize"
	"github.com/buildbuddy-io/buildbuddy/server/config"
	"github.com/buildbuddy-io/buildbuddy/server/environment"
	"github.com/buildbuddy-io/buildbuddy/server/interfaces"
	"github.com/buildbuddy-io/buildbuddy/server/metrics"
	"github.com/buildbuddy-io/buildbuddy/server/resources"
	"github.com/buildbuddy-io/buildbuddy/server/util/log"
	"github.com/buildbuddy-io/buildbuddy/server/util/perms"
	"github.com/buildbuddy-io/buildbuddy/server/util/status"
	"github.com/golang/protobuf/proto"
	"github.com/prometheus/client_golang/prometheus"

	aclpb "github.com/buildbuddy-io/buildbuddy/proto/acl"
	repb "github.com/buildbuddy-io/buildbuddy/proto/remote_execution"
	uidpb "github.com/buildbuddy-io/buildbuddy/proto/user_id"
	wkpb "github.com/buildbuddy-io/buildbuddy/proto/worker"
	dockerclient "github.com/docker/docker/client"
)

const (
	// Runner states

	// initial means the container struct has been created but no actual container
	// has been created yet.
	initial state = iota
	// ready means the container is created and ready to run commands.
	ready
	// paused means the container is frozen and is eligible for addition to the
	// container pool.
	paused
	// removed means the container has been removed and cannot execute any more
	// commands.
	removed

	// How long to spend waiting for a runner to be removed before giving up.
	runnerCleanupTimeout = 30 * time.Second
	// Allowed time to spend trying to pause a runner and add it to the pool.
	runnerRecycleTimeout = 15 * time.Second

	// How big a runner's workspace is allowed to get before we decide that it
	// can't be added to the pool and must be cleaned up instead.
	defaultRunnerDiskSizeLimitBytes = 16e9
	// Memory usage estimate multiplier for pooled runners, relative to the
	// default memory estimate for execution tasks.
	runnerMemUsageEstimateMultiplierBytes = 6.5
	// The maximum fraction of allocated RAM that can be allocated to pooled
	// runners.
	runnerAllocatedRAMFractionBytes = 0.8

	// Label assigned to runner pool request count metric for fulfilled requests.
	hitStatusLabel = "hit"
	// Label assigned to runner pool request count metric for unfulfilled requests.
	missStatusLabel = "miss"
)

var (
	// RunnerMaxMemoryExceeded is returned from Pool.Add if a runner cannot be
	// added to the pool because its current memory consumption exceeds the max
	// configured limit.
	RunnerMaxMemoryExceeded = status.ResourceExhaustedError("runner memory limit exceeded")
	// RunnerMaxDiskSizeExceeded is returned from Pool.Add if a runner cannot be
	// added to the pool because its current disk usage exceeds the max configured
	// limit.
	RunnerMaxDiskSizeExceeded = status.ResourceExhaustedError("runner disk size limit exceeded")

	podIDFromCpusetRegexp = regexp.MustCompile("/kubepods(/.*?)?/pod([a-z0-9\\-]{36})/")

	flagFilePattern           = regexp.MustCompile(`^(?:@|--?flagfile=)(.+)`)
	externalRepositoryPattern = regexp.MustCompile(`^@.*//.*`)
)

func k8sPodID() (string, error) {
	if _, err := os.Stat("/proc/1/cpuset"); err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	buf, err := os.ReadFile("/proc/1/cpuset")
	if err != nil {
		return "", err
	}
	cpuset := string(buf)
	if m := podIDFromCpusetRegexp.FindStringSubmatch(cpuset); m != nil {
		return m[2], nil
	}
	return "", nil
}

// State indicates the current state of a CommandContainer.
type state int

// CommandRunner represents a command container and attached workspace.
type CommandRunner struct {
	// ACL controls who can use this runner.
	ACL *aclpb.ACL
	// PlatformProperties holds the platform properties for the last
	// task executed by this runner.
	PlatformProperties *platform.Properties
	// WorkerKey is the peristent worker key. Only tasks with matching
	// worker key can execute on this runner.
	WorkerKey string
	// InstanceName is the remote instance name specified when creating this
	// runner. Only tasks with matching remote instance names can execute on this
	// runner.
	InstanceName string

	// Container is the handle on the container (possibly the bare /
	// NOP container) that is used to execute commands.
	Container container.CommandContainer
	// Workspace holds the data which is used by this runner.
	Workspace *workspace.Workspace

	// State is the current state of the runner as it pertains to reuse.
	state state

	// Stdin handle to send persistent WorkRequests to.
	stdinWriter io.Writer
	// Stdout handle to read persistent WorkResponses from.
	// N.B. This is a bufio.Reader to support ByteReader required by ReadUvarint.
	stdoutReader *bufio.Reader
	// Keeps track of whether or not we encountered any errors that make the runner non-reusable.
	doNotReuse bool

	// Cached resource usage values from the last time the runner was added to
	// the pool.

	memoryUsageBytes int64
	diskUsageBytes   int64
}

func (r *CommandRunner) PrepareForTask(task *repb.ExecutionTask) error {
	r.Workspace.SetTask(task)
	// Clean outputs for the current task if applicable, in case
	// those paths were written as read-only inputs in a previous action.
	if r.PlatformProperties.RecycleRunner {
		if err := r.Workspace.Clean(); err != nil {
			log.Errorf("Failed to clean workspace: %s", err)
			return err
		}
	}
	if err := r.Workspace.CreateOutputDirs(); err != nil {
		return status.UnavailableErrorf("Error creating output directory: %s", err.Error())
	}
	return nil
}

func (r *CommandRunner) Run(ctx context.Context, command *repb.Command) *interfaces.CommandResult {
	// Get the container to "ready" state so that we can exec commands in it.
	switch r.state {
	case initial:
		if err := r.Container.PullImageIfNecessary(ctx); err != nil {
			return commandutil.ErrorResult(err)
		}
		if err := r.Container.Create(ctx, r.Workspace.Path()); err != nil {
			return commandutil.ErrorResult(err)
		}
		r.state = ready
		break
	case paused:
		if err := r.Container.Unpause(ctx); err != nil {
			return commandutil.ErrorResult(err)
		}
		r.state = ready
		break
	default:
		return commandutil.ErrorResult(status.FailedPreconditionErrorf("unexpected runner state %d; this should never happen", r.state))
	}

	if r.supportsPersistentWorkers(ctx, command) {
		return r.sendPersistentWorkRequest(ctx, command)
	}

	return r.Container.Exec(ctx, command, nil, nil)
}

func (r *CommandRunner) Remove(ctx context.Context) error {
	errs := []error{}
	if err := r.Workspace.Remove(); err != nil {
		errs = append(errs, err)
	}
	if s := r.state; s != initial && s != removed {
		r.state = removed
		if err := r.Container.Remove(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return errSlice(errs)
	}
	return nil
}

func (r *CommandRunner) RemoveWithTimeout(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, runnerCleanupTimeout)
	defer cancel()
	return r.Remove(ctx)
}

func (r *CommandRunner) RemoveInBackground() {
	// TODO: Add to a cleanup queue instead of spawning a goroutine here.
	go func() {
		if err := r.RemoveWithTimeout(context.Background()); err != nil {
			log.Errorf("Failed to remove runner: %s", err)
		}
	}()
}

// ACLForUser returns an ACL that grants anyone in the given user's group to
// Read/Write permissions for a runner.
func ACLForUser(user interfaces.UserInfo) *aclpb.ACL {
	if user == nil {
		return nil
	}
	userID := &uidpb.UserId{Id: user.GetUserID()}
	groupID := user.GetGroupID()
	permBits := perms.OWNER_READ | perms.OWNER_WRITE | perms.GROUP_READ | perms.GROUP_WRITE
	return perms.ToACLProto(userID, groupID, permBits)
}

// Pool keeps track of command runners, both inactive (paused) and running.
//
// In the case of bare command execution, paused runners may not actually
// have their execution suspended. The pool doesn't currently account for CPU
// usage in this case.
type Pool struct {
	env              environment.Env
	podID            string
	buildRoot        string
	dockerClient     *dockerclient.Client
	containerdSocket string

	maxRunnerCount            int
	maxRunnerMemoryUsageBytes int64
	maxRunnerDiskUsageBytes   int64

	mu             sync.RWMutex // protects(isShuttingDown), protects(runners)
	isShuttingDown bool
	runners        []*CommandRunner
}

func NewPool(env environment.Env) (*Pool, error) {
	executorConfig := env.GetConfigurator().GetExecutorConfig()
	if executorConfig == nil {
		return nil, status.FailedPreconditionError("No executor config found")
	}

	podID, err := k8sPodID()
	if err != nil {
		return nil, status.FailedPreconditionErrorf("Failed to determine k8s pod ID: %s", err)
	}

	var dockerClient *dockerclient.Client
	containerdSocket := ""
	if executorConfig.ContainerdSocket != "" {
		_, err := os.Stat(executorConfig.ContainerdSocket)
		if os.IsNotExist(err) {
			return nil, status.FailedPreconditionErrorf("Containerd socket %q not found", executorConfig.ContainerdSocket)
		}
		containerdSocket = executorConfig.ContainerdSocket
		log.Info("Using containerd for execution")
		if executorConfig.DockerSocket != "" {
			log.Warning("containerd_socket and docker_socket both specified. Ignoring docker_socket in favor of containerd.")
		}
	} else if executorConfig.DockerSocket != "" {
		_, err := os.Stat(executorConfig.DockerSocket)
		if os.IsNotExist(err) {
			return nil, status.FailedPreconditionErrorf("Docker socket %q not found", executorConfig.DockerSocket)
		}
		dockerSocket := executorConfig.DockerSocket
		dockerClient, err = dockerclient.NewClientWithOpts(
			dockerclient.WithHost(fmt.Sprintf("unix://%s", dockerSocket)),
			dockerclient.WithAPIVersionNegotiation(),
		)
		if err != nil {
			return nil, status.FailedPreconditionErrorf("Failed to create docker client: %s", err)
		}
		log.Info("Using docker for execution")
	}

	p := &Pool{
		env:              env,
		podID:            podID,
		dockerClient:     dockerClient,
		containerdSocket: containerdSocket,
		buildRoot:        executorConfig.GetRootDirectory(),
		runners:          []*CommandRunner{},
	}
	p.setLimits(&executorConfig.RunnerPool)
	return p, nil
}

func (p *Pool) props() *platform.ExecutorProperties {
	return &platform.ExecutorProperties{
		ContainerType: p.containerType(),
	}
}

func (p *Pool) containerType() platform.ContainerType {
	if p.dockerClient != nil {
		return platform.DockerContainerType
	}
	if p.containerdSocket != "" {
		return platform.ContainerdContainerType
	}
	return platform.BareContainerType
}

// Add adds the given runner into the pool, evicting older runners if needed.
// If an error is returned, the runner was not successfully added to the pool.
func (p *Pool) Add(ctx context.Context, r *CommandRunner) error {
	if err := p.add(ctx, r); err != nil {
		metrics.RunnerPoolFailedRecycleAttempts.With(prometheus.Labels{
			metrics.RunnerPoolFailedRecycleReason: err.Label,
		}).Inc()
		return err.Error
	}
	return nil
}

func (p *Pool) add(ctx context.Context, r *CommandRunner) *labeledError {
	// TODO: once CommandContainer lifecycle methods are available, enforce that
	// the runner's CommandContainer is paused, and return a
	// FailedPreconditionError if not.

	if r.state != ready {
		return &labeledError{
			status.InternalErrorf("unexpected runner state %d; this should never happen", r.state),
			"unexpected_runner_state",
		}
	}
	if err := r.Container.Pause(ctx); err != nil {
		return &labeledError{
			status.WrapError(err, "failed to pause container before adding to the pool"),
			"pause_failed",
		}
	}
	r.state = paused

	stats, err := r.Container.Stats(ctx)
	if err != nil {
		return &labeledError{
			status.WrapError(err, "failed to compute container stats"),
			"stats_failed",
		}
	}
	if stats.MemoryUsageBytes > p.maxRunnerMemoryUsageBytes {
		return &labeledError{
			RunnerMaxMemoryExceeded,
			"max_memory_exceeded",
		}
	}
	du, err := r.Workspace.DiskUsageBytes()
	if err != nil {
		return &labeledError{
			status.WrapError(err, "failed to compute runner disk usage"),
			"compute_disk_usage_failed",
		}
	}
	if du > p.maxRunnerDiskUsageBytes {
		return &labeledError{
			RunnerMaxDiskSizeExceeded,
			"max_disk_usage_exceeded",
		}
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if p.isShuttingDown {
		return &labeledError{
			status.UnavailableError("pool is shutting down; cannot add new runners"),
			"pool_shutting_down",
		}
	}

	if len(p.runners) == p.maxRunnerCount {
		if len(p.runners) == 0 {
			return &labeledError{
				status.InternalError("pool max runner count is 0; this should never happen"),
				"max_runner_count_zero",
			}
		}
		// Evict the first and oldest runner to make room for the new one.
		r := p.runners[0]
		p.runners = p.runners[1:]

		metrics.RunnerPoolEvictions.Inc()
		metrics.RunnerPoolCount.Dec()
		metrics.RunnerPoolDiskUsageBytes.Sub(float64(r.diskUsageBytes))
		metrics.RunnerPoolMemoryUsageBytes.Sub(float64(r.memoryUsageBytes))

		r.RemoveInBackground()
	}

	p.runners = append(p.runners, r)

	// Cache these values so we don't need to recompute them when updating metrics
	// upon removal.
	r.memoryUsageBytes = stats.MemoryUsageBytes
	r.diskUsageBytes = du

	metrics.RunnerPoolDiskUsageBytes.Add(float64(r.diskUsageBytes))
	metrics.RunnerPoolMemoryUsageBytes.Add(float64(r.memoryUsageBytes))
	metrics.RunnerPoolCount.Inc()

	return nil
}

func (p *Pool) hostBuildRoot() string {
	if p.podID == "" {
		// Probably running on bare metal -- return the build root directly.
		return p.buildRoot
	}
	// Running on k8s -- return the path to the build root on the *host* node.
	// TODO(bduffany): Make this configurable in YAML, populating {{.PodID}} via template.
	// People might have conventions other than executor-data for the volume name + remotebuilds
	// for the build root dir.
	return fmt.Sprintf("/var/lib/kubelet/pods/%s/volumes/kubernetes.io~empty-dir/executor-data/remotebuilds", p.podID)
}

func (p *Pool) WarmupDefaultImage() {
	if p.dockerClient == nil {
		return
	}
	cfg := p.env.GetConfigurator().GetExecutorConfig()
	c := docker.NewDockerContainer(
		p.dockerClient, platform.DefaultContainerImage, p.hostBuildRoot(),
		&docker.DockerOptions{
			Socket:                  cfg.DockerSocket,
			EnableSiblingContainers: cfg.DockerSiblingContainers,
			UseHostNetwork:          cfg.DockerNetHost,
			DockerMountMode:         cfg.DockerMountMode,
		},
	)
	start := time.Now()
	// Give the pull up to 1 minute to succeed and 1 minute to create a warm up container.
	// In practice I saw clean pulls take about 30 seconds.
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	err := c.PullImageIfNecessary(ctx)
	if err != nil {
		log.Warningf("Warm up: could not pull default image %q: %s", platform.DefaultContainerImage, err)
		return
	}
	log.Infof("Warm up: pulled default image %q in %s", platform.DefaultContainerImage, time.Since(start))

	tmpDir, err := os.MkdirTemp("", "buildbuddy-warmup-*")
	if err != nil {
		log.Warningf("Warm up: could not create temp directory: %s", err)
		return
	}
	defer func() {
		_ = os.Remove(tmpDir)
	}()
	err = c.Create(ctx, tmpDir)
	if err != nil {
		log.Warningf("Warm up: could not create warm up container: %s", err)
		return
	}
	err = c.Remove(ctx)
	if err != nil {
		log.Warningf("Warm up: could not remove warm up container: %s", err)
		return
	}
	log.Infof("Warm up finished for default image in %s.", time.Since(start))
}

// Get returns a runner that can be used to execute the given task. The caller
// must call TryRecycle on the returned runner when done using it.
//
// If the task has runner recycling enabled then it attempts to find a runner
// from the pool that can execute the task. If runner recycling is disabled or
// if there are no eligible paused runners, it creates and returns a new runner.
//
// The returned runner is considered "active" and will be killed if the
// executor is shut down.
func (p *Pool) Get(ctx context.Context, task *repb.ExecutionTask) (*CommandRunner, error) {
	props, err := platform.ParseProperties(task.GetCommand().GetPlatform(), p.props())
	if err != nil {
		return nil, err
	}
	// TODO: This mutates the task; find a cleaner way to do this.
	platform.ApplyOverrides(p.env, props, task.GetCommand())

	user, err := auth.UserFromTrustedJWT(ctx)
	// PermissionDenied and Unimplemented both imply that this is an
	// anonymous execution, so ignore those.
	if err != nil && !status.IsPermissionDeniedError(err) && !status.IsUnimplementedError(err) {
		return nil, err
	}
	if props.RecycleRunner && err != nil {
		return nil, status.InvalidArgumentError(
			"runner recycling is not supported for anonymous builds " +
				`(recycling was requested via platform property "recycle-runner=true")`)
	}

	instanceName := task.GetExecuteRequest().GetInstanceName()

	workerKey := props.PersistentWorkerKey
	if props.PersistentWorker && workerKey == "" {
		workerArgs, _ := SplitArgsIntoWorkerArgsAndFlagFiles(task.GetCommand().GetArguments())
		workerKey = strings.Join(workerArgs, " ")
	}

	if props.RecycleRunner {
		r := p.take(&query{
			User:           user,
			ContainerImage: props.ContainerImage,
			WorkflowID:     props.WorkflowID,
			InstanceName:   instanceName,
			WorkerKey:      workerKey,
		})
		if r != nil {
			log.Info("Reusing workspace for task.")
			r.PlatformProperties = props
			return r, nil
		}
	}
	wsOpts := &workspace.Opts{Preserve: props.PreserveWorkspace}
	ws, err := workspace.New(p.env, p.buildRoot, wsOpts)
	if err != nil {
		return nil, err
	}
	ctr := p.newContainer(props)
	return &CommandRunner{
		ACL:                ACLForUser(user),
		PlatformProperties: props,
		InstanceName:       instanceName,
		WorkerKey:          workerKey,
		Container:          ctr,
		Workspace:          ws,
	}, nil
}

func (p *Pool) newContainer(props *platform.Properties) container.CommandContainer {
	switch p.containerType() {
	case platform.DockerContainerType:
		cfg := p.env.GetConfigurator().GetExecutorConfig()
		return docker.NewDockerContainer(
			p.dockerClient, props.ContainerImage, p.hostBuildRoot(),
			&docker.DockerOptions{
				Socket:                  cfg.DockerSocket,
				EnableSiblingContainers: cfg.DockerSiblingContainers,
				UseHostNetwork:          cfg.DockerNetHost,
				DockerMountMode:         cfg.DockerMountMode,
				ForceRoot:               props.DockerForceRoot,
			},
		)
	case platform.ContainerdContainerType:
		return containerd.NewContainerdContainer(p.containerdSocket, props.ContainerImage, p.hostBuildRoot())
	default:
		return bare.NewBareCommandContainer()
	}
}

// query specifies a set of search criteria for runners within a pool.
// All criteria must match in order for a runner to be matched.
type query struct {
	// User is the current authenticated user. This query will only match runners
	// that this user can access.
	// Required.
	User interfaces.UserInfo
	// ContainerImage is the image that must have been used to create the
	// container.
	// Required; the zero-value "" matches bare runners.
	ContainerImage string
	// WorkflowID is the BuildBuddy workflow ID, if applicable.
	// Required; the zero-value "" matches non-workflow runners.
	WorkflowID string
	// WorkerKey is the key used to tell if a persistent worker can be reused.
	// Required; the zero-value "" matches non-persistent-worker runners.
	WorkerKey string
	// InstanceName is the remote instance name that must have been used when
	// creating the runner.
	// Required; the zero-value "" corresponds to the default instance name.
	InstanceName string
}

// take takes any runner matching the given query out of the pool. If no
// matching runners are found, `nil` is returned.
func (p *Pool) take(q *query) *CommandRunner {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, r := range p.runners {
		if r.PlatformProperties.ContainerImage != q.ContainerImage ||
			r.PlatformProperties.WorkflowID != q.WorkflowID ||
			r.WorkerKey != q.WorkerKey ||
			r.InstanceName != q.InstanceName {
			continue
		}
		if authErr := perms.AuthorizeWrite(&q.User, r.ACL); authErr != nil {
			continue
		}

		p.runners = append(p.runners[:i], p.runners[i+1:]...)

		metrics.RunnerPoolCount.Dec()
		metrics.RunnerPoolDiskUsageBytes.Sub(float64(r.diskUsageBytes))
		metrics.RunnerPoolMemoryUsageBytes.Sub(float64(r.memoryUsageBytes))
		metrics.RecycleRunnerRequests.With(prometheus.Labels{
			metrics.RecycleRunnerRequestStatusLabel: hitStatusLabel,
		}).Inc()

		return r
	}

	metrics.RecycleRunnerRequests.With(prometheus.Labels{
		metrics.RecycleRunnerRequestStatusLabel: missStatusLabel,
	}).Inc()

	return nil
}

// Size returns the current number of paused runners in the pool.
func (p *Pool) Size() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.runners)
}

// Shutdown removes all runners from the pool and prevents new ones from
// being added.
func (p *Pool) Shutdown(ctx context.Context) error {
	p.mu.Lock()
	p.isShuttingDown = true
	pooledRunners := p.runners
	p.runners = nil
	p.mu.Unlock()

	errs := []error{}
	for _, r := range pooledRunners {
		if err := r.RemoveWithTimeout(ctx); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return status.InternalErrorf("failed to shut down runner pool: %s", errSlice(errs))
	}
	return nil
}

// TryRecycle either adds r back to the pool if appropriate, or removes it,
// freeing up any resources it holds.
func (p *Pool) TryRecycle(r *CommandRunner, finishedCleanly bool) {
	ctx, cancel := context.WithTimeout(context.Background(), runnerRecycleTimeout)
	defer cancel()

	recycled := false
	defer func() {
		if !recycled {
			r.RemoveInBackground()
		}
	}()

	if !r.PlatformProperties.RecycleRunner || !finishedCleanly || r.doNotReuse {
		return
	}
	// Clean the workspace once before adding it to the pool.
	if err := r.Workspace.Clean(); err != nil {
		log.Errorf("Failed to clean workspace: %s", err)
		return
	}
	// This call happens after we send the final stream event back to the
	// client, so background context is appropriate.
	if err := p.Add(ctx, r); err != nil {
		if status.IsResourceExhaustedError(err) || status.IsUnavailableError(err) {
			log.Debug(err.Error())
		} else {
			// If not a resource limit exceeded error, probably it was an error
			// removing the directory contents or a docker daemon error.
			log.Errorf("Failed to recycle runner: %s", err)
		}
		return
	}

	recycled = true
}

func (p *Pool) setLimits(cfg *config.RunnerPoolConfig) {
	totalRAMBytes := int64(float64(resources.GetAllocatedRAMBytes()) * runnerAllocatedRAMFractionBytes)
	estimatedRAMBytes := int64(float64(tasksize.DefaultMemEstimate) * runnerMemUsageEstimateMultiplierBytes)

	count := cfg.MaxRunnerCount
	if count == 0 {
		// Don't allow more paused runners than the max number of tasks that can be
		// executing at once, if they were all using the default memory estimate.
		if estimatedRAMBytes > 0 {
			count = int(float64(totalRAMBytes) / float64(estimatedRAMBytes))
		}
	} else if count < 0 {
		// < 0 means no limit.
		count = int(math.MaxInt32)
	}

	mem := cfg.MaxRunnerMemoryUsageBytes
	if mem == 0 {
		mem = int64(float64(totalRAMBytes) / float64(count))
	} else if mem < 0 {
		// < 0 means no limit.
		mem = math.MaxInt64
	}

	disk := cfg.MaxRunnerDiskSizeBytes
	if disk == 0 {
		disk = defaultRunnerDiskSizeLimitBytes
	} else if disk < 0 {
		// < 0 means no limit.
		disk = math.MaxInt64
	}

	p.maxRunnerCount = count
	p.maxRunnerMemoryUsageBytes = mem
	p.maxRunnerDiskUsageBytes = disk
}

type labeledError struct {
	// Error is the wrapped error.
	Error error
	// Label is a short label for Prometheus.
	Label string
}

type errSlice []error

func (es errSlice) Error() string {
	if len(es) == 1 {
		return es[0].Error()
	}
	msgs := []string{}
	for _, err := range es {
		msgs = append(msgs, err.Error())
	}
	return fmt.Sprintf("[multiple errors: %s]", strings.Join(msgs, "; "))
}

func SplitArgsIntoWorkerArgsAndFlagFiles(args []string) ([]string, []string) {
	workerArgs := make([]string, 0)
	flagFiles := make([]string, 0)
	for _, arg := range args {
		if flagFilePattern.MatchString(arg) {
			flagFiles = append(flagFiles, arg)
		} else {
			workerArgs = append(workerArgs, arg)
		}
	}
	return workerArgs, flagFiles
}

func (r *CommandRunner) supportsPersistentWorkers(ctx context.Context, command *repb.Command) bool {
	if r.PlatformProperties.PersistentWorkerKey != "" {
		return true
	}

	if !r.PlatformProperties.PersistentWorker {
		return false
	}

	_, flagFiles := SplitArgsIntoWorkerArgsAndFlagFiles(command.GetArguments())
	return len(flagFiles) > 0
}

func (r *CommandRunner) sendPersistentWorkRequest(ctx context.Context, command *repb.Command) *interfaces.CommandResult {
	result := &interfaces.CommandResult{
		CommandDebugString: fmt.Sprintf("(persistentworker) %s", command.GetArguments()),
		ExitCode:           commandutil.NoExitCode,
	}

	workerArgs, flagFiles := SplitArgsIntoWorkerArgsAndFlagFiles(command.GetArguments())

	// If it's our first rodeo, create the persistent worker.
	if r.stdinWriter == nil || r.stdoutReader == nil {
		stdinReader, stdinWriter := io.Pipe()
		stdoutReader, stdoutWriter := io.Pipe()
		r.stdinWriter = stdinWriter
		r.stdoutReader = bufio.NewReader(stdoutReader)

		command.Arguments = append(workerArgs, "--persistent_worker")

		go func() {
			res := r.Container.Exec(ctx, command, stdinReader, stdoutWriter)
			stdinWriter.Close()
			stdoutReader.Close()
			log.Debugf("Persistent worker exited with response: %+v, flagFiles: %+v, workerArgs: %+v", res, flagFiles, workerArgs)
			r.doNotReuse = true
		}()
	}

	// We've got a worker - now let's build a work request.
	requestProto := &wkpb.WorkRequest{
		Inputs: make([]*wkpb.Input, 0, len(r.Workspace.Inputs)),
	}

	expandedArguments, err := r.expandArguments(flagFiles)
	if err != nil {
		result.Error = status.WrapError(err, "expanding arguments")
		return result
	}
	requestProto.Arguments = expandedArguments

	// Collect all of the input digests
	for path, digest := range r.Workspace.Inputs {
		digestBuffer := proto.NewBuffer( /* buf */ nil)
		err := digestBuffer.Marshal(digest)
		if err != nil {
			result.Error = status.WrapError(err, "marshalling input digest")
			return result
		}
		requestProto.Inputs = append(requestProto.Inputs, &wkpb.Input{
			Digest: digestBuffer.Bytes(),
			Path:   path,
		})
	}

	// Encode the work requests
	buf := proto.NewBuffer( /* buf */ nil)
	if err := buf.EncodeMessage(requestProto); err != nil {
		result.Error = status.WrapError(err, "request marshalling failed")
		return result
	}

	// Send it to our worker over stdin.
	r.stdinWriter.Write(buf.Bytes())

	// Now we've sent a work request, let's collect our response.
	responseProto := &wkpb.WorkResponse{}

	// Read the response size from stdout as a unsigned varint.
	size, err := binary.ReadUvarint(r.stdoutReader)
	if err != nil {
		result.Error = status.WrapError(err, "reading response length")
		return result
	}
	data := make([]byte, size)

	// Read the response proto from stdout.
	if _, err := io.ReadFull(r.stdoutReader, data); err != nil {
		result.Error = status.WrapError(err, "reading response proto")
		return result
	}

	// Unmarshal the response proto.
	if err := proto.Unmarshal(data, responseProto); err != nil {
		result.Error = status.WrapError(err, "unmarshaling response proto")
		return result
	}

	// Populate the result from the response proto.
	result.Stderr = []byte(responseProto.Output)
	result.ExitCode = int(responseProto.ExitCode)
	return result
}

// Recursively expands arguments by replacing @filename args with the contents of the referenced
// files. The @ itself can be escaped with @@. This deliberately does not expand --flagfile= style
// arguments, because we want to get rid of the expansion entirely at some point in time.
// Based on: https://github.com/bazelbuild/bazel/blob/e9e6978809b0214e336fee05047d5befe4f4e0c3/src/main/java/com/google/devtools/build/lib/worker/WorkerSpawnRunner.java#L324
func (r *CommandRunner) expandArguments(args []string) ([]string, error) {
	expandedArgs := make([]string, 0)
	for _, arg := range args {
		if strings.HasPrefix(arg, "@") && !strings.HasPrefix(arg, "@@") && !externalRepositoryPattern.MatchString(arg) {
			file, err := os.Open(filepath.Join(r.Workspace.Path(), arg[1:]))
			if err != nil {
				return nil, err
			}
			defer file.Close()
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				args, err := r.expandArguments([]string{scanner.Text()})
				if err != nil {
					return nil, err
				}
				expandedArgs = append(expandedArgs, args...)
			}
			if err := scanner.Err(); err != nil {
				return nil, err
			}
		} else {
			expandedArgs = append(expandedArgs, arg)
		}
	}

	return expandedArgs, nil
}
