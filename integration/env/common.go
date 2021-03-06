package env

import (
	"crypto/sha1"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/giantswarm/e2e-harness/pkg/framework"
)

const (
	// EnvVarCircleCI is the process environment variable representing the
	// CIRCLECI env var.
	EnvVarCircleCI = "CIRCLECI"
	// EnvVarCircleSHA is the process environment variable representing the
	// CIRCLE_SHA1 env var.
	EnvVarCircleSHA = "CIRCLE_SHA1"
	// EnvVarClusterID is the process environment variable representing the
	// CLUSTER_NAME env var.
	//
	// TODO rename to CLUSTER_ID. Note this also had to be changed in the
	// framework package of e2e-harness.
	EnvVarClusterID = "CLUSTER_NAME"
	// EnvVarCommonDomain is the process environment variable representing the
	// COMMON_DOMAIN env var.
	EnvVarCommonDomain = "COMMON_DOMAIN"
	// EnvVarGithubBotToken is the process environment variable representing
	// the GITHUB_BOT_TOKEN env var.
	EnvVarGithubBotToken = "GITHUB_BOT_TOKEN"
	// EnvVarKeepResources is the process environment variable representing the
	// KEEP_RESOURCES env var.
	EnvVarKeepResources = "KEEP_RESOURCES"
	// EnvVarRegistryPullSecret is the process environment variable representing the
	// REGISTRY_PULL_SECRET env var.
	EnvVarRegistryPullSecret = "REGISTRY_PULL_SECRET"
	// EnvVarTestedVersion is the process environment variable representing the
	// TESTED_VERSION env var.
	EnvVarTestedVersion = "TESTED_VERSION"
	// EnvVarTestDir is the process environment variable representing the
	// TEST_DIR env var.
	EnvVarTestDir = "TEST_DIR"
	// EnvVaultToken is the process environment variable representing the
	// VAULT_TOKEN env var.
	EnvVaultToken = "VAULT_TOKEN"
	// EnvVarVersionBundleVersion is the process environment variable representing
	// the VERSION_BUNDLE_VERSION env var.
	EnvVarVersionBundleVersion = "VERSION_BUNDLE_VERSION"

	// operator namespace suffix
	operatorNamespaceSuffix = "op"
)

var (
	circleCI             string
	circleSHA            string
	clusterID            string
	commonDomain         string
	githubToken          string
	testDir              string
	testedVersion        string
	keepResources        string
	registryPullSecret   string
	vaultToken           string
	versionBundleVersion string
)

func init() {
	var err error

	circleCI = os.Getenv(EnvVarCircleCI)
	keepResources = os.Getenv(EnvVarKeepResources)

	circleSHA = os.Getenv(EnvVarCircleSHA)
	if circleSHA == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarCircleSHA))
	}

	testedVersion = os.Getenv(EnvVarTestedVersion)
	if testedVersion == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarTestedVersion))
	}

	testDir = os.Getenv(EnvVarTestDir)

	// NOTE that implications of changing the order of initialization here means
	// breaking the initialization behaviour.
	clusterID := os.Getenv(EnvVarClusterID)
	if clusterID == "" {
		os.Setenv(EnvVarClusterID, ClusterID())
	}

	commonDomain = os.Getenv(EnvVarCommonDomain)
	if commonDomain == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarCommonDomain))
	}

	vaultToken = os.Getenv(EnvVaultToken)
	if vaultToken == "" {
		panic(fmt.Sprintf("env var %q must not be empty", EnvVaultToken))
	}

	githubToken = os.Getenv(EnvVarGithubBotToken)
	if githubToken == "" {
		panic(fmt.Sprintf("env var %q must not be empty", EnvVarGithubBotToken))
	}

	registryPullSecret = os.Getenv(EnvVarRegistryPullSecret)
	if registryPullSecret == "" {
		panic(fmt.Sprintf("env var '%s' must not be empty", EnvVarRegistryPullSecret))
	}

	params := &framework.VBVParams{
		Component: "kvm-operator",
		Provider:  "kvm",
		Token:     githubToken,
		VType:     TestedVersion(),
	}
	versionBundleVersion, err = framework.GetVersionBundleVersion(params)
	if err != nil {
		panic(err.Error())
	}
	// TODO there should be a not found error returned by the framework in such
	// cases.
	if VersionBundleVersion() == "" {
		if strings.ToLower(TestedVersion()) == "wip" {
			log.Println("WIP version bundle version not present, exiting.")
			os.Exit(0)
		}
		panic("version bundle version  must not be empty")
	}
	os.Setenv(EnvVarVersionBundleVersion, VersionBundleVersion())
}

func CircleCI() string {
	return circleCI
}

func CircleSHA() string {
	return circleSHA
}

// ClusterID returns a cluster ID unique to a run integration test. It might
// look like ci-wip-3cc75.
//
//     ci is a static identifier stating a CI run of the aws-operator.
//     wip is a version reference which can also be cur for the current version.
//     3cc75 is the Git SHA and the integration test dir combined.
//
func ClusterID() string {
	var parts []string

	parts = append(parts, "ci")
	parts = append(parts, TestedVersion()[0:3])
	shaPart := CircleSHA()[0:4]
	testPart := TestHash()

	h := sha1.New()
	h.Write([]byte(shaPart + testPart))
	s := fmt.Sprintf("%x", h.Sum(nil))[0:5]

	parts = append(parts, s)

	return strings.Join(parts, "-")
}

func CommonDomain() string {
	return commonDomain
}

func GithubToken() string {
	return githubToken
}

func KeepResources() string {
	return keepResources
}

func RegistryPullSecret() string {
	return registryPullSecret
}

func TargetNamespace() string {
	return fmt.Sprintf("%s-%s", ClusterID(), operatorNamespaceSuffix)
}

func TestedVersion() string {
	return testedVersion
}

func TestDir() string {
	return testDir
}

func TestHash() string {
	if TestDir() == "" {
		return ""
	}

	h := sha1.New()
	h.Write([]byte(TestDir()))
	s := fmt.Sprintf("%x", h.Sum(nil))[0:5]

	return s
}

func VaultToken() string {
	return vaultToken
}

func VersionBundleVersion() string {
	return versionBundleVersion
}
