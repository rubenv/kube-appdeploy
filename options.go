package appdeploy

type Mode int

const (
	ApplyKubernetes Mode = iota
	WriteToFolder
)

type Options struct {
	Mode         Mode
	OutputFolder string
}

var CleanTypes = []string{
	"deployment",
	"service",
	"secret",
}
