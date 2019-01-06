# klyn-log

> klyn-log is structured, persistent logging for Go.

## how to use

``` go
import(
    klog "git.yusank.space/yusank/klyn-log"
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

    // other thing
}    
```

### before install
 - go > 1.7

### how to install

``` sh
$ go get git.yusank.space/yusank/klyn-log
```
 
## testing

test with three mode, `...Mode1`:`FlushModeEveryLog`,`...Mode2`:`FlushModeByDuration`,`...Mode3`:`FlushModeBySize`

`-8` means run on 8 core cpu machine

 result: 
``` sh
$ go test -bench=. -count=5 -run=^$
goos: darwin
goarch: amd64
pkg: git.yusank.space/yusank/klyn-log
BenchmarkNewLoggerMode1-8   	  100000	     10095 ns/op
BenchmarkNewLoggerMode1-8   	  200000	     11567 ns/op
BenchmarkNewLoggerMode1-8   	  100000	     10174 ns/op
BenchmarkNewLoggerMode1-8   	  100000	     13317 ns/op
BenchmarkNewLoggerMode1-8   	  200000	     13025 ns/op
BenchmarkNewLoggerMode2-8   	 1000000	      1866 ns/op
BenchmarkNewLoggerMode2-8   	 1000000	      1864 ns/op
BenchmarkNewLoggerMode2-8   	 1000000	      1855 ns/op
BenchmarkNewLoggerMode2-8   	 1000000	      1859 ns/op
BenchmarkNewLoggerMode2-8   	 1000000	      1866 ns/op
BenchmarkNewLoggerMode3-8   	 1000000	      1856 ns/op
BenchmarkNewLoggerMode3-8   	 1000000	      1897 ns/op
BenchmarkNewLoggerMode3-8   	 1000000	      1878 ns/op
BenchmarkNewLoggerMode3-8   	 1000000	      1875 ns/op
BenchmarkNewLoggerMode3-8   	 1000000	      1882 ns/op
PASS
 ```
## Authors
- [yusank](http://git.yusank.space/yusank)

## license
[`MIT` license](http://git.yusank.space/yusank/klyn-log/src/master/LICENSE)