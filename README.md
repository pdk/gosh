# gosh

a simple scripting language to support "object" pipelines

1. interpreted language, `#!/usr/bin/env gosh`
2. dynamically and strongly typed (type is inferred, but variables cannot change type)
3. all statements are expressions
4. support for coroutines (strongly typed)
5. support json, xml, csv
6. strong date/time support
7. easy HTTP client
8. structs with methods, ala go
9. errors are values
10. function/method type is just "takes N parameters"
11. methods are functions bound to type of target object


## variables

Variables are not declared, but just initialized. One a variables type is set by initialization, it cannot be changed.

    x := 1          # x is an integer
    s := "boo"      # s is a string

Multiple assignment is supported:

    x, s := 1, "boo"

## simple types

1. Bool (`true` or `false`)
2. Integer (`int64`)
3. String
4. Character (go `rune`)
5. Float (`float64`)

Variables of other go types are created by explicit conversion, eg.

    x := int32(123)
    c := complex64(1,2)
    b := byte(127)

The type of an variable can be determined with the `type` function.

    x := 2
    type(x)     # int64
    type("foo") # string

## nil

Any variable can take the value `nil`, but `nil` cannot be used to initialize a
variable as it has no particular type. A typed `nil` can be used, eg:

    n := int64(nil)

An alternate syntax is

    x := nil(int64)

## functions

Functions are declared with the `func` keyword

    f := func(a,b,c) {
        ...
    }

Functions will return the value of the last expression evaluated, if there is no explicit return.

Functions can return multiple values.

    f := func() {
        return "alpha", 42
    }
    a, b := f()

Function/method invocation:

    foo(1,2,3)
    x.foo(1,2,3)

Methods may only be defined on structs, at struct declaration. Some internal
types have methods.

Functions are not tightly bound to their names. A function variable is just
another variable. It can be reassigned. But it can only be reassigned to a new
value of the same type. This means that to assign a new function to an existing
variable, it must have the same number of parameters.

    f := func(a,b,c) {
        return a+b+c
    }

    # error:
    f := func(a,b) {
        return a+b
    }

To declare a function variable which will later (possibly) be defined, use an
empty `func` definition:

    f := foo(a,b,c) {}

## named parameters

Named parameters in funcion invocation is supported.

    f := func(a,b,c) {
        # ...
    }

    f(b: 2, c: 1, a: 23)

Omitted parameters can be set with the default assignment operator.

    f := func(a,b,c) {
        a ?= "apple"    # a is assigned "apple" iff a == nil
        # ...
    }

    f(b: 2, c: 7)

It is an error to omit a parameter in a function invocation that does not have a
default value assignment.

## structs

`struct` is the only way to create/define custom types.

Define a `struct`:

    struct myStruct {
        x := 0
        s := ""
        f := func(_, a, c) {
            return a + _.s + c
        }
    }

Methods are functions having 1+ parameters. When a method is invoked, the target
is bound to the first parameter. It is acceptable to use _ (underscore) as a
parameter name.

Create a new instance:

    s := myStruct()

or

    s := myStruct(42, "hello")

which assigns to struct's fields by order.

An explicit constructor may also be defined:

    struct fooBar {
        x := 0
        fooBar := func(me, x) {
            me.x := x + 42
        }

    }

Using a struct literal:

    s := struct{
        x := 1
        w := "foo"
    }

Struct fields can be accessed by offset.

    x := struct {
        a := "apple"
        b := "ball"
        c := 42
    }

    x[0]    # "apple"
    x[2]    # 42

Struct fields can be access by field name.

    x["a"]  # "apple"

Standard struct methods (cannot be redefined)

    dup()       # create and return a copy of the struct
    flds()      # return a list of field names (names not bound to funcs)
    methods()   # return a list of methods (names bound to funcs)

Type expressions of structs:

    type(struct{})  # struct
    type(fooBar())  # fooBar
    type(fooBar)    # type (fooBar is a type, not a thing)

## self-referential structs

