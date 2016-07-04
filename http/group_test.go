package http

import (
	"encoding/json"
	"net/http"
	"testing"

	"time"

	"bytes"
	"fmt"
	"github.com/kylelemons/godebug/pretty"
	"github.com/tecsisa/authorizr/api"
)

func TestWorkerHandler_HandleGetGroup(t *testing.T) {
	now := time.Now()
	testcases := map[string]struct {
		// API method args
		org  string
		name string
		// Expected result
		expectedStatusCode int
		expectedResponse   GetGroupNameResponse
		expectedError      api.Error
		// Manager Results
		getGroupByNameResult *api.Group
		// Manager Errors
		getGroupByNameErr error
	}{
		"OkCase": {
			org:                "org1",
			name:               "group1",
			expectedStatusCode: http.StatusOK,
			expectedResponse: GetGroupNameResponse{
				Group: &api.Group{
					ID:       "groupID",
					Name:     "group1",
					Path:     "Path",
					Urn:      "Urn",
					Org:      "Org",
					CreateAt: now,
				},
			},
			getGroupByNameResult: &api.Group{
				ID:       "groupID",
				Name:     "group1",
				Path:     "Path",
				Urn:      "Urn",
				Org:      "Org",
				CreateAt: now,
			},
		},
		"ErrorCaseGroupNotFound": {
			name:               "group1",
			expectedStatusCode: http.StatusNotFound,
			expectedError: api.Error{
				Code:    api.GROUP_BY_ORG_AND_NAME_NOT_FOUND,
				Message: "Group not found",
			},
			getGroupByNameErr: &api.Error{
				Code:    api.GROUP_BY_ORG_AND_NAME_NOT_FOUND,
				Message: "Group not found",
			},
		},
		"ErrorCaseUnauthorizedResourcesError": {
			name:               "group1",
			expectedStatusCode: http.StatusForbidden,
			expectedError: api.Error{
				Code:    api.UNAUTHORIZED_RESOURCES_ERROR,
				Message: "Unauthorized",
			},
			getGroupByNameErr: &api.Error{
				Code:    api.UNAUTHORIZED_RESOURCES_ERROR,
				Message: "Unauthorized",
			},
		},
		"ErrorCaseInvalidParameterError": {
			name:               "group1",
			expectedStatusCode: http.StatusBadRequest,
			expectedError: api.Error{
				Code:    api.INVALID_PARAMETER_ERROR,
				Message: "Invalid parameter",
			},
			getGroupByNameErr: &api.Error{
				Code:    api.INVALID_PARAMETER_ERROR,
				Message: "Invalid parameter",
			},
		},
		"ErrorCaseUnknownApiError": {
			name:               "group1",
			expectedStatusCode: http.StatusInternalServerError,
			getGroupByNameErr: &api.Error{
				Code:    api.UNKNOWN_API_ERROR,
				Message: "Error",
			},
		},
	}

	client := http.DefaultClient

	for n, test := range testcases {

		testApi.ArgsOut[GetGroupByNameMethod][0] = test.getGroupByNameResult
		testApi.ArgsOut[GetGroupByNameMethod][1] = test.getGroupByNameErr

		url := fmt.Sprintf(server.URL+API_VERSION_1+"/organizations/%v/groups/%v", test.org, test.name)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			t.Errorf("Test case %v. Unexpected error creating http request %v", n, err)
			continue
		}

		res, err := client.Do(req)
		if err != nil {
			t.Errorf("Test case %v. Unexpected error calling server %v", n, err)
			continue
		}

		// Check received parameters
		if testApi.ArgsIn[GetGroupByNameMethod][1] != test.org {
			t.Errorf("Test case %v. Received different org (wanted:%v / received:%v)", n, test.org, testApi.ArgsIn[GetGroupByNameMethod][1])
			continue
		}
		if testApi.ArgsIn[GetGroupByNameMethod][2] != test.name {
			t.Errorf("Test case %v. Received different Name (wanted:%v / received:%v)", n, test.name, testApi.ArgsIn[GetGroupByNameMethod][2])
			continue
		}

		// check status code
		if test.expectedStatusCode != res.StatusCode {
			t.Errorf("Test case %v. Received different http status code (wanted:%v / received:%v)", n, test.expectedStatusCode, res.StatusCode)
			continue
		}

		switch res.StatusCode {
		case http.StatusOK:
			getGroupNameResponse := GetGroupNameResponse{}
			err = json.NewDecoder(res.Body).Decode(&getGroupNameResponse)
			if err != nil {
				t.Errorf("Test case %v. Unexpected error parsing response %v", n, err)
				continue
			}
			// Check result
			if diff := pretty.Compare(getGroupNameResponse, test.expectedResponse); diff != "" {
				t.Errorf("Test %v failed. Received different responses (received/wanted) %v", n, diff)
				continue
			}
		case http.StatusInternalServerError: // Empty message so continue
			continue
		default:
			apiError := api.Error{}
			err = json.NewDecoder(res.Body).Decode(&apiError)
			if err != nil {
				t.Errorf("Test case %v. Unexpected error parsing error response %v", n, err)
				continue
			}
			// Check result
			if diff := pretty.Compare(apiError, test.expectedError); diff != "" {
				t.Errorf("Test %v failed. Received different error response (received/wanted) %v", n, diff)
				continue
			}

		}

	}
}

