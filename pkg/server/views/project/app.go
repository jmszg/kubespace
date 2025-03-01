package project

import (
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/server/config"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	"github.com/kubespace/kubespace/pkg/service/project"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"k8s.io/klog/v2"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type ProjectApp struct {
	Views      []*views.View
	AppService *project.AppService
	models     *model.Models
}

func NewProjectApp(config *config.ServerConfig) *ProjectApp {
	app := &ProjectApp{
		AppService: config.ServiceFactory.Project.AppService,
		models:     config.Models,
	}
	vs := []*views.View{
		views.NewView(http.MethodGet, "", app.listApps),
		views.NewView(http.MethodGet, "/versions", app.listAppVersions),
		views.NewView(http.MethodGet, "/version/:id", app.getAppVersion),
		views.NewView(http.MethodGet, "/status", app.listAppStatus),
		views.NewView(http.MethodGet, "/status_sse", app.statusSSE),
		views.NewView(http.MethodGet, "/download", app.downloadChart),
		views.NewView(http.MethodGet, "/:id", app.getApp),
		views.NewView(http.MethodPost, "", app.create),
		views.NewView(http.MethodPost, "/install", app.install),
		views.NewView(http.MethodPost, "/destroy", app.destroy),
		views.NewView(http.MethodPost, "/import_storeapp", app.importStoreapp),
		views.NewView(http.MethodPost, "/import_custom_app", app.importCustomApp),
		views.NewView(http.MethodPost, "/duplicate_app", app.duplicateApp),
		views.NewView(http.MethodDelete, "/version/:id", app.deleteAppVersion),
		views.NewView(http.MethodDelete, "/:id", app.deleteApp),
	}
	app.Views = vs
	return app
}

func (a *ProjectApp) create(c *views.Context) *utils.Response {
	var ser serializers.ProjectCreateAppSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return a.AppService.CreateProjectApp(c.User, ser)
}

func (a *ProjectApp) listApps(c *views.Context) *utils.Response {
	var ser serializers.ProjectAppListSerializer
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	data, err := a.AppService.ListApp(ser.Scope, ser.ScopeId)
	if err != nil {
		return &utils.Response{Code: code.GetError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success, Data: data}
}

func (a *ProjectApp) listAppStatus(c *views.Context) *utils.Response {
	var ser serializers.ProjectAppListSerializer
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return a.AppService.ListAppStatus(ser)
}

func (a *ProjectApp) getApp(c *views.Context) *utils.Response {
	appId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return a.AppService.GetApp(uint(appId))
}

func (a *ProjectApp) deleteApp(c *views.Context) *utils.Response {
	appId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	err = a.models.ProjectAppManager.DeleteProjectApp(uint(appId))
	if err != nil {
		return &utils.Response{Code: code.DBError, Msg: err.Error()}
	}
	return &utils.Response{Code: code.Success}
}

func (a *ProjectApp) listAppVersions(c *views.Context) *utils.Response {
	var ser serializers.ProjectAppVersionListSerializer
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return a.AppService.ListAppVersions(ser)
}

func (a *ProjectApp) getAppVersion(c *views.Context) *utils.Response {
	appVersionId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return a.AppService.GetAppVersion(uint(appVersionId))
}

func (a *ProjectApp) deleteAppVersion(c *views.Context) *utils.Response {
	appVersionId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	if err = a.models.ProjectAppVersionManager.DeleteVersion(uint(appVersionId)); err != nil {
		return &utils.Response{Code: code.DBError, Msg: "删除应用版本失败：" + err.Error()}
	}
	return &utils.Response{Code: code.Success}
}

func (a *ProjectApp) install(c *views.Context) *utils.Response {
	var ser serializers.ProjectInstallAppSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return a.AppService.InstallApp(c.User, ser)
}

func (a *ProjectApp) destroy(c *views.Context) *utils.Response {
	var ser serializers.ProjectInstallAppSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return a.AppService.DestroyApp(c.User, ser)
}

func (a *ProjectApp) importStoreapp(c *views.Context) *utils.Response {
	var ser serializers.ImportStoreAppSerializers
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return a.AppService.ImportStoreApp(ser, c.User)
}

func (a *ProjectApp) importCustomApp(c *views.Context) *utils.Response {
	var ser serializers.ImportCustomAppSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}

	file, err := c.FormFile("file")
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: "get chart file error: " + err.Error()}
	}
	chartIn, err := file.Open()
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: "get chart file error: " + err.Error()}
	}
	return a.AppService.ImportCustomApp(c.User, ser, chartIn)
}

func (a *ProjectApp) duplicateApp(c *views.Context) *utils.Response {
	var ser serializers.DuplicateAppSerializer
	if err := c.ShouldBind(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	return a.AppService.DuplicateApp(&ser, c.User)
}

func (a *ProjectApp) statusSSE(c *views.Context) *utils.Response {
	var ser serializers.ProjectAppListSerializer
	if err := c.ShouldBindQuery(&ser); err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	w := c.Writer
	clientGone := w.CloseNotify()
	c.SSEvent("message", "{}")
	w.Flush()
	//c.Stream()
	tick := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-clientGone:
			klog.Info("app status client gone")
			return nil
		case <-tick.C:
			tick.Stop()
			res := a.AppService.ListAppStatus(ser)
			c.SSEvent("message", res)
			w.Flush()
			tick.Reset(5 * time.Second)
		}
	}
}

func (a *ProjectApp) downloadChart(c *views.Context) *utils.Response {
	chartPath := c.Query("path")
	if chartPath == "" {
		c.JSON(http.StatusNotFound, &utils.Response{Code: code.GetError, Msg: "not found chart params"})
		return nil
	}

	appChart, err := a.models.ProjectAppVersionManager.GetAppVersionChart(chartPath)
	if err != nil {
		c.JSON(http.StatusNotFound, &utils.Response{Code: code.GetError, Msg: err.Error()})
		return nil
	}
	chartName := chartPath
	s := strings.Split(chartPath, "/")
	if len(s) >= 2 {
		chartName = s[len(s)-1]
	}
	fileContentDisposition := "attachment;filename=\"" + chartName + "\""
	c.Header("Content-Disposition", fileContentDisposition)
	c.Data(http.StatusOK, "application/x-tar", appChart.Content)
	return nil
}
