package handlers

import (
	"TeamTickBackend/app"
	"TeamTickBackend/dal/models"
	"TeamTickBackend/gen"
	appErrors "TeamTickBackend/pkg/errors"
	service "TeamTickBackend/services"
	"context"
	"errors"
	"time"
)

type TaskHandler struct {
	taskService         *service.TaskService
	groupsService       *service.GroupsService
	auditRequestService *service.AuditRequestService
}

func NewTaskHandler(container *app.AppContainer) (gen.CheckinTasksServerInterface, gen.CheckinRecordsServerInterface) {
	TaskService := service.NewTaskService(
		container.DaoFactory.TaskDAO,
		container.DaoFactory.TaskRecordDAO,
		container.DaoFactory.TaskRecordRedisDAO,
		container.DaoFactory.TaskRedisDAO,
		container.DaoFactory.TransactionManager,
		container.DaoFactory.GroupDAO,
		container.DaoFactory.GroupMemberDAO,
	)
	GroupsService := service.NewGroupsService(
		container.DaoFactory.GroupDAO,
		container.DaoFactory.GroupMemberDAO,
		container.DaoFactory.JoinApplicationDAO,
		container.DaoFactory.TransactionManager,
		container.DaoFactory.GroupRedisDAO,
		container.DaoFactory.GroupMemberRedisDAO,
		container.DaoFactory.JoinApplicationRedisDAO,
		container.DaoFactory.TaskRedisDAO,
	)
	AuditRequestService := service.NewAuditRequestService(
		container.DaoFactory.TransactionManager,
		container.DaoFactory.CheckApplicationDAO,
		container.DaoFactory.TaskRecordDAO,
		container.DaoFactory.TaskDAO,
		container.DaoFactory.GroupDAO,
		container.DaoFactory.CheckApplicationRedisDAO,
	)
	handler := &TaskHandler{
		taskService:         TaskService,
		groupsService:       GroupsService,
		auditRequestService: AuditRequestService,
	}
	return gen.NewCheckinTasksStrictHandler(handler, nil), gen.NewCheckinRecordsStrictHandler(handler, nil)
}

// convertToCheckinTask 将 models.Task 转换为 gen.CheckinTask
func convertToCheckinTask(task *models.Task) gen.CheckinTask {
	now := time.Now()
	var status gen.CheckinTaskStatus
	if now.Before(task.StartTime) {
		status = gen.Upcoming
	} else if now.After(task.EndTime) {
		status = gen.Expired
	} else {
		status = gen.Ongoing
	}

	return gen.CheckinTask{
		CreatedAt:   int(task.CreatedAt.Unix()),
		Description: task.Description,
		EndTime:     int(task.EndTime.Unix()),
		GroupId:     task.GroupID,
		StartTime:   int(task.StartTime.Unix()),
		Status:      status,
		TaskId:      task.TaskID,
		TaskName:    task.TaskName,
		VerificationConfig: gen.TaskVerificationConfig{
			CheckinMethods: gen.CheckinMethods{
				Gps:  task.GPS,
				Face: task.Face,
				Wifi: task.WiFi,
				Nfc:  task.NFC,
			},
			LocationInfo: struct {
				Location gen.Location `json:"location"`
				Radius   int          `json:"radius"`
			}{
				Location: gen.Location{
					Latitude:  task.Latitude,
					Longitude: task.Longitude,
				},
				Radius: task.Radius,
			},
			WifiInfo: func() *gen.WifiInfo {
				if task.WiFi {
					return &gen.WifiInfo{
						Ssid:  task.SSID,
						Bssid: task.BSSID,
					}
				}
				return nil
			}(),
			NfcInfo: func() *gen.NFCInfo {
				if task.NFC {
					return &gen.NFCInfo{
						TagId:   task.TagID,
						TagName: task.TagName,
					}
				}
				return nil
			}(),
		},
	}
}

// 管理员删除指定签到任务
func (h *TaskHandler) DeleteCheckinTasksTaskId(ctx context.Context, request gen.DeleteCheckinTasksTaskIdRequestObject) (gen.DeleteCheckinTasksTaskIdResponseObject, error) {
	// 从上下文中获取当前用户ID
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}

	// 获取任务信息以获取组ID
	task, err := h.taskService.GetTaskByTaskID(ctx, request.TaskId)
	if err != nil {
		if errors.Is(err, appErrors.ErrTaskNotFound) {
			return gen.DeleteCheckinTasksTaskId404JSONResponse{
				Code:    "1",
				Message: "任务不存在",
			}, nil
		}
		return nil, err
	}

	// 验证用户是否是组的管理员
	if err := h.groupsService.CheckMemberPermission(ctx, task.GroupID, userID); err != nil {
		if errors.Is(err, appErrors.ErrRolePermissionDenied) {
			return gen.DeleteCheckinTasksTaskId403JSONResponse{
				Code:    "1",
				Message: "没有权限删除该任务",
			}, nil
		}
		if errors.Is(err, appErrors.ErrGroupMemberNotFound) {
			return gen.DeleteCheckinTasksTaskId404JSONResponse{
				Code:    "1",
				Message: "用户不存在",
			}, nil
		}
		return nil, err
	}

	// 调用服务层删除任务
	err = h.taskService.DeleteTask(ctx, request.TaskId)
	if err != nil {
		return nil, err
	}

	return gen.DeleteCheckinTasksTaskId200JSONResponse{
		Code: "0",
		Data: &map[string]interface{}{},
	}, nil
}

