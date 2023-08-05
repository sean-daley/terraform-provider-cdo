package ios

import (
	"context"
	"reflect"
	"testing"

	"github.com/CiscoDevnet/go-client/connector/sdc"
	"github.com/CiscoDevnet/go-client/device"
	"github.com/CiscoDevnet/go-client/device/ios/iosconfig"
	"github.com/CiscoDevnet/go-client/internal/http"
	"github.com/jarcoal/httpmock"
)

func TestIosCreate(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	sdc := sdc.NewSdcResponseBuilder().
		WithName("MyOnPremConnector").
		WithUid("88888888-8888-8888-8888-888888888888").
		WithTenantUid("00000000-0000-0000-0000-000000000000").
		AsOnPremConnector().
		Build()

	iosDevice := device.NewReadOutputBuilder().
		AsIos().
		WithUid("11111111-1111-1111-1111-111111111111").
		WithName("my-ios").
		OnboardedUsingOnPremConnector(sdc.Uid).
		WithLocation("10.10.0.1", 443).
		Build()

	testCases := []struct {
		testName   string
		input      CreateInput
		setupFunc  func(input CreateInput)
		assertFunc func(output *CreateOutput, err *CreateError, t *testing.T)
	}{
		{
			testName: "successfully onboards iOS when using SDC",
			input: CreateInput{
				Name:             iosDevice.Name,
				LarType:          iosDevice.LarType,
				LarUid:           iosDevice.LarUid,
				Ipv4:             iosDevice.Ipv4,
				Username:         "unittestuser",
				Password:         "not a real password",
				IgnoreCertifcate: false,
			},

			setupFunc: func(input CreateInput) {
				configureDeviceCreateToRespondSuccessfully(iosDevice)
				configureSdcReadToRespondSuccessfully(sdc)
				configureIosConfigReadToSucceedWithSubsequentCalls(iosDevice.Uid, []httpmock.Responder{
					httpmock.NewJsonResponderOrPanic(200, iosconfig.ReadOutput{Uid: iosDevice.Uid, State: IosStatePreReadMetadata}),
					httpmock.NewJsonResponderOrPanic(200, iosconfig.ReadOutput{Uid: iosDevice.Uid, State: IosStateDone}),
				})
				configureDeviceUpdateToRespondSuccessfully(iosDevice)
			},

			assertFunc: func(output *CreateOutput, err *CreateError, t *testing.T) {
				if err != nil {
					t.Errorf("unexpected error: %s", err.Error())
				}

				if output == nil {
					t.Fatalf("output is nil!")
				}

				expectedCreatedOutput := CreateOutput{
					Uid:        iosDevice.Uid,
					Name:       iosDevice.Name,
					DeviceType: iosDevice.DeviceType,
					Host:       iosDevice.Host,
					Port:       iosDevice.Port,
					Ipv4:       iosDevice.Ipv4,
					LarType:    iosDevice.LarType,
					LarUid:     iosDevice.LarUid,
				}
				if !reflect.DeepEqual(expectedCreatedOutput, *output) {
					t.Errorf("expected: %+v, got: %+v", expectedCreatedOutput, output)
				}

				assertDeviceCreateWasCalledOnce(t)
				assertSdcReadByUidWasCalledOnce(sdc.Uid, t)
				assertDeviceReadWasCalledTimes(iosDevice.Uid, 2, t)
				assertDeviceUpdateWasCalledOnce(iosDevice.Uid, t)
			},
		},

		{
			testName: "returns error when device create call encounters error",
			input: CreateInput{
				Name:             iosDevice.Name,
				LarType:          iosDevice.LarType,
				LarUid:           iosDevice.LarUid,
				Ipv4:             iosDevice.Ipv4,
				Username:         "unittestuser",
				Password:         "not a real password",
				IgnoreCertifcate: false,
			},

			setupFunc: func(input CreateInput) {
				configureDeviceCreateToRespondWithError()
				configureSdcReadToRespondSuccessfully(sdc)
				configureIosConfigReadToSucceedWithSubsequentCalls(iosDevice.Uid, []httpmock.Responder{
					httpmock.NewJsonResponderOrPanic(200, iosconfig.ReadOutput{Uid: iosDevice.Uid, State: IosStatePreReadMetadata}),
					httpmock.NewJsonResponderOrPanic(200, iosconfig.ReadOutput{Uid: iosDevice.Uid, State: IosStateDone}),
				})
				configureDeviceUpdateToRespondSuccessfully(iosDevice)
			},

			assertFunc: func(output *CreateOutput, err *CreateError, t *testing.T) {
				if err == nil {
					t.Error("error is nil!")
				}

				if output != nil {
					t.Errorf("expected output to be nil, got: %+v", *output)
				}
			},
		},

		{
			testName: "returns error when sdc read call encounters error",
			input: CreateInput{
				Name:             iosDevice.Name,
				LarType:          iosDevice.LarType,
				LarUid:           iosDevice.LarUid,
				Ipv4:             iosDevice.Ipv4,
				Username:         "unittestuser",
				Password:         "not a real password",
				IgnoreCertifcate: false,
			},

			setupFunc: func(input CreateInput) {
				configureDeviceCreateToRespondSuccessfully(iosDevice)
				configureSdcReadToRespondWithError(sdc.Uid)
				configureIosConfigReadToSucceedWithSubsequentCalls(iosDevice.Uid, []httpmock.Responder{
					httpmock.NewJsonResponderOrPanic(200, iosconfig.ReadOutput{Uid: iosDevice.Uid, State: IosStatePreReadMetadata}),
					httpmock.NewJsonResponderOrPanic(200, iosconfig.ReadOutput{Uid: iosDevice.Uid, State: IosStateDone}),
				})
				configureDeviceUpdateToRespondSuccessfully(iosDevice)
			},

			assertFunc: func(output *CreateOutput, err *CreateError, t *testing.T) {
				if err == nil {
					t.Error("error is nil!")
				}

				if output != nil {
					t.Errorf("expected output to be nil, got: %+v", *output)
				}
			},
		},

		{
			testName: "returns error when iOS config read call encounters error",
			input: CreateInput{
				Name:             iosDevice.Name,
				LarType:          iosDevice.LarType,
				LarUid:           iosDevice.LarUid,
				Ipv4:             iosDevice.Ipv4,
				Username:         "unittestuser",
				Password:         "not a real password",
				IgnoreCertifcate: false,
			},

			setupFunc: func(input CreateInput) {
				configureDeviceCreateToRespondSuccessfully(iosDevice)
				configureSdcReadToRespondSuccessfully(sdc)
				configureIosConfigReadToSucceedWithSubsequentCalls(iosDevice.Uid, []httpmock.Responder{
					httpmock.NewJsonResponderOrPanic(200, iosconfig.ReadOutput{Uid: iosDevice.Uid, State: iosconfig.IosConfigStateError}),
				})
				configureDeviceUpdateToRespondSuccessfully(iosDevice)
			},

			assertFunc: func(output *CreateOutput, err *CreateError, t *testing.T) {
				if err == nil {
					t.Error("error is nil!")
				}

				if output != nil {
					t.Errorf("expected output to be nil, got: %+v", *output)
				}
			},
		},

		{
			testName: "returns error when device update call encounters error",
			input: CreateInput{
				Name:             iosDevice.Name,
				LarType:          iosDevice.LarType,
				LarUid:           iosDevice.LarUid,
				Ipv4:             iosDevice.Ipv4,
				Username:         "unittestuser",
				Password:         "not a real password",
				IgnoreCertifcate: false,
			},

			setupFunc: func(input CreateInput) {
				configureDeviceCreateToRespondSuccessfully(iosDevice)
				configureSdcReadToRespondSuccessfully(sdc)
				configureIosConfigReadToSucceedWithSubsequentCalls(iosDevice.Uid, []httpmock.Responder{
					httpmock.NewJsonResponderOrPanic(200, iosconfig.ReadOutput{Uid: iosDevice.Uid, State: IosStatePreReadMetadata}),
					httpmock.NewJsonResponderOrPanic(200, iosconfig.ReadOutput{Uid: iosDevice.Uid, State: IosStateDone}),
				})
				configureDeviceUpdateToRespondWithError(iosDevice.Uid)
			},

			assertFunc: func(output *CreateOutput, err *CreateError, t *testing.T) {
				if err == nil {
					t.Error("error is nil!")
				}

				if output != nil {
					t.Errorf("expected output to be nil, got: %+v", *output)
				}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			httpmock.Reset()

			testCase.setupFunc(testCase.input)

			output, err := Create(context.Background(), *http.NewWithDefault("https://unittest.cdo.cisco.com", "a_valid_token"), testCase.input)

			testCase.assertFunc(output, err, t)
		})
	}
}