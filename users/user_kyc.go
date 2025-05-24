package users

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/timchuks/monieverse/core/controllers/shared"
	"github.com/timchuks/monieverse/core/perms"
	"github.com/timchuks/monieverse/core/server"
	db "github.com/timchuks/monieverse/internal/db/sqlc"
	"github.com/timchuks/monieverse/internal/domain"
	"github.com/timchuks/monieverse/internal/fields"
	"github.com/timchuks/monieverse/internal/validator"
	"github.com/timchuks/monieverse/internal/worker"
	"golang.org/x/exp/slices"
)

func (c *usersController) GetUserKYC(ctx *gin.Context) {

	user := c.srv.ContextGetUser(ctx)

	kyc := c.getUserKYC(ctx, user)

	ctx.JSON(http.StatusOK, gin.H{"status": "ok", "data": kyc})

}

func (c *usersController) GetUserKYCRequirements(ctx *gin.Context) {
	srv := c.srv
	user := srv.ContextGetUser(ctx)

	statusParam := ctx.DefaultQuery("status", "active")
	statusSlice := strings.Split(statusParam, ",")
	for i, status := range statusSlice {
		statusSlice[i] = strings.TrimSpace(status)
	}

	allowedStatuses := map[string]bool{
		"active": true,
		"closed": true,
	}
	for _, status := range statusSlice {
		if !allowedStatuses[status] {
			srv.ErrorJSONResponse(ctx, http.StatusBadRequest, fmt.Errorf("invalid status parameter: %s", status))
			return
		}
	}

	pageParam := ctx.DefaultQuery("page", "1")
	perPageParam := ctx.DefaultQuery("per_page", "10")

	page, err := strconv.Atoi(pageParam)
	if err != nil || page < 1 {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, errors.New("invalid page parameter"))
		return
	}

	perPage, err := strconv.Atoi(perPageParam)
	if err != nil || perPage < 1 || perPage > 100 {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, errors.New("invalid per_page parameter"))
		return
	}

	limit := perPage
	offset := (page - 1) * perPage

	params := db.GetKYCRequirementsForUserParams{
		UserID:      user.ID.String(),
		CountryCode: user.CountryCode,
		AccountType: user.AccountType,
		Statuses:    statusSlice,
		Limit:       limit,
		Offset:      offset,
	}

	kycs, err := srv.Store.GetKYCRequirementsForUser(ctx, params)
	if err != nil {
		srv.Logger.Error(fmt.Errorf("failed to get kyc requirements: %w", err), nil)
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, errors.New("something went wrong! try again"))
		return
	}

	var out []domain.KYCRequirementOutput

	for _, k := range kycs {
		kyc, err := dbKYCToOutput(k)
		if err != nil {
			srv.Logger.Error(fmt.Errorf("failed to parse kyc payload: %w", err), nil)
			srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, errors.New("something went wrong! try again"))
			return
		}
		out = append(out, *kyc)
	}

	response := struct {
		Status  string                        `json:"status"`
		Data    []domain.KYCRequirementOutput `json:"data"`
		Page    int                           `json:"page"`
		PerPage int                           `json:"per_page"`
	}{
		Status:  "ok",
		Data:    out,
		Page:    page,
		PerPage: perPage,
	}

	srv.SuccessJSONResponse(ctx, http.StatusOK, server.ResponseOk, response)
}

func (c *usersController) GetUserKYCRequirementResults(ctx *gin.Context) {
	srv := c.srv

	kycID, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, errors.New("invalid kyc id"))
		return
	}

	user := srv.ContextGetUser(ctx)

	if srv.Store.HasPermission(ctx, *user, perms.AdminPermission) {
		userID, err := uuid.Parse(ctx.Param("user_id"))
		if err != nil {
			srv.ErrorJSONResponse(ctx, http.StatusBadRequest, errors.New("invalid user id"))
			return
		}

		u, err := srv.Store.GetUser(ctx, userID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				srv.ErrorJSONResponse(ctx, http.StatusNotFound, errors.New("user not found"))
				return
			}
			srv.Logger.Error(err, nil)
			srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, errors.New("something went wrong! try again"))
			return
		}
		user = &u
	}

	kycResults, err := c.getUserKYCRequirementResults(ctx, user, kycID)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			srv.ErrorJSONResponse(ctx, http.StatusNotFound, errors.New("kyc requirement not found"))
			return
		}
		srv.Logger.Error(err, nil)
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, errors.New("something went wrong! try again"))
	}

	srv.SuccessJSONResponse(ctx, http.StatusOK, server.ResponseOk, gin.H{
		"kyc_results": kycResults,
	})
}

