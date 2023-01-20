# OpenTelemetry Collector Components

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/newrelic/opentelemetry-collector-components/blob/master/LICENSE)
[![CLA assistant](https://cla-assistant.io/readme/badge/newrelic/developer-toolkit-template-go)](https://cla-assistant.io/newrelic/opentelemetry-collector-components)

OpenTelemetry Collector components is a New Relic's repository that contains custom [Collector's components](https://opentelemetry.io/docs/collector/). Those components are either in process of being accepted by the community or removed (e.g deprecated) from upstream.

Components owners are specified in [CODEOWNERS](./github/CODEOWNERS) file.

## Community

New Relic hosts and moderates an online forum where customers can interact with New Relic employees as well as other customers to get help and share best practices. 

* [Roadmap](https://newrelic.github.io/developer-toolkit/roadmap/) - As part of the Developer Toolkit, the roadmap for this project follows the same RFC process
* [Issues or Enhancement Requests](https://github.com/newrelic/developer-toolkit-template-go/issues) - Issues and enhancement requests can be submitted in the Issues tab of this repository. Please search for and review the existing open issues before submitting a new issue.
* [Contributors Guide](CONTRIBUTING.md) - Contributions are welcome (and if you submit a Enhancement Request, expect to be invited to contribute it yourself :grin:).
* [Community discussion board](https://discuss.newrelic.com/c/build-on-new-relic/developer-toolkit) - Like all official New Relic open source projects, there's a related Community topic in the New Relic Explorers Hub.

Keep in mind that when you submit your pull request, you'll need to sign the CLA via the click-through using CLA-Assistant. If you'd like to execute our corporate CLA, or if you have any questions, please drop us an email at opensource@newrelic.com.


## Development

Before writing any code, check the [development guides](./DEVELOPMENT.md) and open an issue with the corresponding information. For current components, you can check its maintainers in the [CODEOWNERS file](./.github/CODEOWNERS).

[The nopreceiver](./receiver/nopreceiver/) is a "no operation" component of type receiver that can be used as helper package when starting a new component.

### Requirements

* Go 1.18.0+
* GNU Make
* git


### Testing

Before contributing, all linting and tests must pass. Each component must be a Go module and include the [Makefile.Common](./Makefile.Common), in its Makefile:

```make
include ../../Makefile.Common
```

All available targets can be checked with:

```
# See make helpers
$ make help
```

### Commit Messages

Using the following format for commit messages allows for auto-generation of
the [CHANGELOG](CHANGELOG.md):

#### Format:

`<type>(<scope>): <subject>`

| Type | Description | Change log? |
|------| ----------- | :---------: |
| `chore` | Maintenance type work | No |
| `docs` | Documentation Updates | Yes |
| `feat` | New Features | Yes |
| `fix`  | Bug Fixes | Yes |
| `refactor` | Code Refactoring | No |

#### Scope

This refers to what part of the code is the focus of the work.  For example:

**General:**

* `./.github` - Work related to test and build the system (linting, owners, CI/CD, etc).
* `./receiver/` - OpenTelemetry collector receivers.
* `./internal/` - Common testing packages and components import.
* `./cmd/nrotelcomponents/` - Testing collector that includes all custom components.


## Support

New Relic has open-sourced this project. This project is provided AS-IS WITHOUT WARRANTY OR SUPPORT, although you can report issues and contribute to the project here on GitHub.

_Please do not report issues with this software to New Relic Global Technical Support._


## Open Source License

This project is distributed under the [Apache 2 license](LICENSE).
