# Swift rules

Rules for generating Swift protobuf and gRPC `.swift` files and libraries using [Swift Protobuf](https://github.com/apple/swift-protobuf) and [Swift gRPC](https://github.com/grpc/grpc-swift)

| Rule | Description |
| ---: | :--- |
| [swift_proto_compile](#swift_proto_compile) | Generates Swift protobuf `.swift` artifacts |
| [swift_grpc_compile](#swift_grpc_compile) | Generates Swift protobuf+gRPC `.swift` artifacts |
| [swift_proto_library](#swift_proto_library) | Generates a Swift protobuf library using `swift_library` from `rules_swift` |
| [swift_grpc_library](#swift_grpc_library) | Generates a Swift protobuf+gRPC library using `swift_library` from `rules_swift` |

---

## `swift_proto_compile`

Generates Swift protobuf `.swift` artifacts

### `WORKSPACE`

```starlark
load("@rules_proto_grpc//swift:repositories.bzl", rules_proto_grpc_swift_repos="swift_repos")

rules_proto_grpc_swift_repos()

load(
    "@build_bazel_rules_swift//swift:repositories.bzl",
    "swift_rules_dependencies",
)

swift_rules_dependencies()

load(
    "@build_bazel_apple_support//lib:repositories.bzl",
    "apple_support_dependencies",
)

apple_support_dependencies()
```

### `BUILD.bazel`

```starlark
load("@rules_proto_grpc//swift:defs.bzl", "swift_proto_compile")

swift_proto_compile(
    name = "person_swift_proto",
    deps = ["@rules_proto_grpc//example/proto:person_proto"],
)
```

### Attributes

| Name | Type | Mandatory | Default | Description |
| ---: | :--- | --------- | ------- | ----------- |
| `deps` | `list<ProtoInfo>` | true | `[]`    | List of labels that provide a `ProtoInfo` (such as `native.proto_library`)          |
| `verbose` | `int` | false | `0`    | The verbosity level. Supported values and results are 1: *show command*, 2: *show command and sandbox after running protoc*, 3: *show command and sandbox before and after running protoc*, 4. *show env, command, expected outputs and sandbox before and after running protoc*          |

---

## `swift_grpc_compile`

Generates Swift protobuf+gRPC `.swift` artifacts

### `WORKSPACE`

```starlark
load("@rules_proto_grpc//swift:repositories.bzl", rules_proto_grpc_swift_repos="swift_repos")

rules_proto_grpc_swift_repos()

load(
    "@build_bazel_rules_swift//swift:repositories.bzl",
    "swift_rules_dependencies",
)

swift_rules_dependencies()

load(
    "@build_bazel_apple_support//lib:repositories.bzl",
    "apple_support_dependencies",
)

apple_support_dependencies()
```

### `BUILD.bazel`

```starlark
load("@rules_proto_grpc//swift:defs.bzl", "swift_grpc_compile")

swift_grpc_compile(
    name = "greeter_swift_grpc",
    deps = ["@rules_proto_grpc//example/proto:greeter_grpc"],
)
```

### Attributes

| Name | Type | Mandatory | Default | Description |
| ---: | :--- | --------- | ------- | ----------- |
| `deps` | `list<ProtoInfo>` | true | `[]`    | List of labels that provide a `ProtoInfo` (such as `native.proto_library`)          |
| `verbose` | `int` | false | `0`    | The verbosity level. Supported values and results are 1: *show command*, 2: *show command and sandbox after running protoc*, 3: *show command and sandbox before and after running protoc*, 4. *show env, command, expected outputs and sandbox before and after running protoc*          |

---

## `swift_proto_library`

> NOTE: this rule is EXPERIMENTAL.  It may not work correctly or even compile!

Generates a Swift protobuf library using `swift_library` from `rules_swift`

### `WORKSPACE`

```starlark
load("@rules_proto_grpc//swift:repositories.bzl", rules_proto_grpc_swift_repos="swift_repos")

rules_proto_grpc_swift_repos()

load(
    "@build_bazel_rules_swift//swift:repositories.bzl",
    "swift_rules_dependencies",
)

swift_rules_dependencies()

load(
    "@build_bazel_apple_support//lib:repositories.bzl",
    "apple_support_dependencies",
)

apple_support_dependencies()
```

### `BUILD.bazel`

```starlark
load("@rules_proto_grpc//swift:defs.bzl", "swift_proto_library")

swift_proto_library(
    name = "person_swift_library",
    deps = ["@rules_proto_grpc//example/proto:person_proto"],
)
```

### Attributes

| Name | Type | Mandatory | Default | Description |
| ---: | :--- | --------- | ------- | ----------- |
| `deps` | `list<ProtoInfo>` | true | `[]`    | List of labels that provide a `ProtoInfo` (such as `native.proto_library`)          |
| `verbose` | `int` | false | `0`    | The verbosity level. Supported values and results are 1: *show command*, 2: *show command and sandbox after running protoc*, 3: *show command and sandbox before and after running protoc*, 4. *show env, command, expected outputs and sandbox before and after running protoc*          |

---

## `swift_grpc_library`

> NOTE: this rule is EXPERIMENTAL.  It may not work correctly or even compile!

Generates a Swift protobuf+gRPC library using `swift_library` from `rules_swift`

### `WORKSPACE`

```starlark
load("@rules_proto_grpc//swift:repositories.bzl", rules_proto_grpc_swift_repos="swift_repos")

rules_proto_grpc_swift_repos()

load(
    "@build_bazel_rules_swift//swift:repositories.bzl",
    "swift_rules_dependencies",
)

swift_rules_dependencies()

load(
    "@build_bazel_apple_support//lib:repositories.bzl",
    "apple_support_dependencies",
)

apple_support_dependencies()
```

### `BUILD.bazel`

```starlark
load("@rules_proto_grpc//swift:defs.bzl", "swift_grpc_library")

swift_grpc_library(
    name = "person_swift_library",
    deps = ["@rules_proto_grpc//example/proto:person_proto"],
)
```

### Attributes

| Name | Type | Mandatory | Default | Description |
| ---: | :--- | --------- | ------- | ----------- |
| `deps` | `list<ProtoInfo>` | true | `[]`    | List of labels that provide a `ProtoInfo` (such as `native.proto_library`)          |
| `verbose` | `int` | false | `0`    | The verbosity level. Supported values and results are 1: *show command*, 2: *show command and sandbox after running protoc*, 3: *show command and sandbox before and after running protoc*, 4. *show env, command, expected outputs and sandbox before and after running protoc*          |
