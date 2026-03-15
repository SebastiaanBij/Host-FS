package dto

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path"
	"time"

	"github.com/docker/go-plugins-helpers/volume"
)

type HostFSDriverState struct {
	LogLevel string  `json:"LogLevel"`
	HostDir  string  `json:"HostDir"`
	MountDir string  `json:"MountDir"`
	StateDir string  `json:"StateDir"`
	Volumes  Volumes `json:"Volumes"`
}

func NewHostFSDriverStateFromFile(stateDir string) (*HostFSDriverState, error) {
	content, err := os.ReadFile(path.Join(stateDir, "state.json"))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("could not decode state file, due to the following error: %s", err.Error()))
	}

	var state HostFSDriverState
	err = json.Unmarshal(content, &state)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("could not decode state file, due to the following error: %s", err.Error()))
	}

	return &state, nil
}

func NewHostFSDriverStateFromEnv() (*HostFSDriverState, error) {
	logLevel := os.Getenv("LOG_LEVEL")
	if logLevel == "" {
		logLevel = "info"
	}

	hostDir := os.Getenv("HOST_DIR")
	if hostDir == "" {
		hostDir = "/var/lib/host-fs/host"
	}

	mountDir := os.Getenv("MOUNT_DIR")
	if mountDir == "" {
		mountDir = "/var/lib/host-fs/mount"
	}

	stateDir := os.Getenv("STATE_DIR")
	if stateDir == "" {
		stateDir = "/var/lib/host-fs"
	}

	return &HostFSDriverState{
		LogLevel: logLevel,
		HostDir:  hostDir,
		MountDir: mountDir,
		StateDir: stateDir,
		Volumes:  make(Volumes),
	}, nil
}

func (state *HostFSDriverState) WriteHostFSDriverStateToFile() error {
	if _, err := os.Stat(path.Join(state.HostDir, state.StateDir)); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(path.Join(state.HostDir, state.StateDir), 0755); err != nil {
				return errors.New(fmt.Sprintf("could not create state directory, due to the following error: %s", err.Error()))
			}
		} else {
			return errors.New(fmt.Sprintf("could not find state directory, due to the following error: %s", err.Error()))
		}
	}

	content, err := json.Marshal(state)
	if err != nil {
		return errors.New(fmt.Sprintf("could not write state to file, due to the following error: %s", err.Error()))
	}

	if err := os.WriteFile(path.Join(state.HostDir, state.StateDir, "state.json"), content, 0600); err != nil {
		return errors.New(fmt.Sprintf("could not write state to file, due to the following error: %s", err.Error()))
	}

	return nil
}

func (state *HostFSDriverState) SlogLogLevel() (*slog.Level, error) {
	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(state.LogLevel)); err != nil {
		return nil, errors.New(fmt.Sprintf("could not decode log level, due to the following error: %s", err.Error()))
	}

	return &logLevel, nil
}

func (state *HostFSDriverState) VolumeExists(name string) bool {
	_, ok := state.Volumes[name]
	return ok
}

func (state *HostFSDriverState) NewVolume(name string, options VolumeOptions) *Volume {
	return &Volume{
		HostPath:     path.Join(state.HostDir, options.Path),
		MountPath:    path.Join(state.MountDir, name),
		CreationDate: time.Now().Format(time.RFC3339),
		Options:      options,
	}
}

func (state *HostFSDriverState) DockerVolume(name string) (*volume.Volume, error) {
	stateVolume, ok := state.Volumes[name]
	if !ok {
		return nil, errors.New(fmt.Sprintf("could not retrieve volume, due to the fact that it does not exist"))
	}

	return &volume.Volume{
		Name:       name,
		Mountpoint: stateVolume.Options.Path,
		CreatedAt:  stateVolume.CreationDate,
	}, nil
}
