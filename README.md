# ecs-manager

AWS ECS Manager tool written in [Go](https://golang.org).

## Disclaimer

Use this software on your onw risk. None of the authors or contributors, in any way whatsoever, can be responsible for your use of this software or for any damage it could cause.


| **IMPORTANT NOTE** |
| ------------------ |
| Starting with version 0.2.0, ecs-manager is using YAML instead of TOML configuration file format. If you used ecs-manager version prior to 0.2.0, you would need to create new config.yaml file. |


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
❯ cat > ~/.config/ecs-manager/config.yaml <<EOF
ecs:

  ### Example
  #
  # "arn:aws:ecs:us-east-1:111111111111:cluster/test-ecs-1":
  #   # Is it test ECS cluster?
  #   test_cluster: true
  #   # Wait for instances to start task before proceeding to the next one?
  #   wait_for_task: true
  #   # How many instances in cluster are allowed to have 0 tasks running?
  #   number_of_zero_task_instances: 1
  #   # Number of instances to drain and terminate at the same time
  #   drain_and_terminate_batch_size: 1
  #   # Delay in seconds before proceeding to the next instance
  #   drain_and_terminate_delay: 60

EOF
```

For each cluster create config under `ecs:` in `~/.config/ecs-manager/config.yaml`. For example:

```yaml
ecs:

  ### Example
  #
  # "arn:aws:ecs:us-east-1:111111111111:cluster/test-ecs-1":
  #   test_cluster: true
  #   wait_for_task: true

  "arn:aws:ecs:us-east-1:111111111111:cluster/test-ecs-1":
    test_cluster: false
    wait_for_task: true

  "arn:aws:ecs:us-east-1:111111111111:cluster/test-ecs-2":
    test_cluster: true
    wait_for_task: false

  "arn:aws:ecs:us-east-1:111111111111:cluster/test-ecs-3":
    test_cluster: true
    wait_for_task: true
    number_of_zero_tasks_instances: 2
    drain_and_terminate_batch_size: 2
    drain_and_terminate_delay: 60

```

When `test_cluster` is set to `true`, it means if you chose to drain instances in cluster, this tool would not wait for drain to finish, but force stop tasks one by one.

When `wait_for_task` is set to `true`, it means if you chose to drain and terminate instances in cluster, this tool would wait for a new instance to come up and start at least one task before proceeding to the next one.


## Usage

NOTE: Before running this tool, you need to [Configure AWS CLI](https://docs.aws.amazon.com/cli/latest/userguide/cli-chap-configure.html).

Run `ecs-manager` command and follow the menu.