// 获取指定签到任务的详细信息。需要是任务所属组成员
func (h *TaskHandler) GetCheckinTasksTaskId(ctx context.Context, request gen.GetCheckinTasksTaskIdRequestObject) (gen.GetCheckinTasksTaskIdResponseObject, error) {
	// 从上下文中获取当前用户ID
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}

	// 获取任务信息以获取组ID
	task, err := h.taskService.GetTaskByTaskID(ctx, request.TaskId)
	if err != nil {
		if errors.Is(err, appErrors.ErrTaskNotFound) {
			return gen.GetCheckinTasksTaskId404JSONResponse{
				Code:    "1",
				Message: "任务不存在",
			}, nil
		}
		return nil, err
	}

	// 验证用户是否是组的成员
	if err := h.groupsService.CheckUserExistInGroup(ctx, task.GroupID, userID); err != nil {
		if !errors.Is(err, appErrors.ErrRolePermissionDenied) {
			return gen.GetCheckinTasksTaskId403JSONResponse{
				Code:    "1",
				Message: "没有权限查看该任务",
			}, nil
		}
		if errors.Is(err, appErrors.ErrGroupMemberNotFound) {
			return gen.GetCheckinTasksTaskId404JSONResponse{
				Code:    "1",
				Message: "用户不存在",
			}, nil
		}

		return nil, err
	}

	// 转换任务为API响应格式
	checkinTask := convertToCheckinTask(task)
	return gen.GetCheckinTasksTaskId200JSONResponse{
		Code: "0",
		Data: checkinTask,
	}, nil
}

// 更新指定签到任务的信息。需要是任务所属组的管理员。注意：如果当前时间已经到达或超过了签到开始时间，将无法修改任务
func (h *TaskHandler) PutCheckinTasksTaskId(ctx context.Context, request gen.PutCheckinTasksTaskIdRequestObject) (gen.PutCheckinTasksTaskIdResponseObject, error) {
	// 从上下文中获取当前用户ID
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}

	// 获取任务信息以获取组ID
	task, err := h.taskService.GetTaskByTaskID(ctx, request.TaskId)
	if err != nil {
		if errors.Is(err, appErrors.ErrTaskNotFound) {
			return gen.PutCheckinTasksTaskId404JSONResponse{
				Code:    "1",
				Message: "任务不存在",
			}, nil
		}
		return nil, err
	}

	// 验证用户是否是组的管理员
	if err := h.groupsService.CheckMemberPermission(ctx, task.GroupID, userID); err != nil {
		if errors.Is(err, appErrors.ErrGroupMemberNotFound) {
			return gen.PutCheckinTasksTaskId404JSONResponse{
				Code:    "1",
				Message: "用户不存在",
			}, nil
		}
		if errors.Is(err, appErrors.ErrRolePermissionDenied) {
			return gen.PutCheckinTasksTaskId403JSONResponse{
				Code:    "1",
				Message: "没有权限修改该任务",
			}, nil
		}
		return nil, err
	}

	// 调用服务层更新任务
	task, err = h.taskService.UpdateTask(
		ctx,
		request.TaskId,
		request.Body.TaskName,
		request.Body.Description,
		time.Unix(int64(request.Body.StartTime), 0),
		time.Unix(int64(request.Body.EndTime), 0),
		request.Body.VerificationConfig.LocationInfo.Location.Latitude,
		request.Body.VerificationConfig.LocationInfo.Location.Longitude,
		request.Body.VerificationConfig.LocationInfo.Radius,
		request.Body.VerificationConfig.CheckinMethods.Gps,
		request.Body.VerificationConfig.CheckinMethods.Face,
		request.Body.VerificationConfig.CheckinMethods.Wifi,
		request.Body.VerificationConfig.CheckinMethods.Nfc,
	)
	if err != nil {
		if errors.Is(err, appErrors.ErrAuditRequestNotFound) {
			return gen.PutCheckinTasksTaskId404JSONResponse{
				Code:    "1",
				Message: "签到任务不存在",
			}, nil
		}
		return nil, err
	}

	// 转换任务为API响应格式
	checkinTask := convertToCheckinTask(task)
	return gen.PutCheckinTasksTaskId200JSONResponse{
		Code: "0",
		Data: checkinTask,
	}, nil
}

