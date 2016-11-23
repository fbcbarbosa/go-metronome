package metronome


import (
	"encoding/json"
	"errors"
	"fmt"
	"regexp"

)

var constraintViol = errors.New("Bad constraint.  Must be EQ,LIKE,UNLIKE")
var mountViol = errors.New("Mount point must designate RW,RO")
var containerPathViol = errors.New("Bad container path.  Must match `^/[^/].*$`")

func required(msg string) error {
	if len(msg) == 0 {
		return errors.New("Missing Required message")
	}
	return errors.New(fmt.Sprintf("%s is required by metronome api", msg))
}


// Jobs is a slice of jobs
type Artifact struct {
	Uri_        string `json:"uri"`
	Executable_ bool   `json:"executable"`
	Extract_    bool   `json:"extract"`
	Cache_      bool   `json:"cache"`
}

func (self *Artifact) Uri() string {
	return self.Uri_
}
func (self *Artifact) Executable() bool {
	return self.Executable_
}
func (self *Artifact) Extract() bool {
	return self.Extract_
}
func (self *Artifact) Cache() bool {
	return self.Cache_
}

type Docker struct {
	Image_ string `json:"image"`
}

func NewDockerImage(image string) (*Docker, error) {
	if len(image) == 0 {
		return nil, required("Docker.Image requires a value")
	}
	return &Docker{Image_: image}, nil
}

func (self *Docker) Image() string {
	return self.Image_
}

// constraint

type Operator int

const (
	EQ Operator = 1 + iota
	LIKE
	UNLIKE
)

var constraint_operators = [...]string{
	"EQ",
	"LIKE",
	"UNLIKE",
}

func (self *Operator) String() string {
	return constraint_operators[int(*self) - 1]
}

func (self *Operator) UnmarshalJSON(raw []byte) error {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return err
	}
	switch s {
	case "EQ":
		*self = EQ
	case "LIKE":
		*self = LIKE
	case "UNLIKE":
		*self = UNLIKE
	default:
		fmt.Printf("Operator.UnmarshallJSON - unknown value '%s'\n", s)
		return constraintViol
	}
	return nil
}
func (self *Operator) MarshalJSON() ([]byte, error) {
	s := self.String()
	return []byte(fmt.Sprintf("\"%s\"", s)), nil
}

type Constraint struct {
	Attribute_ string `json:"attribute"`
	// operator is EQ, LIKE,UNLIKE
	Operator_  Operator `json:"operator"`
	Value_     string   `json:"value"`
}

func NewConstraint(attribute string, op Operator, value string) (*Constraint, error) {
	if attribute == "" {
		return nil, required("Constraint.attribute")
	}
	return &Constraint{Attribute_: attribute, Operator_: op, Value_: value}, nil
}

func (self *Constraint) Attribute() string {
	return self.Attribute_
}
func (self *Constraint) Operator() Operator {
	return self.Operator_
}
func (self *Constraint) Value() string {
	return self.Value_
}

type Placement struct {
	Constraints_ []Constraint `json:"constraints,omitempty"`
}

func (self *Placement) Constraints() ([]Constraint, error) {
	return self.Constraints_, nil
}

// volumes

type MountMode int
type ContainerPath string

const (
	RO MountMode = 1 + iota
	RW
)

var mount_modes = [...]string{
	"RO",
	"RW",
}

func (self MountMode) String() string {
	return mount_modes[int(self) - 1]
}
func (self *MountMode) MarshalJSON() ([]byte, error) {
	//s := self.String()
	return []byte(fmt.Sprintf("\"%s\"", mount_modes[int(*self) - 1])), nil
}
func (self *MountMode) UnmarshalJSON(raw []byte) error {
	var s string
	if err := json.Unmarshal(raw, &s); err != nil {
		return err
	}
	switch s {
	case "RO":
		*self = RO
	case "RW":
		*self = RW
	default:
		return mountViol
	}
	return nil
}
func (self *ContainerPath) UnmarshalJSON(raw []byte) error {

	if _, err := regexp.MatchString("^/[^/].*$", string(raw)); err != nil {
		return containerPathViol
	}
	*self = ContainerPath(raw)
	return nil
}
func (self *ContainerPath) String() string {
	return string(*self)
}

func NewContainerPath(path string) (self *ContainerPath, err error) {
	if _, err = regexp.MatchString("^/[^/].*$", path); err != nil {
		return nil, err
	}
	vg := ContainerPath(path)

	return &vg, nil

}

type Volume struct {
	// minlength 1; pattern: ^/[^/].*$
	ContainerPath_ ContainerPath `json:"containerPath"`
	HostPath_      string        `json:"hostPath"`
	// Values: RW,RO
	Mode_          MountMode `json:"mode"`
}

func (self *Volume) ContainerPath() (ContainerPath, error) {
	return self.ContainerPath_, nil
}
func (self *Volume) HostPath() (string, error) {
	return self.HostPath_, nil
}
func (self *Volume) Mode() (MountMode, error) {
	return self.Mode_, nil
}

func NewVolume(raw_path string, hostPath string, mode MountMode) (*Volume, error) {
	vol := Volume{Mode_: mode, HostPath_:hostPath}

	// ensure valid path
	if cpath, err := NewContainerPath(raw_path); err != nil {
		return nil, err
	} else {
		vol.ContainerPath_ = *cpath
	}
	if vol.HostPath_ == "" {
		return nil, required("host path")
	}
	return &vol, nil
}

type Restart struct {
	ActiveDeadlineSeconds_ int    `json:"activeDeadlineSeconds"`
	Policy_                string `json:"policy"`
}

func (self *Restart) ActiveDeadlineSeconds() int {
	return self.ActiveDeadlineSeconds_
}