Sometimes it's handy to create recursive `struct`s. A typed `nil` can be used to
initialize such a reference.

    struct binaryTree {
        value := ""     # gonna sort strings
        left := nil(binaryTree)
        right := nil(binaryTree)
    }

## bool

Booleans are of type `bool`. Values are `true` and `false` and `nil`. In most
logical expressions, `nil` is equivalent to `false`. The exception is `==`,
where `false` and `nil` are not equal.

    true && false       # false
    true || false       # true
    !true               # false
    !false              # true
    true && nil         # false
    true || nil         # true
    !nil                # true
    true == false       # false
    false == false      # true
    true == nil         # false
    false == nil        # false
    nil == nil          # true

The function `bool` can be used to convert some types to an actual `bool`. For
numbers, 0 (zero) is false, anything else is true. For strings, if the value is
"true" or "t", then it's `true`; for "" (empty string) or "nil", it's `nil`;
otherwise it's `false`.

    bool(1)             # true
    bool(0)             # false
    bool("true")        # true
    bool("false")       # false
    bool("yes")         # false
    bool("")            # nil

Boolean operators

    !       # not
    &&      # and
    ||      # or
    ^^      # xor

### Logical AND &&

    expr1 && expr2

If the value of `expr1` is a `bool` and it is `true`, then `expr2` will be
evaluated, and its value will be the value of the whole expression. If `expr1`
is not a `bool`, or its value is not `true`, then the value of the expression is
the value of `expr1` (usually `false`).

    true && 42      # 42
    false && 42     # false
    42 && true      # 42
    true && false   # false

### Logical OR ||

    expr1 || expr2
    true || false       # true
    false || true       # true
    true || 42          # true
    false || 42         # 42
    42 || true          # 42 (42 is not false)
    42 || false         # 42
    "false" || true     # "false" ("false" is not false)

