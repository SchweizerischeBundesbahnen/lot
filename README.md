# Lightweight Operator Toolbox (LOT)

#### Table Of Contents

- [Introduction](#introduction)
- [Getting Started](#getting-started)
- [Documentation](#documentation)
- [Contributing](#contributing)
- [Code of Conduct](#code-of-conduct)
- [Coding Standards](#coding-standards)
- [License](#license)

## <a id="introduction"></a> Introduction

LOT is collection of utilities for building Kubernetes Operator without the burden of excessive boilerplate.

LOT is a Go module designed to simplify Kubernetes operator development by abstracting common tasks, providing a simple API, and offering compatibility with underlying libraries.

With LOT, you can rapidly prototype and build robust operators with minimal boilerplate code.

Key Features
- **Abstraction of Common Tasks**: LOT reduces boilerplate code by abstracting away common tasks involved in Kubernetes operator development. This allows you to focus on the core logic of your operator without getting lost in repetitive implementation details.
- **Simple and Powerful API**: LOT provides a simple yet powerful API that encapsulates the most common use cases encountered in Kubernetes operator development. You can easily interact with Kubernetes resources, manage lifecycle events, handle reconciliations, and more, all with intuitive function calls.
- **Compatibility with Underlying Libraries**: LOT is built to be compatible with the underlying libraries used in Kubernetes operator development.
- **Fast Prototyping**: LOT is designed to accelerate your development process, enabling fast prototyping of Kubernetes operators. Its intuitive API and reduced boilerplate allow you to quickly build and iterate on your operator logic, helping you bring your ideas to life in no time.

## <a id="getting-started"></a> Getting Started
To use LOT in your Go project, simply download the package:

```shell
go get github/SchweizerischeBundesbahnen/lot@<VERSION>
```

and import it into your code:

```go
import "github.com/SchweizerischeBundesbahnen/lot"
```

## <a id="documentation"></a> Documentation

- [Examples](https://github.com/SchweizerischeBundesbahnen/lot/blob/main/examples)

### Scope
The current scope of LOT is for building Kubernetes Operators for cases where no API extension via CRDs are planed

> Further features are about to come 😉

## <a id="contributing"></a> Contributing

Contributions are greatly appreciated. The maintainers actively manage the issues list, and try to highlight issues suitable for newcomers. The project follows the typical GitHub pull request model
See [CONTRIBUTING.md](CONTRIBUTING.md) for more details. Before starting any work, please either comment on an existing issue, or file a new one.



## <a id="code-of-conduct"></a> Code of Conduct

To ensure that your project is a welcoming and inclusive environment for all contributors, you should establish a good [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md)

## <a id="license"></a> License

LOT is released under the [Apache License](LICEN
SE.md). Feel free to use, modify, and distribute it as per the terms of the license.