func (self *Restart) Policy() string {
	return self.Policy_
}

func NewRestart(activeDeadlineSeconds int, policy string) (*Restart, error) {
	if len(policy) == 0 {
		return nil, required("length(Restart.policy)>0")
	}
	return &Restart{ActiveDeadlineSeconds_: activeDeadlineSeconds, Policy_: policy}, nil
}

type Run struct {
	Artifacts_      []Artifact  `json:"artifacts,omitempty"`
	Cmd_            string      `json:"cmd,omitempty"`

	Args_           []string          `json:"args,omitempty"`
	Cpus_           float64           `json:"cpus"`
	Mem_            int               `json:"mem"`
	Disk_           int               `json:"disk"`
	Docker_         *Docker           `json:"docker,omitempty"`
	Env_            map[string]string `json:"env,omitempty"`
	MaxLaunchDelay_ int               `json:"maxLaunchDelay"`
	Placement_      *Placement        `json:"placement,omitempty"`
	Restart_        *Restart          `json:"restart,omitempty"`
	User_           string            `json:"user,omitempty"`
	Volumes_        []Volume         `json:"volumes,omitempty"`
}

func (self *Run) String() string {
	rez := fmt.Sprint("cpus: %f disk: %d mem: %d\n", self.Cpus_, self.Disk_, self.Mem_)
	return rez
}

func (self *Run) Artifacts() []Artifact {
	return self.Artifacts_
}

func (self *Run) SetArtifacts(artifacts []Artifact) *Run {
	self.Artifacts_ = artifacts
	return self
}

func (self *Run) Cmd() string {
	return self.Cmd_
}

func (self *Run) SetCmd(cmd string) *Run {
	self.Cmd_ = cmd
	return self
}

func (self *Run) Args() *[]string {
	return &self.Args_
}

func (self *Run) AddArg(item string) {
	self.Args_ = append(self.Args_, item)
}
func (self *Run) SetArgs(newargs [] string) *Run {
	self.Args_ = newargs
	return self
}
// cpu
func (self *Run) Cpus() float64 {
	return self.Cpus_
}
func (self *Run) SetCpus(p float64) *Run {
	self.Cpus_ = p
	return self
}
// memory
func (self *Run) Mem() int {
	return self.Mem_
}
func (self *Run) SetMem(p int) *Run {
	self.Mem_ = p
	return self
}
// disk
func (self *Run) Disk() int {
	return self.Disk_
}
func (self *Run) SetDisk(p int) *Run {
	self.Disk_ = p
	return self
}

func (self *Run) Docker() *Docker {
	return self.Docker_
}
func (self *Run) SetDocker(docker *Docker) *Run {
	self.Docker_ = docker
	return self
}
func (self *Run) Env() map[string]string {
	return self.Env_
}
func (self *Run) SetEnv(mp map[string]string) *Run {
	self.Env_ = mp
	return self
}

func (self *Run) MaxLaunchDelay() int {
	return self.MaxLaunchDelay_
}
func (self *Run) SetMaxLaunchDelay(p int) *Run {
	self.MaxLaunchDelay_ = p
	return self
}

func (self *Run) Placement() *Placement {
	return self.Placement_
}
func (self *Run) SetPlacement(p *Placement) *Run {
	self.Placement_ = p
	return self
}

func (self *Run) Restart() *Restart {
	return self.Restart_
}
func (self *Run) SetRestart(restart *Restart) *Run {
	self.Restart_ = restart
	return self
}

func (self *Run) User() string {
	return self.User_
}
func (self *Run) SetUser(user string) *Run {
	self.User_ = user
	return self
}

// make Volume ifc
func (self *Run) Volumes() *[]Volume {
	return &self.Volumes_
}
func (self *Run) SetVolumes(vols []Volume) *Run {
	self.Volumes_ = vols
	return self
}

func NewRun(cpus float64, mem int, disk int) (*Run, error) {
	if mem <= 0 {
		return nil, required("Run.memory")
	}
	if disk <= 0 {
		return nil, required("Run.disk")
	}
	if cpus <= 0.0 {
		return nil, required("Run.cpus")
	}
	vg := Run{
		Artifacts_: make([]Artifact, 0, 10),
		Args_: make([]string, 0, 0),
		Cpus_: cpus,
		Mem_: mem,
		Disk_: disk,
		Docker_:  nil,
		Env_: make(map[string]string),
		MaxLaunchDelay_: 0,
		Placement_: nil,
		Restart_: nil,
		Volumes_: make([]Volume, 0, 0),
	}
	vg.Args_ = append(vg.Args_, "foo")
	return &vg, nil
}

type Labels struct {
	Location string `json:"location"`
	Owner    string `json:"owner"`
}

type Job struct {
	Description_ string `json:"description"`
	ID_          string `json:"id"`
	Labels_      *Labels`json:"labels,omitempty"`
	Run_         *Run `json:"run"`
}

func NewJob(id string, description string, labels *Labels, run *Run) (*Job, error) {
	if len(id) == 0 {
		return nil, required("Job.Id")
	}
	if run == nil {
		return nil, required("Job.run")
	}
	return &Job{ID_: id,
		Description_: description,
		Labels_: labels,
		Run_: run,
	}, nil
}
func (self *Job) Description() string {
	return self.Description_
}
func (self *Job) SetDescription(desc string) *Job {
	self.Description_ = desc
	return self
}
func (self *Job) Run() *Run {
	return self.Run_
}
func (self *Job) SetRun(run *Run) *Job {
	self.Run_ = run
	return self
}
func (self *Job) Labels() *Labels {
	return self.Labels_
}
func (self *Job) SetLabel(label Labels) *Job {
	self.Labels_ = &label
	return self
}