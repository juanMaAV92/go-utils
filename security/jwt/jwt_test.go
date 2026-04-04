package jwt_test

import (
	"strings"
	"testing"
	"time"

	gjwt "github.com/golang-jwt/jwt/v5"
	"github.com/juanmaAV/go-utils/security/jwt"
)

// Test RSA key pair (2048-bit, for testing only)
const (
	testIssuer  = "test-issuer"
	testPrivKey = `-----BEGIN PRIVATE KEY-----
MIIEvAIBADANBgkqhkiG9w0BAQEFAASCBKYwggSiAgEAAoIBAQDIjA35vss0YS1R
9dPvqUdP0D46xwb6R6kYifVQOef4Rz1sOVaDY3goTFfTLpyAN5W3OcSD6xtJQoMb
fXb/QxeJDNmXBWyYQS4BCiRw/GNZHKWwioKTZBD+tJhXn9Tg+t4uEMVO9h+5PVhK
mNDhNjEbwLgKNAfC/aNUrkCxtsoiTcad69kmJw5FdR2WjuYXafF6fYTkooptbNAr
Cb55R/OQScqrSgyduWAjNSFiyl0OADd39TxpZh8k8fBEqgeEOkkR5QNbwNJj1GDe
tl6Q8utD4hn7JU+2XPn+fsAj3Kl5Fo47E4mJe3t2WYCmjkVHE7r7P0rIlg+d7Zoy
oj6z2dZhAgMBAAECggEAAZCyRTo7fMerYmHgSxUPpOxTqALIp6hqhfIBs6QYDuSD
crZJ2hGrLOlXoCLTft6wMPNm+L6bgmld+5dxl9FuvBeZFSgqLlAH62MoYKdfoSDr
nCKgnUThKxO+wqRRNYZPuJ1R5Olf2wLDDyX9L1zMalKJPS8lxlxTa4RGpfxuvHDK
azkv8Zj+w5IoQ/b4Hn/g+69/qqYx/tKgRDoE/luMZRCA42eipLp2x7msxJEyOyOq
5IMPCNbGDvKcU40FU0STGxICdkzpuLIgSwLaw3ixcklSNTmgixiSTKqBdhQarq7k
gbq1f8/AmP3AjTSJaaGmF10LzDvBLa+Vzu3EMYGBoQKBgQDs0Gd9hyumYK2MfW2T
AALmc3TROdNQf5Q/7XGq5mB9xP4yG9HTWoitouaOniaI+SZTAzzgxlMTgEwJ00/b
SurAJXL6OvymGCGwZTM+7WThzvioYd9zFeZvDxZ+iSedNEIKItE5PUNiZNXFpcfZ
+tLXmliuuJEaYgIxRMdui457CQKBgQDYy3ZCtEiM5ZHYm2PY/OCrpvMBK0ED3gUP
8GApaBdhd8O8pDpjnznzuGJYlx0MN++lzhTs1HfyAeeTyANmYua+52IA9wm64Wug
I9rSZoRRiYxbDbgW0nRXxDZk2LwsqCO/l4aEFop5SmUt4vOnhcdogJz1Anrd/bod
WSwU1p1emQKBgAXlqq4Vj6C1B51YAkKG3YuflGkhZ4G5q5dr8kivM/ftz+aviqoY
tw6b3+HtTkha6/llOz7dsXPq3fngqTxswSvHwvU4QtJgB3a9Dmmiv8BfxxFqXoYf
JX3eglDkWXgwtPRLqaojPGpW2HvzhOaIuHdmAI9ZSUO+7Q8NB2pZT1MBAoGAKnD6
f+iHY3315W/WRj6LRU9zt0Deg4FNgGdQjAqiuSQXH7EO9T3QvJPWPP2oZCH3OoBz
vEGEEc6ppVa8w6iM/8aQexvhvcIvrbQXPKVxNf01iwXOijk9KYlyFKARhrSL+xAQ
937qMQCNekQ56wvXk+/JynVn1Fm9u80fQh0ZxdECgYAq7NWKvDLI1IyKZvtBnknN
nHWFM8cA6z1ca07jMlBSaaByDgTX+QYZsCbr4pSnovzgyU19w8mLNHQV6ajk3l9e
sto3kpGQZhCHYP5CGDT+nYVhYUuZKM39bR7AzRTbPysiusfMq43pqvyRJHn3tAX9
CpKpoQ6FVQIspwwGZAoDUA==
-----END PRIVATE KEY-----`
	testPubKey = `-----BEGIN PUBLIC KEY-----
MIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8AMIIBCgKCAQEAyIwN+b7LNGEtUfXT76lH
T9A+OscG+kepGIn1UDnn+Ec9bDlWg2N4KExX0y6cgDeVtznEg+sbSUKDG312/0MX
iQzZlwVsmEEuAQokcPxjWRylsIqCk2QQ/rSYV5/U4PreLhDFTvYfuT1YSpjQ4TYx
G8C4CjQHwv2jVK5AsbbKIk3GnevZJicORXUdlo7mF2nxen2E5KKKbWzQKwm+eUfz
kEnKq0oMnblgIzUhYspdDgA3d/U8aWYfJPHwRKoHhDpJEeUDW8DSY9Rg3rZekPLr
Q+IZ+yVPtlz5/n7AI9ypeRaOOxOJiXt7dlmApo5FRxO6+z9KyJYPne2aMqI+s9nW
YQIDAQAB
-----END PUBLIC KEY-----`
	wrongPrivKey = `-----BEGIN PRIVATE KEY-----
MIIEvgIBADANBgkqhkiG9w0BAQEFAASCBKgwggSkAgEAAoIBAQDgnbFypAEKtB9C
MSCLhT2N1xY7ZXh4cK4M/eFY8sfof2bsu4zcrx1Cb/IUfcwWimifAXH65SqRXH6b
IadWc0WV6sZH9eymSmxx3SDLraQqYxtrHOrqQ85xKBf/OQvgcnkIns523jccMv6P
ySLaVzka9lAFB3K846JRo1UVx41Gv+dMrSydfpzbXxrc2Q8AozurQpKAWgRiEJkj
ee111sPKv7MHlPHCDjD/wCcEen9jmOA6wU4qMcHwyIKV01hSLuEjElVhK6OpjqcR
4pAFUzPu6XO4zf4H31y6liARMRQ4bRXh50fU6RM9jjlOXKX4QcdDv2SCwfLp0gPe
XReNlbaVAgMBAAECggEAMg8zEwe2K7qFFO16AV5Kn6gGDmrI9J64S7DxXi+NmiV6
vNv3wt9MOBhW7XYz2/ow4u8hhfc5C3h/xvczIjNCmOzgD/2hJlaD4MSVoI8sfT3l
SHQDbv55rgJvhrQiD32zt9Yc0aUoxyIeVdFP8TVrRrUKlHRaf/qDnIo4VkhJFjEX
5OanmHdIkhCkr92woHHuR30akI9mBQVXP9jRzbiV13mZ+TeeK+9ZH7nTz4yvPqK/
+dJXh0ZLzFV1/2oqfFTmpGYT0R2UTPqEpoWMQ9lCRyp2jmeCkTYeFuNhfo3jHsdi
dut3VWuqD2HgZBLwxpLZa6b8PiW4nzGMXtGyDMFCgQKBgQD0zvbXeIihXq7j4xXb
WwKR0xrPRAhhMCH33dz+Qp0OZZ4Tq1XDzlY0h7l+kTsAD0aibnM4M2wVZbCGWSQO
jFvftbh2jD2oASp7orKqrRx0gimMb3/YJ9mkU+38fz4DBuwiZnpZlbWP3/hRYNY4
GjWKYrUyLfCdBqhzkcfPD7u+JQKBgQDq4mm7D+DBqMPd7ehZtp56N02MzXGGyzwh
O8vysi88LUzIo8Oo3AfSI8Y8mYS3HMK4IHbddqcNj9/AsyonjR4TCV07WogFW6S+
7HeIW06VEm/zMIVXUuHxPbkQOK10mFfzeg9C/fuuW8yGMqnanRliGQObuBduIPWw
9o8rJLiTsQKBgE2NFsMxas776JlGgdEIZqr9XhvUqHbpQYl83hopzgkWhdojR7sM
rWBcspV2umMNc8nXBWcDWzT4DtCwgmydaClZLsNXL9z96ZBa/RB2YHJEHZdgZvZ5
wUd+UwDO6j0ZP0qyfgXNGEQopkhZTeNd4iIhnNb1mKiVyF08DDj6+fWFAoGBAKv9
JdZnhffID8PNlLk8U0bCf/J97Ib7AgiDtI79kkDKGtM/tuFKXB+vOlTdRKSgVqRk
gSUg4Km5k+mSR2e2mTLvRGlVnQvvUu7KT6x0z1GwsnCsMrcCZZczzvLlzXz2oFAU
LGCtgUDmzxfkuSLurct67X10ixOE5uKxZ5v7w3vBAoGBAMDCFsEr61diWvbQnGX0
W7RxmRhUyN7ozzVmyStZKe07inuDXJJVS7zNG2/BzV/jQvlCS8aHq89chMNo7cOU
bTfe0m7Rkb29Q8LhDGo/bsdTCh1Pi1oSp3r2egMd+f4dmqpz2qxyvhDUnMHW2qNN
tFzh1zY+5azfddoaRH5j7PxT
-----END PRIVATE KEY-----`
)

