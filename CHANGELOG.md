# ecs-manager change log

This file is used to list changes made in each version of ecs-manager.

## 0.2.1 (December 14 2021)

- Get memory and CPU usage ([@mzdrale](https://gitlab.com/mzdrale))

## 0.2.0 (November 17 2021)

- [Breaking Change] Change configuration file format from TOML to YAML ([@mzdrale](https://gitlab.com/mzdrale) - [issue #13](https://gitlab.com/mzdrale/ecs-manager/-/issues/13))

## 0.1.5 (March 30 2021)

- Check if instances is terminated by checking it's state, rather than wait for number if instances in cluster to decrease by one ([@mzdrale](https://gitlab.com/mzdrale) - [issue #12](https://gitlab.com/mzdrale/ecs-manager/-/issues/12))
- Warn if "excluded list" is not empty ([@mzdrale](https://gitlab.com/mzdrale) - [issue #11](https://gitlab.com/mzdrale/ecs-manager/-/issues/11))
- Add option to wait for at least one task to start on a new instance before proceeding to the next one when doing "Drain and terminate instance, one by one" ([@mzdrale](https://gitlab.com/mzdrale) - [issue #10](https://gitlab.com/mzdrale/ecs-manager/-/issues/10))
- Rename `darwin` platform to `macos` ([@mzdrale](https://gitlab.com/mzdrale) - [issue #9](https://gitlab.com/mzdrale/ecs-manager/-/issues/9))
- Add support for macOS on Apple Silicon ([@mzdrale](https://gitlab.com/mzdrale) - [issue #8](https://gitlab.com/mzdrale/ecs-manager/-/issues/8))
- Minor fixes

## 0.1.4 (July 09 2020)

- Fix issue with terminating instance ([@mzdrale](https://gitlab.com/mzdrale))

## 0.1.3 (July 03 2020)

- Return to menu if selected `N` when prompted ([@mzdrale](https://gitlab.com/mzdrale) - [issue #7](https://gitlab.com/mzdrale/ecs-manager/-/issues/7))
- If draining instance fails, stop there ([@mzdrale](https://gitlab.com/mzdrale) - [issue #6](https://gitlab.com/mzdrale/ecs-manager/-/issues/6))
- Update `GetClusterInfo` function to return results for multiple clusters ([@mzdrale](https://gitlab.com/mzdrale) - [issue #5](https://gitlab.com/mzdrale/ecs-manager/-/issues/5))
- Fix output format of error messages ([@mzdrale](https://gitlab.com/mzdrale) - [issue #4](https://gitlab.com/mzdrale/ecs-manager/-/issues/4))
- Print cluster info when browsing through cluster list ([@mzdrale](https://gitlab.com/mzdrale) - [issue #3](https://gitlab.com/mzdrale/ecs-manager/-/issues/3))
- Add `version` argument ([@mzdrale](https://gitlab.com/mzdrale))
- Various text output fixes ([@mzdrale](https://gitlab.com/mzdrale))
- Code cleanup ([@mzdrale](https://gitlab.com/mzdrale))

## 0.1.2 (June 22 2020)

- Add LICENSE ([@mzdrale](https://gitlab.com/mzdrale))
- Fix some text output ([@mzdrale](https://gitlab.com/mzdrale))

## 0.1.1 (June 19 2020)

- Fix some text output format and color ([@mzdrale](https://gitlab.com/mzdrale))
- Update `.gitlab-ci.yml` and `Makefile` - Add version number to binary name ([@mzdrale](https://gitlab.com/mzdrale))
- Update `README.md` ([@mzdrale](https://gitlab.com/mzdrale))
- Code cleanup ([@mzdrale](https://gitlab.com/mzdrale))

## 0.1.0 (June 18 2020)

- Minor code cleanup ([@mzdrale](https://gitlab.com/mzdrale))

## 0.1.0-pre2 (June 17 2020)

- Fix duration time format ([@mzdrale](https://gitlab.com/mzdrale))

## 0.1.0-pre1 (June 16 2020)

- Initial prerelease ([@mzdrale](https://gitlab.com/mzdrale))
