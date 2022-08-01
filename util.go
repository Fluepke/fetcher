package main

import (
	"syscall"
	"fmt"
)

func setLimits() {
        var rLimit syscall.Rlimit
        err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit)
        if err != nil {
                fmt.Println("Error Getting Rlimit ", err)
        }

        rLimit.Cur = rLimit.Max
        err = syscall.Setrlimit(syscall.RLIMIT_NOFILE, &rLimit)
        if err != nil {
                fmt.Println(err)
        }

        var rLimit2 syscall.Rlimit
        err = syscall.Getrlimit(syscall.RLIMIT_NOFILE, &rLimit2)
        if err != nil {
		fmt.Println("Error Getting Rlimit ", err)
        }
}
