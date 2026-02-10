# `gnobench`  the time consumed for GnoVM OpCode execution and store access

`gnobench` benchmarks the time consumed for each VM CPU OpCode and persistent access to the store, including marshalling and unmarshalling of realm objects.

## Usage

### Simple mode

The benchmark only involves the GnoVM and the persistent store. It benchmarks the bare minimum components, and the results are isolated from other components. We use standardize gno contract to perform the benchmarking.

This mode is the best for benchmarking each major release and/or changes in GnoVM. Build it using the GnoVM Makefile:

    make build.bench.opcode
    make build.bench.storage
    make build.bench.native

### Production mode

It benchmarks the node in the production environment with minimum overhead.
We can not only benchmark with standardize the contract but also capture the live usage in production environment.
It gives us a complete picture of the node perform.


  1. Build the production node with benchmarking flags:

  `go build -tags "benchmarkingstorage benchmarkingops benchmarkingnative" gno.land/cmd/gnoland`

  2. Run the node in the production environment. It will dump benchmark data to a benchmarks.bin file.

  3. call the realm contracts at `gno.land/r/x/benchmark/opcodes`, `gno.land/r/x/benchmark/storage` and `gno.land/r/x/benchmark/native`

  4. Stop the server after the benchmarking session is complete.

  5. Run the following command to convert the binary dump:

  `gnobench -bin path_to_benchmarks.bin`

    it converts the binary dump to results.csv and results_stats.csv.


## Results

The benchmarking results are stored in two files:
  1. The raw results are saved in results.csv.

  | Operation       | Elapsed Time | Disk IO Bytes |
  |-----------------|--------------|---------------|
  | OpEval          | 40333        | 0             |
  | OpPopBlock      | 208          | 0             |
  | OpHalt          | 167          | 0             |
  | OpEval          | 500          | 0             |
  | OpInterfaceType | 458          | 0             |
  | OpPopBlock      | 166          | 0             |
  | OpHalt          | 125          | 0             |
  | OpInterfaceType | 21125        | 0             |
  | OpEval          | 541          | 0             |
  | OpEval          | 209          | 0             |
  | OpInterfaceType | 334          | 0             |



  2. The averages and standard deviations are summarized in results_stats.csv.

  | Operation      | Avg Time | Avg Size | Time Std Dev | Count |
|----------------|----------|----------|--------------|-------|
| OpAdd          | 101      | 0        | 45           | 300   |
| OpAddAssign    | 309      | 0        | 1620         | 100   |
| OpArrayLit     | 242      | 0        | 170          | 700   |
| OpArrayType    | 144      | 0        | 100          | 714   |
| OpAssign       | 136      | 0        | 95           | 2900  |
| OpBand         | 92       | 0        | 30           | 100   |
| OpBandAssign   | 127      | 0        | 62           | 100   |
| OpBandn        | 97       | 0        | 54           | 100   |
| OpBandnAssign  | 125      | 0        | 113          | 100   |
| OpBinary1      | 128      | 0        | 767          | 502   |
| OpBody         | 127      | 0        | 145          | 13700 |

## Design consideration

### Minimum Overhead and Footprint

- Constant build flags enable benchmarking.
- Encode operations and measurements in binary.
- Dump to a local file in binary.
- No logging, printout, or network access involved.

### Accurate

- Pause the timer for storage access while performing VM opcode benchmarking.
- Measure each OpCode execution in nanoseconds.
- Store access includes the duration for Amino marshalling and unmarshalling.
