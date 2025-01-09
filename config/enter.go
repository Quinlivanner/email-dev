package config

type Config struct {
	API               API                `yaml:"Api"`
	AWS               AWS                `yaml:"Aws"`
	Psql              Psql               `yaml:"Psql"`
	System            System             `yaml:"System"`
	Jwt               Jwt                `yaml:"Jwt"`
	Logger            Logger             `yaml:"Logger"`
	DatabseTableNames DatabaseTableNames `yaml:"DatabaseTableNames"`
	RequestsApi       AIRequestsApi      `yaml:"AIRequestsApi"`
	OtherInfo         OtherInfo          `yaml:"OtherInfo"`
	Dovecot           DOVECOT            `yaml:"Dovecot"`
}
