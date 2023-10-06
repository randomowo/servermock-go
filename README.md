configurable mock server

# Installation
## From source
- Clone repository `git clone https://github.com/randomowo/servermock-go`
- cd to repo dir `cd servermock-go`
- Build `go build .`
- Run `./servermock_go -config path/to/config`

> Also, you can set config file path in env via `CONFIG_FILE` param

# Configuration file schema
```yaml
/:                                      # route (must have prefix '/')
  get:
  delete:
  post:
  put:
  default:                              # for any non-configured method
    code: 200                           # response status code
    body:                               # response body
      content_type: "application/json"  # can be "application/json" or "text/plain"
      echo: true                        # echo request body (content_type and value will be ignored)
      value:                            # can be string or object (only if application/json used)
```

> see example [here](./examples/config.yaml) 


## TODO
- [ ] error on event (example: every n requests)
- [ ] tests
- [ ] code documentation
- [ ] openapi hosted documentation
- [ ] more configurable params (addr, port)
- [ ] more mimetypes?
