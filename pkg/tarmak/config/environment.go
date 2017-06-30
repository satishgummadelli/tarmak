package config

import (
	"errors"
	"fmt"
	"net"

	log "github.com/Sirupsen/logrus"
	"github.com/hashicorp/go-multierror"
)

type Environment struct {
	Name string     `yaml:"name,omitempty"` // only alphanumeric lowercase
	AWS  *AWSConfig `yaml:"aws,omitempty"`
	GCP  *GCPConfig `yaml:"gcp,omitempty"`

	Contexts []Context `yaml:"contexts,omitempty"`

	stackState *StackState
	stackVault *StackVault
}

func (e *Environment) Validate() error {
	var result error

	networkCIDRs := []*net.IPNet{}
	e.stackState = nil
	e.stackVault = nil
	for posContext, _ := range e.Contexts {
		context := &e.Contexts[posContext]

		// set myself in the context
		context.environment = e

		// ensure context validates
		if err := context.Validate(); err != nil {
			result = multierror.Append(result, err)
		}

		// get network cidr
		if net := context.NetworkCIDR(); net != nil {
			networkCIDRs = append(networkCIDRs, net)
		}

		// loop through stacks
		for posStack, _ := range context.Stacks {
			stack := context.Stacks[posStack]

			// ensure no multiple state stacks
			if stack.StackName() == StackNameState {
				if e.stackState == nil {
					e.stackState = stack.State
				} else {
					result = multierror.Append(result, fmt.Errorf("environment '%s' has multiple state stacks", e.Name))
				}
			}

			// ensure no multiple vault stacks
			if stack.StackName() == StackNameVault {
				if e.stackVault == nil {
					e.stackVault = stack.Vault
				} else {
					result = multierror.Append(result, fmt.Errorf("environment '%s' has multiple vault stacks", e.Name))
				}
			}

		}
	}

	// ensure there is a state stack
	if e.stackState == nil {
		result = multierror.Append(result, fmt.Errorf("environment '%s' has no state stack", e.Name))
	}

	// ensure there is a vault stack
	if e.stackVault == nil {
		result = multierror.Append(result, fmt.Errorf("environment '%s' has no vault stack", e.Name))
	}

	// validate network overlap
	if err := validateNetworkOverlap(networkCIDRs); err != nil {
		result = multierror.Append(result, err)
	}

	return result
}

func validateNetworkOverlap(netCIDRs []*net.IPNet) error {
	var result error
	for i, _ := range netCIDRs {
		for j := i + 1; j < len(netCIDRs); j++ {
			// check for overlap per network
			if netCIDRs[i].Contains(netCIDRs[j].IP) || netCIDRs[j].Contains(netCIDRs[i].IP) {
				result = multierror.Append(result, fmt.Errorf(
					"network '%s' overlaps with '%s'",
					netCIDRs[i].String(),
					netCIDRs[j].String(),
				))
			}
		}
	}
	return result
}

func (c *Environment) ProviderName() string {
	providerName, err := c.getProviderName()
	if err != nil {
		return ""
	}
	return providerName
}

func (e *Environment) RemoteState(contextName, stackName string) string {
	if e.ProviderName() == ProviderNameAWS {
		return e.AWS.RemoteState(
			e.RemoteStateBucketName(),
			e.Name,
			contextName,
			stackName,
		)
	}

	log.Fatalf("unsupported provider: '%s'", e.ProviderName())
	return ""

}

func (e *Environment) RemoteStateBucketName() string {
	if e.ProviderName() == ProviderNameAWS {
		return fmt.Sprintf(
			"%s%s-%s-terraform-state",
			e.stackState.BucketPrefix,
			e.Name,
			e.AWS.Region,
		)
	}

	log.Fatalf("unsupported provider: '%s'", e.ProviderName())
	return ""
}

func (e *Environment) RemoteStateAvailable() bool {
	if e.ProviderName() == ProviderNameAWS {
		return e.AWS.RemoteStateAvailable(e.RemoteStateBucketName())
	}

	log.Fatalf("unsupported provider: '%s'", e.ProviderName())
	return false
}

func (e *Environment) ProviderEnvironment() ([]string, error) {
	if e.ProviderName() == ProviderNameAWS {
		return e.AWS.Environment()
	}
	return []string{}, fmt.Errorf("unsupported provider: '%s'", e.ProviderName())
}

func (c *Environment) getProviderName() (string, error) {
	providers := []string{}
	if c.AWS != nil {
		providers = append(providers, ProviderNameAWS)
	}
	if c.GCP != nil {
		providers = append(providers, ProviderNameGCP)
	}

	if len(providers) < 1 {
		return "", errors.New("please specify exactly one provider")
	}
	if len(providers) > 1 {
		return "", fmt.Errorf("more than one provider given: %+v", providers)
	}

	return providers[0], nil
}
