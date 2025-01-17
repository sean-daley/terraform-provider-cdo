package asaconfig

import (
	"context"
	"fmt"
	"github.com/CiscoDevnet/terraform-provider-cdo/go-client/model/statemachine/state"
	"strings"

	"github.com/CiscoDevnet/terraform-provider-cdo/go-client/internal/http"
	"github.com/CiscoDevnet/terraform-provider-cdo/go-client/internal/retry"
)

func UntilStateDone(ctx context.Context, client http.Client, specificUid string) retry.Func {

	// create asa config read request
	readReq := NewReadRequest(ctx, client, *NewReadInput(
		specificUid,
	))

	var readOutp ReadOutput

	return func() (bool, error) {
		err := readReq.Send(&readOutp)
		if err != nil {
			return false, err
		}

		client.Logger.Printf("asa config state=%s\n", readOutp.State)
		if strings.EqualFold(readOutp.State, state.DONE) {
			return true, nil
		}
		if strings.EqualFold(readOutp.State, state.ERROR) {
			return false, fmt.Errorf("workflow ended in ERROR")
		}
		if strings.EqualFold(readOutp.State, state.BAD_CREDENTIALS) {
			return false, fmt.Errorf("bad credentials")
		}
		if strings.EqualFold(readOutp.State, state.PRE_WAIT_FOR_USER_TO_UPDATE_CREDS) {
			return false, fmt.Errorf("bad credentials")
		}
		return false, nil
	}
}
