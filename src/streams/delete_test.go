/**
 * (C) Copyright IBM Corp. 2020
 *
 * SPDX-License-Identifier: Apache-2.0
 */
package streams

import (
	"context"
	"errors"
	"fmt"
	"github.com/Alvearie/hri-mgmt-api/common/eventstreams"
	"github.com/Alvearie/hri-mgmt-api/common/path"
	"github.com/Alvearie/hri-mgmt-api/common/response"
	"github.com/Alvearie/hri-mgmt-api/common/test"
	es "github.com/IBM/event-streams-go-sdk-generator/build/generated"
	"github.com/golang/mock/gomock"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"
)

const (
	topicNotFoundMessage string = "Topic not found"
)

func TestDelete(t *testing.T) {
	os.Setenv(response.EnvOwActivationId, "activation123")

	tenantId := "tenant123"
	streamId1 := "dataIntegrator123.qualifier123"
	streamId2 := "dataIntegrator123"
	var streamName = strings.Join([]string{tenantId, streamId1}, ".")
	var streamNameNoQualifier = strings.Join([]string{tenantId, streamId2}, ".")

	validArgs := map[string]interface{}{
		path.ParamOwPath: fmt.Sprintf("/hri/tenants/%s/streams/%s", tenantId, streamId1),
	}
	validArgsNoQualifier := map[string]interface{}{
		path.ParamOwPath: fmt.Sprintf("/hri/tenants/%s/streams/%s", tenantId, streamId2),
	}
	missingPathArgs := map[string]interface{}{}
	missingTenantArgs := map[string]interface{}{
		path.ParamOwPath: fmt.Sprintf("/hri/tenants"),
	}
	missingStreamArgs := map[string]interface{}{
		path.ParamOwPath: fmt.Sprintf("/hri/tenants/%s/streams", tenantId),
	}

	StatusNotFound := http.Response{StatusCode: 404}
	var NotFoundError = es.ModelError{
		Message: topicNotFoundMessage,
	}

	testCases := []struct {
		name                   string
		args                   map[string]interface{}
		validatorResponse      map[string]interface{}
		modelInError           *es.ModelError
		modelNotificationError *es.ModelError
		mockResponse           *http.Response
		expectedStream         string
		expected               map[string]interface{}
	}{
		{
			name:     "missing-path",
			args:     missingPathArgs,
			expected: response.Error(http.StatusBadRequest, "Required parameter '__ow_path' is missing"),
		},
		{
			name:     "missing-tenant-id",
			args:     missingTenantArgs,
			expected: response.Error(http.StatusBadRequest, "The path is shorter than the requested path parameter; path: [ hri tenants], requested index: 3"),
		},
		{
			name:     "missing-stream-id",
			args:     missingStreamArgs,
			expected: response.Error(http.StatusBadRequest, "The path is shorter than the requested path parameter; path: [ hri tenants tenant123 streams], requested index: 5"),
		},
		{
			name:           "not-authorized",
			args:           validArgs,
			modelInError:   &es.ModelError{},
			mockResponse:   &StatusForbidden,
			expected:       response.Error(http.StatusUnauthorized, eventstreams.UnauthorizedMsg),
			expectedStream: streamName,
		},
		{
			name:           "missing-auth-header",
			args:           validArgs,
			modelInError:   &es.ModelError{},
			mockResponse:   &StatusUnauthorized,
			expected:       response.Error(http.StatusUnauthorized, eventstreams.MissingHeaderMsg),
			expectedStream: streamName,
		},
		{
			name:           "in-conn-error",
			args:           validArgs,
			modelInError:   &OtherError,
			mockResponse:   &StatusUnprocessableEntity,
			expected:       response.Error(http.StatusInternalServerError, kafkaConnectionMessage),
			expectedStream: streamName,
		},
		{
			name:                   "notification-conn-error",
			args:                   validArgs,
			modelNotificationError: &OtherError,
			mockResponse:           &StatusUnprocessableEntity,
			expected:               response.Error(http.StatusInternalServerError, kafkaConnectionMessage),
			expectedStream:         streamName,
		},
		{
			name:           "in-topic-not-found",
			args:           validArgs,
			modelInError:   &NotFoundError,
			mockResponse:   &StatusNotFound,
			expected:       response.Error(http.StatusNotFound, topicNotFoundMessage),
			expectedStream: streamName,
		},
		{
			name:                   "notification-topic-not-found",
			args:                   validArgs,
			modelNotificationError: &NotFoundError,
			mockResponse:           &StatusNotFound,
			expected:               response.Error(http.StatusNotFound, topicNotFoundMessage),
			expectedStream:         streamName,
		},
		{
			name:           "good-request-qualifier",
			args:           validArgs,
			expected:       response.Success(http.StatusOK, map[string]interface{}{}),
			expectedStream: streamName,
		},
		{
			name:           "good-request-no-qualifier",
			args:           validArgsNoQualifier,
			expected:       response.Success(http.StatusOK, map[string]interface{}{}),
			expectedStream: streamNameNoQualifier,
		},
	}

	for _, tc := range testCases {

		controller := gomock.NewController(t)
		defer controller.Finish()
		mockService := test.NewMockService(controller)

		var mockInErr error
		var mockNotificationErr error
		if tc.modelInError != nil {
			mockInErr = errors.New(tc.modelInError.Message)
		}
		if tc.modelNotificationError != nil {
			mockNotificationErr = errors.New(tc.modelNotificationError.Message)
		}

		mockService.
			EXPECT().
			DeleteTopic(context.Background(), eventstreams.TopicPrefix+tc.expectedStream+eventstreams.InSuffix).
			Return(nil, tc.mockResponse, mockInErr).
			MaxTimes(1)
		mockService.
			EXPECT().
			DeleteTopic(context.Background(), eventstreams.TopicPrefix+tc.expectedStream+eventstreams.NotificationSuffix).
			Return(nil, tc.mockResponse, mockNotificationErr).
			MaxTimes(1)

		mockService.
			EXPECT().
			HandleModelError(mockInErr).
			Return(tc.modelInError).
			MaxTimes(1)
		mockService.
			EXPECT().
			HandleModelError(mockNotificationErr).
			Return(tc.modelNotificationError).
			MaxTimes(1)

		t.Run(tc.name, func(t *testing.T) {
			if actual := Delete(tc.args, mockService); !reflect.DeepEqual(tc.expected, actual) {
				t.Error(fmt.Sprintf("Expected: [%v], actual: [%v]", tc.expected, actual))
			}
		})
	}
}
