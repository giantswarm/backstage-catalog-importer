// Provides the 'installations' command to export an installations catalog.
package installations

import (
	"fmt"
	"log"
	"os"

	"github.com/spf13/cobra"

	"github.com/giantswarm/backstage-catalog-importer/pkg/input/installations"
	bscatalog "github.com/giantswarm/backstage-catalog-importer/pkg/output/bscatalog/v1alpha1"
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
			"backstage.io/source-location":    fmt.Sprintf("url:https://github.com/giantswarm/installations/blob/master/%s/cluster.yaml", ins.Codename),
			"opsgenie.com/component-selector": fmt.Sprintf("detailsPair(installation:%s)", ins.Codename),
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

	// Account engineer
	if ins.AccountEngineer != "" {
		r.Annotations["giantswarm.io/account-engineer"] = ins.AccountEngineer
	}

	// Escalation matrix
	if ins.EscalationMatrix != "" {
		r.Annotations["giantswarm.io/escalation-matrix"] = ins.EscalationMatrix
	}

	// Happa and Grafana link
	if ins.Provider == "aws" || ins.Provider == "azure" || ins.Provider == "kvm" {
		// Vintage
		r.Links = append(r.Links, []bscatalog.EntityLink{
			{
				URL:   fmt.Sprintf("https://happa.g8s.%s/admin-login", ins.Base),
				Title: "Happa",
				Icon:  "giantswarm",
			}, {
				URL:   fmt.Sprintf("https://grafana.g8s.%s/", ins.Base),
				Title: "Grafana",
				Icon:  "grafana",
			},
		}...)
	} else {
		r.Links = append(r.Links, []bscatalog.EntityLink{
			{
				URL:   fmt.Sprintf("https://happa.%s.%s/admin-login", ins.Codename, ins.Base),
				Title: "Happa",
				Icon:  "giantswarm",
			}, {
				URL:   fmt.Sprintf("https://grafana.%s.%s/", ins.Codename, ins.Base),
				Title: "Grafana",
				Icon:  "grafana",
			},
		}...)
	}

	// AWS Console link
	if ins.Aws != nil {
		if ins.Aws.HostCluster.Account != "" && ins.Aws.HostCluster.AdminRoleARN != "" {
			r.Links = append(r.Links, bscatalog.EntityLink{
				URL:   fmt.Sprintf(awsCloudProviderURLMask, ins.Aws.HostCluster.Account, ins.Aws.HostCluster.AdminRoleARN, fmt.Sprintf("%s+management+cluster", ins.Codename)),
				Title: "AWS Console (management cluster)",
				Icon:  "aws",
			})
		}
		if ins.Aws.GuestCluster.Account != "" && ins.Aws.GuestCluster.AdminRoleARN != "" {
			r.Links = append(r.Links, bscatalog.EntityLink{
				URL:   fmt.Sprintf(awsCloudProviderURLMask, ins.Aws.GuestCluster.Account, ins.Aws.GuestCluster.AdminRoleARN, fmt.Sprintf("%s+workload+clusters", ins.Codename)),
				Title: "AWS Console (workload clusters)",
				Icon:  "aws",
			})
		}

		// Region
		if ins.Aws.Region != "" {
			r.Labels["giantswarm.io/region"] = ins.Aws.Region
		}
	}

	return r.ToEntity()
}
