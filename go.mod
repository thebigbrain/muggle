module github.com/thebigbrain/muggle

go 1.13

replace golang.org/x/sys => github.com/golang/sys v0.0.0-20200212091648-12a6c2dcc1e4

require (
	github.com/gin-gonic/gin v1.5.0
	github.com/go-redis/redis v6.15.7+incompatible
	github.com/google/uuid v1.1.1
)
