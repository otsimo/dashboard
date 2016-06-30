package kuberesolver

import (
	"fmt"
	"runtime"
	"time"

	"google.golang.org/grpc/grpclog"
)

func Until(f func(), period time.Duration, stopCh <-chan struct{}) {
	select {
	case <-stopCh:
		return
	default:
	}
	for {
		func() {
			defer HandleCrash()
			f()
		}()
		select {
		case <-stopCh:
			return
		case <-time.After(period):
		}
	}
}

// HandleCrash simply catches a crash and logs an error. Meant to be called via defer.
func HandleCrash() {
	if r := recover(); r != nil {
		callers := ""
		for i := 0; true; i++ {
			_, file, line, ok := runtime.Caller(i)
			if !ok {
				break
			}
			callers = callers + fmt.Sprintf("%v:%v\n", file, line)
		}
		grpclog.Printf("kuberesolver: recovered from panic: %#v (%v)\n%v", r, r, callers)
	}
}
