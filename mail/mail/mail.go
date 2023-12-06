package mail

import (
	"io"
	"net/mail"
	"mime"
	"mime/multipart"
	"encoding/base64"
	"strings"
)

type Header interface {
	Get(key string) string
}

type Message struct {
	From string
	Subject string
	TextContent string
	XMLContent string
}

func Parse(raw string) (*Message, error) {
	message, err := mail.ReadMessage(strings.NewReader(raw))
	from, err := mail.ParseAddress(message.Header.Get("From"))
	if err != nil {
		return nil, err
	}
	parts := map[string]string{}
	if err := parseBody(message.Header, message.Body, parts); err != nil {
		return nil, err
	}
	var xmlContent string
	if parts["text/xml"] != "" {
		xmlContent = parts["text/xml"]
	} else if parts["application/xml"] != "" {
		xmlContent = parts["application/xml"]		
	} else {
		xmlContent = ""
	}
	parsed := Message{
		From: from.Address,
		Subject: message.Header.Get("Subject"),
		TextContent: strings.TrimSpace(parts["text/plain"]),
		XMLContent: xmlContent,
	}
	return &parsed, nil
}

func parseBody(header Header, body io.Reader, parts map[string]string) error {
	mediaType, params, err := mime.ParseMediaType(header.Get("Content-Type"))
	if err != nil {
		return err
	}
	boundary := params["boundary"]
	switch mediaType {
	case "multipart/alternative", "multipart/related", "multipart/mixed":
		return parseMultipart(boundary, body, parts)
	case "text/plain", "text/xml", "application/xml":
		content, err := io.ReadAll(body)
		if err != nil {
			return err
		}
		if header.Get("Content-Transfer-Encoding") == "base64" {
			decoded, err := base64.StdEncoding.DecodeString(string(content))
			parts[mediaType] = string(decoded)
			return err
		} else {
			parts[mediaType] = string(content)
			return nil
		}
	default: return nil
	}
}

func parseMultipart(boundary string, body io.Reader, parts map[string]string) error {
	r := multipart.NewReader(body, boundary)
	for {
		part, err := r.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if err := parseBody(part.Header, part, parts); err != nil {
			part.Close()
			continue
		}
		part.Close()
		if len(parts) >= 2 {
			break;
		}
	}
	return nil
}