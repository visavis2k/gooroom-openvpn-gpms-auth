package main

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"io/fs"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"

	"github.com/hashicorp/go-hclog"
	"github.com/spf13/viper"
)

var (
	logger hclog.Logger

	caFilePath   string
	certFilePath string
	keyFilePath  string

	authEndpoint string
)

func main() {
	logger = hclog.Default()
	logger.SetLevel(hclog.Debug)

	viper.SetConfigFile("/etc/gooroom/gooroom-client-server-register/gcsr.conf")
	viper.SetConfigType("props")
	viper.SetDefault("caFilePath", "/etc/openvpn/client/root_cacert.pem")
	viper.SetDefault("certFilePath", "/etc/openvpn/client/gooroom_client.crt")
	viper.SetDefault("keyFilePath", "/etc/openvpn/client/gooroom_client.key")
	viper.SetDefault("authEndpoint", "https://glm.javaworld.co.kr/glm/v1/pam/authconfirm")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(*fs.PathError); ok {
			// ignore
		} else if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			// ignore
		} else {
			logger.Error("Error loading config", "error", err)
			os.Exit(5)
		}
	}

	caFilePath = viper.GetString("caFilePath")
	certFilePath = viper.GetString("certFilePath")
	keyFilePath = viper.GetString("keyFilePath")
	authEndpoint = viper.GetString("authEndpoint")

	data := request{
		username: os.Getenv("username"),
		password: os.Getenv("password"),
	}

	controlFile := os.Getenv("auth_control_file")

	success := data.authenticate()

	writeStatus(success, data.username, controlFile)
}

func (req *request) authenticate() bool {
	cert, err := tls.LoadX509KeyPair(certFilePath, keyFilePath)
	if err != nil {
		logger.Error("Error loding cert and key file", "error", err)
	}

	certPool, err := x509.SystemCertPool()
	if err != nil {
		panic(err)
	}

	if ca, err := ioutil.ReadFile(caFilePath); err != nil {
		panic(err)
	} else if ok := certPool.AppendCertsFromPEM(ca); !ok {
		panic("invalid cert in CA PEM")
	}

	tlsConfig := &tls.Config{
		RootCAs:            certPool,
		Certificates:       []tls.Certificate{cert},
		InsecureSkipVerify: true,
	}

	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	client := &http.Client{Transport: tr}

	data := url.Values{
		"user_id": {req.username},
		"user_pw": {sha256Hex(req.username, req.password)},
	}

	resp, err := client.PostForm(authEndpoint, data)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	logger.Debug("Response", "StatusCode", resp.StatusCode, "Status", resp.Status)

	var res Response
	json.NewDecoder(resp.Body).Decode(&res)

	logger.Debug("Decoded Response", "Status", res.Status, "Result", res.Status.Result, "ResultCode", res.Status.ResultCode)

	if res.Status.Result == "SUCCESS" {
		return true
	} else {
		return false
	}
}

func writeStatus(success bool, username, controlFile string) {
	file, err := os.OpenFile(controlFile, os.O_RDWR, 0755)
	if err != nil {
		logger.Debug("Error opening control file", "error", err)
		return
	}
	defer file.Close()

	if success {
		logger.Debug("Authorization was successful", "Username", username)
		file.WriteString("1")
	} else {
		logger.Debug("Authorization WAS NOT successful", "Username", username)
		file.WriteString("0")
	}
}

func sha256Hex(username, password string) string {
	hash := sha256.New()
	hash.Write([]byte(password))
	hashedPassword := hash.Sum(nil)

	hash.Reset()
	hash.Write([]byte(username + hex.EncodeToString(hashedPassword)))

	return hex.EncodeToString(hash.Sum(nil))
}

type request struct {
	username string
	password string
}

type Response struct {
	Status struct {
		Result     string `json:"result"`
		ResultCode string `json:"resultCode"`
		ErrMsg     string `json:"errMsg"`
	} `json:"status"`
}
