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
	"net/http"
	"os"
)

const (
	letterBytes = "abcdefghijklmnopqrstuvwxyz0123456789"
	randNumber  = "0123456789"
)

var (
	mysqlHost     = getenv("REMOTEHOST")
	mysqlPort     = getenv("REMOTEPORT")
	mysqlPassword = getenv("REMOTEPASSWORD")
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
	logger.Info("Request url: GET %v.", r.URL)
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
	logger.Debug("username:%v", username)

	r.ParseForm()
	dbname := r.Form.Get("name")
	logger.Debug("name: %s.", dbname)

	newUsername, newPassword, err := grant(dbname)
	if err != nil {
		logger.Error("Catch err: %v.", err)
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeGrantUser, err.Error()), nil)
	}

	instance := models.Instance{
		Host:              mysqlHost,
		Port:              mysqlPort,
		Instance_data:     dbname,
		Instance_username: newUsername,
		Instance_password: newPassword,
		Uri:               "mysql://" + newUsername + ":" + newPassword + "@" + mysqlHost + ":" + mysqlPort + "/" + dbname,
		Username:          username,
	}

	result, err := models.CreateInstance(db, &instance)
	if err != nil {
		logger.Error("Create plan err: %v", err)
		JsonResult(w, http.StatusBadRequest, GetError2(ErrorCodeCreateInstance, err.Error()), nil)
		return
	}

	logger.Info("End create instance handler.")
	JsonResult(w, http.StatusOK, nil, result)
}

func grant(dbname string) (string, string, error) {
	//初始化mysql的链接串
	db, err := sql.Open("mysql", "root:"+mysqlPassword+"@tcp("+mysqlHost+":"+mysqlPort+")/")

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

	_, err = db.Query("GRANT SELECT ON " + dbname + ".* TO '" + newusername + "'@'%' IDENTIFIED BY '" + newpassword + "'")
	if err != nil {
		logger.Error("db query err: %v", err)
		return "", "", err
	}

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
