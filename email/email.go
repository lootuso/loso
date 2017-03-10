package email

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"math/rand"
	"net/smtp"
	"strings"
	"github.com/uole/gokit/random"
)

type (
	Email struct {
		Auth     smtp.Auth
		Template *template.Template
		address  string
	}
	Message struct {
		From         string
		To           string
		CC           string
		Subject      string
		Text         string
		HTML         string
		TemplateName string
		TemplateData interface{}
		Inlines      []*File
		Attachments  []*File
		buffer       *bytes.Buffer
		boundary     string
	}

	File struct {
		Name    string
		Type    string
		Content string
	}
)

func New(address string) *Email {
	return &Email{
		address: address,
	}
}

func (m *Message) writeText(content string, contentType string) {
	m.buffer.WriteString(fmt.Sprintf("--%s\r\n", m.boundary))
	m.buffer.WriteString(fmt.Sprintf("Content-Type: %s; charset=UTF-8\r\n", contentType))
	m.buffer.WriteString("Content-Transfer-Encoding: quoted-printable\r\n")
	m.buffer.WriteString("\r\n")
	m.buffer.WriteString(content + "\r\n")
}

func (m *Message) writeFile(f *File, disposition string) {
	m.buffer.WriteString(fmt.Sprintf("--%s\r\n", m.boundary))
	m.buffer.WriteString(fmt.Sprintf("Content-Type: %s; name=%s\r\n", f.Type, f.Name))
	m.buffer.WriteString(fmt.Sprintf("Content-Disposition: %s; filename=%s\r\n", disposition, f.Name))
	m.buffer.WriteString("Content-Transfer-Encoding: base64\r\n")
	m.buffer.WriteString("\r\n")
	m.buffer.WriteString(f.Content + "\r\n")
}

func (e Email) generateBoundary(length int) string {
	b := make([]byte, length)
	charset := "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	for i := range b {
		b[i] = charset[rand.Int63()%int64(len(charset))]
	}
	return string(b)
}

func (e *Email) Send(m *Message) error {
	m.buffer = new(bytes.Buffer)
	m.boundary = random.String(16)
	m.buffer.WriteString("MIME-Version: 1.0\r\n")
	m.buffer.WriteString(fmt.Sprintf("TO: %s\r\n", m.To))
	m.buffer.WriteString(fmt.Sprintf("CC: %s\r\n", m.CC))
	m.buffer.WriteString(fmt.Sprintf("From: %s\r\n", m.From))
	m.buffer.WriteString(fmt.Sprintf("Subject: %s\r\n", m.Subject))
	m.buffer.WriteString(fmt.Sprintf("Content-Type: multipart/mixed; boundary=%s\r\n", m.boundary))
	m.buffer.WriteString("\r\n")

	// Message body
	if m.TemplateName != "" {
		buf := new(bytes.Buffer)
		if err := e.Template.ExecuteTemplate(buf, m.TemplateName, m.TemplateData); err != nil {
			return err
		}
		m.writeText(buf.String(), "text/html")
	} else if m.Text != "" {
		m.writeText(m.Text, "text/plain")
	} else if m.HTML != "" {
		m.writeText(m.HTML, "text/html")
	} else {
		return errors.New("email content can't blank")
	}
	// Attachments / inlines
	for _, f := range m.Inlines {
		m.writeFile(f, "inline")
	}
	for _, f := range m.Attachments {
		m.writeFile(f, "disposition")
	}
	m.buffer.WriteString("\r\n")
	m.buffer.WriteString("--" + m.boundary + "--")
	return smtp.SendMail(e.address, e.Auth, m.From, strings.Split(m.To, ";"), m.buffer.Bytes())
}

// send email
func Send(address, username, password string, m *Message) error {
	e := New(address)
	t := strings.Split(address, ":")
	e.Auth = smtp.PlainAuth("", username, password, t[0])
	return e.Send(m)
}
