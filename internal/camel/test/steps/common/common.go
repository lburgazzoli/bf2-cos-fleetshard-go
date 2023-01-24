package common

import (
	"github.com/cucumber/godog"
	"github.com/rs/xid"
	"gitub.com/lburgazzoli/bf2-cos-fleetshard-go/pkg/test/cucumber"
)

func init() {
	cucumber.StepModules = append(cucumber.StepModules, func(ctx *godog.ScenarioContext, s *cucumber.TestScenario) {
		ctx.Step(`^I store an UID as \${([^"]*)}$`, func(as string) error {
			s.Variables[as] = xid.New().String()
			return nil
		})
	})
}