func (c *usersController) getUserKYCRequirementResults(ctx context.Context, user *db.User, kycID uuid.UUID) ([]domain.KYCOutput, error) {

	srv := c.srv

	userKYC, err := srv.Store.GetUserSubmittedKYCRequirement(ctx, db.GetUserSubmittedKYCRequirementParams{
		UserID:           user.ID,
		KycRequirementID: kycID,
	})

	if err != nil {
		return nil, err
	}

	plainTextKYCResults, err := srv.Store.GetUserDynamicKYCResults(ctx, userKYC.ID)
	if err != nil {
		return nil, err
	}

	documents, err := srv.Store.GetDocumentsByModel(ctx, db.GetDocumentsByModelParams{
		Model:   db.DocumentModelDynamicKYCResult,
		ModelID: userKYC.ID,
	})

	if err != nil {
		return nil, err
	}

	var kycResults []domain.KYCOutput
	for _, r := range plainTextKYCResults {
		kycResults = append(kycResults, domain.KYCOutput{
			ID:    r.ID,
			Field: r.Field,
			Value: r.Value,
		})
	}

	for _, doc := range documents {
		docURL, err := srv.Uploader.GetTempURL(doc.Bucket, doc.DocumentPath)
		if err != nil {
			srv.Logger.Error(fmt.Errorf("failed to get temp url from %s: %w", srv.Config.FileUploadProvider, err), nil)
			continue
		}
		kycResults = append(kycResults, domain.KYCOutput{
			ID:     doc.ID,
			Field:  doc.DocumentNumber,
			Value:  docURL,
			IsFile: true,
		})
	}

	return kycResults, nil

}

func dbKYCToOutput(k db.KycRequirement) (*domain.KYCRequirementOutput, error) {
	var out domain.KYCRequirementOutput
	var err error
	out.ID = k.ID
	out.Target = k.Target
	out.TargetID = k.TargetID
	out.Deadline = k.Deadline
	out.Title = k.Title
	out.Description = k.Description
	var payload domain.KYCRequirement
	err = json.Unmarshal(k.Payload, &payload)
	if err != nil {
		return nil, err
	}
	out.Fields = getKYCFieldsFromDBFields(payload.Fields)
	return &out, nil

}

func getKYCFieldsFromDBFields(dbFields []string) []fields.Field {

	var kycFields []fields.Field
	for k, f := range fields.KYCFieldsMap {
		if slices.Contains(dbFields, k) {
			kycFields = append(kycFields, f)
		}
	}
	return kycFields
}

func (c *usersController) ProcessUserKYCRequirements(ctx *gin.Context) {
	srv := c.srv

	if err := ctx.Request.ParseMultipartForm(srv.Config.MaxBodyPayloadSize); err != nil {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, err)
		return
	}

	user := srv.ContextGetUser(ctx)

	requirementID := ctx.Request.FormValue("requirement_id")
	kyc, err := c.findUserKYCRequirement(ctx, requirementID, user)

	if err != nil {
		srv.Logger.Error(fmt.Errorf("failed to get kyc requirement: %w", err), nil)
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, errors.New("invalid requirement id"))
		return
	}

	ukyc, err := srv.Store.GetUserKYCRequirement(ctx, db.GetUserKYCRequirementParams{
		UserID:           user.ID,
		KycRequirementID: kyc.ID,
	})

	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		srv.Logger.Error(err, nil)
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, errors.New("unable to get user kyc requirement"))
		return
	}

	if ukyc.ID != uuid.Nil && ukyc.Status != domain.KYCRequirementStatusPending {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest, errors.New("you cannot update a completed kyc requirement: current status is "+ukyc.Status))
		return
	}

	domainKYC, err := domain.ConvertDBKYCRequirementKYCRequirement(kyc, false)
	if err != nil {
		srv.Logger.Error(fmt.Errorf("failed to parse kyc payload: %w", err), nil)
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, errors.New("something went wrong! try again"))
		return
	}

	submittedFields := extractDynamicFields(ctx.Request.MultipartForm, "requirement_id")

	if !slices.Equal(domainKYC.Fields, submittedFields) {
		srv.ErrorJSONResponse(ctx, http.StatusBadRequest,
			errors.New("invalid fields submitted: valid fields are"+fmt.Sprintf("%s", domainKYC.Fields)))
		return
	}

	textFieldMaps := make(map[string]string)

	for ky, values := range ctx.Request.MultipartForm.Value {
		value := values[0]
		if value == "" {
			srv.ErrorJSONResponse(ctx, http.StatusBadRequest, fmt.Errorf("invalid value for field %s", ky))
			return
		}
		if slices.Contains(domainKYC.Fields, ky) {
			textFieldMaps[ky] = value
		}

	}

	if ukyc.ID == uuid.Nil {
		ukyc, err = srv.Store.CreateUserKYCRequirement(ctx, db.CreateUserKYCRequirementParams{
			KycRequirementID: kyc.ID,
			UserID:           user.ID,
			Payload:          []byte("{}"),
		})

		if err != nil {
			srv.Logger.Error(err, nil)
			srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, errors.New("something went wrong! try again during creation of user kyc requirement"))
		}
	}

	_ = srv.Store.DeleteUserKYCRequirementResults(ctx, ukyc.ID)

	err = srv.Store.CreateKYCRequirementResultEntry(ctx, ukyc.ID, textFieldMaps)
	if err != nil {
		srv.Logger.Error(err, map[string]interface{}{
			"ukyc": textFieldMaps,
		})
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, err)
		return
	}

	for key, files := range ctx.Request.MultipartForm.File {
		f := files[0]
		_, err = c.processFileKYCUploadOperation(ctx, user, ukyc, f, key)
		if err != nil {
			srv.Logger.Error(err, nil)
			srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, errors.New("something went wrong! try again during upload of user kyc requirement"))
			return
		}
	}

	err = srv.Store.UpdateUserKYCRequirementStatus(ctx, db.UpdateUserKYCRequirementStatusParams{
		ID:     ukyc.ID,
		Status: domain.KYCRequirementStatusSubmitted,
	})
	if err != nil {
		srv.Logger.Error(err, nil)
		srv.ErrorJSONResponse(ctx, http.StatusInternalServerError, errors.New("something went wrong! try again please"))
		return
	}

	_ = srv.TaskDistributor.Fire(ctx, worker.TaskKYCSubmittedByUser, user.ID.String(), nil)

	srv.SuccessJSONResponse(ctx, http.StatusOK, server.ResponseOk, nil)

}

