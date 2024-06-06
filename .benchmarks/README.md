# Benchmarks

This folder is where benchmarks are configured to be added on the dashboard generated in [benchmarks](https://gnoland.github.io/benchmarks).

We are using the [gobenchdata](https://github.com/bobheadxi/gobenchdata) GitHub action to run all our benchmarks and generate the graphs. Use its documentation if you need to do something more complicated than adding some benchmarks from a new package.

We have two types of benchmarks; slow and fast ones. Slow ones can also be executed as checks on every PR.

Now let's see how to add your tests to the generated benchmark graphs and also add as checks if they are fast enough on every PR:

## Add new benchmarks to generated graphs.

All benchmarks can be added to these graphs to keep track of the performance evolution on different parts of the code. This is done adding new lines on [gobenchdata-web.yml](https://github.com/gnolang/gno/blob/gh-benchmarks/gobenchdata-web.yml)

This is eventually copied into [benchmark](https://github.com/gnolang/benchmarks/tree/gh-pages) gh-pages branch and it will be rendered [here](https://gnolang.github.io/benchmarks/).

Things to take into account:

- All benchmarks on a package will be shown on the same graph.
- The value on `package` and `benchmarks` are regular expressions.
- You have to explicitly add your new package here to make it appears on generated graphs.
- If you have benchmarks on the same package that takes much more time per op than the rest, you should divide it into a separate graph for visibility. In this example we can see how we separated tests from the gnolang package into the ones finishing with `Equality` and `LoopyMain`, because `LoopyMain` is taking an order of magnitude more time per operation than the other tests:
```yaml
      - name: Equality benchmarks (gnovm)
        benchmarks: [ '.Equality' ]
        package: github.com\/gnolang\/gno\/gnovm\/pkg\/gnolang
      - name: LoopyMain benchmarks (gnovm)
        benchmarks: [ '.LoopyMain' ]
        package: github.com\/gnolang\/gno\/gnovm\/pkg\/gnolang
```

## Add new checks for PRs

If we want to add a new package to check all the fast benchmarks on it on every PR, we should have a look into [gobenchdata-checks.yml](./gobenchdata-checks.yml).
