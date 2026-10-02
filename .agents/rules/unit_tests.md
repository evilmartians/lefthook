---
paths:
  - "internal/**/*_test.go"
---

# Go unit tests

Prefer having tests in a separate package so only public API is tested. Refactor the code if needed to stick with this rule only if it doesn't make code more complicated and hard to read and understand for humans.

## Test functions

For test func names use the following scheme:
- If testing a simple func, name it `TestSimpleFunc` for exported and `Test_simpleFunc` for unexported func.
- If testing a method, name it `TestStructName_MethodName` or `TestStructName_methodName`.

## Assertions

When writing or rewriting a unit test don't use asserts or any helpers except:
- Test helper to prepare the data
- "github.com/google/go-cmp/cmp" to compare the data

Stick to the following assertion style:

```go
if !cmp.Equal(result, want) {
    t.Errorf("object.FunctionName() = %v, want %v", result, want)
}
```

When a variable can be compared using `==` use it instead of `cmp.Equal`.

## Table tests

When a function can behave differently regarding the input or other conditions use table tests:

```go
for name, tt := range map[string]struct{
    // <fields required for testing different inputs and results>
} {
    "condition-1": {
        // ...
    },
    "condition-2": {
        // ...
    },
} {
    t.Run(name, func(t *testing.T) {
        // setup

        // function call

        // assertions
    })
}
```

"condition-1" and "condition-2" ideally should be short and grep-able, still reflect the testcase. Use dashes `-` to separate words, but try not to go beyond 3 words.

Example unit test: @internal/git/wrapper/diff_test.go

## Test helpers

Test helpers live in `tests/helpers/`.

