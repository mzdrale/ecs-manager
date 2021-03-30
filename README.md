# ecs-manager

AWS ECS Manager tool written in [Go](https://golang.org).

## Disclaimer

Use this software on your onw risk. None of the authors or contributors, in any way whatsoever, can be responsible for your use of this software or for any damage it could cause.

## Install

Go to [Releases page](https://gitlab.com/mzdrale/ecs-manager/-/releases) and download the latest binary version for your OS (look for link `ecs-manager-<version>-<platform>-amd64.tar.gz`). Linux amd64, macOS amd64 and macOS arm64 are available at the moment. If you need to run this tool on some other platform, you can download source code and build binary.
In this README, we will use `ecs-manager-0.1.5
-macos-amd64.tar.gz` as example.

```bash
❯ wget https://gitlab.com/mzdrale/ecs-manager/-/jobs/1139578084/artifacts/raw/target/ecs-manager-0.1.5-macos-amd64.tar.gz
```

Uncompress archive:
```bash
❯ tar xzvf ecs-manager-0.1.5-macos-amd64.tar.gz
```

Copy binary to some directory in `$PATH`, for example `/usr/local/bin`:

```bash
❯ mv ecs-manager-0.1.5-macos-amd64 /usr/local/bin/ecs-manager
❯ chmod 755 /usr/local/bin/ecs-manager
```

## Configure

Create configuration directory:

```bash
❯ mkdir -p ~/.config/ecs-manager
```

Create configuration file:

```bash
❯ cat > ~/.config/ecs-manager/config.toml <<EOF
[ecs]

# When draining instances in these clusters, we'll not wait drain to
# finish, but force task stop to speed up the process
test_clusters = [
]

# Wait for task to start before draining next instance
wait_for_task_on_instance_clusters = [
]
EOF
```

If you have some test ECS cluster and you don't want to wait for instance to finish draining tasks when in DRAINING mode, just add cluster ARN into `test_clusters` list in `~/.config/ecs-manager/config.toml` configuration file. For example:

```
test_clusters = [
    "arn:aws:ecs:us-east-1:111111111111:cluster/test-ecs-1",
    "arn:aws:ecs:us-east-1:222222222222:cluster/test-ecs-2"
]
```

If you have a cluster that needs at least one task to run on instance before draining and terminating next instance ("Drain and terminate instances, one by one"), add cluster ARN into `wait_for_task_on_instance_clusters` list in `~/.config/ecs-manager/config.toml` configuration file. For example:

```
wait_for_task_on_instance_clusters = [
    "arn:aws:ecs:us-east-1:111111111111:cluster/test-ecs-1"
]
```

## Usage

NOTE: Before running this tool, you need to [Configure AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html).

Run `ecs-manager` command and follow the menu.
