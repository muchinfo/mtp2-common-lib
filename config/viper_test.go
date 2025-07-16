package config

import (
	"os"
	"testing"

	"github.com/fsnotify/fsnotify"
)

type testConfig struct {
	Name string `yaml:"name"`
	Age  int    `yaml:"age"`
}

func writeTestYaml(filename, content string) error {
	return os.WriteFile(filename, []byte(content), 0644)
}

func TestInitViper(t *testing.T) {
	file1 := "test1.yaml"
	file2 := "test2.yaml"
	defer os.Remove(file1)
	defer os.Remove(file2)

	// The rest of this function appears to be a duplicate and can be removed.
}

func TestInitViper_Success(t *testing.T) {
	file1 := "test1.yaml"
	file2 := "test2.yaml"
	defer os.Remove(file1)
	defer os.Remove(file2)

	yaml1 := "name: Alice\nage: 20\n"
	yaml2 := "age: 30\n"
	if err := writeTestYaml(file1, yaml1); err != nil {
		t.Fatalf("failed to write test yaml1: %v", err)
	}
	if err := writeTestYaml(file2, yaml2); err != nil {
		t.Fatalf("failed to write test yaml2: %v", err)
	}

	var cfg testConfig
	err := InitViper([]string{file1, file2}, &cfg, func(e fsnotify.Event) {
		// callback for config change
	})
	if err != nil {
		t.Fatalf("InitViper failed: %v", err)
	}
	if cfg.Name != "Alice" || cfg.Age != 30 {
		t.Errorf("unexpected config: %+v", cfg)
	}

	// 配置变更回调的测试依赖于文件系统事件，通常集成测试中验证
	// 这里只验证回调函数能被正常设置
}

func TestInitViper_InvalidFile(t *testing.T) {
	var cfg testConfig
	err := InitViper([]string{"not_exist.yaml"}, &cfg, func(e fsnotify.Event) {})
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

func TestInitViper_UnmarshalError(t *testing.T) {
	file := "bad.yaml"
	defer os.Remove(file)
	if err := writeTestYaml(file, "name: Alice\nage: not_a_number\n"); err != nil {
		t.Fatalf("failed to write test yaml: %v", err)
	}
	var cfg testConfig
	err := InitViper([]string{file}, &cfg, func(e fsnotify.Event) {})
	if err == nil {
		t.Error("expected unmarshal error, got nil")
	}
}
