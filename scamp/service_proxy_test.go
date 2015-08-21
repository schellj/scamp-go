package scamp

import "testing"

func TestNewServiceProxy(t *testing.T) {
	classRecordsRaw := []byte(`[3,"compute-lxmgvhZpYbl+f/MFpNM5l4Mr","compute",1,5000,"beepish+tls://10.8.1.30:30388",["json","jsonstore","extdirect"],[["FeedIn",["preprocessChunk",""],["preprocessEnd",""],["preprocessStart",""],["processBegin",""],["processChunk",""],["processEnd",""]],["TSV",["feedChunk","","2"],["feedPrepare","","2"]],["_meta",["documentation","noauth"]]],1434680388.47995]`)
	certRaw := []byte(`-----BEGIN CERTIFICATE-----
MIIDpTCCAo2gAwIBAgIJAITp2ZFHlCLrMA0GCSqGSIb3DQEBBQUAMGkxCzAJBgNV
BAYTAlVTMRMwEQYDVQQIDApDYWxpZm9ybmlhMRIwEAYDVQQHDAlTYW4gRGllZ28x
EzARBgNVBAoMCkfDg8K8ZFRlY2gxHDAaBgNVBAMME0F1dGhTZXJ2aWNlL3NvcmVh
cjEwHhcNMTMwMTA4MDMxNDA4WhcNMTMwMjA3MDMxNDA4WjBpMQswCQYDVQQGEwJV
UzETMBEGA1UECAwKQ2FsaWZvcm5pYTESMBAGA1UEBwwJU2FuIERpZWdvMRMwEQYD
VQQKDApHw4PCvGRUZWNoMRwwGgYDVQQDDBNBdXRoU2VydmljZS9zb3JlYXIxMIIB
IjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAqOcmIFeF8oTHUBYbOJSjCjWx
7tED4F4IfMPIbXDgZsqfRzAUP4b5E2HYVbMnWsMcKhlG1446mJyCuz8UmD+710vK
8/pGfuW+2GnRyAU3mtJ73P2QY32LkqHKA56mnBSAPXRXvDfIV6wNRVeAjt+qqBJX
s8baRjKDNae+YCtnUFiyRKR/aA4uKWlzdYaL1fLWn2Y+3KLW8TWdvMd14leo4Jgi
3OHO1QkuvV2s2xKzUbEKz5XFU4v5qce9vPN5IKl08SpKlLFwPckMkbaL2KPM7138
EVa26BUmhjAdrPNxffkyqpzo4WRYt6sKB5pfpEeDKOprPTbbuJLrkQUBLiArbQID
AQABo1AwTjAdBgNVHQ4EFgQUdmhgT1PYaFiZzFpcSCAxxnRSrNkwHwYDVR0jBBgw
FoAUdmhgT1PYaFiZzFpcSCAxxnRSrNkwDAYDVR0TBAUwAwEB/zANBgkqhkiG9w0B
AQUFAAOCAQEABGiODpXGKRk3+n9XiyzgMEvyn+c+iu+tPc57czqMIChR+VH/bUW8
1AxoW3GTkdpP4vnNANzDe/u7iyrLM16re6RPP4lZrOk3dWAQJJOoetJ22ohgREYg
ZZWQ/G7LU4/aoUO8AMQ/64K35jGdQiRxTsELC0Axvb1QPXeeWMOFRZUU315VE2gE
5KWowYQOsF2O2rQEyOV9rFGR7EfYGonmVq0nNPi9GrjPPBOQJ92bUclp4il/3scE
aeaCAeVN4OSatmqWeQE+E+bWNAZgtkX3qPHgrVk2RkS0+jUy4ykiZVnZ4+6PlzOi
w9ljw9ZdZtL7E2XHxXYfEPwcpvF2MsmHLg==
-----END CERTIFICATE-----`)
	sigRaw := []byte(`cKJxGCMJKvdBtY9uI533aIXN1tW6ATaScjk0ecXlwfaYwaL01uxiI926sxi8v/PuxQM0DN7EcgF3
SSySlX65JVo4XdYMlcAhe0Qtci/ottFdh7e5f064Yd6X9A/Y/+1BEyETRNNTUfFwZ/28OGeZ39my
YQjwcpITPewereAlhAR3mvsIx3dYVuF0WAzWYl7hgCVBHktfVW2zVqYKsfpheM6uNglSU44jqB3H
y1yC87J2pwgkT1OKg++uDYunR0OMaX0qC3L1KlZR+zjvkGASpfnltIO//O5RQOoh8p53fezCfc0Y
ufE+JwYJoqxEszo6Adg+VTqt0gPNYu+zdElSTg==`)
	proxy,err := NewServiceProxy(classRecordsRaw, certRaw, sigRaw)
	if err != nil {
		t.Errorf("failed to instantiate new proxy: %s/%s", proxy, err)
	}
}