// PostCheckinTasksTaskIdVerify 验证用户提供的签到信息是否符合签到任务要求
func (h *TaskHandler) PostCheckinTasksTaskIdVerify(ctx context.Context, request gen.PostCheckinTasksTaskIdVerifyRequestObject) (gen.PostCheckinTasksTaskIdVerifyResponseObject, error) {
	// 从上下文中获取当前用户ID
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}

	// 获取任务信息以获取组ID
	task, err := h.taskService.GetTaskByTaskID(ctx, request.TaskId)
	if err != nil {
		if errors.Is(err, appErrors.ErrTaskNotFound) {
			return &gen.PostCheckinTasksTaskIdVerify404JSONResponse{
				Code:    "1",
				Message: "任务不存在",
			}, nil
		}
		return nil, err
	}

	// 验证用户是否是组的成员
	if err := h.groupsService.CheckUserExistInGroup(ctx, task.GroupID, userID); err != nil {
		if errors.Is(err, appErrors.ErrGroupMemberNotFound) {
			return &gen.PostCheckinTasksTaskIdVerify403JSONResponse{
				Code:    "1",
				Message: "您不是该组成员",
			}, nil
		}
		if !errors.Is(err, appErrors.ErrRolePermissionDenied) {
			return nil, err
		}
	}

	// 验证请求的验证类型是否在任务配置中启用
	switch request.Body.VerifyType {
	case gen.Gps:
		if !task.GPS {
			return &gen.PostCheckinTasksTaskIdVerify400JSONResponse{
				Code:    "1",
				Message: "该任务未启用GPS验证",
			}, nil
		}
	case gen.Wifi:
		if !task.WiFi {
			return &gen.PostCheckinTasksTaskIdVerify400JSONResponse{
				Code:    "1",
				Message: "该任务未启用WiFi验证",
			}, nil
		}
	case gen.Nfc:
		if !task.NFC {
			return &gen.PostCheckinTasksTaskIdVerify400JSONResponse{
				Code:    "1",
				Message: "该任务未启用NFC验证",
			}, nil
		}
	default:
		return &gen.PostCheckinTasksTaskIdVerify400JSONResponse{
			Code:    "1",
			Message: "无效的验证类型",
		}, nil
	}

	// 根据验证类型执行具体的验证逻辑
	var isValid bool
	var verifyType gen.PostCheckinTasksTaskIdVerifyJSONBodyVerifyType
	var message string

	switch request.Body.VerifyType {
	case gen.Gps:
		if request.Body.VerificationData.LocationInfo == nil {
			return &gen.PostCheckinTasksTaskIdVerify400JSONResponse{
				Code:    "1",
				Message: "缺少位置信息",
			}, nil
		}
		isValid = h.taskService.VerifyLocation(
			ctx,
			request.Body.VerificationData.LocationInfo.Location.Latitude,
			request.Body.VerificationData.LocationInfo.Location.Longitude,
			request.TaskId,
		)
		if !isValid {
			return &gen.PostCheckinTasksTaskIdVerify200JSONResponse{
				Code: "0",
				Data: struct {
					Message    string                                             `json:"message"`
					Valid      bool                                               `json:"valid"`
					VerifyType gen.PostCheckinTasksTaskIdVerifyJSONBodyVerifyType `json:"verifyType"`
				}{
					Message:    "位置验证失败",
					Valid:      false,
					VerifyType: gen.Gps,
				},
			}, nil
		}
		verifyType = gen.Gps
		message = "位置验证"
	case gen.Wifi:
		if request.Body.VerificationData.WifiInfo == nil {
			return &gen.PostCheckinTasksTaskIdVerify400JSONResponse{
				Code:    "1",
				Message: "缺少WiFi信息",
			}, nil
		}
		isValid = h.taskService.VerifyWiFi(
			ctx,
			request.Body.VerificationData.WifiInfo.Ssid,
			request.Body.VerificationData.WifiInfo.Bssid,
			request.TaskId,
		)
		if !isValid {
			return &gen.PostCheckinTasksTaskIdVerify200JSONResponse{
				Code: "0",
				Data: struct {
					Message    string                                             `json:"message"`
					Valid      bool                                               `json:"valid"`
					VerifyType gen.PostCheckinTasksTaskIdVerifyJSONBodyVerifyType `json:"verifyType"`
				}{
					Message:    "WiFi验证失败",
					Valid:      false,
					VerifyType: gen.Wifi,
				},
			}, nil
		}
		verifyType = gen.Wifi
		message = "WiFi验证"
	case gen.Nfc:
		if request.Body.VerificationData.NfcInfo == nil {
			return &gen.PostCheckinTasksTaskIdVerify400JSONResponse{
				Code:    "1",
				Message: "缺少NFC信息",
			}, nil
		}
		isValid = h.taskService.VerifyNFC(
			ctx,
			request.Body.VerificationData.NfcInfo.TagId,
			request.Body.VerificationData.NfcInfo.TagName,
			request.TaskId,
		)
		if !isValid {
			return &gen.PostCheckinTasksTaskIdVerify200JSONResponse{
				Code: "0",
				Data: struct {
					Message    string                                             `json:"message"`
					Valid      bool                                               `json:"valid"`
					VerifyType gen.PostCheckinTasksTaskIdVerifyJSONBodyVerifyType `json:"verifyType"`
				}{
					Message:    "NFC验证失败",
					Valid:      false,
					VerifyType: gen.Nfc,
				},
			}, nil
		}
		verifyType = gen.Nfc
		message = "NFC验证"
	}

	message += "成功"

	return &gen.PostCheckinTasksTaskIdVerify200JSONResponse{
		Code: "0",
		Data: struct {
			Message    string                                             `json:"message"`
			Valid      bool                                               `json:"valid"`
			VerifyType gen.PostCheckinTasksTaskIdVerifyJSONBodyVerifyType `json:"verifyType"`
		}{
			Message:    message,
			Valid:      true,
			VerifyType: verifyType,
		},
	}, nil
}

