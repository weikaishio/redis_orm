// +build !windows

package sync2db

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime/pprof"
	"strings"
	"syscall"
)

// (DOC)
// 监听 os.Signal 信号进行处理
// 杀进程的时候带上信号可以在进程结束前做清理操作
// 最简单的用法是使用 ListenQuitAndDump 函数,这个函数定死了使用的信号参数: INT/USR1 杀进程,USR2 dump进程状态
//
//	func main() {
//		// 启动服务(不能阻塞否则下面的信号监听不能执行)
//		// .......
//
//		// 等待杀进程信号(阻塞直到INT/USR1信号到来)
//		osutil.ListenQuitAndDump()
//
//		// 然后开始清理
//		// .......
//	}
//
// 或者调用 ListenSignal 做更高级的定制
//
//	func main() {
//		// 启动服务(不能阻塞否则下面的信号监听不能执行)
//		// .......
//
//		// 监听信号直到 INT 信号到来
//		ListenSignal(func(sig os.Signal) bool {
//			switch sig {
//			case syscall.SIGINT:
//				// 结束信号监听
//				// 也可以直接在这里启动清理操作
//				return true
//			case syscall.SIGUSR1:
//				// 不结束信号监听
//				return false
//			}
//			return false
//		}, syscall.SIGINT, syscall.SIGUSR1)
//
//		// 然后开始清理
//		// .......
//	}

type SignalHandleFunc func(sig os.Signal, reload SignalRelodFunc) (ret bool)
type SignalRelodFunc func()

// 监听信号 signals, 当收到其中一个信号时调用 handler
// (NOTE): ListenSignal 函数将会阻塞直到指定的信号到来且 handler 处理信号后返回true(如果返回false,会继续接受信号)
func ListenSignal(handler SignalHandleFunc, reload SignalRelodFunc, signals ...os.Signal) {
	sigChan := make(chan os.Signal)
	signal.Notify(sigChan, signals...)
	for {
		sig := <-sigChan
		if handler(sig, reload) {
			break
		}
	}
}

// (NOTE): ListenQuitAndDump 函数将会阻塞直到 INT/USR1/USR2 信号到来
func ListenQuitAndDump() {
	ListenSignal(QuitAndDumpAndReload, nil, syscall.SIGINT, syscall.SIGUSR1, syscall.SIGUSR2)
}

func ListenQuitAndDumpAndReload(reload SignalRelodFunc) {
	ListenSignal(QuitAndDumpAndReload, reload, syscall.SIGINT, syscall.SIGUSR1, syscall.SIGUSR2)
}

// 这是一个 SignalHandleFunc,用于退出或dump进程
// 退出监听: syscall.SIGINT, syscall.SIGUSR1
// dump监听: syscall.SIGUSR2
// 使用 kill 命令时可以带上信号参数:
//
//	kill -s INT <pid> 杀进程
//	kill -s USR1 <pid> reload配置
//	kill -s USR2 <pid> dump内存堆栈
func QuitAndDumpAndReload(sig os.Signal, reload SignalRelodFunc) bool {
	switch sig {
	case syscall.SIGINT:
		return true
	case syscall.SIGUSR1:
		//reload config
		if reload != nil {
			reload()
		}
	case syscall.SIGUSR2:
		// dump goroutine stack trace
		filename := filepath.Base(os.Args[0]) + ".dump"
		dumpOut, err := os.OpenFile(filename, os.O_CREATE|os.O_TRUNC|os.O_RDWR, 0777)
		if err == nil {
			for _, name := range []string{"goroutine", "heap", "block"} {
				p := pprof.Lookup(name)
				if p == nil {
					continue
				}
				name = strings.ToUpper(name)
				fmt.Fprintf(dumpOut, "-----BEGIN %s-----\n", name)
				p.WriteTo(dumpOut, 2)
				fmt.Fprintf(dumpOut, "\n-----END %s-----\n", name)
			}
			dumpOut.Close()
		}
	}
	return false
}
