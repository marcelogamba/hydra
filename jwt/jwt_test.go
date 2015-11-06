package jwt

import (
	"github.com/ory-am/hydra/Godeps/_workspace/src/github.com/stretchr/testify/assert"
	"github.com/ory-am/hydra/Godeps/_workspace/src/github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestMerge(t *testing.T) {
	for k, c := range [][]map[string]interface{}{
		[]map[string]interface{}{
			map[string]interface{}{"foo": "bar"},
			map[string]interface{}{"baz": "bar"},
			map[string]interface{}{"foo": "bar", "baz": "bar"},
		},
		[]map[string]interface{}{
			map[string]interface{}{"foo": "bar"},
			map[string]interface{}{"foo": "baz"},
			map[string]interface{}{"foo": "bar"},
		},
		[]map[string]interface{}{
			map[string]interface{}{},
			map[string]interface{}{"foo": "baz"},
			map[string]interface{}{"foo": "baz"},
		},
		[]map[string]interface{}{
			map[string]interface{}{"foo": "bar"},
			map[string]interface{}{"foo": "baz", "bar": "baz"},
			map[string]interface{}{"foo": "bar", "bar": "baz"},
		},
	} {
		assert.EqualValues(t, c[2], merge(c[0], c[1]), "Case %d", k)
	}
}

func TestLoadCertificate(t *testing.T) {
	for _, c := range TestCertificates {
		out, err := LoadCertificate(c[0])
		assert.Nil(t, err)
		assert.Equal(t, c[1], string(out))
	}
	_, err := LoadCertificate("")
	assert.NotNil(t, err)
	_, err = LoadCertificate("foobar")
	assert.NotNil(t, err)
}

func TestSignRejectsAlgAndTypHeader(t *testing.T) {
	j := New([]byte(TestCertificates[0][1]), []byte(TestCertificates[1][1]))
	for _, c := range []map[string]interface{}{
		map[string]interface{}{"alg": "foo"},
		map[string]interface{}{"typ": "foo"},
		map[string]interface{}{"typ": "foo", "alg": "foo"},
	} {
		_, err := j.SignToken(map[string]interface{}{}, c)
		assert.NotNil(t, err)
	}
}

func TestVerifyPassesHeaderAlgInjection(t *testing.T) {
	j := New([]byte(TestCertificates[0][1]), []byte(TestCertificates[1][1]))
	_, err := j.VerifyToken([]byte("eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.e30.BZXqpeQKnhMtyln2NnoNTUoz_BmyNR-vPHmCxfEpnzCegPZJeCPQiFmn6k7hYhYeWFhH0NhH7-c22-bAf656Esy5qdcxCrwgYSyXAbGQ4C9YsinGcliXeQYcYgOmj8gS2K5Xbj4g9StOB7KywZ_QTJc6FVOqqcgikYVtVA6bMKRrYB4ZS6ZFPdWYTWZ-qOyEg6V7o6-IWmCpEZXlyBgyfAanQkTISMyYuJFPCnFhjnmBUyz0JrWE4gQutOk1-Yw2ikym4GQDrkxrKnnmC_lSJ5I1daxq09oMNj4WRsckktOU64Wuk0PRq_CEpSIA7uHE-Ecgn4ZvRgyLaR1B8S2pAw"))
	assert.NotNil(t, err)
}

func TestSignAndVerify(t *testing.T) {
	for i, c := range []struct {
		private []byte
		public  []byte
		header  map[string]interface{}
		claims  map[string]interface{}
		valid   bool
		signOk  bool
	}{
		{
			[]byte(""),
			[]byte(TestCertificates[1][1]),
			map[string]interface{}{"foo": "bar"},
			map[string]interface{}{"nbf": time.Now().Add(time.Hour)},
			false,
			false,
		},
		{
			[]byte(TestCertificates[0][1]),
			[]byte(""),
			map[string]interface{}{"foo": "bar"},
			map[string]interface{}{"nbf": time.Now().Add(time.Hour)},
			false,
			true,
		},
		{
			[]byte(TestCertificates[0][1]),
			[]byte(TestCertificates[1][1]),
			map[string]interface{}{"foo": "bar"},
			map[string]interface{}{"nbf": time.Now().Add(-time.Hour)},
			false,
			true,
		},
		{
			[]byte(TestCertificates[0][1]),
			[]byte(TestCertificates[1][1]),
			map[string]interface{}{"foo": "bar"},
			map[string]interface{}{"nbf": time.Now().Add(time.Hour)},
			false,
			true,
		},
		{
			[]byte(TestCertificates[0][1]),
			[]byte(TestCertificates[1][1]),
			map[string]interface{}{"foo": "bar"},
			map[string]interface{}{"exp": time.Now().Add(-time.Hour)},
			false,
			true,
		},
		{
			[]byte(TestCertificates[0][1]),
			[]byte(TestCertificates[1][1]),
			map[string]interface{}{"foo": "bar"},
			map[string]interface{}{
				"nbf": time.Now().Add(-time.Hour),
				"iat": time.Now().Add(-time.Hour),
				"exp": time.Now().Add(time.Hour),
			},
			true,
			true,
		},
		{
			[]byte(TestCertificates[0][1]),
			[]byte(TestCertificates[1][1]),
			map[string]interface{}{"foo": "bar"},
			map[string]interface{}{
				"nbf": time.Now().Add(-time.Hour),
			},
			false,
			true,
		},
		{
			[]byte(TestCertificates[0][1]),
			[]byte(TestCertificates[1][1]),
			map[string]interface{}{"foo": "bar"},
			map[string]interface{}{
				"exp": time.Now().Add(time.Hour),
			},
			true,
			true,
		},
		{
			[]byte(TestCertificates[0][1]),
			[]byte(TestCertificates[1][1]),
			map[string]interface{}{"foo": "bar"},
			map[string]interface{}{},
			false,
			true,
		},
	} {
		j := New(c.private, c.public)
		data, err := j.SignToken(c.claims, c.header)
		if c.signOk {
			require.Nil(t, err, "Case %d", i)
		} else {
			require.NotNil(t, err, "Case %d", i)
		}
		tok, err := j.VerifyToken([]byte(data))
		if c.valid {
			require.Nil(t, err, "Case %d", i)
			require.Equal(t, c.valid, tok.Valid, "Case %d", i)
		} else {
			require.NotNil(t, err, "Case %d", i)
		}
	}
}
