package smtpserver

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/SongZihuan/huan-springboard/src/config"
	"github.com/SongZihuan/huan-springboard/src/utils"
	"gopkg.in/gomail.v2"
	"net"
	"net/mail"
	"net/smtp"
	"os"
	"strings"
	"sync"
	"time"
)

var smtpAddress string = ""
var smtpUser string = ""
var smtpPassword string = ""
var smtpRecipient []*mail.Address

var once sync.Once

func InitSmtp() (err error) {
	once.Do(func() {
		recipientList := config.GetConfig().SMTP.Recipient

		smtpAddress = config.GetConfig().SMTP.Address
		smtpUser = config.GetConfig().SMTP.User
		smtpPassword = config.GetConfig().SMTP.Password
		smtpRecipient = make([]*mail.Address, 0, len(recipientList))

		if !config.IsReady() {
			panic("config is not ready")
		} else if smtpAddress == "" || smtpUser == "" {
			return
		} else if len(recipientList) == 0 {
			err = fmt.Errorf("not smt recopient")
			return
		}

		for _, rec := range recipientList {
			addr, err := mail.ParseAddress(strings.TrimSpace(rec))
			if err != nil {
				fmt.Printf("%s parser failled, ignore\n", rec)
				continue
			}

			if !utils.IsValidEmail(addr.Address) {
				fmt.Printf("%s is not a valid email, ignore\n", addr.Address)
				continue
			}

			smtpRecipient = append(smtpRecipient, addr)
		}

		if len(smtpRecipient) == 0 {
			err = fmt.Errorf("not any valid email address to be self recipient")
			return
		}
	})
	return err
}

func Send(subject string, msg string) error {
	if !config.IsReady() {
		panic("config is not ready")
	} else if smtpAddress == "" || smtpUser == "" {
		return nil
	}

	subject = fmt.Sprintf("【%s 消息提醒】 %s", config.GetConfig().SystemName, subject)
	now := time.Now()

	err := _sendTo(subject, msg, nil, nil, smtpRecipient, "", now)
	if err != nil {
		return err
	}

	return nil
}

func _sendTo(subject string, msg string, fromAddr *mail.Address, replyToAddr *mail.Address, toAddr []*mail.Address, messageID string, t time.Time) (err error) {
	if smtpAddress == "" || smtpUser == "" {
		return nil
	}

	defer func() {
		r := recover()
		if r != nil && err == nil {
			if _err, ok := r.(error); ok {
				err = _err
			} else {
				err = fmt.Errorf("panic: %v", r)
			}
		}
	}()

	sender := smtpUser

	if fromAddr == nil {
		fromAddr = &mail.Address{
			Name:    config.GetConfig().SystemName,
			Address: smtpUser,
		}
	}

	if replyToAddr == nil {
		replyToAddr = &mail.Address{
			Name:    fromAddr.Name,
			Address: fromAddr.Address,
		}
	}

	const missingPort = "missing port in address"
	host, port, err := net.SplitHostPort(smtpAddress)
	var addrErr *net.AddrError
	if errors.As(err, &addrErr) {
		if addrErr.Err == missingPort {
			host = smtpAddress
			port = "25"
		} else {
			return err
		}
	} else if err != nil {
		return err
	}

	tlsconfig := &tls.Config{
		ServerName:         host,
		InsecureSkipVerify: false,
	}

	conn, err := net.Dial("tcp", net.JoinHostPort(host, port))
	if err != nil {
		return err
	}
	defer func() {
		_ = conn.Close()
	}()

	isSecureConn := false
	_conn := tls.Client(conn, tlsconfig)
	err = _conn.Handshake()
	if err == nil {
		conn = _conn
		isSecureConn = true
	}

	smtpClient, err := smtp.NewClient(conn, host)
	if err != nil {
		return fmt.Errorf("new smtp client: %v", err)
	}
	defer func() {
		_ = smtpClient.Quit()
		smtpClient = nil
	}()

	hostname, err := os.Hostname()
	if err != nil {
		return err
	}

	if err = smtpClient.Hello(hostname); err != nil {
		return fmt.Errorf("hello: %v", err)
	}

	// If not using SMTPS, always use STARTTLS if available
	hasStartTLS, _ := smtpClient.Extension("STARTTLS")
	if !isSecureConn && hasStartTLS {
		if err = smtpClient.StartTLS(tlsconfig); err != nil {
			return fmt.Errorf("start tls: %v", err)
		}
	}

	canAuth, options := smtpClient.Extension("AUTH")
	if canAuth {
		var auth smtp.Auth
		if strings.Contains(options, "CRAM-MD5") {
			auth = smtp.CRAMMD5Auth(sender, smtpPassword)
		} else if strings.Contains(options, "PLAIN") {
			auth = smtp.PlainAuth("", sender, smtpPassword, host)
		} else if strings.Contains(options, "LOGIN") {
			auth = LoginAuth(sender, smtpPassword)
		}

		if auth != nil {
			if err = smtpClient.Auth(auth); err != nil {
				return fmt.Errorf("auth: %s", err.Error())
			}
		}
	}

	err = smtpClient.Mail(sender)
	if err != nil {
		return fmt.Errorf("mail: %v", err)
	}

	recList := make([]string, 0, len(toAddr))

	for _, addr := range toAddr {
		if addr.Address == "" || !utils.IsValidEmail(addr.Address) {
			fmt.Printf("%s is not a valid email, ignore\n", addr.Address)
			continue
		}

		err = smtpClient.Rcpt(addr.Address)
		if err != nil {
			fmt.Printf("%s set rcpt error: %s, ignore\n", addr.String(), err.Error())
			continue
		}

		recList = append(recList, addr.String())
	}

	if len(recList) == 0 {
		return fmt.Errorf("no any valid recipient")
	}

	if fromAddr.Address == "" {
		fromAddr.Address = smtpUser
	}

	gomsg := gomail.NewMessage()
	gomsg.SetHeader("From", fromAddr.String())
	gomsg.SetHeader("To", recList...)
	gomsg.SetHeader("Reply-To", replyToAddr.String())
	gomsg.SetHeader("Subject", subject)
	gomsg.SetDateHeader("Date", t)
	if messageID != "" {
		gomsg.SetHeader("In-Reply-To", messageID)
		gomsg.SetHeader("References", messageID)
	}
	gomsg.SetBody("text/plain", msg)

	w, err := smtpClient.Data()
	if err != nil {
		return fmt.Errorf("data: %v", err)
	}

	if _, err = gomsg.WriteTo(w); err != nil {
		return fmt.Errorf("write to: %v", err)
	}

	err = w.Close()
	if err != nil {
		return fmt.Errorf("close: %v", err)
	}

	return nil
}

type loginAuth struct {
	username, password string
}

func (*loginAuth) Start(_ *smtp.ServerInfo) (string, []byte, error) {
	return "LOGIN", []byte{}, nil
}

func (a *loginAuth) Next(fromServer []byte, more bool) ([]byte, error) {
	if more {
		switch string(fromServer) {
		case "Username:":
			return []byte(a.username), nil
		case "Password:":
			return []byte(a.password), nil
		default:
			return nil, fmt.Errorf("unknwon fromServer: %s", string(fromServer))
		}
	}
	return nil, nil
}

func LoginAuth(username, password string) smtp.Auth {
	return &loginAuth{username, password}
}

type Message struct {
	Info string // Message information for log purpose.
	*gomail.Message
	confirmChan chan struct{}
}
