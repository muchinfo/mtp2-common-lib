package main

import (
	"fmt"

	"github.com/muchinfo/mtp2-common-lib/config"

	"github.com/fsnotify/fsnotify"
)

type AppConfig struct {
	Name string `yaml:"name"`
	Age  int    `yaml:"age"`
}

func RunViperExample() {
	cfg := AppConfig{}
	files := []string{"config.yaml"}
	err := config.InitViper(files, &cfg, func(e fsnotify.Event) {
		fmt.Println("配置文件变更:", e.Name)
		fmt.Printf("新配置: %+v\n", cfg)
	})
	if err != nil {
		panic(err)
	}
	fmt.Printf("初始配置: %+v\n", cfg)
}
