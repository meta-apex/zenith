package prometheus

// A Config is a prometheus config.
type Config struct {
	Host string `meta:",optional"`
	Port int    `meta:",default=9101"`
	Path string `meta:",default=/metrics"`
}