// 管理员查看用户组内的所有签到任务
func (h *TaskHandler) GetGroupsGroupIdCheckinTasks(ctx context.Context, request gen.GetGroupsGroupIdCheckinTasksRequestObject) (gen.GetGroupsGroupIdCheckinTasksResponseObject, error) {
	// 从上下文中获取当前用户ID
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}

	// 验证用户是否是组的管理员
	if err := h.groupsService.CheckMemberPermission(ctx, request.GroupId, userID); err != nil {
		if errors.Is(err, appErrors.ErrGroupMemberNotFound) {
			return gen.GetGroupsGroupIdCheckinTasks404JSONResponse{
				Code:    "1",
				Message: "用户不存在",
			}, nil
		}
		if errors.Is(err, appErrors.ErrRolePermissionDenied) {
			return gen.GetGroupsGroupIdCheckinTasks403JSONResponse{
				Code:    "1",
				Message: "没有权限查看该组",
			}, nil
		}
		return nil, err
	}

	// 调用服务层获取任务列表
	tasks, err := h.taskService.GetTasksByGroupID(ctx, request.GroupId)
	if err != nil {
		if errors.Is(err, appErrors.ErrTaskNotFound) {
			return gen.GetGroupsGroupIdCheckinTasks200JSONResponse{
				Code: "0",
				Data: []gen.CheckinTask{},
			}, nil
		}
		return nil, err
	}

	// 转换任务列表为API响应格式
	checkinTasks := make([]gen.CheckinTask, len(tasks))
	for i, task := range tasks {
		// 判断任务状态
		now := time.Now()
		var status gen.CheckinTaskStatus
		if now.Before(task.StartTime) {
			status = gen.Upcoming
		} else if now.After(task.EndTime) {
			status = gen.Expired
		} else {
			status = gen.Ongoing
		}

		checkinTasks[i] = gen.CheckinTask{
			CreatedAt:   int(task.CreatedAt.Unix()),
			Description: task.Description,
			EndTime:     int(task.EndTime.Unix()),
			GroupId:     task.GroupID,
			StartTime:   int(task.StartTime.Unix()),
			Status:      status,
			TaskId:      task.TaskID,
			TaskName:    task.TaskName,
			VerificationConfig: gen.TaskVerificationConfig{
				CheckinMethods: gen.CheckinMethods{
					Gps:  task.GPS,
					Face: task.Face,
					Wifi: task.WiFi,
					Nfc:  task.NFC,
				},
				LocationInfo: struct {
					Location gen.Location `json:"location"`
					Radius   int          `json:"radius"`
				}{
					Location: gen.Location{
						Latitude:  task.Latitude,
						Longitude: task.Longitude,
					},
					Radius: task.Radius,
				},
				WifiInfo: func() *gen.WifiInfo {
					if task.WiFi {
						return &gen.WifiInfo{
							Ssid:  task.SSID,
							Bssid: task.BSSID,
						}
					}
					return nil
				}(),
				NfcInfo: func() *gen.NFCInfo {
					if task.NFC {
						return &gen.NFCInfo{
							TagId:   task.TagID,
							TagName: task.TagName,
						}
					}
					return nil
				}(),
			},
		}
	}

	return gen.GetGroupsGroupIdCheckinTasks200JSONResponse{
		Code: "0",
		Data: checkinTasks,
	}, nil
}

