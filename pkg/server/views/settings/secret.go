package settings

import (
	"github.com/kubespace/kubespace/pkg/model"
	"github.com/kubespace/kubespace/pkg/model/types"
	"github.com/kubespace/kubespace/pkg/server/views"
	"github.com/kubespace/kubespace/pkg/server/views/serializers"
	"github.com/kubespace/kubespace/pkg/utils"
	"github.com/kubespace/kubespace/pkg/utils/code"
	"net/http"
	"strconv"
	"time"
)

type SettingsSecret struct {
	Views  []*views.View
	models *model.Models
}

func NewSettingsSecret(models *model.Models) *SettingsSecret {
	secret := &SettingsSecret{
		models: models,
	}
	vs := []*views.View{
		views.NewView(http.MethodGet, "", secret.list),
		views.NewView(http.MethodPost, "", secret.create),
		views.NewView(http.MethodPut, "/:id", secret.update),
		views.NewView(http.MethodDelete, "/:id", secret.delete),
	}
	secret.Views = vs
	return secret
}

func (s *SettingsSecret) create(c *views.Context) *utils.Response {
	var ser serializers.SecretsSerializers
	resp := &utils.Response{Code: code.Success}
	if err := c.ShouldBind(&ser); err != nil {
		resp.Code = code.ParamsError
		resp.Msg = err.Error()
		return resp
	}
	secret := &types.SettingsSecret{
		Name:        ser.Name,
		Description: ser.Description,
		Type:        ser.SecretType,
		User:        ser.User,
		Password:    ser.Password,
		PrivateKey:  ser.PrivateKey,
		AccessToken: ser.AccessToken,
		CreateUser:  c.User.Name,
		UpdateUser:  c.User.Name,
		CreateTime:  time.Time{},
		UpdateTime:  time.Time{},
	}
	_, err := s.models.SettingsSecretManager.Create(secret)
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = "创建密钥失败: " + err.Error()
		return resp
	}
	return resp
}

func (s *SettingsSecret) update(c *views.Context) *utils.Response {
	var ser serializers.SecretsSerializers
	resp := &utils.Response{Code: code.Success}
	if err := c.ShouldBind(&ser); err != nil {
		resp.Code = code.ParamsError
		resp.Msg = err.Error()
		return resp
	}
	secretId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	secret, err := s.models.SettingsSecretManager.Get(uint(secretId))
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = "获取密钥失败: " + err.Error()
		return resp
	}
	secret.Description = ser.Description
	secret.Type = ser.SecretType
	secret.User = ser.User
	secret.Password = ser.Password
	secret.PrivateKey = ser.PrivateKey
	secret.AccessToken = ser.AccessToken
	secret.UpdateUser = c.User.Name
	secret.UpdateTime = time.Now()
	_, err = s.models.SettingsSecretManager.Update(secret)
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = "更新密钥失败: " + err.Error()
		return resp
	}
	return resp
}

func (s *SettingsSecret) delete(c *views.Context) *utils.Response {
	resp := &utils.Response{Code: code.Success}
	secretId, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return &utils.Response{Code: code.ParamsError, Msg: err.Error()}
	}
	secret, err := s.models.SettingsSecretManager.Get(uint(secretId))
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = "获取密钥失败: " + err.Error()
		return resp
	}
	err = s.models.SettingsSecretManager.Delete(secret)
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = "删除密钥失败: " + err.Error()
		return resp
	}
	return resp
}

func (s *SettingsSecret) list(c *views.Context) *utils.Response {
	resp := &utils.Response{Code: code.Success}
	secrets, err := s.models.SettingsSecretManager.List()
	if err != nil {
		resp.Code = code.DBError
		resp.Msg = err.Error()
		return resp
	}
	var data []map[string]interface{}

	for _, secret := range secrets {
		data = append(data, map[string]interface{}{
			"id":          secret.ID,
			"name":        secret.Name,
			"description": secret.Description,
			"type":        secret.Type,
			"user":        secret.User,
			"create_user": secret.CreateUser,
			"update_user": secret.UpdateUser,
			"create_time": secret.CreateTime,
			"update_time": secret.UpdateTime,
		})
	}
	resp.Data = data
	return resp
}
