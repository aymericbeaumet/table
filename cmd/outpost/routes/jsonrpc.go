package routes

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/pion/ion-sfu/cmd/signal/json-rpc/server"
	"github.com/sourcegraph/jsonrpc2"
)

type JSONSignalAuthMiddleware struct {
	overwriteSID string
	*server.JSONSignal
}

func (p *JSONSignalAuthMiddleware) Handle(ctx context.Context, conn *jsonrpc2.Conn, req *jsonrpc2.Request) {
	replyError := func(err error) {
		_ = conn.ReplyWithError(ctx, req.ID, &jsonrpc2.Error{
			Code:    500,
			Message: fmt.Sprintf("%s", err),
		})
	}

	if req.Method == "join" {
		var join server.Join
		if err := json.Unmarshal(*req.Params, &join); err != nil {
			p.Logger.Error(err, "connect: error parsing offer")
			replyError(err)
			return
		}

		join.SID = p.overwriteSID
		messsage, err := json.Marshal(join)
		if err != nil {
			p.Logger.Error(err, "connect: error when setting session id")
			replyError(err)
			return
		}
		rawMessage := json.RawMessage(messsage)

		req.Params = &rawMessage
	}

	p.JSONSignal.Handle(ctx, conn, req)
}
