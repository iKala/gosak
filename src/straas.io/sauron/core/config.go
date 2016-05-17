package core

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	yaml "gopkg.in/yaml.v2"

	"straas.io/sauron"
)

// NewFileConfig creates a file config loader
func NewFileConfig(cfgRoot string) (sauron.Config, error) {
	// TODO: check existence of root

	return &fileConfigImpl{
		cfgRoot: cfgRoot,
		futil:   &fileUtilImpl{},
	}, nil
}

// fileConfigImpl loads configration from file directly
type fileConfigImpl struct {
	cfgRoot string
	futil   fileUtil
}

// fileUtil defines file operations for testing purpose
type fileUtil interface {
	// Read reads file content
	Read(string) ([]byte, error)
	// Exists checks if file exist
	Exist(string) bool
}

// LoadJobs loads jobs of the given env
func (c *fileConfigImpl) LoadJobs(env string) ([]sauron.JobMeta, error) {
	// TODO: what format ?!
	return nil, nil
}

// LoadCOnfig load config of the path
func (c *fileConfigImpl) LoadConfig(path string, v interface{}) error {
	path = filepath.Join(c.cfgRoot, path)
	jsonPath := path + ".json"
	yamlPath := path + ".yaml"

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
