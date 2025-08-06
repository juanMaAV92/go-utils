package env

const (
	Enviroment   = "ENVIRONMENT"
	Port         = "PORT"
	GracefulTime = "GRACEFUL_TIME"
)

const (
	OTLP_ENDPOINT = "OTLP_ENDPOINT"
)

const (
	LocalEnvironment = "local"
)

const (
	PostgresHost        = "DB_HOST_POSTGRES"
	PostgresPassword    = "DB_PASSWORD_POSTGRES"
	PostgresUser        = "DB_USER_POSTGRES"
	PostgresPort        = "DB_PORT_POSTGRES"
	PostgresName        = "DB_NAME_POSTGRES"
	PostgresMaxPoolSize = "DB_MAX_POOL_SIZE_POSTGRES"
	PostgresMaxLifeTime = "DB_MAX_LIFE_TIME_POSTGRES"
)

const (
	JWTSecretKey       = "JWT_SECRET_KEY"
	JWTAccessTokenTTL  = "JWT_ACCESS_TOKEN_TTL"
	JWTRefreshTokenTTL = "JWT_REFRESH_TOKEN_TTL"
)