If `expr1` is a bool and it's value is `false`/`nil`, then `expr2` will be
evaluated, and be the value of the expression. If `expr1` is not a bool or if
it's value is `true`, then the value of the expression will be the value of
`expr1` and `expr2 will not be evaluate.

### &&/|| expressions

Precedence for `&&` and `||` is the same, and they are evaluated left-to-right.

    p && q || r     # ((p && q) || r)
    p || q && r     # ((p || q) && r)

    true && "aa" || "bb"       # "aa"
    false && "aa" || "bb"      # "bb"
    true && ("aa" || "bb")     # "aa"
    false && ("aa" || "bb")    # false
    true || "aa" && "bb"       # "bb"
    false || "aa" && "bb"      # false

    "yes" && true       # "yes"
    "no" || true        # "no"

    x := a == 1 && "one" || "something else"

If a is 1, then x will get "one", otherwise "something else".

### quick returns

    err == nil && return "meh", err
    err != nil || return "meg", err

## math

Standard numerical operators:

    *, /, %, +, -

## relational

Standard relational operators:

    ==, !=, >, <, >=, <=

Structs can define an `equals(a,b)` method, and it will be invoked for `==` and
`!=` operations. If no such method exists, then the structs will be checked for
equality by comparing type, then field values (but not methods).

Lists are equal if they are the same length and all values are `==`.

Maps are equal if they have exactly the same keys, in the same order, and values
that are `==`.

## assignment

The standard assignment operator is `:=`.

    x := 1

An assignment is an expression, so this is valid:

    x := y := z := 1

There is also the accumulator operator:

    x += 1
    y += -1
    s += " moar"

And the nil-assignment:

    v ?= 42

This is a conditional assignment. Only if the current value of the variable is
`nil` will the value be assigned.

## lists

Lists can contain any/mixed types.

    l := []                     # empty list
    l := [1,2,3]                # list of ints
    l := [1,"foo",myStruct()]   # mixed list

`append` is a method, and updates the list.

    l.append("apple")

Values in lists are retrieved with 0-based offset:

    l := [1,"foo",true]
    l[0]        # 1
    l[1]        # "foo"
    l[5]        # error!

Sub-lists:

    l := [1,2,3,4]
    l[1:3]      # [2,3,4]
    l[3:]       # [4]
    l[:1]       # [1,2]

Standard list methods:

    append(x)       # add x to the end
    len()           # return the current size of the list
    pop()           # return the last item appended, remove from the list
    dup()           # create and return a copy of the list

## maps

Maps provide a mapping from strings to values. All maps use keys of type string.

    myMap["foo"] := 23

Maps are also accessible by order assigned (0-based).

    data["name"] := "george"
    data["city"] := "san fran"
    data["age"] := 42
    data["name"] := "fred"

    data["age"]     # 42
    data[1]         # "san fran"
    data[0]         # "fred"

Maps can be accessed with dot notation. (This only works for keys having only
identifier-safe characters.)

    data.city       # "san fran"

To use a struct as a key, the method `hash` (with 0 extra parameters) must be
defined, and it must return a string.

    struct morp {
        x := ""
        hash := func(m) {
            return m.x
        }
    }

Standard map methods:

    del(key)        # remove key from the map
    len()           # return the number of keys in the map
    dup()           # create and return a copy of the map
    keys()          # return an in-order list of the keys
    values()        # return an in-order list of the values

## if

Conditional statements are go-style:

    if ... {
        ...
    }

or

    if ... {
        ...
    } else if ... {
        ...
    } else {
        ...
    }

Note that since everything is an expression, it's acceptable to write something
like this:

    x := if a == b { 1 } else { 2 }

## while

    while ... {
        ...
    }

The conditional expression of a `while` loop must evaluate to a `bool`. Anything
other than a `bool` will cause an error.

A `while` loop may be terminated with a `break`. If a `while` loop completes
without a `break`, its value is `true`. If the loop terminates on a break, then
the value is `false`.

    v := while true {
        break
    }
    # v is false

## for

Coroutine/sequence iteration:

    for v in expr {
        ...
    }

Integers can be used to iterate:

    for v in 3 {
        # v has values 0, 1, 2
    }

Strings:

    for v in "hello" {
        # v has character values 'h','e','l','l','o'
    }

Lists

    for v in [1,2,3] {
        # v has values 1, 2, 3
    }

In `for` loops, the variable is initialized on each iteration, so it's possible
to process lists of mixed types:

    l := [1,"apple"]
    for x in l {
        if x isa int64 {
            printf("int! %d\n", x)
        } else if x isa string {
            printf("string! %s", x)
        }
    }

A `for` loop may also include an index variable

    for i, v in ['a','b','c'] {
        # i has values 0, 1, 2
        # v has values 'a', 'b', 'c'
    }

When using an iterator/coroutine which returns n values, the for loop must have
n or n+1 named variables.


## switch

Multibranch logic:

    switch {
        case a == b {
            ...
        }
        case x isa string {
            ...
        }
        # default case
        ...
    }


## type conditionals

Since function parameters are not typed, functions may receive input parameters
of any type. The keyword `isa` is used for type checking.

    foo := func(x) {
        if x isa [int,int32] {
            return x + 2
        }
        if x isa string {
            return x + " two"
        }
        # implicit nil return
    }

There is no type hierarchy, but lists of types can be used to check multiple
types. There are some predefined type lists.

    x isa std.Number

It is an error to use `isa` against a non-type expression.

    x isa "string" # error. "string" is not string.

A less specific method of type checking is with `hasa` which checks if the
object has a field or method.

    x := struct {
        foo := func() {}
        biff := 1
    }

    x hasa foo                              # true
    x hasa bar                              # false
    x hasa foo && x.foo isa func            # true
    x hasa foo && x.foo isa int             # false
    x hasa biff && x.biff isa std.Number    # true

## enum

`enum` allows the creation of a distinct set of "symbolic" values. A variable
initialized with an enum value may only take other values of the same enum.

    enum Color {
        blue
        green
        red
    }

If an enum value is ambiguous it can be specified with the particular type.

    x := blue           # error. ambiguous
    x := Color.blue     # ok

The standard function `string()` will return the string representation of the
enum value.

An optional value may be specified when defining enums. Values must be of the
same type.

    enum Color {
        blue: "B"
        green: "G"
        red: "R"
    }

or

    enum Color {
        blue:  3
        green: 1
        red:   2
    }

The standard function `int()` will return the int64 value of the particular enum.


## imports

The `import` keyword allows inclusion of other gosh files.

If a second invocation of the same path is found, it will not be re-processed.

    import "../lib/mylib.gosh"

## pkg

A `pkg` is similar to a struct, except there can be only one. `pkg` be thought of
as a static struct, or as a singleton.

This can be used in conjunction with `import` to provide modularity.

There can only be 1 `pkg` command in a file, and it must be the first command,
if present.

    pkg mystuff

    # ... any kinds of definitions/code

    foo := func() {
        printf("yahoo")
    }

    mystuff.foo()


It is an error to redefine a `pkg`. Since an `import` statement will only load a
given file once, it's not a problem to "reimport" a `pkg` multiple times. It
will only be evaluated once.

## generators

Generators are a special case of functions that produce a series of values.

    g := func() {
        << "a"
        << "b"
        << "c"
    }

    for x in g() {
        print(x)
    }
    # prints a, b, c

A generater may use `return` to terminate, but it cannot return any values.

Generators with more than one series can be defined.

    g := func() [i, j] {
        i << "a"
        j << 1
        i << "b"
        j << 2
    }

## consumers

A consumer is a function which iterates on an consumer.

    f := func(dataStream) {
        for item in dataStream {
            # do stuff
        }
    }

To process multiple input streams:

    f := func(s1, s2) {
        for item in s1, s2 {
            # item is from either s1 or s2
        }
    }

## pipelines

A pipeline is a series of generators and consumers where the products of
generators are inputs to consumers.

    numbers := func() {
        for i in 10 {
            << i
        }
    }

    printer := func(input) {
        for n in input {
            printf("%d\n", n)
        }
    }

    numbers() >> printer(~)

If a generator produces more than a single stream (see `g()` above), then the
pipeline can specify where to feed each output stream.

    g() i >> [ addOne() >> printer() ], j >> devNull()

Multiple series can be merged with `^`.

    [ genAlpha() ^ genBeta() ] >> printer()

## multiple workers

How to specify that a consumer can be replicated? Use `#` as a multiple
operator?

    readLines(stdin) >> 5 # lineProcessor() >> printer()

