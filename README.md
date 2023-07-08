[![Go Reference](https://pkg.go.dev/badge/github.com/bitfield/gotestdox.svg)](https://pkg.go.dev/github.com/bitfield/gotestdox)
[![Go Report Card](https://goreportcard.com/badge/github.com/bitfield/gotestdox)](https://goreportcard.com/report/github.com/bitfield/gotestdox)
[![Mentioned in Awesome Go](https://awesome.re/mentioned-badge-flat.svg)](https://github.com/avelino/awesome-go)
![Tests](https://github.com/bitfield/gotestdox/actions/workflows/test.yml/badge.svg)

![Writing gopher logo](img/gotestdox.png)

`gotestdox` is a command-line tool for turning Go test names into readable sentences. Here's how to install it:

```
go install github.com/bitfield/gotestdox/cmd/gotestdox@latest
```

# What?

For example, suppose we have some tests named like this:

```
TestRelevantIsFalseForNonPassFailEvents
TestRelevantIsTrueForTestPassOrFailEvents
```

We can transform them into straightforward sentences that express the desired behaviour, by running `gotestdox`:

**`gotestdox`**

This will run the tests, and print:

```
 ✔ Relevant is false for non pass fail events (0.00s)
 ✔ Relevant is true for test pass or fail events (0.00s)
```

# Why?

I read a blog post by Dan North, which says:

> My first “Aha!” moment occurred as I was being shown a deceptively simple utility called `agiledox`, written by my colleague, Chris Stevenson. It takes a JUnit test class and prints out the method names as plain sentences.
>
> The word “test” is stripped from both the class name and the method names, and the camel-case method name is converted into regular text. That’s all it does, but its effect is amazing.
>
> Developers discovered it could do at least some of their documentation for them, so they started to write test methods that were real sentences.\
—Dan North, [Introducing BDD](https://dannorth.net/introducing-bdd/)

# How?

The original [`testdox`](https://github.com/astubbs/testdox) tool (part of `agiledox`) was very simple, as Dan describes: it just turned a camel-case JUnit test name like `testFailsForDuplicateCustomers` into a space-separated sentence like `fails for duplicate customers`.

And that's what I find neat about it: it's so simple that it hardly seems like it could be of any value, but it is. I've already used the idea to improve a lot of my test names.

There are implementations of `testdox` for various languages other than Java: for example, [PHP](https://phpunit.readthedocs.io/en/9.5/textui.html#testdox), [Python](https://pypi.org/project/pytest-testdox/), and [.NET](https://testdox.wordpress.com/). I haven't found one for Go, so here it is.

`gotestdox` reads the JSON output generated by the `go test -json` command. This is easier than trying to parse Go source code, for example, and also gives us pass/fail information for the tests. It ignores all events except pass/fail events for individual tests (including subtests).

# Getting fancy

Some more advanced ways to use `gotestdox`:

## Exit status

If there are any test failures, `gotestdox` will print the output messages from the offending test and report status 1 on exit.

## Colour

`gotestdox` indicates a passing test with a `✔` (check mark emoji), and a failing test with an `x`. These are displayed as green and red respectively, using the [`color`](https://github.com/fatih/color) library, which automagically detects if it's talking to a colour-capable terminal.

If not (for example, when you redirect output to a file), or if the [`NO_COLOR`](https://no-color.org/) environment variable is set to any value, colour output will be disabled.

## Test flags and arguments

`gotestdox`, with no arguments, will run the command `go test -json` and process its output.

Any arguments you supply will be passed on to `go test`. For example:

**`gotestdox -run ParseJSON`**

will run the command:

`go test -json -run ParseJSON`

You can supply a list of packages to test, or any other arguments or flags understood by `go test`. However, `gotestdox` only prints events about *tests* (ignoring benchmarks and examples). It doesn't report fuzz tests, since they don't tend to have useful names.

## Multiple packages

To test all the packages in the current tree, run:

**`gotestdox ./...`**

Each package's test results will be prefixed by the fully-qualified name of the package. For example:

```
github.com/octocat/mymodule/api:
 ✔ NewServer errors on invalid config options (0.00s)
 ✔ NewServer returns a correctly configured server (0.00s)

github.com/octocat/mymodule/util:
 x LeftPad adds the correct number of leading spaces (0.00s)
    util_test.go:133: want "  dummy", got " dummy"
 ```

## Multi-word function names

There's an ambiguity about test names involving functions whose names contain more than one word. For example, suppose we're testing a function `HandleInput`, and we write a test like this:

```
TestHandleInputClosesInputAfterReading
```

Unless we do something, this will be rendered as:

```
 ✔ Handle input closes input after reading
```

To let us give `gotestdox` a hint about this, there's one extra transformation rule: the first underscore marks the end of the function name. So we can name our test like this:

```
TestHandleInput_ClosesInputAfterReading
```

and this becomes:

```
 ✔ HandleInput closes input after reading
```

I think this is an acceptable compromise: the `gotestdox` output is much more readable, while the extra underscore in the test name doesn't seriously interfere with its readability.

The intent is not to *perfectly* render all sensible test names as sentences, in any case, but to do *something* useful with them, primarily to encourage developers to write test names that are informative descriptions of the unit's behaviour, and thus (as a side effect) read well when formatted by `gotestdox`.

In other words, `gotestdox` is not the thing. It's the thing that gets us to the thing, the end goal being meaningful test names (I like the term _literate_ test names).

## Filtering standard input

If you want to run `go test -json` yourself, for example as part of a shell pipeline, and pipe its output into `gotestdox`, you can do that too:

**`go test -json | gotestdox`**

In this case, any flags or arguments to `gotestdox` will be ignored, and it won't *run* the tests; instead, it will act purely as a text filter. However, just as when it runs the tests itself, it will report exit status 1 if there are any test failures.

## As a package

See [pkg.go.dev/github.com/bitfield/gotestdox](https://pkg.go.dev/github.com/bitfield/gotestdox) for the full documentation on using `gotestdox` as a package in your own programs.

# So what?

Why should you care, then? What's interesting about `gotestdox`, or any `testdox`-like tool, I find, is the way its output makes you think about your tests, how you name them, and what they do.

As Dan says in his blog post, turning test names into sentences is a very simple idea, but it has a powerful effect. Test names *should* be sentences.

## Test names should be sentences

I don't know about you, but I've wasted a lot of time and energy over the years trying to choose good names for tests. I didn't really have a way to evaluate whether the name I chose was good or not. Now I do!

In fact, I wrote a whole blog post about it:

* [Test names should be sentences](https://bitfieldconsulting.com/golang/test-names)

It might be interesting to show your `gotestdox` output to users, customers, or business folks, and see if it makes sense to them. If so, you're on the right lines. And it's quite likely to generate some interesting conversations (“Is that really what it does? But that's not what we asked for!”)

It seems that I'm not the only one who finds this idea useful. I hear that `gotestdox` is already being used in some fairly major Go projects and companies, helping their developers to get more value out of their existing tests, and encouraging them to think in interesting new ways about what tests are really for. How nice!

# Links

- [Bitfield Consulting](https://bitfieldconsulting.com/)
- [Test names should be sentences](https://bitfieldconsulting.com/golang/test-names)
- [The Power of Go: Tests](https://bitfieldconsulting.com/books/tests)

<small>Gopher image by [MariaLetta](https://github.com/MariaLetta/free-gophers-pack)</small>
