package ca

import (
	"bytes"
	"crypto/x509"
	"fmt"

	"github.com/hashicorp/consul/agent/connect"
)

func validateSetIntermediate(
	intermediatePEM, rootPEM string,
	currentPrivateKey string, // optional
	spiffeID *connect.SpiffeIDSigning,
) error {
	// Get the key from the incoming intermediate cert so we can compare it
	// to the currently stored key.
	intermediate, err := connect.ParseCert(intermediatePEM)
	if err != nil {
		return fmt.Errorf("error parsing intermediate PEM: %v", err)
	}

	if currentPrivateKey != "" {
		privKey, err := connect.ParseSigner(currentPrivateKey)
		if err != nil {
			return err
		}

		// Compare the two keys to make sure they match.
		b1, err := x509.MarshalPKIXPublicKey(intermediate.PublicKey)
		if err != nil {
			return err
		}
		b2, err := x509.MarshalPKIXPublicKey(privKey.Public())
		if err != nil {
			return err
		}
		if !bytes.Equal(b1, b2) {
			return fmt.Errorf("intermediate cert is for a different private key")
		}
	}

	// Validate the remaining fields and make sure the intermediate validates against
	// the given root cert.
	if !intermediate.IsCA {
		return fmt.Errorf("intermediate is not a CA certificate")
	}
	if uriCount := len(intermediate.URIs); uriCount != 1 {
		return fmt.Errorf("incoming intermediate cert has unexpected number of URIs: %d", uriCount)
	}
	if got, want := intermediate.URIs[0].String(), spiffeID.URI().String(); got != want {
		return fmt.Errorf("incoming cert URI %q does not match current URI: %q", got, want)
	}

	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM([]byte(rootPEM))
	_, err = intermediate.Verify(x509.VerifyOptions{
		Roots: pool,
	})
	if err != nil {
		return fmt.Errorf("could not verify intermediate cert against root: %v", err)
	}

	return nil
}
