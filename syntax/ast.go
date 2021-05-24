package syntax

// Event interface represents workflow events in 'on' section
type Event interface {
	Name() string
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#onevent_nametypes
type WebhookEvent struct {
	Hook  string
	Types []string
}

func (e *WebhookEvent) Name() string {
	return e.Hook
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#onpushpull_requestbranchestags
type PushEvent struct {
	Branches []string
	Tags     []string
}

func (e *PushEvent) Name() string {
	return "push"
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#onpushpull_requestbranchestags
type PullRequestEvent struct {
	Branches []string
	Tags     []string
}

func (e *PullRequestEvent) Name() string {
	return "pull_request"
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#onschedule
type ScheduledEvent struct {
	Cron string
}

func (e *ScheduledEvent) Name() string {
	return "schedule"
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#permissions
type Permission uint8

const (
	PermissionNone = iota
	PermissionRead
	PermissionWrite
)

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#permissions
type Permissions struct {
	// All represents read-all or write-all, which define permissions of all scope at once.
	All Permission
	// Scopes is mappings from permission name to permission value
	Scopes map[string]Permission
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#defaultsrun
type DefaultsRun struct {
	Shell            string
	WorkingDirectory string
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#defaults
type Defaults struct {
	Run *DefaultsRun
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#concurrency
type Concurrency struct {
	Group            string
	CancelInProgress bool
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idenvironment
type Environment struct {
	Name string
	// URL is the URL mapped to 'environment_url' in the deployments API. Empty value means no value was specified.
	URL string
}

type ExecKind uint8

const (
	ExecKindAction = iota
	ExecKindRun
)

// Exec is an interface how the step is executed. Step in workflow runs either an action or a script
type Exec interface {
	Kind() ExecKind
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepsrun
type ExecRun struct {
	Script string
	// Shell represents optional 'shell' field. Empty means nothing specified
	Shell string
	// WorkingDirectory represents optional 'working-directory' field. Empty means nothing specified
	WorkingDirectory string
}

func (r *ExecRun) Kind() ExecKind {
	return ExecKindRun
}

type ExecAction struct {
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepsuses
	Uses string
	// Entrypoint represents optional 'entrypoint' field in 'with' section. Empty field means nothing specified
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepswithentrypoint
	Entrypoint string
	// Args represents optional 'args' field in 'with' section. Empty field means nothing specified
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepswithargs
	Args string
}

func (r *ExecAction) Kind() ExecKind {
	return ExecKindAction
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstrategymatrix
type Matrix struct {
	Values map[string]string
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#example-including-additional-values-into-combinations
	Include []map[string]string
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#example-excluding-configurations-from-a-matrix
	Exclude []map[string]string
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstrategy
type Strategy struct {
	Matrix *Matrix
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstrategyfail-fast
	FailFast bool
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstrategymax-parallel
	MaxParallel int
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idsteps
type Step struct {
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepsid
	ID string
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepsif
	If string
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepsname
	Name string
	Exec Exec
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepsenv
	Env map[string]string
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepscontinue-on-error
	ContinuesOnError bool
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepstimeout-minutes
	TimeoutMinutes float64
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idcontainercredentials
type Credentials struct {
	Username string
	Password string
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idcontainer
type Container struct {
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idcontainerimage
	Image       string
	Credentials *Credentials
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idcontainerenv
	Env map[string]string
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idcontainerports
	Ports []string
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idcontainervolumes
	Volumes []string
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idcontaineroptions
	Options string
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idservices
type Service struct {
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idservicesservice_idimage
	Image string
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idservicesservice_idcredentials
	Credentials *Credentials
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idservicesservice_idenv
	Env map[string]string
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idservicesservice_idports
	Ports []string
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idservicesservice_idvolumes
	Volumes []string
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idservicesservice_idoptions
	Options string
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobs
type Job struct {
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idname
	Name string
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idneeds
	Needs []string
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idruns-on
	RunsOn      string
	Permissions *Permissions
	Environment *Environment
	Concurrency *Concurrency
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idoutputs
	Outputs map[string]string
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idenv
	Env      map[string]string
	Defaults *Defaults
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idif
	If string
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idsteps
	Steps []*Step
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idtimeout-minutes
	TimeoutMinutes float64
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstrategy
	Strategy *Strategy
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idcontinue-on-error
	ContinueOnError bool
	Container       *Container
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idservices
	Services map[string]*Service
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions
type Workflow struct {
	Name string
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#onpushpull_requestbranchestags
	On          []Event
	Permissions *Permissions
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#env
	Env         map[string]string
	Defaults    *Defaults
	Concurrency *Concurrency
	Jobs        map[string]*Job
}
