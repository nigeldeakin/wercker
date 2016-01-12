package main

import (
	"io/ioutil"
	"os"
	"os/user"
	"path/filepath"
	"testing"

	"github.com/codegangsta/cli"
	"github.com/pborman/uuid"
	"github.com/stretchr/testify/suite"
)

func run(s suite.TestingSuite, gFlags []cli.Flag, cFlags []cli.Flag, action func(c *cli.Context), args []string) {
	rootLogger.SetLevel("debug")
	os.Clearenv()
	app := cli.NewApp()
	app.Flags = gFlags
	app.Commands = []cli.Command{
		{
			Name:      "test",
			ShortName: "t",
			Usage:     "test command",
			Action:    action,
			Flags:     cFlags,
		},
	}
	app.CommandNotFound = func(c *cli.Context, command string) {
		s.T().Fatalf("Command not found: %s", command)
	}
	app.Action = func(c *cli.Context) {
		s.T().Fatal("No command specified")
	}
	app.Run(args)
}

func defaultArgs(more ...string) []string {
	args := []string{
		"wercker",
		"--debug",
		"--wercker-endpoint", "http://example.com/wercker-endpoint",
		"--base-url", "http://example.com/base-url/",
		"--auth-token", "test-token",
		"--auth-token-store", "/tmp/.wercker/test-token",
		"test",
	}
	return append(args, more...)
}

type OptionsSuite struct {
	*TestSuite
}

func TestOptionsSuite(t *testing.T) {
	suiteTester := &OptionsSuite{&TestSuite{}}
	suite.Run(t, suiteTester)
}

func (s *OptionsSuite) TestGlobalOptions() {
	args := defaultArgs()
	test := func(c *cli.Context) {
		opts, err := NewGlobalOptions(c, emptyEnv())
		s.Nil(err)
		s.Equal(true, opts.Debug)
		s.Equal("http://example.com/base-url", opts.BaseURL)
		s.Equal("test-token", opts.AuthToken)
		s.Equal("/tmp/.wercker/test-token", opts.AuthTokenStore)
	}
	run(s, globalFlags, emptyFlags, test, args)
}

func (s *OptionsSuite) TestGuessAuthToken() {
	tmpFile, err := ioutil.TempFile("", "test-auth-token")
	s.Nil(err)

	token := uuid.NewRandom().String()
	_, err = tmpFile.Write([]byte(token))
	s.Nil(err)

	tokenStore := tmpFile.Name()
	defer os.Remove(tokenStore)
	defer tmpFile.Close()

	args := []string{
		"wercker",
		"--auth-token-store", tokenStore,
		"test",
	}

	test := func(c *cli.Context) {
		opts, err := NewGlobalOptions(c, emptyEnv())
		s.Nil(err)
		s.Equal(token, opts.AuthToken)
	}

	run(s, globalFlags, emptyFlags, test, args)
}

func (s *OptionsSuite) TestEmptyPipelineOptionsEmptyDir() {
	tmpDir, err := ioutil.TempDir("", "empty-directory")
	s.Nil(err)
	defer os.RemoveAll(tmpDir)

	basename := filepath.Base(tmpDir)
	currentUser, err := user.Current()
	s.Nil(err)
	username := currentUser.Username

	// Target the path
	args := defaultArgs(tmpDir)
	test := func(c *cli.Context) {
		opts, err := NewPipelineOptions(c, emptyEnv())
		s.Nil(err)
		s.Equal(basename, opts.ApplicationID)
		s.Equal(basename, opts.ApplicationName)
		s.Equal(username, opts.ApplicationOwnerName)
		s.Equal(username, opts.ApplicationStartedByName)
		s.Equal(tmpDir, opts.ProjectPath)
		s.Equal(basename, opts.ProjectID)
		// Pretty much all the git stuff should be empty
		s.Equal("", opts.GitBranch)
		s.Equal("", opts.GitCommit)
		s.Equal("", opts.GitDomain)
		s.Equal(username, opts.GitOwner)
		s.Equal("", opts.GitRepository)
		dumpOptions(opts)
	}
	run(s, globalFlags, pipelineFlags, test, args)
}

func (s *OptionsSuite) TestEmptyBuildOptions() {
	args := defaultArgs()
	test := func(c *cli.Context) {
		opts, err := NewBuildOptions(c, emptyEnv())
		s.Nil(err)
		s.NotEqual("", opts.BuildID)
		s.Equal(opts.BuildID, opts.PipelineID)
		s.Equal("", opts.DeployID)
	}
	run(s, globalFlags, pipelineFlags, test, args)
}

func (s *OptionsSuite) TestBuildOptions() {
	args := defaultArgs("--build-id", "fake-build-id")
	test := func(c *cli.Context) {
		opts, err := NewBuildOptions(c, emptyEnv())
		s.Nil(err)
		s.Equal("fake-build-id", opts.PipelineID)
		s.Equal("fake-build-id", opts.BuildID)
		s.Equal("", opts.DeployID)
	}
	run(s, globalFlags, pipelineFlags, test, args)
}