func TestWorkerHandler_HandleCreateGroup(t *testing.T) {
	now := time.Now()
	testcases := map[string]struct {
		// API method args
		org     string
		request *CreateGroupRequest
		// Expected result
		expectedStatusCode int
		expectedResponse   CreateGroupResponse
		expectedError      api.Error
		// Manager Results
		addGroupResult *api.Group
		// Manager Errors
		addGroupErr error
	}{
		"OkCase": {
			org: "org1",
			request: &CreateGroupRequest{
				Name: "group1",
				Path: "Path",
			},
			expectedStatusCode: http.StatusCreated,
			expectedResponse: CreateGroupResponse{
				Group: &api.Group{
					ID:       "GroupID",
					Name:     "group1",
					Path:     "Path",
					Urn:      "Urn",
					Org:      "org1",
					CreateAt: now,
				},
			},
			addGroupResult: &api.Group{
				ID:       "GroupID",
				Name:     "group1",
				Path:     "Path",
				Urn:      "Urn",
				Org:      "org1",
				CreateAt: now,
			},
		},
		"ErrorCaseMalformedRequest": {
			expectedStatusCode: http.StatusBadRequest,
			expectedError: api.Error{
				Code:    api.INVALID_PARAMETER_ERROR,
				Message: "EOF",
			},
		},
		"ErrorCaseGroupAlreadyExist": {
			org: "org1",
			request: &CreateGroupRequest{
				Name: "group1",
				Path: "Path",
			},
			expectedStatusCode: http.StatusConflict,
			expectedError: api.Error{
				Code:    api.GROUP_ALREADY_EXIST,
				Message: "Group already exist",
			},
			addGroupErr: &api.Error{
				Code:    api.GROUP_ALREADY_EXIST,
				Message: "Group already exist",
			},
		},
		"ErrorCaseInvalidParameterError": {
			org: "org1",
			request: &CreateGroupRequest{
				Name: "group1",
				Path: "Path",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError: api.Error{
				Code:    api.INVALID_PARAMETER_ERROR,
				Message: "Invalid Parameter",
			},
			addGroupErr: &api.Error{
				Code:    api.INVALID_PARAMETER_ERROR,
				Message: "Invalid Parameter",
			},
		},
		"ErrorCaseUnauthorizedResourcesError": {
			org: "org1",
			request: &CreateGroupRequest{
				Name: "group1",
				Path: "Path",
			},
			expectedStatusCode: http.StatusForbidden,
			expectedError: api.Error{
				Code:    api.UNAUTHORIZED_RESOURCES_ERROR,
				Message: "Unauthorized",
			},
			addGroupErr: &api.Error{
				Code:    api.UNAUTHORIZED_RESOURCES_ERROR,
				Message: "Unauthorized",
			},
		},
		"ErrorCaseUnknownApiError": {
			org: "org1",
			request: &CreateGroupRequest{
				Name: "group1",
				Path: "Path",
			},
			expectedStatusCode: http.StatusInternalServerError,
			addGroupErr: &api.Error{
				Code:    api.UNKNOWN_API_ERROR,
				Message: "Error",
			},
		},
	}

	client := http.DefaultClient

	for n, test := range testcases {

		testApi.ArgsOut[AddGroupMethod][0] = test.addGroupResult
		testApi.ArgsOut[AddGroupMethod][1] = test.addGroupErr

		var body *bytes.Buffer
		if test.request != nil {
			jsonObject, err := json.Marshal(test.request)
			if err != nil {
				t.Errorf("Test case %v. Unexpected marshalling api request %v", n, err)
				continue
			}
			body = bytes.NewBuffer(jsonObject)
		}
		if body == nil {
			body = bytes.NewBuffer([]byte{})
		}

		url := fmt.Sprintf(server.URL+API_VERSION_1+"/organizations/%v/groups", test.org)
		req, err := http.NewRequest(http.MethodPost, url, body)
		if err != nil {
			t.Errorf("Test case %v. Unexpected error creating http request %v", n, err)
			continue
		}

		res, err := client.Do(req)
		if err != nil {
			t.Errorf("Test case %v. Unexpected error calling server %v", n, err)
			continue
		}

		if test.request != nil {
			// Check received parameters
			if testApi.ArgsIn[AddGroupMethod][1] != test.org {
				t.Errorf("Test case %v. Received different Org (wanted:%v / received:%v)", n, test.org, testApi.ArgsIn[AddGroupMethod][1])
				continue
			}
			if testApi.ArgsIn[AddGroupMethod][2] != test.request.Name {
				t.Errorf("Test case %v. Received different Name (wanted:%v / received:%v)", n, test.request.Name, testApi.ArgsIn[AddGroupMethod][2])
				continue
			}
			if testApi.ArgsIn[AddGroupMethod][3] != test.request.Path {
				t.Errorf("Test case %v. Received different Path (wanted:%v / received:%v)", n, test.request.Path, testApi.ArgsIn[AddGroupMethod][3])
				continue
			}
		}

		// check status code
		if test.expectedStatusCode != res.StatusCode {
			t.Errorf("Test case %v. Received different http status code (wanted:%v / received:%v)", n, test.expectedStatusCode, res.StatusCode)
			continue
		}

		switch res.StatusCode {
		case http.StatusCreated:
			createGroupResponse := CreateGroupResponse{}
			err = json.NewDecoder(res.Body).Decode(&createGroupResponse)
			if err != nil {
				t.Errorf("Test case %v. Unexpected error parsing response %v", n, err)
				continue
			}
			// Check result
			if diff := pretty.Compare(createGroupResponse, test.expectedResponse); diff != "" {
				t.Errorf("Test %v failed. Received different responses (received/wanted) %v", n, diff)
				continue
			}
		case http.StatusInternalServerError: // Empty message so continue
			continue
		default:
			apiError := api.Error{}
			err = json.NewDecoder(res.Body).Decode(&apiError)
			if err != nil {
				t.Errorf("Test case %v. Unexpected error parsing error response %v", n, err)
				continue
			}
			// Check result
			if diff := pretty.Compare(apiError, test.expectedError); diff != "" {
				t.Errorf("Test %v failed. Received different error response (received/wanted) %v", n, diff)
				continue
			}
		}
	}
}

func TestWorkerHandler_HandleDeleteGroup(t *testing.T) {
	testcases := map[string]struct {
		// API method args
		org  string
		name string
		// Expected result
		expectedStatusCode int
		expectedError      api.Error
		// Manager Errors
		removeGroupErr error
	}{
		"OkCase": {
			org:                "org1",
			name:               "group1",
			expectedStatusCode: http.StatusNoContent,
		},
		"ErrorCaseGroupNotFound": {
			org:                "org1",
			name:               "group1",
			expectedStatusCode: http.StatusNotFound,
			expectedError: api.Error{
				Code:    api.GROUP_BY_ORG_AND_NAME_NOT_FOUND,
				Message: "Group not found",
			},
			removeGroupErr: &api.Error{
				Code:    api.GROUP_BY_ORG_AND_NAME_NOT_FOUND,
				Message: "Group not found",
			},
		},
		"ErrorCaseInvalidParameterError": {
			org:                "org1",
			name:               "InvalidID",
			expectedStatusCode: http.StatusBadRequest,
			expectedError: api.Error{
				Code:    api.INVALID_PARAMETER_ERROR,
				Message: "Invalid parameter",
			},
			removeGroupErr: &api.Error{
				Code:    api.INVALID_PARAMETER_ERROR,
				Message: "Invalid parameter",
			},
		},
		"ErrorCaseUnauthorizedResourcesError": {
			org:                "org1",
			name:               "UnauthorizedID",
			expectedStatusCode: http.StatusForbidden,
			expectedError: api.Error{
				Code:    api.UNAUTHORIZED_RESOURCES_ERROR,
				Message: "Unauthorized",
			},
			removeGroupErr: &api.Error{
				Code:    api.UNAUTHORIZED_RESOURCES_ERROR,
				Message: "Unauthorized",
			},
		},
		"ErrorCaseUnknownApiError": {
			org:                "org1",
			name:               "ExceptionID",
			expectedStatusCode: http.StatusInternalServerError,
			removeGroupErr: &api.Error{
				Code:    api.UNKNOWN_API_ERROR,
				Message: "Error",
			},
		},
	}

	client := http.DefaultClient

	for n, test := range testcases {

		testApi.ArgsOut[RemoveGroupMethod][0] = test.removeGroupErr

		url := fmt.Sprintf(server.URL+API_VERSION_1+"/organizations/%v/groups/%v", test.org, test.name)
		req, err := http.NewRequest(http.MethodDelete, url, nil)
		if err != nil {
			t.Errorf("Test case %v. Unexpected error creating http request %v", n, err)
			continue
		}

		res, err := client.Do(req)
		if err != nil {
			t.Errorf("Test case %v. Unexpected error calling server %v", n, err)
			continue
		}

		// Check received parameters
		if testApi.ArgsIn[RemoveGroupMethod][1] != test.org {
			t.Errorf("Test case %v. Received different Org (wanted:%v / received:%v)", n, test.org, testApi.ArgsIn[RemoveGroupMethod][1])
			continue
		}
		if testApi.ArgsIn[RemoveGroupMethod][2] != test.name {
			t.Errorf("Test case %v. Received different Name (wanted:%v / received:%v)", n, test.name, testApi.ArgsIn[RemoveGroupMethod][2])
			continue
		}

		// check status code
		if test.expectedStatusCode != res.StatusCode {
			t.Errorf("Test case %v. Received different http status code (wanted:%v / received:%v)", n, test.expectedStatusCode, res.StatusCode)
			continue
		}

		switch res.StatusCode {
		case http.StatusNoContent:
			// No message expected
			continue
		case http.StatusInternalServerError: // Empty message so continue
			continue
		default:
			apiError := api.Error{}
			err = json.NewDecoder(res.Body).Decode(&apiError)
			if err != nil {
				t.Errorf("Test case %v. Unexpected error parsing error response %v", n, err)
				continue
			}
			// Check result
			if diff := pretty.Compare(apiError, test.expectedError); diff != "" {
				t.Errorf("Test %v failed. Received different error response (received/wanted) %v", n, diff)
				continue
			}

		}

	}
}

func TestWorkerHandler_HandleListGroups(t *testing.T) {
	testcases := map[string]struct {
		// API method args
		org        string
		pathPrefix string
		// Expected result
		expectedStatusCode int
		expectedResponse   GetGroupsResponse
		expectedError      api.Error
		// Manager Results
		getListGroupResult []api.GroupIdentity
		// Manager Errors
		getListGroupsErr error
	}{
		"OkCase": {
			org:                "org1",
			pathPrefix:         "path",
			expectedStatusCode: http.StatusOK,
			expectedResponse: GetGroupsResponse{
				[]api.GroupIdentity{
					api.GroupIdentity{
						Org:  "org1",
						Name: "group1",
					},
				},
			},
			getListGroupResult: []api.GroupIdentity{
				api.GroupIdentity{
					Org:  "org1",
					Name: "group1",
				},
			},
		},
		"ErrorCaseUnauthorizedError": {
			org:                "org1",
			pathPrefix:         "Path",
			expectedStatusCode: http.StatusForbidden,
			expectedError: api.Error{
				Code:    api.UNAUTHORIZED_RESOURCES_ERROR,
				Message: "Unauthorized",
			},
			getListGroupsErr: &api.Error{
				Code:    api.UNAUTHORIZED_RESOURCES_ERROR,
				Message: "Unauthorized",
			},
		},
		"ErrorCaseInvalidParameterError": {
			org:                "org1",
			pathPrefix:         "Invalid",
			expectedStatusCode: http.StatusBadRequest,
			expectedError: api.Error{
				Code:    api.INVALID_PARAMETER_ERROR,
				Message: "Invalid Path",
			},
			getListGroupsErr: &api.Error{
				Code:    api.INVALID_PARAMETER_ERROR,
				Message: "Invalid Path",
			},
		},
		"ErrorCaseUnknownApiError": {
			org:                "org1",
			pathPrefix:         "path",
			expectedStatusCode: http.StatusInternalServerError,
			getListGroupsErr: &api.Error{
				Code:    api.UNKNOWN_API_ERROR,
				Message: "Error",
			},
		},
	}

	client := http.DefaultClient

	for n, test := range testcases {

		testApi.ArgsOut[GetListGroupsMethod][0] = test.getListGroupResult
		testApi.ArgsOut[GetListGroupsMethod][1] = test.getListGroupsErr

		url := fmt.Sprintf(server.URL+API_VERSION_1+"/organizations/%v/groups?PathPrefix=", test.org, test.pathPrefix)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			t.Errorf("Test case %v. Unexpected error creating http request %v", n, err)
			continue
		}

		if test.pathPrefix != "" {
			q := req.URL.Query()
			q.Add("PathPrefix", test.pathPrefix)
			req.URL.RawQuery = q.Encode()
		}

		res, err := client.Do(req)
		if err != nil {
			t.Errorf("Test case %v. Unexpected error calling server %v", n, err)
			continue
		}

		// Check received parameter
		if testApi.ArgsIn[GetListGroupsMethod][1] != test.org {
			t.Errorf("Test case %v. Received different Org (wanted:%v / received:%v)", n, test.org, testApi.ArgsIn[GetListGroupsMethod][1])
			continue
		}
		if testApi.ArgsIn[GetListGroupsMethod][2] != test.pathPrefix {
			t.Errorf("Test case %v. Received different PathPrefix (wanted:%v / received:%v)", n, test.pathPrefix, testApi.ArgsIn[GetListGroupsMethod][2])
			continue
		}

		// check status code
		if test.expectedStatusCode != res.StatusCode {
			t.Errorf("Test case %v. Received different http status code (wanted:%v / received:%v)", n, test.expectedStatusCode, res.StatusCode)
			continue
		}

		switch res.StatusCode {
		case http.StatusOK:
			getGroupsResponse := GetGroupsResponse{}
			err = json.NewDecoder(res.Body).Decode(&getGroupsResponse)
			if err != nil {
				t.Errorf("Test case %v. Unexpected error parsing response %v", n, err)
				continue
			}
			// Check result
			if diff := pretty.Compare(getGroupsResponse, test.expectedResponse); diff != "" {
				t.Errorf("Test %v failed. Received different responses (received/wanted) %v", n, diff)
				continue
			}
		case http.StatusInternalServerError: // Empty message so continue
			continue
		default:
			apiError := api.Error{}
			err = json.NewDecoder(res.Body).Decode(&apiError)
			if err != nil {
				t.Errorf("Test case %v. Unexpected error parsing error response %v", n, err)
				continue
			}
			// Check result
			if diff := pretty.Compare(apiError, test.expectedError); diff != "" {
				t.Errorf("Test %v failed. Received different error response (received/wanted) %v", n, diff)
				continue
			}

		}
	}
}

