package mailer

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"text/template"
	"time"

	"github.com/timchuks/monieverse/internal/config"

	"github.com/ainsleyclark/go-mail/drivers"
	apimail "github.com/ainsleyclark/go-mail/mail"
	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
)

const (
	DriverMailGun   = "mailgun"
	DriverSendGrid  = "sendgrid"
	DriverSMTP      = "smtp"
	DriverSparkPost = "sparkpost"
)

// Mail holds the information necessary to connect to an SMTP server
type Mail struct {
	Domain      string
	Templates   string
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
	API         string
	APIKey      string
	APIUrl      string
}

// Message is the type for an email message
type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	BCC         []string
	CC          []string
	Template    string
	Attachments []string
	Data        interface{}
}

func New(cfg config.Config) *Mail {
	return &Mail{
		Domain:      cfg.MailDomain,
		Templates:   cfg.MailTemplates,
		Host:        cfg.MailHost,
		Port:        cfg.MailPort,
		Username:    cfg.MailUsername,
		Password:    cfg.MailPassword,
		Encryption:  cfg.MailEncryption,
		FromAddress: cfg.MailFromAddress,
		FromName:    cfg.MailFromName,
		API:         cfg.MailAPI,
		APIKey:      cfg.MailAPIKey,
		APIUrl:      cfg.MailAPIUrl,
	}
}

// Result contains information regarding the status of the sent email message
type Result struct {
	Success bool
	Error   error
}

// Send sends an email message using correct method. If API values are set,
// it will send using the appropriate api; otherwise, it sends via smtp
func (m *Mail) Send(msg Message) error {
	if len(m.API) > 0 && len(m.APIKey) > 0 && len(m.APIUrl) > 0 && m.API != "smtp" {
		return m.ChooseAPI(msg)
	}
	return m.SendSMTPMessage(msg)
}

// ChooseAPI chooses api to use (specified in .env)
func (m *Mail) ChooseAPI(msg Message) error {
	switch m.API {
	case DriverMailGun, DriverSendGrid, DriverSparkPost, DriverSMTP:
		return m.SendUsingAPI(msg, m.API)
	default:
		return fmt.Errorf("unknown api %s; only mailgun, sparkpost or sendgrid accepted", m.API)
	}
}

// SendUsingAPI sends a message using the appropriate API. It can be called directly, if necessary.
// transport can be one of sparkpost, sendgrid, or mailgun
func (m *Mail) SendUsingAPI(msg Message, transport string) error {
	if msg.From == "" {
		msg.From = m.FromAddress
	}

	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	cfg := apimail.Config{
		URL:         m.APIUrl,
		APIKey:      m.APIKey,
		Domain:      m.Domain,
		FromAddress: msg.From,
		FromName:    msg.FromName,
	}

	var err error
	var driver apimail.Mailer

	switch transport {
	case DriverMailGun:
		driver, err = drivers.NewMailgun(cfg)
	case DriverSendGrid:
		driver, err = drivers.NewSendGrid(cfg)
	case DriverSparkPost:
		driver, err = drivers.NewSparkPost(cfg)
	default:
		driver, err = drivers.NewSMTP(cfg)
	}

	if err != nil {
		return err
	}

	formattedMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		return err
	}

	plainMessage, err := m.buildPlainTextMessage(msg)
	if err != nil {
		return err
	}

	tx := &apimail.Transmission{
		Recipients: []string{msg.To},
		CC:         msg.CC,
		BCC:        msg.BCC,
		Subject:    msg.Subject,
		HTML:       formattedMessage,
		PlainText:  plainMessage,
	}

	// add attachments
	err = m.addAPIAttachments(msg, tx)
	if err != nil {
		return err
	}

	_, err = driver.Send(tx)
	if err != nil {
		return err
	}

	return nil
}

// addAPIAttachments adds attachments, if any, to mail being sent via api
func (m *Mail) addAPIAttachments(msg Message, tx *apimail.Transmission) error {
	if len(msg.Attachments) > 0 {
		var attachments []apimail.Attachment

		for _, x := range msg.Attachments {
			var attach apimail.Attachment
			content, err := os.ReadFile(x)
			if err != nil {
				return err
			}

			fileName := filepath.Base(x)
			attach.Bytes = content
			attach.Filename = fileName
			attachments = append(attachments, attach)
		}

		tx.Attachments = attachments
	}

	return nil
}

// SendSMTPMessage builds and sends an email message using SMTP. This is called by ListenForMail,
// and can also be called directly when necessary
func (m *Mail) SendSMTPMessage(msg Message) error {
	formattedMessage, err := m.buildHTMLMessage(msg)
	if err != nil {
		return err
	}

	plainMessage, err := m.buildPlainTextMessage(msg)
	if err != nil {
		return err
	}

	if msg.From == "" {
		msg.From = m.FromAddress
	}

	if msg.FromName == "" {
		msg.FromName = m.FromName
	}

	server := mail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.Username
	server.Password = m.Password
	server.Encryption = m.getEncryption(m.Encryption)
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	smtpClient, err := server.Connect()
	if err != nil {
		return err
	}

	email := mail.NewMSG()
	email.SetFrom(msg.From).
		AddTo(msg.To).
		SetSubject(msg.Subject)

	email.AddAlternative(mail.TextPlain, plainMessage)
	email.SetBody(mail.TextHTML, formattedMessage)

	if len(msg.Attachments) > 0 {
		for _, x := range msg.Attachments {
			email.AddAttachment(x)
		}
	}

	err = email.Send(smtpClient)
	if err != nil {
		return err
	}

	return nil
}

// getEncryption returns the appropriate encryption type based on a string value
func (m *Mail) getEncryption(e string) mail.Encryption {
	switch e {
	case "tls":
		return mail.EncryptionSTARTTLS
	case "ssl":
		return mail.EncryptionSSL
	case "none":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
}

// buildHTMLMessage creates the html version of the message
func (m *Mail) buildHTMLMessage(msg Message) (string, error) {
	templateToRender := fmt.Sprintf("%s/%s.html.tmpl", m.Templates, msg.Template)

	t, err := template.New("email-html").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.Data); err != nil {
		return "", err
	}

	formattedMessage := tpl.String()
	formattedMessage, err = m.inlineCSS(formattedMessage)
	if err != nil {
		return "", err
	}

	return formattedMessage, nil
}

// buildPlainTextMessage creates the plaintext version of the message
func (m *Mail) buildPlainTextMessage(msg Message) (string, error) {
	templateToRender := fmt.Sprintf("%s/%s.plain.tmpl", m.Templates, msg.Template)

	t, err := template.New("email-plain").ParseFiles(templateToRender)
	if err != nil {
		return "", err
	}

	var tpl bytes.Buffer
	if err = t.ExecuteTemplate(&tpl, "body", msg.Data); err != nil {
		return "", err
	}

	plainMessage := tpl.String()

	return plainMessage, nil
}

// inlineCSS takes html input as a string, and inlines css where possible
func (m *Mail) inlineCSS(s string) (string, error) {
	options := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}

	prem, err := premailer.NewPremailerFromString(s, &options)
	if err != nil {
		return "", err
	}

	html, err := prem.Transform()
	if err != nil {
		return "", err
	}

	return html, nil
}
