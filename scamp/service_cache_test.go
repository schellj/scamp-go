package scamp

import "testing"
import "bytes"
import "bufio"
import "os"

func TestScanCertificate(t *testing.T) {
	reader := bytes.NewReader(testCertificateToScan)

	scanner := bufio.NewScanner(reader)
	scanner.Split(scanCertficates)

	if !scanner.Scan() {
		t.Errorf("failed to scan")
	}

	if !bytes.Equal(scanner.Bytes(), testCertificateToScan) {
		t.Errorf("certificate did not match:\n`%v`\n`%v`", scanner.Bytes(), testCertificateToScan)
	}
}

func BenchmarkReadingProductionAnnounceCache(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			file,err := os.Open("/Users/xavierlange/code/gudtech/scamp-patrol/fixtures/discovery.sample")
			if err != nil {
				panic("could not open file")
			}
			
			cache := NewServiceCache()
			err = cache.LoadAnnounceCache(bufio.NewScanner(file))
			if err != nil {
				panic("failed to load announce cache")
			}
		}
	})
}

func TestReadAnnounceCache(t *testing.T) {
	initSCAMPLogger()

	file,err := os.Open("/Users/xavierlange/code/gudtech/scamp-go/fixtures/announce_cache")
	if err != nil {
		return
	}

	s := bufio.NewScanner(file)

	cache := NewServiceCache()
	err = cache.LoadAnnounceCache(s)
	if err != nil {
		t.Errorf("unexpected error parsing announce cache: `%s`", err)
	}

	if cache.Size() != 32 {
		t.Errorf("expected cache size to be 32 but was %d", cache.Size())
	}
}

func TestRegisterOnServiceCache(t *testing.T) {
	cache := NewServiceCache()
	serviceInstance := new(ServiceProxy)
	serviceInstance.ident = "bob"

	cache.Store(serviceInstance)
	retrieved := cache.Retrieve("bob")
	if retrieved == nil {
		t.Errorf("retrieved nothing")
	}
}

var testCertificateToScan = []byte(`
-----BEGIN CERTIFICATE-----
MIIFqzCCA5OgAwIBAgIJALlP+LRar+aUMA0GCSqGSIb3DQEBBQUAMGwxJzAlBgNV
BAMMHmtzYXdoMS5kZXYuZ3VkdGVjaC5jb20gcGF5bWVudDELMAkGA1UECgwCR1Qx
CzAJBgNVBAYTAlVTMRMwEQYDVQQIDApDYWxpZm9ybmlhMRIwEAYDVQQHDAlTYW4g
RGllZ28wHhcNMTQwMjIyMDY1MjUyWhcNMjQwMjIwMDY1MjUyWjBsMScwJQYDVQQD
DB5rc2F3aDEuZGV2Lmd1ZHRlY2guY29tIHBheW1lbnQxCzAJBgNVBAoMAkdUMQsw
CQYDVQQGEwJVUzETMBEGA1UECAwKQ2FsaWZvcm5pYTESMBAGA1UEBwwJU2FuIERp
ZWdvMIICIjANBgkqhkiG9w0BAQEFAAOCAg8AMIICCgKCAgEA0uoAPSpP5Gdq9vut
MIyhWUwwbTr8VXpdMJcdmCESQd/6dbas2F7bujusblujed/CWdPRomjZPwuq7z/9
vlUnIvN7ZKCc1q7vnV0hDMfBeZAuMHYhHgywU4O/DumvxYsPUayr6Bj1rfOjVkyY
QsiF8pGAZa7+/H2IdrdTNAxgLLlQmaYDCU4ckvyLcOm/MICKfyIORn3bmhcpTbmG
rYtJbR8FgsBAQxhv1GiYQXCNYGVbERGSVw8Pw4Zk4S8yHG3wiJoEvoc2WBYhBDGA
ZhkrdH1/UNscjsIiabF4+E1mAahB9k0+4844vuktLbQtDpJLp5upeWDAVN5pCduw
946C8U2wma23WMTFC9NDwYOHJpDeXxu4K8F3GJIvN1YVzIjLqVV0AU6dXHHJSCx2
pAc7ZxwcT8XgJ17AgFyfRmLNr2TfItozoLXaHMmWWLDPdJEkiLVtBRowQkocbDPc
Ih2y4iEM3cO9hcNwuU4oyjM6asudB9PeUS3jk7OAoZE9BYv3gkZuoxWk3FoCwpoL
2QW26hiaO5fIhTcwEKjb/Qm9wDNkUlSgxExJ2adV+psXZTwk3HlrHL6k323E8YEC
THHnFvYXWrXrGw5Vg0OyLl5vQUqSbzNW5yQtFhIhOc5rFONTHuTnLZ74Vlm17DyF
6o+VNFGkHdDwDS68bdaTz7VIS20CAwEAAaNQME4wHQYDVR0OBBYEFMha7Tsi+kBH
0GyrSGUJxzP/GdVNMB8GA1UdIwQYMBaAFMha7Tsi+kBH0GyrSGUJxzP/GdVNMAwG
A1UdEwQFMAMBAf8wDQYJKoZIhvcNAQEFBQADggIBAAAZaoQJCwIDLeic7ps8UO8f
UmN2gK/APHEsY5Ev4mNxFcIiUAJlHDMJ5MvinacakanIkQMYrDPmLPCeR9rBYk5d
EOjhdrUwfgx+cxlJmt5Q4zHYjtfHvIINVIsnUus4mb9ONHGEwYFPCZHCHpgmW/qd
HWu7huXsnpyHhA29lmoOtxJIULR4xPNWh09pouc+YXdTK9F5KoPCskWI0z4Xvo3K
NmttiNERg86W6S7yQO1Zj5Bwu5t2uBAmeJsOpIJvSW5MOWPDvCBSAWy085aO7eoE
ttRl+KGbLVbVRR3Sbg3fnweR5hEA/C50i5iVTZUERQ+c2jgvcG6WM5uJ1xCkxxsb
7bn+vUwWDyHv86vQbt8Uq02KIfp1MGC3la12M7v6UNgYq/Ruhbr6E4HjHjswo3Uc
yAdMtpSgXVj+rczQLNO/NCGVnpkEge8qcBAgoC2EfLpCsOhclhwtnwkom8WXi4ga
s5chuQjibCKkBh+dEPDnjjM/WWmWzoQtK+esmPLp1sNRJDc/dwgcaxAHTG5yiS5r
GXM51rweYWKZZ3TbShekXaNFDct97dFgXBWuTvB/5X7bd1Uu0X9U5GqINbjRSoUL
0FPRAZnwPS8XQRj76fO0EiKyHZMZMnfS78vzFuOxR+vOaFqWzZ0TVdOuyQRZqW7w
r/6q4/7qnfRgT6Clnxtf
-----END CERTIFICATE-----`)