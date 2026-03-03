package acctest

import (
	"github.com/filipowm/go-unifi/unifi"
	"github.com/p3terp4N/terraform-provider-unifi-express/internal/provider"
	pt "github.com/p3terp4N/terraform-provider-unifi-express/internal/provider/testing"
	"github.com/hashicorp/go-version"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"os"
	"sync"
	"testing"
)

type providersMap map[string]func() (tfprotov6.ProviderServer, error)

var (
	providers  = createProviders()
	testClient unifi.Client
)

type Steps []resource.TestStep

type AcceptanceTestCase struct {
	CheckDestroy      resource.TestCheckFunc
	VersionConstraint string
	MinVersion        *version.Version
	PreCheck          func()
	Steps             Steps
	Lock              *sync.Mutex
}

func AcceptanceTest(t *testing.T, testCase AcceptanceTestCase) {
	t.Helper()
	if len(testCase.Steps) == 0 {
		t.Fatal("missing test steps")
	}
	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			pt.PreCheck(t)
			if testCase.VersionConstraint != "" {
				PreCheckVersionConstraint(t, testCase.VersionConstraint)
			}
			if testCase.MinVersion != nil {
				PreCheckMinVersion(t, testCase.MinVersion)
			}
			if testCase.PreCheck != nil {
				testCase.PreCheck()
			}
			if testCase.Lock != nil {
				testCase.Lock.Lock()
				t.Cleanup(func() {
					testCase.Lock.Unlock()
				})
			}
		},
		ProtoV6ProviderFactories: providers,
		CheckDestroy:             testCase.CheckDestroy,
		Steps:                    testCase.Steps,
	})
}

func TestMain(m *testing.M) {
	providers = createProviders()
	os.Exit(pt.Run(m, func(env *pt.TestEnvironment) {
		testClient = env.Client
	}))
}

func createProviders() providersMap {
	return map[string]func() (tfprotov6.ProviderServer, error){
		"unifi": providerserver.NewProtocol6WithError(provider.NewV2("acctest")()),
	}
}