## filters

Filters are used in pipelines to marshal/unmarshal streams of data.

1. `json` -- parse input as a series (0+) of JSON documents
2. `xml` -- parse input as a series (0+) of XML documents
3. `csv` -- parse input as a header, followed by a series (0+) of data rows
4. `yaml` -- ??
5. `words` -- parse each line into a list, whitespace separated
6. `lines` -- parse input as a series of lines, one string per line

Each filter will produce a sequence (0 or more) of items to be processed.

unix syntax:

    ps | grep foo | awk '{ print $1 }' | sort | uniq

functional syntax:

    uniq(sort(awk('{ print $1 }', grep('foo', ps()))))

Each item has an implicit input (`stdin`) and the result of each is fed to the
implicit input.

In gosh `~` is used to denote "product of previous pipeline stage"

    ${ ps -auxw } >> words(~) >> firstWords(~) >> regexp("foo", ~) >> sortStrings(~) >> onlyUniq(~)

but if an item in a pipeline takes only a single argument, the `(~)` can be omitted:

    ${ ps -auxw } >> words >> firstWords >> regexp("foo", ~) >> sortStrings >> onlyUniq

In gosh, the pipeline operator is `>>`.

To run an external command, use the `${ ... }` notation. To run a multiline
command (via `bash`) use `$${ ... }`. If you need command line argument
expansion (e.g. `*.foo`) used the `$$`.

The value of an external command invocation like this is an input stream. That
is then used as the last parameter to the next item in the pipeline.

## credit

built with _Writing An Interpreter In Go_ by Thorsten Ball.
