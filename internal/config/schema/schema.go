package schema

type NpmConfig struct {
	Enabled *bool    `yaml:"enabled,omitempty"`
	Service string   `yaml:"service"`
	Scripts []string `yaml:"scripts"`
}

type GoConfig struct {
	Enabled *bool  `yaml:"enabled,omitempty"`
	Service string `yaml:"service"`
}

type PythonConfig struct {
	Enabled        *bool  `yaml:"enabled,omitempty"`
	Django         bool   `yaml:"django"`
	DjangoService  string `yaml:"django_service"`
	PackageManager string `yaml:"package_manager"`
}

type GCPConfig struct {
	ProjectName               string `yaml:"project_name"`
	Account                   string `yaml:"account,omitempty"`
	ImpersonateServiceAccount string `yaml:"impersonate_service_account,omitempty"`
}

type TerraformConfig struct {
	Dir        string   `yaml:"dir,omitempty"`
	UseFolders bool     `yaml:"-"`
	Envs       []string `yaml:"envs,omitempty"`
}

type ModuleConfig struct {
	Python *PythonConfig `yaml:"python,omitempty"`
	Npm    *NpmConfig    `yaml:"npm,omitempty"`
	Go     *GoConfig     `yaml:"go,omitempty"`
}

type ServiceConfig struct {
	Name       string         `yaml:"name"`
	Dir        string         `yaml:"dir"`
	Docker     *bool          `yaml:"docker,omitempty"`
	Dockerfile string         `yaml:"dockerfile,omitempty"`
	Image      string         `yaml:"image,omitempty"`
	Command    string         `yaml:"command,omitempty"`
	Modules    []ModuleConfig `yaml:"modules"`
	AppYaml    string         `yaml:"app_yaml,omitempty"`
}

func (s *ServiceConfig) IsDocker() bool {
	if s == nil {
		return false
	}
	return s.Docker != nil && *s.Docker
}

func (p *PythonConfig) IsEnabled() bool {
	if p == nil {
		return false
	}
	return p.Enabled == nil || *p.Enabled
}

func (n *NpmConfig) IsEnabled() bool {
	if n == nil {
		return false
	}
	return n.Enabled == nil || *n.Enabled
}

func (g *GoConfig) IsEnabled() bool {
	if g == nil {
		return false
	}
	return g.Enabled == nil || *g.Enabled
}

type Workflow struct {
	ID       string   `yaml:"id" json:"id"`
	Name     string   `yaml:"name" json:"name"`
	Commands []string `yaml:"commands" json:"commands"`
}

type Config struct {
	Version             int              `yaml:"version"`
	Docker              bool             `yaml:"docker"`
	GoogleCloudPlatform *GCPConfig       `yaml:"google_cloud_platform,omitempty"`
	Terraform           *TerraformConfig `yaml:"terraform,omitempty"`
	Envs                []string         `yaml:"envs,omitempty"`
	Services            []ServiceConfig  `yaml:"services"`
	AppYaml             string           `yaml:"app_yaml,omitempty"`
	Workflows           []Workflow       `yaml:"workflows,omitempty"`

	// Inputs stores transient values collected during execution
	Inputs map[string]string `yaml:"-"`

	// SourcePath is the absolute path to the loaded config file
	SourcePath string `yaml:"-"`
}
