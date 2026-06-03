package shell

import (
	"github.com/MunifTanjim/tmux-ctrl/internal/health"
)

func Health() []health.Checker {
	return []health.Checker{
		{
			Name: "shell - completion",
			Check: func() health.Check {
				check := health.Check{
					Meta: []health.Meta{},
				}
				shellName := DetectShell()
				switch shellName {
				case "zsh":
					enabledMeta := health.Meta{
						Name: "enabled",
					}
					isEnabled, err := IsCompletionEnabled(shellName)
					if err != nil {
						enabledMeta.Passed = false
						enabledMeta.Reason = err.Error()
					} else {
						enabledMeta.Passed = isEnabled
						if isEnabled {
							enabledMeta.Value = "YES"
						} else {
							enabledMeta.Value = "NO"
						}
					}
					check.Meta = append(check.Meta, enabledMeta)

					installedMeta := health.Meta{
						Name: "installed",
					}
					isInstalled, err := IsCompletionInstalled(shellName)
					if err != nil {
						installedMeta.Passed = false
						installedMeta.Reason = err.Error()
					} else {
						installedMeta.Passed = isInstalled
						if isInstalled {
							installedMeta.Value = CompletionFilename(shellName)
						} else {
							installedMeta.Value = "NO"
						}
					}
					check.Meta = append(check.Meta, installedMeta)

					check.Passed = installedMeta.Passed && enabledMeta.Passed

					return check
				default:
					return health.Check{
						Passed: true,
						Meta:   []health.Meta{{Name: "shell", Value: "unsupported", Passed: true}},
					}
				}
			},
		},
	}
}
