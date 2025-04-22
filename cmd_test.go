package the

import (
	"context"
	"syscall"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"

	"github.com/heffcodex/the/tcfg"
)

type testConfig struct {
	tcfg.BaseConfig `mapstructure:",squash"` //nolint:tagliatelle // test
}
type testApp struct {
	*BaseApp[testConfig]
}

func newTestApp() (*testApp, error) {
	v := viper.New()
	v.SetConfigFile("config.test.yaml")

	configLoader := tcfg.NewLoader[testConfig](v)

	baseApp, err := NewBaseApp(configLoader)
	if err != nil {
		return nil, err
	}

	return &testApp{
		BaseApp: baseApp,
	}, nil
}

func TestCmdWaitInterrupt(t *testing.T) {
	t.Parallel()

	for name, intFunc := range map[string]func(cmd *cobra.Command){
		"sigint": func(*cobra.Command) { _ = syscall.Kill(syscall.Getpid(), syscall.SIGINT) },
		"soft":   func(cmd *cobra.Command) { CmdSoftInterrupt(cmd) },
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			var seq []int

			cmd := NewCmd(
				newTestApp,
				SilenceErrors(true),
				SilenceUsage(true),
				Args("run"),
				Commands(&cobra.Command{
					Use: "run",
					RunE: func(cmd *cobra.Command, _ []string) error {
						a := CmdApp[testConfig, *testApp](cmd)

						a.AddCloser(func(context.Context) error {
							seq = append(seq, 4)
							t.Log("app close 1")
							return nil
						})
						a.AddCloser(func(context.Context) error {
							seq = append(seq, 3)
							t.Log("app close 2")
							return nil
						})

						go func() {
							time.Sleep(time.Second)
							intFunc(cmd)
						}()

						CmdWaitInterrupt(cmd)

						seq = append(seq, 1)
						t.Log("got interrupt")

						time.Sleep(time.Second)

						seq = append(seq, 2)
						t.Log("returning")

						return nil
					},
				}),
			)

			require.NoError(t, cmd.Execute())
			require.Equal(t, []int{1, 2, 3, 4}, seq)
		})
	}
}
