package devserver

// Config is config for inner http server.
type Config struct {
	Enabled        bool   `meta:",default=true"`
	ListenOn       string `meta:",default=:5555"`
	MetricsPath    string `meta:",default=/metrics"`
	HealthPath     string `meta:",default=/health"`
	EnableMetrics  bool   `meta:",default=true"`
	EnablePprof    bool   `meta:",default=true"`
	HealthResponse string `meta:",default=OK"`
}
