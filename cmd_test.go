package the

import (
	"context"
	"syscall"
	"testing"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"

	"github.com/heffcodex/the/tcf"
	"github.com/heffcodex/the/tzap"
)

type testConfig struct {
	tcf.Config `mapstructure:",squash"`
}
type testApp struct {
	*App[testConfig]
}

func newTestApp() (*testApp, error) {
	v := viper.New()
	v.SetConfigFile("config.test.yaml")

	configLoader := tcf.NewLoader[testConfig](v)
	zapCoreFunc := tzap.DefaultCore(zapcore.NewConsoleEncoder)

	baseApp, err := NewApp(configLoader, zapCoreFunc)
	if err != nil {
		return nil, err
	}

	return &testApp{
		App: baseApp,
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
				Silence(),
				Args("run"),
				Commands(&cobra.Command{
					Use: "run",
					RunE: func(cmd *cobra.Command, _ []string) error {
						a := CmdApp[*testApp](cmd)

						a.D().OnClose(func(context.Context) error {
							seq = append(seq, 4)
							t.Log("close 1")
							return nil
						})
						a.D().OnClose(func(context.Context) error {
							seq = append(seq, 3)
							t.Log("close 2")
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
