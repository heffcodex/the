package tcfg

type Env string

const (
	EnvDev   Env = "dev"
	EnvTest  Env = "test"
	EnvStage Env = "staging"
	EnvProd  Env = "production"
)

func (e Env) String() string {
	return string(e)
}