func (s *OptionsSuite) TestEmptyDeployOptions() {
	args := defaultArgs()
	test := func(c *cli.Context) {
		opts, err := NewDeployOptions(c, emptyEnv())
		s.Nil(err)
		s.NotEqual("", opts.DeployID)
		s.Equal(opts.DeployID, opts.PipelineID)
		s.Equal("", opts.BuildID)
	}
	run(s, globalFlags, pipelineFlags, test, args)
}

func (s *OptionsSuite) TestDeployOptions() {
	args := defaultArgs("--deploy-id", "fake-deploy-id")
	test := func(c *cli.Context) {
		opts, err := NewDeployOptions(c, emptyEnv())
		s.Nil(err)
		s.Equal("fake-deploy-id", opts.PipelineID)
		s.Equal("fake-deploy-id", opts.DeployID)
		s.Equal("", opts.BuildID)
	}
	run(s, globalFlags, pipelineFlags, test, args)
}

func (s *OptionsSuite) TestKeenOptions() {
	args := defaultArgs(
		"--keen-metrics",
		"--keen-project-id", "test-id",
		"--keen-project-write-key", "test-key",
	)
	test := func(c *cli.Context) {
		e := emptyEnv()
		gOpts, err := NewGlobalOptions(c, e)
		opts, err := NewKeenOptions(c, e, gOpts)
		s.Nil(err)
		s.Equal(true, opts.ShouldKeenMetrics)
		s.Equal("test-key", opts.KeenProjectWriteKey)
		s.Equal("test-id", opts.KeenProjectID)
	}
	run(s, globalFlags, pipelineFlags, test, args)
}

func (s *OptionsSuite) TestKeenMissingOptions() {
	test := func(c *cli.Context) {
		e := emptyEnv()
		gOpts, err := NewGlobalOptions(c, e)
		_, err = NewKeenOptions(c, e, gOpts)
		s.NotNil(err)
	}

	missingID := defaultArgs(
		"--keen-metrics",
		"--keen-project-write-key", "test-key",
	)

	missingKey := defaultArgs(
		"--keen-metrics",
		"--keen-project-id", "test-id",
	)

	run(s, globalFlags, keenFlags, test, missingID)
	run(s, globalFlags, keenFlags, test, missingKey)
}

func (s *OptionsSuite) TestReporterOptions() {
	args := defaultArgs(
		"--report",
		"--wercker-host", "http://example.com/wercker-host",
		"--wercker-token", "test-token",
	)
	test := func(c *cli.Context) {
		e := emptyEnv()
		gOpts, err := NewGlobalOptions(c, e)
		opts, err := NewReporterOptions(c, e, gOpts)
		s.Nil(err)
		s.Equal(true, opts.ShouldReport)
		s.Equal("http://example.com/wercker-host", opts.ReporterHost)
		s.Equal("test-token", opts.ReporterKey)
	}
	run(s, globalFlags, pipelineFlags, test, args)
}

func (s *OptionsSuite) TestReporterMissingOptions() {
	test := func(c *cli.Context) {
		e := emptyEnv()
		gOpts, err := NewGlobalOptions(c, e)
		_, err = NewReporterOptions(c, e, gOpts)
		s.NotNil(err)
	}

	missingHost := defaultArgs(
		"--report",
		"--wercker-token", "test-token",
	)

	missingKey := defaultArgs(
		"--report",
		"--wercker-host", "http://example.com/wercker-host",
	)

	run(s, globalFlags, reporterFlags, test, missingHost)
	run(s, globalFlags, reporterFlags, test, missingKey)
}

func (s *OptionsSuite) TestTagEscaping() {
	args := defaultArgs("--tag", "feature/foo")
	test := func(c *cli.Context) {
		opts, err := NewPipelineOptions(c, emptyEnv())
		s.Nil(err)
		s.Equal("feature_foo", opts.Tag)
	}
	run(s, globalFlags, pipelineFlags, test, args)
}

func (s *OptionsSuite) TestWorkingDir() {
	tempDir, err := ioutil.TempDir("", "wercker-test-")
	s.Nil(err)
	defer os.RemoveAll(tempDir)

	args := defaultArgs("--working-dir", tempDir)

	test := func(c *cli.Context) {
		opts, err := NewPipelineOptions(c, emptyEnv())
		s.Nil(err)
		s.Equal(tempDir, opts.WorkingDir)
	}

	run(s, globalFlags, pipelineFlags, test, args)
}

func (s *OptionsSuite) TestWorkingDirCWD() {
	args := defaultArgs()
	cwd, err := filepath.Abs(".")
	s.Nil(err)

	test := func(c *cli.Context) {
		opts, err := NewPipelineOptions(c, emptyEnv())
		s.Nil(err)
		s.Equal(cwd, opts.WorkingDir)
	}

	run(s, globalFlags, pipelineFlags, test, args)
}

func (s *OptionsSuite) TestWorkingDirGetsSet() {
	tempDir, err := ioutil.TempDir("", "wercker-test-")
	s.Nil(err)
	defer os.RemoveAll(tempDir)

	// This ignores the _build part, we're only concerned about the working dir
	args := defaultArgs("--build-dir", filepath.Join(tempDir, "_build"))

	test := func(c *cli.Context) {
		opts, err := NewPipelineOptions(c, emptyEnv())
		s.Nil(err)
		s.Equal(tempDir, opts.WorkingDir)
	}

	run(s, globalFlags, pipelineFlags, test, args)
}