type testClaims struct {
	gjwt.RegisteredClaims
	UserID string   `json:"user_id"`
	Roles  []string `json:"roles"`
}

func newSvc(t *testing.T) *jwt.TokenService {
	t.Helper()
	svc, err := jwt.New(testPrivKey, testPubKey, testIssuer)
	if err != nil {
		t.Fatalf("jwt.New: %v", err)
	}
	return svc
}

func TestNew_InvalidKey(t *testing.T) {
	_, err := jwt.New("not-a-pem", "", "issuer")
	if err == nil {
		t.Error("expected error for invalid private key PEM")
	}
	_, err = jwt.New("", "not-a-pem", "issuer")
	if err == nil {
		t.Error("expected error for invalid public key PEM")
	}
}

func TestGenerateAndValidate(t *testing.T) {
	svc := newSvc(t)

	claims := testClaims{
		RegisteredClaims: svc.RegisteredClaims(time.Hour),
		UserID:           "usr_123",
		Roles:            []string{"admin"},
	}

	token, err := svc.GenerateToken(&claims)
	if err != nil || token == "" {
		t.Fatalf("GenerateToken: %v", err)
	}

	got, err := jwt.ValidateToken[testClaims, *testClaims](svc, token)
	if err != nil {
		t.Fatalf("ValidateToken: %v", err)
	}
	if got.UserID != "usr_123" {
		t.Errorf("UserID = %q, want \"usr_123\"", got.UserID)
	}
	if got.Issuer != testIssuer {
		t.Errorf("Issuer = %q, want %q", got.Issuer, testIssuer)
	}
	if len(got.Roles) != 1 || got.Roles[0] != "admin" {
		t.Errorf("Roles = %v", got.Roles)
	}
}

