/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package hyperkube

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path"

	"github.com/kubernetes-sigs/service-catalog/pkg/version"
	"k8s.io/klog/v2"

	"github.com/spf13/pflag"

	utiltemplate "github.com/kubernetes-sigs/service-catalog/pkg/kubernetes/pkg/util/template"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/apiserver/pkg/server"
	utilflag "k8s.io/component-base/cli/flag"
	"k8s.io/component-base/logs"

	"github.com/kubernetes-sigs/service-catalog/pkg"
)

// HyperKube represents a single binary that can morph/manage into multiple
// servers.
type HyperKube struct {
	Name string // The executable name, used for help and soft-link invocation
	Long string // A long description of the binary.  It will be word-wrapped before output.

	servers             []Server
	baseFlags           *pflag.FlagSet
	out                 io.Writer
	helpFlagVal         bool
	printVersionFlagVal bool
	makeSymlinksFlagVal bool
}

// AddServer adds a server to the HyperKube object.
func (hk *HyperKube) AddServer(s *Server) {
	hk.servers = append(hk.servers, *s)
	hk.servers[len(hk.servers)-1].hk = hk
}

// FindServer will find a specific server named name.
func (hk *HyperKube) FindServer(name string) (*Server, error) {
	for _, s := range hk.servers {
		if s.Name() == name || s.AlternativeName == name {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("Server not found: %s", name)
}

// Servers returns a list of all of the registered servers
func (hk *HyperKube) Servers() []Server {
	return hk.servers
}

// Flags returns a flagset for "global" flags.
func (hk *HyperKube) Flags() *pflag.FlagSet {
	if hk.baseFlags == nil {
		hk.baseFlags = pflag.NewFlagSet(hk.Name, pflag.ContinueOnError)
		hk.baseFlags.SetOutput(io.Discard)
		hk.baseFlags.SetNormalizeFunc(utilflag.WordSepNormalizeFunc)
		hk.baseFlags.BoolVarP(&hk.helpFlagVal, "help", "h", false, "help for "+hk.Name)
		hk.baseFlags.BoolVar(&hk.printVersionFlagVal, "version", false, "Print version information and quit")
		hk.baseFlags.BoolVar(&hk.makeSymlinksFlagVal, "make-symlinks", false, "create a symlink for each server in current directory")
		hk.baseFlags.MarkHidden("make-symlinks") // hide this flag from appearing in servers' usage output

		// These will add all of the "global" flags (defined with both the
		// flag and pflag packages) to the new flag set we have.
		hk.baseFlags.AddGoFlagSet(flag.CommandLine)
		hk.baseFlags.AddFlagSet(pflag.CommandLine)

	}
	return hk.baseFlags
}

// Out returns the io.Writer that is used for all usage/error information
func (hk *HyperKube) Out() io.Writer {
	if hk.out == nil {
		hk.out = os.Stderr
	}
	return hk.out
}

// SetOut sets the output writer for all usage/error information
func (hk *HyperKube) SetOut(w io.Writer) {
	hk.out = w
}

// Print is a convenience method to Print to the defined output
func (hk *HyperKube) Print(i ...interface{}) {
	fmt.Fprint(hk.Out(), i...)
}

// Println is a convenience method to Println to the defined output
func (hk *HyperKube) Println(i ...interface{}) {
	fmt.Fprintln(hk.Out(), i...)
}

// Printf is a convenience method to Printf to the defined output
func (hk *HyperKube) Printf(format string, i ...interface{}) {
	fmt.Fprintf(hk.Out(), format, i...)
}

// Run the server.  This will pick the appropriate server and run it.
func (hk *HyperKube) Run(args []string, stopCh <-chan struct{}) error {
	// If we are called directly, parse all flags up to the first real
	// argument.  That should be the server to run.
RunAgain:
	command := args[0]
	serverName := path.Base(command)
	args = args[1:]
	if serverName == hk.Name {

		baseFlags := hk.Flags()
		baseFlags.SetInterspersed(false) // Only parse flags up to the next real command
		err := baseFlags.Parse(args)
		if err != nil || hk.helpFlagVal {
			if err != nil {
				hk.Println("Error:", err)
			}
			hk.Usage()
			return err
		}

		if hk.makeSymlinksFlagVal {
			return hk.MakeSymlinks(command)
		}

		if hk.printVersionFlagVal {
			pkg.PrintAndExit()
		}

		args = baseFlags.Args()
		if len(args) > 0 && len(args[0]) > 0 {
			serverName = args[0]
			args = args[1:]
		} else {
			err = errors.New("no server specified")
			hk.Printf("Error: %v\n\n", err)
			hk.Usage()
			return err
		}
	}

	s, err := hk.FindServer(serverName)
	if err != nil {
		if len(args) > 0 {
			goto RunAgain // the first args was popped off at start of Run, try again with new args
		}
		hk.Printf("Error: %v\n\n", err)
		hk.Usage()
		return err
	}

	s.Flags().AddFlagSet(hk.Flags())
	err = s.Flags().Parse(args)
	if err != nil || hk.helpFlagVal {
		if err != nil {
			hk.Printf("Error: %v\n\n", err)
		}
		s.Usage()
		return err
	}

	if hk.printVersionFlagVal {
		pkg.PrintAndExit()
	}

	logs.InitLogs()
	defer logs.FlushLogs()
	klog.Infof("Service Catalog version %s (built %s)", version.Get().String(), version.Get().BuildDate)
	if !s.RespectsStopCh {
		// For commands that do not respect the stopCh, we run them in a go
		// routine and leave them running when stopCh is closed.
		errCh := make(chan error)
		go func() {
			errCh <- s.Run(s, s.Flags().Args(), wait.NeverStop)
		}()
		select {
		case <-stopCh:
			return errors.New("interrupted") // This error text is ignored.
		case err = <-errCh:
			// fall-through
		}
	} else {
		err = s.Run(s, s.Flags().Args(), stopCh)
	}
	if err != nil {
		hk.Println("Error:", err)
	}

	return err
}

// RunToExit will run the hyperkube and then call os.Exit with an appropriate exit code.
func (hk *HyperKube) RunToExit(args []string) {
	stopCh := server.SetupSignalHandler()
	if err := hk.Run(args, stopCh); err != nil {
		os.Exit(1)
	}
}

// Usage will write out a summary for all servers that this binary supports.
func (hk *HyperKube) Usage() {
	tt := `{{if .Long}}{{.Long | trim | wrap ""}}
{{end}}Usage

  {{.Name}} <server> [flags]

Servers
{{range .Servers}}
  {{.Name}}
{{.Long | trim | wrap "    "}}{{end}}
Call '{{.Name}} --make-symlinks' to create symlinks for each server in the local directory.
Call '{{.Name}} <server> --help' for help on a specific server.
`
	utiltemplate.ExecuteTemplate(hk.Out(), tt, hk)
}

// MakeSymlinks will create a symlink for each registered hyperkube server in the local directory.
func (hk *HyperKube) MakeSymlinks(command string) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	var errs bool
	for _, s := range hk.servers {
		link := path.Join(wd, s.Name())

		err := os.Symlink(command, link)
		if err != nil {
			errs = true
			hk.Println(err)
		}
	}

	if errs {
		return errors.New("error creating one or more symlinks")
	}
	return nil
}
