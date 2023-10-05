# Gnoland Testing Guide

This guide provides an overview of our testing practices and conventions. While most of our testing aligns with typical Go practices, there are exceptions and specifics you should be aware of.

## Standard Package Testing

For most packages, tests are written and executed in the standard Go manner:

- Tests are located alongside the code they test.
- The `go test` command can be used to execute tests.

However, as mentioned earlier, there are some exceptions. In the following sections, we will explore our specialized tests and how to work with them.

## Gno Filetests

**Location:** `gnovm/test/files`

These are our custom file-based tests tailored specifically for this project.

**Execution:** 

From the gnovm directory, There are two main commands to run Gno filetests:

1. To test native files, use:
```
make _test.gnolang.native
```

2. To test standard libraries, use:
```
make _test.gnolang.stdlibs
```

**Golden Files Update:** 

Golden files are references for expected outputs. Sometimes, after certain updates, these need to be synchronized. To do so:

1. For native tests:
```
make _test.gnolang.native.sync
```

2. For standard library tests:
```
make _test.gnolang.stdlibs.sync
```

## Integration Tests

**Location:** `gno.land/**/testdata`

From the gno.land directory, Integration tests are designed to ensure different parts of the project work cohesively. Specifically:

1. **InMemory Node Integration Testing:**  
   Found in `gno.land/cmd/gnoland/testdata`, these are dedicated to running integration tests against a genuine `gnoland` node.

2. **Integration Features Testing:**  
   Located in `gno.land/pkg/integration/testdata`, these tests target integrations specific commands.

These integration tests utilize the `testscript` package and follow the `txtar` file specifications. 

**Documentation:**

- For general `testscript` package documentation, refer to: [testscript documentation](https://github.com/rogpeppe/go-internal/blob/v1.11.0/testscript/doc.go)
  
- For more specific details about our integration tests, consult our extended documentation: [gnoland integration documentation](https://github.com/gnolang/gno/blob/master/gno.land/pkg/integration/doc.go)

**Execution:** 

To run the integration tests (alongside other packages):

```
make _test.pkgs
```

**Golden Files Update within txtar:** 

For tests utilizing the `cmp` command inside `txtar` files, golden files can be synchronized using:

```
make _test.pkgs.sync
```

---

As the project evolves, this guide might be updated. 
