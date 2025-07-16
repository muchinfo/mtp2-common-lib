package config

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// ConfigMapError represents an error when reading a config map from a file.
type ConfigMapError struct {
	File string
}

func (e *ConfigMapError) Error() string {
	return fmt.Sprintf("failed to read config map from file: %s", e.File)
}

// cfg 必须是带有 mapstructure 标签字段的结构体指针（例如：`mapstructure:"field"`）。
// 如果 cfg 不是指针，Unmarshal 会失败；请确保传递的是配置结构体的指针。
func InitViper(configFiles []string, cfg any, onConfigChange func(e fsnotify.Event)) error {
	var err error

	v := viper.New()
	v.SetConfigType("yaml")

	// 依次加载并合并所有配置文件
	for i, file := range configFiles {
		if i == 0 {
			v.SetConfigFile(file)
			err = v.ReadInConfig()
		} else {
			configMap := readConfigMap(file)
			if configMap == nil {
				err = &ConfigMapError{File: file}
				return err
			}
			err = v.MergeConfigMap(configMap)
		}
		if err != nil {
			return err
		}
	}

	// 如果 Unmarshal 失败（如配置结构体与配置文件不匹配），直接返回错误，调用方需处理该错误
	if err = v.Unmarshal(cfg); err != nil {
		return err
	}

	// 侦听配置文件变更（仅监听第一个配置文件）
	v.WatchConfig()
	v.OnConfigChange(func(e fsnotify.Event) {
		// 配置文件发生变更之后会调用的回调函数

		var callbackErr error

		// 依次加载并合并所有配置文件
		for i, file := range configFiles {
			if i == 0 {
				v.SetConfigFile(file)
				callbackErr = v.ReadInConfig()
				if callbackErr != nil {
					fmt.Printf("failed to read config in callback: %v\n", callbackErr)
					return
				}
			} else {
				configMap := readConfigMap(file)
				if configMap == nil {
					callbackErr = &ConfigMapError{File: file}
					fmt.Printf("failed to read config map in callback: %v\n", callbackErr)
					return
				}
				// Only merge if configMap is not nil
				callbackErr = v.MergeConfigMap(configMap)
				if callbackErr != nil {
					fmt.Printf("failed to merge config map in callback: %v\n", callbackErr)
					return
				}
			}
		}

		if callbackErr = v.Unmarshal(cfg); callbackErr != nil {
			fmt.Printf("failed to unmarshal config in callback: %v\n", callbackErr)
			return
		}
		onConfigChange(e)
	})

	return nil
}

// 辅助函数：读取单个yaml配置文件为map[string]interface{}
func readConfigMap(filename string) map[string]interface{} {
	v := viper.New()
	v.SetConfigFile(filename)
	v.SetConfigType("yaml")
	if err := v.ReadInConfig(); err != nil {
		return nil
	}
	return v.AllSettings()
}
