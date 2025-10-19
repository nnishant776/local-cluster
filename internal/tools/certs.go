package tools

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"time"

	"dario.cat/mergo"
	"github.com/helmfile/helmfile/pkg/yaml"
	"github.com/nnishant776/errstack"
	"github.com/nnishant776/local-cluster/internal/utils"
	"github.com/spf13/cobra"
)

func newGenerateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate signed root CA certificate and private key",
		Long:  "Generate signed root CA certificate and private key. This will print the generated private key and the CA certificate to the stdout as well as update the secrets section of the config with the same value and write the key and certificate to their respective files (named root.crt and root.key)",
		RunE: func(cmd *cobra.Command, args []string) error {
			configDir := utils.GetAppConfigDir()
			configPath := filepath.Join(configDir, "config.yaml")
			cfg, rawConfigData, err := utils.ParseConfig(configPath)
			if err != nil {
				return err
			}

			privKey, err := rsa.GenerateKey(rand.Reader, 4096)
			if err != nil {
				return errstack.NewChainString(
					"certificate: failed to generate private key",
				).Chain(err)
			}

			caSpec := &x509.Certificate{
				SerialNumber: big.NewInt(2025),
				Subject: pkix.Name{
					Country:            []string{cfg.Infra.TLS.CACerificateParams.CountryCode},
					Organization:       []string{cfg.Infra.TLS.CACerificateParams.Organization},
					OrganizationalUnit: []string{cfg.Infra.TLS.CACerificateParams.Unit},
					Locality:           []string{cfg.Infra.TLS.CACerificateParams.Locality},
					Province:           []string{cfg.Infra.TLS.CACerificateParams.State},
					StreetAddress:      []string{cfg.Infra.TLS.CACerificateParams.StreetAddress},
					PostalCode:         []string{cfg.Infra.TLS.CACerificateParams.PostalCode},
					SerialNumber:       "",
					CommonName:         cfg.Infra.TLS.CACerificateParams.CommonName,
				},
				KeyUsage:                    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
				Extensions:                  []pkix.Extension{},
				ExtraExtensions:             []pkix.Extension{},
				UnhandledCriticalExtensions: []asn1.ObjectIdentifier{},
				BasicConstraintsValid:       true,
				IsCA:                        true,
				ExtKeyUsage: []x509.ExtKeyUsage{
					x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth,
				},
			}

			caSpec.NotBefore, err = time.Parse(
				time.DateOnly, cfg.Infra.TLS.CACerificateParams.IssueDate,
			)
			if err != nil {
				return err
			}

			caSpec.NotAfter, err = time.Parse(
				time.DateOnly, cfg.Infra.TLS.CACerificateParams.ExpiryDate,
			)
			if err != nil {
				return err
			}

			caCertBytes, err := x509.CreateCertificate(rand.Reader, caSpec, caSpec, &privKey.PublicKey, privKey)
			if err != nil {
				return errstack.NewChainString(
					"certificate: failed to certificate",
				).Chain(err)
			}

			privKeyBytes := x509.MarshalPKCS1PrivateKey(privKey)
			caPem := bytes.NewBuffer(nil)
			privKeyPem := bytes.NewBuffer(nil)

			pem.Encode(caPem, &pem.Block{
				Type:  "CERTIFICATE",
				Bytes: caCertBytes,
			})
			pem.Encode(privKeyPem, &pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: privKeyBytes,
			})
			os.WriteFile("root.key", privKeyPem.Bytes(), 0o644)
			os.WriteFile("root.crt", caPem.Bytes(), 0o644)
			fmt.Printf("Private Key:\n%s\n", privKeyPem)
			fmt.Printf("CA Certificate:\n%s\n", caPem)

			type dict map[string]any
			overrideData := dict{
				"secrets": dict{
					"tls": dict{
						"ca": dict{
							"certificate": caPem.String(),
							"privateKey":  privKeyPem.String(),
						},
					},
				},
			}

			err = mergo.MergeWithOverwrite(&rawConfigData, (map[string]any)(overrideData), func(c *mergo.Config) {
				c.Overwrite = true
			})
			if err != nil {
				return errstack.NewChainString(
					"certificate: failed to merge configuration",
				).Chain(err)
			}

			buf, err := yaml.Marshal(rawConfigData)
			if err != nil {
				return errstack.NewChainString(
					"certificate: failed to marshal configuration",
				).Chain(err)
			}

			err = os.WriteFile(configPath, buf, 0o644)
			if err != nil {
				return errstack.NewChainString(
					"certificate: failed to marshal configuration",
				).Chain(err)
			}

			return nil
		},
	}

	return cmd
}

func NewCertificateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:                "certificate",
		Long:               "Commands to do various certificate operations",
		DisableFlagParsing: true,
	}

	cmd.AddCommand(newGenerateCommand())

	return cmd
}
