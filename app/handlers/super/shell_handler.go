package super

import (
	"context"
	"io"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"github.com/askasoft/pango/cog"
	"github.com/askasoft/pango/cog/linkedhashmap"
	"github.com/askasoft/pango/str"
	"github.com/askasoft/pango/tbs"
	"github.com/askasoft/pango/xin"
	"github.com/askasoft/pangox-xdemo/app/handlers"
	"github.com/askasoft/pangox-xdemo/app/utils/tbsutil"
)

func ShellIndex(c *xin.Context) {
	h := handlers.H(c)

	h["OS"] = str.Capitalize(runtime.GOOS)
	h["Timeouts"] = tbsutil.GetStrings(c.Locale, "super.shell.timeouts")

	labels := linkedhashmap.NewLinkedHashMap(
		cog.KV("code", tbs.GetText(c.Locale, "super.shell.label.code")),
		cog.KV("time", tbs.GetText(c.Locale, "super.shell.label.time")),
		cog.KV("output", tbs.GetText(c.Locale, "super.shell.label.output")),
	)
	h["Labels"] = labels

	c.HTML(http.StatusOK, "super/shell", h)
}

type ShellArg struct {
	Command string        `form:"command,strip"`
	Timeout time.Duration `form:"timeout"`
}

type ShellResult struct {
	Code   int    `json:"code,omitempty"`
	Time   string `json:"time,omitempty"`
	Output string `json:"output,omitempty"`
}

func ShellExec(c *xin.Context) {
	arg := &ShellArg{}
	_ = c.Bind(arg)

	sr := shellExec(c, arg.Command, arg.Timeout)

	c.JSON(http.StatusOK, sr)
}

func shellExec(c context.Context, command string, timeout time.Duration) (sr ShellResult) {
	if command == "" {
		return
	}

	var (
		exe   string
		arg   []string
		stdin io.Reader
	)

	if runtime.GOOS == "windows" {
		exe = "cmd.exe"
		command += "\r\nexit\r\n"
		stdin = strings.NewReader(command)
	} else {
		exe = "sh"
		command = str.RemoveByte(command, '\r')
		arg = []string{"-e", "-x", "-c", command}
	}

	switch {
	case timeout < time.Second:
		timeout = time.Second
	case timeout > 300*time.Second:
		timeout = 300 * time.Second
	}

	ctx, cancel := context.WithTimeout(c, timeout)
	defer cancel()

	start := time.Now()
	stdout := &strings.Builder{}

	cmd := exec.CommandContext(ctx, exe, arg...)
	cmd.Stdin = stdin
	cmd.Stdout = stdout
	cmd.Stderr = stdout

	if err := cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok { //nolint: errorlint
			sr.Code = exitErr.ExitCode()
		}
		sr.Time = time.Since(start).String()
		sr.Output = err.Error()
		return
	}

	sr.Time = time.Since(start).String()
	sr.Output = str.Strip(stdout.String())
	return
}
