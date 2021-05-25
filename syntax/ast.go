package syntax

type Pos struct {
	Line int
	Col  int
}

type String struct {
	Value string
	Pos   *Pos
}

type Bool struct {
	Value bool
	Pos   *Pos
}

type Int struct {
	Value int
	Pos   *Pos
}

type Float struct {
	Value float64
	Pos   *Pos
}

// Event interface represents workflow events in 'on' section
type Event interface {
	Name() string
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#onevent_nametypes
type WebhookEvent struct {
	Hook  *String
	Types []*String
	Pos   *Pos
}

func (e *WebhookEvent) Name() string {
	return e.Hook.Value
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#onpushpull_requestbranchestags
type PushEvent struct {
	Branches []*String
	Tags     []*String
	Pos      *Pos
}

func (e *PushEvent) Name() string {
	return "push"
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#onpushpull_requestbranchestags
type PullRequestEvent struct {
	Branches []*String
	Tags     []*String
	Pos      *Pos
}

func (e *PullRequestEvent) Name() string {
	return "pull_request"
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#onschedule
type ScheduledEvent struct {
	Cron *String
	Pos  *Pos
}

func (e *ScheduledEvent) Name() string {
	return "schedule"
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#permissions
type PermKind uint8

const (
	PermKindNone = iota
	PermKindRead
	PermKindWrite
)

type Permission struct {
	Kind PermKind
	Pos  *Pos
	Name *String
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#permissions
type Permissions struct {
	// All represents read-all or write-all, which define permissions of all scope at once.
	All *Permission
	// Scopes is mappings from permission name to permission value
	Scopes map[string]*Permission
	Pos    *Pos
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#defaultsrun
type DefaultsRun struct {
	Shell            *String
	WorkingDirectory *String
	Pos              *Pos
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#defaults
type Defaults struct {
	Run *DefaultsRun
	Pos *Pos
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#concurrency
type Concurrency struct {
	Group            *String
	CancelInProgress *Bool
	Pos              *Pos
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idenvironment
type Environment struct {
	Name *String
	// URL is the URL mapped to 'environment_url' in the deployments API. Empty value means no value was specified.
	URL *String
	Pos *Pos
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
	Run *String
	// Shell represents optional 'shell' field. Nil means nothing specified
	Shell *String
	// WorkingDirectory represents optional 'working-directory' field. Nil means nothing specified
	WorkingDirectory *String
}

func (r *ExecRun) Kind() ExecKind {
	return ExecKindRun
}

type ExecAction struct {
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepsuses
	Uses *String
	// Entrypoint represents optional 'entrypoint' field in 'with' section. Nil field means nothing specified
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepswithentrypoint
	Entrypoint *String
	// Args represents optional 'args' field in 'with' section. Nil field means nothing specified
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepswithargs
	Args *String
}

func (r *ExecAction) Kind() ExecKind {
	return ExecKindAction
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstrategymatrix
type MatrixElement struct {
	Key   *String
	Value *String
}
type Matrix struct {
	// Values stores mappings from name to values
	Values map[string][]*MatrixElement
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#example-including-additional-values-into-combinations
	Include []map[string]*MatrixElement
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#example-excluding-configurations-from-a-matrix
	Exclude []map[string]*MatrixElement
	Pos     *Pos
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstrategy
type Strategy struct {
	Matrix *Matrix
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstrategyfail-fast
	FailFast *Bool
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstrategymax-parallel
	MaxParallel *Int
	Pos         *Pos
}

type EnvVar struct {
	Name  *String
	Value *String
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idsteps
type Step struct {
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepsid
	ID *String
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepsif
	If *String
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepsname
	Name *String
	Exec Exec
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepsenv
	Env map[string]*EnvVar
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepscontinue-on-error
	ContinuesOnError *Bool
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstepstimeout-minutes
	TimeoutMinutes *Float
	Pos            *Pos
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idcontainercredentials
type Credentials struct {
	Username *String
	Password *String
	Pos      *Pos
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idcontainer
type Container struct {
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idcontainerimage
	Image       *String
	Credentials *Credentials
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idcontainerenv
	Env map[string]*EnvVar
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idcontainerports
	Ports []*String
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idcontainervolumes
	Volumes []*String
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idcontaineroptions
	Options *String
	Pos     *Pos
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idservices
type Service struct {
	Name *String
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idservicesservice_idimage
	Image *String
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idservicesservice_idcredentials
	Credentials *Credentials
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idservicesservice_idenv
	Env map[string]*EnvVar
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idservicesservice_idports
	Ports []*String
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idservicesservice_idvolumes
	Volumes []*String
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idservicesservice_idoptions
	Options *String
	Pos     *Pos
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobs
type Output struct {
	Name  *String
	Value *String
}
type Job struct {
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_id
	ID *String
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idname
	Name *String
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idneeds
	Needs []*String
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idruns-on
	RunsOn      *String
	Permissions *Permissions
	Environment *Environment
	Concurrency *Concurrency
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idoutputs
	Outputs map[string]*Output
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idenv
	Env      map[string]*EnvVar
	Defaults *Defaults
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idif
	If *String
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idsteps
	Steps []*Step
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idtimeout-minutes
	TimeoutMinutes *Float
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idstrategy
	Strategy *Strategy
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idcontinue-on-error
	ContinueOnError *Bool
	Container       *Container
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#jobsjob_idservices
	Services map[string]*Service
	Pos      *Pos
}

// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions
type Workflow struct {
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#name
	Name *String
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#onpushpull_requestbranchestags
	On          []Event
	Permissions *Permissions
	// https://docs.github.com/en/actions/reference/workflow-syntax-for-github-actions#env
	Env         map[string]*EnvVar
	Defaults    *Defaults
	Concurrency *Concurrency
	// Jobs is mappings from job ID to the job object
	Jobs map[string]*Job
}
