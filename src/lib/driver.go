package lib

import (
	"errors"
	"fmt"
	"host-fs/src/dto"
	"log/slog"
	"os"
	"sync"

	"github.com/docker/go-plugins-helpers/volume"
	"golang.org/x/sys/unix"
)

type HostFSDriver struct {
	State *dto.HostFSDriverState
	Mutex *sync.RWMutex
}

func NewHostFSDriver(stateDir string) (*HostFSDriver, error) {
	state, err := dto.NewHostFSDriverStateFromFile(stateDir)
	if err != nil {
		state, err = dto.NewHostFSDriverStateFromEnv()
		if err != nil {
			return nil, err
		}
	}

	if level, err := state.SlogLogLevel(); err == nil {
		slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: level,
		})))
	} else {
		return nil, errors.New(fmt.Sprintf("could not instantiate logger, due to the following error: %s", err.Error()))
	}

	slog.Info("instantiated the HostFS Driver")

	return &HostFSDriver{
		State: state,
		Mutex: &sync.RWMutex{},
	}, nil
}

func (driver *HostFSDriver) Get(request *volume.GetRequest) (*volume.GetResponse, error) {
	slog.Info("retrieving volume", "name", request.Name)

	slog.Debug("claiming read lock", "name", request.Name)
	driver.Mutex.RLock()
	defer driver.Mutex.RUnlock()
	slog.Debug("claimed read lock", "name", request.Name)

	dockerVolume, err := driver.State.DockerVolume(request.Name)
	if err != nil {
		slog.Error("could not retrieve volume", "name", request.Name, "error", err.Error())
		slog.Debug("released read lock", "name", request.Name)
		return nil, err
	}

	slog.Info("retrieved volume", "name", request.Name)
	slog.Debug("released read lock", "name", request.Name)
	return &volume.GetResponse{
		Volume: dockerVolume,
	}, nil
}

func (driver *HostFSDriver) List() (*volume.ListResponse, error) {
	slog.Info("retrieving all volumes")

	slog.Debug("claiming read lock")
	driver.Mutex.RLock()
	defer driver.Mutex.RUnlock()
	slog.Debug("claimed read lock")

	var dockerVolumes []*volume.Volume
	for name := range driver.State.Volumes {
		dockerVolume, err := driver.State.DockerVolume(name)
		if err != nil {
			slog.Error("could not retrieve all volumes", "error", err.Error())
			slog.Debug("released read lock")
			return nil, err
		}

		dockerVolumes = append(dockerVolumes, dockerVolume)
	}

	slog.Info("retrieved all volumes")
	slog.Debug("released read lock")
	return &volume.ListResponse{
		Volumes: dockerVolumes,
	}, nil
}