func (c *usersController) processFileKYCUploadOperation(
	ctx context.Context,
	user *db.User,
	ukyc db.KycRequirementsUser,
	f *multipart.FileHeader,
	field string) (*db.Document, error) {

	srv := c.srv

	file, err := f.Open()
	if err != nil {
		return nil, err
	}

	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			srv.Logger.Error(fmt.Errorf("failed to close file: %w", err), nil)
		}
	}(file)

	oldDoc, err := srv.Store.GetDocumentByModel(ctx, db.GetDocumentByModelParams{
		Model:   db.DocumentModelDynamicKYCResult,
		ModelID: ukyc.ID,
	})
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	req := domain.UploadDocumentRequest{
		DocumentType:   "kyc",
		DocumentNumber: field,
		ContentType:    f.Header.Get("Content-Type"),
		Size:           f.Size,
		FileName:       f.Filename,
		Bucket:         srv.Config.FileBucket,
		Model:          db.DocumentModelDynamicKYCResult,
		ModelID:        ukyc.ID,
	}

	v := validator.New()
	if !req.Validate(v, srv.Config) {
		return nil, validator.NewValidationError("validation failed", v.Errors)
	}

	dbRes, err := srv.Store.CreateDocumentTx(ctx, shared.UploadDocumentToSource(srv, user, file, req, "kyc"))

	if err != nil {
		return nil, err
	}

	if err = shared.CleanUpOldDocument(ctx, srv, oldDoc); err != nil {
		srv.Logger.Error(err, map[string]interface{}{
			"doc": oldDoc,
		})
	}

	return dbRes, nil
}

func (c *usersController) findUserKYCRequirement(ctx context.Context, requirementID string, user *db.User) (*db.KycRequirement, error) {
	if user == nil {
		return nil, errors.New("user not found")
	}

	id, err := uuid.Parse(requirementID)
	if err != nil {
		return nil, err
	}

	kyc, err := c.srv.Store.GetActiveKYCRequirementByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if kyc.Target == domain.TargetUser && kyc.TargetID != user.ID.String() {
		return nil, errors.New("user does not match requirement")
	}

	return &kyc, err

}

func extractDynamicFields(form *multipart.Form, skip string) []string {
	var out []string
	if form == nil {
		return out
	}
	for k := range form.Value {
		if k != skip {
			out = append(out, k)
		}
	}

	for k := range form.File {
		out = append(out, k)
	}

	slices.Sort(out)
	return out
}

func (c *usersController) getUserKYC(ctx *gin.Context, user *db.User) *domain.CustomerKYC {
	return shared.GetUserKYC(ctx, c.srv, user)
}
