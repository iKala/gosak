package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	yaml "gopkg.in/yaml.v2"

	"straas.io/sauron"
)

const (
	defaultInterval = time.Minute
	yamlExtension   = ".yaml"
	jsonExtension   = ".json"
)

// NewFileConfig creates a file config loader
func NewFileConfig(cfgRoot string, dryRun bool) (sauron.Config, error) {
	futil := &fileUtilImpl{}
	if !futil.Exist(cfgRoot) {
		return nil, fmt.Errorf("config root %s does not exist", cfgRoot)
	}
	return &fileConfigImpl{
		cfgRoot: cfgRoot,
		dryRun:  dryRun,
		futil:   futil,
	}, nil
}

// fileConfigImpl loads configration from file directly
type fileConfigImpl struct {
	cfgRoot string
	dryRun  bool
	futil   fileUtil
}

// fileUtil defines file operations for testing purpose
type fileUtil interface {
	// Read reads file content
	Read(string) ([]byte, error)
	// Exists checks if file exist
	Exist(string) bool
	// Walk visits the path
	Walk(string, filepath.WalkFunc) error
}

// LoadJobs loads jobs of the given env
func (c *fileConfigImpl) LoadJobs(envs ...string) ([]sauron.JobMeta, error) {
	result := []sauron.JobMeta{}
	for _, env := range envs {
		subResult, err := c.loadJobs(env)
		if err != nil {
			return nil, err
		}
		result = append(result, subResult...)
	}
	return result, nil
}

func (c *fileConfigImpl) loadJobs(env string) ([]sauron.JobMeta, error) {
	alertRoot := filepath.Join(c.cfgRoot, "alert")
	scanRoot := filepath.Join(alertRoot, env)
	result := []sauron.JobMeta{}

	if !c.futil.Exist(scanRoot) {
		return nil, fmt.Errorf("path %s does not exist", scanRoot)
	}
	c.futil.Walk(scanRoot, func(path string, info os.FileInfo, err error) error {
		// only care about yaml files
		if err != nil {
			return err
		}
		if info.IsDir() || !strings.HasSuffix(path, yamlExtension) {
			return nil
		}
		data, err := c.futil.Read(path)
		if err != nil {
			return fmt.Errorf("fail to read file %s, err:%v", path, err)
		}
		meta := sauron.JobMeta{}
		if err := yaml.Unmarshal(data, &meta); err != nil {
			return fmt.Errorf("fail to parse file %s, err:%v", path, err)
		}

		// assign attributes
		meta.DryRun = c.dryRun
		meta.Env = env
		meta.JobID = toJobID(alertRoot, path)
		// use default interval
		if meta.Interval == 0 {
			meta.Interval = defaultInterval
		}

		result = append(result, meta)
		return nil
	})
	return result, nil
}

// LoadCOnfig load config of the path
func (c *fileConfigImpl) LoadConfig(path string, v interface{}) error {
	path = filepath.Join(c.cfgRoot, path)
	jsonPath := path + jsonExtension
	yamlPath := path + yamlExtension

	// try parse yaml first then json
	if c.futil.Exist(yamlPath) {
		bs, err := c.futil.Read(yamlPath)
		if err != nil {
			return err
		}
		return yaml.Unmarshal(bs, v)
	}

	if c.futil.Exist(jsonPath) {
		bs, err := c.futil.Read(jsonPath)
		if err != nil {
			return err
		}
		return json.Unmarshal(bs, v)
	}

	// load configuration
	return fmt.Errorf("unable to find config file for %s", path)
}

func (c *fileConfigImpl) AddChangeListener(func()) {
	// file config never fire this
}

func toJobID(base, path string) string {
	path, _ = filepath.Rel(base, path)
	return strings.TrimSuffix(path, yamlExtension)
}

type fileUtilImpl struct{}

func (*fileUtilImpl) Read(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

func (*fileUtilImpl) Exist(path string) bool {
	if _, err := os.Stat(path); err == nil {
		return true
	}
	return false
}

func (*fileUtilImpl) Walk(path string, walker filepath.WalkFunc) error {
	return filepath.Walk(path, walker)
}
