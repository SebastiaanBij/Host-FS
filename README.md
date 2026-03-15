# Host FS

Host FS is a Docker Volume Driver inspired by [local-persist](https://github.com/MatchbookLab/local-persist). It allows volumes to be created directly on the host filesystem, automatically creating directories with the correct permissions and ownership — eliminating the need to set these up manually before creating a volume.

> **Note**
> * This project is in early development and has no full test suite. Use it at your own risk. Issues are welcome and will be addressed on a best-effort basis.
> * Currently only `amd64` is supported. `arm64` support is planned for a future release.

## Table of Contents

- [Installation](#installation)
- [Configuration](#configuration)
- [Usage](#usage)
    - [CLI](#cli)
    - [Docker Compose](#docker-compose)

---

## Installation

```shell
curl -fsSL https://raw.githubusercontent.com/SebastiaanBij/Host-FS/master/scripts/install.sh | sh
```

---

## Configuration

The plugin behaviour can be customised using the following environment variables:

| Variable    | Description                                                      | Default                  |
|-------------|------------------------------------------------------------------|--------------------------|
| `LOG_LEVEL` | Log level (`debug`, `info`, `warn`, `error`)                     | `info`                   |
| `HOST_DIR`  | Host filesystem mount path inside the plugin                     | `/var/lib/host-fs/host`  |
| `MOUNT_DIR` | Docker propagation mount path inside the plugin                  | `/var/lib/host-fs/mount` |
| `STATE_DIR` | State directory inside the plugin                                | `/var/lib/host-fs`       |

Variables can be set at install time but require manual building and installation at this point in time.

> **Warning**
> Changing `HOST_DIR` or `MOUNT_DIR` requires a custom build of the plugin with corresponding modifications to `config.json`.

---

## Usage

Volumes accept the following options:

| Option | Description                                  | Required | Default |
|--------|----------------------------------------------|----------|---------|
| `path` | Path on the host filesystem to mount         | Yes      | -       |
| `perm` | Permissions of the directory                 | No       | `0755`  |
| `uid`  | User ID to assign as owner of the directory  | No       | `0`     |
| `gid`  | Group ID to assign as owner of the directory | No       | `0`     |

### CLI

```shell
docker volume create \
  --driver host-fs \
  --opt path=/path/on/host \
  --opt perm=0755 \
  --opt uid=1000 \
  --opt gid=1000 \
  myvolume
```

### Docker Compose

```yaml
services:
  myapp:
    image: myapp
    volumes:
      - myvolume:/data

volumes:
  myvolume:
    driver: host-fs
    driver_opts:
      path: /path/on/host
      perm: "0755"
      uid: "1000"
      gid: "1000"
```