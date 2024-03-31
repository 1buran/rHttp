# rHttp - REPL for HTTP
[![codecov](https://codecov.io/gh/mrBuran/rHttp/graph/badge.svg?token=20IW0GY8R9)](https://codecov.io/gh/mrBuran/rHttp)
[![goreportcard](https://goreportcard.com/badge/github.com/mrBuran/rHttp)](https://goreportcard.com/report/github.com/mrBuran/redmine)
![Main demo](https://i.imgur.com/rezbXW9.gif)

#### Responses with minified JSON 
![JSON min](https://i.imgur.com/YW7GrFu.gif)

#### Load session
![Load session](https://i.imgur.com/rAIuhZC.gif)

## Introduction

This is project was created when I needed something like REPL for http request in terminal,
there are many great tools exist: Postman, Insomnia, httpie, curl etc
but i wanted something little bit different, something lightweight, simple and fast,
something like REPL when I can quickly modify request, send it and see the response
with all the details. I wanted such utility with minimal magic,
ideally without any black magic under the hood.

The project has no ambitious goals. It is not killer of Postamn or httpie or curl.
I hope you may find it useful.

## Features

Currently implemented:
- https, http/2 support
- Auto following the redirects
- Easy manipulation of request cookies, headers, params (query string) and form values
- Automatic syntax highlighting of the body of http responses
- Auto format JSON responses (useful for inspection of minified responses)
- Save & load sessions (useful for complex request setup)

In progress:
- Easy manipulation of JSON request payload (crud of simple data structure)
- Load JSON request payload from file
- Config file for change key bindings, default settings

> [!CAUTION]
> The project is under active development, features or how do they work may change!

## Installation

```sh
go install github.com/mrBuran/rHttp@latest
```

## Key Bindings

| Keys                 | Action                                     |
|:---------------------|:-------------------------------------------|
| `Shift+Right`        | next item of menu                          |
| `Shift+Left`         | prev item of menu                          |
| `Enter`              | set value of text intput                   |
| `Ctrl+g`             | run request                                |
| `Ctrl+d`             | delete item  (param, header or form value) |
| `Space`              | toggle checkbox                            |
| `PageDown`           | scroll down body of response               |
| `PageUp`             | scroll up body of response                 |
| `Tab`                | autocomplete                               |
| `Ctrl+f`             | toggle fullscreen mode                     |
| `Ctl+h`              | toggle full help                           |
| `Ctrl+l`             | load session                               |
| `Ctrl+s`             | save session                               |
| `Ctrl+q` or `Ctrl+c` | quit                                       |
