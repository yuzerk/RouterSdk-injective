// Package build provide customized methods to build project.
// It can add external infos (eg. gitCommit, gitDate) to the version sub command.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"
)

var gobin, _ = filepath.Abs(filepath.Join("build", "bin"))

func main() {
	log.SetFlags(log.Lshortfile)

	if _, err := os.Stat(filepath.Join("build", "ci.go")); os.IsNotExist(err) {
		log.Fatal("this script must be run from the root of the repository")
	}
	if len(os.Args) < 2 {
		log.Fatal("need subcommand as first argument")
	}
	switch os.Args[1] {
	case "install":
		doInstall(os.Args[2:])
	default:
		log.Fatal("unknown command ", os.Args[1])
	}
}

// Compiling

func doInstall(cmdline []string) {
	_ = flag.CommandLine.Parse(cmdline)
	env := Env()

	// Check Go version. People regularly open issues about compilation
	// failure with outdated Go. This should save them the trouble.
	if !strings.Contains(runtime.Version(), "devel") {
		// Figure out the minor version number since we can't textually compare (1.10 < 1.9)
		var minor int
		_, _ = fmt.Sscanf(strings.TrimPrefix(runtime.Version(), "go1."), "%d", &minor)

		if minor < 18 {
			log.Println("You have Go version", runtime.Version())
			log.Println("requires at least Go version 1.18 and cannot")
			log.Println("be compiled with an earlier version. Please upgrade your Go installation.")
			os.Exit(1)
		}
	}
	// Compile packages given as arguments, or everything if there are no arguments.
	packages := []string{"./..."}
	if flag.NArg() > 0 {
		packages = flag.Args()
	}

	goinstall := goTool("install", buildFlags(env)...)
	if runtime.GOARCH == "arm64" {
		goinstall.Args = append(goinstall.Args, "-p", "1")
	}
	goinstall.Args = append(goinstall.Args, "-v")
	goinstall.Args = append(goinstall.Args, packages...)
	MustRun(goinstall)
}

func buildFlags(env *Environment) (flags []string) {
	var ld []string
	if env.Commit != "" {
		ld = append(ld,
			"-X", "main.gitCommit="+env.Commit,
			"-X", "main.gitDate="+env.Date,
		)
	}
	if runtime.GOOS == "darwin" {
		ld = append(ld, "-s")
	}

	if len(ld) > 0 {
		flags = append(flags, "-ldflags", strings.Join(ld, " "))
	}
	return flags
}

func goTool(subcmd string, args ...string) *exec.Cmd {
	return goToolArch(runtime.GOARCH, os.Getenv("CC"), subcmd, args...)
}

func goToolArch(arch, cc, subcmd string, args ...string) *exec.Cmd {
	cmd := GoTool(subcmd, args...)
	if arch == "" || arch == runtime.GOARCH {
		cmd.Env = append(cmd.Env, "GOBIN="+gobin)
	} else {
		cmd.Env = append(cmd.Env, "CGO_ENABLED=1", "GOARCH="+arch)
	}
	if cc != "" {
		cmd.Env = append(cmd.Env, "CC="+cc)
	}
	for _, e := range os.Environ() {
		if strings.HasPrefix(e, "GOBIN=") {
			continue
		}
		cmd.Env = append(cmd.Env, e)
	}
	return cmd
}

// build flags
var (
	// These flags override values in build env.
	GitCommitFlag   = flag.String("git-commit", "", `Overrides git commit hash embedded into executables`)
	GitBranchFlag   = flag.String("git-branch", "", `Overrides git branch being built`)
	GitTagFlag      = flag.String("git-tag", "", `Overrides git tag being built`)
	BuildnumFlag    = flag.String("buildnum", "", `Overrides CI build number`)
	PullRequestFlag = flag.Bool("pull-request", false, `Overrides pull request status of the build`)
	CronJobFlag     = flag.Bool("cron-job", false, `Overrides cron job status of the build`)
)

// Environment contains metadata provided by the build environment.
type Environment struct {
	Name                      string // name of the environment
	Repo                      string // name of GitHub repo
	Commit, Date, Branch, Tag string // Git info
	Buildnum                  string
	IsPullRequest             bool
	IsCronJob                 bool
}

func (env *Environment) String() string {
	return fmt.Sprintf("%s env (commit:%s date:%s branch:%s tag:%s buildnum:%s pr:%t)",
		env.Name, env.Commit, env.Date, env.Branch, env.Tag, env.Buildnum, env.IsPullRequest)
}

func isTrueEnv(name string) bool {
	return strings.EqualFold(os.Getenv(name), "true")
}