func TestWorkerHandler_HandleUpdateGroup(t *testing.T) {
	now := time.Now()
	testcases := map[string]struct {
		// API method args
		org     string
		request *UpdateGroupRequest
		// Expected result
		expectedStatusCode int
		expectedResponse   UpdateGroupResponse
		expectedError      api.Error
		// Manager Results
		updateGroupResult *api.Group
		// Manager Errors
		updateGroupErr error
	}{
		"OkCase": {
			org: "org1",
			request: &UpdateGroupRequest{
				Name: "newName",
				Path: "NewPath",
			},
			expectedStatusCode: http.StatusOK,
			expectedResponse: UpdateGroupResponse{
				Group: &api.Group{
					ID:       "GroupID",
					Name:     "group1",
					Path:     "Path",
					Urn:      "urn",
					CreateAt: now,
				},
			},
			updateGroupResult: &api.Group{
				ID:       "GroupID",
				Name:     "group1",
				Path:     "Path",
				Urn:      "urn",
				CreateAt: now,
			},
		},
		"ErrorCaseMalformedRequest": {
			expectedStatusCode: http.StatusBadRequest,
			expectedError: api.Error{
				Code:    api.INVALID_PARAMETER_ERROR,
				Message: "EOF",
			},
		},
		"ErrorCaseGroupNotFound": {
			request: &UpdateGroupRequest{
				Name: "newName",
				Path: "NewPath",
			},
			expectedStatusCode: http.StatusNotFound,
			expectedError: api.Error{
				Code:    api.GROUP_BY_ORG_AND_NAME_NOT_FOUND,
				Message: "Group not found",
			},
			updateGroupErr: &api.Error{
				Code:    api.GROUP_BY_ORG_AND_NAME_NOT_FOUND,
				Message: "Group not found",
			},
		},
		"ErrorCaseInvalidParameterError": {
			request: &UpdateGroupRequest{
				Name: "newName",
				Path: "InvalidPath",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedError: api.Error{
				Code:    api.INVALID_PARAMETER_ERROR,
				Message: "Invalid parameter",
			},
			updateGroupErr: &api.Error{
				Code:    api.INVALID_PARAMETER_ERROR,
				Message: "Invalid parameter",
			},
		},
		"ErrorCaseGroupAlreadyExistError": {
			request: &UpdateGroupRequest{
				Name: "newName",
				Path: "newPath",
			},
			expectedStatusCode: http.StatusConflict,
			expectedError: api.Error{
				Code:    api.GROUP_ALREADY_EXIST,
				Message: "Group already exist",
			},
			updateGroupErr: &api.Error{
				Code:    api.GROUP_ALREADY_EXIST,
				Message: "Group already exist",
			},
		},
		"ErrorCaseUnauthorizedResourcesError": {
			request: &UpdateGroupRequest{
				Name: "newName",
				Path: "NewPath",
			},
			expectedStatusCode: http.StatusForbidden,
			expectedError: api.Error{
				Code:    api.UNAUTHORIZED_RESOURCES_ERROR,
				Message: "Unauthorized",
			},
			updateGroupErr: &api.Error{
				Code:    api.UNAUTHORIZED_RESOURCES_ERROR,
				Message: "Unauthorized",
			},
		},
		"ErrorCaseUnknownApiError": {
			request: &UpdateGroupRequest{
				Name: "newName",
				Path: "NewPath",
			},
			expectedStatusCode: http.StatusInternalServerError,
			updateGroupErr: &api.Error{
				Code:    api.UNKNOWN_API_ERROR,
				Message: "Error",
			},
		},
	}

	client := http.DefaultClient

	for n, test := range testcases {

		testApi.ArgsOut[UpdateGroupMethod][0] = test.updateGroupResult
		testApi.ArgsOut[UpdateGroupMethod][1] = test.updateGroupErr

		var body *bytes.Buffer
		if test.request != nil {
			jsonObject, err := json.Marshal(test.request)
			if err != nil {
				t.Errorf("Test case %v. Unexpected marshalling api request %v", n, err)
				continue
			}
			body = bytes.NewBuffer(jsonObject)
		}
		if body == nil {
			body = bytes.NewBuffer([]byte{})
		}

		url := fmt.Sprintf(server.URL+API_VERSION_1+"/organizations/%v/groups/group1", test.org)
		req, err := http.NewRequest(http.MethodPut, url, body)
		if err != nil {
			t.Errorf("Test case %v. Unexpected error creating http request %v", n, err)
			continue
		}

		res, err := client.Do(req)
		if err != nil {
			t.Errorf("Test case %v. Unexpected error calling server %v", n, err)
			continue
		}

		if test.request != nil {
			// Check received parameters
			if testApi.ArgsIn[UpdateGroupMethod][1] != test.org {
				t.Errorf("Test case %v. Received different Org (wanted:%v / received:%v)", n, test.org, testApi.ArgsIn[UpdateGroupMethod][1])
				continue
			}
			if testApi.ArgsIn[UpdateGroupMethod][2] != "group1" {
				t.Errorf("Test case %v. Received different Name (wanted:%v / received:%v)", n, "group1", testApi.ArgsIn[UpdateGroupMethod][2])
				continue
			}
			if testApi.ArgsIn[UpdateGroupMethod][3] != test.request.Name {
				t.Errorf("Test case %v. Received different newName (wanted:%v / received:%v)", n, test.request.Name, testApi.ArgsIn[UpdateGroupMethod][3])
				continue
			}
			if testApi.ArgsIn[UpdateGroupMethod][4] != test.request.Path {
				t.Errorf("Test case %v. Received different Path (wanted:%v / received:%v)", n, test.request.Path, testApi.ArgsIn[UpdateGroupMethod][4])
				continue
			}

		}

		// check status code
		if test.expectedStatusCode != res.StatusCode {
			t.Errorf("Test case %v. Received different http status code (wanted:%v / received:%v)", n, test.expectedStatusCode, res.StatusCode)
			continue
		}

		switch res.StatusCode {
		case http.StatusOK:
			updateGroupResponse := UpdateGroupResponse{}
			err = json.NewDecoder(res.Body).Decode(&updateGroupResponse)
			if err != nil {
				t.Errorf("Test case %v. Unexpected error parsing response %v", n, err)
				continue
			}
			// Check result
			if diff := pretty.Compare(updateGroupResponse, test.expectedResponse); diff != "" {
				t.Errorf("Test %v failed. Received different responses (received/wanted) %v", n, diff)
				continue
			}
		case http.StatusInternalServerError: // Empty message so continue
			continue
		default:
			apiError := api.Error{}
			err = json.NewDecoder(res.Body).Decode(&apiError)
			if err != nil {
				t.Errorf("Test case %v. Unexpected error parsing error response %v", n, err)
				continue
			}
			// Check result
			if diff := pretty.Compare(apiError, test.expectedError); diff != "" {
				t.Errorf("Test %v failed. Received different error response (received/wanted) %v", n, diff)
				continue
			}

		}
	}
}

func TestWorkerHandler_HandleListMembers(t *testing.T) {
	testcases := map[string]struct {
		// API method args
		org  string
		name string
		// Expected result
		expectedStatusCode int
		expectedResponse   GetGroupMembersResponse
		expectedError      api.Error
		// Manager Results
		getListMembersResult []string
		// Manager Errors
		getListMembersErr error
	}{
		"OkCase": {
			org:                "org1",
			name:               "group1",
			expectedStatusCode: http.StatusOK,
			expectedResponse: GetGroupMembersResponse{
				Members: []string{"member1", "member2"},
			},
			getListMembersResult: []string{"member1", "member2"},
		},
		"ErrorCaseGroupNotFoundErr": {
			org:                "org1",
			name:               "group1",
			expectedStatusCode: http.StatusNotFound,
			expectedError: api.Error{
				Code:    api.GROUP_BY_ORG_AND_NAME_NOT_FOUND,
				Message: "Group Not Found",
			},
			getListMembersErr: &api.Error{
				Code:    api.GROUP_BY_ORG_AND_NAME_NOT_FOUND,
				Message: "Group Not Found",
			},
		},
		"ErrorCaseInvalidParameterErr": {
			org:                "org1",
			name:               "group1",
			expectedStatusCode: http.StatusBadRequest,
			expectedError: api.Error{
				Code:    api.INVALID_PARAMETER_ERROR,
				Message: "Invalid Parameter",
			},
			getListMembersErr: &api.Error{
				Code:    api.INVALID_PARAMETER_ERROR,
				Message: "Invalid Parameter",
			},
		},
		"ErrorCaseUnauthorizedError": {
			org:                "org1",
			name:               "group1",
			expectedStatusCode: http.StatusForbidden,
			expectedError: api.Error{
				Code:    api.UNAUTHORIZED_RESOURCES_ERROR,
				Message: "Unauthorized",
			},
			getListMembersErr: &api.Error{
				Code:    api.UNAUTHORIZED_RESOURCES_ERROR,
				Message: "Unauthorized",
			},
		},
		"ErrorCaseUnknownApiError": {
			expectedStatusCode: http.StatusInternalServerError,
			getListMembersErr: &api.Error{
				Code:    api.UNKNOWN_API_ERROR,
				Message: "Error",
			},
		},
	}

	client := http.DefaultClient

	for n, test := range testcases {

		testApi.ArgsOut[ListMembersMethod][0] = test.getListMembersResult
		testApi.ArgsOut[ListMembersMethod][1] = test.getListMembersErr

		url := fmt.Sprintf(server.URL+API_VERSION_1+"/organizations/%v/groups/%v/users", test.org, test.name)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			t.Errorf("Test case %v. Unexpected error creating http request %v", n, err)
			continue
		}

		res, err := client.Do(req)
		if err != nil {
			t.Errorf("Test case %v. Unexpected error calling server %v", n, err)
			continue
		}

		// Check received parameter
		if testApi.ArgsIn[ListMembersMethod][1] != test.org {
			t.Errorf("Test case %v. Received different Org (wanted:%v / received:%v)", n, test.org, testApi.ArgsIn[ListMembersMethod][1])
			continue
		}
		if testApi.ArgsIn[ListMembersMethod][2] != test.name {
			t.Errorf("Test case %v. Received different Name (wanted:%v / received:%v)", n, test.name, testApi.ArgsIn[ListMembersMethod][2])
			continue
		}

		// check status code
		if test.expectedStatusCode != res.StatusCode {
			t.Errorf("Test case %v. Received different http status code (wanted:%v / received:%v)", n, test.expectedStatusCode, res.StatusCode)
			continue
		}

		switch res.StatusCode {
		case http.StatusOK:
			getGroupMembersResponse := GetGroupMembersResponse{}
			err = json.NewDecoder(res.Body).Decode(&getGroupMembersResponse)
			if err != nil {
				t.Errorf("Test case %v. Unexpected error parsing response %v", n, err)
				continue
			}
			// Check result
			if diff := pretty.Compare(getGroupMembersResponse, test.expectedResponse); diff != "" {
				t.Errorf("Test %v failed. Received different responses (received/wanted) %v", n, diff)
				continue
			}
		case http.StatusInternalServerError: // Empty message so continue
			continue
		default:
			apiError := api.Error{}
			err = json.NewDecoder(res.Body).Decode(&apiError)
			if err != nil {
				t.Errorf("Test case %v. Unexpected error parsing error response %v", n, err)
				continue
			}
			// Check result
			if diff := pretty.Compare(apiError, test.expectedError); diff != "" {
				t.Errorf("Test %v failed. Received different error response (received/wanted) %v", n, diff)
				continue
			}

		}
	}
}

