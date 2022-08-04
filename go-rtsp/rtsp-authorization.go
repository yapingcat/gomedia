package rtsp

import (
	"crypto/md5"
	"encoding/base64"
	"fmt"
	"strings"
)

type authenticate interface {
	setUri(uri string)
	setMethod(method string)
	setUserInfo(userName, passwd string)
	authenticateInfo() string
	decode(string)
}

func createAuthByAuthenticate(auth string) authenticate {
	if strings.HasPrefix(auth, "Basic") {
		return &basicAuth{}
	} else if strings.HasPrefix(auth, "Digest") {
		return &digestAuth{}
	} else {
		panic("unsupport Authorization")
	}
}

type basicAuth struct {
	userName string
	passwd   string
}

func (basic *basicAuth) setUri(uri string)           {}
func (basic *basicAuth) setMethod(method string)     {}
func (basic *basicAuth) decode(authorization string) {}

func (basic *basicAuth) setUserInfo(userName, passwd string) {
	basic.userName = userName
	basic.passwd = passwd
}

func (basic *basicAuth) authenticateInfo() string {
	return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(basic.userName+":"+basic.passwd)))
}

type digestAuth struct {
	userName string
	passwd   string
	realm    string
	nonce    string
	uri      string
	method   string
}

func (digest *digestAuth) setUri(url string) {
	digest.uri = url
}
func (digest *digestAuth) setMethod(method string) {
	digest.uri = method
}

func (digest *digestAuth) setUserInfo(userName, passwd string) {
	digest.userName = userName
	digest.passwd = passwd
}

func (digest *digestAuth) decode(authorization string) {
	elems := strings.Split(strings.TrimPrefix(authorization, "Digest"), ",")
	for i := 0; i < len(elems); i++ {
		elem := strings.TrimSpace(elems[i])
		if strings.Contains(elem, "realm") {
			fmt.Sscanf(elem, "realm=\"%s\"", &digest.realm)
			digest.realm = digest.realm[:len(digest.realm)-1]
		} else if strings.Contains(elem, "nonce") {
			fmt.Sscanf(elem, "nonce=\"%s\"", &digest.nonce)
			digest.nonce = digest.nonce[:len(digest.nonce)-1]
		}
	}
}

//response=md5(md5(username:realm:password):nonce:md5(method:url));
func (digest *digestAuth) authenticateInfo() string {
	str1 := digest.userName + ":" + digest.realm + ":" + digest.passwd
	str2 := digest.method + ":" + digest.uri
	md5Bytes1 := md5.Sum([]byte(str1))
	md5Bytes2 := md5.Sum([]byte(str2))
	md5str1 := fmt.Sprintf("%x", md5Bytes1)
	md5str2 := fmt.Sprintf("%x", md5Bytes2)
	str3 := md5str1 + ":" + digest.nonce + ":" + md5str2
	md5Bytes3 := md5.Sum([]byte(str3))
	response := fmt.Sprintf("%x", md5Bytes3)
	digestInfo := fmt.Sprintf("Digest username=\"%s\", realm=\"%s\", nonce=\"%s\", uri=\"%s\", response=\"%s\"",
		digest.userName, digest.realm, digest.nonce, digest.uri, response)
	return digestInfo
}
