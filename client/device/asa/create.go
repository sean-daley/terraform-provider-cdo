package asa

import (
	"context"
	"fmt"
	"github.com/CiscoDevnet/terraform-provider-cdo/go-client/connector"
	"github.com/CiscoDevnet/terraform-provider-cdo/go-client/device"
	"github.com/CiscoDevnet/terraform-provider-cdo/go-client/device/asa/asaconfig"
	"github.com/CiscoDevnet/terraform-provider-cdo/go-client/internal/http"
	"github.com/CiscoDevnet/terraform-provider-cdo/go-client/internal/retry"
	"github.com/CiscoDevnet/terraform-provider-cdo/go-client/model"
	"github.com/CiscoDevnet/terraform-provider-cdo/go-client/model/devicetype"
	"github.com/CiscoDevnet/terraform-provider-cdo/go-client/model/featureflag"
	"github.com/CiscoDevnet/terraform-provider-cdo/go-client/user"
	"strings"
)

type CreateInput struct {
	Name          string
	ConnectorUid  string
	ConnectorType string
	SocketAddress string

	Username string
	Password string

	IgnoreCertificate bool
}

type CreateOutput struct {
	Uid           string          `json:"uid"`
	Name          string          `json:"Name"`
	DeviceType    devicetype.Type `json:"deviceType"`
	Host          string          `json:"host"`
	Port          string          `json:"port"`
	SocketAddress string          `json:"ipv4"`
	ConnectorType string          `json:"larType"`
	ConnectorUid  string          `json:"larUid"`
}

type Metadata struct {
	IsNewPolicyObjectModel string `json:"isNewPolicyObjectModel"` // yes it is a string, but it should be either "true" or "false" :/
}

type CreateError struct {
	Err               error
	CreatedResourceId *string
}

func (r *CreateError) Error() string {
	return r.Err.Error()
}

func NewCreateRequestInput(name, connectorUid, connectorType, socketAddress, username, password string, ignoreCertificate bool) *CreateInput {
	return &CreateInput{
		Name:              name,
		ConnectorUid:      connectorUid,
		ConnectorType:     connectorType,
		SocketAddress:     socketAddress,
		Username:          username,
		Password:          password,
		IgnoreCertificate: ignoreCertificate,
	}
}

func Create(ctx context.Context, client http.Client, createInp CreateInput) (*CreateOutput, *CreateError) {

	client.Logger.Println("creating asa device")

	// 1. create a general CDO device with device type ASA
	// 1.1 check if we should use the new asa model
	userInfo, err := user.GetTokenInfo(ctx, client, user.NewGetTokenInfoInput())
	if err != nil {
		return nil, &CreateError{
			CreatedResourceId: nil,
			Err:               err,
		}
	}
	// 1.2 set metadata according to whether we want to use new asa model
	var metadata *Metadata = nil
	if userInfo.HasFeatureFlagEnabled(featureflag.AsaConfigurationObjectMigration) {
		metadata = &Metadata{IsNewPolicyObjectModel: "true"}
	}
	// 1.3 create the device
	deviceCreateOutp, err := device.Create(ctx, client, *device.NewCreateRequestInput(
		createInp.Name, "ASA", createInp.ConnectorUid, createInp.ConnectorType, createInp.SocketAddress, false, createInp.IgnoreCertificate, metadata,
	))
	var createdResourceId *string = nil
	if deviceCreateOutp != nil {
		createdResourceId = &deviceCreateOutp.Uid
	}
	if err != nil {
		return nil, &CreateError{
			CreatedResourceId: createdResourceId,
			Err:               err,
		}
	}

	// 2. wait for creation to be done, by waiting until asa specific device state is done
	client.Logger.Println("reading specific device uid")

	asaReadSpecOutp, err := device.ReadSpecific(ctx, client, *device.NewReadSpecificInput(
		deviceCreateOutp.Uid,
	))
	if err != nil {
		return nil, &CreateError{
			CreatedResourceId: createdResourceId,
			Err:               err,
		}
	}

	client.Logger.Println("waiting for asa config state done")

	// poll until asa config state done
	err = retry.Do(asaconfig.UntilStateDone(ctx, client, asaReadSpecOutp.SpecificUid), *retry.NewOptionsWithLogger(client.Logger))

	// error during polling, but we maybe able to handle it
	if err != nil {
		// no idea what I am doing here, but that is what the cdo frontend ui is doing
		if createInp.IgnoreCertificate {
			// update device with ignore certificate
			client.Logger.Println("retrying with ignore certificate")
			_, err := device.Update(ctx, client, device.UpdateInput{
				Uid:               deviceCreateOutp.Uid,
				IgnoreCertificate: true,
			})
			if err != nil {
				return nil,
					&CreateError{
						CreatedResourceId: createdResourceId,
						Err:               fmt.Errorf("error while updating config to ignore certificate, cause=%w", err),
					}

			}
		} else {
			return nil, &CreateError{
				CreatedResourceId: createdResourceId,
				Err:               err,
			}
		}
	}

	// 3. encrypt credentials for on prem connector (sdc)
	var publicKey *model.PublicKey
	if strings.EqualFold(deviceCreateOutp.ConnectorType, "SDC") {

		// on-prem connector requires encryption
		client.Logger.Println("decrypting public key from connector for encryption")

		if deviceCreateOutp.ConnectorUid == "" {
			return nil, &CreateError{
				CreatedResourceId: createdResourceId,
				Err:               fmt.Errorf("connector uid not found"),
			}

		}

		// read connector public key
		connectorReadRes, err := connector.ReadByUid(ctx, client, connector.ReadByUidInput{
			ConnectorUid: deviceCreateOutp.ConnectorUid,
		})
		if err != nil {
			return nil, &CreateError{
				CreatedResourceId: createdResourceId,
				Err:               err,
			}
		}
		publicKey = &connectorReadRes.PublicKey
	}

	// update asa config credentials
	client.Logger.Println("updating asa config credentials")

	asaConfigUpdateInp := asaconfig.NewUpdateInput(
		asaReadSpecOutp.SpecificUid,
		createInp.Username,
		createInp.Password,
		publicKey,
		asaReadSpecOutp.State,
	)
	_, err = asaconfig.Update(ctx, client, *asaConfigUpdateInp)
	if err != nil {
		return nil, &CreateError{
			CreatedResourceId: createdResourceId,
			Err:               err,
		}
	}

	// poll until asa config state done
	client.Logger.Println("waiting for device to reach state done")

	err = retry.Do(asaconfig.UntilStateDone(ctx, client, asaReadSpecOutp.SpecificUid), *retry.NewOptionsWithLogger(client.Logger))
	if err != nil {
		return nil, &CreateError{
			CreatedResourceId: createdResourceId,
			Err:               err,
		}
	}

	// successful
	createOutp := CreateOutput{
		Uid:           deviceCreateOutp.Uid,
		Name:          deviceCreateOutp.Name,
		DeviceType:    deviceCreateOutp.DeviceType,
		Host:          deviceCreateOutp.Host,
		Port:          deviceCreateOutp.Port,
		SocketAddress: deviceCreateOutp.SocketAddress,
		ConnectorUid:  deviceCreateOutp.ConnectorUid,
		ConnectorType: deviceCreateOutp.ConnectorType,
	}
	return &createOutp, nil
}
