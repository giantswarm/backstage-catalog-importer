package installations

type Installation struct {
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