// Env returns metadata about the current CI environment, falling back to LocalEnv
// if not running on CI.
func Env() *Environment {
	switch {
	case isTrueEnv("CI") && isTrueEnv("TRAVIS"):
		commit := os.Getenv("TRAVIS_PULL_REQUEST_SHA")
		if commit == "" {
			commit = os.Getenv("TRAVIS_COMMIT")
		}
		return &Environment{
			Name:          "travis",
			Repo:          os.Getenv("TRAVIS_REPO_SLUG"),
			Commit:        commit,
			Date:          getDate(commit),
			Branch:        os.Getenv("TRAVIS_BRANCH"),
			Tag:           os.Getenv("TRAVIS_TAG"),
			Buildnum:      os.Getenv("TRAVIS_BUILD_NUMBER"),
			IsPullRequest: os.Getenv("TRAVIS_PULL_REQUEST") != "false",
			IsCronJob:     os.Getenv("TRAVIS_EVENT_TYPE") == "cron",
		}
	case isTrueEnv("CI") && isTrueEnv("APPVEYOR"):
		commit := os.Getenv("APPVEYOR_PULL_REQUEST_HEAD_COMMIT")
		if commit == "" {
			commit = os.Getenv("APPVEYOR_REPO_COMMIT")
		}
		return &Environment{
			Name:          "appveyor",
			Repo:          os.Getenv("APPVEYOR_REPO_NAME"),
			Commit:        commit,
			Date:          getDate(commit),
			Branch:        os.Getenv("APPVEYOR_REPO_BRANCH"),
			Tag:           os.Getenv("APPVEYOR_REPO_TAG_NAME"),
			Buildnum:      os.Getenv("APPVEYOR_BUILD_NUMBER"),
			IsPullRequest: os.Getenv("APPVEYOR_PULL_REQUEST_NUMBER") != "",
			IsCronJob:     isTrueEnv("APPVEYOR_SCHEDULED_BUILD"),
		}
	default:
		return LocalEnv()
	}
}

// LocalEnv returns build environment metadata gathered from git.
func LocalEnv() *Environment {
	env := applyEnvFlags(&Environment{Name: "local", Repo: "local"})

	head := readGitFile("HEAD")
	if fields := strings.Fields(head); len(fields) == 2 {
		head = fields[1]
	} else {
		// In this case we are in "detached head" state
		// see: https://git-scm.com/docs/git-checkout#_detached_head
		// Additional check required to verify, that file contains commit hash
		commitRe := regexp.MustCompile("^([0-9a-f]{40})$")
		if commit := commitRe.FindString(head); commit != "" && env.Commit == "" {
			env.Commit = commit
		}
		return env
	}
	if env.Commit == "" {
		env.Commit = readGitFile(head)
	}
	env.Date = getDate(env.Commit)
	if env.Branch == "" {
		if head != "HEAD" {
			env.Branch = strings.TrimPrefix(head, "refs/heads/")
		}
	}
	if info, err := os.Stat(".git/objects"); err == nil && info.IsDir() && env.Tag == "" {
		env.Tag = firstLine(RunGit("tag", "-l", "--points-at", "HEAD"))
	}
	return env
}

func firstLine(s string) string {
	return strings.Split(s, "\n")[0]
}

func getDate(commit string) string {
	if commit == "" {
		return ""
	}
	out := RunGit("show", "-s", "--format=%ct", commit)
	if out == "" {
		return ""
	}
	date, err := strconv.ParseInt(strings.TrimSpace(out), 10, 64)
	if err != nil {
		panic(fmt.Sprintf("failed to parse git commit date: %v", err))
	}
	return time.Unix(date, 0).Format("20060102")
}

func applyEnvFlags(env *Environment) *Environment {
	if !flag.Parsed() {
		panic("you need to call flag.Parse before Env or LocalEnv")
	}
	if *GitCommitFlag != "" {
		env.Commit = *GitCommitFlag
	}
	if *GitBranchFlag != "" {
		env.Branch = *GitBranchFlag
	}
	if *GitTagFlag != "" {
		env.Tag = *GitTagFlag
	}
	if *BuildnumFlag != "" {
		env.Buildnum = *BuildnumFlag
	}
	if *PullRequestFlag {
		env.IsPullRequest = true
	}
	if *CronJobFlag {
		env.IsCronJob = true
	}
	return env
}

// DryRunFlag dry run flag
var DryRunFlag = flag.Bool("n", false, "dry run, don't execute commands")

// MustRun executes the given command and exits the host process for
// any error.
func MustRun(cmd *exec.Cmd) {
	fmt.Println(">>>", strings.Join(cmd.Args, " "))
	if !*DryRunFlag {
		cmd.Stderr = os.Stderr
		cmd.Stdout = os.Stdout
		if err := cmd.Run(); err != nil {
			log.Fatal(err)
		}
	}
}

var warnedAboutGit bool

// RunGit runs a git subcommand and returns its output.
// The command must complete successfully.
func RunGit(args ...string) string {
	cmd := exec.Command("git", args...)
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		if e, ok := err.(*exec.Error); ok && e.Err == exec.ErrNotFound {
			if !warnedAboutGit {
				log.Println("Warning: can't find 'git' in PATH")
				warnedAboutGit = true
			}
			return ""
		}
		log.Fatal(strings.Join(cmd.Args, " "), ": ", err, "\n", stderr.String())
	}
	return strings.TrimSpace(stdout.String())
}

// readGitFile returns content of file in .git directory.
func readGitFile(file string) string {
	content, err := ioutil.ReadFile(path.Join(".git", file))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(content))
}

// GoTool returns the command that runs a go tool. This uses go from GOROOT instead of PATH
// so that go commands executed by build use the same version of Go as the 'host' that runs
// build code. e.g.
//
//     /usr/lib/go-1.12.1/bin/go run build/ci.go ...
//
// runs using go 1.12.1 and invokes go 1.12.1 tools from the same GOROOT. This is also important
// because runtime.Version checks on the host should match the tools that are run.
func GoTool(tool string, args ...string) *exec.Cmd {
	args = append([]string{tool}, args...)
	return exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), args...)
}
