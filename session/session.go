package session

import (
	"crypto/md5"
	"fmt"
	"strings"

	"github.com/creditoreal-hub/goimobiliar/actions/login"
	"github.com/creditoreal-hub/goimobiliar/actions/logout"
	"github.com/creditoreal-hub/goimobiliar/erros"
)

type Session struct {
	SessionId      string `json:"sessionId,omitempty"`
	Endpoint       string `json:"endpoint,omitempty"`
	NomeImob       string `json:"nomeImob,omitempty"`
	ImobId         string `json:"imobId,omitempty"`
	UsuarioId      string `json:"usuarioId,omitempty"`
	Nome           string `json:"nome,omitempty"`
	Versao         string `json:"versao,omitempty"`
	ClientIP       string `json:"clientIP,omitempty"`
	CodFilial      int    `json:"codFilial,omitempty"`
	NomeFilial     string `json:"nomeFilial,omitempty"`
	Cidade         string `json:"cidade,omitempty"`
	Uf             string `json:"uf,omitempty"`
	MaxSessions    int    `json:"maxSessions,omitempty"`
	ServerDateTime string `json:"serverDateTime,omitempty"`
}

type NewInput struct {
	Endpoint string
	ImobId   string
	UserId   string
	UserPass string
}

func NewSession(input *NewInput) (*Session, error) {
	if input.UserId == "" {
		return nil, erros.ErrCampoVazio("userId")
	} else if input.UserPass == "" {
		return nil, erros.ErrCampoVazio("userPass")
	}

	password := fmt.Sprintf("%x", md5.Sum([]byte(strings.ToUpper(input.UserPass))))

	loginResponse, err := login.Run(&login.RunInput{
		Endpoint: input.Endpoint,
		ActionInput: &login.ActionInput{
			UserId:   &input.UserId,
			UserPass: &password,
			ImobId:   &input.ImobId,
		},
	})
	if err != nil {
		return nil, err
	}

	sess := Session{
		SessionId:      loginResponse.Header.SessionId,
		Endpoint:       input.Endpoint,
		NomeImob:       *loginResponse.Body.NomeImob,
		ImobId:         input.ImobId,
		UsuarioId:      *loginResponse.Body.UsuarioId,
		Nome:           *loginResponse.Body.Nome,
		Versao:         *loginResponse.Body.Versao,
		ClientIP:       *loginResponse.Body.ClientIP,
		CodFilial:      *loginResponse.Body.CodFilial,
		NomeFilial:     *loginResponse.Body.NomeFilial,
		Cidade:         *loginResponse.Body.Cidade,
		Uf:             *loginResponse.Body.Uf,
		MaxSessions:    *loginResponse.Body.MaxSessions,
		ServerDateTime: *loginResponse.Body.ServerDateTime,
	}

	return &sess, nil
}

func (s *Session) EndSession() error {
	if s == nil || s.SessionId == "" {
		return nil
	}

	_, err := logout.Run(&logout.RunInput{
		Endpoint:  s.Endpoint,
		SessionId: s.SessionId,
	})
	if err != nil {
		return err
	}

	s = nil

	return nil
}
