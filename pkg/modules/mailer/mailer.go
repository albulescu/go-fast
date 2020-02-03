package mailer

import (
	"bytes"
	"os"
	"text/template"

	"github.com/albulescu/go-fast/internal/core"
	"github.com/albulescu/go-fast/internal/def"
	"github.com/defval/inject/v2"
)

type MailerAdaptor interface {
	def.Mailer
}

func adaptorFactory() MailerAdaptor {
	return nil
}

func Setup() def.ModuleFactory {
	return func(di *inject.Container) interface{} {
		return &MailerModule{
			Adaptor: adaptorFactory(),
		}
	}
}

type MailerModule struct {
	core.AbstractModule
	Adaptor MailerAdaptor
}

func (m *MailerModule) Send(templateName string, params map[string]interface{}, subject, to string) error {

	tmpl, err := template.New(templateName).Parse(`
	Template: {{$.template}}
	To: {{$.to}}
	Subject: {{$.subject}}

	{{ range $k, $v := $.params }}Key:{{ $k }}, Value:{{ $v }}
	{{ end }}
	`)
	if err != nil {
		panic(err)
	}

	subjectStr := bytes.NewBufferString("")
	tplSubject, _ := template.New("subject").Parse(subject)
	tplSubject.Execute(subjectStr, params)

	err = tmpl.Execute(os.Stdout, def.M{
		"template": templateName,
		"subject":  subjectStr.String(),
		"to":       to,
		"params":   params,
	})
	if err != nil {
		panic(err)
	}

	return nil
}