func TestWorkerHandler_HandleAddMember(t *testing.T) {
	testcases := map[string]struct {
		// API method args
		org       string
		userID    string
		groupName string
		// Expected result
		expectedStatusCode int
		expectedError      api.Error
		// Manager Errors
		addMemberErr error
	}{
		"OkCase": {
			org:                "org1",
			userID:             "user1",
			groupName:          "group1",
			expectedStatusCode: http.StatusNoContent,
		},
		"ErrorCaseGroupNotFoundErr": {
			org:                "org1",
			userID:             "user1",
			groupName:          "Invalid Group",
			expectedStatusCode: http.StatusNotFound,
			expectedError: api.Error{
				Code:    api.GROUP_BY_ORG_AND_NAME_NOT_FOUND,
				Message: "Group Not Found",
			},
			addMemberErr: &api.Error{
				Code:    api.GROUP_BY_ORG_AND_NAME_NOT_FOUND,
				Message: "Group Not Found",
			},
		},
		"ErrorCaseUserNotFoundErr": {
			org:                "org1",
			userID:             "Invalid User",
			groupName:          "group1",
			expectedStatusCode: http.StatusNotFound,
			expectedError: api.Error{
				Code:    api.USER_BY_EXTERNAL_ID_NOT_FOUND,
				Message: "User Not Found",
			},
			addMemberErr: &api.Error{
				Code:    api.USER_BY_EXTERNAL_ID_NOT_FOUND,
				Message: "User Not Found",
			},
		},
		"ErrorCaseUnauthorizedError": {
			org:                "org1",
			userID:             "user1",
			groupName:          "group1",
			expectedStatusCode: http.StatusForbidden,
			expectedError: api.Error{
				Code:    api.UNAUTHORIZED_RESOURCES_ERROR,
				Message: "Unauthorized",
			},
			addMemberErr: &api.Error{
				Code:    api.UNAUTHORIZED_RESOURCES_ERROR,
				Message: "Unauthorized",
			},
		},
		"ErrorCaseInvalidParameterErr": {
			org:                "org1",
			userID:             "user1",
			groupName:          "group1",
			expectedStatusCode: http.StatusBadRequest,
			expectedError: api.Error{
				Code:    api.INVALID_PARAMETER_ERROR,
				Message: "Invalid Parameter",
			},
			addMemberErr: &api.Error{
				Code:    api.INVALID_PARAMETER_ERROR,
				Message: "Invalid Parameter",
			},
		},
		"ErrorCaseUserIsAlreadyMemberErr": {
			org:                "org1",
			userID:             "user1",
			groupName:          "group1",
			expectedStatusCode: http.StatusConflict,
			expectedError: api.Error{
				Code:    api.USER_IS_ALREADY_A_MEMBER_OF_GROUP,
				Message: "User is already a member of group",
			},
			addMemberErr: &api.Error{
				Code:    api.USER_IS_ALREADY_A_MEMBER_OF_GROUP,
				Message: "User is already a member of group",
			},
		},
		"ErrorCaseUnknownApiError": {
			org:                "org1",
			userID:             "user1",
			groupName:          "group1",
			expectedStatusCode: http.StatusInternalServerError,
			addMemberErr: &api.Error{
				Code:    api.UNKNOWN_API_ERROR,
				Message: "Error",
			},
		},
	}

	client := http.DefaultClient

	for n, test := range testcases {

		testApi.ArgsOut[AddMemberMethod][0] = test.addMemberErr

		url := fmt.Sprintf(server.URL+API_VERSION_1+"/organizations/%v/groups/%v/users/%v", test.org, test.groupName, test.userID)
		req, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			t.Errorf("Test case %v. Unexpected error creating http request %v", n, err)
			continue
		}

		res, err := client.Do(req)
		if err != nil {
			t.Errorf("Test case %v. Unexpected error calling server %v", n, err)
			continue
		}

		// Check received parameters
		if testApi.ArgsIn[AddMemberMethod][1] != test.userID {
			t.Errorf("Test case %v. Received different UserID (wanted:%v / received:%v)", n, test.userID, testApi.ArgsIn[AddMemberMethod][1])
			continue
		}
		if testApi.ArgsIn[AddMemberMethod][2] != test.groupName {
			t.Errorf("Test case %v. Received different GroupName (wanted:%v / received:%v)", n, test.groupName, testApi.ArgsIn[AddMemberMethod][2])
			continue
		}
		if testApi.ArgsIn[AddMemberMethod][3] != test.org {
			t.Errorf("Test case %v. Received different Org (wanted:%v / received:%v)", n, test.org, testApi.ArgsIn[AddMemberMethod][3])
			continue
		}

		// check status code
		if test.expectedStatusCode != res.StatusCode {
			t.Errorf("Test case %v. Received different http status code (wanted:%v / received:%v)", n, test.expectedStatusCode, res.StatusCode)
			continue
		}

		switch res.StatusCode {
		case http.StatusNoContent:
			// No message expected
			continue
		case http.StatusInternalServerError: // Empty message so continue
			continue
		default:
			apiError := api.Error{}
			err = json.NewDecoder(res.Body).Decode(&apiError)
			if err != nil {
				t.Errorf("Test case %v. Unexpected error parsing error response %v", n, err)
				continue
			}
			// Check result
			if diff := pretty.Compare(apiError, test.expectedError); diff != "" {
				t.Errorf("Test %v failed. Received different error response (received/wanted) %v", n, diff)
				continue
			}

		}
	}
}

