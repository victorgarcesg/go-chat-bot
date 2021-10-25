package settings

type Config struct {
	RabbitMQ struct {
		User        string `yaml:"user"`
		Pass        string `yaml:"pass"`
		Host        string `yaml:"host"`
		Port        string `yaml:"port"`
		ClientQueue string `yaml:"clientQueue"`
		StooqQueue  string `yaml:"stooqQueue"`
	}
	Database struct {
		User       string `yaml:"user"`
		Pass       string `yaml:"pass"`
		Protocol   string `yaml:"protocol"`
		Host       string `yaml:"host"`
		Port       string `yaml:"port"`
		DataSource string `yaml:"dataSource"`
	}
}
