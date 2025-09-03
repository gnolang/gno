# stress-test-many-posts

This is a utility which adds millions of posts to the (local) boards realm to test how memory and
transaction time change with lots of realm storage.

## Run

Start a local gno.land:

```
cd gno/gno.land
make install
gnoland start -lazy -skip-genesis-sig-verification
```

Start gnoweb. In a separate terminal enter:

```
gnoweb
```

Run the utility. In a separate terminal enter:

```
cd gno/misc/stress-test/stress-test-many-posts
go run .
```

## Monitor

This calls r/demo/boards to add `testboard` and a thread if they don't exist. This utility adds millions of replies to the first post on board #1.
To monitor progress, in a web browser go to http://127.0.0.1:8888/r/demo/boards:testboard . The first post says something like "(100 replies)".
Refresh your browser to see this number increase.

## Optional speedup

This utility adds 50 replies per transaction (as allowed by the maximum gas of 100 million). By default gno.land does one transaction every
5 seconds. It is optional but recommended to decrease this time to 1 second as follows:

* In the terminal where you started gno.land, hit ctrl-C
* In a text editor, open `gnoland-data/config/config.toml`
* Change `timeout_commit` to "1s"
* Save and close the text editor.
* Restart gno.land by entering `gnoland start -lazy -skip-genesis-sig-verification`
* In the terminal where you started the test, restart it by entering `go run .`

Now data is added 5 times faster. This can reduce adding a million replies from days to hours.

## Test output

The utility prints results to the terminal in CSV which you can paste into a spreadsheet to analyze and graph.
Below is a sample output. On average the transaction time is 1 second (if you did the optional speedup), but
can be longer if the system becomes burdened with lots of data in memory.

```
nPosts, avg. for 50 posts [s], min for 50 posts [s], max for 50 posts [s]
1000, 1.044250, 1.015000, 1.136000
2000, 1.048650, 0.999000, 1.106000
3000, 1.047950, 1.021000, 1.079000
4000, 1.046450, 1.028000, 1.072000
```
