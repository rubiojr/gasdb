package api

import (
	"crypto/tls"
	"crypto/x509"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFuelPriceAPI_TLS(t *testing.T) {
	t.Setenv("GODEBUG", "tlsrsakex=0")

	tests := []struct {
		name    string
		version uint16
		cipher  uint16
	}{
		{"legacy RSA", tls.VersionTLS12, tls.TLS_RSA_WITH_AES_256_GCM_SHA384},
		{"ECDHE", tls.VersionTLS12, tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256},
		{"TLS 1.3", tls.VersionTLS13, tls.TLS_AES_128_GCM_SHA256},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = io.WriteString(w, `{"ResultadoConsulta":"OK","ListaEESSPrecio":[{"IDEESS":"1"}]}`)
			}))
			server.Config.ErrorLog = log.New(io.Discard, "", 0)
			server.TLS = &tls.Config{
				MinVersion:   tt.version,
				MaxVersion:   tt.version,
				CipherSuites: []uint16{tt.cipher},
			}
			server.StartTLS()
			t.Cleanup(server.Close)

			defaultTransport := http.DefaultTransport.(*http.Transport)
			client := NewFuelPriceAPI()
			client.baseURL = server.URL
			transport := client.httpClient.Transport.(*http.Transport)
			t.Cleanup(transport.CloseIdleConnections)
			require.NotSame(t, defaultTransport, transport)

			// The compatibility exception must not disable certificate verification.
			_, err := client.FetchPrices()
			var certificateError *tls.CertificateVerificationError
			require.ErrorAs(t, err, &certificateError)

			roots := x509.NewCertPool()
			roots.AddCert(server.Certificate())
			transport.TLSClientConfig.RootCAs = roots
			prices, err := client.FetchPrices()
			require.NoError(t, err)
			require.Equal(t, ApiResultOK, prices.ResultadoConsulta)
			require.Len(t, prices.ListaEESSPrecio, 1)

			if tt.cipher == tls.TLS_RSA_WITH_AES_256_GCM_SHA384 {
				// A separate client using default cipher suites must still reject RSA key exchange.
				otherTransport := defaultTransport.Clone()
				otherTransport.TLSClientConfig = &tls.Config{RootCAs: roots, MinVersion: tls.VersionTLS12}
				t.Cleanup(otherTransport.CloseIdleConnections)
				otherClient := &http.Client{Transport: otherTransport, Timeout: DefaultTimeout}
				resp, err := otherClient.Get(server.URL)
				if resp != nil {
					_ = resp.Body.Close()
				}
				require.Error(t, err)
			}
		})
	}
}
