package config

import "os"

type Services struct {
	RedisURL string

	MinioEndpoint string
	MinioUser     string
	MinioPassword string
	MinioBucket   string
	MinioUseSSL   bool
}

// LoadServices reads optional infra config from the environment.
func LoadServices() Services {
	return Services{
		RedisURL: os.Getenv("REDIS_URL"),

		MinioEndpoint: os.Getenv("MINIO_ENDPOINT"),
		MinioUser:     os.Getenv("MINIO_ROOT_USER"),
		MinioPassword: os.Getenv("MINIO_ROOT_PASSWORD"),
		MinioBucket:   os.Getenv("MINIO_BUCKET"),
		MinioUseSSL:   os.Getenv("MINIO_USE_SSL") == "true",
	}
}

func (s Services) RedisEnabled() bool { return s.RedisURL != "" }
func (s Services) MinioEnabled() bool { return s.MinioEndpoint != "" }
