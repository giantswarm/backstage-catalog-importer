package personio

import (
	"context"
	"sort"

	personiov1 "github.com/giantswarm/personio-go/v1"
)

type Employee struct {
	FirstName    string
	LastName     string
	Email        string
	GithubHandle string
}

// Returns information on active employees from personio.
func GetActiveEmployees(ctx context.Context, clientID, clientSecret string) ([]Employee, error) {
	client, err := personiov1.NewClient(ctx, personiov1.DefaultBaseUrl, personiov1.Credentials{
		ClientId:     clientID,
		ClientSecret: clientSecret,
	})
	if err != nil {
		return nil, err
	}

	employees, err := client.GetEmployees()
	if err != nil {
		return nil, err
	}

	var result []Employee
	for _, employee := range employees {
		// only return active employees
		if *employee.GetStringAttribute("status") != "active" {
			continue
		}

		result = append(result, Employee{
			FirstName:    *employee.GetStringAttribute("first_name"),
			LastName:     *employee.GetStringAttribute("last_name"),
			Email:        *employee.GetStringAttribute("email"),
			GithubHandle: *employee.GetStringAttribute("dynamic_3196204"),
		})
	}

	// Sort the slice by email in ascending order
	sort.Slice(result, func(i, j int) bool {
		return result[i].Email < result[j].Email
	})

	return result, nil
}
