package handlers

import (
	"TeamTickBackend/app"
	"TeamTickBackend/gen"
	appErrors "TeamTickBackend/pkg/errors"
	service "TeamTickBackend/services"
	"context"
	"errors"
)

type AuditRequestHandler struct {
	auditRequestService service.AuditRequestService
}

func NewAuditRequestHandler(container *app.AppContainer) gen.AuditRequestServerInterface {
	auditRequestService := service.NewAuditRequestService(
		container.DaoFactory.TransactionManager,
		container.DaoFactory.CheckApplicationDAO,
		container.DaoFactory.TaskRecordDAO,
		container.DaoFactory.TaskDAO,
		container.DaoFactory.GroupDAO,
	)
	handler := &AuditRequestHandler{
		auditRequestService: *auditRequestService,
	}
	return gen.NewAuditRequestStrictHandler(handler, nil)
}
