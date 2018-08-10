package main
import (
	"gopkg.in/urfave/cli.v1"
	"bytes"
	"fmt"
	"github.com/DEL-ORG/del/params"
	"runtime"
	"github.com/DEL-ORG/del/cmd/internal/browser"
	"net/url"
	"io"
	"io/ioutil"
	"os/exec"
	"strings"
)

const issueUrl = "https://github.com/deladmin/del/issues/new"
func reportBug(ctx *cli.Context) error {
	var buff bytes.Buffer
	fmt.Fprintln(&buff, header)
	fmt.Fprintln(&buff, "Version:", params.Version)
	fmt.Fprintln(&buff, "Go Version:", runtime.Version())
	fmt.Fprintln(&buff, "OS:", runtime.GOOS)
	printOSDetails(&buff)
	if !browser.Open(issueUrl + "?body=" + url.QueryEscape(buff.String())) {
		fmt.Printf("Please file a new issue at %s using this template:\n%s", issueUrl, buff.String())
	}
	return nil
}
func printOSDetails(w io.Writer) {
	switch runtime.GOOS {
	case "darwin":
		printCmdOut(w, "uname -v: ", "uname", "-v")
		printCmdOut(w, "", "sw_vers")
	case "linux":
		printCmdOut(w, "uname -sr: ", "uname", "-sr")
		printCmdOut(w, "", "lsb_release", "-a")
	case "openbsd", "netbsd", "freebsd", "dragonfly":
		printCmdOut(w, "uname -v: ", "uname", "-v")
	case "solaris":
		out, err := ioutil.ReadFile("/etc/release")
		if err == nil {
			fmt.Fprintf(w, "/etc/release: %s\n", out)
		} else {
			fmt.Printf("failed to read /etc/release: %v\n", err)
		}
	}
}
func printCmdOut(w io.Writer, prefix, path string, args ...string) {
	cmd := exec.Command(path, args...)
	out, err := cmd.Output()
	if err != nil {
		fmt.Printf("%s %s: %v\n", path, strings.Join(args, " "), err)
		return
	}
	fmt.Fprintf(w, "%s%s\n", prefix, bytes.TrimSpace(out))
}
const header = `Please answer these questions before submitting your issue. Thanks!
#### What did you do?
#### What did you expect to see?
#### What did you see instead?
#### System details
`
