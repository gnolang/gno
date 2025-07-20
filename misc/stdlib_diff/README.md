# stdlib_diff

stdlibs_diff is a tool that generates an html report indicating differences between gno standard libraries and go standard libraries.

## Usage

Compare the `go` standard libraries the `gno` standard libraries

```shell
./stdlibs_diff -src <path to go standard libraries> -dst <path to gno standard libraries> -out <output directory>
```

Compare the `gno` standard libraries the `go` standard libraries

```shell
./stdlibs_diff -src <path to gno standard libraries> -dst <path to go standard libraries> -out <output directory>
```


## Parameters

| Flag       | Description                                                        | Default value |
| ---------- | ------------------------------------------------------------------ | ------------- |
| src        | Directory containing packages that will be compared to destination | None          |
| dst        | Directory containing packages; used to compare src packages        | None          |
| out        | Directory where the report will be created                         | None          |

## Tips

An index.html is generated at the root of the report location. Utilize it to navigate easily through the report.
