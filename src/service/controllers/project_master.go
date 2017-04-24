package controllers

import (
	"encoding/json"
	"errors"
	"service/models"
	"strconv"
	"strings"
	"github.com/astaxie/beego"
	"time"
)

// Записаться и отписаться от проекта, получить список записанных
type ProjectMasterController struct {
	ControllerWithAuthorization
}

// URLMapping ...
func (c *ProjectMasterController) URLMapping() {
	c.Mapping("Post", c.Post)
	c.Mapping("GetOne", c.GetOne)
	c.Mapping("GetAll", c.GetAll)
	c.Mapping("Put", c.Put)
	c.Mapping("Delete", c.Delete)
}
// Param   project_id      query   int     true    "ID проекта, на который нужно записаться"

// TODO: слишком много вложений if-в, нужно как-то разделить на несколько функций
// Post ...
// @Title Post
// @Description Добавить мастера к проекту
// @Param   user_id         query   string  true    "ID пользователя, корого нужно сделать мастером"
// @Param   project_id      query   string  true    "ID проекта, на который нужно добавить мастера"
// @Param   Bearer-token    header  string  true    "Токен доступа мастера проекта или админа"
// @Success 201 {int} "Created"
// @Failure 403 body is empty
// @router / [post]
func (c *ProjectMasterController) Post() {
	if c.CurrentUser.PermissionLevel == -1 {
		beego.Debug(c.Ctx.Input.IP(), "Access denied for `Post` new master for project")
		c.Ctx.Output.SetStatus(403)
		c.Data["json"] = "Forbidden"
	} else {
		// получить id проекта, на который нужно добавить мастера
		project_id, err := c.GetInt("project_id")
		if err != nil {
			beego.Debug(c.Ctx.Input.IP(), "Not an int param. Should be int", err.Error())
			c.Ctx.Output.SetStatus(400)
			c.Data["json"] = err.Error()
		} else {
			masters_of_this_project, err := models.GetMastersOfTheProject(project_id)
			if err != nil {
				beego.Debug(c.Ctx.Input.IP(), "Wrong project id", err.Error())
				c.Ctx.Output.SetStatus(400)
				c.Data["json"] = err.Error()
			} else if c.CurrentUser.PermissionLevel == 2 ||
					models.IsUserInArray(c.CurrentUser.UserId, masters_of_this_project) {

				// проект, на который записывается пользователь
				project, err := models.GetProjectById(project_id)
				if err != nil {
					beego.Debug("Wrong project id", err.Error())
					c.Data["json"] = err.Error()
					c.Ctx.Output.SetStatus(400)
				} else {
					// пользователь, которого назначают мастером
					user_id, err := c.GetInt("user_id")
					if err != nil {
						beego.Debug(c.Ctx.Input.IP(), "Not an int param. Should be int", err.Error())
						c.Ctx.Output.SetStatus(400)
						c.Data["json"] = err.Error()
					} else {
						user, err := models.GetUserById(user_id)
						if err != nil {
							beego.Debug("Wrong user id", err.Error())
							c.Ctx.Output.SetStatus(400)
							c.Data["json"] = err.Error()
						} else {
							// записать пользователя
							beego.Trace("Good user_id and project_id and Bearer-token")
							v := models.ProjectMaster{
								MasterId: user,
								ProjectId: project,
								SignedDate: time.Now(),
							}

							err := models.AddMasterToProject(&v)
							if err != nil {
								beego.Critical("Can't add master", err.Error())
								c.Ctx.Output.SetStatus(500)
								c.Data["json"] = err.Error()
							} else {
								beego.Trace("Welcome, master")
								c.Ctx.Output.SetStatus(201)
								c.Data["json"] = "Created"
							}
						}
					}
				}
			}
		}
	}
	c.ServeJSON()
}

// GetOne ...
// @Title Get One
// @Description Получить список записанных пользователей
// @Param   id  path    string  true    "ID проекта, на которых нужно узнать список записанных"
// @Success 200 []{int} список из ID пользователей
// @Failure 403 :id is empty
// @router /:id [get]
func (c *ProjectMasterController) GetOne() {
	beego.Trace("New GET for singed up users")
	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)
	v, err := models.GetAllSignedUpOnProject(id)
	if err != nil {
		beego.Debug("GET list of signed up users error", err.Error())
		c.Data["json"] = err.Error()
		c.Ctx.Output.SetStatus(500)
	} else {
		beego.Trace("Success GET")
		c.Data["json"] = v
	}
	c.ServeJSON()
}

