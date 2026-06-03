package config

import (
	"github.com/MunifTanjim/tmux-ctrl/internal/health"
	"github.com/MunifTanjim/tmux-ctrl/internal/util"
)

func Health() []health.Checker {
	return []health.Checker{
		{
			Name: "config",
			Check: func() health.Check {
				configFilePath := GetConfigPath(ConfigFileName)
				if exists, err := util.FileExists(configFilePath); err != nil || !exists {
					return health.Check{Passed: false, Reason: "not found"}
				}
				return health.Check{
					Passed: true,
					Meta:   []health.Meta{{Name: "path", Value: configFilePath, Passed: true}},
				}
			},
		},
	}
}
