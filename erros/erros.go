package erros

import (
	"encoding/json"
	"errors"
	"fmt"
)

type RequestResponse struct {
	Header Header `json:"Header,omitempty"`
	Body   Body   `json:"Body,omitempty"`
}

type Header struct {
	SessionID string `json:"SessionId,omitempty"`
	Action    string `json:"Action,omitempty"`
	Status    string `json:"Status,omitempty"`
	Error     bool   `json:"Error,omitempty"`
	ErrorCode int    `json:"ErrorCode,omitempty"`
}

type Body struct {
	Erros []*Erro `json:"Erros,omitempty"`
}

type Erro struct {
	Campo    string `json:"Campo,omitempty"`
	Mensagem string `json:"Mensagem,omitempty"`
}

func CheckResponseError(body *[]byte) error {
	if string(*body) == "468 - session expired, new login required" {
		return errors.New("imobiliar: sessão inválida")
	}

	var response map[string]map[string]interface{}
	if err := json.Unmarshal(*body, &response); err != nil {
		return err
	}

	hasError, ok := response["Header"]["Error"].(bool)
	if !ok {
		return errors.New("failed type assertion on 'Header.Error'")
	}

	if hasError {
		var responseBody RequestResponse
		if err := json.Unmarshal(*body, &responseBody); err != nil {
			return err
		}

		if len(responseBody.Body.Erros) <= 0 {
			return errors.New("imobiliar: erro lançado, mas nenhum encontrado")
		}

		return fmt.Errorf("%v: %v", responseBody.Body.Erros[0].Mensagem, responseBody.Body.Erros[0].Campo)
	}

	return nil
}
