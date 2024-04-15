# rHttp - REPL for HTTP
[![codecov](https://codecov.io/gh/1buran/rHttp/graph/badge.svg?token=20IW0GY8R9)](https://codecov.io/gh/1buran/rHttp)
[![goreportcard](https://goreportcard.com/badge/github.com/1buran/rHttp)](https://goreportcard.com/report/github.com/1buran/redmine)
![Main demo](https://i.imgur.com/I0vIcFS.gif)

#### Responses with minified JSON
![JSON min](https://i.imgur.com/FFrxom5.gif)

#### Edit JSON request payload
![Edit JSON Payload](https://i.imgur.com/08bJisW.gif)

#### Load JSON request payload from file
![Attach file](https://i.imgur.com/CtpGGvZ.gif)

#### Load session
![Load session](https://i.imgur.com/iPGhPGI.gif)

#### Redirects
![Redirects](https://i.imgur.com/rRA4vVy.gif)

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
- Easy manipulation of JSON request payload (through the built-in mini editor)
- Load JSON request payload from file
- Automatic syntax highlighting of the body of http responses
- Auto format JSON responses (useful for inspection of minified responses)
- Save & load sessions (useful for complex request setup)

In progress:
- Load binary data of upload form from file
- Config file for change key bindings, default settings

> [!CAUTION]
> The project is under active development, features or how do they work may change!

## Installation

```sh
go install github.com/1buran/rhttp@latest
```

## Key Bindings

| Keys              | Action                                                  |
|:------------------|:--------------------------------------------------------|
| `Shift+Right`     | next item of menu                                       |
| `Shift+Left`      | prev item of menu                                       |
| `Enter`           | set value of text intput                                |
| `Ctrl+g`          | run request                                             |
| `Ctrl+d`          | delete item  (param, header, form value, attached file) |
| `Space`           | toggle checkbox                                         |
| `PageDown`        | scroll down body of response                            |
| `PageUp`          | scroll up body of response                              |
| `Tab`             | autocomplete                                            |
| `Ctrl+f`          | toggle fullscreen mode                                  |
| `Ctl+h`           | toggle full help                                        |
| `Ctrl+l`          | load session                                            |
| `Ctrl+s`          | save session                                            |
| `Ctrl+q / Ctrl+c` | quit                                                    |
| `Ctrl+j`          | toggle editor (edit JSON request payload)               |
| `Alt+Enter`       | save JSON request payload                               |
| `Ctrl+p`          | load jSON request payload from file                     |

> [!WARNING]
> Some of rHttp key bindigs may overriden by system settings or terminal emulator
> settings, please check them if you face with not working key binding.

## Mini editor

Editor key bindings are from the plugin, they are most tipical shortcuts for common editors,
here they are:

| Keys                                           | Action                                    |
|:-----------------------------------------------|:------------------------------------------|
| `right / ctrl+f`, `left / ctrl+b`              | forward, backward                         |
| `alt+right / alt+f`, `alt+left / alt+b`        | word forward, word backward               |
| `down / ctrl+n`, `up / ctrl+p`                 | line next, line previous                  |
| `alt+backspace / ctrl+w`, `alt+delete / alt+d` | delete word backward, delete word forward |
| `ctrl+k`, `ctrl+u`                             | delete after cursor, delete before cursor |
| `enter / ctrl+m`                               | insert new line                           |
| `backspace`                                    | delete character backward                 |
| `delete`                                       | delete character forward                  |
| `home / ctrl+a`, `end / ctrl+e`                | line start, line end                      |
| `ctrl+v` (depends on your terminal settings)   | paste                                     |
| `alt+< / ctrl+home`, `alt+> / ctrl+end`        | input begin, input end                    |
| `alt+c`, `alt+l`, `alt+u`                      | capitalize, lowercase and uppercase word  |
| `ctrl+t`                                       | transpose character backward              |

> [!IMPORTANT]
> Some of original texarea key bindigs are overriden by the rHttp key bindings, e.g. `ctrl+h` will
> open the help instead of delete character backward or `ctrl+d` will remove JSON request payload
> at all instead of delete character forward.

> [!WARNING]
> Some of original texarea key bindigs may overriden by system settings or terminal emulator
> settings, please check them if you face with not working key binding.

[texarea key bindings](https://pkg.go.dev/github.com/charmbracelet/bubbles/textarea#pkg-variables)

## Tasks

These are tasks of [xc](https://github.com/joerdav/xc) runner.

### vhs

Run VHS fo update gifs.

```
vhs demo/main.tape
vhs demo/json-min.tape
vhs demo/load-session.tape
vhs demo/redirects.tape
vhs demo/edit-json-payload.tape
vhs demo/attach-file.tape
```
### imgur

Upload to Imgur and update readme.

```
declare -A demo=()
demo["main"]="Main demo"
demo["json-min"]="JSON min"
demo["load-session"]="Load session"
demo["redirects"]="Redirects"
demo["edit-json-payload"]="Edit JSON Payload"
demo["attach-file"]="Attach file"

for i in ${!demo[@]}; do
    . .env && url=`curl --location https://api.imgur.com/3/image \
        --header "Authorization: Client-ID ${clientId}" \
        --form image=@demo/$i.gif \
        --form type=image \
        --form title=rHttp \
        --form description=Demo | jq -r '.data.link'`
    sed -i "s#^\!\[${demo[$i]}\].*#![${demo[$i]}]($url)#" README.md
done
```
