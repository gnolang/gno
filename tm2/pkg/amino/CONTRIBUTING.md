## Compatibility

### Protobuf

Amino currently aims to be and stay proto3 compatible. Please, ensure that any
change you add retains proto3 compatibility. Basic compatibility is ensured by
tests. Proto3 may eventually become deprecated for Amino2.

## Fuzzers

Amino is fuzzed using several fuzzers. At least run [gofuzz] by running the command:
```
make test
```
If [go-fuzzer] isn't installed on your system, make sure to run:
```
go get -u github.com/dvyukov/go-fuzz/go-fuzz-build github.com/dvyukov/go-fuzz/go-fuzz
```
The fuzzers are run by:
```
make gofuzz_json
```
and
```
make gofuzz_binary
```
respectively. Both fuzzers will run in an endless loop and you have to quit them manually. They will output 
any problems (crashers) on the commandline. You'll find details of those crashers in the project directories 
`tests/fuzz/binary/crashers` and `tests/fuzz/json/crashers` respectively. 

If you find a crasher related to your changes please fix it, or file an issue containing the crasher information.

[gofuzz]: https://github.com/google/gofuzz
[go-fuzzer]: https://github.com/dvyukov/go-fuzz
[protobuf]: https://developers.google.com/protocol-buffers/
