package mailer

// NewMessage creates a new message
func (m *Mail) NewMessage(to, subject, template string, data map[string]interface{}) Message {
	return Message{
		To:       to,
		Subject:  subject,
		Template: template,
		Data:     data,
	}
}

func (m *Mail) NewMessageWithAttachments(to, subject, template string, data map[string]interface{}, attachments []string) Message {

	msg := m.NewMessage(to, subject, template, data)
	msg.Attachments = attachments

	return msg
}
