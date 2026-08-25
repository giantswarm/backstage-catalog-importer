// Provides the 'installations' command to export an installations catalog.
package installations

import (
	"fmt"
	"log"
	"os"
	"sort"

	"github.com/spf13/cobra"

	"github.com/giantswarm/backstage-catalog-importer/pkg/input/installations"
	bscatalog "github.com/giantswarm/backstage-catalog-importer/pkg/output/bscatalog/v1alpha1"
	"github.com/giantswarm/backstage-catalog-importer/pkg/output/catalog/group"
	"github.com/giantswarm/backstage-catalog-importer/pkg/output/catalog/resource"
	"github.com/giantswarm/backstage-catalog-importer/pkg/output/export"
)

var Command = &cobra.Command{
	Use:   "installations",
	Short: "Export installations catalog",
	Long:  `Exports Giant Swarm installations for GS-internal use.`,
	RunE:  run,
}

const (
	orgFlag    = "org"
	repoFlag   = "repo"
	outputFlag = "output"

	awsCloudProviderURLMask = "https://signin.aws.amazon.com/switchrole?account=%s&roleName=%s&displayName=%s"

	providerAWS = "aws"

	iconAWS     = "aws"
	iconGrafana = "grafana"
)

func init() {
	Command.Flags().String(orgFlag, "giantswarm", "GitHub organization to export users from")
	Command.Flags().String(repoFlag, "installations", "Name of the repository containing installation data")

	Command.PersistentFlags().StringP(outputFlag, "o", ".", "Output directory path")
}

func run(cmd *cobra.Command, args []string) error {
	token := os.Getenv("GITHUB_TOKEN")
	if token == "" {
		log.Fatal("Please set environment variable GITHUB_TOKEN to a personal GitHub access token (PAT).")
	}

	org, err := cmd.Flags().GetString(orgFlag)
	if err != nil {
		log.Fatalf("Error: could not access '--org' flag - %s", err)
	}
	repo, err := cmd.Flags().GetString(repoFlag)
	if err != nil {
		log.Fatalf("Error: could not access '--repo' flag - %s", err)
	}
	path, err := cmd.PersistentFlags().GetString(outputFlag)
	if err != nil {
		log.Fatalf("Error: could not access '--output' flag - %s", err)
	}

	installationsExporter := export.New(export.Config{TargetPath: path + "/installations.yaml"})

	insService, err := installations.New(installations.Config{
		GithubOrganization:   org,
		GithubRepositoryName: repo,
		GithubAuthToken:      token,
	})
	if err != nil {
		log.Fatalf("Error: could not create service -- %v", err)
	}

	ins, err := insService.GetInstallations()
	if err != nil {
		log.Fatalf("Error: could not write installations -- %v", err)
	}

	// Installations are owned by customer groups which exist nowhere else in
	// the catalog, so export one Group entity per distinct customer to avoid
	// dangling owner relations.
	for _, e := range toCustomerGroupEntities(ins) {
		err = installationsExporter.AddEntity(e)
		if err != nil {
			log.Fatalf("Error: could not add customer group -- %v", err)
		}
	}

	for _, installation := range ins {
		e := toResourceEntity(installation)
		err = installationsExporter.AddEntity(e)
		if err != nil {
			log.Fatalf("Error: could add installation resource -- %v", err)
		}
	}

	err = installationsExporter.WriteFile()
	if err != nil {
		log.Fatalf("Error: could not write installations -- %v", err)
	}

	return nil
}

// toCustomerGroupEntities returns one Group entity of type "customer" per
// distinct customer owning at least one installation, sorted by name.
func toCustomerGroupEntities(ins []*installations.Installation) []*bscatalog.Entity {
	customers := make(map[string]bool)
	for _, i := range ins {
		if i.Customer != "" {
			customers[i.Customer] = true
		}
	}

	names := make([]string, 0, len(customers))
	for name := range customers {
		names = append(names, name)
	}
	sort.Strings(names)

	entities := make([]*bscatalog.Entity, 0, len(names))
	for _, name := range names {
		g, err := group.New(name,
			group.WithType("customer"),
			group.WithDescription(fmt.Sprintf("Customer %s", name)),
		)
		if err != nil {
			log.Fatalf("Error: could not create customer group %q -- %v", name, err)
		}
		entities = append(entities, g.ToEntity())
	}

	return entities
}

