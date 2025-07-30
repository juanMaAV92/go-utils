package log

type Config = func(*config)

type config struct {
	level Level
}

func WithLevel(level Level) Config {
	return func(conf *config) {
		conf.level = level
	}
}

func applyConfig(cf ...Config) *config {
	configuration := &config{
		level: InfoLevel,
	}
	for _, c := range cf {
		c(configuration)
	}

	return configuration
}
