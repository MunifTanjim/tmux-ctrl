package config

import (
	"os"
	"strings"
	"time"

	"github.com/MunifTanjim/tmux-ctrl/internal/util"
	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
	"go.yaml.in/yaml/v3"
)

// KnownKeys lists config keys offered for shell completion.
var KnownKeys = []string{
	"extract.patterns",
	"extract.hint_alphabet",
	"session.hidden_name_pattern",
}

func Get[T any](key string) T {
	var zero T
	switch any(zero).(type) {
	case string:
		return any(viper.GetString(key)).(T)
	case int:
		return any(viper.GetInt(key)).(T)
	case int32:
		return any(viper.GetInt32(key)).(T)
	case int64:
		return any(viper.GetInt64(key)).(T)
	case uint:
		return any(viper.GetUint(key)).(T)
	case uint16:
		return any(viper.GetUint16(key)).(T)
	case uint32:
		return any(viper.GetUint32(key)).(T)
	case uint64:
		return any(viper.GetUint64(key)).(T)
	case float64:
		return any(viper.GetFloat64(key)).(T)
	case bool:
		return any(viper.GetBool(key)).(T)
	case time.Time:
		return any(viper.GetTime(key)).(T)
	case time.Duration:
		return any(viper.GetDuration(key)).(T)
	case []string:
		return any(viper.GetStringSlice(key)).(T)
	case []int:
		return any(viper.GetIntSlice(key)).(T)
	case map[string]any:
		return any(viper.GetStringMap(key)).(T)
	case map[string]string:
		return any(viper.GetStringMapString(key)).(T)
	case map[string][]string:
		return any(viper.GetStringMapStringSlice(key)).(T)
	default:
		var value T
		if err := viper.UnmarshalKey(key, &value, viper.DecodeHook(mapstructure.TextUnmarshallerHookFunc())); err != nil {
			return zero // never return a partially-decoded value
		}
		return value
	}
}

func Has(key string) bool {
	return viper.InConfig(key)
}

func Set(key, value string) error {
	configFile := GetConfigPath(ConfigFileName)
	if err := util.EnsureFileExists(configFile); err != nil {
		return err
	}

	config := make(map[string]any)

	if content, err := os.ReadFile(configFile); err != nil {
		return err
	} else if err := yaml.Unmarshal(content, &config); err != nil {
		return err
	}

	util.SetNestedMapValue(config, key, value)

	var buf strings.Builder
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(config); err != nil {
		return err
	}
	if err := encoder.Close(); err != nil {
		return err
	}

	if err := os.WriteFile(configFile, []byte(buf.String()), 0644); err != nil {
		return err
	}

	return nil
}

func SetDefault(key, value string) error {
	if Has(key) {
		return nil
	}
	return Set(key, value)
}