// 用户组管理员在指定用户组内创建签到任务
func (h *TaskHandler) PostGroupsGroupIdCheckinTasks(ctx context.Context, request gen.PostGroupsGroupIdCheckinTasksRequestObject) (gen.PostGroupsGroupIdCheckinTasksResponseObject, error) {
	// 从上下文中获取当前用户ID
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}
	// 验证用户组是否存在
	_, err := h.groupsService.GetGroupByGroupID(ctx, request.GroupId)
	if err != nil {
		if errors.Is(err, appErrors.ErrGroupNotFound) {
			return &gen.PostGroupsGroupIdCheckinTasks404JSONResponse{
				Code:    "1",
				Message: "用户组不存在",
			}, nil
		}
		return nil, err
	}

	// 验证用户是否是组的管理员
	if err := h.groupsService.CheckMemberPermission(ctx, request.GroupId, userID); err != nil {
		if errors.Is(err, appErrors.ErrRolePermissionDenied) {
			return &gen.PostGroupsGroupIdCheckinTasks403JSONResponse{
				Code:    "1",
				Message: "没有权限创建任务",
			}, nil
		}
		if errors.Is(err, appErrors.ErrGroupMemberNotFound) {
			return &gen.PostGroupsGroupIdCheckinTasks404JSONResponse{
				Code:    "1",
				Message: "用户不存在",
			}, nil
		}
		return nil, err
	}

	// 验证时间参数
	now := int(time.Now().Unix())
	if request.Body.StartTime <= now {
		return &gen.PostGroupsGroupIdCheckinTasks400JSONResponse{
			Code:    "1",
			Message: "开始时间必须大于当前时间",
		}, nil
	}
	if request.Body.EndTime <= request.Body.StartTime {
		return &gen.PostGroupsGroupIdCheckinTasks400JSONResponse{
			Code:    "1",
			Message: "结束时间必须大于开始时间",
		}, nil
	}

	// 验证 WiFi 和 NFC 相关参数
	if request.Body.VerificationConfig.CheckinMethods.Wifi {
		if request.Body.VerificationConfig.WifiInfo == nil {
			return &gen.PostGroupsGroupIdCheckinTasks400JSONResponse{
				Code:    "1",
				Message: "启用WiFi验证时必须提供WiFi信息",
			}, nil
		}
		if request.Body.VerificationConfig.WifiInfo.Ssid == "" || request.Body.VerificationConfig.WifiInfo.Bssid == "" {
			return &gen.PostGroupsGroupIdCheckinTasks400JSONResponse{
				Code:    "1",
				Message: "WiFi信息不完整，必须提供SSID和BSSID",
			}, nil
		}
	}

	if request.Body.VerificationConfig.CheckinMethods.Nfc {
		if request.Body.VerificationConfig.NfcInfo == nil {
			return &gen.PostGroupsGroupIdCheckinTasks400JSONResponse{
				Code:    "1",
				Message: "启用NFC验证时必须提供NFC信息",
			}, nil
		}
		if request.Body.VerificationConfig.NfcInfo.TagId == "" {
			return &gen.PostGroupsGroupIdCheckinTasks400JSONResponse{
				Code:    "1",
				Message: "NFC信息不完整，必须提供标签ID",
			}, nil
		}
	}

	if request.Body.VerificationConfig.CheckinMethods.Gps {
		if request.Body.VerificationConfig.LocationInfo.Location.Latitude == 0 && request.Body.VerificationConfig.LocationInfo.Location.Longitude == 0 {
			return &gen.PostGroupsGroupIdCheckinTasks400JSONResponse{
				Code:    "1",
				Message: "启用GPS验证时必须提供有效的位置信息",
			}, nil
		}

		if request.Body.VerificationConfig.LocationInfo.Radius == 0 {
			return &gen.PostGroupsGroupIdCheckinTasks400JSONResponse{
				Code:    "1",
				Message: "位置信息不完整，必须提供有效半径",
			}, nil
		}

		if request.Body.VerificationConfig.LocationInfo.Radius < 1 || request.Body.VerificationConfig.LocationInfo.Radius > 10000 {
			return &gen.PostGroupsGroupIdCheckinTasks400JSONResponse{
				Code:    "1",
				Message: "有效半径必须在1-10000米之间",
			}, nil
		}
	}

	// 调用服务层创建任务
	task, err := h.taskService.CreateTask(
		ctx,
		request.Body.TaskName,
		request.Body.Description,
		request.GroupId,
		time.Unix(int64(request.Body.StartTime), 0),
		time.Unix(int64(request.Body.EndTime), 0),
		request.Body.VerificationConfig.LocationInfo.Location.Latitude,
		request.Body.VerificationConfig.LocationInfo.Location.Longitude,
		request.Body.VerificationConfig.LocationInfo.Radius,
		request.Body.VerificationConfig.CheckinMethods.Gps,
		request.Body.VerificationConfig.CheckinMethods.Face,
		request.Body.VerificationConfig.CheckinMethods.Wifi,
		request.Body.VerificationConfig.CheckinMethods.Nfc,
	)
	if err != nil {
		return nil, err
	}

	// 转换任务为API响应格式
	checkinTask := convertToCheckinTask(task)
	return &gen.PostGroupsGroupIdCheckinTasks201JSONResponse{
		Code: "0",
		Data: checkinTask,
	}, nil
}

