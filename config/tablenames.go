package config

type DatabaseTableNames struct {
	Domains       string `yaml:"domains"`
	EmailAccounts string `yaml:"email_accounts"`
	RecivedEmails string `yaml:"received_emails"`
	SentEmails    string `yaml:"sent_emails"`
	Attachments   string `yaml:"attachments"`
}
