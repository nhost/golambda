# Example

Golang function in file `hello.go`. And it's output zip file in `outpu.zip`.

## Requirements

- Inside your function, the first exported "Handler" function will be attached to the router.
- Multiple Handler functions are not allowed.
- Handler funcs with any other name than "Handler" are also not allowed
- "Handler" function should be exported, it cannot be "handler".
