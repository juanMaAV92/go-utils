package log

type Opts struct {
	field map[string]interface{}
}

func newOpts() Opts {
	return Opts{field: make(map[string]interface{})}
}

func Fields(fields map[string]interface{}) Opts {
	if fields == nil {
		fields = make(map[string]interface{})
	}
	return Opts{field: fields}
}

func Field(key string, value interface{}) Opts {
	opts := newOpts()
	opts.field[key] = value
	return opts
}

func (o *Opts) AddField(key string, value interface{}) {
	if o.field == nil {
		o.field = make(map[string]interface{})
	}
	o.field[key] = value
}
