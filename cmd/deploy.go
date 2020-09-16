package cmd

import (
	"fmt"

	"github.com/ory/viper"
	"github.com/spf13/cobra"

	"github.com/boson-project/faas"
	"github.com/boson-project/faas/docker"
	"github.com/boson-project/faas/knative"
	"github.com/boson-project/faas/prompt"
)

func init() {
	root.AddCommand(deployCmd)
	deployCmd.Flags().BoolP("confirm", "c", false, "Prompt to confirm all configuration options - $FAAS_CONFIRM")
	deployCmd.Flags().StringP("namespace", "n", "", "Override namespace into which the Function is deployed (on supported platforms).  Default is to use currently active underlying platform setting - $FAAS_NAMESPACE")
}

var deployCmd = &cobra.Command{
	Use:        "deploy",
	Short:      "Deploy an existing Function project to a cluster",
	SuggestFor: []string{"delpoy", "deplyo"},
	PreRunE:    bindEnv("namespace", "confirm"),
	RunE:       runDeploy,
}

func runDeploy(cmd *cobra.Command, _ []string) (err error) {
	config := newDeployConfig()
	function, err := faas.LoadFunction(config.Path)
	if err != nil {
		return
	}
	if err = function.OverrideNamespace(config.Namespace); err != nil {
		return
	}

	if function.Image == "" {
		return fmt.Errorf("Cannot determine the Function image name. Have you built it yet?")
	}

	// Confirm or print configuration
	config.Prompt()

	pusher := docker.NewPusher()
	pusher.Verbose = config.Verbose

	deployer := knative.NewDeployer()
	deployer.Verbose = config.Verbose

	client := faas.New(
		faas.WithVerbose(config.Verbose),
		faas.WithPusher(pusher),
		faas.WithDeployer(deployer))

	return client.Deploy(config.Path)

	// NOTE: Namespace is optional, default is that used by k8s client
	// (for example kubectl usually uses ~/.kube/config)
}

type deployConfig struct {
	// Namespace override for the deployed function.  If provided, the
	// underlying platform will be instructed to deploy the function to the given
	// namespace (if such a setting is applicable; such as for Kubernetes
	// clusters).  If not provided, the currently configured namespace will be
	// used.  For instance, that which would be used by default by `kubectl`
	// (~/.kube/config) in the case of Kubernetes.
	Namespace string

	// Path of the Function implementation on local disk. Defaults to current
	// working directory of the process.
	Path string

	// Verbose logging.
	Verbose bool

	// Confirm: confirm values arrived upon from environment plus flags plus defaults,
	// with interactive prompting (only applicable when attached to a TTY).
	Confirm bool
}

// newDeployConfig creates a buildConfig populated from command flags and
// environment variables; in that precedence.
func newDeployConfig() deployConfig {
	return deployConfig{
		Namespace: viper.GetString("namespace"),
		Path:      cwd(),
		Verbose:   viper.GetBool("verbose"), // defined on root
		Confirm:   viper.GetBool("confirm"),
	}
}

// Prompt the user with value of config members, allowing for interaractive changes.
// Skipped if not in an interactive terminal (non-TTY), or if --yes (agree to
// all prompts) was explicitly set.
func (c deployConfig) Prompt() deployConfig {
	if !interactiveTerminal() || !c.Confirm {
		return c
	}
	return deployConfig{
		Namespace: prompt.ForString("Namespace", c.Namespace),
		Path:      prompt.ForString("Project path", c.Path),
		Verbose:   c.Verbose,
	}
}
