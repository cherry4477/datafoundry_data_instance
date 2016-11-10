package api

import (
	"crypto/md5"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/asiainfoLDP/datafoundry_data_instance/log"
	"github.com/asiainfoLDP/datafoundry_data_instance/models"
	"github.com/julienschmidt/httprouter"
	"io"
	mathrand "math/rand"
	"net/http"
	"os"
)

const (
	letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"
	randNumber  = "0123456789"
)

var (
//mysqlHost     = getenv("REMOTEHOST")
//mysqlPort     = getenv("REMOTEPORT")
//mysqlPassword = getenv("REMOTEPASSWORD")
)

var logger = log.GetLogger()

func QueryServiceList(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	logger.Info("Request url: GET %v.", r.URL)

	logger.Info("Begin query service list handler.")

	username, e := validateAuth(r.Header.Get("Authorization"))
	if e != nil {
		JsonResult(w, http.StatusUnauthorized, e, nil)
		return
	}
	logger.Debug("username:%v", username)

	db := models.GetDB()
	if db == nil {
		logger.Warn("Get db is nil.")
		JsonResult(w, http.StatusInternalServerError, GetError(ErrorCodeDbNotInitlized), nil)
		return
	}

	r.ParseForm()
	class := r.Form.Get("class")
	provider := r.Form.Get("provider")

	offset, size := OptionalOffsetAndSize(r, 30, 1, 100)
	orderBy := models.ValidateOrderBy(r.Form.Get("orderby"))
	sortOrder := models.ValidateSortOrder(r.Form.Get("sortorder"), false)

	count, apps, err := models.QueryServices(db, class, provider, orderBy, sortOrder, offset, size)
	if err != nil {
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeQueryServices, err.Error()), nil)
		return
	}

	logger.Info("End query service list handler.")
	JsonResult(w, http.StatusOK, nil, NewQueryListResult(count, apps))
}

func CreateInstance(w http.ResponseWriter, r *http.Request, params httprouter.Params) {
	logger.Info("Request url: POST %v.", r.URL)
	logger.Info("Begin create instance handler.")

	db := models.GetDB()
	if db == nil {
		logger.Warn("Get db is nil.")
		JsonResult(w, http.StatusInternalServerError, GetError(ErrorCodeDbNotInitlized), nil)
		return
	}

	username, e := validateAuth(r.Header.Get("Authorization"))
	if e != nil {
		JsonResult(w, http.StatusUnauthorized, e, nil)
		return
	}
	logger.Info("username:%v", username)

	serviceId := params.ByName("id")
	serviceinfo, err := models.GetServiceInfo(db, serviceId)
	if err != nil {
		logger.Error("Catch err: %v", err)
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeGetServiceInfo, err.Error()), nil)
		return
	}

	//newUsername, newPassword, err := grant(serviceinfo)
	if err != nil {
		logger.Error("Catch err: %v.", err)
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeGrantUser, err.Error()), nil)
		return
	}

	//instance := models.Instance{
	//	Host:              serviceinfo.Address,
	//	Port:              serviceinfo.Port,
	//	Instance_data:     serviceinfo.Service_data,
	//	Instance_username: newUsername,
	//	Instance_password: newPassword,
	//	Uri:               "mysql://" + newUsername + ":" + newPassword + "@" + serviceinfo.Address + ":" + serviceinfo.Port + "/" + serviceinfo.Service_data,
	//	Username:          username,
	//}
	result := models.CreateResult{
		Uri: "mysql://" + serviceinfo.Username + ":" + serviceinfo.Password +
			"@" + serviceinfo.Address + ":" + serviceinfo.Port + "/" + serviceinfo.Service_data,
		Hostname: serviceinfo.Address,
		Port:     serviceinfo.Port,
		Name:     serviceinfo.Service_data,
		Username: serviceinfo.Username,
		Password: serviceinfo.Password,
	}

	//result, err := models.CreateInstance(db, &instance)
	//if err != nil {
	//	logger.Error("Create plan err: %v", err)
	//	JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeCreateInstance, err.Error()), nil)
	//	return
	//}

	logger.Info("End create instance handler.")
	JsonResult(w, http.StatusOK, nil, result)
}

func grant(info *models.ServiceInfo) (string, string, error) {
	//初始化mysql的链接串
	db, err := sql.Open("mysql", info.Username+":"+info.Password+"@tcp("+info.Address+":"+info.Port+")/")

	if err != nil {
		logger.Error("sql open err: %v", err)
		return "", "", err
	}
	//测试是否能联通
	err = db.Ping()
	if err != nil {
		logger.Error("ping err: %v", err)
		return "", "", err
	}
	defer db.Close()

	newusername := getguid()[0:15]
	newpassword := getguid()[0:15]

	_, err = db.Query("GRANT SELECT ON " + info.Service_data + ".* TO '" + newusername + "'@'%' IDENTIFIED BY '" + newpassword + "'")
	if err != nil {
		logger.Error("db query err: %v", err)
		return "", "", err
	}

	logger.Info("username: %s, password: %s.", newusername, newpassword)
	return newusername, newpassword, nil
}

func getguid() string {
	b := make([]byte, 48)

	if _, err := io.ReadFull(rand.Reader, b); err != nil {
		return ""
	}
	return getmd5string(base64.URLEncoding.EncodeToString(b))
}

func getmd5string(s string) string {
	h := md5.New()
	h.Write([]byte(s))
	return hex.EncodeToString(h.Sum(nil))
}

func getenv(env string) string {
	env_value := os.Getenv(env)
	if env_value == "" {
		logger.Emergency("Need env %s.", env)
		os.Exit(2)
	}
	fmt.Println("ENV:", env, env_value)
	return env_value
}

func validateAuth(token string) (string, *Error) {
	if token == "" {
		return "", GetError(ErrorCodeAuthFailed)
	}

	username, err := getDFUserame(token)
	if err != nil {
		return "", GetError2(ErrorCodeAuthFailed, err.Error())
	}

	return username, nil
}

func canEditSaasApps(username string) bool {
	return username == "datafoundry"
}

func genUUID() string {
	bs := make([]byte, 16)
	_, err := rand.Read(bs)
	if err != nil {
		logger.Warn("genUUID error: ", err.Error())

		mathrand.Read(bs)
	}

	return fmt.Sprintf("%X-%X-%X-%X-%X", bs[0:4], bs[4:6], bs[6:8], bs[8:10], bs[10:])
}
