# envsubst

`envsubst` is a Go package for expanding variables in a string using `${var}` syntax.
Includes support for bash string replacement functions.

## Documentation

[Documentation can be found on GoDoc][doc].

## Supported Functions

| __Expression__                | __Meaning__                                                     |
| -----------------             | --------------                                                  |
| `${var}`                      | Value of `$var`
| `${var-${DEFAULT}}`           | If `$var` is not set, evaluate expression as `${DEFAULT}`
| `${var:-${DEFAULT}}`          | If `$var` is not set or is empty, evaluate expression as `${DEFAULT}`
| `${var=${DEFAULT}}`           | If `$var` is not set, evaluate expression as `${DEFAULT}`
| `${var:=${DEFAULT}}`          | If `$var` is not set or is empty, evaluate expression as `${DEFAULT}`
| `$$var`                       | Escape expressions. Result will be the string `$var`
| `${var^}`                     | Uppercase first character of `$var`
| `${var^^}`                    | Uppercase all characters in `$var`
| `${var,}`                     | Lowercase first character of `$var`
| `${var,,}`                    | Lowercase all characters in `$var`
| `${#var}`                     | String length of `$var`
| `${var:n}`                    | Offset `$var` `n` characters from start
| `${var: -n}`                  | Offset `$var` `n` characters from end
| `${var:n:len}`                | Offset `$var` `n` characters with max length of `len`
| `${var#pattern}`              | Strip shortest `pattern` match from start
| `${var##pattern}`             | Strip longest `pattern` match from start
| `${var%pattern}`              | Strip shortest `pattern` match from end
| `${var%%pattern}`             | Strip longest `pattern` match from end
| `${var/pattern/replacement}`  | Replace as few `pattern` matches as possible with `replacement`
| `${var//pattern/replacement}` | Replace as many `pattern` matches as possible with `replacement`
| `${var/#pattern/replacement}` | Replace `pattern` match with `replacement` from `$var` start
| `${var/%pattern/replacement}` | Replace `pattern` match with `replacement` from `$var` end

## Unsupported Functions

* `${var-default}`
* `${var+default}`
* `${var:?default}`
* `${var:+default}`

  [doc]: http://godoc.org/github.com/drone/envsubst
