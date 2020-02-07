package k8s

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/util/cert"
)

const (
	rsaBits = 2048
)

type TLSGenerator struct{}

func CreateTLSCertSecret(namespaceName, secretName, commonName string) (*v1.Secret, error) {
	signingPriv, err := rsa.GenerateKey(cryptorand.Reader, rsaBits)
	if err != nil {
		return nil, err
	}
	config := cert.Config{
		CommonName: commonName,
	}
	cacert, err := cert.NewSelfSignedCACert(config, signingPriv)
	if err != nil {
		return nil, err
	}
	var serverKey, serverCert bytes.Buffer

	if err := pem.Encode(&serverCert, &pem.Block{Type: "CERTIFICATE", Bytes: cacert.Raw}); err != nil {
		return nil, fmt.Errorf("failed creating cert: %v", err)
	}
	if err := pem.Encode(&serverKey, &pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(signingPriv)}); err != nil {
		return nil, fmt.Errorf("failed creating key: %v", err)
	}

	data := map[string][]byte{
		v1.TLSCertKey:       serverCert.Bytes(),
		v1.TLSPrivateKeyKey: serverKey.Bytes(),
	}

	secret := &v1.Secret{
		Data: data,
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespaceName,
		},
	}

	return clientset.CoreV1().Secrets(namespaceName).Create(secret)
}
