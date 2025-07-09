package devserver

// Config is config for inner http server.
type Config struct {
	Enabled        bool   `meta:",default=true"`
	Host           string `meta:",optional"`
	Port           int    `meta:",default=6060"`
	MetricsPath    string `meta:",default=/metrics"`
	HealthPath     string `meta:",default=/healthz"`
	EnableMetrics  bool   `meta:",default=true"`
	EnablePprof    bool   `meta:",default=true"`
	HealthResponse string `meta:",default=OK"`
}
