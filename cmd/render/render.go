// render is a frentend to POV-Ray for rendering many POV files in parallel.
package main

// QPov
//
// Copyright (C) Thomas Habets <thomas@habets.se> 2015
// https://github.com/ThomasHabets/qpov
//
//   This program is free software; you can redistribute it and/or modify
//   it under the terms of the GNU General Public License as published by
//   the Free Software Foundation; either version 2 of the License, or
//   (at your option) any later version.
//
//   This program is distributed in the hope that it will be useful,
//   but WITHOUT ANY WARRANTY; without even the implied warranty of
//   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
//   GNU General Public License for more details.
//
//   You should have received a copy of the GNU General Public License along
//   with this program; if not, write to the Free Software Foundation, Inc.,
//   51 Franklin Street, Fifth Floor, Boston, MA 02110-1301 USA.

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path"
	"runtime"
	"sync"
	"time"
)

var (
	povray      = flag.String("povray", "/usr/bin/povray", "Path to povray.")
	schedtool   = flag.String("schedtool", "/usr/bin/schedtool", "Path to schedtool.")
	concurrency = flag.Int("concurrency", -1, "Run this many povrays in parallel. <0 means set to number of CPUs.")
	fast        = flag.Bool("fast", false, "Fast rendering.")
	hq          = flag.Bool("hq", true, "High quality.")
	idle        = flag.Bool("idle", true, "Use idle priority.")

	mutex                  sync.Mutex
	totalUser, totalSystem time.Duration
)

func doRender(files <-chan string, done chan<- bool) {
	for f := range files {
		func() {
			ext := path.Ext(f)
			base := f[:len(f)-len(ext)]
			if _, err := os.Stat(base + ".png"); err == nil {
				return
			}

			stdout, err := os.Create(fmt.Sprintf("%s.stdout", f))
			if err != nil {
				log.Fatalf("Failed to open stdout file: %v", err)
			}
			defer stdout.Close()

			stats, err := os.Create(fmt.Sprintf("%s.stats", f))
			if err != nil {
				log.Fatalf("Failed to open stdout file: %v", err)
			}
			defer stats.Close()

			stderr, err := os.Create(fmt.Sprintf("%s.stderr", f))
			if err != nil {
				log.Fatalf("Failed to open stderr file: %v", err)
			}
			defer stderr.Close()

			var bin string
			var args []string
			if *idle {
				bin = *schedtool
				args = append(args, "-D", "-e", *povray)
			} else {
				bin = *povray
			}
			args = append(args, "-D")
			if *fast {
				args = append(args, "+Q2", "+W400", "+H225")
			} else if *hq {
				args = append(args, "+Q11", "+A0.3", "+R4", "+W1600", "+H900")
			} else {
				args = append(args, "+W1600", "+H900")
			}
			args = append(args, path.Base(f))
			cmd := exec.Command(bin, args...)
			cmd.Stdout = stdout
			cmd.Stderr = stderr
			cmd.Dir = path.Dir(f)

			st := time.Now()
			if err := cmd.Run(); err != nil {
				log.Fatalf("Failed to render %q: %v", f, err)
			}
			fmt.Fprintf(stats, "Sys: %+v\n", cmd.ProcessState.Sys())
			fmt.Fprintf(stats, "SysUsage: %+v\n", cmd.ProcessState.SysUsage())
			fmt.Fprintf(stats, "System time: %v\n", cmd.ProcessState.SystemTime())
			fmt.Fprintf(stats, "User time: %v\n", cmd.ProcessState.UserTime())
			fmt.Fprintf(stats, "Real time: %v\n", time.Since(st))
			func() {
				mutex.Lock()
				defer mutex.Unlock()
				totalUser += cmd.ProcessState.UserTime()
				totalSystem += cmd.ProcessState.SystemTime()
			}()
		}()
	}
	done <- true
}

func main() {
	flag.Parse()
	done := make(chan bool)

	if *concurrency < 0 {
		*concurrency = runtime.NumCPU()
	}

	st := time.Now()
	files := make(chan string)
	for i := 0; i < *concurrency; i++ {
		go doRender(files, done)
	}
	for _, f := range flag.Args() {
		files <- f
	}
	close(files)

	finished := 0
	for _ = range flag.Args() {
		<-done
		finished++
		fmt.Printf("Finished %d of %d\n", finished, len(flag.Args()))
		if finished == len(flag.Args()) {
			break
		}
	}
	mutex.Lock()
	defer mutex.Unlock()
	totalTime := time.Since(st)
	fmt.Printf("Total time: %v (%v user + %v system = %g parallelism)\n", totalTime, totalUser, totalSystem, float64(totalUser+totalSystem)/float64(totalTime))
}
