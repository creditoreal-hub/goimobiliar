# goimobiliar

Biblioteca em Go para consumir o webservice Imobiliar de forma simples e padronizada.
Ela expõe actions organizadas por pacote com entradas e saídas tipadas, além de utilitários de autenticação de sessão, execução unitária (Run) e em lote (RunMulti) com controle de paralelismo.

## Instalação

```bash
go get github.com/creditoreal-hub/goimobiliar
```

## Autenticação

Antes de consumir qualquer action, crie uma sessão autenticada:

```go
package main

import (
	"fmt"

	"github.com/creditoreal-hub/goimobiliar/session"
)

func main() {
	sess, err := session.NewSession(&session.NewInput{
		Endpoint: "http://base.imobiliar.com.br:porta/webservice/Imobiliar2",
		ImobId:   "IMOB_ID",
		UserId:   "USUARIO",
		UserPass: "SENHA", // A biblioteca gera o hash MD5 da senha automaticamente
	})
	if err != nil {
		panic(err)
	}
	defer sess.EndSession()

	fmt.Println("Sessão criada! ID:", sess.SessionId)
}
```

## Exemplo de uso de uma Action com Run (execução unitária)

Abaixo, um exemplo com a action CONDOM_CONDOMINIO_CONSULTAR:

```go
package main

import (
	"fmt"
	"log"

	"github.com/creditoreal-hub/goimobiliar/actions/condom_condominio_consultar"
	"github.com/creditoreal-hub/goimobiliar/session"
)

func main() {
	sess, err := session.NewSession(&session.NewInput{
		Endpoint: "http://base.imobiliar.com.br:porta/webservice/Imobiliar2",
		ImobId:   "IMOB_ID",
		UserId:   "USUARIO",
		UserPass: "SENHA",
	})
	if err != nil {
		log.Fatal(err)
	}
	defer sess.EndSession()

	out, err := condom_condominio_consultar.Run(&condom_condominio_consultar.RunInput{
		Session: sess,
		ActionInput: &condom_condominio_consultar.ActionInput{
			CodCondominio: 12345,
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Nome do condomínio: %s\n", out.NomeCondominio)
}
```

## Execução em lote com RunMulti

`RunMulti` executa a action para várias entradas, cuidando da autenticação e do encerramento de sessão por item, e pode rodar em paralelo.

- `RunMultiInput` contém os seguintes parâmetros:
  - `Parallel` (bool): indica se as entradas serão processadas em paralelo.
  - `Entries` (slice): cada item com `Endpoint`, `ImobId`, `UserId`, `UserPass` e `Input` (o `ActionInput` daquela action).

- `RunMultiOutput` retorna uma lista com um item por entrada, contendo campos como:
  - `ImobId`
  - `Success` (bool)
  - `Error.Message` (se falhou)
  - `Data` (resultado tipado da action, quando `Success == true`)

Exemplo prático com `CONDOM_CONDOMINIO_CONSULTAR`

```go
package main

import (
	"fmt"
	"log"

	"github.com/creditoreal-hub/goimobiliar/actions/condom_condominio_consultar"
	"github.com/creditoreal-hub/goimobiliar/consts"
)

func main() {
	out, err := condom_condominio_consultar.RunMulti(&condom_condominio_consultar.RunMultiInput{
		Parallel: true, // ou false para sequencial
		Entries: []*consts.RunMultiInputEntry[*condom_condominio_consultar.ActionInput]{
			{
				Endpoint: "ENDPOINT-1",
				ImobId:   "IMOB_1",
				UserId:   "USER_1",
				UserPass: "SENHA_1",
				Input: &condom_condominio_consultar.ActionInput{
					CodCondominio: 11111,
				},
			},
			{
				Endpoint: "ENDPOINT-2",
				ImobId:   "IMOB_2",
				UserId:   "USER_2",
				UserPass: "SENHA_2",
				Input: &condom_condominio_consultar.ActionInput{
					CodCondominio: 22222,
				},
			},
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	for _, entry := range *out {
		if entry.Success {
			fmt.Printf("[Imob %s] OK: %s (%s/%s)\n",
				entry.ImobId,
				entry.Data.NomeCondominio,
				entry.Data.Cidade,
				entry.Data.UF,
			)
		} else {
			fmt.Printf("[Imob %s] ERRO: %s\n", entry.ImobId, entry.Error.Message)
		}
	}
}
```

### Parâmetro `Parallel`

- `Parallel: true`
  Processa todas as entradas simultaneamente.

- `Parallel: false`
  Processa uma por vez, na ordem fornecida.

> Dica: ao usar `Parallel: true`, garanta que sua infraestrutura e o Imobiliar suportam o volume de requisições concorrentes desejado.

### FAQ

- **Posso reaproveitar a mesma sessão em várias chamadas `Run`?**

  Sim. Enquanto a sessão estiver válida, você pode chamar quantas actions precisar. Finalize ao fim.

- **`RunMulti` cria/encerra sessão automaticamente?**

  Sim. Para cada entrada, ele autentica e encerra a sessão ao terminar aquele item.

- **O que muda entre as actions?**

  Apenas o `ActionInput` (campos de entrada) e o `RunOutput` (campos de saída). O padrão de uso é o mesmo.
