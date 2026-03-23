package dkim

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	b64 "encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/midoks/imail/internal/tools"
)

func makeRsa() ([]byte, []byte, error) {
	privatekey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err == nil {
		var publickey *rsa.PublicKey
		publickey = &privatekey.PublicKey
		Priv := x509.MarshalPKCS1PrivateKey(privatekey)
		Pub, err := x509.MarshalPKIXPublicKey(publickey)

		if err == nil {
			return Priv, Pub, nil
		}
	}
	return []byte{}, []byte{}, err
}

func CheckDomainA(domain string) error {
	findIp, err := net.LookupIP(domain)
	if err != nil {
		return err
	}

	ip, err := tools.GetPublicIP()
	if err != nil {
		return err
	}

	var isFind = false
	for _, fIp := range findIp {
		if strings.EqualFold(fIp.String(), ip) {
			isFind = true
			break
		}
	}

	if !isFind {
		return errors.New("IP not configured by domain name!")
	}

	return nil
}

func sanitizeDomain(domain string) string {
	// Validate that the domain is a syntactically valid DNS name and contains no path separators.
	if domain == "" {
		return ""
	}
	// Reject obvious path separators early.
	if strings.ContainsAny(domain, "/\\") {
		return ""
	}

	labels := strings.Split(domain, ".")
	if len(labels) == 0 {
		return ""
	}

	for _, label := range labels {
		// Each label must be 1-63 characters.
		if len(label) == 0 || len(label) > 63 {
			return ""
		}
		// Labels must start and end with a letter or digit.
		if !((label[0] >= 'a' && label[0] <= 'z') ||
			(label[0] >= 'A' && label[0] <= 'Z') ||
			(label[0] >= '0' && label[0] <= '9')) {
			return ""
		}
		last := label[len(label)-1]
		if !((last >= 'a' && last <= 'z') ||
			(last >= 'A' && last <= 'Z') ||
			(last >= '0' && last <= '9')) {
			return ""
		}
		// The rest of the label may contain letters, digits, or hyphens.
		for i := 1; i < len(label)-1; i++ {
			ch := label[i]
			if !((ch >= 'a' && ch <= 'z') ||
				(ch >= 'A' && ch <= 'Z') ||
				(ch >= '0' && ch <= '9') ||
				ch == '-') {
				return ""
			}
		}
	}

	// Overall domain length must not exceed 253 characters.
	if len(domain) > 253 {
		return ""
	}

	return domain
}

func MakeDkimFile(path, domain string) (string, error) {
	// Sanitize domain to prevent path traversal
	safeDomain := sanitizeDomain(domain)
	if safeDomain == "" {
		return "", errors.New("invalid domain name")
	}

	priFile := filepath.Join(path, "dkim", safeDomain, "default.private")
	defalutTextFile := filepath.Join(path, "dkim", safeDomain, "default.txt")
	defalutValFile := filepath.Join(path, "dkim", safeDomain, "default.val")

	if tools.IsExist(priFile) {
		pubContent, _ := tools.ReadFile(defalutTextFile)
		return pubContent, nil
	}

	Priv, Pub, err := makeRsa()
	if err != nil {
		return "", err
	}

	block := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: Priv,
	}

	// pri := b64.StdEncoding.EncodeToString(Priv)
	file, err := os.Create(priFile)
	if err != nil {
		return "", err
	}

	err = pem.Encode(file, block)
	if err != nil {
		return "", err
	}

	pub := b64.StdEncoding.EncodeToString(Pub)
	pubContent := fmt.Sprintf("default._domainkey\tIN\tTXT\t(\r\nv=DKIM1;k=rsa;p=%s\r\n)\r\n----- DKIM key default for %s", pub, domain)

	err = tools.WriteFile(defalutTextFile, pubContent)
	if err == nil {
		err = tools.WriteFile(defalutValFile, fmt.Sprintf("v=DKIM1;k=rsa;p=%s", pub))
	}

	return pubContent, err
}

func MakeDkimConfFile(path, domain string) (string, error) {
	// Sanitize domain to prevent path traversal
	safeDomain := sanitizeDomain(domain)
	if safeDomain == "" {
		return "", errors.New("invalid domain name")
	}

	pDir := filepath.Join(path, "dkim")
	if b := tools.IsExist(pDir); !b {
		err := os.MkdirAll(pDir, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	pathDir := filepath.Join(path, "dkim", safeDomain)
	if b := tools.IsExist(pathDir); !b {
		err := os.MkdirAll(pathDir, os.ModePerm)
		if err != nil {
			return "", err
		}
	}

	return MakeDkimFile(path, safeDomain)
}

func GetDomainDkimVal(path, domain string) (string, error) {
	// Sanitize domain to prevent path traversal
	safeDomain := sanitizeDomain(domain)
	if safeDomain == "" {
		return "", errors.New("invalid domain name")
	}

	_, _ = MakeDkimConfFile(path, safeDomain)
	defalutValFile := filepath.Join(path, "dkim", safeDomain, "default.val")
	pubContentRecord, err := tools.ReadFile(defalutValFile)
	return pubContentRecord, err
}