func TestValidateToken_Expired(t *testing.T) {
	svc := newSvc(t)

	claims := testClaims{
		RegisteredClaims: gjwt.RegisteredClaims{
			ExpiresAt: gjwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
		UserID: "usr_expired",
	}
	token, _ := svc.GenerateToken(&claims)

	_, err := jwt.ValidateToken[testClaims, *testClaims](svc, token)
	if err == nil {
		t.Error("expected error for expired token")
	}
}

func TestValidateTokenIgnoringExpiration(t *testing.T) {
	svc := newSvc(t)

	claims := testClaims{
		RegisteredClaims: gjwt.RegisteredClaims{
			ExpiresAt: gjwt.NewNumericDate(time.Now().Add(-time.Hour)),
		},
		UserID: "usr_refresh",
	}
	token, _ := svc.GenerateToken(&claims)

	got, err := jwt.ValidateTokenIgnoringExpiration[testClaims, *testClaims](svc, token)
	if err != nil {
		t.Fatalf("ValidateTokenIgnoringExpiration: %v", err)
	}
	if got.UserID != "usr_refresh" {
		t.Errorf("UserID = %q, want \"usr_refresh\"", got.UserID)
	}
}

func TestValidateToken_TamperedPayload(t *testing.T) {
	svc := newSvc(t)

	claims := testClaims{
		RegisteredClaims: svc.RegisteredClaims(time.Hour),
		UserID:           "usr_legit",
	}
	token, _ := svc.GenerateToken(&claims)

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		t.Fatal("expected 3-part JWT")
	}
	parts[1] += "tampered"
	tampered := strings.Join(parts, ".")

	_, err := jwt.ValidateToken[testClaims, *testClaims](svc, tampered)
	if err == nil {
		t.Error("expected error for tampered token")
	}
}

func TestValidateToken_WrongKey(t *testing.T) {
	svc := newSvc(t)
	other, _ := jwt.New(wrongPrivKey, testPubKey, testIssuer)

	claims := testClaims{
		RegisteredClaims: other.RegisteredClaims(time.Hour),
		UserID:           "usr_wrong",
	}
	token, _ := other.GenerateToken(&claims)

	_, err := jwt.ValidateToken[testClaims, *testClaims](svc, token)
	if err == nil {
		t.Error("expected error: token signed with different key")
	}
}

func TestGenerateToken_NoPrivateKey(t *testing.T) {
	svc, _ := jwt.New("", testPubKey, testIssuer)
	_, err := svc.GenerateToken(&testClaims{})
	if err == nil {
		t.Error("expected error when no private key")
	}
}

func TestValidateToken_NoPublicKey(t *testing.T) {
	svc, _ := jwt.New(testPrivKey, "", testIssuer)
	_, err := jwt.ValidateToken[testClaims, *testClaims](svc, "any.token.here")
	if err == nil {
		t.Error("expected error when no public key")
	}
}