// 获取当前登录用户需要参与的所有签到任务列表
func (h *TaskHandler) GetUsersMeCheckinTasks(ctx context.Context, request gen.GetUsersMeCheckinTasksRequestObject) (gen.GetUsersMeCheckinTasksResponseObject, error) {
	// 从上下文中获取当前用户ID
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}

	// 调用服务层获取任务列表
	tasks, err := h.taskService.GetTasksByUserID(ctx, userID)
	if err != nil {
		if errors.Is(err, appErrors.ErrTaskNotFound) {
			return &gen.GetUsersMeCheckinTasks200JSONResponse{
				Code: "0",
				Data: []struct {
					CreatedAt          int                   `json:"createdAt,omitempty"`
					Description        string                `json:"description,omitempty"`
					EndTime            int                   `json:"endTime"`
					GroupId            int                   `json:"groupId,omitempty"`
					GroupName          *string               `json:"groupName,omitempty"`
					MyCheckinStatus    gen.UserCheckinStatus `json:"myCheckinStatus,omitempty"`
					StartTime          int                   `json:"startTime"`
					Status             gen.CheckinTaskStatus `json:"status"`
					TaskId             int                   `json:"taskId,omitempty"`
					TaskName           string                `json:"taskName"`
					VerificationConfig struct {
						CheckinMethods gen.CheckinMethods `json:"checkinMethods"`
					} `json:"verificationConfig"`
				}{},
			}, nil
		}
		return nil, err
	}

	// 获取用户的签到记录
	records, err := h.taskService.GetTaskRecordsByUserID(ctx, userID)
	if err != nil {
		if !errors.Is(err, appErrors.ErrTaskNotFound) {
			return nil, err
		}
		records = []*models.TaskRecord{}
	}

	// 获取用户的审核申请记录
	auditRequests, err := h.auditRequestService.GetAuditRequestListByUserID(ctx, userID)
	if err != nil {
		if !errors.Is(err, appErrors.ErrTaskNotFound) {
			return nil, err
		}
		auditRequests = []*models.CheckApplication{}
	}

	// 创建任务ID到记录的映射
	recordMap := make(map[int]*models.TaskRecord)
	for _, record := range records {
		recordMap[record.TaskID] = record
	}

	// 创建任务ID到审核申请的映射
	auditMap := make(map[int]*models.CheckApplication)
	for _, audit := range auditRequests {
		auditMap[audit.TaskID] = audit
	}

	// 转换任务列表为API响应格式
	checkinTasks := make([]struct {
		CreatedAt          int                   `json:"createdAt,omitempty"`
		Description        string                `json:"description,omitempty"`
		EndTime            int                   `json:"endTime"`
		GroupId            int                   `json:"groupId,omitempty"`
		GroupName          *string               `json:"groupName,omitempty"`
		MyCheckinStatus    gen.UserCheckinStatus `json:"myCheckinStatus,omitempty"`
		StartTime          int                   `json:"startTime"`
		Status             gen.CheckinTaskStatus `json:"status"`
		TaskId             int                   `json:"taskId,omitempty"`
		TaskName           string                `json:"taskName"`
		VerificationConfig struct {
			CheckinMethods gen.CheckinMethods `json:"checkinMethods"`
		} `json:"verificationConfig"`
	}, len(tasks))

	for i, task := range tasks {
		// 获取组信息
		group, err := h.groupsService.GetGroupByGroupID(ctx, task.GroupID)
		if err != nil {
			return nil, err
		}

		// 判断签到状态
		var myCheckinStatus gen.UserCheckinStatus
		if recordMap[task.TaskID] != nil {
			myCheckinStatus = gen.UserCheckinStatusSuccess
		} else if auditMap[task.TaskID] != nil {
			switch auditMap[task.TaskID].Status {
			case "pending":
				myCheckinStatus = gen.UserCheckinStatusPendingAudit
			case "approved":
				myCheckinStatus = gen.UserCheckinStatusAuditApproved
			case "rejected":
				myCheckinStatus = gen.UserCheckinStatusAuditRejected
			}
		} else {
			myCheckinStatus = "unchecked"
		}

		// 判断任务状态
		now := time.Now()
		var status gen.CheckinTaskStatus
		if now.Before(task.StartTime) {
			status = gen.Upcoming
		} else if now.After(task.EndTime) {
			status = gen.Expired
		} else {
			status = gen.Ongoing
		}

		checkinTasks[i] = struct {
			CreatedAt          int                   `json:"createdAt,omitempty"`
			Description        string                `json:"description,omitempty"`
			EndTime            int                   `json:"endTime"`
			GroupId            int                   `json:"groupId,omitempty"`
			GroupName          *string               `json:"groupName,omitempty"`
			MyCheckinStatus    gen.UserCheckinStatus `json:"myCheckinStatus,omitempty"`
			StartTime          int                   `json:"startTime"`
			Status             gen.CheckinTaskStatus `json:"status"`
			TaskId             int                   `json:"taskId,omitempty"`
			TaskName           string                `json:"taskName"`
			VerificationConfig struct {
				CheckinMethods gen.CheckinMethods `json:"checkinMethods"`
			} `json:"verificationConfig"`
		}{
			CreatedAt:       int(task.CreatedAt.Unix()),
			Description:     task.Description,
			EndTime:         int(task.EndTime.Unix()),
			GroupId:         task.GroupID,
			GroupName:       &group.GroupName,
			MyCheckinStatus: myCheckinStatus,
			StartTime:       int(task.StartTime.Unix()),
			Status:          status,
			TaskId:          task.TaskID,
			TaskName:        task.TaskName,
			VerificationConfig: struct {
				CheckinMethods gen.CheckinMethods `json:"checkinMethods"`
			}{
				CheckinMethods: gen.CheckinMethods{
					Gps:  task.GPS,
					Face: task.Face,
					Wifi: task.WiFi,
					Nfc:  task.NFC,
				},
			},
		}
	}

	return &gen.GetUsersMeCheckinTasks200JSONResponse{
		Code: "0",
		Data: checkinTasks,
	}, nil
}

