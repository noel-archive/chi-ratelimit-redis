# 🔬 chi-ratelimit-redis
> *Implements Redis support for [chi-ratelimit](https://github.com/Noelware/chi-ratelimit)*

## Why is this its own separate package?
I wanted to keep **chi-ratelimit** without any dependencies, so this was the only solution. If only Go had support
for peer-dependencies like NPM. :pleading_face:

## Example
```shell
$ go get github.com/noelware/chi-ratelimit/redis
```

```go
package main

import (
	"net/http"
	"github.com/go-chi/chi/v5"
	"github.com/noelware/chi-ratelimit"
)

func main() {
	ratelimiter := ratelimiter.NewRatelimiter(
		ratelimiter.WithProvider(redis.New(
			redis.WithKeyPrefix("owo:"),
			redis.WithClient(<redis client here>),
		)),
	)
	
	router := chi.NewRouter()
	router.Use(ratelimiter.Middleware)
	router.Get("/", func (w http.ResponseWriter, req *http.Request) {
		// do stuff here
	})
	
	http.ListenAndServe(":3030", router)
}
```

## License
**chi-ratelimit-redis** is released under the [MIT License](https://github.com/Noelware/chi-ratelimit-redis/blob/master/LICENSE)
by **Noelware**.
