# klyn-log

> klyn-log is structured, persistent logging for Go.

## how to use

``` go
import(
    klog "git.yusank.cn/yusank/klyn-log"
)

var logger klog.Logger
func main() {
    // default
    logger = klog.DefaultLogger()

    // or 
    conf := &klog.LoggerConfig{
        // custom
    }

    logger = klog.NewLogger(conf)

    logger.Warn(map[string]interface{}{
				"name":   "hello world",
				"userId": 1234,
				"event": map[string]interface{}{
					"gameId": "dddjs",
				},
			})
}    
```

### before install
 - go > 1.7

### how to install

``` sh
$ go get git.yusank.cn/yusank/klyn-log
```
 
## testing

test with three mode, `...Mode1`:`FlushModeEveryLog`,`...Mode2`:`FlushModeByDuration`,`...Mode3`:`FlushModeBySize`

`-8` means run on 8 core cpu machine

 result: 
``` sh
$ go test -bench=. -run=^$
goos: darwin
goarch: amd64
pkg: git.yusank.cn/yusank/klyn-log
BenchmarkNewLoggerMode1-8         200000              8799 ns/op
BenchmarkNewLoggerMode2-8        1000000              1912 ns/op
BenchmarkNewLoggerMode3-8        1000000              1853 ns/op
PASS
ok      git.yusank.cn/yusank/klyn-log   5.681s
 ```
> suggest use mode 2 or 3 for now .

## Authors
- [yusank](https://git.yusank.cn/yusank)

## license
[`MIT` license](https://git.yusank.cn/yusank/klyn-log/src/master/LICENSE)
