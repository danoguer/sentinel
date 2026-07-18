package contextbuilder

type Collector interface {
	Name() string

	Supports() bool

	Collect() (Context, error)
}

var Registry = map[string]Collector{
	"nginx":    &NginxCollector{},
	"docker":   &DockerCollector{},

}
