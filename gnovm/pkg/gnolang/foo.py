def main():
    rounds = 1000000000
    n = rounds
    pi = 4 * sum(1 / i for i in range(1 - 2*n, 2*n + 1, 4))

    print("{:.16f}".format(pi))

main()
