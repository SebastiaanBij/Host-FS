package dto

import (
	"errors"
	"fmt"
	"strconv"
)

type VolumeOptions struct {
	Path string
	Perm int
	UID  int
	GID  int
}

func NewVolumeOptionsFromStringMap(options map[string]string) (*VolumeOptions, error) {
	volumePath := options["path"]
	if volumePath == "" {
		return nil, errors.New("missing required option: 'path'")
	}

	volumePerm := options["perm"]
	if volumePerm == "" {
		volumePerm = "0755"
	}
	volumePermInt, err := strconv.Atoi(volumePerm)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("could not decode option: 'perm', due to the following error: %s", err.Error()))
	}

	volumeUID := options["uid"]
	if volumeUID == "" {
		volumeUID = "0"
	}
	volumeUIDInt, err := strconv.Atoi(volumeUID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("could not decode option: 'uid', due to the following error: %s", err.Error()))
	}

	volumeGID := options["gid"]
	if volumeGID == "" {
		volumeGID = "0"
	}
	volumeGIDInt, err := strconv.Atoi(volumeGID)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("could not decode option: 'gid', due to the following error: %s", err.Error()))
	}

	return &VolumeOptions{
		Path: volumePath,
		Perm: volumePermInt,
		UID:  volumeUIDInt,
		GID:  volumeGIDInt,
	}, nil
}

type Volume struct {
	HostPath     string        `json:"HostPath"`
	MountPath    string        `json:"MountPath"`
	CreationDate string        `json:"CreationDate"`
	Options      VolumeOptions `json:"Options"`
}

type Volumes = map[string]Volume
