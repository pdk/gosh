# pipes and dags

A pipe is a degenerate form of a DAG.

    a -> b -> c -> d

what about branching

        +-> b
    a --+
        +-> c

and merges

    a --+
        |-> c
    b --+

can we have multiple yields?

    func() {
        # blah
        yield(1) "hello"
        yield(2) "world"
    }

or named output channels

    func() [x,y] {
        x << "hello"
        y << "hello"
    }

named input channels

    func(a,b) {
        for v in a {
            print v
        }
        for v in b {
            print v
        }
    }

auto-switching on input channels:

    func(a,b) {
        for v in a,b {
            print v
        }
    }
