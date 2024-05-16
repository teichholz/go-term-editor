package config

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"os"

	"github.com/fsnotify/fsnotify"
)

//go:embed config.json
var config embed.FS
var confDir string
var confName string = "config.json"
var confFile string

type EditorConfig struct {
	RelativeLineNumbers bool `json:"relativeLineNumbers"`
	TrimFiles   bool   `json:"trimFiles"`
}

type Config struct {
	log *log.Logger
	watcher *fsnotify.Watcher
	EditorConfig *EditorConfig
}

func NewConfig(log *log.Logger) *Config {
	return &Config{log: log}
}

// TODO make this a component and inject things like the logger
func (cfg *Config) Init() {
	if os.Getenv("XDG_CONFIG_HOME") == "" {
		confDir = os.Getenv("HOME") + "/.goditor"
	} else {
		confDir = os.Getenv("XDG_CONFIG_HOME") + "/goditor"
	}
	confFile = confDir + "/" + confName

	cfg.writeConfigIfMissing()

	var editorConfig EditorConfig
	cfg.EditorConfig = &editorConfig
	cfg.readConfigIntoMemory()

	go cfg.rereadConfigOnFileChange()
}

func (cfg *Config) writeConfigIfMissing() {
	_, err := os.DirFS(confDir).Open("config.json")
	// write config file if it does not exist
	if err != nil {
		content, err := fs.ReadFile(config, confName)
		if err != nil {
			cfg.log.Fatalf("Could not read embedded config file: %v", err)
		}

		derr := os.Mkdir(confDir, 0755)
		if derr != nil && derr.(*os.PathError).Err.Error() != "file exists" {
			cfg.log.Fatalf("Could not create config directory: %v", derr)
		}

		ferr := os.WriteFile(confFile, content, 0664)
		if ferr != nil && ferr.(*os.PathError).Err.Error() != "file exists" {
			cfg.log.Fatalf("Could not write config file: %v", ferr)
		}
	}
}

func (cfg *Config) rereadConfigOnFileChange() {
	watcher, err := fsnotify.NewWatcher()
	cfg.watcher = watcher
	if err != nil {
		cfg.log.Fatalf("Could not create file watcher: %v", err)
	}
	defer watcher.Close()

	err = watcher.Add(confDir)
	if err != nil {
		cfg.log.Fatalf("Could not watch config file: %v", err)
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Has(fsnotify.Create) && event.Name == confFile {
				cfg.log.Printf("Config file changed, reloading")
				cfg.readConfigIntoMemory()
			}
		case err := <-watcher.Errors:
			cfg.log.Printf("Error watching config file: %v", err)
			return;
		}
	}
}

func (cfg *Config) Cleanup() {
	if cfg.watcher != nil {
		cfg.watcher.Close()
	}
}

func (cfg *Config) readConfigIntoMemory() {
	configContent, err := os.ReadFile(confFile)
	if err != nil {
		cfg.log.Fatalf("Could not read config file into memory: %v", err)
	}
	uerr := json.Unmarshal(configContent, cfg.EditorConfig)
	if uerr != nil {
		cfg.log.Fatalf("Could not unmarshal config file: %v", uerr)
	}
}