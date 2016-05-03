package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"strings"

	"github.com/julienschmidt/httprouter"
	"github.com/satori/go.uuid"
	"github.com/tecsisa/authorizr/api"
	"github.com/tecsisa/authorizr/authorizr"
)

// Requests

type CreateGroupRequest struct {
	Name string `json:"Name, omitempty"`
	Path string `json:"Path, omitempty"`
}

// Responses

type CreateGroupResponse struct {
	Group *api.Group
}

type GetGroupNameResponse struct {
	Group *api.Group
}

type GroupHandler struct {
	core *authorizr.Core
}

func (g *GroupHandler) handleCreateGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Decode request
	request := CreateGroupRequest{}
	err := json.NewDecoder(r.Body).Decode(&request)
	if err != nil {
		g.core.Logger.Errorln(err)
		RespondBadRequest(w)
		return
	}

	// Check parameters
	if len(strings.TrimSpace(request.Name)) == 0 ||
		len(strings.TrimSpace(request.Path)) == 0 {
		g.core.Logger.Errorf("There are mising parameters: Name %v, Path %v", request.Name, request.Path)
		RespondBadRequest(w)
		return
	}

	org := ps.ByName(ORG_ID)
	// Call group API to create an group
	result, err := g.core.GroupApi.AddGroup(createGroupFromRequest(request, org))

	// Error handling
	if err != nil {
		g.core.Logger.Errorln(err)
		RespondInternalServerError(w)
		return
	}

	response := &CreateGroupResponse{
		Group: result,
	}

	// Write group to response
	RespondOk(w, response)
}

func (g *GroupHandler) handleDeleteGroup(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}

func (g *GroupHandler) handleGetGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Retrieve group org and name from path
	org := ps.ByName(ORG_ID)
	name := ps.ByName(GROUP_ID)

	// Call group API to retrieve group
	result, err := g.core.GroupApi.GetGroupByName(org, name)

	// Error handling
	if err != nil {
		g.core.Logger.Errorln(err)
		// Transform to API errors
		apiError := err.(*api.Error)
		if apiError.Code == api.GROUP_BY_ORG_AND_NAME_NOT_FOUND {
			RespondNotFound(w)
			return
		} else { // Unexpected API error
			RespondInternalServerError(w)
			return
		}
	}

	response := GetGroupNameResponse{
		Group: result,
	}

	// Write group to response
	RespondOk(w, response)
}

func (g *GroupHandler) handleListGroups(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}

func (g *GroupHandler) handleUpdateGroup(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}

func (g *GroupHandler) handleListMembers(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}

func (g *GroupHandler) handleAddMember(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}

func (g *GroupHandler) handleRemoveMember(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}

func (g *GroupHandler) handleAttachGroupPolicy(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}

func (g *GroupHandler) handleDetachGroupPolicy(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}

func (g *GroupHandler) handleListAtachhedGroupPolicies(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}

func (g *GroupHandler) handleListAllGroups(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {

}

func createGroupFromRequest(request CreateGroupRequest, org string) api.Group {
	path := request.Path + "/" + request.Name
	urn := fmt.Sprintf("urn:iws:iam:%v:group/%v", org, path)
	group := api.Group{
		ID:   uuid.NewV4().String(),
		Name: request.Name,
		Path: path,
		Urn:  urn,
		Org:  org,
	}

	return group
}