func TestWorkerHandler_HandleRemoveMember(t *testing.T) {
	testcases := map[string]struct {
		// API method args
		org       string
		userID    string
		groupName string
		// Expected result
		expectedStatusCode int
		expectedError      api.Error
		// Manager Errors
		removeMemberErr error
	}{
		"OkCase": {
			org:                "org1",
			userID:             "user1",
			groupName:          "group1",
			expectedStatusCode: http.StatusNoContent,
		},
		"ErrorCaseGroupNotFoundErr": {
			org:                "org1",
			userID:             "user1",
			groupName:          "Invalid Group",
			expectedStatusCode: http.StatusNotFound,
			expectedError: api.Error{
				Code:    api.GROUP_BY_ORG_AND_NAME_NOT_FOUND,
				Message: "Group Not Found",
			},
			removeMemberErr: &api.Error{
				Code:    api.GROUP_BY_ORG_AND_NAME_NOT_FOUND,
				Message: "Group Not Found",
			},
		},
		"ErrorCaseUserNotFoundErr": {
			org:                "org1",
			userID:             "Invalid User",
			groupName:          "group1",
			expectedStatusCode: http.StatusNotFound,
			expectedError: api.Error{
				Code:    api.USER_BY_EXTERNAL_ID_NOT_FOUND,
				Message: "User Not Found",
			},
			removeMemberErr: &api.Error{
				Code:    api.USER_BY_EXTERNAL_ID_NOT_FOUND,
				Message: "User Not Found",
			},
		},
		"ErrorCaseUserIsNotMemberErr": {
			org:                "org1",
			userID:             "user1",
			groupName:          "group1",
			expectedStatusCode: http.StatusNotFound,
			expectedError: api.Error{
				Code:    api.USER_IS_NOT_A_MEMBER_OF_GROUP,
				Message: "User is not a member",
			},
			removeMemberErr: &api.Error{
				Code:    api.USER_IS_NOT_A_MEMBER_OF_GROUP,
				Message: "User is not a member",
			},
		},
		"ErrorCaseUnauthorizedError": {
			org:                "org1",
			userID:             "user1",
			groupName:          "group1",
			expectedStatusCode: http.StatusForbidden,
			expectedError: api.Error{
				Code:    api.UNAUTHORIZED_RESOURCES_ERROR,
				Message: "Unauthorized",
			},
			removeMemberErr: &api.Error{
				Code:    api.UNAUTHORIZED_RESOURCES_ERROR,
				Message: "Unauthorized",
			},
		},
		"ErrorCaseInvalidParameterErr": {
			org:                "org1",
			userID:             "user1",
			groupName:          "group1",
			expectedStatusCode: http.StatusBadRequest,
			expectedError: api.Error{
				Code:    api.INVALID_PARAMETER_ERROR,
				Message: "Invalid Parameter",
			},
			removeMemberErr: &api.Error{
				Code:    api.INVALID_PARAMETER_ERROR,
				Message: "Invalid Parameter",
			},
		},
		"ErrorCaseUnknownApiError": {
			org:                "org1",
			userID:             "user1",
			groupName:          "group1",
			expectedStatusCode: http.StatusInternalServerError,
			removeMemberErr: &api.Error{
				Code:    api.UNKNOWN_API_ERROR,
				Message: "Error",
			},
		},
	}

	client := http.DefaultClient

	for n, test := range testcases {

		testApi.ArgsOut[RemoveMemberMethod][0] = test.removeMemberErr

		url := fmt.Sprintf(server.URL+API_VERSION_1+"/organizations/%v/groups/%v/users/%v", test.org, test.groupName, test.userID)
		req, err := http.NewRequest(http.MethodDelete, url, nil)
		if err != nil {
			t.Errorf("Test case %v. Unexpected error creating http request %v", n, err)
			continue
		}

		res, err := client.Do(req)
		if err != nil {
			t.Errorf("Test case %v. Unexpected error calling server %v", n, err)
			continue
		}

		// Check received parameters
		if testApi.ArgsIn[RemoveMemberMethod][1] != test.userID {
			t.Errorf("Test case %v. Received different UserID (wanted:%v / received:%v)", n, test.userID, testApi.ArgsIn[RemoveMemberMethod][2])
			continue
		}
		if testApi.ArgsIn[RemoveMemberMethod][2] != test.groupName {
			t.Errorf("Test case %v. Received different GroupName (wanted:%v / received:%v)", n, test.groupName, testApi.ArgsIn[RemoveMemberMethod][2])
			continue
		}
		if testApi.ArgsIn[RemoveMemberMethod][3] != test.org {
			t.Errorf("Test case %v. Received different Org (wanted:%v / received:%v)", n, test.org, testApi.ArgsIn[RemoveMemberMethod][1])
			continue
		}

		// check status code
		if test.expectedStatusCode != res.StatusCode {
			t.Errorf("Test case %v. Received different http status code (wanted:%v / received:%v)", n, test.expectedStatusCode, res.StatusCode)
			continue
		}

		switch res.StatusCode {
		case http.StatusNoContent:
			// No message expected
			continue
		case http.StatusInternalServerError: // Empty message so continue
			continue
		default:
			apiError := api.Error{}
			err = json.NewDecoder(res.Body).Decode(&apiError)
			if err != nil {
				t.Errorf("Test case %v. Unexpected error parsing error response %v", n, err)
				continue
			}
			// Check result
			if diff := pretty.Compare(apiError, test.expectedError); diff != "" {
				t.Errorf("Test %v failed. Received different error response (received/wanted) %v", n, diff)
				continue
			}

		}
	}
}