// GetAll ...
// @Title Get All
// @Description get ProjectSignUp
// @Param	query	query	string	false	"Filter. e.g. col1:v1,col2:v2 ..."
// @Param	fields	query	string	false	"Fields returned. e.g. col1,col2 ..."
// @Param	sortby	query	string	false	"Sorted-by fields. e.g. col1,col2 ..."
// @Param	order	query	string	false	"Order corresponding to each sortby field, if single value, apply to all sortby fields. e.g. desc,asc ..."
// @Param	limit	query	string	false	"Limit the size of result set. Must be an integer"
// @Param	offset	query	string	false	"Start position of result set. Must be an integer"
// @Success 200 {object} models.ProjectSignUp
// @Failure 403
// @router / [get]

// wtf
func (c *ProjectMasterController) GetAll() {
	var fields []string
	var sortby []string
	var order []string
	var query = make(map[string]string)
	var limit int64 = 10
	var offset int64

	// fields: col1,col2,entity.col3
	if v := c.GetString("fields"); v != "" {
		fields = strings.Split(v, ",")
	}
	// limit: 10 (default is 10)
	if v, err := c.GetInt64("limit"); err == nil {
		limit = v
	}
	// offset: 0 (default is 0)
	if v, err := c.GetInt64("offset"); err == nil {
		offset = v
	}
	// sortby: col1,col2
	if v := c.GetString("sortby"); v != "" {
		sortby = strings.Split(v, ",")
	}
	// order: desc,asc
	if v := c.GetString("order"); v != "" {
		order = strings.Split(v, ",")
	}
	// query: k:v,k:v
	if v := c.GetString("query"); v != "" {
		for _, cond := range strings.Split(v, ",") {
			kv := strings.SplitN(cond, ":", 2)
			if len(kv) != 2 {
				c.Data["json"] = errors.New("Error: invalid query key/value pair")
				c.ServeJSON()
				return
			}
			k, v := kv[0], kv[1]
			query[k] = v
		}
	}

	l, err := models.GetAllProjectAuthor(query, fields, sortby, order, offset, limit)
	if err != nil {
		c.Data["json"] = err.Error()
	} else {
		c.Data["json"] = l
	}
	c.ServeJSON()
}

// Put ...
// @Title Put
// @Description update the ProjectSignUp
// @Param	id		path 	string	true		"The id you want to update"
// @Param	body		body 	models.ProjectSignUp	true		"body for ProjectSignUp content"
// @Success 200 {object} models.ProjectSignUp
// @Failure 403 :id is not int
// @router /:id [put]

// wtf
func (c *ProjectMasterController) Put() {
	idStr := c.Ctx.Input.Param(":id")
	id, _ := strconv.Atoi(idStr)
	v := models.ProjectSignUp{Id: id}
	if err := json.Unmarshal(c.Ctx.Input.RequestBody, &v); err == nil {
		if err := models.UpdateProjectAuthorById(&v); err == nil {
			c.Data["json"] = "OK"
		} else {
			c.Data["json"] = err.Error()
		}
	} else {
		c.Data["json"] = err.Error()
	}
	c.ServeJSON()
}

// Delete ...
// @Title Delete
// @Description Отписаться от проекта
// @Param   id              path    string  true    "ID проекта, из которого нужно удалить заявку"
// @Param   Bearer-token    header  string  true    "Токен доступа любого зарегистрированного пользователя"
// @Success 200 {string} delete success!
// @Failure 403 id is empty
// @router /:id [delete]
func (c *ProjectMasterController) Delete() {
	beego.Trace("User want to sign out from project")
	if c.CurrentUser.PermissionLevel == -1 {
		beego.Debug(c.Ctx.Input.IP(), "Access denied for `Delete` project_sign_up")
		c.Ctx.Output.SetStatus(403)
		c.Data["json"] = "Forbidden"
	} else {
		idStr := c.Ctx.Input.Param(":id")
		id, err := strconv.Atoi(idStr)
		if err != nil {
			beego.Debug("Can't parse", err.Error())
			c.Ctx.Output.SetStatus(400)
			c.Data["json"] = err.Error()
		}
		if err := models.DeleteProjectSignUp(c.CurrentUser.UserId, id); err == nil {
			beego.Trace("Success sign out from project")
			c.Data["json"] = "OK"
		} else {
			beego.Debug("Can't delete from ProjectSingUp", err.Error())
			c.Ctx.Output.SetStatus(400)
			c.Data["json"] = err.Error()
		}
	}
	c.ServeJSON()
}
