# rHttp - REPL for HTTP
[![codecov](https://codecov.io/gh/1buran/rHttp/graph/badge.svg?token=20IW0GY8R9)](https://codecov.io/gh/1buran/rHttp)
[![goreportcard](https://goreportcard.com/badge/github.com/1buran/rHttp)](https://goreportcard.com/report/github.com/1buran/redmine)
![Main demo](https://i.imgur.com/6ao55dy.gif)

#### Responses with minified JSON
![JSON min](https://i.imgur.com/Ii6CzZK.gif)

#### Edit JSON request payload
![Edit JSON Payload](https://i.imgur.com/VAdcP65.gif)

#### Load session
![Load session](https://i.imgur.com/TQ3uKG3.gif)

#### Redirects
![Redirects](https://i.imgur.com/Dm9XCJh.gif)

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
- Automatic syntax highlighting of the body of http responses
- Auto format JSON responses (useful for inspection of minified responses)
- Save & load sessions (useful for complex request setup)

In progress:
- Load JSON request payload from file
- Config file for change key bindings, default settings

> [!CAUTION]
> The project is under active development, features or how do they work may change!

## Installation

```sh
go install github.com/1buran/rhttp@latest
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

## Tasks

These are tasks of [xc](https://github.com/joerdav/xc) runner.

### vhs

Run VHS fo update gifs.

```
vhs demo/main.tape
vhs demo/json-min.tape
vhs demo/load-session.tape
vhs demo/redirects.tape
```
### imgur

Upload to Imgur and update readme.

```
url=`curl --location https://api.imgur.com/3/image \
     --header "Authorization: Client-ID ${clientId}" \
     --form image=@demo/main.gif \
     --form type=image \
     --form title=rHttp \
     --form description=Demo | jq -r '.data.link'`
sed -i "s#^\!\[Main demo\].*#![Main demo]($url)#" README.md

url=`curl --location https://api.imgur.com/3/image \
     --header "Authorization: Client-ID ${clientId}" \
     --form image=@demo/json-min.gif \
     --form type=image \
     --form title=rHttp \
     --form description=Demo | jq -r '.data.link'`
sed -i "s#^\!\[JSON min\].*#![JSON min]($url)#" README.md

url=`curl --location https://api.imgur.com/3/image \
     --header "Authorization: Client-ID ${clientId}" \
     --form image=@demo/load-session.gif \
     --form type=image \
     --form title=rHttp \
     --form description=Demo | jq -r '.data.link'`
sed -i "s#^\!\[Load session\].*#![Load session]($url)#" README.md

url=`curl --location https://api.imgur.com/3/image \
     --header "Authorization: Client-ID ${clientId}" \
     --form image=@demo/redirects.gif \
     --form type=image \
     --form title=rHttp \
     --form description=Demo | jq -r '.data.link'`
sed -i "s#^\!\[Redirects\].*#![Redirects]($url)#" README.md

url=`curl --location https://api.imgur.com/3/image \
     --header "Authorization: Client-ID ${clientId}" \
     --form image=@demo/edit-json-payload.gif \
     --form type=image \
     --form title=rHttp \
     --form description=Demo | jq -r '.data.link'`
sed -i "s#^\!\[Edit JSON Payload\].*#![[Edit JSON Payload]($url)#" README.md
```