func TestWorkerHandler_HandleAttachGroupPolicy(t *testing.T) {
	testcases := map[string]struct {
		// API method args
		org        string
		groupName  string
		policyName string
		// Expected result
		expectedStatusCode int
		expectedError      api.Error
		// Manager Errors
		attachGroupPolicyErr error
	}{
		"OkCase": {
			org:                "org1",
			groupName:          "group1",
			policyName:         "policy1",
			expectedStatusCode: http.StatusNoContent,
		},
		"ErrorCaseGroupNotFoundErr": {
			org:                "org1",
			groupName:          "Invalid Group",
			policyName:         "policy1",
			expectedStatusCode: http.StatusNotFound,
			expectedError: api.Error{
				Code:    api.GROUP_BY_ORG_AND_NAME_NOT_FOUND,
				Message: "Group Not Found",
			},
			attachGroupPolicyErr: &api.Error{
				Code:    api.GROUP_BY_ORG_AND_NAME_NOT_FOUND,
				Message: "Group Not Found",
			},
		},
		"ErrorCasePolicyNotFoundErr": {
			org:                "org1",
			groupName:          "group1",
			policyName:         "Invalid Policy",
			expectedStatusCode: http.StatusNotFound,
			expectedError: api.Error{
				Code:    api.POLICY_BY_ORG_AND_NAME_NOT_FOUND,
				Message: "User Not Found",
			},
			attachGroupPolicyErr: &api.Error{
				Code:    api.POLICY_BY_ORG_AND_NAME_NOT_FOUND,
				Message: "User Not Found",
			},
		},
		"ErrorCaseUnauthorizedError": {
			org:                "org1",
			groupName:          "group1",
			policyName:         "policy1",
			expectedStatusCode: http.StatusForbidden,
			expectedError: api.Error{
				Code:    api.UNAUTHORIZED_RESOURCES_ERROR,
				Message: "Unauthorized",
			},
			attachGroupPolicyErr: &api.Error{
				Code:    api.UNAUTHORIZED_RESOURCES_ERROR,
				Message: "Unauthorized",
			},
		},
		"ErrorCaseInvalidParameterErr": {
			org:                "org1",
			groupName:          "group1",
			policyName:         "policy1",
			expectedStatusCode: http.StatusBadRequest,
			expectedError: api.Error{
				Code:    api.INVALID_PARAMETER_ERROR,
				Message: "Invalid Parameter",
			},
			attachGroupPolicyErr: &api.Error{
				Code:    api.INVALID_PARAMETER_ERROR,
				Message: "Invalid Parameter",
			},
		},
		"ErrorCasePolicyIsAlreadyAttachedErr": {
			org:                "org1",
			groupName:          "group1",
			policyName:         "policy1",
			expectedStatusCode: http.StatusConflict,
			expectedError: api.Error{
				Code:    api.POLICY_IS_ALREADY_ATTACHED_TO_GROUP,
				Message: "Policy is already attached to group",
			},
			attachGroupPolicyErr: &api.Error{
				Code:    api.POLICY_IS_ALREADY_ATTACHED_TO_GROUP,
				Message: "Policy is already attached to group",
			},
		},
		"ErrorCaseUnknownApiError": {
			org:                "org1",
			groupName:          "group1",
			policyName:         "policy1",
			expectedStatusCode: http.StatusInternalServerError,
			attachGroupPolicyErr: &api.Error{
				Code:    api.UNKNOWN_API_ERROR,
				Message: "Error",
			},
		},
	}

	client := http.DefaultClient

	for n, test := range testcases {

		testApi.ArgsOut[AttachPolicyToGroupMethod][0] = test.attachGroupPolicyErr

		url := fmt.Sprintf(server.URL+API_VERSION_1+"/organizations/%v/groups/%v/policies/%v", test.org, test.groupName, test.policyName)
		req, err := http.NewRequest(http.MethodPost, url, nil)
		if err != nil {
			t.Errorf("Test case %v. Unexpected error creating http request %v", n, err)
			continue
		}

		res, err := client.Do(req)
		if err != nil {
			t.Errorf("Test case %v. Unexpected error calling server %v", n, err)
			continue
		}

		// Check received parameters
		if testApi.ArgsIn[AttachPolicyToGroupMethod][1] != test.org {
			t.Errorf("Test case %v. Received different Org (wanted:%v / received:%v)", n, test.org, testApi.ArgsIn[AttachPolicyToGroupMethod][1])
			continue
		}
		if testApi.ArgsIn[AttachPolicyToGroupMethod][2] != test.groupName {
			t.Errorf("Test case %v. Received different GroupName (wanted:%v / received:%v)", n, test.groupName, testApi.ArgsIn[AttachPolicyToGroupMethod][2])
			continue
		}
		if testApi.ArgsIn[AttachPolicyToGroupMethod][3] != test.policyName {
			t.Errorf("Test case %v. Received different policyName (wanted:%v / received:%v)", n, test.policyName, testApi.ArgsIn[AttachPolicyToGroupMethod][3])
			continue
		}

		// check status code
		if test.expectedStatusCode != res.StatusCode {
			t.Errorf("Test case %v. Received different http status code (wanted:%v / received:%v)", n, test.expectedStatusCode, res.StatusCode)
			continue
		}

		switch res.StatusCode {
		case http.StatusNoContent:
			// No message expected
			continue
		case http.StatusInternalServerError: // Empty message so continue
			continue
		default:
			apiError := api.Error{}
			err = json.NewDecoder(res.Body).Decode(&apiError)
			if err != nil {
				t.Errorf("Test case %v. Unexpected error parsing error response %v", n, err)
				continue
			}
			// Check result
			if diff := pretty.Compare(apiError, test.expectedError); diff != "" {
				t.Errorf("Test %v failed. Received different error response (received/wanted) %v", n, diff)
				continue
			}

		}
	}
}

