package steps

import (
	"github.com/cucumber/godog"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/test/cucumber"
)

func init() {
	cucumber.StepModules = append(cucumber.StepModules, func(ctx *godog.ScenarioContext, s *cucumber.TestScenario) {
	})
}
