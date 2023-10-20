# Remo

[![Go Reference](https://pkg.go.dev/badge/github.com/itpey/remo.svg)](https://pkg.go.dev/github.com/itpey/remo)
[![license](https://img.shields.io/github/license/itpey/remo)](https://github.com/itpey/remo/blob/main/LICENSE)

Remo is a Go package that provides in-memory key-value storage with expiration capabilities. It is designed to be a simple and efficient way to manage data with an optional time-to-live (TTL) duration for keys.

## Table of Contents

- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
  - [Creating a Storage Instance](#creating-a-storage-instance)
  - [Setting and Retrieving Values](#setting-and-retrieving-values)
  - [Deleting Keys](#deleting-keys)
  - [Automatic Cleanup](#automatic-cleanup)
  - [Resetting the Storage](#resetting-the-storage)
- [Running Tests](#running-tests)
- [Running Benchmarks](#running-benchmarks)
- [License](#license)

## Features

- In-memory key-value storage with optional TTL (time-to-live).
- Thread-safe operations with efficient read and write locking.
- Automatic cleanup of expired keys.
- Simple and straightforward API.

## Installation

You can install Remo using the Go module system:

```bash
go get github.com/itpey/remo
```

## Usage

### Creating a Storage Instance

To get started, create a new instance of the Storage struct:

```go
store := remo.New()
```

### Setting and Retrieving Values

You can use the `Set` and `Get` methods to store and retrieve key-value pairs, with the option to set a time-to-live (TTL) duration for keys.

#### Setting a Key with No Expiration

To set a key with no expiration, simply pass a TTL of 0:

```go
// Set a key with a value and no expiration
store.Set("myKey", "myValue", 0)
```

#### Setting a Key with a Specific TTL

You can also set a key with a specific TTL duration, which represents the time the key will be retained in the storage. For example, to set a key that expires in 30 minutes:

```go
// Set a key with a value and a 30-minute TTL
store.Set("myKey", "myValue", 30 * time.Minute)
```

#### Retrieving a Value

To retrieve a value by key, use the `Get` method. It returns the value associated with the key and an error if the key does not exist or has expired:

```go
value, err := store.Get("myKey")
if err != nil {
    // Handle error (e.g., key not found or key has expired)
} else {
    // Use the retrieved value
}
```

### Deleting Keys

You can delete keys using the `Delete` method:
```go
store.Delete("myKey")
```

### Automatic Cleanup

Remo includes an automatic cleanup feature that removes expired keys at a specified interval. You can start and stop this feature using the following methods:

```go
// Start automatic cleanup (e.g., every 10 minutes)
store.StartCleanup(10 * time.Minute)

// Stop automatic cleanup
store.StopCleanup()
```

### Resetting the Storage

Remo provides a convenient `Reset` method that allows you to clear all keys from the storage. This is useful when you need to start with an empty key-value store. Here's how to use the `Reset` method:

```go
store.Reset()
```

## Running Tests

To run tests for Remo, use the following command:

```bash
go test github.com/itpey/remo
```

## Running Benchmarks

To run benchmarks for Remo, use the following command:

```bash
go test -bench=. github.com/itpey/remo
```

## License

This package is distributed under the Apache License, Version 2.0. See the [LICENSE](https://github.com/itpey/remo/blob/main/LICENSE) file for more details.
