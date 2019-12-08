package appdeployers

import (
	"strings"
)

const DefaultStrategie = "default"

type Deployer struct {
	strategies []Strategy
}

func NewDeployer(strategies ...Strategy) *Deployer {
	return &Deployer{
		strategies: strategies,
	}
}

func (d Deployer) Strategy(strategyName string) Strategy {
	strategyName = strings.ToLower(strategyName)
	var defaultStrategy Strategy
	for _, strategy := range d.strategies {
		for _, name := range strategy.Names() {
			if name == strategyName {
				return strategy
			}
			if name == DefaultStrategie {
				defaultStrategy = strategy
			}
		}
	}
	return defaultStrategy
}

func ValidStrategy(strategyName string) ([]string, bool) {
	strategyName = strings.ToLower(strategyName)
	names := append(Standard{}.Names(), BlueGreenV2{}.Names()...)
	for _, name := range names {
		if name == strategyName {
			return names, true
		}
	}
	return names, false
}
