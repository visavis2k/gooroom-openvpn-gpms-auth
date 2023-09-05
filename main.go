package main

import (
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"github.com/hashicorp/go-hclog"
	"net/http"
	"net/url"
	"os"
)

var (
	logger hclog.Logger

	caFilePath   = "/etc/openvpn/client/root_cacert.pem"
	certFilePath = "/etc/openvpn/client/gooroom_client.crt"
	keyFilePath  = "/etc/openvpn/client/gooroom_client.key"

	authEndpoint = "https://glm.javaworld.co.kr/glm/v1/pam/authconfirm"
)

func main() {
	logger = hclog.Default()
	logger.SetLevel(hclog.Debug)

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
		logger.Error("Error lading cert and key file", "error", err)
	}

	certPool, err := x509.SystemCertPool()
	if err != nil {
		panic(err)
	}

	if ca, err := os.ReadFile(caFilePath); err != nil {
		panic(err)
	} else if ok := certPool.AppendCertsFromPEM(ca); !ok {
		panic("invalid cert in CA PEM")
	}

	tlsConfig := &tls.Config{
		RootCAs:      certPool,
		Certificates: []tls.Certificate{cert},
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

	logger.Debug("Response", "StatusCode", resp.StatusCode, "Status")
	logger.Debug("Response", "Body", resp.Body)

	var res response
	json.NewDecoder(resp.Body).Decode(&res)

	if res.status.Result == "SUCCESS" {
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

type response struct {
	status struct {
		Result     string `json:"result"`
		ResultCode string `json:"resultCode"`
		ErrMsg     string `json:"errMsg"`
	}
}