func TestWorkerHandler_HandleDetachGroupPolicy(t *testing.T) {
	testcases := map[string]struct {
		// API method args
		org        string
		groupName  string
		policyName string
		// Expected result
		expectedStatusCode int
		expectedError      api.Error
		// Manager Errors
		detachGroupPolicyErr error
	}{
		"OkCase": {
			org:                "org1",
			groupName:          "group1",
			policyName:         "policy1",
			expectedStatusCode: http.StatusNoContent,
		},
		"ErrorCaseGroupNotFoundErr": {
			org:                "org1",
			groupName:          "Invalid Group",
			policyName:         "policy1",
			expectedStatusCode: http.StatusNotFound,
			expectedError: api.Error{
				Code:    api.GROUP_BY_ORG_AND_NAME_NOT_FOUND,
				Message: "Group Not Found",
			},
			detachGroupPolicyErr: &api.Error{
				Code:    api.GROUP_BY_ORG_AND_NAME_NOT_FOUND,
				Message: "Group Not Found",
			},
		},
		"ErrorCasePolicyNotFoundErr": {
			org:                "org1",
			groupName:          "group1",
			policyName:         "Invalid Policy",
			expectedStatusCode: http.StatusNotFound,
			expectedError: api.Error{
				Code:    api.POLICY_BY_ORG_AND_NAME_NOT_FOUND,
				Message: "User Not Found",
			},
			detachGroupPolicyErr: &api.Error{
				Code:    api.POLICY_BY_ORG_AND_NAME_NOT_FOUND,
				Message: "User Not Found",
			},
		},
		"ErrorCasePolicyIsNotAttachedErr": {
			org:                "org1",
			groupName:          "group1",
			policyName:         "Invalid Policy",
			expectedStatusCode: http.StatusNotFound,
			expectedError: api.Error{
				Code:    api.POLICY_IS_NOT_ATTACHED_TO_GROUP,
				Message: "Policy is not attached to group",
			},
			detachGroupPolicyErr: &api.Error{
				Code:    api.POLICY_IS_NOT_ATTACHED_TO_GROUP,
				Message: "Policy is not attached to group",
			},
		},
		"ErrorCaseUnauthorizedError": {
			org:                "org1",
			groupName:          "group1",
			policyName:         "policy1",
			expectedStatusCode: http.StatusForbidden,
			expectedError: api.Error{
				Code:    api.UNAUTHORIZED_RESOURCES_ERROR,
				Message: "Unauthorized",
			},
			detachGroupPolicyErr: &api.Error{
				Code:    api.UNAUTHORIZED_RESOURCES_ERROR,
				Message: "Unauthorized",
			},
		},
		"ErrorCaseInvalidParameterErr": {
			org:                "org1",
			groupName:          "group1",
			policyName:         "policy1",
			expectedStatusCode: http.StatusBadRequest,
			expectedError: api.Error{
				Code:    api.INVALID_PARAMETER_ERROR,
				Message: "Invalid Parameter",
			},
			detachGroupPolicyErr: &api.Error{
				Code:    api.INVALID_PARAMETER_ERROR,
				Message: "Invalid Parameter",
			},
		},
		"ErrorCaseUnknownApiError": {
			org:                "org1",
			groupName:          "group1",
			policyName:         "policy1",
			expectedStatusCode: http.StatusInternalServerError,
			detachGroupPolicyErr: &api.Error{
				Code:    api.UNKNOWN_API_ERROR,
				Message: "Error",
			},
		},
	}

	client := http.DefaultClient

	for n, test := range testcases {

		testApi.ArgsOut[DetachPolicyToGroupMethod][0] = test.detachGroupPolicyErr

		url := fmt.Sprintf(server.URL+API_VERSION_1+"/organizations/%v/groups/%v/policies/%v", test.org, test.groupName, test.policyName)
		req, err := http.NewRequest(http.MethodDelete, url, nil)
		if err != nil {
			t.Errorf("Test case %v. Unexpected error creating http request %v", n, err)
			continue
		}

		res, err := client.Do(req)
		if err != nil {
			t.Errorf("Test case %v. Unexpected error calling server %v", n, err)
			continue
		}

		// Check received parameters
		if testApi.ArgsIn[DetachPolicyToGroupMethod][1] != test.org {
			t.Errorf("Test case %v. Received different Org (wanted:%v / received:%v)", n, test.org, testApi.ArgsIn[AttachPolicyToGroupMethod][1])
			continue
		}
		if testApi.ArgsIn[DetachPolicyToGroupMethod][2] != test.groupName {
			t.Errorf("Test case %v. Received different GroupName (wanted:%v / received:%v)", n, test.groupName, testApi.ArgsIn[AttachPolicyToGroupMethod][2])
			continue
		}
		if testApi.ArgsIn[DetachPolicyToGroupMethod][3] != test.policyName {
			t.Errorf("Test case %v. Received different policyName (wanted:%v / received:%v)", n, test.policyName, testApi.ArgsIn[AttachPolicyToGroupMethod][3])
			continue
		}

		// check status code
		if test.expectedStatusCode != res.StatusCode {
			t.Errorf("Test case %v. Received different http status code (wanted:%v / received:%v)", n, test.expectedStatusCode, res.StatusCode)
			continue
		}

		switch res.StatusCode {
		case http.StatusNoContent:
			// No message expected
			continue
		case http.StatusInternalServerError: // Empty message so continue
			continue
		default:
			apiError := api.Error{}
			err = json.NewDecoder(res.Body).Decode(&apiError)
			if err != nil {
				t.Errorf("Test case %v. Unexpected error parsing error response %v", n, err)
				continue
			}
			// Check result
			if diff := pretty.Compare(apiError, test.expectedError); diff != "" {
				t.Errorf("Test %v failed. Received different error response (received/wanted) %v", n, diff)
				continue
			}

		}
	}
}

