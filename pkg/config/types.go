package config

type Config struct {
	Provider Provider     `yaml:"provider"`              // provider configuration
	Env      *Environment `yaml:"environment,omitempty"` // enviorment variables shared across hosts
	Hosts    []Host       `yaml:"hosts"`                 // pool of available machines/servers
	Pipeline Pipeline     `yaml:"pipeline"`              // task pipeline
}

type Provider struct {
	Github Github `yaml:"github"`
}

type Github struct {
	Repository string `yaml:"repository"`     // repository name in the format: user/repo
	Branch     string `yaml:"branch"`         // on what branch to build on
	Auth       *Auth  `yaml:"auth,omitempty"` // PAT Token
}

type Auth struct {
	Token string `yaml:"token"` // PAT token
}

type Environment struct {
	GlobalEnv map[string]string `yaml:"global,omitempty"` // variables shared across hosts
	LocalEnv  map[string]string `yaml:"local,omitempty"`  // variable only on producer server
}

type Host struct {
	Name         string    `yaml:"name"`              // host human readable name
	Address      string    `yaml:"address"`           // local/public accesible address
	InstallSteps *[]string `yaml:"install,omitempty"` // Bootstraping host machine
}

type Pipeline struct {
	Build BuildTaskProducer  `yaml:"build"` // build instructions after cloning the repository
	Tasks []TaskConsumerJobs `yaml:"tasks"` // jobs for the TaskConsumer to execute, runs in parallel by default.
}
type BuildTaskProducer struct {
	Name       string   `yaml:"name"`  // name given for the build task
	BuildSteps []string `yaml:"steps"` // commands to run, sequentially
}
type TaskConsumerJobs struct {
	Name   string   `yaml:"name"`    // name given to each job
	RunsOn []string `yaml:"runs_on"` // array of machines the job will run on

	// if the job runs in parallel to other tasks, set by default to true. if false, the consumer waits for dependents task to finish.
	// we use a pointer since we want to default it to true and we need to know if the field was set.
	RunsInParallel *bool    `yaml:"parallel,omitempty"`
	Commands       []string `yaml:"cmd"`                  // commands to run
	DependsOn      []string `yaml:"depends_on,omitempty"` // on what tasks does this job depends on

	//Option A: build by regex pattern
	Pattern string `yaml:"pattern,omitempty"`

	// Option B: build by explictly specifying files.
	File []string `yaml:"files,omitempty"`
}

type ValidatedConfig struct {
	*Config
	endpoints []EndpointInfo
}
