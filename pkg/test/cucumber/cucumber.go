// Package cucumber allows you to use cucumber to execute Gherkin based
// BDD test scenarios with some helpful API testing step implementations.
//
// Some steps allow you store variables or use those variables.  The variables
// are scoped to the Scenario.  The http response state is stored in the users
// session.  Switching users will switch the session.  Scenarios are executed
// concurrently.  The same user can be logged into two scenarios, but each scenario
// has a different session.
//
// Note: be careful using the same user/organization across different scenarios since
// they will likely see unexpected API mutations done in the other scenarios.
//
// Using in a test
//  func TestMain(m *testing.M) {
//
//	ocmServer := mocks.NewMockConfigurableServerBuilder().Build()
//	defer ocmServer.Close()
//
//	h, _, teardown := test.RegisterIntegration(&testing.T{}, ocmServer)
//	defer teardown()
//
//	cucumber.TestMain(h)
//
//}

package cucumber

import (
	"os"
	"sync"
	"time"

	"github.com/cucumber/godog"
	"github.com/cucumber/godog/colors"
)

func NewTestSuite() *TestSuite {
	return &TestSuite{}
}

func DefaultOptions() godog.Options {
	opts := godog.Options{
		Output:      colors.Colored(os.Stdout),
		Format:      "progress",
		Paths:       []string{"features"},
		Randomize:   time.Now().UTC().UnixNano(), // randomize TestScenario execution order
		Concurrency: 10,
	}

	return opts
}

// TestSuite holds the state global to all the test scenarios.
// It is accessed concurrently from all test scenarios.
type TestSuite struct {
	Mu sync.Mutex
}

// TestScenario holds that state of single scenario.  It is not accessed
// concurrently.
type TestScenario struct {
	Suite           *TestSuite
	Variables       map[string]interface{}
	hasTestCaseLock bool
}

func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}

// StepModules is the list of functions used to add steps to a godog.ScenarioContext, you can
// add more to this list if you need test TestSuite specific steps.
var StepModules []func(ctx *godog.ScenarioContext, s *TestScenario)

func (suite *TestSuite) InitializeScenario(ctx *godog.ScenarioContext) {
	s := &TestScenario{
		Suite:     suite,
		Variables: map[string]interface{}{},
	}

	for _, module := range StepModules {
		module(ctx, s)
	}
}
