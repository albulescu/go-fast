package mailer

type MailgunAdaptor struct {
	domain     string
	privateKey string
}

func NewMailgunAdaptor(domain, privateKey string) MailerAdaptor {
	return &MailgunAdaptor{}
}

func (m *MailgunAdaptor) Send(template string, params map[string]interface{}, subject, to string) error {
	return nil
}
