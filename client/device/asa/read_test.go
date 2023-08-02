package asa

import (
	"context"
	"reflect"
	"testing"

	"github.com/cisco-lockhart/go-client/device"
	"github.com/cisco-lockhart/go-client/internal/http"
	"github.com/jarcoal/httpmock"
)

func TestAsaRead(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	asaDevice := device.NewReadOutputBuilder().
		WithUid("11111111-1111-1111-1111-111111111111").
		WithName("my-asa").
		OnboardedUsingCdg("88888888-8888-8888-8888-888888888888").
		WithLocation("10.10.0.1", 443).
		Build()

	testCases := []struct {
		testName   string
		input      ReadInput
		setupFunc  func()
		assertFunc func(output *ReadOutput, err error, t *testing.T)
	}{
		{
			testName: "successfully reads ASA",
			input: ReadInput{
				Uid: asaDevice.Uid,
			},

			setupFunc: func() {
				configureDeviceReadToRespondSuccessfully(asaDevice)
			},

			assertFunc: func(output *ReadOutput, err error, t *testing.T) {
				if err != nil {
					t.Errorf("unexpected error: %s", err.Error())
				}

				if output == nil {
					t.Fatalf("output is nil!")
				}

				expectedReadOutput := ReadOutput{
					Uid:             asaDevice.Uid,
					Name:            asaDevice.Name,
					CreatedDate:     asaDevice.CreatedDate,
					LastUpdatedDate: asaDevice.LastUpdatedDate,
					DeviceType:      asaDevice.DeviceType,
					LarUid:          asaDevice.LarUid,
					LarType:         asaDevice.LarType,
					Ipv4:            asaDevice.Ipv4,
					Host:            asaDevice.Host,
					Port:            asaDevice.Port,
				}
				if !reflect.DeepEqual(expectedReadOutput, *output) {
					t.Errorf("expected: %+v, got: %+v", asaDevice, *output)
				}
			},
		},

		{
			testName: "returns error when the remote service reading the ASA encounters an issue",
			input: ReadInput{
				Uid: asaDevice.Uid,
			},

			setupFunc: func() {
				configureDeviceReadToRespondWithError(asaDevice.Uid)
			},

			assertFunc: func(output *ReadOutput, err error, t *testing.T) {
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

			testCase.setupFunc()

			output, err := Read(context.Background(), *http.NewWithDefault("https://unittest.cdo.cisco.com", "a_valid_token"), testCase.input)

			testCase.assertFunc(output, err, t)
		})
	}
}