// 用户针对某个签到任务执行签到操作。根据任务配置的校验方式提供相应数据。需要登录且是任务所属组成员
func (h *TaskHandler) PostCheckinTasksTaskIdCheckin(ctx context.Context, request gen.PostCheckinTasksTaskIdCheckinRequestObject) (gen.PostCheckinTasksTaskIdCheckinResponseObject, error) {
	// 从上下文中获取用户ID
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}

	// 获取任务信息以获取组ID
	task, err := h.taskService.GetTaskByTaskID(ctx, request.TaskId)
	if err != nil {
		if errors.Is(err, appErrors.ErrTaskNotFound) {
			return &gen.PostCheckinTasksTaskIdCheckin404JSONResponse{
				Code:    "1",
				Message: "任务不存在",
			}, nil
		}
		return nil, err
	}
	// 检查任务的校验策略并验证提供的信息是否完整
	if task.GPS {
		if request.Body.VerificationData.LocationInfo == nil {
			return &gen.PostCheckinTasksTaskIdCheckin400JSONResponse{
				Code:    "1",
				Message: "需要提供位置信息",
			}, nil
		}
		if request.Body.VerificationData.LocationInfo.Location.Latitude == 0 && request.Body.VerificationData.LocationInfo.Location.Longitude == 0 {
			return &gen.PostCheckinTasksTaskIdCheckin400JSONResponse{
				Code:    "1",
				Message: "位置信息不完整",
			}, nil
		}
	}

	if task.WiFi {
		if request.Body.VerificationData.WifiInfo == nil {
			return &gen.PostCheckinTasksTaskIdCheckin400JSONResponse{
				Code:    "1",
				Message: "需要提供WiFi信息",
			}, nil
		}
		if request.Body.VerificationData.WifiInfo.Ssid == "" || request.Body.VerificationData.WifiInfo.Bssid == "" {
			return &gen.PostCheckinTasksTaskIdCheckin400JSONResponse{
				Code:    "1",
				Message: "WiFi信息不完整",
			}, nil
		}
	}

	if task.NFC {
		if request.Body.VerificationData.NfcInfo == nil {
			return &gen.PostCheckinTasksTaskIdCheckin400JSONResponse{
				Code:    "1",
				Message: "需要提供NFC信息",
			}, nil
		}
		if request.Body.VerificationData.NfcInfo.TagId == "" || request.Body.VerificationData.NfcInfo.TagName == "" {
			return &gen.PostCheckinTasksTaskIdCheckin400JSONResponse{
				Code:    "1",
				Message: "NFC信息不完整",
			}, nil
		}
	}

	if task.Face {
		if request.Body.VerificationData.FaceData == nil {
			return &gen.PostCheckinTasksTaskIdCheckin400JSONResponse{
				Code:    "1",
				Message: "需要提供人脸信息",
			}, nil
		}
	}
	GroupId := task.GroupID

	// 检查用户是否是该组成员
	if err := h.groupsService.CheckUserExistInGroup(ctx, GroupId, userID); err != nil {
		if errors.Is(err, appErrors.ErrGroupMemberNotFound) {
			return &gen.PostCheckinTasksTaskIdCheckin403JSONResponse{
				Code:    "1",
				Message: "您不是该组成员",
			}, nil
		}
		return nil, err
	}
	// 调用服务执行签到
	record, err := h.taskService.CheckInTask(ctx, request.TaskId, userID, request.Body.VerificationData.LocationInfo.Location.Latitude, request.Body.VerificationData.LocationInfo.Location.Longitude, time.Now())
	if err != nil {
		if errors.Is(err, appErrors.ErrTaskRecordAlreadyExists) {
			return &gen.PostCheckinTasksTaskIdCheckin409JSONResponse{
				Code:    "1",
				Message: "您已经签到过该任务",
			}, nil
		}
		if errors.Is(err, appErrors.ErrTaskNotFound) {
			return &gen.PostCheckinTasksTaskIdCheckin404JSONResponse{
				Code:    "1",
				Message: "任务不存在",
			}, nil
		}
		if errors.Is(err, appErrors.ErrGroupNotFound) {
			return &gen.PostCheckinTasksTaskIdCheckin404JSONResponse{
				Code:    "1",
				Message: "用户组不存在",
			}, nil
		}
		if errors.Is(err, appErrors.ErrTaskHasEnded) {
			return &gen.PostCheckinTasksTaskIdCheckin200JSONResponse{
				Code: "0",
				Data: struct {
					RecordId   int  `json:"recordId"`
					SignedTime int  `json:"signedTime"`
					Success    bool `json:"success"`
				}{
					RecordId:   record.RecordID,
					SignedTime: int(record.SignedTime.Unix()),
					Success:    false,
				},
			}, nil
		}
		if errors.Is(err, appErrors.ErrTaskNotInRange) {
			return &gen.PostCheckinTasksTaskIdCheckin200JSONResponse{
				Code: "0",
				Data: struct {
					RecordId   int  `json:"recordId"`
					SignedTime int  `json:"signedTime"`
					Success    bool `json:"success"`
				}{
					RecordId:   record.RecordID,
					SignedTime: int(record.SignedTime.Unix()),
					Success:    false,
				},
			}, nil
		}
		return nil, err
	}

	return &gen.PostCheckinTasksTaskIdCheckin200JSONResponse{
		Code: "0",
		Data: struct {
			RecordId   int  `json:"recordId"`
			SignedTime int  `json:"signedTime"`
			Success    bool `json:"success"`
		}{
			RecordId:   record.RecordID,
			SignedTime: int(record.SignedTime.Unix()),
			Success:    true,
		},
	}, nil
}

