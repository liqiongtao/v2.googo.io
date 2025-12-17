package goohttp

var (
	DefaultConfig = &Config{
		Addr:            ":8080",
		TraceIdHeader:   DefaultTraceIdHeader,
		EnableLog:       true,
		Logger:          &DefaultLogger{},
		EnableCORS:      true,
		CORSConfig:      DefaultCORSConfig,
		EnableRateLimit: true,
		RateLimiters:    []*RateLimiter{},
	}
)

type Config struct {
	Addr            string         `yaml:"addr" json:"addr"`                           // 监听端口
	TraceIdHeader   string         `yaml:"trace_id_header" json:"trace_id_header"`     // TraceId 请求头名称，默认为 X-Request-Id
	EnableLog       bool           `yaml:"enable_log" json:"enable_log"`               // 是否启用日志
	Logger          Logger         `yaml:"-" json:"-"`                                 // 日志对象
	EnableCORS      bool           `yaml:"enable_cors" json:"enable_cors"`             // 是否启用CORS
	CORSConfig      *CORSConfig    `yaml:"cors" json:"cors"`                           // CORS配置
	EnableRateLimit bool           `yaml:"enable_rate_limit" json:"enable_rate_limit"` // 是否启用限流
	RateLimiters    []*RateLimiter `yaml:"-" json:"-"`                                 // 限流对象
	EnableEncrypt   bool           `yaml:"enable_encrypt" json:"enable_encrypt"`       // 是否启用加密传输
	Encryptor       Encryptor      `yaml:"-" json:"-"`                                 // 加解密对象
	ResponseHooks   []ResponseHook `yaml:"-" json:"-"`                                 // 响应钩子函数
}

type ConfigOption func(*Config)

func (o ConfigOption) Apply(c *Config) {
	o(c)
}

func WithAddr(addr string) ConfigOption {
	return func(c *Config) {
		c.Addr = addr
	}
}

func WithTraceIdHeader(traceIdHeader string) ConfigOption {
	return func(c *Config) {
		c.TraceIdHeader = traceIdHeader
	}
}

func WithEnableLog(enableLog bool) ConfigOption {
	return func(c *Config) {
		c.EnableLog = enableLog
	}
}

func WithLogger(logger Logger) ConfigOption {
	return func(c *Config) {
		c.Logger = logger
	}
}

func WithEnableCORS(enableCORS bool) ConfigOption {
	return func(c *Config) {
		c.EnableCORS = enableCORS
	}
}

func WithCORSConfig(corsConfig *CORSConfig) ConfigOption {
	return func(c *Config) {
		c.CORSConfig = corsConfig
	}
}

func WithEnableRateLimit(enableRateLimit bool) ConfigOption {
	return func(c *Config) {
		c.EnableRateLimit = enableRateLimit
	}
}

func WithRateLimit(rateLimiters ...*RateLimiter) ConfigOption {
	return func(c *Config) {
		c.RateLimiters = rateLimiters
	}
}

func WithEnableEncrypt(enableEncrypt bool) ConfigOption {
	return func(c *Config) {
		c.EnableEncrypt = enableEncrypt
	}
}

func WithEncryptor(encryptor Encryptor) ConfigOption {
	return func(c *Config) {
		c.Encryptor = encryptor
	}
}

func WithResponseHooks(responseHooks []ResponseHook) ConfigOption {
	return func(c *Config) {
		c.ResponseHooks = responseHooks
	}
}