func toResourceEntity(ins *installations.Installation) *bscatalog.Entity {
	r := resource.Resource{
		Name:        ins.Codename,
		Title:       ins.Codename,
		Owner:       ins.Customer,
		Type:        "installation",
		Description: fmt.Sprintf("%s installation on %s owned by %s", ins.Pipeline, ins.Provider, ins.Customer),
		Labels: map[string]string{
			"giantswarm.io/provider": ins.Provider,
			"giantswarm.io/customer": ins.Customer,
			"giantswarm.io/pipeline": ins.Pipeline,
		},
		Annotations: map[string]string{
			"backstage.io/source-location": fmt.Sprintf("url:https://github.com/giantswarm/installations/blob/master/%s/cluster.yaml", ins.Codename),
		},
		Links: []bscatalog.EntityLink{
			{
				URL:   fmt.Sprintf("https://github.com/giantswarm/%s", ins.CmcRepository),
				Title: "Customer management clusters (CMC)",
				Icon:  "github",
				Type:  "CMC",
			},
			{
				URL:   fmt.Sprintf("https://github.com/giantswarm/%s", ins.CcrRepository),
				Title: "Customer config (CCR)",
				Icon:  "github",
				Type:  "CCR",
			},
		},
		Spec: bscatalog.ComponentSpec{},
	}

	// Base domain
	if ins.Base != "" {
		r.Annotations["giantswarm.io/base"] = ins.Base
	}
	// Access docs
	if ins.AccessMarkdown != "" {
		r.Annotations["giantswarm.io/access-docs-markdown"] = ins.AccessMarkdown
	}

	// Custom CA
	if ins.CustomCA != "" {
		r.Annotations["giantswarm.io/custom-ca"] = ins.CustomCA
	}

	// Region
	if ins.Region != "" {
		r.Labels["giantswarm.io/region"] = ins.Region
	}

	// Account engineer
	if ins.AccountEngineer != "" {
		r.Annotations["giantswarm.io/account-engineer"] = ins.AccountEngineer
	}

	// Escalation matrix
	if ins.EscalationMatrix != "" {
		r.Annotations["giantswarm.io/escalation-matrix"] = ins.EscalationMatrix
	}

	// Happa and Grafana link
	if ins.Provider == providerAWS || ins.Provider == "azure" || ins.Provider == "kvm" {
		// Vintage
		r.Links = append(r.Links, []bscatalog.EntityLink{
			{
				URL:   fmt.Sprintf("https://happa.g8s.%s/admin-login", ins.Base),
				Title: "Happa",
				Icon:  "giantswarm",
			}, {
				URL:   fmt.Sprintf("https://grafana.g8s.%s/", ins.Base),
				Title: "Grafana",
				Icon:  iconGrafana,
			},
		}...)
	} else {
		r.Links = append(r.Links, []bscatalog.EntityLink{
			{
				URL:   fmt.Sprintf("https://grafana-%s.teleport.giantswarm.io", ins.Codename),
				Title: "Grafana (via Teleport)",
				Icon:  iconGrafana,
			}, {
				URL:   fmt.Sprintf("https://grafana.%s.%s/", ins.Codename, ins.Base),
				Title: "Grafana (customer URL)",
				Icon:  iconGrafana,
			}, {
				URL:   fmt.Sprintf("https://happa.%s.%s/admin-login", ins.Codename, ins.Base),
				Title: "Happa",
				Icon:  "giantswarm",
			}, {
				URL:   fmt.Sprintf("https://kyverno-%s.teleport.giantswarm.io", ins.Codename),
				Title: "Policy Reporter (via Teleport)",
				Icon:  "dashboard",
			},
		}...)
	}

	// AWS Console link
	if ins.Aws != nil {
		if ins.Aws.HostCluster.Account != "" && ins.Aws.HostCluster.AdminRoleARN != "" {
			r.Links = append(r.Links, bscatalog.EntityLink{
				URL:   fmt.Sprintf(awsCloudProviderURLMask, ins.Aws.HostCluster.Account, ins.Aws.HostCluster.AdminRoleARN, fmt.Sprintf("%s+management+cluster", ins.Codename)),
				Title: "AWS Console (management cluster)",
				Icon:  iconAWS,
			})
		}
		if ins.Aws.GuestCluster.Account != "" && ins.Aws.GuestCluster.AdminRoleARN != "" {
			r.Links = append(r.Links, bscatalog.EntityLink{
				URL:   fmt.Sprintf(awsCloudProviderURLMask, ins.Aws.GuestCluster.Account, ins.Aws.GuestCluster.AdminRoleARN, fmt.Sprintf("%s+workload+clusters", ins.Codename)),
				Title: "AWS Console (workload clusters)",
				Icon:  iconAWS,
			})
		}
	}

	return r.ToEntity()
}