func TestWorkerHandler_HandleListAttachedGroupPolicies(t *testing.T) {
	testcases := map[string]struct {
		// API method args
		org  string
		name string
		// Expected result
		expectedStatusCode int
		expectedResponse   GetGroupPoliciesResponse
		expectedError      api.Error
		// Manager Results
		getListAttachedGroupPoliciesResult []api.PolicyIdentity
		// Manager Errors
		getListAttachedGroupPoliciesErr error
	}{
		"OkCase": {
			org:                "org1",
			name:               "group1",
			expectedStatusCode: http.StatusOK,
			expectedResponse: GetGroupPoliciesResponse{
				AttachedPolicies: []api.PolicyIdentity{
					api.PolicyIdentity{
						Org:  "org1",
						Name: "policy1",
					},
					api.PolicyIdentity{
						Org:  "org1",
						Name: "policy2",
					},
				},
			},
			getListAttachedGroupPoliciesResult: []api.PolicyIdentity{
				api.PolicyIdentity{
					Org:  "org1",
					Name: "policy1",
				},
				api.PolicyIdentity{
					Org:  "org1",
					Name: "policy2",
				},
			},
		},
		"ErrorCaseGroupNotFoundErr": {
			org:                "org1",
			name:               "group1",
			expectedStatusCode: http.StatusNotFound,
			expectedError: api.Error{
				Code:    api.GROUP_BY_ORG_AND_NAME_NOT_FOUND,
				Message: "Group Not Found",
			},
			getListAttachedGroupPoliciesErr: &api.Error{
				Code:    api.GROUP_BY_ORG_AND_NAME_NOT_FOUND,
				Message: "Group Not Found",
			},
		},
		"ErrorCaseInvalidParameterErr": {
			org:                "org1",
			name:               "group1",
			expectedStatusCode: http.StatusBadRequest,
			expectedError: api.Error{
				Code:    api.INVALID_PARAMETER_ERROR,
				Message: "Invalid Parameter",
			},
			getListAttachedGroupPoliciesErr: &api.Error{
				Code:    api.INVALID_PARAMETER_ERROR,
				Message: "Invalid Parameter",
			},
		},
		"ErrorCaseUnauthorizedError": {
			org:                "org1",
			name:               "group1",
			expectedStatusCode: http.StatusForbidden,
			expectedError: api.Error{
				Code:    api.UNAUTHORIZED_RESOURCES_ERROR,
				Message: "Unauthorized",
			},
			getListAttachedGroupPoliciesErr: &api.Error{
				Code:    api.UNAUTHORIZED_RESOURCES_ERROR,
				Message: "Unauthorized",
			},
		},
		"ErrorCaseUnknownApiError": {
			expectedStatusCode: http.StatusInternalServerError,
			getListAttachedGroupPoliciesErr: &api.Error{
				Code:    api.UNKNOWN_API_ERROR,
				Message: "Error",
			},
		},
	}

	client := http.DefaultClient

	for n, test := range testcases {

		testApi.ArgsOut[ListAttachedGroupPoliciesMethod][0] = test.getListAttachedGroupPoliciesResult
		testApi.ArgsOut[ListAttachedGroupPoliciesMethod][1] = test.getListAttachedGroupPoliciesErr

		url := fmt.Sprintf(server.URL+API_VERSION_1+"/organizations/%v/groups/%v/policies", test.org, test.name)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			t.Errorf("Test case %v. Unexpected error creating http request %v", n, err)
			continue
		}

		res, err := client.Do(req)
		if err != nil {
			t.Errorf("Test case %v. Unexpected error calling server %v", n, err)
			continue
		}

		// Check received parameter
		if testApi.ArgsIn[ListAttachedGroupPoliciesMethod][1] != test.org {
			t.Errorf("Test case %v. Received different Org (wanted:%v / received:%v)", n, test.org, testApi.ArgsIn[ListAttachedGroupPoliciesMethod][1])
			continue
		}
		if testApi.ArgsIn[ListAttachedGroupPoliciesMethod][2] != test.name {
			t.Errorf("Test case %v. Received different Name (wanted:%v / received:%v)", n, test.name, testApi.ArgsIn[ListAttachedGroupPoliciesMethod][2])
			continue
		}

		// check status code
		if test.expectedStatusCode != res.StatusCode {
			t.Errorf("Test case %v. Received different http status code (wanted:%v / received:%v)", n, test.expectedStatusCode, res.StatusCode)
			continue
		}

		switch res.StatusCode {
		case http.StatusOK:
			getGroupPoliciesResponse := GetGroupPoliciesResponse{}
			err = json.NewDecoder(res.Body).Decode(&getGroupPoliciesResponse)
			if err != nil {
				t.Errorf("Test case %v. Unexpected error parsing response %v", n, err)
				continue
			}
			// Check result
			if diff := pretty.Compare(getGroupPoliciesResponse, test.expectedResponse); diff != "" {
				t.Errorf("Test %v failed. Received different responses (received/wanted) %v", n, diff)
				continue
			}
		case http.StatusInternalServerError: // Empty message so continue
			continue
		default:
			apiError := api.Error{}
			err = json.NewDecoder(res.Body).Decode(&apiError)
			if err != nil {
				t.Errorf("Test case %v. Unexpected error parsing error response %v", n, err)
				continue
			}
			// Check result
			if diff := pretty.Compare(apiError, test.expectedError); diff != "" {
				t.Errorf("Test %v failed. Received different error response (received/wanted) %v", n, diff)
				continue
			}

		}
	}
}

func TestWorkerHandler_HandleListAllGroups(t *testing.T) {
	testcases := map[string]struct {
		// API method args
		org        string
		pathPrefix string
		// Expected result
		expectedStatusCode int
		expectedResponse   GetGroupsResponse
		expectedError      api.Error
		// Manager Results
		getListAllGroupResult []api.GroupIdentity
		// Manager Errors
		getListAllGroupErr error
	}{
		"OkCase": {
			org:                "org1",
			pathPrefix:         "/path/",
			expectedStatusCode: http.StatusOK,
			expectedResponse: GetGroupsResponse{
				[]api.GroupIdentity{
					api.GroupIdentity{
						Org:  "org1",
						Name: "group1",
					},
				},
			},
			getListAllGroupResult: []api.GroupIdentity{
				api.GroupIdentity{
					Org:  "org1",
					Name: "group1",
				},
			},
		},
		"ErrorCaseUnauthorizedError": {
			org:                "org1",
			pathPrefix:         "/path/",
			expectedStatusCode: http.StatusForbidden,
			expectedError: api.Error{
				Code:    api.UNAUTHORIZED_RESOURCES_ERROR,
				Message: "Unauthorized",
			},
			getListAllGroupErr: &api.Error{
				Code:    api.UNAUTHORIZED_RESOURCES_ERROR,
				Message: "Unauthorized",
			},
		},
		"ErrorCaseInvalidParameterError": {
			org:                "org1",
			pathPrefix:         "Invalid",
			expectedStatusCode: http.StatusBadRequest,
			expectedError: api.Error{
				Code:    api.INVALID_PARAMETER_ERROR,
				Message: "Invalid Path",
			},
			getListAllGroupErr: &api.Error{
				Code:    api.INVALID_PARAMETER_ERROR,
				Message: "Invalid Path",
			},
		},
		"ErrorCaseUnknownApiError": {
			org:                "org1",
			pathPrefix:         "/path/",
			expectedStatusCode: http.StatusInternalServerError,
			getListAllGroupErr: &api.Error{
				Code:    api.UNKNOWN_API_ERROR,
				Message: "Error",
			},
		},
	}

	client := http.DefaultClient

	for n, test := range testcases {

		testApi.ArgsOut[GetListGroupsMethod][0] = test.getListAllGroupResult
		testApi.ArgsOut[GetListGroupsMethod][1] = test.getListAllGroupErr

		url := fmt.Sprintf(server.URL+API_VERSION_1+"/groups?Org=%v&PathPrefix=%v", test.org, test.pathPrefix)
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			t.Errorf("Test case %v. Unexpected error creating http request %v", n, err)
			continue
		}

		res, err := client.Do(req)
		if err != nil {
			t.Errorf("Test case %v. Unexpected error calling server %v", n, err)
			continue
		}

		// Check received parameter
		// Check received parameter
		if testApi.ArgsIn[GetListGroupsMethod][1] != test.org {
			t.Errorf("Test case %v. Received different Org (wanted:%v / received:%v)", n, test.org, testApi.ArgsIn[GetListGroupsMethod][1])
			continue
		}
		if testApi.ArgsIn[GetListGroupsMethod][2] != test.pathPrefix {
			t.Errorf("Test case %v. Received different PathPrefix (wanted:%v / received:%v)", n, test.pathPrefix, testApi.ArgsIn[GetListGroupsMethod][2])
			continue
		}

		// check status code
		if test.expectedStatusCode != res.StatusCode {
			t.Errorf("Test case %v. Received different http status code (wanted:%v / received:%v)", n, test.expectedStatusCode, res.StatusCode)
			continue
		}

		switch res.StatusCode {
		case http.StatusOK:
			getGroupsResponse := GetGroupsResponse{}
			err = json.NewDecoder(res.Body).Decode(&getGroupsResponse)
			if err != nil {
				t.Errorf("Test case %v. Unexpected error parsing response %v", n, err)
				continue
			}
			// Check result
			if diff := pretty.Compare(getGroupsResponse, test.expectedResponse); diff != "" {
				t.Errorf("Test %v failed. Received different responses (received/wanted) %v", n, diff)
				continue
			}
		case http.StatusInternalServerError: // Empty message so continue
			continue
		default:
			apiError := api.Error{}
			err = json.NewDecoder(res.Body).Decode(&apiError)
			if err != nil {
				t.Errorf("Test case %v. Unexpected error parsing error response %v", n, err)
				continue
			}
			// Check result
			if diff := pretty.Compare(apiError, test.expectedError); diff != "" {
				t.Errorf("Test %v failed. Received different error response (received/wanted) %v", n, diff)
				continue
			}

		}
	}
}