func (driver *HostFSDriver) Create(request *volume.CreateRequest) error {
	slog.Info("creating volume", "name", request.Name, "options", request.Options)

	slog.Debug("decoding volume options", "name", request.Name, "options", request.Options)
	volumeOptions, err := dto.NewVolumeOptionsFromStringMap(request.Options)
	if err != nil {
		slog.Error("could not create volume", "name", request.Name, "options", request.Options, "error", err.Error())
		slog.Debug("released write lock", "name", request.Name, "options", request.Options)
		return err
	}
	stateVolume := driver.State.NewVolume(request.Name, *volumeOptions)
	slog.Debug("decoded volume options", "name", request.Name, "options", request.Options)

	slog.Debug("claiming write lock", "name", request.Name, "options", request.Options)
	driver.Mutex.Lock()
	defer driver.Mutex.Unlock()
	slog.Debug("claimed write lock", "name", request.Name, "options", request.Options)

	slog.Debug("checking volume existence", "name", request.Name, "options", request.Options)
	if driver.State.VolumeExists(request.Name) {
		slog.Error("could not create volume, due to the fact that it already exists", "name", request.Name, "options", request.Options)
		slog.Debug("released write lock", "name", request.Name, "options", request.Options)
		return errors.New("could not retrieve volume, due to the fact that it already exists")
	}
	slog.Debug("checked volume existence", "name", request.Name, "options", request.Options)

	slog.Debug("creating volume host directories with corresponding permissions", "name", request.Name, "options", request.Options)
	if err := os.MkdirAll(stateVolume.HostPath, os.FileMode(stateVolume.Options.Perm)); err != nil {
		slog.Error("could not create volume", "name", request.Name, "options", request.Options, "error", err.Error())
		slog.Debug("released write lock", "name", request.Name, "options", request.Options)
		return err
	}
	slog.Debug("created volume host directories with corresponding permissions", "name", request.Name, "options", request.Options)

	slog.Debug("assigning volume host directories ownership", "name", request.Name, "options", request.Options)
	if err := os.Chown(stateVolume.HostPath, stateVolume.Options.UID, stateVolume.Options.GID); err != nil {
		slog.Error("could not create volume", "name", request.Name, "options", request.Options, "error", err.Error())
		slog.Debug("released write lock", "name", request.Name, "options", request.Options)
		return err
	}
	slog.Debug("assigned volume host directories ownership", "name", request.Name, "options", request.Options)

	slog.Debug("saving state", "name", request.Name, "options", request.Options)
	driver.State.Volumes[request.Name] = *stateVolume
	if err := driver.State.WriteHostFSDriverStateToFile(); err != nil {
		slog.Error("could not create volume", "name", request.Name, "options", request.Options, "error", err.Error())
		slog.Debug("released write lock", "name", request.Name, "options", request.Options)
		return err
	}
	slog.Debug("saved state", "name", request.Name, "options", request.Options)

	slog.Info("created volume", "name", request.Name, "options", request.Options)
	slog.Debug("released write lock", "name", request.Name, "options", request.Options)
	return nil
}

func (driver *HostFSDriver) Remove(request *volume.RemoveRequest) error {
	slog.Info("removing volume", "name", request.Name)

	slog.Debug("claiming read lock", "name", request.Name)
	driver.Mutex.RLock()
	slog.Debug("claimed read lock", "name", request.Name)

	if !driver.State.VolumeExists(request.Name) {
		driver.Mutex.RUnlock()
		slog.Error("could not remove volume, due to the fact that it does not exist", "name", request.Name)
		slog.Debug("released read lock", "name", request.Name)
		return errors.New("could not remove volume, due to the fact that it does not exist")
	}
	driver.Mutex.RUnlock()
	slog.Debug("released read lock", "name", request.Name)

	slog.Debug("claiming write lock", "name", request.Name)
	driver.Mutex.Lock()
	defer driver.Mutex.Unlock()
	slog.Debug("claimed write lock", "name", request.Name)

	slog.Debug("saving state", "name", request.Name)
	delete(driver.State.Volumes, request.Name)
	if err := driver.State.WriteHostFSDriverStateToFile(); err != nil {
		slog.Error("could not remove volume", "name", request.Name, "error", err.Error())
		slog.Debug("released write lock", "name", request.Name)
		return err
	}
	slog.Debug("saved state", "name", request.Name)

	slog.Info("removed volume", "name", request.Name)
	slog.Debug("released write lock", "name", request.Name)
	return nil
}

func (driver *HostFSDriver) Path(request *volume.PathRequest) (*volume.PathResponse, error) {
	slog.Info("retrieving volume path", "name", request.Name)

	slog.Debug("claiming read lock", "name", request.Name)
	driver.Mutex.RLock()
	defer driver.Mutex.RUnlock()
	slog.Debug("claimed read lock", "name", request.Name)

	dockerVolume, err := driver.State.DockerVolume(request.Name)
	if err != nil {
		slog.Error("could not retrieve volume path", "name", request.Name, "error", err.Error())
		slog.Debug("released read lock", "name", request.Name)
		return nil, err
	}

	slog.Info("retrieved volume path", "name", request.Name)
	slog.Debug("released read lock", "name", request.Name)
	return &volume.PathResponse{
		Mountpoint: dockerVolume.Mountpoint,
	}, nil
}

