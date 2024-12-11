package personio

import (
	"context"
	"log"

	personiov1 "github.com/giantswarm/personio-go/v1"
)

func Lookup() error {
	var personioCredentials personiov1.Credentials

	client, err := personiov1.NewClient(context.TODO(), personiov1.DefaultBaseUrl, personioCredentials)
	if err != nil {
		return err
	}

	employees, err := client.GetEmployees()
	if err != nil {
		return err
	}

	for _, employee := range employees {
		log.Printf("Employee: %s %s email=%s github=%s",
			*employee.GetStringAttribute("first_name"),
			*employee.GetStringAttribute("last_name"),
			*employee.GetStringAttribute("email"),
			*employee.GetStringAttribute("github_username"),
		)
	}

	return nil
}
