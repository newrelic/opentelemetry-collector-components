# Development guides

## Adding new components

**Before** any code is written, open an issue providing the following information:

- Who will be the owner for your component. A owner must be a New Relic's team, it will
be the maintainer of the corresponding component.
- Some information about your component, such the reasoning behind it.
- Links refereing to upstream issues about your component. Note that the components in
this repository should be **temporal**, all components should be pushed upstream.

Components comprise of exporters, extensions, receivers, and processors. The key criteria to implementing a component is to:

* Implement the [component.Component](https://pkg.go.dev/go.opentelemetry.io/collector/component#Component) interface
* Provide a configuration structure which defines the configuration of the component
* Provide the implementation which performs the component operation


Generally, maintenance of components is the responsibility of its owners.

- Create your component under the proper folder and use Go standard package naming recommendations.
- Use a boiler-plate Makefile that just references the one at top level, ie.: `include ../../Makefile.Common` - this
  allows you to use the main helper functions during development.
- Each component has its own go.mod file. This allows custom builds of the collector to take a limited sets of
  dependencies - so run `go mod` commands as appropriate for your component.
- Implement the needed interface on your component by importing the appropriate component from the core repo. Follow the
  pattern of existing components regarding config and factory source files and tests.
- Implement your component as appropriate. Target is to get 80% or more of code coverage.
- Add a README.md on the root of your component describing its configuration and usage, likely referencing some of the
  yaml files used in the component tests. We also suggest that the yaml files used in tests have comments for all
  available configuration settings so users can copy and modify them as needed.
- Add a `replace` directive at the root `go.mod` file so your component is included in the build of the contrib
  executable.
- All components must be included in [`internal/components/`](./internal/components) and in the respective testing
  harnesses. To align with the test goal of the project, components must be testable within the framework defined within
  the folder. If a component can not be properly tested within the existing framework, it must increase the non testable
  components number with a comment within the PR explaining as to why it can not be tested.
- Add the team for your component to a new line for your component in the [`.github/CODEOWNERS`](./.github/CODEOWNERS) file.


## Testing

Different teams, shared responsibility. The repository contains a [common testing pipeline](./.github/workflows/build-and-test.yml)
that will run and **must succeed** for all developed components. The following checks will run for all components:

- Code owners check.
- Linter: `make golint`
- Documentation and best practises: checks that the corresponding component contains documentation, licenses, valid links, etc.
- Unit tests: `make gotest`
- Cross compile: verifies that a collector can be build with the current components for different architectures, thus the component 
must be included in [./internal/components/components.go](./internal/components/components.go).
