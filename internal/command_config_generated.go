// Code generated by "goconfig -field StdinReader io.Reader|StdoutWriter io.Writer|StderrWriter io.Writer|Dir string -option -output command_config_generated.go -configOption Option"; DO NOT EDIT.

package internal

import "io"

type ConfigItem[T any] struct {
	modified     bool
	value        T
	defaultValue T
}

func (s *ConfigItem[T]) Set(value T) {
	s.modified = true
	s.value = value
}
func (s *ConfigItem[T]) Get() T {
	if s.modified {
		return s.value
	}
	return s.defaultValue
}
func (s *ConfigItem[T]) Default() T {
	return s.defaultValue
}
func (s *ConfigItem[T]) IsModified() bool {
	return s.modified
}
func NewConfigItem[T any](defaultValue T) *ConfigItem[T] {
	return &ConfigItem[T]{
		defaultValue: defaultValue,
	}
}

type Config struct {
	StdinReader  *ConfigItem[io.Reader]
	StdoutWriter *ConfigItem[io.Writer]
	StderrWriter *ConfigItem[io.Writer]
	Dir          *ConfigItem[string]
}
type ConfigBuilder struct {
	stdinReader  io.Reader
	stdoutWriter io.Writer
	stderrWriter io.Writer
	dir          string
}

func (s *ConfigBuilder) StdinReader(v io.Reader) *ConfigBuilder {
	s.stdinReader = v
	return s
}
func (s *ConfigBuilder) StdoutWriter(v io.Writer) *ConfigBuilder {
	s.stdoutWriter = v
	return s
}
func (s *ConfigBuilder) StderrWriter(v io.Writer) *ConfigBuilder {
	s.stderrWriter = v
	return s
}
func (s *ConfigBuilder) Dir(v string) *ConfigBuilder {
	s.dir = v
	return s
}
func (s *ConfigBuilder) Build() *Config {
	return &Config{
		StdinReader:  NewConfigItem(s.stdinReader),
		StdoutWriter: NewConfigItem(s.stdoutWriter),
		StderrWriter: NewConfigItem(s.stderrWriter),
		Dir:          NewConfigItem(s.dir),
	}
}

func NewConfigBuilder() *ConfigBuilder { return &ConfigBuilder{} }
func (s *Config) Apply(opt ...Option) {
	for _, x := range opt {
		x(s)
	}
}

type Option func(*Config)

func WithStdinReader(v io.Reader) Option {
	return func(c *Config) {
		c.StdinReader.Set(v)
	}
}
func WithStdoutWriter(v io.Writer) Option {
	return func(c *Config) {
		c.StdoutWriter.Set(v)
	}
}
func WithStderrWriter(v io.Writer) Option {
	return func(c *Config) {
		c.StderrWriter.Set(v)
	}
}
func WithDir(v string) Option {
	return func(c *Config) {
		c.Dir.Set(v)
	}
}