func (driver *HostFSDriver) Mount(request *volume.MountRequest) (*volume.MountResponse, error) {
	slog.Info("mounting volume", "id", request.ID, "name", request.Name)

	slog.Debug("claiming read lock", "id", request.ID, "name", request.Name)
	driver.Mutex.RLock()
	defer driver.Mutex.RUnlock()
	slog.Debug("claimed read lock", "id", request.ID, "name", request.Name)

	stateVolume, ok := driver.State.Volumes[request.Name]
	if !ok {
		slog.Error("could not mount volume, due to the fact that it does not exist", "id", request.ID, "name", request.Name)
		slog.Debug("released read lock", "id", request.ID, "name", request.Name)
		return nil, errors.New("could not mount volume, due to the fact that it does not exist")
	}

	slog.Debug("creating volume mount directories with corresponding permissions", request.ID, "name", request.Name)
	if err := os.MkdirAll(stateVolume.MountPath, os.FileMode(0755)); err != nil {
		slog.Error("could not mount volume", request.ID, "name", request.Name, "error", err.Error())
		slog.Debug("released read lock", "id", request.ID, "name", request.Name)
		return nil, err
	}
	slog.Debug("created volume mount directories with corresponding permissions", request.ID, "name", request.Name)

	slog.Debug("binding volume mount directories", request.ID, "name", request.Name)
	if err := unix.Mount(stateVolume.HostPath, stateVolume.MountPath, "", unix.MS_BIND, ""); err != nil {
		slog.Error("could not mount volume", request.ID, "name", request.Name, "error", err.Error())
		slog.Debug("released read lock", "id", request.ID, "name", request.Name)
		return nil, err
	}
	slog.Debug("bound volume mount directories", request.ID, "name", request.Name)

	slog.Info("mounted volume", "id", request.ID, "name", request.Name)
	slog.Debug("released read lock", "id", request.ID, "name", request.Name)
	return &volume.MountResponse{
		Mountpoint: stateVolume.MountPath,
	}, nil
}

func (driver *HostFSDriver) Unmount(request *volume.UnmountRequest) error {
	slog.Info("unmounting volume", "id", request.ID, "name", request.Name)

	slog.Debug("claiming read lock", "id", request.ID, "name", request.Name)
	driver.Mutex.RLock()
	defer driver.Mutex.RUnlock()
	slog.Debug("claimed read lock", "id", request.ID, "name", request.Name)

	stateVolume, ok := driver.State.Volumes[request.Name]
	if !ok {
		slog.Error("could not unmount volume, due to the fact that it does not exist", "id", request.ID, "name", request.Name)
		slog.Debug("released read lock", "id", request.ID, "name", request.Name)
		return errors.New("could not unmount volume, due to the fact that it does not exist")
	}

	slog.Debug("unbinding volume mount directories", request.ID, "name", request.Name)
	if err := unix.Unmount(stateVolume.MountPath, unix.MNT_DETACH); err != nil {
		slog.Error("could not unmount volume", request.ID, "name", request.Name, "error", err.Error())
		slog.Debug("released read lock", "id", request.ID, "name", request.Name)
		return err
	}
	slog.Debug("unbound volume mount directories", request.ID, "name", request.Name)

	slog.Info("unmounted volume", "id", request.ID, "name", request.Name)
	slog.Debug("released read lock", "id", request.ID, "name", request.Name)
	return nil
}

func (driver *HostFSDriver) Capabilities() *volume.CapabilitiesResponse {
	slog.Info("retrieving volume capabilities")
	slog.Info("retrieved volume capabilities")
	return &volume.CapabilitiesResponse{
		Capabilities: volume.Capability{
			Scope: "local",
		},
	}
}
