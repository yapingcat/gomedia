package rtsp

import (
    "crypto/md5"
    "encoding/base64"
    "encoding/binary"
    "encoding/hex"
    "fmt"
    "strings"
    "sync/atomic"
    "time"
)

type authenticate interface {
    setUri(uri string)
    setMethod(method string)
    setUserInfo(userName, passwd string)
    setRealm(realm string)
    authenticateInfo() string
    decode(string)
    check(string) bool
    wwwAuthenticate() string
}

func createAuthByAuthenticate(auth string) authenticate {
    if strings.HasPrefix(auth, "Basic") {
        return &basicAuth{}
    } else if strings.HasPrefix(auth, "Digest") {
        return &digestAuth{nonceCounter: 0}
    } else {
        panic("unsupport Authorization")
    }
}

type basicAuth struct {
    userName string
    passwd   string
    realm    string
}

func (basic *basicAuth) setUri(uri string)           {}
func (basic *basicAuth) setMethod(method string)     {}
func (basic *basicAuth) setRealm(realm string)       {}
func (basic *basicAuth) decode(authorization string) {}

func (basic *basicAuth) setUserInfo(userName, passwd string) {
    basic.userName = userName
    basic.passwd = passwd
}

func (basic *basicAuth) authenticateInfo() string {
    return fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(basic.userName+":"+basic.passwd)))
}

func (basic *basicAuth) check(authorization string) bool {
    parts := strings.SplitN(authorization, " ", 2)
    if parts[0] != "Basic" {
        return false
    }
    response := strings.TrimSpace(parts[1])
    dst := base64.StdEncoding.EncodeToString([]byte(basic.userName + ":" + basic.passwd))
    if response == dst {
        return true
    } else {
        return false
    }
}

func (basic *basicAuth) wwwAuthenticate() string {
    return "Basic realm=\"" + basic.realm + "\""
}

type digestAuth struct {
    userName     string
    passwd       string
    realm        string
    nonce        string
    uri          string
    method       string
    nonceCounter int32
}

func (digest *digestAuth) setUri(url string) {
    digest.uri = url
}
func (digest *digestAuth) setMethod(method string) {
    digest.method = method
}

func (digest *digestAuth) setUserInfo(userName, passwd string) {
    digest.userName = userName
    digest.passwd = passwd
}

func (digest *digestAuth) setRealm(realm string) {
    digest.realm = realm
}

func (digest *digestAuth) createNonce() string {
    atomic.AddInt32(&digest.nonceCounter, 1)
    t := time.Now().UnixMilli()

    data := make([]byte, 12)
    binary.BigEndian.PutUint32(data, uint32(digest.nonceCounter))
    binary.BigEndian.PutUint64(data[4:], uint64(t))
    result := md5.Sum(data)
    digest.nonce = hex.EncodeToString(result[:])
    return digest.nonce
}

func (digest *digestAuth) wwwAuthenticate() string {
    return "Digest realm=\"" + digest.realm + "\",nonce=\"" + digest.createNonce() + "\""
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
    response := digest.makeResponse()
    digestInfo := fmt.Sprintf("Digest username=\"%s\", realm=\"%s\", nonce=\"%s\", uri=\"%s\", response=\"%s\"",
        digest.userName, digest.realm, digest.nonce, digest.uri, response)
    return digestInfo
}

func (digest *digestAuth) makeResponse() string {
    str1 := digest.userName + ":" + digest.realm + ":" + digest.passwd
    str2 := digest.method + ":" + digest.uri
    md5Bytes1 := md5.Sum([]byte(str1))
    md5Bytes2 := md5.Sum([]byte(str2))
    md5str1 := fmt.Sprintf("%x", md5Bytes1)
    md5str2 := fmt.Sprintf("%x", md5Bytes2)
    str3 := md5str1 + ":" + digest.nonce + ":" + md5str2
    md5Bytes3 := md5.Sum([]byte(str3))
    return fmt.Sprintf("%x", md5Bytes3)
}

func (digest *digestAuth) check(authorization string) bool {
    parts := strings.SplitN(authorization, " ", 2)
    if parts[0] != "Digest" {
        return false
    }
    digestField := strings.TrimSpace(parts[1])
    response := ""
    items := strings.Split(digestField, ",")
    for _, item := range items {
        kv := strings.SplitN(strings.TrimSpace(item), "=", 2)
        if len(kv) < 2 {
            continue
        }
        v := strings.Trim(kv[1], "\"")
        switch kv[0] {
        case "username":
            if v != digest.userName {
                return false
            }
        case "realm":
        case "nonce":
            if digest.nonce != v {
                return false
            }
        case "uri":
            digest.uri = v
        case "response":
            response = v
        }
    }
    if response == digest.makeResponse() {
        return true
    } else {
        return false
    }
}
