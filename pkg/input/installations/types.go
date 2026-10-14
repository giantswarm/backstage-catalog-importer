package installations

type Installation struct {
	// Active is false for a decommissioned installation whose record is kept
	// for a re-creation. Absent means active, as in default-cluster.yaml.
	Active                 *bool         `yaml:"active,omitempty"`
	Base                   string        `yaml:"base"`
	Codename               string        `yaml:"codename"`
	Customer               string        `yaml:"customer"`
	CmcRepository          string        `yaml:"cmc_repository"`
	CcrRepository          string        `yaml:"ccr_repository"`
	AccountEngineer        string        `yaml:"accountEngineer"`
	AccountEngineersHandle string        `yaml:"accountEngineersHandle"`
	EscalationMatrix       string        `yaml:"escalation_matrix"`
	Slack                  *SlackDetails `yaml:"slack,omitempty"`
	Pipeline               string        `yaml:"pipeline"`
	Provider               string        `yaml:"provider"`
	Region                 string        `yaml:"region"`
	Aws                    *AwsDetails   `yaml:"aws,omitempty"`
	CustomCA               string
	AccessMarkdown         string
}

// Inactive says whether the installation is marked decommissioned
// (active: false).
func (i *Installation) Inactive() bool {
	return i.Active != nil && !*i.Active
}

type SlackDetails struct {
	Support []string `yaml:"support"`
}

type AwsDetails struct {
	Region       string      `yaml:"region"`
	HostCluster  AwsIdentity `yaml:"hostCluster"`
	GuestCluster AwsIdentity `yaml:"guestCluster"`
}

type AwsIdentity struct {
	Account          string `yaml:"account"`
	AdminRoleARN     string `yaml:"adminRoleARN"`
	CloudtrailBucket string `yaml:"cloudtrailBucket"`
	GuardDuty        bool   `yaml:"guardDuty"`
}
