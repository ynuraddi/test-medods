package smtp

type Config struct {
	Host string
	Port int

	From string

	Username string
	Password string
}
