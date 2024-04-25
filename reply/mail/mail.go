package mail

import (
	"github.com/google/uuid"
	"time"
	"strings"
	"fmt"
	"encoding/base64"
	"mime/quotedprintable"
)

type Message struct {
	ID string
	Domain string
	From string
	To string
	Subject string
	TextContent string
	HTMLContent string
	Attachments []Attachment
}

type Attachment struct {
	ContentType string
	Filename string
	Content string
}

func NewMessage(domain string, from string, to string) *Message {
	return &Message{
		ID: uuid.New().String(),
		Domain: domain,
		From: from,
		To: to,
	}
}

func (m *Message) AddAttachment(a Attachment) {
	m.Attachments = append(m.Attachments, a)
}

func (a *Attachment) Unmarshal(buf *strings.Builder) {
	fmt.Fprintf(buf, "Content-Type: %s; name=%s\n", a.ContentType, a.Filename)
	fmt.Fprintf(buf, "Content-Description: %s\n", a.Filename)
	fmt.Fprintf(buf, "Content-Disposition: attachment; filename=%s\n", a.Filename)
	buf.WriteString("Content-Transfer-Encoding: base64\n\n")
	buf.WriteString(base64.StdEncoding.EncodeToString([]byte(a.Content)))
	buf.WriteString("\n")
}

func (m *Message) UnmarshalMixed(buf *strings.Builder) {
	buf.WriteString("Content-Type: multipart/mixed;boundary=\"mixed\"\n")
	buf.WriteString("\n")
	buf.WriteString("--mixed\n")
	if m.HTMLContent == "" {
		m.UnmarshalText(buf)
	} else {
		m.UnmarshalRelated(buf)
	}
	for _, a := range m.Attachments {
		buf.WriteString("\n--mixed\n")
		a.Unmarshal(buf)
	}
	buf.WriteString("\n--mixed--\n")
}

func (m *Message) UnmarshalRelated(buf *strings.Builder) {
	buf.WriteString("Content-Type: multipart/related;boundary=\"related\"\n")
	buf.WriteString("\n")
	buf.WriteString("\n--related\n")
	m.UnmarshalAlternative(buf)
	//buf.WriteString("\n--related\n")
	//m.UnmarshalCrest(buf)
	buf.WriteString("\n--related--\n")
}

func (m *Message) UnmarshalAlternative(buf *strings.Builder) {
	buf.WriteString("Content-Type: multipart/alternative;boundary=\"alternative\"\n")
	buf.WriteString("\n")	
	buf.WriteString("--alternative\n")
	m.UnmarshalText(buf)
	buf.WriteString("\n--alternative\n")
	m.UnmarshalHTML(buf)
	buf.WriteString("\n--alternative--\n")
}

/*func (m *Message) UnmarshalCrest(buf *strings.Builder) {
	buf.WriteString("Content-Type: image/png; name=\"hmrc_crest\"\n")
	buf.WriteString("Content-Transfer-Encoding: base64\n")
	buf.WriteString("Content-ID: <hmrc_crest>\n")
	buf.WriteString("Content-Disposition: inline; filename=\"hmrc_crest.png\">\n")
	buf.WriteString("\n")
	buf.WriteString(m.Crest)
}*/

func (m *Message) UnmarshalHTML(buf *strings.Builder) {
	buf.WriteString("Content-Type: text/html; charset=utf-8\n")
	buf.WriteString("Content-Transfer-Encoding: quoted-printable\n")
	buf.WriteString("\n")
	w := quotedprintable.NewWriter(buf)
	w.Write([]byte(m.HTMLContent))
	w.Close()
}

func (m *Message) UnmarshalText(buf *strings.Builder) {
	buf.WriteString("Content-Type: text/plain; charset=utf-8\n")
	buf.WriteString("Content-Transfer-Encoding: quoted-printable\n")
	buf.WriteString("\n")
	w := quotedprintable.NewWriter(buf)
	w.Write([]byte(m.TextContent))
	w.Close()
}

func (m *Message) Unmarshal() []byte {
	var buf strings.Builder

	fmt.Fprintf(&buf, "From: %s\n", m.From)
	fmt.Fprintf(&buf, "To: %s\n", m.To)
	fmt.Fprintf(&buf, "Subject: %s\n", m.Subject)
	fmt.Fprintf(&buf, "Date: %s\n", time.Now().UTC().Format(time.RFC1123Z))
	fmt.Fprintf(&buf, "Message-ID: <%s@%s>\n", m.ID, m.Domain)
	buf.WriteString("MIME-Version: 1.0\n")
	if m.Attachments != nil {
		m.UnmarshalMixed(&buf)
	} else if m.HTMLContent != "" {
		m.UnmarshalRelated(&buf)
	} else {
		m.UnmarshalText(&buf)
	}
	buf.WriteString("\n")

	return []byte(buf.String())
}