// 用户组管理员查看某个签到任务的所有成功签到记录
func (h *TaskHandler) GetCheckinTasksTaskIdRecords(ctx context.Context, request gen.GetCheckinTasksTaskIdRecordsRequestObject) (gen.GetCheckinTasksTaskIdRecordsResponseObject, error) {
	// 从上下文中获取用户ID
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}
	// 获取任务信息以获取组ID
	task, err := h.taskService.GetTaskByTaskID(ctx, request.TaskId)
	if err != nil {
		if errors.Is(err, appErrors.ErrTaskNotFound) {
			return &gen.GetCheckinTasksTaskIdRecords404JSONResponse{
				Code:    "1",
				Message: "任务不存在",
			}, nil
		}
		return nil, err
	}

	// 检查用户是否是组的管理员
	if err := h.groupsService.CheckMemberPermission(ctx, task.GroupID, userID); err != nil {
		if errors.Is(err, appErrors.ErrGroupMemberNotFound) {
			return &gen.GetCheckinTasksTaskIdRecords404JSONResponse{
				Code:    "1",
				Message: "用户不存在",
			}, nil
		}
		if errors.Is(err, appErrors.ErrRolePermissionDenied) {
			return &gen.GetCheckinTasksTaskIdRecords403JSONResponse{
				Code:    "1",
				Message: "没有权限查看该任务",
			}, nil
		}
		return nil, err
	}
	// 调用服务获取签到记录列表
	records, err := h.taskService.GetTaskRecordsByTaskID(ctx, request.TaskId)
	if err != nil {
		if errors.Is(err, appErrors.ErrTaskNotFound) {
			return &gen.GetCheckinTasksTaskIdRecords404JSONResponse{
				Code:    "1",
				Message: "任务不存在",
			}, nil
		}
		return nil, err
	}

	// 获取任务信息
	task, err = h.taskService.GetTaskByTaskID(ctx, request.TaskId)
	if err != nil {
		if errors.Is(err, appErrors.ErrTaskNotFound) {
			return &gen.GetCheckinTasksTaskIdRecords404JSONResponse{
				Code:    "1",
				Message: "任务不存在",
			}, nil
		}
		return nil, err
	}

	// 转换为API响应格式
	var response []gen.CheckinRecord
	for _, record := range records {
		response = append(response, gen.CheckinRecord{
			RecordId:   record.RecordID,
			TaskId:     record.TaskID,
			TaskName:   record.TaskName,
			GroupId:    record.GroupID,
			GroupName:  record.GroupName,
			UserId:     record.UserID,
			Username:   record.Username,
			SignedTime: int(record.SignedTime.Unix()),
			CreatedAt:  int(record.CreatedAt.Unix()),
			CheckinMethods: gen.CheckinMethods{
				Gps:  task.GPS,
				Face: task.Face,
				Wifi: task.WiFi,
				Nfc:  task.NFC,
			},
			LocationInfo: &struct {
				Location *gen.Location `json:"location,omitempty"`
			}{
				Location: &gen.Location{
					Latitude:  record.Latitude,
					Longitude: record.Longitude,
				},
			},
		})
	}

	if len(response) == 0 {
		return &gen.GetCheckinTasksTaskIdRecords200JSONResponse{
			Code: "0",
			Data: []gen.CheckinRecord{},
		}, nil
	}

	return &gen.GetCheckinTasksTaskIdRecords200JSONResponse{
		Code: "0",
		Data: response,
	}, nil
}

// 获取当前用户的所有签到记录
func (h *TaskHandler) GetUsersMeCheckinRecords(ctx context.Context, request gen.GetUsersMeCheckinRecordsRequestObject) (gen.GetUsersMeCheckinRecordsResponseObject, error) {
	// 从上下文中获取用户ID
	userID, ok := ctx.Value("userID").(int)
	if !ok {
		return nil, appErrors.ErrJwtParseFailed
	}

	// 调用服务获取签到记录列表
	records, err := h.taskService.GetTaskRecordsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// 转换为API响应格式
	var response []gen.CheckinRecord
	for _, record := range records {
		// 获取对应任务的信息
		task, err := h.taskService.GetTaskByTaskID(ctx, record.TaskID)
		if err != nil {
			continue // 如果任务不存在，跳过该记录
		}

		response = append(response, gen.CheckinRecord{
			RecordId:   record.RecordID,
			TaskId:     record.TaskID,
			TaskName:   record.TaskName,
			GroupId:    record.GroupID,
			GroupName:  record.GroupName,
			UserId:     record.UserID,
			Username:   record.Username,
			SignedTime: int(record.SignedTime.Unix()),
			CreatedAt:  int(record.CreatedAt.Unix()),
			CheckinMethods: gen.CheckinMethods{
				Gps:  task.GPS,
				Face: task.Face,
				Wifi: task.WiFi,
				Nfc:  task.NFC,
			},
			LocationInfo: &struct {
				Location *gen.Location `json:"location,omitempty"`
			}{
				Location: &gen.Location{
					Latitude:  record.Latitude,
					Longitude: record.Longitude,
				},
			},
		})
	}
	if len(response) == 0 {
		return &gen.GetUsersMeCheckinRecords200JSONResponse{
			Code: "0",
			Data: []gen.CheckinRecord{},
		}, nil
	}

	return &gen.GetUsersMeCheckinRecords200JSONResponse{
		Code: "0",
		Data: response,
	}, nil
}
